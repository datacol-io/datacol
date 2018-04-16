package kube

import (
	"errors"
	"fmt"
	"sort"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/datacol-io/datacol/cloud"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	klabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

//ScalePodReplicas can scale up/down pods as per replicas count
func ScalePodReplicas(c *kubernetes.Clientset, ns, app, proctype, image string, command []string,
	replicas int32, sqlproxy bool, envVars map[string]string, provider cloud.CloudProvider,
) error {
	runner, _ := NewDeployer(c)

	req := &DeployRequest{
		ServiceID:           fmt.Sprintf("%s-%s", app, proctype),
		Namespace:           ns,
		Image:               image,
		Args:                command,
		Replicas:            &replicas,
		App:                 app,
		Proctype:            proctype,
		EnableCloudSqlProxy: sqlproxy,
		EnvVars:             envVars,
		Provider:            provider,
	}

	_, err := runner.Run(req)

	return err
}

func GetAllPods(c *kubernetes.Clientset, ns, app string) ([]v1.Pod, error) {
	labels := map[string]string{appLabel: app, managedBy: heritage}
	selector := klabels.Set(labels).AsSelector()
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
	tags := map[string]string{appLabel: app, managedBy: heritage}
	selector := klabels.Set(tags).AsSelector()
	res, err := c.Extensions().Deployments(ns).List(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return nil, err
	}

	return res.Items, nil
}

func getPodsForDeployment(c *kubernetes.Clientset, dp *v1beta1.Deployment) ([]v1.Pod, error) {
	selector := klabels.Set(dp.Spec.Selector.MatchLabels).AsSelector()
	res, err := c.Core().Pods(dp.Namespace).List(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return nil, err
	}

	return res.Items, nil
}

func getLatestPodsForDeployment(c *kubernetes.Clientset, dp *v1beta1.Deployment) ([]v1.Pod, error) {
	selector := klabels.Set(dp.Spec.Selector.MatchLabels).AsSelector()
	res, err := c.Core().Pods(dp.Namespace).List(metav1.ListOptions{
		LabelSelector:   selector.String(),
		ResourceVersion: dp.ResourceVersion,
	})
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

func podEvents(c *kubernetes.Clientset, pod *v1.Pod) (*v1.EventList, error) {
	fields := map[string]string{
		"involvedObject.name":      pod.ObjectMeta.Name,
		"involvedObject.namespace": pod.Namespace,
		"involvedObject.uid":       string(pod.ObjectMeta.UID),
	}

	res, err := c.Core().Events(pod.Namespace).List(metav1.ListOptions{
		FieldSelector: klabels.Set(fields).AsSelector().String(),
		// ResourceVersion: pod.ObjectMeta.ResourceVersion,
	})
	if err != nil {
		return res, err
	}

	sort.Slice(res.Items, func(i, j int) bool {
		return res.Items[j].LastTimestamp.Before(&res.Items[i].LastTimestamp)
	})

	return res, err
}

//Read https://github.com/kubernetes/kubernetes/issues/1899
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

		name := fmt.Sprintf("%s-%s", labels["app"], labels["type"])
		cstatus := v1.ContainerStatus{}
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.Name == name {
				cstatus = cs
				break
			}
		}

		log.Debugf("pod %s status: %s", pod.Name, toJson(cstatus))
		if cstatus.Ready {
			continue
		}

		res, err := podEvents(c, &pod)
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

// Return timeout if Pod is fetching the image from registry. I am returning 0 always since there is not suppor for it atm.
func handlePendingPods(c *kubernetes.Clientset, namespace string, labels map[string]string) (int, error) {
	timeout := 0
	resp, err := c.Core().Pods(namespace).List(metav1.ListOptions{
		LabelSelector: klabels.Set(labels).AsSelector().String(),
	})

	if err != nil {
		return timeout, err
	}

	log.Infof("Number of pods %d for %v", len(resp.Items), labels)

	for _, item := range resp.Items {
		log.Debugf("pod %s phase=%s", item.Name, item.Status.Phase)

		if item.Status.Phase != v1.PodPending && item.Status.Phase != v1.PodRunning {
			continue
		}
		reason, message := podPendingStatus(c, &item)

		if err := podErrors(c, &item, reason, message); err != nil {
			return timeout, err
		}
	}

	return timeout, nil
}

// Handle potential pod errors based on the Pending
// reason passed into the function
func podErrors(c *kubernetes.Clientset, pod *v1.Pod, reason, message string) error {
	containerErrors := []string{"CrashLoopBackOff", "ErrImagePull"}
	for _, e := range containerErrors {
		if e == reason {
			return errors.New(message)
		}
	}

	events, err := podEvents(c, pod)
	if err != nil {
		return err
	}

	eventErrors := []string{"Failed", "InspectFailed", "ErrImageNeverPull", "FailedScheduling"}
	for _, event := range events.Items {
		for _, e := range eventErrors {
			if e == event.Reason {
				return errors.New(event.Message)
			}
		}
	}

	return nil
}

// Introspect the pod containers when pod is in Pending state
func podPendingStatus(c *kubernetes.Clientset, pod *v1.Pod) (reason string, message string) {
	reason = "Pending"
	statuses := pod.Status.ContainerStatuses

	name := containerNameInPod(pod)

	for _, cs := range statuses {
		if cs.Name == name {
			if waiting := cs.State.Waiting; waiting != nil {
				reason = waiting.Reason
				message = waiting.Message

				if reason == "ContainerCreating" {
					if events, err := podEvents(c, pod); err == nil {
						if ev := events.Items; len(ev) > 0 {
							reason = ev[len(ev)-1].Reason
							message = ev[len(ev)-1].Message
							return
						}
					}
				}
			}
		}
	}

	return
}

type podPhaseResponse struct {
	done  bool
	phase v1.PodPhase
	err   error
}

func waitUntilPodRunning(c *kubernetes.Clientset, ns, name string) error {
	// give pod 20 minutes to execute (after it got into ready state)
	pollAttempts := 20
	pollInterval := 1

	var status v1.PodPhase
	for i := 0; i <= pollAttempts; i++ {
		res := getPodPhase(c, ns, name)

		if !res.done {
			time.Sleep(time.Duration(pollInterval) * time.Second)
			continue
		}
		if res.err != nil {
			return res.err
		}

		status = res.phase

		if status == v1.PodRunning {
			break
		}
	}

	if status != v1.PodRunning {
		return fmt.Errorf("pod %s failed to enter running state (%s)", name, status)
	}

	return nil
}

func waitUntilPodCreated(c *kubernetes.Clientset, ns, name string) error {
	// give pod 20 minutes to execute (after it got into ready state)
	pollAttempts := 20
	pollInterval := 1

	var status v1.PodPhase
	for i := 0; i <= pollAttempts; i++ {
		res := getPodPhase(c, ns, name)
		log.Debugf("checking pod %s status: %+v", name, res)

		if !res.done {
			time.Sleep(time.Duration(pollInterval) * time.Second)
			continue
		}
		if res.err != nil {
			return res.err
		}

		status = res.phase

		if status == v1.PodRunning {
			break
		}
	}

	if status != v1.PodRunning {
		return fmt.Errorf("pod failed to enter running state: %s", status)
	}

	return nil
}

func getPodPhase(c *kubernetes.Clientset, ns, name string) podPhaseResponse {
	pod, err := c.Core().Pods(ns).Get(name, metav1.GetOptions{})
	if err != nil {
		return podPhaseResponse{true, v1.PodUnknown, err}
	}

	running, err := isRunning(pod)
	if err != nil {
		return podPhaseResponse{true, pod.Status.Phase, err}
	}

	if running {
		return podPhaseResponse{true, pod.Status.Phase, nil}
	}

	// check status of containers
	for _, container := range pod.Status.ContainerStatuses {
		if container.Ready || container.State.Waiting == nil {
			continue
		}

		switch container.State.Waiting.Reason {
		case "ErrImagePull", "ImagePullBackOff", "InvalidImageName":
			err = fmt.Errorf("image pull failed: %s", container.State.Waiting.Message)
			return podPhaseResponse{true, v1.PodUnknown, err}
		}
	}

	fmt.Printf("Waiting for pod %s/%s to be running, status is %s\n", pod.Namespace, pod.Name, pod.Status.Phase)
	return podPhaseResponse{false, pod.Status.Phase, nil}
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

func containerNameInPod(pod *v1.Pod) string {
	return fmt.Sprintf("%s-%s", pod.ObjectMeta.Labels[appLabel], pod.ObjectMeta.Labels[typeLabel])
}

type podStatus int

const (
	podInitializing podStatus = iota + 1
	podCreating
	podStarting
	podUp
	podTerminating
	podDown
	podCrashed
	podError
	podDestroyed
)

var podStatesMap = map[string]podStatus{
	"Pending":           podInitializing,
	"ContainerCreating": podCreating,
	"Starting":          podStarting,
	"Running":           podUp,
	"Terminating":       podTerminating,
	"Succeeded":         podDown,
	"Failed":            podCrashed,
	"Unknown":           podError,
}

func getPodStatusStr(c *kubernetes.Clientset, pod *v1.Pod) string {
	status := getPodStatus(c, pod)
	for key, val := range podStatesMap {
		if val == status {
			return key
		}
	}

	return ""
}

func getPodStatus(c *kubernetes.Clientset, pod *v1.Pod) podStatus {
	if pod == nil {
		return podDestroyed
	}

	var status string
	if pod.Status.Phase == v1.PodPending {
		status, _ = podPendingStatus(c, pod)
	} else if pod.Status.Phase == v1.PodRunning {
		status = podReadinessStatus(pod)
		if status == "Starting" || status == "Terminating" {
			return podStatesMap[status]
		} else if status == "Running" && podLivenessStatus(pod) {
			return podStatesMap[status]
		}
	} else {
		status = string(pod.Status.Phase)
	}

	if v, ok := podStatesMap[status]; ok {
		return v
	} else {
		return podError
	}
}

func podReadinessStatus(pod *v1.Pod) string {
	status := "Unknown"
	name := containerNameInPod(pod)

	for _, cs := range pod.Status.ContainerStatuses {
		if cs.Name == name {
			if cs.Ready {
				if pod.ObjectMeta.DeletionTimestamp != nil {
					return "Terminating"
				}
				return "Running"
			}

			if cs.State.Running != nil {
				return "Starting"
			}

			if cs.State.Terminated != nil && pod.ObjectMeta.DeletionTimestamp != nil {
				return "Terminating"
			}
		}
	}

	return status
}

func podLivenessStatus(pod *v1.Pod) bool {
	for _, cond := range pod.Status.Conditions {
		if cond.Type == v1.PodReady && cond.Status != v1.ConditionTrue {
			return false
		}
	}

	return true
}
