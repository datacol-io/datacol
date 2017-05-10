package google

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"cloud.google.com/go/datastore"
	log "github.com/Sirupsen/logrus"
	pb "github.com/dinesh/datacol/api/models"
	"google.golang.org/api/compute/v1"
	dm "google.golang.org/api/deploymentmanager/v2"
	"google.golang.org/api/googleapi"
	sql "google.golang.org/api/sqladmin/v1beta4"
)

const stackKind = "Stack"

type InitOptions struct {
	Name, Project, ClusterName, MachineType, Zone, Bucket string
	DiskSize, ProjectNumber                               int64
	NumNodes, Port                                        int
	SAEmail                                               string
	ClusterNotExists, Preemptible                         bool
	API_KEY, Version, Region                              string
}

func teardown(dmService *dm.Service, name, project string) error {
	log.Infof("Deleting stack %s", name)

	op, err := dmService.Deployments.Delete(project, name).Do()
	if err != nil {
		return err
	}

	if err = waitForDpOp(dmService, op, project, false, nil); err != nil {
		return fmt.Errorf("deleting stack %v", err)
	}

	return nil
}

func TeardownStack(name, project, bucket string) error {
	log.Infof("Deleting stack %s", name)

	gsService := storageService(name)
	resp, err := gsService.Objects.List(bucket).Do()
	if err != nil {
		return err
	}

	for _, obj := range resp.Items {
		if err := gsService.Objects.Delete(obj.Bucket, obj.Name).Do(); err != nil {
			return err
		}
	}

	dmsvc := dmService(name)
	op, err := dmsvc.Deployments.Delete(project, name).Do()
	if err != nil {
		return err
	}

	if err = waitForDpOp(dmsvc, op, project, false, nil); err != nil {
		return fmt.Errorf("deleting stack %v", err)
	}

	return nil
	// return g.resetDatabase()
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

	if err = waitForDpOp(service, op, g.Project, false, nil); err != nil {
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
	return nameKey(stackKind, g.DeploymentName, nil)
}

func (g *GCPCloud) nestedKey(kind, key string) *datastore.Key {
	return nameKey(kind, key, g.stackKey())
}

func waitForDpOp(svc *dm.Service, op *dm.Operation, project string, interrupt bool, teardown func() error) error {
	log.Infof("Waiting on %s [%v]", op.OperationType, op.Name)

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
				return teardown()
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
					if err := teardown(); err != nil {
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

func resourceFromStack(service *dm.Service, project, stack, name string) (*pb.Resource, error) {
	return &pb.Resource{Name: name}, nil
}

type initResponse struct {
	Host, Password string
}

func InitializeStack(opts *InitOptions) (*initResponse, error) {
	if len(opts.MachineType) == 0 {
		opts.MachineType = ditermineMachineType(opts.NumNodes)
	}

	if len(opts.API_KEY) == 0 {
		opts.API_KEY = generateId("A", 19)
	}

	opts.Region = getGcpRegion(opts.Zone)
	opts.Port = 8080

	imports := []*dm.ImportFile{
		{
			Name:    "registry.jinja",
			Content: registryYAML,
		},
		{
			Name:    "compute.jinja",
			Content: apiVmYAML,
		},
	}

	if opts.ClusterNotExists {
		imports = append(imports, &dm.ImportFile{
			Name:    "container-vm.jinja",
			Content: kubeClusterYAML,
		})
	}

	name := opts.Name

	req := &dm.Deployment{
		Name: name,
		Target: &dm.TargetConfiguration{
			Imports: imports,
			Config: &dm.ConfigFile{
				Content: compileTmpl(configYAML, &opts),
			},
		},
	}

	service := dmService(name)
	log.Infof("creating new stack %s", name)
	log.Debugf(toJson(opts))

	op, err := service.Deployments.Insert(opts.Project, req).Do()
	if err != nil {
		if gerr, ok := err.(*googleapi.Error); ok && gerr.Code == 409 {
			return nil, fmt.Errorf("Failed: Deployment %s is already created. Please destroy.", name)
		} else {
			return nil, err
		}
	}

	tearfn := func() error {
		return teardown(service, opts.Name, opts.Project)
	}

	if err = waitForDpOp(service, op, opts.Project, true, tearfn); err != nil {
		return nil, err
	}

	instanceName := fmt.Sprintf("%s-compute", opts.Name)
	ret, err := compute.NewInstancesService(computeService(opts.Name)).Get(
		opts.Project,
		opts.Zone,
		instanceName,
	).Do()

	if err != nil {
		return nil, fmt.Errorf("failed to get IP err: %v", err)
	}

	host := externalIp(ret)
	return &initResponse{Host: host, Password: opts.API_KEY}, nil
}

const registryYAML = `
resources:
- name: {{ env['name'] }}
  type: storage.v1.bucket
  properties:
    location: {{ properties['bucketLocation'] }}
    projectNumber: {{ properties['projectNumber'] }}
`

const kubeClusterYAML = `
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

- type: compute.jinja
  name: {{ .Name }}-datacol-vm
  properties:
    machineType: f1-micro
    zone: {{ .Zone }}
    bucket: {{ .Bucket }}
    sva_email: {{ .SAEmail }}
    api_key: {{ .API_KEY }}
    stack_name: {{ .Name }}
    version: {{ .Version }}
    region: {{ .Region }}
    port: {{ .Port }}

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

const apiVmYAML = `
resources:
- name: {{ env["deployment"] }}-fw
  type: compute.v1.firewall
  properties:
    region: {{ properties['region'] }}
    allowed:
    - IPProtocol: tcp
      ports:
      - 80
      - 22
      - 8080
      - 10000
    sourceRange: 0.0.0.0/0
    targetTags:
    - datacol

- name: {{ env['deployment'] }}-compute
  type: compute.v1.instance
  properties:
    machineType: https://www.googleapis.com/compute/v1/projects/{{ env["project"] }}/zones/{{ properties["zone"] }}/machineTypes/{{ properties["machineType"] }}
    zone: {{ properties['zone'] }}
    disks:
    - autoDelete: true
      boot: true
      initializeParams:
        sourceImage: https://www.googleapis.com/compute/v1/projects/debian-cloud/global/images/debian-8-jessie-v20170426
    networkInterfaces:
    - accessConfigs:
      - name: external-nat
        type: ONE_TO_ONE_NAT
    serviceAccounts:
    - email: {{ properties['sva_email'] }}
      scopes:
        - https://www.googleapis.com/auth/cloud-platform
        - https://www.googleapis.com/auth/devstorage.read_write
        - https://www.googleapis.com/auth/logging.write
        - https://www.googleapis.com/auth/monitoring
    tags:
      items:
        - datacol
    metadata:
      items:
      - key: DATACOL_API_KEY
        value: {{ properties['api_key'] }}
      - key: DATACOL_STACK
        value: {{ properties['stack_name'] }}
      - key: DATACOL_BUCKET
        value: {{ properties['bucket'] }}
      - key: startup-script
        value: |
          #! /bin/bash
          apt-get update
          apt-get install -y unzip curl
          curl -Ls /tmp https://storage.googleapis.com/datacol-distros/binaries/{{ properties['version'] }}/apictl.zip > /tmp/apictl.zip
          unzip /tmp/apictl.zip -d /usr/local/bin && chmod +x /usr/local/bin/apictl
          nohup apictl &
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
