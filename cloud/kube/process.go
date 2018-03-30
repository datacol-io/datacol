package kube

import (
	"fmt"
	"io"
	"net/url"

	log "github.com/Sirupsen/logrus"
	"github.com/appscode/go/crypto/rand"
	pb "github.com/datacol-io/datacol/api/models"
	"github.com/datacol-io/datacol/cloud"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
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

	return exec.Stream(remotecommand.StreamOptions{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
		Tty:    tty,
	})
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
		Param("container", containerName)

	req.VersionedParams(&corev1.PodExecOptions{
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
		pods, err := getPodsForDeployment(c, &dp)
		if err != nil {
			return nil, err
		}

		var status string
		if len(pods) > 0 {
			//FIXME: ideally should report status of all pods
			status = getPodStatusStr(c, &pods[len(pods)-1])
		}

		items = append(items, &pb.Process{
			Proctype: dp.ObjectMeta.Labels[typeLabel],
			Workers:  *dp.Spec.Replicas,
			Name:     dp.Name,
			Status:   status,
		})
	}

	return items, nil
}

func ProcessRun(
	c *kubernetes.Clientset,
	cfg *rest.Config,
	ns, name, image string,
	command []string,
	envVars map[string]string,
	sqlproxy bool,
	stream io.ReadWriter,
	provider cloud.CloudProvider,
) error {
	proctype := rand.Characters(6)
	podName := fmt.Sprintf("%s-%s", name, proctype)

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

	return processRun(c, cfg, ns, command, req, stream)
}

func deletePodByName(c *kubernetes.Clientset, ns, name string) error {
	return c.Core().Pods(ns).Delete(name, &metav1.DeleteOptions{})
}

func processRun(c *kubernetes.Clientset, cfg *rest.Config, ns string, command []string, req *DeployRequest, stream io.ReadWriter) error {
	spec := newPodSpec(req)
	spec.Spec.RestartPolicy = corev1.RestartPolicyNever

	log.Debugf("creating pod with spec %s", toJson(spec))
	pod, err := c.Core().Pods(ns).Create(spec)
	if err != nil {
		return err
	}

	if err = waitUntilPodRunning(c, ns, pod.ObjectMeta.Name); err != nil {
		return err
	}

	log.Debugf("Running command %v inside pod: %v", command, pod.ObjectMeta.Name)

	executer := &ExecOptions{
		Namespace: ns,
		PodName:   req.ServiceID,
		Command:   command,
		Stdin:     true,
		In:        stream,
		Out:       stream,
		Err:       stream,
		Client:    c,
		Config:    cfg,
	}

	return executer.Run()
}
