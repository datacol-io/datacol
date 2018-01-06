package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"
	"strings"

	log "github.com/Sirupsen/logrus"
	term "github.com/appscode/go/term"
	"github.com/dinesh/datacol/cmd/stdcli"
	semver "github.com/hashicorp/go-version"
	"github.com/inconshreveable/go-update"
	"github.com/urfave/cli"
)

func init() {
	stdcli.AddCommand(cli.Command{
		Name:   "upgrade",
		Usage:  "upgrade datacol CLI to latest version",
		Action: cmdUpgrade,
	})

	cli.VersionPrinter = cmdVersion
}

func cmdVersion(c *cli.Context) {
	upgradeNudge(c)
	fmt.Fprintf(c.App.Writer, "version=%s\n", c.App.Version)
}

func cmdUpgrade(c *cli.Context) error {
	currentv, err := semver.NewVersion(c.App.Version)
	stdcli.ExitOnError(err)

	lv := latestVersion()
	newv, err := semver.NewVersion(lv)
	stdcli.ExitOnError(err)

	if newv.GreaterThan(currentv) {
		url := binaryURL(lv)
		fmt.Printf("Updating from %s to %s ...", c.App.Version, lv)
		log.Debugf("\nDownloading from %s", url)

		resp, err := http.Get(url)
		stdcli.ExitOnError(err)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				term.Errorln(err)
			}

			err = fmt.Errorf("fetching latest version err:%s", string(b))
			stdcli.ExitOnError(err)
		}

		if err := update.Apply(resp.Body, update.Options{}); err != nil {
			stdcli.ExitOnError(err)
		}

		fmt.Printf(" DONE\n")
	} else {
		fmt.Printf("You already have latest version: %s\n", lv)
	}

	return nil
}

func upgradeNudge(c *cli.Context) {
	newv, err := semver.NewVersion(latestVersion())
	if err != nil {
		log.Fatal(err)
	}

	currv, err := semver.NewVersion(c.App.Version)
	if err != nil {
		log.Fatal(err)
	}

	if newv.GreaterThan(currv) {
		fmt.Printf("New version %s is available, run `datacol upgrade` to update.\n", newv.String())
	}
}

func latestVersion() string {
	resp, err := http.Get("https://storage.googleapis.com/datacol-distros/binaries/latest.txt")
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatal(fmt.Errorf("fetching latest version err:%s", string(b)))
	}

	return strings.Replace(string(b), "\n", "", -1)
}

func binaryURL(v string) string {
	os := runtime.GOOS
	arch := runtime.GOARCH
	ext := ""

	if os == "windows" {
		ext = ".exe"
	}

	return fmt.Sprintf("https://storage.googleapis.com/datacol-distros/binaries/%s/datacol-%s-%s%s", v, os, arch, ext)
}
