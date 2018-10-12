package common

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	docker "github.com/fsouza/go-dockerclient"
)

type imageManifest []struct {
	RepoTags []string
}

func BuildDockerLoad(target, tag string, dkr *docker.Client, r io.Reader, w io.Writer, auth *docker.AuthConfiguration) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("gz reader: %v", err)
	}
	tr := tar.NewReader(gz)
	tmpfile, err := ioutil.TempFile("", "")
	if err != nil {
		return err
	}
	defer os.Remove(tmpfile.Name())

	tw := tar.NewWriter(tmpfile)
	var manifestData []byte

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		logrus.Println("GOT", header.Name, header.Size)
		tw.WriteHeader(header)

		if header.Name == "manifest.json" {
			mfb := &bytes.Buffer{}
			io.Copy(mfb, tr)
			manifestData = mfb.Bytes()
			tw.Write(manifestData)
		} else {
			io.Copy(tw, tr)
		}

		if header.Name == "repositories" {
			break // a hack to break the loook as repositories should be the last item for docker images
		}
	}

	if err = tw.Close(); err != nil {
		return fmt.Errorf("tar writer: %v", err)
	}
	if err = gz.Close(); err != nil {
		return fmt.Errorf("gz writer: %v", err)
	}

	manifest, err := unmarshalImageManifest(manifestData)
	if err != nil {
		return err
	}

	if len(manifest) == 0 {
		logrus.Error("invalid image manifest: no data")
		return fmt.Errorf("invalid image manifest: no data")
	}

	outb := &bytes.Buffer{}
	cmd := exec.Command("docker", "load", "-i", tmpfile.Name())
	cmd.Stdout = outb
	cmd.Stderr = outb

	if err := cmd.Run(); err != nil {
		out := strings.TrimSpace(outb.String())
		return fmt.Errorf("%s: %v", out, err)
	}

	if auth == nil {
		auth = &docker.AuthConfiguration{}
	}

	// cleanup the docker images after completion
	defer func() {
		for _, tags := range manifest {
			for _, image := range tags.RepoTags {
				dkr.RemoveImage(image)
				dkr.RemoveImage(target + ":" + tag)
			}
		}
	}()

	for _, tags := range manifest {
		for _, image := range tags.RepoTags {
			if err := dkr.TagImage(image, docker.TagImageOptions{Repo: target, Tag: tag}); err != nil {
				return err
			}
			logrus.Printf("pushing image %s to %s:%s", image, target, tag)

			if err := dkr.PushImage(docker.PushImageOptions{
				Name:              target,
				Tag:               tag,
				OutputStream:      w,
				InactivityTimeout: 5 * time.Minute,
			}, *auth); err != nil {
				return fmt.Errorf("failed to push image=%s:%s err:%v", target, tag, err)
			}
		}
	}

	return nil
}

func unmarshalImageManifest(mdata []byte) (imageManifest, error) {
	var manifest imageManifest
	if err := json.Unmarshal(mdata, &manifest); err != nil {
		return nil, err
	}

	return manifest, nil
}

func extractImageManifest(r io.Reader) (imageManifest, error) {
	mtr := tar.NewReader(r)

	var manifest imageManifest

	for {
		mh, err := mtr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		fmt.Printf(mh.Name + " ")
		if mh.Name == "manifest.json" {
			var mdata bytes.Buffer

			if _, err := io.Copy(&mdata, mtr); err != nil {
				return nil, err
			}

			if err := json.Unmarshal(mdata.Bytes(), &manifest); err != nil {
				return nil, err
			}

			return manifest, nil
		}
	}

	return nil, fmt.Errorf("unable to locate manifest")
}
