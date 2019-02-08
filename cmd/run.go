package cmd

import (
	"fmt"
	"io"
	"os"

	pb "github.com/datacol-io/datacol/api/models"
	"github.com/datacol-io/datacol/cmd/stdcli"
	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh/terminal"
)

func init() {
	stdcli.AddCommand(cli.Command{
		Name:   "run",
		Usage:  "execute a command in an app",
		Action: cmdAppRun,
		Flags: []cli.Flag{
			&appFlag,
			cli.BoolFlag{
				Name:  "detach, d",
				Usage: "run the command in detached mode",
			},
		},
	})
}

// follow https://github.com/openshift/origin/search?utf8=%E2%9C%93&q=exec+arrow&type=Issues
func cmdAppRun(c *cli.Context) error {
	name, err := getCurrentApp(c)
	stdcli.ExitOnError(err)

	client, close := getApiClient(c)
	defer close()
	_, err = client.GetApp(name)
	stdcli.ExitOnError(err)

	var opts pb.ProcessRunOptions
	opts.Detach = c.Bool("detach")

	if !opts.Detach {
		if w, h, err := terminal.GetSize(int(os.Stdout.Fd())); err == nil {
			opts.Width = w
			opts.Height = h
		}

		opts.Tty = isTerminal(os.Stdin)
		if opts.Tty {
			restore := restoreTerm(os.Stdin, os.Stdout) // restore terminal raw state
			defer restore()
		}
	}

	args := c.Args()
	stdcli.ExitOnError(client.RunProcess(name, args, opts))
	return nil
}

func isTerminal(f *os.File) bool {
	return terminal.IsTerminal(int(f.Fd()))
}

func restoreTerm(in io.Reader, out io.Writer) func() {
	fn := terminalRaw(in)
	return func() {
		if fn() {
			fmt.Fprint(out, '\n')
		}
	}
}

func terminalRaw(reader io.Reader) func() bool {
	var fd int
	var state *terminal.State

	if f, ok := reader.(*os.File); ok {
		fd = int(f.Fd())
		if s, err := terminal.MakeRaw(fd); err == nil {
			state = s
		}
	}

	return func() bool {
		if state != nil {
			terminal.Restore(fd, state)
			return true
		}
		return false
	}
}
