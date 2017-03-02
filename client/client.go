package client

import (
  "os"
  "fmt"
  "time"
  "context"
  "bytes"
  "log"
  "io/ioutil"
  "text/template"
  "net/http"

  "github.com/pkg/errors"
  "golang.org/x/oauth2/google"
  dm "google.golang.org/api/deploymentmanager/v2"
  csm "google.golang.org/api/cloudresourcemanager/v1"
)

type ConfigOption struct {
  RackName, Zone, ProjectId, BucketLocation, Bucket string
  ProjectNumber, DiskSize int64
  NumNodes int
  MachineType string
}

func getProject(client *http.Client, projectId string) int64 {
  service, err := csm.New(client)
  if err != nil { 
    checkErr(errors.Wrap(err, "failed to get cloudresourcemanager client"))
  }

  resp, err := service.Projects.Get(projectId).Do()

  if err != nil { 
    checkErr(errors.Wrap(err, fmt.Sprintf("failed to fetch project %s", projectId)))
  }

  return resp.ProjectNumber
}

func DeleteStack(auth *Auth) error {
  hc := GetClientOrDie(auth)

  service, err := dm.New(hc)
  if err != nil {
    return errors.Wrap(err, "failed to get deploymentmanager client")
  }

  op, err := service.Deployments.Delete(auth.ProjectId, auth.RackName).Do()
  if err != nil {
    return errors.Wrap(err, fmt.Sprintf("failed to get deployment %s", auth.RackName))
  }

  if err = awaitOp(auth.ProjectId, service, op); err != nil {
    return errors.Wrap(err, "failed to delete stack")
  }

  return nil
}

func GetClientOrDie(auth *Auth) *http.Client {
  jwtConfig, err := google.JWTConfigFromJSON(auth.ServiceKey, csm.CloudPlatformScope)
  checkErr(errors.Wrap(err, "failed to get jwt client"))

  return jwtConfig.Client(context.TODO())
}

func CreateStack(auth *Auth, numNodes int) error {
  name := auth.RackName
  projectId := auth.ProjectId
  hc := GetClientOrDie(auth)

  service, err := dm.New(hc)
  checkErr(err)

  co := &ConfigOption {
    RackName: name,
    Zone: auth.Zone,
    ProjectId: projectId,
    ProjectNumber: getProject(hc, projectId),
    MachineType: "f1-micro",
    BucketLocation: "asia-east1",
    Bucket: auth.BucketName,
    DiskSize: 10,
    NumNodes: numNodes,
  }

  log.Printf("creating stack: %s with %+v", name, *co)

  deployment := &dm.Deployment {
    Name: name,
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
        Content: compileConfig("config.j2", co),
      },
    },
  }

  op, err := service.Deployments.Insert(projectId, deployment).Do()
  checkErr(err)

  return awaitOp(projectId, service, op)
}

func loadTemplate(name string) string {
  content, err := ioutil.ReadFile("client/templates/" + name)
  checkErr(err)
  return string(content)
}

func compileConfig(name string, co *ConfigOption) string {
  tmpl, err := template.New("ct").Parse(loadTemplate(name))
  if err != nil { panic(err) }

  var doc bytes.Buffer
  if err := tmpl.Execute(&doc, co); err != nil { panic(err) }
  return doc.String()
}

func checkErr(err error) {
  if err != nil { 
    fmt.Printf("%+v\n", err)
    os.Exit(1)
  }
}

func awaitOp(projectId string, svc *dm.Service, op *dm.Operation) error {
  opName := op.Name
  log.Printf("Waiting on %s [%v]\n", op.Kind, opName)
  for {
    time.Sleep(2 * time.Second)
    op, err := svc.Operations.Get(projectId, opName).Do()
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
      log.Printf("Done.")
      return nil
    default:
      return fmt.Errorf("Unknown status %q: %+v", op.Status, op)
    }
  }
}