package aws

import (
	"fmt"
	pb "github.com/dinesh/datacol/api/models"
	"io"
	"k8s.io/client-go/kubernetes"
	kapi "k8s.io/client-go/pkg/api/v1"
	klabels "k8s.io/client-go/pkg/labels"
)

func (a *AwsCloud) EnvironmentGet(name string) (pb.Environment, error) {
	return nil, nil
}

func (a *AwsCloud) EnvironmentSet(name string, body io.Reader) error {
	return nil
}

func (a *AwsCloud) GetRunningPods(app string) (string, error) {
	ns := a.DeploymentName
	c, err := getKubeClientset(ns)
	if err != nil {
		return "", err
	}

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
