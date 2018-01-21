package kube

import (
	"fmt"

	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"

	log "github.com/Sirupsen/logrus"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	klabels "k8s.io/apimachinery/pkg/labels"
)

const (
	k8sAPIVersion     string = "v1"
	k8sBetaAPIVersion string = "extensions/v1beta1"
)

type Deployer struct {
	Client *kubernetes.Clientset
}

type DeployRequest struct {
	Entrypoint    []string
	Args          []string
	ContainerPort intstr.IntOrString
	Environment   string
	EnvVars       map[string]string
	Heartbeat     struct {
		Path                         string
		Port                         intstr.IntOrString
		InitialDelayLivenessSeconds  int
		InitialDelayReadinessSeconds int
		TimeoutSeconds               int32
	}
	Image     string
	Replicas  *int32
	ServiceID string
	Secrets   []struct {
		Name  string
		Value string
	}
	Domains []string
	Tags    map[string]string
	Zone    string
	Tier    string // to specify pods belonging to an App
}

type DeployResponse struct {
	Request  DeployRequest
	NodePort int
}

func NewDeployer(c *kubernetes.Clientset) (*Deployer, error) {
	return &Deployer{Client: c}, nil
}

func (d *Deployer) Run(payload *DeployRequest) (*DeployResponse, error) {
	res := &DeployResponse{Request: *payload}

	if payload.Environment == "" {
		return nil, fmt.Errorf("environment not set for DeployRequest.")
	}

	// create namespace if needed
	if _, err := d.Client.Core().Namespaces().Create(newNamespace(payload)); err != nil {
		if !kerrors.IsAlreadyExists(err) {
			return nil, fmt.Errorf("creating namespace %v err: %v", payload.Environment, err)
		}
	}

	// create deployment
	dp, err := d.CreateOrUpdateDeployment(payload)
	if err != nil {
		return res, fmt.Errorf("failed to create deployment %v", err)
	}

	if payload.ContainerPort.IntVal > 0 {
		// create service only of we have a contanerPort which can be exposed
		svc, err := d.CreateOrUpdateService(newService(payload), payload.Environment)
		if err != nil {
			return res, fmt.Errorf("failed to create service %v", err)
		}

		if len(svc.Spec.Ports) > 0 {
			res.NodePort = int(svc.Spec.Ports[0].NodePort)
		}

		if len(payload.Domains) > 0 {
			_, err = d.CreateOrUpdateIngress(newIngress(res, payload.Domains), payload.Environment)
			if err != nil {
				return res, err
			}
		}
	}

	dpname := dp.ObjectMeta.Name

	if err := waitUntilDeploymentUpdated(d.Client, payload.Environment, dpname); err != nil {
		return res, err
	}

	if err := waitUntilDeploymentReady(d.Client, payload.Environment, dpname); err != nil {
		return res, err
	}

	log.Infof("Deployment completed: %+v", dpname)
	return res, nil
}

func (d *Deployer) Remove(r *DeployRequest) error {
	return nil
}

func newNamespace(payload *DeployRequest) *v1.Namespace {
	return &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: payload.Environment},
		TypeMeta:   metav1.TypeMeta{APIVersion: k8sAPIVersion, Kind: "Namespace"},
	}
}

// CreateOrUpdateService creates or updates a service
func (r *Deployer) CreateOrUpdateService(svc *v1.Service, env string) (*v1.Service, error) {
	newsSvc, err := r.Client.Core().Services(env).Create(svc)
	if err != nil {
		if !kerrors.IsAlreadyExists(err) {
			return nil, err
		}
		oldSvc, err := r.Client.Core().Services(env).Get(svc.ObjectMeta.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		svc.ObjectMeta.ResourceVersion = oldSvc.ObjectMeta.ResourceVersion
		svc.Spec.ClusterIP = oldSvc.Spec.ClusterIP
		svc.Spec.Ports[0].NodePort = oldSvc.Spec.Ports[0].NodePort
		svc, err = r.Client.Core().Services(env).Update(svc)

		if err != nil {
			return nil, err
		}
		log.Debugf("Service updated: %+v", svc.ObjectMeta.Name)
		return svc, nil
	}
	log.Debugf("Service created: %+v", svc.ObjectMeta.Name)
	return newsSvc, nil
}

func newService(payload *DeployRequest) *v1.Service {
	serviceType := v1.ServiceTypeLoadBalancer

	// we will create an Ingress for if domain is provided
	if len(payload.Domains) > 0 {
		serviceType = v1.ServiceTypeNodePort
	}

	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: payload.Tags,
			Labels:      map[string]string{"name": payload.ServiceID},
			Name:        payload.ServiceID,
			Namespace:   payload.Environment,
		},
		Spec: v1.ServiceSpec{
			Type: serviceType,
			Ports: []v1.ServicePort{{
				Port: payload.ContainerPort.IntVal,
			}},
			Selector: map[string]string{"name": payload.ServiceID},
		},
		TypeMeta: metav1.TypeMeta{APIVersion: k8sAPIVersion, Kind: "Service"},
	}
}

