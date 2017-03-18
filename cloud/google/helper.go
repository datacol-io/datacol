package google 

import (
  "fmt"
  "log"
  "runtime"
  "path"
  "io/ioutil"
  "html/template"
  "encoding/json"
  "bytes"
  "bufio"
  "strings"

  "github.com/dinesh/rz/client/models"
)

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
