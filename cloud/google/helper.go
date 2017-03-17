package google 

import (
  "context"
  "net/http"
  "fmt"
  "log"
  "runtime"
  "path"
  "io/ioutil"
  "html/template"
  "encoding/json"
  "bytes"

  csm "google.golang.org/api/cloudresourcemanager/v1"
  oauth2_google "golang.org/x/oauth2/google"
)

func getProjectNumber(client *http.Client, id string) (int64, error) {
  service, err := csm.New(client)
  if err != nil { 
    return 0, fmt.Errorf("cloudresourcemanager client %s", err)
  }

  op, err := service.Projects.Get(id).Do()

  if err != nil {
    return 0, fmt.Errorf("fetching %s: %s",id, err)
  }

  return op.ProjectNumber, nil
}

func JwtClient(sva []byte) *http.Client {
  jwtConfig, err := oauth2_google.JWTConfigFromJSON(sva, csm.CloudPlatformScope)
  if err != nil {
    log.Fatal(fmt.Errorf("JWT client %s", err))
  }

  return jwtConfig.Client(context.TODO())
}

func BearerToken(sva []byte) (string, error) {
  jwtConfig, err := oauth2_google.JWTConfigFromJSON(sva, csm.CloudPlatformScope)
  if err != nil { return "", err }

  source := jwtConfig.TokenSource(context.TODO())
  tk, err := source.Token()
  if err != nil { return "", err }
  
  return tk.AccessToken, nil
}

func loadTemplate(name string) string {
  _, filename, _, _ := runtime.Caller(1)
  dir := path.Join(path.Dir(filename), "templates")
  content, err := ioutil.ReadFile(dir + "/" + name)

  if err != nil { log.Fatal(err) }

  return string(content)
}

func compileConfig(name string, opts *initOptions) string {
  tmpl, err := template.New("ct").Parse(loadTemplate(name))
  if err != nil { log.Fatal(err) }

  var doc bytes.Buffer
  if err := tmpl.Execute(&doc, opts); err != nil {
    log.Fatal(err) 
  }

  return doc.String()
}

func ditermineMachineType(num int) string {
  return "f1-micro"
}

func dumpJson(object interface {}) {
  dump, err := json.MarshalIndent(object, " ", "  ")
  if err != nil { 
    log.Fatal(fmt.Errorf("dumping json: %v", err)) 
  }
  fmt.Println(string(dump))
}