func newMetadata(payload *DeployRequest) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Annotations: payload.Tags,
		Labels: map[string]string{
			"name":          payload.ServiceID,
			ServiceLabelKey: payload.Tier,
		},
		Name:      payload.ServiceID,
		Namespace: payload.Environment,
	}
}

func findContainer(dp *v1beta1.Deployment, name string) (int, *v1.Container) {
	for i, c := range dp.Spec.Template.Spec.Containers {
		if c.Name == name {
			return i, &c
		}
	}
	return -1, nil
}

// CreateOrUpdateDeployment creates or updates a service
func (r *Deployer) CreateOrUpdateDeployment(payload *DeployRequest) (*v1beta1.Deployment, error) {
	env := payload.Environment
	var d *v1beta1.Deployment

	found := false
	d, err := r.Client.Extensions().Deployments(env).Get(payload.ServiceID, metav1.GetOptions{})

	if err == nil {
		found = true
		i, _ := findContainer(d, payload.ServiceID)
		if i >= 0 {
			d.Spec.Template.Spec.Containers[i] = newContainer(payload)
			//TODO: we are only updating containers schema for existing deployment. Add support for updating any any schema change
			//Below is one workaround of it.

			if payload.Replicas != nil {
				d.Spec.Replicas = payload.Replicas
			}
		}
	} else {
		d = newDeployment(payload)
	}

	if !found {
		d, err := r.Client.Extensions().Deployments(env).Create(d)
		if err != nil {
			return nil, err
		}

		log.Debugf("Deployment created: %+v", d.ObjectMeta.Name)
	} else {
		d, err = r.Client.Extensions().Deployments(env).Update(d)
		if err != nil {
			return nil, err
		}

		log.Debugf("Deployment updated: %+v", d.ObjectMeta.Name)
	}

	log.WithField("image", d.Spec.Template.Spec.Containers[0].Image).Info("Deployment")
	log.Debugf("Deployment:\n %s", toJson(d))

	return d, nil
}

// CreateOrUpdateIngress creates or updates an ingress rule
func (r *Deployer) CreateOrUpdateIngress(ingress *v1beta1.Ingress, env string) (*v1beta1.Ingress, error) {
	newIngress, err := r.Client.Extensions().Ingresses(env).Create(ingress)
	if err != nil {
		if !kerrors.IsAlreadyExists(err) {
			return nil, err
		}

		ingress, err = r.Client.Extensions().Ingresses(env).Update(ingress)
		if err != nil {
			return nil, err
		}

		log.Debugf("Ingress updated: %+v", ingress.ObjectMeta.Name)
		return ingress, nil
	}
	log.Debugf("Ingress created: %+v", ingress.ObjectMeta.Name)
	return newIngress, nil
}

func checkforFailedEvents(c *kubernetes.Clientset, ns string, labels map[string]string) {
	selector := klabels.Set(labels).AsSelector()
	res, err := c.Extensions().ReplicaSets(ns).List(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		log.Fatal(err)
	}

	fields := map[string]string{
		"involvedObject.kind":      "ReplicaSet",
		"involvedObject.name":      res.Items[0].ObjectMeta.Name,
		"involvedObject.namespace": ns,
		"involvedObject.uid":       string(res.Items[0].ObjectMeta.UID),
	}

	selector = klabels.Set(fields).AsSelector()
	response, err := c.Core().Events(ns).List(metav1.ListOptions{FieldSelector: selector.String()})

	for _, event := range response.Items {
		log.Debugf("event %s reason:%s", event.Message, event.Reason)

		if event.Reason == "FailedCreate" {
			log.Fatal(fmt.Errorf(
				"Message:%s lastTimestamp:%v reason:%s count:%d",
				event.Message, event.LastTimestamp,
				event.Reason, event.Count,
			))
		}
	}
}
