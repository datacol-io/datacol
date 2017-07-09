package kube

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"sort"
	"time"

	"k8s.io/client-go/kubernetes"
	kerrors "k8s.io/client-go/pkg/api/errors"
	"k8s.io/client-go/pkg/api/unversioned"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	klabels "k8s.io/client-go/pkg/labels"
	"k8s.io/client-go/pkg/util/intstr"
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
	Image     string `json:"image"`
	Replicas  int32  `json:"replicas"`
	ServiceID string `json:"serviceId"`
	Secrets   []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	} `json:"secrets"`
	SSL  bool              `json:"ssl"`
	Tags map[string]string `json:"tags"`
	Zone string            `json:"zone"`
}

type DeployResponse struct {
	Request  DeployRequest `json:"request"`
	NodePort int           `json:"nodePort"`
}

func NewDeployer(c *kubernetes.Clientset) (*Deployer, error) {
	return &Deployer{Client: c}, nil
}

func (d *Deployer) Run(payload *DeployRequest) (*DeployResponse, error) {
	res := &DeployResponse{Request: *payload}

	if payload.Environment == "" {
		return nil, fmt.Errorf("environment not found.")
	}

	// create namespace if needed
	if _, err := d.Client.Core().Namespaces().Create(newNamespace(payload)); err != nil {
		if !kerrors.IsAlreadyExists(err) {
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
	_, err = d.CreateOrUpdateDeployment(payload)
	if err != nil {
		return res, fmt.Errorf("failed to create deployment %v", err)
	}

	if payload.SSL {
		_, err = d.CreateOrUpdateIngress(newIngress(res), payload.Environment)
		if err != nil {
			return res, err
		}
	}

	dpname := svc.ObjectMeta.Name

	WaitUntilUpdated(d.Client, payload.Environment, dpname)
	WaitUntilReady(d.Client, payload.Environment, dpname)

	log.Infof("Deployment completed: %+v", svc.ObjectMeta.Name)
	return res, nil
}

func (d *Deployer) Remove(r *DeployRequest) error {
	return nil
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
		if !kerrors.IsAlreadyExists(err) {
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
		log.Debugf("Service updated: %+v", svc.ObjectMeta.Name)
		return svc, nil
	}
	log.Debugf("Service created: %+v", svc.ObjectMeta.Name)
	return newsSvc, nil
}

func newService(payload *DeployRequest) *v1.Service {
	serviceType := v1.ServiceTypeLoadBalancer
	if payload.SSL {
		serviceType = v1.ServiceTypeNodePort
	}

	return &v1.Service{
		ObjectMeta: v1.ObjectMeta{
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
	d, err := r.Client.Deployments(env).Get(payload.ServiceID)

	if err == nil {
		found = true
		i, _ := findContainer(d, payload.ServiceID)
		if i >= 0 {
			d.Spec.Template.Spec.Containers[i] = newContainer(payload)
		}
	} else {
		d = newDeployment(payload)
	}

	if !found {
		d, err := r.Client.Deployments(env).Create(d)
		if err != nil {
			return nil, err
		}

		log.Debugf("Deployment created: %+v", d.ObjectMeta.Name)
	} else {
		d, err = r.Client.Deployments(env).Update(d)
		if err != nil {
			return nil, err
		}

		log.Debugf("Deployment updated: %+v", d.ObjectMeta.Name)
	}

	log.WithField("image", d.Spec.Template.Spec.Containers[0].Image).Info("Deployment")
	log.Debugf("Deployment:\n %s", toJson(d))

	return d, nil
}

func newDeployment(payload *DeployRequest) *v1beta1.Deployment {
	maxunavailable := intstr.FromString("25%")
	maxsurge := intstr.FromString("25%")

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
						newContainer(payload),
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

func newContainer(payload *DeployRequest) v1.Container {
	envVars := []v1.EnvVar{}
	for k, v := range payload.EnvVars {
		envVars = append(envVars, v1.EnvVar{Name: k, Value: v})
	}

	return v1.Container{
		Args:            payload.Args,
		Name:            payload.ServiceID,
		Image:           payload.Image,
		ImagePullPolicy: "Always",
		Ports: []v1.ContainerPort{{
			Name:          "http",
			ContainerPort: int32(payload.ContainerPort.IntVal),
		}},
		ReadinessProbe: &v1.Probe{
			Handler: v1.Handler{
				TCPSocket: &v1.TCPSocketAction{
					Port: payload.ContainerPort,
				},
			},
			InitialDelaySeconds: 5,
			TimeoutSeconds:      1,
			PeriodSeconds:       5,
			FailureThreshold:    1,
		},
		Env: envVars,
	}
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

func WaitUntilUpdated(c *kubernetes.Clientset, ns, name string) {
	log.Debugf("waiting for Deployment %s to get a newer generation (30s timeout)", name)
	for i := 0; i < 30; i++ {
		dp, err := c.Extensions().Deployments(ns).Get(name)
		if err != nil {
			if kerrors.IsNotFound(err) {
				time.Sleep(1 * time.Second)
				continue
			}
			log.Fatal(err)
		}

		if dp.Status.ObservedGeneration >= dp.ObjectMeta.Generation {
			log.Debugf("A newer generation was found for Deployment %s", name)
			break
		}
		time.Sleep(1 * time.Second)
	}
}

func WaitUntilReady(c *kubernetes.Clientset, ns, name string) {
	dp, err := c.Extensions().Deployments(ns).Get(name)
	if err != nil {
		log.Fatal(err)
	}

	labels := dp.Spec.Template.ObjectMeta.Labels
	checkforFailedEvents(c, ns, labels)

	timeout := 120
	waited := 0

	log.Infof("waiting for pods to get ready in Deployment %s (%ds timeout)", name, timeout)

	for {
		time.Sleep(1 * time.Second)

		if waited >= timeout {
			break
		}

		ready, availablePods := areReplicaReady(c, ns, name, labels)
		if ready {
			break
		}

		if waited > 0 && (waited%10) == 0 {
			log.Infof("waited %ds and %d pods", waited, availablePods)
		}

		waited += 1
	}

	ready, _ := areReplicaReady(c, ns, name, labels)
	if !ready {
		handleNotReadyPods(c, ns, labels)
	}
}

func handleNotReadyPods(c *kubernetes.Clientset, ns string, labels map[string]string) {
	selector := klabels.Set(labels).AsSelector()
	res, err := c.Core().Pods(ns).List(v1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		log.Fatal(err)
	}

	for _, pod := range res.Items {
		if pod.Status.Phase != v1.PodRunning {
			continue
		}

		name, ok := labels["name"]
		if !ok {
			log.Fatal(fmt.Errorf("name not found in %+v", labels))
		}

		cstatus := v1.ContainerStatus{}
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.Name == name {
				cstatus = cs
				break
			}
		}

		log.Debugf("status: %s", toJson(cstatus))
		if cstatus.Ready {
			continue
		}

		res, err := podEvents(c, ns, &pod)
		if err != nil {
			log.Fatal(err)
		}

		for _, ev := range res.Items {
			if ev.Reason == "Unhealthy" || ev.Reason == "Failed" {
				log.Fatal(fmt.Errorf(ev.Message))
			}
		}
	}
}

func podEvents(c *kubernetes.Clientset, ns string, pod *v1.Pod) (*v1.EventList, error) {
	fields := map[string]string{
		"involvedObject.name":      pod.ObjectMeta.Name,
		"involvedObject.namespace": ns,
		"involvedObject.uid":       string(pod.ObjectMeta.UID),
	}

	res, err := c.Core().Events(ns).List(v1.ListOptions{
		FieldSelector:   klabels.Set(fields).AsSelector().String(),
		ResourceVersion: pod.ObjectMeta.ResourceVersion,
	})
	if err != nil {
		return res, err
	}

	sort.Slice(res.Items, func(i, j int) bool {
		return res.Items[j].LastTimestamp.Before(res.Items[i].LastTimestamp)
	})

	return res, err
}

func areReplicaReady(c *kubernetes.Clientset, ns, name string, labels map[string]string) (bool, int32) {
	dp, err := c.Extensions().Deployments(ns).Get(name)
	if err != nil {
		log.Fatal(err)
	}

	desired := dp.Spec.Replicas
	if desired == nil {
		return true, 0
	}

	status := dp.Status
	pods := status.UpdatedReplicas

	if status.UnavailableReplicas > 0 ||
		status.Replicas != *desired ||
		status.UpdatedReplicas != *desired ||
		status.AvailableReplicas != *desired {
		return false, pods
	}

	return true, pods
}

func checkforFailedEvents(c *kubernetes.Clientset, ns string, labels map[string]string) {
	selector := klabels.Set(labels).AsSelector()
	res, err := c.Extensions().ReplicaSets(ns).List(v1.ListOptions{LabelSelector: selector.String()})
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
	response, err := c.Core().Events(ns).List(v1.ListOptions{FieldSelector: selector.String()})

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

func RunningPods(ns, app string, c *kubernetes.Clientset) (string, error) {
	selector := klabels.Set(map[string]string{"name": app}).AsSelector()
	res, err := c.Core().Pods(ns).List(v1.ListOptions{LabelSelector: selector.String()})
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
