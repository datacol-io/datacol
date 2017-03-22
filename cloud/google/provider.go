package google

import (
  "context"
  "net/http"
  "fmt"
  "io"
  "time"
  "math"
  "strings"
  "strconv"
  "io/ioutil"
  "path/filepath"

  log "github.com/Sirupsen/logrus"
  oauth2_google "golang.org/x/oauth2/google"
  "google.golang.org/api/storage/v1"
  "google.golang.org/api/cloudbuild/v1"
  "google.golang.org/api/deploymentmanager/v2"
  "google.golang.org/api/container/v1"
  csm "google.golang.org/api/cloudresourcemanager/v1"
  iam "google.golang.org/api/iam/v1"
  "k8s.io/client-go/tools/clientcmd"
  "k8s.io/client-go/kubernetes"
  kerrors "k8s.io/client-go/pkg/api/errors"
  kapi "k8s.io/client-go/pkg/api/v1"

  "github.com/dinesh/datacol/client/models"
)

var _jwtClient *http.Client

type GCPCloud struct {
  Project        string
  DeploymentName string
  BucketName     string
  Zone           string
  ServiceKey     []byte
}

func (g *GCPCloud) AppCreate(app *models.App) error {
 
  return nil
}

func (g *GCPCloud) AppGet(name string) (*models.App, error) {
  return nil, nil 
}

func (g *GCPCloud) AppDelete(name string) error {
  ns := g.DeploymentName
  kube, err := getKubeClientset(ns)
  if err != nil { return err }

  if _, err := kube.Core().Services(ns).Get(name); err != nil {
    if !kerrors.IsNotFound(err) {
      return err
    }
  } else if err := kube.Core().Services(ns).Delete(name, &kapi.DeleteOptions{}); err != nil {
    return err
  }

  if _, err = kube.Extensions().Deployments(ns).Get(name); err != nil {
    if !kerrors.IsNotFound(err) {
      return err
    }
  } else if err = kube.Extensions().Deployments(ns).Delete(name, &kapi.DeleteOptions{}); err != nil {
    return err
  }

  if _, err = kube.Extensions().Ingresses(ns).Get(name); err != nil {
    if !kerrors.IsNotFound(err) {
      return err
    }
  } else if err = kube.Extensions().Ingresses(ns).Delete(name, &kapi.DeleteOptions{}); err != nil {
    return err
  }

  return nil
}

func (g *GCPCloud) EnvironmentGet(name string) (models.Environment, error) {
  gskey := fmt.Sprintf("%s.env", name)  
  data, err := g.gsGet(g.BucketName, gskey)
  if err != nil {
    return nil, err
  }

  return loadEnv(data), nil
}

func (g *GCPCloud) EnvironmentSet(name string, body io.Reader) error {
  gskey := fmt.Sprintf("%s.env", name)
  return g.gsPut(g.BucketName, gskey, body)
}

func (g *GCPCloud) LogStream(cfgpath, podId string, out io.Writer, opts models.LogStreamOptions) error {
  c, err := getKubeClientset(g.DeploymentName)
  if err != nil { return err }

  req := c.Core().RESTClient().Get().
    Namespace(g.DeploymentName).
    Name(podId).
    Resource("pods").
    SubResource("log").
    Param("follow", strconv.FormatBool(opts.Follow))

  if opts.Since > 0 {
    sec := int64(math.Ceil(float64(opts.Since) / float64(time.Second)))
    req = req.Param("sinceSeconds", strconv.FormatInt(sec, 10))
  }

  rc, err := req.Stream()
  if err != nil { return err }

  defer rc.Close()
  _, err = io.Copy(out, rc)
  return err
}

