package kube

import (
	"fmt"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"

	log "github.com/Sirupsen/logrus"
	"github.com/datacol-io/datacol/cloud"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newDeployment(payload *DeployRequest) *v1beta1.Deployment {
	maxunavailable := intstr.FromString("25%")
	maxsurge := intstr.FromString("25%")

	labels := map[string]string{
		appLabel:  payload.App,
		typeLabel: payload.Proctype,
		managedBy: heritage,
	}

	return &v1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:   payload.ServiceID,
			Labels: labels,
		},
		Spec: v1beta1.DeploymentSpec{
			Replicas: payload.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Strategy: v1beta1.DeploymentStrategy{
				Type: v1beta1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &v1beta1.RollingUpdateDeployment{
					MaxUnavailable: &maxunavailable,
					MaxSurge:       &maxsurge,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: newPodMetadata(payload),
				Spec:       newPodSpec(payload).Spec,
			},
		},
		TypeMeta: metav1.TypeMeta{APIVersion: k8sBetaAPIVersion, Kind: "Deployment"},
	}
}

func newPodMetadata(req *DeployRequest) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Annotations: req.Tags,
		Labels: map[string]string{
			"app":     req.App,
			"version": req.Version,
			"type":    req.Proctype,
			managedBy: heritage,
		},
		Name:      req.ServiceID,
		Namespace: req.Namespace,
	}
}

func newPodSpec(req *DeployRequest) *v1.Pod {
	pod := &v1.Pod{
		ObjectMeta: newPodMetadata(req),
		Spec: v1.PodSpec{
			Containers: []v1.Container{newContainer(req)},
		},
	}

	if req.EnableCloudSqlProxy {
		MergeCloudSQLManifest(&pod.Spec, req.App, req.EnvVars)
	}

	return pod
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

	container := v1.Container{
		Command:         payload.Entrypoint,
		Args:            payload.Args,
		Name:            payload.ServiceID,
		Image:           payload.Image,
		ImagePullPolicy: v1.PullIfNotPresent,
		Env:             envVars,
	}

	if payload.ContainerPort.IntVal > 0 {
		container.Ports = []v1.ContainerPort{{
			Name:          "http",
			ContainerPort: int32(payload.ContainerPort.IntVal),
		}}

		container.ReadinessProbe = &v1.Probe{
			Handler: v1.Handler{
				TCPSocket: &v1.TCPSocketAction{
					Port: payload.ContainerPort,
				},
			},
			InitialDelaySeconds: 5,
			TimeoutSeconds:      1,
			PeriodSeconds:       5,
			FailureThreshold:    1,
		}
	}

	return container
}

func newIngress(payload *DeployResponse, domains []string) *v1beta1.Ingress {
	r := payload.Request

	if len(domains) == 0 {
		domains = []string{fmt.Sprintf("%s.%s", r.ServiceID, defaultIngressDomain)}
	}

	rules := make([]v1beta1.IngressRule, len(domains))
	for i, domain := range domains {
		rules[i] = v1beta1.IngressRule{
			Host: domain,
			IngressRuleValue: v1beta1.IngressRuleValue{HTTP: &v1beta1.HTTPIngressRuleValue{
				Paths: []v1beta1.HTTPIngressPath{{
					// It's important to have * after / since GCP GLBC load balancer doesn't support subresources automatically.
					Path: "/*",
					Backend: v1beta1.IngressBackend{
						ServiceName: r.ServiceID,
						ServicePort: r.ContainerPort,
					},
				}},
			}},
		}
	}

	//Note: making name dependent on namespace i.e. stackName will only provision one load-balancer per stack
	// change this if you want to allocate individual load balanacer for each app and use Name = payload.Request.ServiceID
	// Also if you change this remember to chage in AppGet to fetch IP of load balancer code
	name := fmt.Sprintf("%s-ing", payload.Request.Namespace)

	ing := &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Labels:      map[string]string{managedBy: heritage},
			Annotations: make(map[string]string),
		},
		Spec: v1beta1.IngressSpec{
			Rules: rules,
		},
		TypeMeta: metav1.TypeMeta{APIVersion: k8sBetaAPIVersion, Kind: "Ingress"},
	}

	if payload.Request.Provider == cloud.AwsProvider {
		ing.Annotations[ingressAnnotationName] = ingressClassName
	}

	return ing
}

func waitUntilDeploymentUpdated(c *kubernetes.Clientset, ns, name string) error {
	log.Debugf("waiting for Deployment %s to get a newer generation (30s timeout)", name)
	for i := 0; i < 30; i++ {
		dp, err := c.Extensions().Deployments(ns).Get(name, metav1.GetOptions{})
		if err != nil {
			if kerrors.IsNotFound(err) {
				time.Sleep(1 * time.Second)
				continue
			}
			return err
		}

		if dp.Status.ObservedGeneration >= dp.ObjectMeta.Generation {
			log.Debugf("A newer generation was found for Deployment %s", name)
			break
		}
		time.Sleep(1 * time.Second)
	}

	return nil
}

func waitUntilDeploymentReady(c *kubernetes.Clientset, ns, name string) error {
	dp, err := c.Extensions().Deployments(ns).Get(name, metav1.GetOptions{})
	if err != nil {
		return err
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

		ready, availablePods := areReplicaReady(c, ns, name, dp.ObjectMeta.ResourceVersion, labels)
		if ready {
			break
		}

		// check every 10 seconds for pod failures.
		// Depend on Deployment checks for ready pods
		if waited > 0 && (waited%10) == 0 {
			additionalTimeout, err := handlePendingPods(c, ns, labels)
			if err != nil {
				log.Errorf("checking pending pods %v", err)
				return err
			}

			timeout += additionalTimeout
			log.Infof("waited %ds and %d pods", waited, availablePods)
		}

		waited += 1
	}

	ready, _ := areReplicaReady(c, ns, name, dp.ObjectMeta.ResourceVersion, labels)
	if !ready {
		if err := handleNotReadyPods(c, ns, labels); err != nil {
			return err
		}
	}

	return nil
}

// Verify the status of a Deployment and if it is fully deployed
func areReplicaReady(c *kubernetes.Clientset, ns, name, resourceVersion string, labels map[string]string) (bool, int32) {
	dp, err := c.Extensions().Deployments(ns).Get(name, metav1.GetOptions{
		ResourceVersion: resourceVersion,
	})
	if err != nil {
		log.Error(err)
		return false, 0
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
