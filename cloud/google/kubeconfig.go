package google

import (
  "bytes"
  "os"
  "fmt"
  "html/template"
  "path/filepath"
  "io/ioutil"
  "encoding/base64"

   container "google.golang.org/api/container/v1"
   homeDir "github.com/mitchellh/go-homedir"
)

var (
  home, _   = homeDir.Dir()
)

const  kubeconfigTemplate = `apiVersion: v1
clusters:
- cluster:
    certificate-authority: {{.CA}}
    server: {{.Server}}
  name: {{.Cluster}}
contexts:
- context:
    cluster: {{.Cluster}}
    user: {{.User}}
  name: {{.Context}}
current-context: {{.Context}}
kind: Config
preferences: {}
users:
- name: {{.User}}
  tokenFile: {{.TokenFile}}
  user:
    auth-provider:
      name: gcp
`

type configOptions struct {
  CA      string
  Server  string
  Cluster string
  User    string
  Context string
  TokenFile string
}

func GenerateClusterConfig(rackName, baseDir string, c *container.Cluster) error {
  tmpl, err := template.New("kubeconfig").Parse(kubeconfigTemplate)
  if err != nil {
    return fmt.Errorf("error reading config template: %v", err)
  }

  if err := os.MkdirAll(baseDir, 0700); err != nil {
    return err
  }

  kubeconfigFile := filepath.Join(baseDir, "kubeconfig")
  certsDir := baseDir

  // Base64 encoded ca
  caPath := filepath.Join(certsDir, "ca.pem")
  if err := ioutil.WriteFile(caPath, []byte(c.MasterAuth.ClusterCaCertificate), 0700); err != nil {
    return err
  }

  caDecodedPath := filepath.Join(certsDir, "ca-decoded.pem")
  caDecoded, err := base64.StdEncoding.DecodeString(c.MasterAuth.ClusterCaCertificate)
  if err != nil {
    return fmt.Errorf("error decoding ca file for kubeconfig: %v", err)
  }

  if err := ioutil.WriteFile(caDecodedPath, caDecoded, 0700); err != nil {
    return err
  }

  copts := &configOptions {
    CA:       caDecodedPath,
    Server:   "https://" + c.Endpoint,
    User:     rackName,
    Context:  rackName,
    Cluster:  rackName,
    TokenFile:  getTokenFile(rackName),
  }

  var kubeconfig bytes.Buffer
  if err = tmpl.Execute(&kubeconfig, copts); err != nil { 
    return err 
  }

  if err = ioutil.WriteFile(kubeconfigFile, kubeconfig.Bytes(), 0644); err != nil {
    return fmt.Errorf("error writing kubeconfig file: %v", err)
  }

  return nil
}
