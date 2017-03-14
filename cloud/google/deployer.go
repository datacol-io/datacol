package google

import (
  "fmt"
  log "github.com/Sirupsen/logrus"

  "k8s.io/client-go/kubernetes"
  "k8s.io/client-go/pkg/api/v1"
  "k8s.io/client-go/pkg/api/unversioned"
  "k8s.io/client-go/pkg/util/intstr"
  "k8s.io/client-go/pkg/apis/extensions/v1beta1"
  "k8s.io/client-go/pkg/watch"
  "k8s.io/client-go/pkg/labels"
  "k8s.io/client-go/pkg/selection"
  "k8s.io/client-go/tools/clientcmd"
  apierrs "k8s.io/client-go/pkg/api/errors"
)

const (
  k8sAPIVersion     string = "v1"
  k8sBetaAPIVersion string = "extensions/v1beta1"
)

type Deployer struct {
  Client *kubernetes.Clientset
}

type DeployRequest struct {
  Args          []string           `json:"arguments"`
  ContainerPort intstr.IntOrString `json:"containerPort"`
  Environment   string             `json:"environment"`
  EnvVars       map[string]string  `json:"envVars"`
  Heartbeat     struct {
    Path                         string             `json:"path"`
    Port                         intstr.IntOrString `json:"port"`
    InitialDelayLivenessSeconds  int                `json:"initialDelayLivenessSeconds"`
    InitialDelayReadinessSeconds int                `json:"initialDelayReadinessSeconds"`
    TimeoutSeconds               int32              `json:"timeoutSeconds"`
  } `json:"heartbeat"`
  Image     string  `json:"image"`
  Replicas  int32   `json:"replicas"`
  ServiceID string  `json:"serviceId"`
  Secrets   []struct {
    Name  string `json:"name"`
    Value string `json:"value"`
  } `json:"secrets"`
  Tags map[string]string `json:"tags"`
  Zone string            `json:"zone"`
}

type DeployResponse struct {
  Request  DeployRequest `json:"request"`
  NodePort int           `json:"nodePort"`
}

func NewDeployer(cfgpath string, token string) (*Deployer, error) {
  config, err := clientcmd.BuildConfigFromFlags("", cfgpath)
  if err != nil {
    return nil, err 
  }

  config.BearerToken = token
  fmt.Printf("rest config: %+v\n", config)

  c, err := kubernetes.NewForConfig(config)
  if err != nil {
    return nil, fmt.Errorf("cluster connection %v", err)
  }

  return &Deployer{c}, nil
}


func (d *Deployer) Run(payload *DeployRequest) (*DeployResponse, error) {
    res := &DeployResponse{Request: *payload}
    
    if payload.Environment == "" {
      return nil, fmt.Errorf("environment not found.")
    }

    // create namespace if needed
    if _, err := d.Client.Core().Namespaces().Create(newNamespace(payload)); err != nil {
      if !apierrs.IsAlreadyExists(err) {
        return nil, fmt.Errorf("creating namespace %v err: %v", payload.Environment, err)
      }
    }

    // create service
    svc, err := d.CreateOrUpdateService(newService(payload), payload.Environment)
    if err != nil {
      return res, fmt.Errorf("failed to create service %v", err)
    }

    if len(svc.Spec.Ports) > 0 {
      res.NodePort = int(svc.Spec.Ports[0].NodePort)
    }

    // create deployment
    deployment, err := d.CreateOrUpdateDeployment(newDeployment(payload), payload.Environment)
    if err != nil {
      return res, fmt.Errorf("failed to create deployment %v", err)
    }

    // get deployment status
    r, err := labels.NewRequirement("name", selection.Equals, []string{payload.ServiceID})
    if err != nil {
      return res, err
    }

    watcher, err := d.Client.Deployments(payload.Environment).Watch(v1.ListOptions{
      LabelSelector:  r.String(),
      ResourceVersion: deployment.ResourceVersion,
    })

    if err != nil {
      return res, err
    }

    // TODO: timeout?
    d.WatchLoop(watcher, func(e watch.Event) bool {
      switch e.Type {
      case watch.Modified:
        d, err := d.Client.Deployments(payload.Environment).Get(payload.ServiceID)
        if err != nil {
          log.Errorf("Error getting deployment: %+v", err)
          return true
        }
        if d.Spec.Replicas != nil && d.Status.Replicas == *d.Spec.Replicas {
          return true
        }
      }
      return false
    })

    _, err = d.CreateOrUpdateIngress(newIngress(res), payload.Environment)
    if err != nil {
      return res, err
    }

    log.Infof("Deployment completed: %+v", svc)
    return res, nil
}

