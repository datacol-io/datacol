package google

import (
  "fmt"
  "time"
  "os"
  "os/signal"
  "syscall"

  log "github.com/Sirupsen/logrus"
  "google.golang.org/api/googleapi"
   dm "google.golang.org/api/deploymentmanager/v2"
)

type initOptions struct {
  ClusterName   string
  MachineType   string
  DiskSize      int64
  NumNodes      int
  ClusterNotExists bool
  ProjectNumber int64
  Zone, Bucket  string
}

func (g *GCPCloud) Initialize(cluster string, nodes int, cfgroot string) error {
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

  options := initOptions {
    ClusterName:  cluster,
    DiskSize:     10,
    NumNodes:     nodes,
    ClusterNotExists: cnexists,
    MachineType: ditermineMachineType(nodes),
    Zone: g.Zone,
    Bucket: g.BucketName,
    ProjectNumber: resp.ProjectNumber,
  }

  log.Debug(toJson(options))

  imports := []*dm.ImportFile {
    {
      Name: "registry.jinja",
      Content: loadTemplate("registry.j2"),
    },
  }

  if cnexists {
    imports = append(imports, &dm.ImportFile{
       Name: "container-vm.jinja",
       Content: loadTemplate("container-vm.j2"),
    })
  }

  req := &dm.Deployment {
    Name: name,
    Target: &dm.TargetConfiguration {
      Imports: imports,
      Config: &dm.ConfigFile {
        Content: compileConfig("config.j2", &options),
      },
    },
  }

  service := g.deploymentmanager()

  op, err := service.Deployments.Insert(g.Project, req).Do()
  if err != nil { 
    if gerr, ok := err.(*googleapi.Error); ok && gerr.Code == 409 {
      return fmt.Errorf("Failed: Deployment %s is already created.", name)
    } else {
      return err 
    }
  }
  
  if err = g.waitForDpOp(service, op, true); err != nil {
    return err
  }

  kube, err := g.getCluster(cluster)
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

  log.Infoln("OK")
  return nil
}

func (g *GCPCloud) waitForDpOp(svc *dm.Service, op *dm.Operation, interrupt bool) error {
  log.Infof("Waiting on %s [%v]", op.Kind, op.Name)
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
          last = fmt.Errorf("%v", operr)
        }
        return last
      }
      fmt.Printf("\nDone.")
      return nil
    default:
      return fmt.Errorf("Unknown status %q: %+v", op.Status, op)
    }
  }
}
