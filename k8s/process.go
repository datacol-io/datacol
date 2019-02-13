package kube

import (
	"fmt"
	"io"
	"net/url"

	log "github.com/Sirupsen/logrus"
	"github.com/appscode/go/crypto/rand"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"

	pb "github.com/datacol-io/datacol/api/models"
	"github.com/datacol-io/datacol/cloud"
)

/*
	This code is inspired from gitlab-runner/executers/kubernetes and being customized to work wth Datacol use case.
*/

type RemoteExecutor interface {
	Execute(method string, url *url.URL, config *rest.Config, stdin io.Reader, stdout, stderr io.Writer, tty bool) error
}

// DefaultRemoteExecutor is the standard implementation of remote command execution
type DefaultRemoteExecutor struct{}

func (*DefaultRemoteExecutor) Execute(method string, url *url.URL, config *rest.Config, stdin io.Reader, stdout, stderr io.Writer, tty bool) error {
	exec, err := remotecommand.NewSPDYExecutor(config, method, url)
	if err != nil {
		return fmt.Errorf("failed to create SPDY executor: %v", err)
	}

	streamOptions := remotecommand.StreamOptions{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
		Tty:    tty,
	}

	return exec.Stream(streamOptions)
}

// ExecOptions declare the arguments accepted by the Exec command
type ExecOptions struct {
	Namespace     string
	PodName       string
	ContainerName string
	Stdin         bool
	Command       []string

	In  io.Reader
	Out io.Writer
	Err io.Writer

	Executor RemoteExecutor
	Client   *kubernetes.Clientset
	Config   *rest.Config
}

// Run executes a validated remote execution against a pod.
func (p *ExecOptions) Run() error {
	pod, err := p.Client.Core().Pods(p.Namespace).Get(p.PodName, metav1.GetOptions{})
	if err != nil {
		log.Errorf("fetching pod status: %v", err)
		return err
	}

	if pod.Status.Phase != corev1.PodRunning {
		return fmt.Errorf("Pod '%s' (on namespace '%s') is not running and cannot execute commands; current phase is '%s'",
			p.PodName, p.Namespace, pod.Status.Phase)
	}

	containerName := p.ContainerName
	if len(containerName) == 0 {
		log.Infof("defaulting container name to '%s'", pod.Spec.Containers[0].Name)
		containerName = pod.Spec.Containers[0].Name
	}

	var stdin io.Reader
	if p.Stdin {
		stdin = p.In
	}

	req := p.Client.Core().RESTClient().Post().
		Resource("pods").
		Name(pod.Name).
		Namespace(pod.Namespace).
		SubResource("exec").
		Param("container", containerName).
		VersionedParams(&corev1.PodExecOptions{
			Container: containerName,
			Command:   p.Command,
			Stdin:     stdin != nil,
			Stdout:    p.Out != nil,
			Stderr:    p.Err != nil,
			TTY:       true,
		}, scheme.ParameterCodec)

	if p.Executor == nil {
		p.Executor = &DefaultRemoteExecutor{}
	}

	return p.Executor.Execute("POST", req.URL(), p.Config, stdin, p.Out, p.Err, true)
}

// ProcessList will fetch the list of processes running based on deployments Labels
// TODO: fetch the status as well
func ProcessList(c *kubernetes.Clientset, ns, app string) ([]*pb.Process, error) {
	deployments, err := getAllDeployments(c, ns, app)
	if err != nil {
		return nil, err
	}
	var items []*pb.Process

	for _, dp := range deployments {
		pods, err := getLatestPodsForDeployment(c, &dp)
		if err != nil {
			return nil, err
		}

		log.Debugf("got %d pods for deployment:%s", len(pods), dp.Name)

		if len(pods) == 0 {
			continue
		}

		//FIXME: ideally should report status of all pods for latest release/version
		targetPod := pods[len(pods)-1]
		status := getPodStatusStr(c, &targetPod)
		proctype := dp.ObjectMeta.Labels[typeLabel]
		_, container := findContainer(dp.Spec.Template, fmt.Sprintf("%s-%s", app, proctype))

		items = append(items, &pb.Process{
			Proctype: proctype,
			Count:    *dp.Spec.Replicas,
			Status:   status,
			Command:  container.Args,
			Cpu:      getRequestLimit(container.Resources, corev1.ResourceCPU),
			Memory:   getRequestLimit(container.Resources, corev1.ResourceMemory),
		})
	}

	jobs, err := getAllCronJobs(c, ns, app)
	if err != nil {
		return nil, err
	}

	for _, cj := range jobs {
		pods, err := getLatestPodsForCronJob(c, &cj)
		if err != nil {
			return nil, err
		}

		log.Debugf("got %d pods for cronjob:%s", len(pods), cj.Name)

		if len(pods) == 0 {
			continue
		}
		targetPod := pods[len(pods)-1]
		status := getPodStatusStr(c, &targetPod)
		proctype := cj.ObjectMeta.Labels[typeLabel]
		_, container := findContainer(cj.Spec.JobTemplate.Spec.Template, fmt.Sprintf("%s-%s", app, proctype))

		proc := &pb.Process{
			Proctype: proctype,
			Status:   status,
			Command:  container.Args,
			Cpu:      getRequestLimit(container.Resources, corev1.ResourceCPU),
			Memory:   getRequestLimit(container.Resources, corev1.ResourceMemory),
			CronExpr: cj.Spec.Schedule,
		}

		if count := cj.Spec.JobTemplate.Spec.Parallelism; count != nil {
			proc.Count = *count
		}

		items = append(items, proc)
	}

	return items, nil
}

