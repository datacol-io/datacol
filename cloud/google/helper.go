package google 

import (
  "fmt"
  "path"
  "bytes"
  "bufio"
  "strings"
  "runtime"
  "io/ioutil"
  "html/template"
  "encoding/json"
  "path/filepath"
  "encoding/hex"
  "crypto/rand"
  "crypto/sha256"

  log "github.com/Sirupsen/logrus"
  "github.com/dinesh/datacol/client/models"
)

func kubecfgPath(name string) string {
  return filepath.Join(models.ConfigPath, name, "kubeconfig")
}

func getTokenFile(name string) string {
  return filepath.Join(models.ConfigPath, name, models.SvaFilename)
}

func getCachedToken(name string) string {
  value, err := ioutil.ReadFile(filepath.Join(models.ConfigPath, name, "token"))
  if err != nil {
    return ""
  }
  return strings.TrimSpace(string(value))
}

func loadTmpl(name string) string {
  _, filename, _, _ := runtime.Caller(1)
  dir := path.Join(path.Dir(filename), "templates")

  content, err := ioutil.ReadFile(dir + "/" + name)
  if err != nil { log.Fatal(err) }

  return string(content)
}

func compileTmpl(content string, opts interface{}) string {
  tmpl, err := template.New("ct").Parse(content)
  if err != nil { log.Fatal(err) }

  var doc bytes.Buffer
  if err := tmpl.Execute(&doc, opts); err != nil {
    log.Fatal(err)
  }

  return doc.String()
}

func ditermineMachineType(num int) string {
  return "n1-standard-1"
}

func toJson(object interface {}) string {
  dump, err := json.MarshalIndent(object, " ", "  ")
  if err != nil { 
    log.Fatal(fmt.Errorf("dumping json: %v", err)) 
  }
  return string(dump)
}


func loadEnv(data []byte) models.Environment {
  e := models.Environment{}

  scanner := bufio.NewScanner(bytes.NewReader(data))
  for scanner.Scan() {
    parts := strings.SplitN(scanner.Text(), "=", 2)

    if len(parts) == 2 {
      if key := strings.TrimSpace(parts[0]); key != "" {
        e[key] = parts[1]
      }
    }
  }

  return e
}

func getGcpRegion(zone string) string {
  return zone[0:len(zone)-2]
}

func generatePassword() (string, error) {
  data := make([]byte, 1024)

  if _, err := rand.Read(data); err != nil {
    return "", err
  }

  hash := sha256.Sum256(data)
  return hex.EncodeToString(hash[:])[0:30], nil
}

func dpToResourceType(dpname, name string) string {
  kind := "unknown"

  switch dpname {
  case "storage.v1.bucket":
    kind = "gs"
  case "container.v1.cluster":
    kind = "cluster"
  case "sqladmin.v1beta4.instance":
    kind = strings.Split(name, "-")[0]
  }
  
  return kind
}