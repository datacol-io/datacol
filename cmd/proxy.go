package cmd

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/datacol-io/datacol/client"
	"github.com/datacol-io/datacol/cmd/stdcli"
	"github.com/urfave/cli"
)

func init() {
	stdcli.AddCommand(cli.Command{
		Name:        "proxy",
		Description: "proxy local ports into a stack",
		Action:      cmdStackProxy,
		Flags:       []cli.Flag{stackFlag},
	})
}

func cmdStackProxy(c *cli.Context) error {
	for _, arg := range c.Args() {
		parts := strings.SplitN(arg, ":", 2)

		host := "localhost" // hardcoding it since we only have 1 bastion host

		var port, hostport int
		switch len(parts) {
		case 1:
			port = parseInt(parts[0])
			hostport = port
		case 2:
			port = parseInt(parts[0])
			hostport = parseInt(parts[1])
		default:
			stdcli.ExitOnError(fmt.Errorf("invalid argument: %s", arg))
		}

		client, close := getApiClient(c)
		defer close()

		go proxy("127.0.0.1", port, host, hostport, client)
	}

	select {}
}

func parseInt(str string) int {
	p, err := strconv.Atoi(str)
	if err != nil {
		stdcli.ExitOnError(err)
	}
	return p
}

func proxy(localhost string, localport int, remotehost string, remoteport int, client *client.Client) {
	fmt.Printf("proxying %s:%d to %s:%d\n", localhost, localport, remotehost, remoteport)

	listener, err := net.Listen("tcp4", fmt.Sprintf("%s:%d", localhost, localport))
	if err != nil {
		fmt.Printf("error: %s\n", err)
		return
	}

	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("error: %s\n", err)
			return
		}

		defer conn.Close()

		fmt.Printf("connect: %d\n", localport)

		go func() {
			err := client.ProxyRemote(remotehost, remoteport, conn)
			if err != nil {
				fmt.Printf("error: %s\n", err)
				conn.Close()
				return
			}
		}()
	}
}
