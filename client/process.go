package client

import (
	"os"
	"strconv"
	"strings"

	"github.com/appscode/go/term"
	pbs "github.com/datacol-io/datacol/api/controller"
	pb "github.com/datacol-io/datacol/api/models"
	"github.com/golang/protobuf/ptypes/empty"
)

func (c *Client) ListProcess(name string) ([]*pb.Process, error) {
	resp, err := c.ProviderServiceClient.ProcessList(ctx, &pbs.AppRequest{
		Name: name,
	})
	if err != nil {
		return nil, err
	}

	return resp.Items, nil
}

func (c *Client) SaveProcess(name string, options map[string]string) error {
	formation := pb.Formation{
		App:       name,
		Structure: make(map[string]int32),
	}

	for key, value := range options {
		num, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		formation.Structure[key] = int32(num)
	}

	term.Printf("scaling processs %v ...", formation.Structure)

	_, err := c.ProviderServiceClient.ProcessSave(ctx, &formation)
	return err
}

func (c *Client) UpdateProcessLimits(name, resource string, limits map[string]string) error {
	rl := pb.ResourceLimits{
		App:      name,
		Proctype: resource,
		Limits:   limits,
	}

	term.Printf("setting %s limits %v in %s ...", resource, rl.Limits, name)

	_, err := c.ProviderServiceClient.ProcessLimits(ctx, &rl)
	return err
}

func (c *Client) RunProcess(name string, args []string, opts pb.ProcessRunOptions) error {
	headers := map[string]string{
		"app":     name,
		"command": strings.Join(args, "#"),
		"tty":     strconv.FormatBool(opts.Tty),
		"detach":  strconv.FormatBool(opts.Detach),
		"width":   strconv.Itoa(opts.Width),
		"height":  strconv.Itoa(opts.Height),
	}

	if opts.Detach {
		resp, err := c.ProviderServiceClient.ProcessRun(ctx, &pbs.ProcessRunReq{
			Name:    name,
			Command: args,
		})
		if err != nil {
			return err
		}
		term.Printf("Started %s ...\n", resp.GetName())
		return nil
	} else {
		return c.Stream("/ws/v1/exec", headers, os.Stdin, os.Stdout)
	}
}

func (c *Client) GetDockerCreds() (*pb.DockerCred, error) {
	et := &empty.Empty{}
	return c.ProviderServiceClient.DockerCredsGet(ctx, et)
}
