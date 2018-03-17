package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	pbs "github.com/datacol-io/datacol/api/controller"
	pb "github.com/datacol-io/datacol/api/models"
	"github.com/datacol-io/datacol/client"
	"github.com/datacol-io/datacol/cmd/stdcli"

	"github.com/docker/docker/builder/dockerignore"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/fileutils"
	"github.com/urfave/cli"
	"golang.org/x/net/context"
	"golang.org/x/net/websocket"
)

func init() {
	stdcli.AddCommand(cli.Command{
		Name:   "build",
		Usage:  "build an app from Dockerfile or app.yaml (App-Engine)",
		Action: cmdBuild,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "ref",
				Usage: "branch or commit Id of git repository",
			},
		},
		Subcommands: []cli.Command{
			{
				Name:   "list",
				Usage:  "get builds for an app",
				Action: cmdBuildList,
			},
			{
				Name:   "delete",
				Usage:  "delete a build",
				Action: cmdBuildDelete,
			},
		},
	})
}

func cmdBuildList(c *cli.Context) error {
	_, name, err := getDirApp(".")
	if err != nil {
		return err
	}
	api, close := getApiClient(c)
	defer close()

	builds, err := api.GetBuilds(name)
	stdcli.ExitOnError(err)

	fmt.Println(toJson(builds))
	return nil
}

func cmdBuildDelete(c *cli.Context) error {
	_, name, err := getDirApp(".")
	stdcli.ExitOnError(err)

	api, close := getApiClient(c)
	defer close()

	if c.NArg() == 0 {
		stdcli.ExitOnError(fmt.Errorf("Please provide id of the build"))
	}

	bid := c.Args().First()

	stdcli.ExitOnError(api.DeleteBuild(name, bid))

	fmt.Println("DONE")
	return nil
}

func cmdBuild(c *cli.Context) error {
	api, close := getApiClient(c)
	defer close()

	dir, name, err := getDirApp(".")
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
	b, err := api.CreateBuildGit(app, version)
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
	dkrexists := true

	if _, err := os.Stat(dockerfileName); os.IsNotExist(err) {
		filename := "Dockerfile"
		dkrexists = false

		if _, err = os.Stat("app.yaml"); err == nil {
			fmt.Printf("Trying to generate %s from app.yaml ...", filename)
			if err = gnDockerFromGAE(filename); err != nil {
				fmt.Println(" FAILED")
				log.Warn(err)
			} else {
				fmt.Println(" DONE")
				dockerfileName = filename
			}
		}

		if err = mkBuildPackDockerfile(sym, env); err != nil {
			return nil, fmt.Errorf("generating Dockerfile err: %v", err)
		}
	}

	if !dkrexists {
		defer func() {
			if err := os.Remove(filepath.Join(sym, dockerfileName)); err != nil {
				log.Errorf("removing Dockerfile err: %v", err)
			}
		}()
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
		wg  sync.WaitGroup
		out = os.Stdout
	)

	if out != nil {
		wg.Add(1)
		go copyAsync(out, ws, &wg)
	}

	// We need to pool the build status to short-circuit the streaming logs
	// FIXME: waitForAwsBuild might get into zombie goroutine if copyAsync returns quick
	go waitforAwsBuild(api, b, &wg, ws, 5*time.Second)

	wg.Wait()

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
				return fmt.Errorf("Build status: %s", b.Status)
			}
		}
	}

	return nil
}

type herokuishOpts struct {
	Env map[string]string
}

func mkBuildPackDockerfile(dir string, env map[string]string) error {
	dockerfile := filepath.Join(dir, "Dockerfile")
	content := compileTmpl(herokuishTmpl, herokuishOpts{env})
	log.Debugf("--- generated Dockerfile ------\n %s --------", content)

	return ioutil.WriteFile(dockerfile, []byte(content), 0644)
}

var herokuishTmpl = `FROM gliderlabs/herokuish:v0.3.29
ADD . /app
{{- $burl := index .Env "BUILDPACK_URL" }} {{ if gt (len $burl) 0 }}
ENV BUILDPACK_URL {{ $burl }}
{{- end }}
RUN /bin/herokuish buildpack build
ENV PORT 8080
EXPOSE 8080
CMD ["/start", "web"]`

func copyAsync(dst io.Writer, src io.Reader, wg *sync.WaitGroup) {
	defer wg.Done()
	io.Copy(dst, src)
}

func waitforAwsBuild(api *client.Client, b *pb.Build, wg *sync.WaitGroup, ws *websocket.Conn, interval time.Duration) {
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
	wg.Done()
}
