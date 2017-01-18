
package client

import (
  "fmt"
  "time"
  "context"
  "bytes"
  "log"
  "io/ioutil"
  "text/template"
  "net/http"

  "golang.org/x/oauth2/google"
  dm "google.golang.org/api/deploymentmanager/v2"
  csm "google.golang.org/api/cloudresourcemanager/v1"
)

type ConfigOption struct {
  Zone, ProjectId, BucketLocation string
  ProjectNumber int64
  MachineType string
}

func getProject(client *http.Client, projectId string) int64 {
  service, err := csm.New(client)
  if err != nil { panic(err) }

  resp, err := service.Projects.Get(projectId).Do()
  if err != nil { panic(err) }

  return resp.ProjectNumber
}

func CreateStack(name string, zone string, projectId string) {
  if name == "" { name = generateId("dm-", 5) }
  ctx := context.Background()
  hc, err := google.DefaultClient(ctx)
  checkErr(err)

  service, err := dm.New(hc)
  checkErr(err)

  co := &ConfigOption {
    Zone: zone,
    ProjectId: projectId,
    ProjectNumber: getProject(hc, projectId),
    MachineType: "f1-micro",
    BucketLocation: "asia-east1",
  }

  log.Printf("creating stack: %s with %+v", name, *co)

  deployment := &dm.Deployment {
    Name: name,
    Target: &dm.TargetConfiguration {
      Imports: [] *dm.ImportFile {
        {
          Name: "vm.jinja",
          Content: loadTemplate("vm.j2"),
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

  err = awaitOp(projectId, service, op)
  checkErr(err)
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
  if err != nil { panic(err) }
}

func awaitOp(projectId string, svc *dm.Service, op *dm.Operation) error {
  opName := op.Name
  log.Printf("Waiting on operation %v\n", opName)
  for {
    time.Sleep(2 * time.Second)
    op, err := svc.Operations.Get(projectId, opName).Do()
    if err != nil { return err }

    switch op.Status {
    case "PENDING", "RUNNING":
      log.Printf("Waiting on operation %v", opName)
      continue
    case "DONE":
      if op.Error != nil {
        var last error
        for _, operr := range op.Error.Errors {
          log.Printf("Error: %+v", operr)
          last = fmt.Errorf("%v", operr)
        }
        return last
      }
      log.Printf("Success. %+v", op)
      return nil
    default:
      return fmt.Errorf("Unknown status %q: %+v", op.Status, op)
    }
  }
}