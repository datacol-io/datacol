package main

import (
	"os"
	"fmt"
	"log"
	"bufio"
	"bytes"
	"errors"
	"regexp"
	"strings"
	"io/ioutil"
	"text/template"
  "gopkg.in/yaml.v2"

	"github.com/dinesh/datacol/cmd/stdcli"
)

var (
	crashing = false
	re       = regexp.MustCompile("[^a-z0-9]+")
	dkrYAML  = `FROM gcr.io/google-appengine/{{ .Runtime }}
ADD . /app
WORKDIR /app

ENV PORT 8080
EXPOSE 8080
{{- range $key, $value := .EnvVariables }}
ENV {{ $key }} {{ $value }}
{{- end }}

CMD {{ .Entrypoint }}
`
)

func handlePanic() {
	if crashing {
		return
	}
	crashing = true

	if rec := recover(); rec != nil {
		err, ok := rec.(error)
		if !ok {
			err = errors.New(rec.(string))
		}

		stdcli.HandlePanicErr(err)
		os.Exit(1)
	}
}

func confirm(s string, tries int) bool {
	r := bufio.NewReader(os.Stdin)

	for ; tries > 0; tries-- {
		fmt.Printf("%s [y/n]: ", s)

		res, err := r.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		// Empty input (i.e. "\n")
		if len(res) < 2 {
			continue
		}

		return strings.ToLower(strings.TrimSpace(res))[0] == 'y'
	}

	return false
}

func slug(s string) string {
	return strings.Trim(re.ReplaceAllString(strings.ToLower(s), "-"), "-")
}

func consoleURL(api, pid string) string {
	return fmt.Sprintf("https://console.developers.google.com/apis/api/%s/overview?project=%s", api, pid)
}

type appYAMLConfig struct {
	Runtime 			string 	`yaml:"runtime"`
	Env     			string 	`yaml:"env"`
	Entrypoint  	string  `yaml:"entrypoint"`
	EnvVariables 	map[string]string `yaml:"env_variables"`
	RuntimeConfig map[string]string `yaml:"runtime_config"`
}

func gaeTodocker() error {
	data, err := ioutil.ReadFile("app.yaml")
	if err != nil { return err }

	var appyaml appYAMLConfig
  if err := yaml.Unmarshal(data, &appyaml); err != nil {
		return err
	}

	if len(appyaml.Entrypoint) > 0 {
		appyaml.Entrypoint = entrypoint(appyaml.Entrypoint)
	} else {
		appyaml.Entrypoint = "/bin/sh -c"
	}

	fmt.Printf("%+v\n", appyaml)
  tmpl, err := template.New("ct").Parse(dkrYAML)
	if err != nil { return err }

	var doc bytes.Buffer
	if err := tmpl.Execute(&doc, appyaml); err != nil {
		return err
	}

	err = ioutil.WriteFile("Dockerfile", doc.Bytes(), 0700)
	if err != nil {
		return err
	}

	return nil
}

func entrypoint(cmd string) string {
	parts := strings.Split(cmd, " ")
	return `["` + strings.Join(parts, `", "`) + `"]`
}
