package google

import (
	"bytes"
	"encoding/base64"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"

	"cloud.google.com/go/compute/metadata"
	container "google.golang.org/api/container/v1"
)

const kubeconfigTemplate = `apiVersion: v1
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
  user:
    auth-provider:
      name: gcp
`

type configOptions struct {
	CA        string
	Server    string
	Cluster   string
	User      string
	Context   string
	TokenFile string
}

func (g *GCPCloud) K8sConfigPath() (string, error) {
	c, err := metadata.InstanceAttributeValue("DATACOL_CLUSTER")
	if err != nil {
		return "", err
	}

	return cacheKubeConfig(g.DeploymentName, g.Project, g.Zone, c)
}

func cacheKubeConfig(sName, project, zone, cname string) (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	filename := filepath.Join(dir, "kubeconfig")

	if _, err = os.Stat(filename); err != nil {
		if os.IsNotExist(err) {
			log.Debugf("creating kubeconfig in %s ...", dir)
			svc, err := container.New(httpClient(sName))
			if err != nil {
				return filename, fmt.Errorf("container client %s", err)
			}

			ret, err := svc.Projects.Zones.Clusters.Get(project, zone, cname).Do()
			if err != nil {
				return filename, err
			}

			return filename, generateClusterConfig(sName, dir, ret)
		} else {
			return filename, err
		}
	}

	return filename, nil
}

func generateClusterConfig(rackName, baseDir string, c *container.Cluster) error {
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
	caDecodedPath := filepath.Join(certsDir, "ca.pem")
	caDecoded, err := base64.StdEncoding.DecodeString(c.MasterAuth.ClusterCaCertificate)
	if err != nil {
		return fmt.Errorf("error decoding ca file for kubeconfig: %v", err)
	}

	if err := ioutil.WriteFile(caDecodedPath, caDecoded, 0700); err != nil {
		return err
	}

	copts := &configOptions{
		CA:      caDecodedPath,
		Server:  "https://" + c.Endpoint,
		User:    rackName,
		Context: rackName,
		Cluster: rackName,
		// TokenFile: getTokenFile(rackName),
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
