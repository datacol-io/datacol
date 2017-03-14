package google

import (
  "fmt"
  "net/http"
  "time"

  "google.golang.org/api/googleapi"
   dm "google.golang.org/api/deploymentmanager/v2"
   container "google.golang.org/api/container/v1"
)

type Deployment struct {
  Name, ClusterName string
  Zone, ProjectId, BucketLocation, Bucket, MachineType string
  ProjectNumber, DiskSize int64
  NumNodes int
  htc *http.Client
}

func NewDeployment(cred []byte,n,prid,zn,b string,nodes int) (*Deployment, error) {
  htc := JwtClient(cred)
  number, err := getProjectNumber(htc, prid)
  if err != nil { return nil, err }

  return &Deployment {
    Name:         n,
    ClusterName:  fmt.Sprintf("%v-cluster", n),
    ProjectId:    prid,
    Zone:         zn,
    DiskSize:     10,
    NumNodes:     nodes,
    Bucket:       b,
    ProjectNumber: number,
    htc:          htc,
    MachineType:  ditermineMachineType(nodes),
  }, nil
}

func NewDeploymentByCred(cred []byte) *Deployment {
  return &Deployment{
    htc: JwtClient(cred),
  }
}

func (dp *Deployment) Run(force bool) error {
  fmt.Printf("creating new stack: %+v\n", dp)

  req := &dm.Deployment {
    Name: dp.Name,
    Target: &dm.TargetConfiguration {
      Imports: [] *dm.ImportFile {
        {
          Name: "container-vm.jinja",
          Content: loadTemplate("container-vm.j2"),
        },
        {
          Name: "registry.jinja",
          Content: loadTemplate("registry.j2"),
        },
      },
      Config: &dm.ConfigFile {
        Content: compileConfig("config.j2", dp),
      },
    },
  }

  service, err := dm.New(dp.htc)
  if err != nil { return err }

  op, err := service.Deployments.Insert(dp.ProjectId, req).Do()
  if err != nil { 
    if gerr, ok := err.(*googleapi.Error); ok && gerr.Code == 409 {
      return fmt.Errorf("%s is already created.", dp.Name)
    } else {
      return err 
    }
  }
  
  if err = dp.waitForOp(service, op); err != nil {
    return err
  }
  return nil
}

func (dp *Deployment) Delete() error {
  service, err := dm.New(dp.htc)
  fmt.Printf("Deleting deployment %+v\n", dp)

  if err != nil {
    return fmt.Errorf("deploymentmanager client %v", err)
  }

  op, err := service.Deployments.Delete(dp.ProjectId, dp.Name).Do()
  if err != nil {
    return err
  }

  if err = dp.waitForOp(service, op); err != nil {
    return fmt.Errorf("deleting stack %v", err)
  }

  return nil
}

func (dp *Deployment) waitForOp(svc *dm.Service, op *dm.Operation) error {
  fmt.Printf("Waiting on %s [%v]\n", op.Kind, op.Name)

  for {
    time.Sleep(2 * time.Second)
    op, err := svc.Operations.Get(dp.ProjectId, op.Name).Do()
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
      fmt.Printf("Done.")
      return nil
    default:
      return fmt.Errorf("Unknown status %q: %+v", op.Status, op)
    }
  }
}

func (dp *Deployment) GetCluster() (*container.Cluster, error) {
  svc, err := container.New(dp.htc)
  if err != nil {
    return nil, fmt.Errorf("container service %s", err.Error()) 
  }

  return svc.Projects.Zones.Clusters.Get(
      dp.ProjectId,
      dp.Zone,
      dp.ClusterName,
  ).Do()

}