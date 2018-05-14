package cmd

import (
	"os"
	"time"

	"github.com/datacol-io/datacol/cmd/stdcli"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"
)

func init() {
	stdcli.AddCommand(cli.Command{
		Name:   "releases",
		Action: cmdReleaseList,
		Subcommands: []cli.Command{
			{
				Name:  "promote",
				Usage: "promote a release",
			},
		},
	})
}

func cmdReleaseList(c *cli.Context) error {
	name, err := getCurrentApp(c)
	stdcli.ExitOnError(err)

	client, close := getApiClient(c)
	defer close()

	releases, err := client.GetReleases(name)
	stdcli.ExitOnError(err)

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "BUILD", "STATUS", "CREATED"})
	for _, r := range releases {
		delta := elaspedDuration(time.Unix(int64(r.CreatedAt), 0))
		table.Append([]string{r.Id, r.BuildId, r.Status, delta})
	}

	table.Render()
	return nil
}

func cmdReleasePromote(c *cli.Context) error {
	return nil
}
