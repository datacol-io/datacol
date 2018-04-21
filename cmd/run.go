package cmd

import (
	"fmt"
	"os"

	"github.com/datacol-io/datacol/cmd/stdcli"
	"github.com/urfave/cli"
)

func init() {
	stdcli.AddCommand(cli.Command{
		Name:   "run",
		Usage:  "execute a command in an app",
		Action: cmdAppRun,
		Flags:  []cli.Flag{&appFlag},
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

	args := prepareCommand(c)
	stdcli.ExitOnError(client.RunProcess(name, args))

	return nil
}

func prepareCommand(c *cli.Context) []string {
	args := c.Args()

	return append([]string{"env", fmt.Sprintf("TERM=%s", getTermEnv())}, args...)
}

func getTermEnv() string {
	term := os.Getenv("TERM")
	if term == "" {
		term = "xterm"
	}

	return term
}
