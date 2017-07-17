package kube

import (
	"fmt"

	"k8s.io/client-go/kubernetes"

	log "github.com/Sirupsen/logrus"
	kerrors "k8s.io/client-go/pkg/api/errors"
	kapi "k8s.io/client-go/pkg/api/v1"
)

func SetPodEnv(c *kubernetes.Clientset, ns, app string, env map[string]string) error {
	dp, err := c.Extensions().Deployments(ns).Get(app)
	if err != nil {
		return err
	}

	for i, c := range dp.Spec.Template.Spec.Containers {
		if c.Name == app {
			envVars := []kapi.EnvVar{}
			for key, value := range env {
				if len(key) > 0 {
					envVars = append(envVars, kapi.EnvVar{Name: key, Value: value})
				}
			}
			log.Debugf("setting env vars:\n %s", toJson(env))
			c.Env = envVars

			dp.Spec.Template.Spec.Containers[i] = c
		}
	}

	if _, err := c.Extensions().Deployments(ns).Update(dp); err != nil {
		return err
	}

	WaitUntilUpdated(c, ns, app)
	WaitUntilReady(c, ns, app)

	return nil
}

func GetServiceEndpoint(c *kubernetes.Clientset, ns, name string) (string, error) {
	var endpoint = ""

	svc, err := c.Core().Services(ns).Get(name)
	if err != nil {
		if kerrors.IsNotFound(err) {
			return endpoint, nil
		}
		return endpoint, err
	}

	if svc.Spec.Type == kapi.ServiceTypeLoadBalancer && len(svc.Status.LoadBalancer.Ingress) > 0 {
		ing := svc.Status.LoadBalancer.Ingress[0]
		if len(ing.Hostname) > 0 {
			endpoint = ing.Hostname
		} else {
			port := 80
			if len(svc.Spec.Ports) > 0 {
				port = int(svc.Spec.Ports[0].Port)
			}
			endpoint = fmt.Sprintf("%s:%d", ing.IP, port)
		}
	}

	return endpoint, nil
}
