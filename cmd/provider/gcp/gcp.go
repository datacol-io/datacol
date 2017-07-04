package gcp

import (
	"cloud.google.com/go/datastore"
	"context"
	"fmt"
	log "github.com/Sirupsen/logrus"
	term "github.com/appscode/go-term"
	"golang.org/x/oauth2/google"
	csm "google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/compute/v1"
	dm "google.golang.org/api/deploymentmanager/v2"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	smm "google.golang.org/api/servicemanagement/v1"
	"google.golang.org/api/storage/v1"
	"google.golang.org/grpc"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	appKind      = "App"
	buildKind    = "Build"
	releaseKind  = "Release"
	resourceKind = "Resource"
)

type InitOptions struct {
	Name, Project, ClusterName, MachineType, Zone, Bucket string
	ApiKey, Version, Region, ArtifactBucket               string
	SAEmail, ClusterVersion                               string
	ProjectNumber                                         int64
	DiskSize, NumNodes, ControllerPort                    int
	ClusterNotExists, Preemptible                         bool
}

type initResponse struct {
	Host, Password string
}

func InitializeStack(opts *InitOptions) (*initResponse, error) {
	if len(opts.MachineType) == 0 {
		opts.MachineType = ditermineMachineType(opts.NumNodes)
	}

	opts.Region = getGcpRegion(opts.Zone)
	opts.ControllerPort = 8080

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

	instanceName := fmt.Sprintf("%s-datacol-bastion", opts.Name)
	ret, err := compute.NewInstancesService(computeService(opts.Name)).Get(
		opts.Project,
		opts.Zone,
		instanceName,
	).Do()

	if err != nil {
		return nil, fmt.Errorf("failed to get IP err: %v", err)
	}

	host := externalIp(ret)
	return &initResponse{Host: host, Password: opts.ApiKey}, nil
}

func servicemanagement(name string) *smm.APIService {
	svc, err := smm.New(httpClient(name))
	if err != nil {
		log.Fatal(fmt.Errorf("creating servicemanagement client err: %v", err))
	}
	return svc
}

func httpClient(name string) *http.Client {
	cfg, err := google.JWTConfigFromJSON(serviceKey(serviceKeyPath(name)), csm.CloudPlatformScope)

	if err != nil {
		log.Fatal(fmt.Errorf("creating JWT config err: %v", err))
	}

	return cfg.Client(context.TODO())
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
	term.Printf("Deleting stack %s\n", name)

	gsService := storageService(name)
	resp, err := gsService.Objects.List(bucket).Do()
	if err != nil {
		return fmt.Errorf("listing items inside bucket err: %v", err)
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

	return resetDatabase(name, project)
}

func resetDatabase(name, project string) error {
	s, close := datastoreClient(name, project)
	defer close()

	ctx := datastore.WithNamespace(context.TODO(), name)

	if err := deleteFromQuery(s, ctx, datastore.NewQuery(appKind)); err != nil {
		return fmt.Errorf("deleting apps err: %v", err)
	}

	if err := deleteFromQuery(s, ctx, datastore.NewQuery(buildKind)); err != nil {
		return fmt.Errorf("deleting builds err: %v", err)
	}

	if err := deleteFromQuery(s, ctx, datastore.NewQuery(releaseKind)); err != nil {
		return fmt.Errorf("deleting releases err: %v", err)
	}

	if err := deleteFromQuery(s, ctx, datastore.NewQuery(resourceKind)); err != nil {
		return fmt.Errorf("deleting resources err: %v", err)
	}

	return nil
}

func waitForDpOp(svc *dm.Service, op *dm.Operation, project string, interrupt bool, teardown func() error) error {
	term.Printf("Waiting on %s [%v]", op.OperationType, op.Name)

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

func datastoreClient(name, project string) (*datastore.Client, func()) {
	opts := []option.ClientOption{
		option.WithGRPCDialOption(grpc.WithBackoffMaxDelay(5 * time.Second)),
		option.WithGRPCDialOption(grpc.WithTimeout(30 * time.Second)),
	}

	opts = append(opts, option.WithServiceAccountFile(serviceKeyPath(name)))

	client, err := datastore.NewClient(context.TODO(), project, opts...)
	if err != nil {
		log.Fatal(fmt.Errorf("datastore client %s", err))
	}

	return client, client.Close
}

func storageService(name string) *storage.Service {
	svc, err := storage.New(httpClient(name))
	if err != nil {
		log.Fatal(fmt.Errorf("storage client %s", err))
	}

	return svc
}

func dmService(name string) *dm.Service {
	svc, err := dm.New(httpClient(name))
	if err != nil {
		log.Fatal(fmt.Errorf("deploymentmanager client %s", err))
	}

	return svc
}

func computeService(name string) *compute.Service {
	svc, err := compute.New(httpClient(name))
	if err != nil {
		log.Fatal(fmt.Errorf("compute client %s", err))
	}

	return svc
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
      initialClusterVersion: {{ properties['version'] }}
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
    api_key: {{ .ApiKey }}
    stack_name: {{ .Name }}
    version: {{ .Version }}
    region: {{ .Region }}
    port: {{ .ControllerPort }}
    cluster_name: {{ .ClusterName }}
    artifact_bucket: {{ .ArtifactBucket }}

{{ if .ClusterNotExists }}
- type: container-vm.jinja
  name: {{ .ClusterName }}
  properties:
    zone: {{ .Zone }}
    numNodes: {{ .NumNodes }}
    diskSize: {{ .DiskSize }}
    machineType: {{ .MachineType }}
    preemptible: {{ .Preemptible }}
    version: {{ .ClusterVersion }}
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

- name: {{ env['deployment'] }}-datacol-bastion
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
      - key: DATACOL_CLUSTER
        value: {{ properties['cluster_name'] }}
      - key: startup-script
        value: |
          #! /bin/bash
          apt-get update
          apt-get install -y unzip curl
          curl -LO https://storage.googleapis.com/kubernetes-release/release/v1.6.3/bin/linux/amd64/kubectl > kubectl &&
            chmod +x kubectl && \
            mv kubectl /usr/local/bin
          mkdir -p /opt/datacol && \
          curl -Ls /tmp https://storage.googleapis.com/{{ properties['artifact_bucket'] }}/binaries/{{ properties['version'] }}/apictl.zip > /tmp/apictl.zip
            unzip /tmp/apictl.zip -d /opt/datacol && chmod +x /opt/datacol/apictl
          cd /opt/datacol && nohup ./apictl -log-file log.txt &
`
