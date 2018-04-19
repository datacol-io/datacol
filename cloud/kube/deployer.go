package kube

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/datacol-io/datacol/cloud"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	klabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

const (
	k8sAPIVersion     string = "v1"
	k8sBetaAPIVersion string = "extensions/v1beta1"
	managedBy         string = "managed-by"
	heritage          string = "datacol"
	appLabel          string = "app"
	typeLabel         string = "type"
	versionLabel      string = "version"
	runProcessKind    string = "run"
)

type Deployer struct {
	Client *kubernetes.Clientset
}

type DeployRequest struct {
	Entrypoint    []string
	Args          []string
	ContainerPort intstr.IntOrString
	Namespace     string
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
	Domains  []string
	Tags     map[string]string
	App      string // to specify pods belonging to an App
	Version  string // to specify the version of pod to deploy
	Proctype string // to specify the type of process web, worker, or other

	// For GCP, we need to provision cloudsql-proxy as a sidecar container.
	EnableCloudSqlProxy bool

	Provider cloud.CloudProvider // cloud provider aws, gcp or local
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

	if payload.Namespace == "" {
		return nil, fmt.Errorf("Namespace not set for DeployRequest.")
	}

	if payload.App == "" {
		return nil, fmt.Errorf("App not set for DeployRequest.")
	}

	if payload.Proctype == "" {
		return nil, fmt.Errorf("Proctype not set for DeployRequest.")
	}

	if string(payload.Provider) == "" {
		log.Debugf("DeploymentRequest %s", toJson(payload))
		return nil, fmt.Errorf("Provider is not set for DeployRequest")
	}

	// create namespace if needed
	if _, err := d.Client.Core().Namespaces().Create(newNamespace(payload)); err != nil {
		if !kerrors.IsAlreadyExists(err) {
			return nil, fmt.Errorf("creating namespace %v err: %v", payload.Namespace, err)
		}
	}

	// create deployment
	dp, err := d.CreateOrUpdateDeployment(payload)
	if err != nil {
		return res, fmt.Errorf("failed to create deployment %v", err)
	}

	if payload.ContainerPort.IntVal > 0 {
		// create service only of we have a contanerPort which can be exposed
		svc, err := d.CreateOrUpdateService(newService(payload), payload.Namespace)
		if err != nil {
			return res, err
		}

		if len(svc.Spec.Ports) > 0 {
			res.NodePort = int(svc.Spec.Ports[0].NodePort)
		}

		if len(payload.Domains) > 0 {
			_, err = d.CreateOrUpdateIngress(newIngress(res, payload.Domains), payload.Namespace, payload.Provider)
			if err != nil {
				return res, err
			}
		}
	}

	dpname := dp.ObjectMeta.Name

	if err := waitUntilDeploymentUpdated(d.Client, payload.Namespace, dpname); err != nil {
		return res, err
	}

	if err := waitUntilDeploymentReady(d.Client, payload.Namespace, dpname); err != nil {
		return res, err
	}

	log.Infof("Deployed %s in %s", payload.Proctype, payload.App)

	return res, nil
}

func (d *Deployer) Remove(r *DeployRequest) error {
	return nil
}

func newNamespace(payload *DeployRequest) *v1.Namespace {
	return &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: payload.Namespace},
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

		updatedSvc, err := r.Client.Core().Services(env).Update(svc)

		if err != nil {
			log.Debugf("was trying to update service %s with %s", svc.Name, toJson(svc))
			return nil, fmt.Errorf("Failed to update service %s: %v", svc.Name, err)
		}

		log.Debugf("Service updated: %+v", svc.Name)
		return updatedSvc, nil
	}

	log.Debugf("Service created: %+v", svc.Name)
	return newsSvc, nil
}

func newService(payload *DeployRequest) *v1.Service {
	serviceType := v1.ServiceTypeLoadBalancer

	// we will create an Ingress for if domain is provided
	if len(payload.Domains) > 0 {
		serviceType = v1.ServiceTypeNodePort
	}

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: payload.Tags,
			Labels:      map[string]string{appLabel: payload.App, managedBy: heritage},
			Name:        payload.ServiceID,
			Namespace:   payload.Namespace,
		},
		Spec: v1.ServiceSpec{
			Type: serviceType,
			Ports: []v1.ServicePort{{
				TargetPort: payload.ContainerPort,
				Port:       payload.ContainerPort.IntVal,
			}},
			Selector: map[string]string{appLabel: payload.App},
		},
		TypeMeta: metav1.TypeMeta{APIVersion: k8sAPIVersion, Kind: "Service"},
	}

	if len(payload.Domains) == 0 {
		svc.Spec.Ports[0].Port = 80
	}

	return svc
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
	env := payload.Namespace
	var d *v1beta1.Deployment

	found := false
	d, err := r.Client.Extensions().Deployments(env).Get(payload.ServiceID, metav1.GetOptions{})

	if err == nil {
		found = true
		if i, _ := findContainer(d, payload.ServiceID); i >= 0 {
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

// Values can be taken from https://github.com/giantswarm/k8scloudconfig/blob/master/v_2_0_0/master_template.go
// OR https://github.com/Electroid/infrastructure/tree/master/kubernetes/ingress
func (r *Deployer) setupAWSIngressController(ns string) (err error) {
	runner := &awsIngress{Client: r.Client, Namespace: ingressDefaultNamespace, ParentNamespace: ns}
	return runner.CreateOrUpdate()
}

// CreateOrUpdateIngress creates or updates an ingress rule
func (r *Deployer) CreateOrUpdateIngress(ingress *v1beta1.Ingress, env string, provider cloud.CloudProvider) (*v1beta1.Ingress, error) {
	if provider == cloud.AwsProvider {
		log.Debugf("Will try to setup nginx ingress controller for ingress:%s/%s", env, ingress.Name)
		if err := r.setupAWSIngressController(env); err != nil {
			log.Error(err)
			return nil, fmt.Errorf("nginx ingress controller setup error: %v", err)
		}
	}

	existing, err := r.Client.Extensions().Ingresses(env).Get(ingress.Name, metav1.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			log.Debugf("Ingress created: %+v", ingress.Name)
			return r.Client.Extensions().Ingresses(env).Create(ingress)
		}
		return nil, err
	}

	ingress = mergeIngressRules(existing, ingress)

	newIngress, err := r.Client.Extensions().Ingresses(env).Update(ingress)
	if err != nil {
		return nil, err
	}

	log.Debugf("Ingress spec updated: %v", toJson(newIngress))
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
