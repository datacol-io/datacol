package main

import (
	"bytes"
	log "github.com/Sirupsen/logrus"
	pbs "github.com/dinesh/datacol/api/controller"
	"golang.org/x/net/context"
	"os/exec"
	"syscall"
)

func (s *Server) Kubectl(ctx context.Context, req *pbs.KubectlReq) (*pbs.CmdResponse, error) {
	// todo: remove this call as we are caching the kubeconfig on first run.
	cfg, err := s.Provider.K8sConfigPath()
	if err != nil {
		return nil, internalError(err, "failed to fetch k8s config")
	}

	args := append([]string{"--kubeconfig", cfg, "-n", s.StackName}, req.Args...)
	out := cmd_execute("kubectl", args)
	return makeResponse(out), nil
}

func (s *Server) ProcessRun(ctx context.Context, req *pbs.ProcessRunReq) (*pbs.CmdResponse, error) {
	cfg, err := s.Provider.K8sConfigPath()
	if err != nil {
		return nil, internalError(err, "failed to fetch k8s config")
	}

	pod, err := s.Provider.GetRunningPods(req.Name)
	if err != nil {
		return nil, err
	}

	args := append([]string{"--kubeconfig", cfg, "-n", s.StackName, "--pod", pod, "exec"}, req.Command...)
	out := cmd_execute("kubectl", args)
	return makeResponse(out), nil
}

func makeResponse(out *Output) *pbs.CmdResponse {
	result := &pbs.CmdResponse{
		ExitCode: int32(out.ExitCode),
		StdErr:   out.StdErr,
		StdOut:   out.StdOut,
	}

	if out.Err != nil {
		result.Err = out.Err.Error()
	}

	return result
}

type Output struct {
	ExitCode int    `json:"exitCode"`
	StdOut   string `json:"stdOut"`
	StdErr   string `json:"stdErr"`
	Err      error  `json:"error"`
}

func cmd_execute(cmd string, args []string) *Output {
	var outStream bytes.Buffer
	var errStream bytes.Buffer

	log.Debugf("running %s %v", cmd, args)

	c := exec.Command(cmd, args...)
	c.Stdin = nil
	c.Stdout = &outStream
	c.Stderr = &errStream

	output := Output{}
	output.Err = c.Run()
	output.StdOut = string(outStream.Bytes())
	output.StdErr = string(errStream.Bytes())

	if exitErr, ok := output.Err.(*exec.ExitError); ok {
		if waitStatus, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			output.ExitCode = waitStatus.ExitStatus()
		}
	}

	log.Debugf("out: %s", toJson(output))

	return &output
}