func (g *GCPCloud) CacheCredentials() (string, error) {
  setting := "token"
  cfgpath := filepath.Join(models.ConfigPath, g.DeploymentName, setting)
  value, err := ioutil.ReadFile(cfgpath)

  if err != nil {
    jwtConfig, err := oauth2_google.JWTConfigFromJSON(g.ServiceKey, csm.CloudPlatformScope)
    if err != nil { return "", err }

    source := jwtConfig.TokenSource(context.TODO())
    tk, err := source.Token()
    if err != nil { 
      return "", err 
    }
    value = []byte(tk.AccessToken)
    err = ioutil.WriteFile(cfgpath, value, 0777)
    if err != nil { return "", err }
  }

  return strings.TrimSpace(string(value)), nil
}

func (g *GCPCloud) storage() *storage.Service {
  svc, err := storage.New(jwtClient(g.ServiceKey))
  if err != nil {
    log.Fatal(fmt.Errorf("storage client %s", err))
  }

  return svc
}

func (g *GCPCloud) cloudbuilder() *cloudbuild.Service {
  svc, err := cloudbuild.New(jwtClient(g.ServiceKey))
  if err != nil {
    log.Fatal(fmt.Errorf("cloudbuilder client %s", err))
  }

  return svc
}

func (g *GCPCloud) csmanager() *csm.Service {
  svc, err := csm.New(jwtClient(g.ServiceKey))
  if err != nil {
    log.Fatal(fmt.Errorf("cloudresourcemanager client %s", err))
  }

  return svc
}

func (g *GCPCloud) deploymentmanager() *deploymentmanager.Service {
  svc, err := deploymentmanager.New(jwtClient(g.ServiceKey))
  if err != nil {
    log.Fatal(fmt.Errorf("deploymentmanager client %s", err))
  }

  return svc
}

func (g *GCPCloud) container() *container.Service {
  svc, err := container.New(jwtClient(g.ServiceKey))
  if err != nil {
    log.Fatal(fmt.Errorf("container client %s", err))
  }

  return svc
}

func (g *GCPCloud) iam() *iam.Service {
  svc, err := iam.New(jwtClient(g.ServiceKey))
  if err != nil {
    log.Fatal(fmt.Errorf("iam client %s", err))
  }

  return svc
}


func (g *GCPCloud) projectNumber() (int64, error) {
  service := g.csmanager()
  op, err := service.Projects.Get(g.Project).Do()

  if err != nil {
    return 0, fmt.Errorf("fetching %s: %s",g.Project,err)
  }

  return op.ProjectNumber, nil
}

func (g *GCPCloud) getCluster(name string) (*container.Cluster, error) {
  service := g.container()
  return service.Projects.Zones.Clusters.Get(g.Project, g.Zone, name).Do()
}

func jwtClient(sva []byte) *http.Client {
  if _jwtClient != nil {
    return _jwtClient
  }

  jwtConfig, err := oauth2_google.JWTConfigFromJSON(sva, csm.CloudPlatformScope)
  if err != nil {
    log.Fatal(fmt.Errorf("JWT client %s", err))
  }

  _jwtClient = jwtConfig.Client(context.Background())
  return _jwtClient
}

func (g *GCPCloud) gsGet(bucket, key string) ([]byte, error) {
  service := g.storage()
  resp, err := service.Objects.Get(bucket, key).Download()
  if err != nil {
    return nil, fmt.Errorf("downloading gs://%s/%s: %v", bucket, key, err)
  }
  
  defer resp.Body.Close()
  return ioutil.ReadAll(resp.Body)
}

func (g *GCPCloud) gsPut(bucket, key string, body io.Reader) error {
  service := g.storage()
  _, err := service.Objects.Insert(bucket, &storage.Object{Name: key}).Media(body).Do()
  return err
}

func getKubeClientset(name string) (*kubernetes.Clientset, error) {
  config, err := clientcmd.BuildConfigFromFlags("", kubecfgPath(name))
  if err != nil {
    return nil, err
  }

  config.BearerToken = getCachedToken(name)

  c, err := kubernetes.NewForConfig(config)
  if err != nil {
    return nil, fmt.Errorf("cluster connection %v", err)
  }

  return c, nil
}
