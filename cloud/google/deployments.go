package google

import (
  "fmt"
  "time"
  "os"
  "os/signal"
  "path/filepath"
  "syscall"

  log "github.com/Sirupsen/logrus"
  "google.golang.org/api/googleapi"
  dm "google.golang.org/api/deploymentmanager/v2"
  "github.com/dinesh/datacol/client/models"
)

type initOptions struct {
  ClusterName   string
  MachineType   string
  DiskSize      int64
  NumNodes      int
  ProjectNumber int64
  Zone, Bucket  string
  ClusterNotExists, Preemptible bool
}

func (g *GCPCloud) Initialize(cluster, machineType string, nodes int, preemt bool) error {
  name := g.DeploymentName
  log.Infof("creating new stack %s", name)
  cnexists := true

  if len(cluster) == 0 {
    cluster = fmt.Sprintf("%v-cluster", g.DeploymentName)
  } else {
    cnexists = false
  }

  resp, err := g.csmanager().Projects.Get(g.Project).Do()
  if err != nil { return err }

  if len(machineType) == 0 {
    machineType = ditermineMachineType(nodes)
  }

  options := initOptions {
    ClusterName:  cluster,
    DiskSize:     10,
    NumNodes:     nodes,
    MachineType:  machineType,
    Zone:         g.Zone,
    Bucket:       g.BucketName,
    ProjectNumber:    resp.ProjectNumber,
    ClusterNotExists: cnexists,
    Preemptible: preemt,
  }

  log.Debug(toJson(options))

  imports := []*dm.ImportFile {
    {
      Name: "registry.jinja",
      Content: registryYAML,
    },
  }

  if cnexists {
    imports = append(imports, &dm.ImportFile{
       Name: "container-vm.jinja",
       Content: vmYAML,
    })
  }

  req := &dm.Deployment {
    Name: name,
    Target: &dm.TargetConfiguration {
      Imports: imports,
      Config: &dm.ConfigFile {
        Content: compileConfig(configYAML, &options),
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
  if err != nil { return err }

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

  return nil
}

func (g *GCPCloud) waitForDpOp(svc *dm.Service, op *dm.Operation, interrupt bool) error {
  log.Infof("Waiting on %s [%v]", op.OperationType, op.Name)
  project := g.Project

  cancelCh := make(chan os.Signal, 1)
  signal.Notify(cancelCh, os.Interrupt, syscall.SIGTERM)

  for {
    time.Sleep(2 * time.Second)
    op, err := svc.Operations.Get(project, op.Name).Do()
    if err != nil { return err }

    select {
      case <- cancelCh:
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

        if err := g.Teardown(); err != nil {
          log.Debugf("deleting stack: %+v", err)
        }
        return last
      }
      return nil
    default:
      return fmt.Errorf("Unknown status %q: %+v", op.Status, op)
    }
  }
}

var registryYAML = `
resources:
- name: {{ env['name'] }}
  type: storage.v1.bucket
  properties:
    location: {{ properties['bucketLocation'] }}
    projectNumber: {{ properties['projectNumber'] }}
`

var vmYAML = `
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

var configYAML = `
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