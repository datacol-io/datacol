package main

import (
	"fmt"
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
	tar, err := createTarball(dir)
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

func createTarball(base string) ([]byte, error) {
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
		filename := "Dockerfile"
		if _, err = os.Stat("app.yaml"); err == nil {
			fmt.Printf("Trying to generate %s from app.yaml ...", filename)
			if err = gnDockerFromGAE(filename); err != nil {
				fmt.Println(" failed")
				log.Warn(err)
			} else {
				fmt.Println(" done")
				dockerfileName = filename
			}
		}
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
	index := int32(0)

	for {
		time.Sleep(2 * time.Second)

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

		if len(lines) > 0 && lines[len(lines)-1] == "DONE" {
			break
		}
	}

	return nil
}