// WatchLoop loops, passing events in w to fn.
func (r *Deployer) WatchLoop(w watch.Interface, fn func(watch.Event) bool) {
  for {
    select {
    case event, ok := <-w.ResultChan():
      if !ok {
        log.Info("No more events")
        return
      }
      if stop := fn(event); stop {
        w.Stop()
      }
    }
  }
}

func newNamespace(payload *DeployRequest) *v1.Namespace {
  return &v1.Namespace{
    ObjectMeta: v1.ObjectMeta{Name: payload.Environment},
    TypeMeta:   unversioned.TypeMeta{APIVersion: k8sAPIVersion, Kind: "Namespace"},
  }
}

// CreateOrUpdateService creates or updates a service
func (r *Deployer) CreateOrUpdateService(svc *v1.Service, env string) (*v1.Service, error) {
  newsSvc, err := r.Client.Services(env).Create(svc)
  if err != nil {
    if !apierrs.IsAlreadyExists(err) {
      return nil, err
    }
    oldSvc, err := r.Client.Services(env).Get(svc.ObjectMeta.Name)
    if err != nil {
      return nil, err
    }
    svc.ObjectMeta.ResourceVersion = oldSvc.ObjectMeta.ResourceVersion
    svc.Spec.ClusterIP = oldSvc.Spec.ClusterIP
    svc.Spec.Ports[0].NodePort = oldSvc.Spec.Ports[0].NodePort
    svc, err = r.Client.Services(env).Update(svc)
    if err != nil {
      return nil, err
    }
    log.Infof("Service updated: %+v", svc)
    return svc, nil
  }
  log.Infof("Service created: %+v", svc)
  return newsSvc, nil
}

func newService(payload *DeployRequest) *v1.Service {
  return &v1.Service{
    ObjectMeta: v1.ObjectMeta{
      Annotations: payload.Tags,
      Labels:      map[string]string{"name": payload.ServiceID},
      Name:        payload.ServiceID,
      Namespace:   payload.Environment,
    },
    Spec: v1.ServiceSpec{
      Type: v1.ServiceTypeNodePort,
      Ports: []v1.ServicePort{{
        Port: payload.ContainerPort.IntVal,
      }},
      Selector: map[string]string{"name": payload.ServiceID},
    },
    TypeMeta: unversioned.TypeMeta{APIVersion: k8sAPIVersion, Kind: "Service"},
  }
}

func newMetadata(payload *DeployRequest) v1.ObjectMeta {
  return v1.ObjectMeta{
    Annotations: payload.Tags,
    Labels:      map[string]string{"name": payload.ServiceID},
    Name:        payload.ServiceID,
    Namespace:   payload.Environment,
  }
}

// CreateOrUpdateDeployment creates or updates a service
func (r *Deployer) CreateOrUpdateDeployment(d *v1beta1.Deployment, env string) (*v1beta1.Deployment, error) {
  log.WithField("image", d.Spec.Template.Spec.Containers[0].Image).Info("Deployment")

  newD, err := r.Client.Deployments(env).Create(d)
  if err != nil {
    if !apierrs.IsAlreadyExists(err) {
      return nil, err
    }
    d, err = r.Client.Deployments(env).Update(d)
    if err != nil {
      return nil, err
    }
    log.Infof("Deployment updated: %+v", d)
    return d, nil

  }
  log.Infof("Deployment created: %+v", d)
  return newD, nil
}

