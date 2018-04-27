package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"

	log "github.com/Sirupsen/logrus"
	term "github.com/appscode/go/term"
	pbs "github.com/datacol-io/datacol/api/controller"
	pb "github.com/datacol-io/datacol/api/models"
	"github.com/datacol-io/datacol/client"
	"github.com/datacol-io/datacol/cmd/stdcli"

	"github.com/docker/docker/builder/dockerignore"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/fileutils"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"
	"golang.org/x/net/context"
	"golang.org/x/net/websocket"
)

func init() {
	stdcli.AddCommand(cli.Command{
		Name:   "build",
		Usage:  "build an app from Dockerfile",
		Action: cmdBuild,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "ref",
				Usage: "branch or commit Id of git repository",
			},
			&appFlag,
		},
	})

	stdcli.AddCommand(cli.Command{
		Name:   "builds",
		Usage:  "manage the builds for an app",
		Action: cmdBuildList,
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:  "limit, n",
				Usage: "Limit the number of recent builds to fetch",
				Value: 5,
			},
			&appFlag,
		},
		Subcommands: []cli.Command{
			{
				Name:      "delete",
				Usage:     "delete a build",
				ArgsUsage: "<Id>",
				Action:    cmdBuildDelete,
			},
		},
	})
}

func cmdBuildList(c *cli.Context) error {
	name, err := getCurrentApp(c)
	if err != nil {
		return err
	}
	api, close := getApiClient(c)
	defer close()

	builds, err := api.GetBuilds(name, c.Int("limit"))
	stdcli.ExitOnError(err)

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "COMMIT", "STATUS", "CREATED"})
	for _, b := range builds {
		delta := elaspedDuration(time.Unix(int64(b.CreatedAt), 0))
		table.Append([]string{b.Id, b.Version, b.Status, delta})
	}

	table.Render()
	return nil
}

func cmdBuildDelete(c *cli.Context) error {
	_, name, err := getDirApp(".", c)
	stdcli.ExitOnError(err)

	api, close := getApiClient(c)
	defer close()

	if c.NArg() == 0 {
		term.Warningln("No build Id provided")
		stdcli.Usage(c)
	}

	stdcli.ExitOnError(api.DeleteBuild(name, c.Args().First()))

	fmt.Println("DONE")
	return nil
}

func cmdBuild(c *cli.Context) error {
	api, close := getApiClient(c)
	defer close()

	dir, name, err := getDirApp(".", c)
	stdcli.ExitOnError(err)

	app, err := api.GetApp(name)
	if err != nil {
		log.Warn(err)
		return app404Err(name)
	}

	ref := c.String("ref")
	if ref == "" {
		_, err = executeBuildDir(api, app, dir)
	} else {
		_, err = executeBuildGitSource(api, app, ref)
	}

	stdcli.ExitOnError(err)
	return err
}

func executeBuildGitSource(api *client.Client, app *pb.App, version string) (*pb.Build, error) {
	var procfile []byte
	if _, err := os.Stat("Procfile"); err == nil {
		content, err := parseProcfile()
		stdcli.ExitOnError(err)
		procfile = content
	}

	b, err := api.CreateBuildGit(app, version, procfile)
	if err != nil {
		return nil, err
	}
	return b, finishBuild(api, b)
}

func executeBuildDir(api *client.Client, app *pb.App, dir string) (*pb.Build, error) {
	env, err := api.GetEnvironment(app.Name)
	if err != nil {
		return nil, err
	}

	tar, err := createTarball(dir, env)
	if err != nil {
		return nil, err
	}

	fmt.Println("OK")

	var procfile []byte
	if _, err := os.Stat("Procfile"); err == nil {
		content, err := parseProcfile()
		stdcli.ExitOnError(err)
		procfile = content
	}

	b, err := api.CreateBuild(app, tar, procfile)
	if err != nil {
		return nil, err
	}

	return b, finishBuild(api, b)
}

