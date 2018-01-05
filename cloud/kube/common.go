package kube

import (
	"fmt"
	"io"
	"math"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	pb "github.com/dinesh/datacol/api/models"
	core_v1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	klabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

const (
	ServiceLabelKey = "datacol.io/service"
)

func DeleteService(c *kubernetes.Clientset, ns, name string) error {
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

	waitUntilDeploymentUpdated(c, ns, name)

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
	deployments, err := getAllDeployments(c, ns, app)
	if err != nil {
		return err
	}

	var containersToWatch []string

	for _, dp := range deployments {
		//TODO: should have better logic for selecting the container name
		container := dp.Spec.Template.Spec.Containers[0]
		containerName := container.Name

		envVars := []core_v1.EnvVar{}
		for key, value := range env {
			if len(key) > 0 {
				envVars = append(envVars, core_v1.EnvVar{Name: key, Value: value})
			}
		}
		log.Debugf("setting env vars=\n %s in container=%s", toJson(env), containerName)
		container.Env = envVars

		dp.Spec.Template.Spec.Containers[0] = container

		if _, err := c.Extensions().Deployments(ns).Update(&dp); err != nil {
			return err
		}

		containersToWatch = append(containersToWatch, containerName)
	}

	for _, name := range containersToWatch {
		go func(cname string) {
			waitUntilDeploymentUpdated(c, ns, cname)
			waitUntilDeploymentReady(c, ns, cname)
		}(name)
	}

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

func LogStreamReq(c *kubernetes.Clientset, w io.Writer, ns, app string, opts pb.LogStreamOptions) error {
	podNames, err := GetAllPodNames(c, ns, app)

	//TODO: consider using https://github.com/djherbis/stream for reading multiple streams
	var readers []io.Reader

	log.Debugf("streaming logs from %v", podNames)

	for _, name := range podNames {
		req := c.Core().RESTClient().Get().
			Namespace(ns).
			Name(name).
			Resource("pods").
			SubResource("log").
			Param("follow", strconv.FormatBool(opts.Follow))

		if opts.Since > 0 {
			sec := int64(math.Ceil(float64(opts.Since) / float64(time.Second)))
			req = req.Param("sinceSeconds", strconv.FormatInt(sec, 10))
		}

		if r, err := req.Stream(); err == nil {
			readers = append(readers, r)
		}
	}

	_, err = io.Copy(w, io.MultiReader(readers...))
	return err
}
