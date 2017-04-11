package google

import (
  "context"
  "net/http"
  "fmt"
  "os"
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
  _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

  kerrors "k8s.io/client-go/pkg/api/errors"
  kapi "k8s.io/client-go/pkg/api/v1"
  klabels "k8s.io/client-go/pkg/labels"
  "google.golang.org/api/googleapi"

  "github.com/dinesh/datacol/client/models"
)

var _jwtClient *http.Client

type GCPCloud struct {
  Project        string
  ProjectNumber  int64
  DeploymentName string
  BucketName     string
  Zone           string
  ServiceKey     []byte
}

func (g *GCPCloud) AppCreate(app *models.App) error { 
  return nil
}

func (g *GCPCloud) AppGet(name string) (*models.App, error) {
  app := &models.App{Name: name}

  ns := g.DeploymentName
  kube, err := getKubeClientset(ns)
  if err != nil { return app, err }

  svc, err := kube.Core().Services(ns).Get(name)
  if err != nil { return app, err }

  if svc.Spec.Type == kapi.ServiceTypeLoadBalancer && len(svc.Status.LoadBalancer.Ingress) > 0 {
    ing := svc.Status.LoadBalancer.Ingress[0]
    if len(ing.Hostname) > 0 {
      app.HostPort = ing.Hostname
    } else {
      port := 80
      if len(svc.Spec.Ports) > 0 {
        port = int(svc.Spec.Ports[0].Port)
      }
      app.HostPort = fmt.Sprintf("%s:%d", ing.IP, port)
    }
  }

  return app, nil
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
    if gerr, ok := err.(*googleapi.Error); ok && gerr.Code == 404 {
      return models.Environment{}, nil
    }
    return nil, err
  }

  return loadEnv(data), nil
}

func (g *GCPCloud) EnvironmentSet(name string, body io.Reader) error {
  gskey := fmt.Sprintf("%s.env", name)
  return g.gsPut(g.BucketName, gskey, body)
}


func (g *GCPCloud) GetRunningPods(app string) (string, error) {
  ns := g.DeploymentName
  c, err := getKubeClientset(ns)
  if err != nil { return "", err }
  
  return runningPods(ns, app, c)
}

func runningPods(ns, app string, c *kubernetes.Clientset) (string, error) {
  selector := klabels.Set(map[string]string{"name": app}).AsSelector()
  res, err := c.Core().Pods(ns).List(kapi.ListOptions{LabelSelector: selector.String()})
  if err != nil {
    return "", err
  }

  var podNames []string
  for _, p := range res.Items {
    podNames = append(podNames, p.Name)
  }
  
  if len(podNames) < 1 {
    return "", fmt.Errorf("No pod running for %s", app)
  }

  return podNames[0], nil
}

func (g *GCPCloud) LogStream(app string, out io.Writer, opts models.LogStreamOptions) error {
  ns := g.DeploymentName
  c, err := getKubeClientset(ns)
  if err != nil { return err }
  
  pod, err := runningPods(ns, app, c)
  if err != nil { return err }

  req := c.Core().RESTClient().Get().
    Namespace(g.DeploymentName).
    Name(pod).
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

func (g *GCPCloud) cacheCredentials() (string, error) {
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

  _jwtClient = jwtConfig.Client(context.TODO())
  return _jwtClient
}

func (g *GCPCloud) gsGet(bucket, key string) ([]byte, error) {
  service := g.storage()
  resp, err := service.Objects.Get(bucket, key).Download()
  if err != nil {
    return nil, err
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
  svapath := filepath.Join(models.ConfigPath, name, models.SvaFilename)
  
  if err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", svapath); err != nil {
    return nil, err
  }

  config, err := clientcmd.BuildConfigFromFlags("", kubecfgPath(name))
  if err != nil {
    return nil, err
  }

  c, err := kubernetes.NewForConfig(config)
  if err != nil {
    return nil, fmt.Errorf("cluster connection %v", err)
  }

  return c, nil
}
