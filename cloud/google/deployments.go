package google

import (
  "fmt"
  "time"

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

func (g *GCPCloud) Initialize(cluster string, nodes int) error {
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

  name := g.DeploymentName
  fmt.Printf("creating new stack %s with: \n", name)
  dumpJson(options)

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
      return fmt.Errorf("%s is already created.", name)
    } else {
      return err 
    }
  }
  
  if err = waitForDpOp(service, op, g.Project); err != nil {
    return err
  }

  kube, err := g.getCluster(cluster)
  return GenerateClusterConfig(name, cluster, kube)
}

func (g *GCPCloud) Teardown() error {
  name := g.DeploymentName
  fmt.Printf("Deleting stack %s\n", name)

  gsService := g.storage()
  resp, err := gsService.Objects.List(g.BucketName).Do()
  if err != nil { return err }

  for _, obj := range resp.Items {
    if err := gsService.Objects.Delete(obj.Bucket, obj.Name).Do(); err != nil {
      return err
    }
  }

  dmService := g.deploymentmanager()
  op, err := dmService.Deployments.Delete(g.Project, name).Do()
  if err != nil { 
    return err 
  } 

  if err = waitForDpOp(dmService, op, g.Project); err != nil {
    return fmt.Errorf("deleting stack %v", err)
  }

  fmt.Printf("OK\n")
  return nil
}

func waitForDpOp(svc *dm.Service, op *dm.Operation, project string) error {
  fmt.Printf("Waiting on %s [%v]\n", op.Kind, op.Name)

  for {
    time.Sleep(2 * time.Second)
    op, err := svc.Operations.Get(project, op.Name).Do()
    if err != nil { return err }

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
