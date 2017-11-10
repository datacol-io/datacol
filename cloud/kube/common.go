package kube

import (
	"fmt"

	"k8s.io/client-go/kubernetes"

	log "github.com/Sirupsen/logrus"
	// kerrors "k8s.io/client-go/pkg/api/errors"
	// kapi "k8s.io/client-go/pkg/api/v1"
	// klabels "k8s.io/client-go/pkg/labels"
)

func DeleteApp(c *kubernetes.Clientset, ns, name string) error {
	if _, err := c.Core().Services(ns).Get(name); err != nil {
		if !kerrors.IsNotFound(err) {
			return err
		}
	} else if err := c.Core().Services(ns).Delete(name, &kapi.DeleteOptions{}); err != nil {
		return err
	}

	labels := klabels.Set(map[string]string{"name": name}).AsSelector()

	dp, err := c.Extensions().Deployments(ns).Get(name)
	if err != nil {
		if !kerrors.IsNotFound(err) {
			return err
		}
	}

	zerors := int32(0)
	dp.Spec.Replicas = &zerors

	if dp, err = c.Extensions().Deployments(ns).Update(dp); err != nil {
		return err
	}

	WaitUntilUpdated(c, ns, name)

	if err = c.Extensions().Deployments(ns).Delete(name, &kapi.DeleteOptions{}); err != nil {
		return err
	}

	// delete replicasets by label name=app
	res, err := c.Extensions().ReplicaSets(ns).List(kapi.ListOptions{LabelSelector: labels.String()})
	if err != nil {
		return err
	}

	for _, rs := range res.Items {
		if err := c.Extensions().ReplicaSets(ns).Delete(rs.Name, &kapi.DeleteOptions{}); err != nil {
			log.Warn(err)
		}
	}

	if _, err = c.Extensions().Ingresses(ns).Get(name); err != nil {
		if !kerrors.IsNotFound(err) {
			return err
		}
	} else if err = c.Extensions().Ingresses(ns).Delete(name, &kapi.DeleteOptions{}); err != nil {
		return err
	}

	return nil
}

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

	if svc.Spec.Type == kapi.ServiceTypeNodePort {
		ing, err := c.Extensions().Ingresses(ns).Get(name)
		if err != nil {
			if kerrors.IsNotFound(err) {
				return endpoint, nil
			}
			return endpoint, err
		}

		lBIngresses := ing.Status.LoadBalancer.Ingress
		if len(lBIngresses) > 0 {
			endpoint = lBIngresses[0].IP
		}

	}

	return endpoint, nil
}
