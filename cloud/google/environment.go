package google

import (
  "io"
  "fmt"
  "k8s.io/client-go/kubernetes"
  _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

  kapi "k8s.io/client-go/pkg/api/v1"
  klabels "k8s.io/client-go/pkg/labels"

  pb "github.com/dinesh/datacol/api/models"
  "google.golang.org/api/googleapi"
)

func (g *GCPCloud) EnvironmentGet(name string) (pb.Environment, error) {
  gskey := fmt.Sprintf("%s.env", name)
  data, err := g.gsGet(g.BucketName, gskey)
  if err != nil {
    if gerr, ok := err.(*googleapi.Error); ok && gerr.Code == 404 {
      return pb.Environment{}, nil
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
