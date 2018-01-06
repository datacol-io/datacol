package client

import (
	"io"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/appscode/go/term"
	pbs "github.com/dinesh/datacol/api/controller"
	pb "github.com/dinesh/datacol/api/models"
	"google.golang.org/grpc/metadata"
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

func (c *Client) RunProcess(name string, args []string) error {
	newctx := metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
		"app":     name,
		"command": strings.Join(args, " "),
	}))

	stream, err := c.ProviderServiceClient.ProcessRun(newctx)
	if err != nil {
		return err //FIXME: not able to make it work for now. Not sure why.
	}

	defer stream.CloseSend()

	wg := sync.WaitGroup{}
	wg.Add(1)

	r, w := os.Stdin, os.Stdout

	go func(out io.Writer) {
		defer wg.Done()

		for {
			ret, rerr := stream.Recv()
			if rerr == io.EOF {
				break
			}

			if rerr != nil {
				err = rerr
				break
			}

			if _, err = out.Write(ret.Data); err != nil {
				break
			}
		}
	}(w)

	go func(r io.Reader) {
		defer wg.Done()
		buf := make([]byte, 1024*1024)

		for {
			n, serr := r.Read(buf)
			if serr == io.EOF {
				break
			}
			if serr != nil {
				err = serr
				break
			}

			if serr := stream.Send(&pbs.StreamMsg{Data: buf[:n]}); serr != nil {
				err = serr
				break
			}
		}
	}(r)

	wg.Wait()

	return err
}
