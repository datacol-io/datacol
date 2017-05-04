package google

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"cloud.google.com/go/datastore"
	log "github.com/Sirupsen/logrus"
	"github.com/dinesh/datacol/client/models"
	dm "google.golang.org/api/deploymentmanager/v2"
	"google.golang.org/api/googleapi"
	sql "google.golang.org/api/sqladmin/v1beta4"
)

const stackKind = "Stack"

type initOptions struct {
	ClusterName                   string
	MachineType                   string
	DiskSize                      int64
	NumNodes                      int
	ProjectNumber                 int64
	Zone, Bucket                  string
	ClusterNotExists, Preemptible bool
}

func (g *GCPCloud) StackSave(s *models.Stack) error {
	key := datastore.NameKey(stackKind, s.Name, nil)
	_, err := g.datastore().Put(context.TODO(), key, s)

	return err
}

func (g *GCPCloud) fetchStack() {
	if len(g.BucketName) > 0 {
		return
	}

	log.Debugf("fetching stack details(%s) from datastore", g.DeploymentName)

	st := new(models.Stack)
	store := g.datastore()
  defer store.Close()

	if err := store.Get(context.TODO(), g.stackKey(), st); err != nil {
		if err.Error() == datastore.ErrNoSuchEntity.Error() {
			log.Fatal(fmt.Errorf("Unable to find stack by name: %s", g.DeploymentName))
		} else {
			log.Fatal(err)
		}
	}

	g.BucketName = st.Bucket
	g.Zone = st.Zone
	g.ProjectNumber = st.PNumber
	g.ServiceKey = st.ServiceKey
}

func (g *GCPCloud) Initialize(cluster, machineType string, nodes int, preemt bool) error {
	name := g.DeploymentName
	log.Infof("creating new stack %s", name)

	g.fetchStack()
	cnexists := true

	if len(cluster) == 0 {
		cluster = fmt.Sprintf("%v-cluster", g.DeploymentName)
	} else {
		cnexists = false
	}

	if len(machineType) == 0 {
		machineType = ditermineMachineType(nodes)
	}

	options := initOptions{
		ClusterName:      cluster,
		DiskSize:         10,
		NumNodes:         nodes,
		MachineType:      machineType,
		Zone:             g.Zone,
		Bucket:           g.BucketName,
		ProjectNumber:    g.ProjectNumber,
		ClusterNotExists: cnexists,
		Preemptible:      preemt,
	}

	log.Debug(toJson(options))

	imports := []*dm.ImportFile{
		{
			Name:    "registry.jinja",
			Content: registryYAML,
		},
	}

	if cnexists {
		imports = append(imports, &dm.ImportFile{
			Name:    "container-vm.jinja",
			Content: vmYAML,
		})
	}

	req := &dm.Deployment{
		Name: name,
		Target: &dm.TargetConfiguration{
			Imports: imports,
			Config: &dm.ConfigFile{
				Content: compileTmpl(configYAML, &options),
			},
		},
	}

	service := g.deploymentmanager()

	op, err := service.Deployments.Insert(g.Project, req).Do()
	if err != nil {
		if gerr, ok := err.(*googleapi.Error); ok && gerr.Code == 409 {
			return fmt.Errorf("Failed: Deployment %s is already created. Please destroy.", name)
		} else {
			return err
		}
	}

	if err = g.waitForDpOp(service, op, true); err != nil {
		return err
	}

	kube, err := g.getCluster(cluster)
	cfgroot := filepath.Join(models.ConfigPath, g.DeploymentName)
	return GenerateClusterConfig(name, cfgroot, kube)
}

func (g *GCPCloud) Teardown() error {
	g.fetchStack()

	name := g.DeploymentName
	dmService := g.deploymentmanager()

	// add fingerprint of in stop request
	// if op, err := dmService.Deployments.Stop(g.Project, name, &dm.DeploymentsStopRequest{}).Do(); err != nil {
	//   fmt.Printf(err.Error())
	//   gerr, ok := err.(*googleapi.Error)
	//   if !ok  || (ok && gerr.Code != 412) {
	//     return fmt.Errorf("waiting to stop deployment: %v", err)
	//   }
	// }

	log.Infof("Deleting stack %s", name)

	gsService := g.storage()
	resp, err := gsService.Objects.List(g.BucketName).Do()
	if err != nil {
		return err
	}

	for _, obj := range resp.Items {
		if err := gsService.Objects.Delete(obj.Bucket, obj.Name).Do(); err != nil {
			return err
		}
	}

	op, err := dmService.Deployments.Delete(g.Project, name).Do()
	if err != nil {
		return err
	}

	if err = g.waitForDpOp(dmService, op, false); err != nil {
		return fmt.Errorf("deleting stack %v", err)
	}

	return g.resetDatabase()
}

