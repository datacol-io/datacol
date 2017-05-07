package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/docker/docker/builder/dockerignore"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/fileutils"
	"gopkg.in/urfave/cli.v2"

	log "github.com/Sirupsen/logrus"

	pb "github.com/dinesh/datacol/api/models"
	"github.com/dinesh/datacol/cmd/stdcli"
)

func init() {
	stdcli.AddCommand(&cli.Command{
		Name:   "build",
		Usage:  "build an app from Dockerfile or app.yaml(App-Engine)",
		Action: cmdBuild,
	})
}

func cmdBuild(c *cli.Context) error {
	client := getClient(c)

	dir, name, err := getDirApp(".")
	if err != nil {
		return err
	}

	app, err := client.GetApp(name)
	if err != nil {
		log.Warn(err)
		return app404Err(name)
	}

	build := client.NewBuild(app)
	return executeBuildDir(c, build, dir)
}

func executeBuildDir(c *cli.Context, b *pb.Build, dir string) error {
	tar, err := createTarball(dir)
	if err != nil {
		return err
	}

	fmt.Println("OK")

	objectName, err := uploadBuildSource(c, b, tar)
	if err != nil {
		return err
	}

	return finishBuild(c, b, objectName)
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

func uploadBuildSource(c *cli.Context, b *pb.Build, tarf []byte) (string, error) {
	client := getClient(c)
	source := fmt.Sprintf("%s.tar.gz", b.Id)

	if err := client.Provider().BuildImport(source, tarf); err != nil {
		return "", nil
	}
	return source, nil
}

func finishBuild(c *cli.Context, b *pb.Build, objectName string) error {
	bopts := &pb.BuildOptions{Key: objectName, Id: b.Id}

	return getClient(c).Provider().BuildCreate(b.App, objectName, bopts)
}