func createTarball(base string, env map[string]string) ([]byte, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	sym, err := filepath.EvalSymlinks(base)
	if err != nil {
		return nil, err
	}

	err = os.Chdir(sym)
	if err != nil {
		return nil, err
	}

	dockerfileName := "Dockerfile"

	if _, err := os.Stat(dockerfileName); os.IsNotExist(err) {
		stdcli.ExitOnError(fmt.Errorf("Dockerfile not found."))
	}

	fmt.Print("Creating tarball ...")

	var includes = []string{"."}
	var excludes []string

	dockerIgnorePath := path.Join(sym, ".dockerignore")
	dockerIgnore, err := os.Open(dockerIgnorePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		excludes = make([]string, 0)
	} else {
		excludes, err = dockerignore.ReadAll(dockerIgnore)
		if err != nil {
			return nil, err
		}
	}

	excludes = append(excludes, stdcli.LocalAppDir)

	keepThem1, _ := fileutils.Matches(".dockerignore", excludes)
	keepThem2, _ := fileutils.Matches("Dockerfile", excludes)
	if keepThem1 || keepThem2 {
		includes = append(includes, ".dockerignore", "Dockerfile")
	}

	options := &archive.TarOptions{
		Compression:     archive.Gzip,
		ExcludePatterns: excludes,
		IncludeFiles:    includes,
	}

	out, err := archive.TarWithOptions(sym, options)
	if err != nil {
		return nil, err
	}

	bytes, err := ioutil.ReadAll(out)
	if err != nil {
		return nil, err
	}

	err = os.Chdir(cwd)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func finishBuild(api *client.Client, b *pb.Build) (err error) {
	// try to sync local state
	b, err = api.GetBuild(b.App, b.Id)
	if err != nil {
		return err
	}

	if api.IsGCP() {
		return finishBuildGCP(api, b)
	}

	return finishBuildAwsWs(api, b)
}

func finishBuildAwsWs(api *client.Client, b *pb.Build) error {
	ws, err := api.StreamClient("/ws/v1/builds/logs", map[string]string{"id": b.Id})
	if err != nil {
		return err
	}

	defer ws.Close()

	var (
		done = make(chan bool, 2)
		out  = os.Stdout
	)

	if out != nil {
		go copyAsync(out, ws)
	}

	// We need to pool the build status to short-circuit the streaming logs
	// FIXME: waitForAwsBuild might get into zombie goroutine if copyAsync returns quick
	go waitforAwsBuild(api, b, done, ws, 5*time.Second)

	<-done

	return nil
}

func finishBuildGCP(api *client.Client, b *pb.Build) (err error) {
	index := int32(0)

	b, err = api.GetBuild(b.App, b.Id)
	if err != nil {
		return err
	}

OUTER:
	for {
		time.Sleep(1 * time.Second)

		ret, err := api.BuildLogs(context.TODO(), &pbs.BuildLogRequest{
			App: b.App,
			Id:  b.RemoteId,
			Pos: index,
		})

		if err != nil {
			return fmt.Errorf("Getting logs for build: %s err: %v", b.RemoteId, err)
		}

		index = ret.Pos
		lines := ret.Lines

		for _, line := range lines {
			fmt.Println(line)
		}

		b, err := api.GetBuild(b.App, b.Id)
		if err != nil {
			return err
		}

		switch b.Status {
		case "SUCCESS":
			break OUTER
		case "WORKING":
		default:
			if b.Status != "" {
				gcrBuildURL := fmt.Sprintf("https://console.cloud.google.com/gcr/builds/%s?project=%s", b.RemoteId, api.Project)
				return fmt.Errorf("Build status: %s\n. Please go to %s to see complete build logs.", b.Status, gcrBuildURL)
			}
		}
	}

	return nil
}

func copyAsync(dst io.Writer, src io.Reader) {
	io.Copy(dst, src)
}

func waitforAwsBuild(api *client.Client, b *pb.Build, done chan bool, ws *websocket.Conn, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		newb, _ := api.GetBuild(b.App, b.Id)
		if newb.Status != "IN_PROGRESS" {
			fmt.Println("Build Id:", newb.Id)
			fmt.Println("Build status:", newb.Status)
			break
		}
	}

	ws.Close()
	done <- true
}