func ProcessRun(
	c *kubernetes.Clientset,
	cfg *rest.Config,
	ns, name, image string,
	options pb.ProcessRunOptions,
	envVars map[string]string,
	sqlproxy bool,
	stream io.ReadWriter,
	provider cloud.CloudProvider,
) error {
	proctype := runProcessKind
	podName := fmt.Sprintf("%s-%s-%s", name, proctype, rand.Characters(6))

	// Create a app Pod.  By running `sleep infinity`, the Pod will sit and do nothing.
	req := &DeployRequest{
		ServiceID:           podName,
		Image:               image,
		Args:                []string{"sleep", "infinity"}, //FIXME: To let the pod stay around until being told to exit
		EnvVars:             envVars,
		App:                 name,
		Proctype:            proctype,
		EnableCloudSqlProxy: sqlproxy,
		Provider:            provider,
	}

	// Delete the pod sunce it's ephemeral
	defer deletePodByName(c, ns, podName)

	return processRun(c, cfg, ns, options, req, stream)
}

func deletePodByName(c *kubernetes.Clientset, ns, name string) error {
	return c.Core().Pods(ns).Delete(name, &metav1.DeleteOptions{})
}

// Inspired from https://docs.okd.io/latest/go_client/executing_remote_processes.html
func processRun(c *kubernetes.Clientset, cfg *rest.Config, ns string, options pb.ProcessRunOptions, req *DeployRequest, stream io.ReadWriter) error {
	spec, err := newPodSpec(req)
	if err != nil {
		return fmt.Errorf("creating container manifest for process: %v", err)
	}

	var zero int64
	spec.Spec.RestartPolicy = corev1.RestartPolicyNever
	spec.Spec.TerminationGracePeriodSeconds = &zero
	spec.Spec.Containers[0].Stdin = options.Tty

	if options.Width > 0 {
		spec.Spec.Containers[0].Env = append(spec.Spec.Containers[0].Env, corev1.EnvVar{
			Name:  "COLUMNS",
			Value: fmt.Sprintf("%d", options.Width),
		})
	}

	if options.Height > 0 {
		spec.Spec.Containers[0].Env = append(spec.Spec.Containers[0].Env, corev1.EnvVar{
			Name:  "LINES",
			Value: fmt.Sprintf("%d", options.Height),
		})
	}

	log.Debugf("creating pod with spec %s", toJson(spec))
	pod, err := c.Core().Pods(ns).Create(spec)
	if err != nil {
		return err
	}

	if err = waitUntilPodRunning(c, ns, pod.ObjectMeta.Name); err != nil {
		return err
	}

	log.Debugf("Running command %v inside pod: %v", options.Entrypoint, pod.ObjectMeta.Name)

	executer := &ExecOptions{
		Namespace: ns,
		PodName:   req.ServiceID,
		Command:   options.Entrypoint,
		Stdin:     options.Tty,
		In:        stream,
		Out:       stream,
		Err:       stream,
		Client:    c,
		Config:    cfg,
	}

	return executer.Run()
}

func ProcessLimits(c *kubernetes.Clientset, ns, app, resource string, limits map[string]string) error {
	deployments, err := getAllDeployments(c, ns, app)
	if err != nil {
		return err
	}

	log.Debugf("setting %s limits %v", resource, limits)
	resourceName := corev1.ResourceName(resource)

	for _, dp := range deployments {
		for proctype, rl := range limits {
			if dp.ObjectMeta.Labels[typeLabel] == proctype {
				cName := fmt.Sprintf("%s-%s", app, proctype)

				if idx, container := findContainer(dp.Spec.Template, cName); idx >= 0 {
					if err := mergeResourceConstraints(resourceName, container, rl); err != nil {
						return err
					}

					dp.Spec.Template.Spec.Containers[idx].Resources = container.Resources
					if _, err := c.Extensions().Deployments(ns).Update(&dp); err != nil {
						return err
					}
				} else {
					log.Warnf("Didn't find container %s in %s deployment.", cName, dp.Name)
				}
			}
		}
	}

	return nil
}