func newDeployment(payload *DeployRequest) *v1beta1.Deployment {
  envVars := []v1.EnvVar{}
  for k, v := range payload.EnvVars {
    envVars = append(envVars, v1.EnvVar{Name: k, Value: v})
  }

  maxunavailable := intstr.FromString("10%")
  maxsurge := intstr.FromString("10%")

  return &v1beta1.Deployment{
    ObjectMeta: newMetadata(payload),
    Spec: v1beta1.DeploymentSpec{
      Replicas: &payload.Replicas,
      Selector: &unversioned.LabelSelector{MatchLabels: map[string]string{"name": payload.ServiceID}},
      Strategy: v1beta1.DeploymentStrategy{
        Type: v1beta1.RollingUpdateDeploymentStrategyType,
        RollingUpdate: &v1beta1.RollingUpdateDeployment{
          MaxUnavailable: &maxunavailable,
          MaxSurge:       &maxsurge,
        },
      },
      Template: v1.PodTemplateSpec{
        ObjectMeta: newMetadata(payload),
        Spec: v1.PodSpec{
          Containers: []v1.Container{
            {
              Args:  payload.Args,
              Name:  payload.ServiceID,
              Image: payload.Image,
              ImagePullPolicy: "IfNotPresent",
              Ports: []v1.ContainerPort{{
                Name:          "http",
                ContainerPort: int32(payload.ContainerPort.IntVal),
              }},
              Env: envVars,
            },
          },
          RestartPolicy: "Always",
        },
      },
    },
    TypeMeta: unversioned.TypeMeta{APIVersion: k8sBetaAPIVersion, Kind: "Deployment"},
  }
}

func newProbe(payload *DeployRequest, delay int32) *v1.Probe {
  return &v1.Probe{
    Handler: v1.Handler{HTTPGet: &v1.HTTPGetAction{
      Path: payload.Heartbeat.Path,
      Port: payload.ContainerPort,
    }},
    InitialDelaySeconds: delay,
    TimeoutSeconds:      payload.Heartbeat.TimeoutSeconds,
  }
}

// CreateOrUpdateIngress creates or updates an ingress rule
func (r *Deployer) CreateOrUpdateIngress(ingress *v1beta1.Ingress, env string) (*v1beta1.Ingress, error) {
  newIngress, err := r.Client.Extensions().Ingresses(env).Create(ingress)
  if err != nil {
    if !apierrs.IsAlreadyExists(err) {
      return nil, err
    }
    ingress, err = r.Client.Extensions().Ingresses(env).Update(ingress)
    if err != nil {
      return nil, err
    }
    log.Infof("Ingress updated: %+v", ingress)
    return ingress, nil

  }
  log.Infof("Ingress created: %+v", ingress)
  return newIngress, nil
}

func newIngress(payload *DeployResponse) *v1beta1.Ingress {
  r := payload.Request
  return &v1beta1.Ingress{
    ObjectMeta: newMetadata(&payload.Request),
    Spec: v1beta1.IngressSpec{
      Rules: []v1beta1.IngressRule{{
        Host: fmt.Sprintf("%s.%s", r.ServiceID, r.Zone),
        IngressRuleValue: v1beta1.IngressRuleValue{HTTP: &v1beta1.HTTPIngressRuleValue{
          Paths: []v1beta1.HTTPIngressPath{{Path: "/", Backend: v1beta1.IngressBackend{
            ServiceName: r.ServiceID,
            ServicePort: r.ContainerPort,
          }}},
        }},
      }},
    },
    TypeMeta: unversioned.TypeMeta{APIVersion: k8sBetaAPIVersion, Kind: "Ingress"},
  }
}