func (g *GCPCloud) updateDeployment(
	service *dm.Service,
	dp *dm.Deployment,
	manifest *dm.Manifest,
	content string,
) error {
	dp.Target = &dm.TargetConfiguration{
		Config:  &dm.ConfigFile{Content: content},
		Imports: manifest.Imports,
	}

	op, err := service.Deployments.Update(g.Project, g.DeploymentName, dp).Do()
	if err != nil {
		if gerr, ok := err.(*googleapi.Error); ok && gerr.Code == 403 {
			// TODO: better error message
			return err
		}
		return err
	}

	if err = g.waitForDpOp(service, op, false); err != nil {
		return err
	}
	return err
}

func waitForSqlOp(svc *sql.Service, op *sql.Operation, project string) error {
	log.Debugf("Waiting for %s [%v]", op.OperationType, op.Name)

	for {
		time.Sleep(2 * time.Second)
		op, err := svc.Operations.Get(project, op.Name).Do()
		if err != nil {
			return err
		}

		switch op.Status {
		case "PENDING", "RUNNING":
			fmt.Print(".")
			continue
		case "DONE":
			if op.Error != nil {
				var last error
				for _, operr := range op.Error.Errors {
					last = fmt.Errorf("%v", operr.Message)
				}
				// try to teardown if, just ignore error if any
				log.Errorf("sqlAdmin Operation failed: %v, Canceling ..", last)
				return last
			} else {
        return nil
      }
		default:
			return fmt.Errorf("Unknown status %q: %+v", op.Status, op)
		}
	}
}

func (g *GCPCloud) stackKey() *datastore.Key {
	return datastore.NameKey(stackKind, g.DeploymentName, nil)
}

func (g *GCPCloud) nestedKey(kind, key string) *datastore.Key {
	return datastore.NameKey(kind, key, g.stackKey())
}

func (g *GCPCloud) waitForDpOp(svc *dm.Service, op *dm.Operation, interrupt bool) error {
	log.Infof("Waiting on %s [%v]", op.OperationType, op.Name)
	project := g.Project

	cancelCh := make(chan os.Signal, 1)
	signal.Notify(cancelCh, os.Interrupt, syscall.SIGTERM)

	for {
		time.Sleep(2 * time.Second)
		op, err := svc.Operations.Get(project, op.Name).Do()
		if err != nil {
			return err
		}

		select {
		case <-cancelCh:
			if interrupt {
				return g.Teardown()
			} else {
				return nil
			}
		default:
		}

		switch op.Status {
		case "PENDING", "RUNNING":
			fmt.Print(".")
			continue
		case "DONE":
			if op.Error != nil {
				var last error
				for _, operr := range op.Error.Errors {
					last = fmt.Errorf("%v", operr.Message)
				}
				// try to teardown if, just ignore error if any
				log.Errorf("Deployment failed: %v, Canceling ..", last)

				if interrupt {
					if err := g.Teardown(); err != nil {
						log.Debugf("deleting stack: %+v", err)
					}
				}
				return last
			}
			return nil
		default:
			return fmt.Errorf("Unknown status %q: %+v", op.Status, op)
		}
	}
}

func (g *GCPCloud) resetDatabase() error {
	apps, err := g.AppList()
	if err != nil {
		return err
	}

	store := g.datastore()
	ctx := context.TODO()

	// delete apps, builds, releases
	for _, app := range apps {
		q := datastore.NewQuery(buildKind).Ancestor(g.stackKey()).
			Filter("app =", app.Name).KeysOnly()
		keys, err := store.GetAll(ctx, q, nil)
		if err != nil {
			return err
		}
		err = store.DeleteMulti(ctx, keys)
		if err != nil {
			return err
		}

		q = datastore.NewQuery(releaseKind).Ancestor(g.stackKey()).
			Filter("app =", app.Name).KeysOnly()
		keys, err = store.GetAll(ctx, q, nil)
		if err != nil {
			return err
		}
		err = store.DeleteMulti(ctx, keys)
		if err != nil {
			return err
		}

		if err = store.Delete(ctx, g.nestedKey(appKind, app.Name)); err != nil {
			return err
		}
	}

	// delete resources
	q := datastore.NewQuery(resourceKind).Ancestor(g.stackKey()).KeysOnly()
	keys, err := store.GetAll(ctx, q, nil)
	if err != nil {
		return err
	}
	if err = store.DeleteMulti(ctx, keys); err != nil {
		return err
	}

	// delete stack
	return store.Delete(ctx, g.stackKey())
}

