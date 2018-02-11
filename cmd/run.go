package cmd

import (
	"fmt"
	"os"

	"github.com/dinesh/datacol/cmd/stdcli"
	"github.com/urfave/cli"
)

func init() {
	stdcli.AddCommand(cli.Command{
		Name:   "run",
		Usage:  "execute a command in an app",
		Action: cmdAppRun,
	})
}

// follow https://github.com/openshift/origin/search?utf8=%E2%9C%93&q=exec+arrow&type=Issues
func cmdAppRun(c *cli.Context) error {
	_, name, err := getDirApp(".")
	stdcli.ExitOnError(err)

	client, close := getApiClient(c)
	defer close()

	_, err = client.GetApp(name)
	stdcli.ExitOnError(err)

	args := prepareCommand(c)
	stdcli.ExitOnError(client.RunProcess(name, args))

	return nil
}

func prepareCommand(c *cli.Context) []string {
	args := c.Args()
	term := os.Getenv("TERM")
	if term == "" {
		term = "xterm"
	}

	return append([]string{"env", fmt.Sprintf("TERM=%s", term)}, args...)
}
