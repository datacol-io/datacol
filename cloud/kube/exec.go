package kube

import (
	"fmt"
	"io"
	"net/url"

	log "github.com/Sirupsen/logrus"
	"k8s.io/kubernetes/pkg/api"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	"k8s.io/kubernetes/pkg/client/restclient"

	"k8s.io/kubernetes/pkg/client/unversioned/remotecommand"
	remotecommandserver "k8s.io/kubernetes/pkg/kubelet/server/remotecommand"
)

/*
	This code is inspired from gitlab-runner/executers/kubernetes and being customized to work wth Datacol use case.
*/

type RemoteExecutor interface {
	Execute(method string, url *url.URL, config *restclient.Config, stdin io.Reader, stdout, stderr io.Writer, tty bool) error
}

// DefaultRemoteExecutor is the standard implementation of remote command execution
type DefaultRemoteExecutor struct{}

func (*DefaultRemoteExecutor) Execute(method string, url *url.URL, config *restclient.Config, stdin io.Reader, stdout, stderr io.Writer, tty bool) error {
	exec, err := remotecommand.NewExecutor(config, method, url)
	if err != nil {
		return err
	}
	return exec.Stream(remotecommand.StreamOptions{
		SupportedProtocols: remotecommandserver.SupportedStreamingProtocols,
		Stdin:              stdin,
		Stdout:             stdout,
		Stderr:             stderr,
		Tty:                tty,
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
	Client   *clientset.Clientset
	Config   *restclient.Config
}

// Run executes a validated remote execution against a pod.
func (p *ExecOptions) Run() error {
	pod, err := p.Client.Core().Pods(p.Namespace).Get(p.PodName)
	if err != nil {
		return err
	}

	if pod.Status.Phase != api.PodRunning {
		return fmt.Errorf("Pod '%s' (on namespace '%s') is not running and cannot execute commands; current phase is '%s'",
			p.PodName, p.Namespace, pod.Status.Phase)
	}

	containerName := p.ContainerName
	if len(containerName) == 0 {
		log.Infof("defaulting container name to '%s'", pod.Spec.Containers[0].Name)
		containerName = pod.Spec.Containers[0].Name
	}

	// TODO: refactor with terminal helpers from the edit utility once that is merged
	var stdin io.Reader
	if p.Stdin {
		stdin = p.In
	}

	// TODO: consider abstracting into a client invocation or client helper
	req := p.Client.Core().RESTClient().Post().
		Resource("pods").
		Name(pod.Name).
		Namespace(pod.Namespace).
		SubResource("exec").
		Param("container", containerName)
	req.VersionedParams(&api.PodExecOptions{
		Container: containerName,
		Command:   p.Command,
		Stdin:     stdin != nil,
		Stdout:    p.Out != nil,
		Stderr:    p.Err != nil,
	}, api.ParameterCodec)

	return p.Executor.Execute("POST", req.URL(), p.Config, stdin, p.Out, p.Err, false)
}