func getManifest(service *dm.Service, project, stack string) (*dm.Deployment, *dm.Manifest, error) {
	dp, err := service.Deployments.Get(project, stack).Do()
	if err != nil {
		return nil, nil, err
	}

	parts := strings.Split(dp.Manifest, "/")
	mname := parts[len(parts)-1]
	m, err := service.Manifests.Get(project, stack, mname).Do()

	return dp, m, err
}

func resourceFromStack(service *dm.Service, project, stack, name string) (*models.Resource, error) {
	return &models.Resource{Name: name }, nil
}

const registryYAML = `
resources:
- name: {{ env['name'] }}
  type: storage.v1.bucket
  properties:
    location: {{ properties['bucketLocation'] }}
    projectNumber: {{ properties['projectNumber'] }}
`

const vmYAML = `
{% set CLUSTER_NAME = env['name'] %}
{% set TYPE_NAME = CLUSTER_NAME + '-type' %}

resources:
- name: {{ CLUSTER_NAME }}
  type: container.v1.cluster
  properties:
    zone: {{ properties['zone'] }}
    cluster:
      name: {{ CLUSTER_NAME }}
      description: "cluster created by datacol.io"
      initialNodeCount: {{ properties['numNodes'] }}
      enableKubernetesAlpha: false
      nodeConfig:
        preemptible: {{ properties['preemptible'] }}
        machineType: {{ properties['machineType'] }}
        imageType: CONTAINER_VM
        diskSizeGb: {{ properties['diskSize'] }}
        oauthScopes:
          - https://www.googleapis.com/auth/compute
          - https://www.googleapis.com/auth/devstorage.read_only
          - https://www.googleapis.com/auth/logging.write
          - https://www.googleapis.com/auth/monitoring

outputs:
- name: clusterType
  value: {{ TYPE_NAME }}
`

const configYAML = `
resources:
- type: registry.jinja
  name: {{ .Bucket }}
  properties:
    projectNumber: {{ .ProjectNumber }}
    zone: {{ .Zone }}

{{ if .ClusterNotExists }}
- type: container-vm.jinja
  name: {{ .ClusterName }}
  properties:
    zone: {{ .Zone }}
    numNodes: {{ .NumNodes }}
    diskSize: {{ .DiskSize }}
    machineType: {{ .MachineType }}
    preemptible: {{ .Preemptible }}
{{ end }}
`

const mysqlInstanceYAML = `
- type: sqladmin.v1beta4.instance
  name: '{{ .name }}'
  properties:
    region: '{{ .region }}'
    databaseVersion: '{{ .db_version }}'
    instanceType: CLOUD_SQL_INSTANCE
    backendType: SECOND_GEN
    settings:
      tier: '{{ .tier }}'
      backupConfiguration:
        enabled: false
        binaryLogEnabled: false
      ipConfiguration:
        ipv4Enabled: true
        requireSsl: true
      dataDiskSizeGb: 10
      dataDiskType: PD_SSD
      activationPolicy: '{{ .activation_policy }}'
      locationPreference:
        zone: {{ .zone }}
- type: sqladmin.v1beta4.database
  name: {{ .name }}-{{ .database }}
  properties:
    name: {{ .database }}
    instance: $(ref.{{ .name }}.name)
    charset: utf8mb4
    collation: utf8mb4_general_ci
`

var pgsqlInstanceYAML = `
- type: sqladmin.v1beta4.instance
  name: '{{ .name }}'
  properties:
    region: '{{ .region }}'
    databaseVersion: '{{ .db_version }}'
    settings:
      memory: '{{ .memory }}'
      cpu: '{{ .cpu }}'
      backupConfiguration:
        enabled: true
        binaryLogEnabled: true
      ipConfiguration:
        ipv4Enabled: false
        requireSsl: false
      dataDiskSizeGb: 10
      dataDiskType: PD_SSD
      activationPolicy: '{{ .activation_policy }}'
      locationPreference:
        zone: {{ .zone }}
- type: sqladmin.v1beta4.database
  name: {{ .name }}-{{ .database }}
  properties:
    name: {{ .database }}
    instance: $(ref.{{ .name }}.name)
    charset: utf8mb4
    collation: utf8mb4_general_ci
`
