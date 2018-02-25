package kube

import (
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bjaglin/multiplexio"
	pb "github.com/dinesh/datacol/api/models"
	"github.com/dinesh/datacol/cloud"
	core_v1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	klabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

func DeleteApp(c *kubernetes.Clientset, ns, app string, provider cloud.CloudProvider) error {
	listSelector := meta_v1.ListOptions{
		LabelSelector: klabels.Set(map[string]string{appLabel: app, managedBy: heritage}).AsSelector().String(),
	}

	ret, err := c.Core().Services(ns).List(listSelector)
	if err != nil {
		return err
	}

	for _, svc := range ret.Items {
		if err := c.Core().Services(ns).Delete(svc.Name, &meta_v1.DeleteOptions{}); err != nil {
			return err
		}
	}

	dps, err := c.Extensions().Deployments(ns).List(listSelector)

	for _, dp := range dps.Items {
		zerors := int32(0)
		dp.Spec.Replicas = &zerors

		if _, err = c.Extensions().Deployments(ns).Update(&dp); err != nil {
			return err
		}

		waitUntilDeploymentUpdated(c, ns, dp.Name)

		if err = c.Extensions().Deployments(ns).Delete(dp.Name, &meta_v1.DeleteOptions{}); err != nil {
			return err
		}

		// delete replicasets by label app=<appname>
		res, err := c.Extensions().ReplicaSets(ns).List(listSelector)
		if err != nil {
			return err
		}

		for _, rs := range res.Items {
			if err := c.Extensions().ReplicaSets(ns).Delete(rs.Name, &meta_v1.DeleteOptions{}); err != nil {
				log.Warn(err)
			}
		}
	}

	commonListSelector := meta_v1.ListOptions{
		LabelSelector: klabels.Set(map[string]string{managedBy: heritage}).String(),
	}
	resp, err := c.Core().Services(ns).List(commonListSelector)
	if err != nil {
		return err
	}

	if len(resp.Items) == 0 {
		// Delete the ingress common to this namespace
		resp, err := c.Extensions().Ingresses(ns).List(commonListSelector)
		if err != nil {
			return err
		}

		for _, ing := range resp.Items {
			log.Debugf("deleting the ingress %s", ing.Name)

			if err := c.Extensions().Ingresses(ns).Delete(ing.Name, &meta_v1.DeleteOptions{}); err != nil {
				log.Warn(err)
			}
		}

		if provider == cloud.AwsProvider {
			awsing := &awsIngress{Client: c, Namespace: ingressDefaultNamespace, ParentNamespace: ns}
			awsing.Remove()
		}
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

	log.Infof("restarted containers: %v", containersToWatch)

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

	log.Debugf("service %s", toJson(svc))

	// If a service is deployed without domainName. We use ServceType = LoadBalancer and cloud load balancer will expose the service
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

	// If a service is of type NodePort.
	if svc.Spec.Type == core_v1.ServiceTypeNodePort {
		ingName := fmt.Sprintf("%s-ing", ns)
		ing, err := c.Extensions().Ingresses(ns).Get(ingName, meta_v1.GetOptions{})
		if err != nil {
			if kerrors.IsNotFound(err) {
				log.Error(err)
				return endpoint, nil
			}
			return endpoint, err
		}

		log.Debugf("ingress %s", toJson(ing))

		if lBIngresses := ing.Status.LoadBalancer.Ingress; len(lBIngresses) > 0 {
			return lBIngresses[0].IP, nil
		}

		if _, ok := ing.Annotations[ingressAnnotationName]; ok {
			svc, err := c.Core().Services(ingressNamespace(ns)).Get(nginxAppName, meta_v1.GetOptions{})
			if err != nil {
				return endpoint, fmt.Errorf("No Load Balancer found for %s. err: %v", name, err)
			}

			log.Debugf("Got ingress service in %s: %s", ingressNamespace, toJson(svc))

			ings := svc.Status.LoadBalancer.Ingress
			if len(ings) > 0 {
				return ings[0].Hostname, nil
			}
			return endpoint, fmt.Errorf("Load balancer don't have any IP yet.")
		}
	}

	log.Warnf("no load balancer IP found for app: %s", name)

	return endpoint, nil
}

func LogStreamReq(c *kubernetes.Clientset, w io.Writer, ns, app string, opts pb.LogStreamOptions) error {
	pods, err := GetAllPods(c, ns, app)

	//TODO: consider using https://github.com/djherbis/stream for reading multiple streams
	var sources []multiplexio.Source

	for _, pod := range pods {
		name := pod.Name
		log.Debugf("streaming logs from %v", name)

		req := c.Core().RESTClient().Get().
			Namespace(ns).
			Name(name).
			Resource("pods").
			SubResource("log").
			Param("follow", strconv.FormatBool(opts.Follow))

		if len(pod.Spec.Containers) > 0 {
			cntName := pod.Spec.Containers[0].Name
			log.Debugf("defaulting to container: %s since we have multiple", cntName)
			req = req.Param("container", cntName)
		}

		if opts.Since > 0 {
			sec := int64(math.Ceil(float64(opts.Since) / float64(time.Second)))
			req = req.Param("sinceSeconds", strconv.FormatInt(sec, 10))
		}

		if r, err := req.Stream(); err == nil {
			prefix := fmt.Sprintf("[%s] ", strings.TrimPrefix(name, app+"-"))
			sources = append(sources, multiplexio.Source{
				Reader: r,
				Write: func(dest io.Writer, token []byte) (int, error) {
					return io.WriteString(dest, prefix+string(token)+"\n")
				},
			})
		}
	}

	_, err = io.Copy(w, multiplexio.NewReader(multiplexio.Options{}, sources...))
	return err
}

func PruneCloudSQLManifest(spec *core_v1.PodSpec) {
	containers := spec.Containers

	for i, ctnr := range spec.Containers {
		if ctnr.Name == cloud.CloudsqlContainerName {
			spec.Containers = append(containers[:i], containers[i+1:]...)
			break
		}
	}

	vls := []core_v1.Volume{}
	for _, v := range spec.Volumes {
		if v.Name != cloud.CloudsqlCredVolName && v.Name != "cloudsql" {
			vls = append(vls, v)
		}
	}

	spec.Volumes = vls
}

func MergeCloudSQLManifest(spec *core_v1.PodSpec, app string, env map[string]string) {
	secretName := fmt.Sprintf("%s-%s", app, cloud.CloudsqlSecretName)

	containers := spec.Containers
	var port int

	if dburl, ok := env["DATABASE_URL"]; ok {
		parts := strings.Split(dburl, "://")
		port = getDefaultPort(parts[0])
	} else {
		log.Errorf("Skipping Cloudsql sidekar manifest since DATABASE_URL is not present in %v", env)
		return
	}

	sqlContainer := core_v1.Container{
		Command: []string{"/cloud_sql_proxy", "--dir=/cloudsql",
			fmt.Sprintf("-instances=%s=tcp:%d", env["INSTANCE_NAME"], port),
			"-credential_file=/secrets/cloudsql/credentials.json"},
		Name:            cloud.CloudsqlContainerName,
		Image:           cloud.CloudsqlImage,
		ImagePullPolicy: "IfNotPresent",
		VolumeMounts: []core_v1.VolumeMount{
			{
				Name:      cloud.CloudsqlCredVolName,
				MountPath: "/secrets/cloudsql",
				ReadOnly:  true,
			},
			{
				Name:      "cloudsql",
				MountPath: "/cloudsql",
			},
		},
	}

	found := false
	for i, c := range containers {
		if c.Name == cloud.CloudsqlContainerName {
			containers[i] = sqlContainer
			found = true
		}
	}

	if !found {
		containers = append(containers, sqlContainer)
	}

	spec.Containers = containers

	if !found {
		volfound := false
		dpvolumes := spec.Volumes

		for _, v := range dpvolumes {
			if v.Name == cloud.CloudsqlCredVolName {
				volfound = true
			}
		}

		if !volfound {
			volumes := []core_v1.Volume{
				{
					Name: cloud.CloudsqlCredVolName,
					VolumeSource: core_v1.VolumeSource{
						Secret: &core_v1.SecretVolumeSource{
							SecretName: secretName,
						},
					},
				},
				{
					Name: "cloudsql",
					VolumeSource: core_v1.VolumeSource{
						EmptyDir: &core_v1.EmptyDirVolumeSource{},
					},
				},
			}

			spec.Volumes = append(dpvolumes, volumes...)
		}
	}
}

func getDefaultPort(kind string) int {
	var port int
	switch kind {
	case "mysql":
		port = 3306
	case "postgres":
		port = 5432
	default:
		log.Fatal(fmt.Errorf("No default port defined for kind=%s", kind))
	}

	return port
}
