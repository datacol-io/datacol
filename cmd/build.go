package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/docker/docker/builder/dockerignore"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/fileutils"
	"golang.org/x/net/context"
	"gopkg.in/urfave/cli.v2"

	log "github.com/Sirupsen/logrus"

	pbs "github.com/dinesh/datacol/api/controller"
	pb "github.com/dinesh/datacol/api/models"
	"github.com/dinesh/datacol/client"
	"github.com/dinesh/datacol/cmd/stdcli"
)

func init() {
	stdcli.AddCommand(&cli.Command{
		Name:   "build",
		Usage:  "build an app from Dockerfile or app.yaml(App-Engine)",
		Action: cmdBuild,
		Subcommands: []*cli.Command{
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
	if err != nil {
		return err
	}

	fmt.Println(toJson(builds))
	return nil
}

func cmdBuildDelete(c *cli.Context) error {
	_, name, err := getDirApp(".")
	if err != nil {
		return err
	}
	api, close := getApiClient(c)
	defer close()

	if c.Args().Len() == 0 {
		return fmt.Errorf("Please provide id of the build")
	}

	bid := c.Args().First()

	if err = api.DeleteBuild(name, bid); err != nil {
		return err
	}

	fmt.Println("DONE")
	return nil
}

func cmdBuild(c *cli.Context) error {
	api, close := getApiClient(c)
	defer close()

	dir, name, err := getDirApp(".")
	if err != nil {
		return err
	}

	app, err := api.GetApp(name)
	if err != nil {
		log.Warn(err)
		return app404Err(name)
	}

	_, err = executeBuildDir(api, app, dir)
	return err
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

	b, err := api.CreateBuild(app, tar)
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

func finishBuild(api *client.Client, b *pb.Build) error {
	if api.IsGCP() {
		return finishBuildGCP(api, b)
	} else {
		return finishBuildAws(api, b)
	}
}

func finishBuildAws(api *client.Client, b *pb.Build) error {
	stream, err := api.ProviderServiceClient.BuildLogsStream(context.TODO(), &pbs.BuildLogStreamReq{Id: b.RemoteId})
	if err != nil {
		return err
	}
	defer stream.CloseSend()

	out := os.Stdout
	for {
		ret, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if _, err := out.Write(ret.Data); err != nil {
			return err
		}
	}
}

func finishBuildGCP(api *client.Client, b *pb.Build) error {
	index := int32(0)

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
			return fmt.Errorf("Build status: %s", b.Status)
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

	if err := ioutil.WriteFile(dockerfile, []byte(content), 0644); err != nil {
		return err
	}
	return nil
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
