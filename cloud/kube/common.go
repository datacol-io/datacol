package kube

import (
	"fmt"

	"k8s.io/client-go/kubernetes"

	log "github.com/Sirupsen/logrus"
	core_v1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	klabels "k8s.io/apimachinery/pkg/labels"
)

func DeleteApp(c *kubernetes.Clientset, ns, name string) error {
	if _, err := c.Core().Services(ns).Get(name, meta_v1.GetOptions{}); err != nil {
		if !kerrors.IsNotFound(err) {
			return err
		}
	} else if err := c.Core().Services(ns).Delete(name, &meta_v1.DeleteOptions{}); err != nil {
		return err
	}

	labels := klabels.Set(map[string]string{"name": name}).AsSelector()

	dp, err := c.Extensions().Deployments(ns).Get(name, meta_v1.GetOptions{})
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

	if err = c.Extensions().Deployments(ns).Delete(name, &meta_v1.DeleteOptions{}); err != nil {
		return err
	}

	// delete replicasets by label name=app
	res, err := c.Extensions().ReplicaSets(ns).List(meta_v1.ListOptions{LabelSelector: labels.String()})
	if err != nil {
		return err
	}

	for _, rs := range res.Items {
		if err := c.Extensions().ReplicaSets(ns).Delete(rs.Name, &meta_v1.DeleteOptions{}); err != nil {
			log.Warn(err)
		}
	}

	if _, err = c.Extensions().Ingresses(ns).Get(name, meta_v1.GetOptions{}); err != nil {
		if !kerrors.IsNotFound(err) {
			return err
		}
	} else if err = c.Extensions().Ingresses(ns).Delete(name, &meta_v1.DeleteOptions{}); err != nil {
		return err
	}

	return nil
}

func SetPodEnv(c *kubernetes.Clientset, ns, app string, env map[string]string) error {
	dp, err := c.Extensions().Deployments(ns).Get(app, meta_v1.GetOptions{})
	if err != nil {
		return err
	}

	for i, c := range dp.Spec.Template.Spec.Containers {
		if c.Name == app {
			envVars := []core_v1.EnvVar{}
			for key, value := range env {
				if len(key) > 0 {
					envVars = append(envVars, core_v1.EnvVar{Name: key, Value: value})
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

	svc, err := c.Core().Services(ns).Get(name, meta_v1.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return endpoint, nil
		}
		return endpoint, err
	}

	if svc.Spec.Type == core_v1.ServiceTypeLoadBalancer && len(svc.Status.LoadBalancer.Ingress) > 0 {
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

	if svc.Spec.Type == core_v1.ServiceTypeNodePort {
		ing, err := c.Extensions().Ingresses(ns).Get(name, meta_v1.GetOptions{})
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
