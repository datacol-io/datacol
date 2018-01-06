package kube

import (
	"fmt"
	"sort"
	"time"

	"github.com/appscode/go/log"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	klabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

func ScalePodReplicas(c *kubernetes.Clientset, ns, name, image string, command []string, replicas int32) error {
	runner, _ := NewDeployer(c)

	req := &DeployRequest{
		ServiceID:   name,
		Environment: ns,
		Image:       image,
		Args:        command,
		Replicas:    replicas,
	}

	_, err := runner.Run(req)

	return err
}

func GetAllPods(c *kubernetes.Clientset, ns, app string) ([]v1.Pod, error) {
	tags := map[string]string{ServiceLabelKey: app}
	selector := klabels.Set(tags).AsSelector()
	res, err := c.Core().Pods(ns).List(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return nil, err
	}

	return res.Items, nil
}

func GetAllPodNames(c *kubernetes.Clientset, ns, app string) ([]string, error) {
	pods, err := GetAllPods(c, ns, app)
	if err != nil {
		return nil, err
	}

	var podNames []string
	for _, p := range pods {
		podNames = append(podNames, p.Name)
	}

	return podNames, nil
}

func getAllDeployments(c *kubernetes.Clientset, ns, app string) ([]v1beta1.Deployment, error) {
	tags := map[string]string{ServiceLabelKey: app}
	selector := klabels.Set(tags).AsSelector()
	res, err := c.Extensions().Deployments(ns).List(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return nil, err
	}

	return res.Items, nil
}

func getPodByName(c *kubernetes.Clientset, ns, app string) (*v1.Pod, error) {
	pods, err := GetAllPods(c, ns, app)
	if err != nil {
		return nil, err
	}

	if len(pods) < 1 {
		return nil, fmt.Errorf("No Pod found by name=%s", app)
	}

	return &pods[0], nil
}

func podEvents(c *kubernetes.Clientset, ns string, pod *v1.Pod) (*v1.EventList, error) {
	fields := map[string]string{
		"involvedObject.name":      pod.ObjectMeta.Name,
		"involvedObject.namespace": ns,
		"involvedObject.uid":       string(pod.ObjectMeta.UID),
	}

	res, err := c.Core().Events(ns).List(metav1.ListOptions{
		FieldSelector:   klabels.Set(fields).AsSelector().String(),
		ResourceVersion: pod.ObjectMeta.ResourceVersion,
	})
	if err != nil {
		return res, err
	}

	sort.Slice(res.Items, func(i, j int) bool {
		return res.Items[j].LastTimestamp.Before(&res.Items[i].LastTimestamp)
	})

	return res, err
}

func handleNotReadyPods(c *kubernetes.Clientset, ns string, labels map[string]string) error {
	selector := klabels.Set(labels).AsSelector()
	res, err := c.Core().Pods(ns).List(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return err
	}

	for _, pod := range res.Items {
		if pod.Status.Phase != v1.PodRunning {
			continue
		}

		name, ok := labels["name"]
		if !ok {
			return fmt.Errorf("name not found in %+v", labels)
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
			return err
		}

		for _, ev := range res.Items {
			if ev.Reason == "Unhealthy" || ev.Reason == "Failed" {
				return fmt.Errorf(ev.Message)
			}
		}
	}

	return nil
}

type podPhaseResponse struct {
	done  bool
	phase v1.PodPhase
	err   error
}

func waitUntilPodRunning(c *kubernetes.Clientset, ns, name string) error {
	// give pod 20 minutes to execute (after it got into ready state)
	pollAttempts := 10
	pollInterval := 1

	var status v1.PodPhase
	for i := 0; i <= pollAttempts; i++ {
		res := getPodPhase(c, ns, name)
		if !res.done {
			time.Sleep(time.Duration(pollInterval) * time.Second)
			continue
		}
		status = res.phase
	}

	if status != v1.PodRunning {
		return fmt.Errorf("pod failed to enter running state: %s", status)
	}

	return nil
}

func isRunning(pod *v1.Pod) (bool, error) {
	switch pod.Status.Phase {
	case v1.PodRunning:
		return true, nil
	case v1.PodSucceeded:
		return false, fmt.Errorf("pod already succeeded before it begins running")
	case v1.PodFailed:
		return false, fmt.Errorf("pod status is failed")
	default:
		return false, nil
	}
}

func getPodPhase(c *kubernetes.Clientset, ns, name string) podPhaseResponse {
	pod, err := c.Core().Pods(ns).Get(name, metav1.GetOptions{})
	if err != nil {
		return podPhaseResponse{true, v1.PodUnknown, err}
	}

	ready, err := isRunning(pod)
	if err != nil {
		return podPhaseResponse{true, pod.Status.Phase, err}
	}

	if ready {
		return podPhaseResponse{true, pod.Status.Phase, nil}
	}

	// check status of containers
	for _, container := range pod.Status.ContainerStatuses {
		if container.Ready {
			continue
		}
		if container.State.Waiting == nil {
			continue
		}

		switch container.State.Waiting.Reason {
		case "ErrImagePull", "ImagePullBackOff":
			err = fmt.Errorf("image pull failed: %s", container.State.Waiting.Message)
			return podPhaseResponse{true, v1.PodUnknown, err}
		}
	}

	fmt.Printf("Waiting for pod %s/%s to be running, status is %s\n", pod.Namespace, pod.Name, pod.Status.Phase)
	return podPhaseResponse{false, pod.Status.Phase, nil}
}
