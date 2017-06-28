package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"text/template"

	"github.com/dinesh/datacol/cmd/stdcli"
)

var (
	crashing = false
	re       = regexp.MustCompile("[^a-z0-9]+")
	dkrYAML  = `FROM gcr.io/google-appengine/{{ .Runtime }}

{{- range .RuntimeSteps}}
{{.}}
{{- end }}

ENV PORT 8080
{{- range $key, $value := .EnvVariables }}
ENV {{ $key }} {{ $value }}
{{- end }}

{{- range .Network.Ports }}
EXPOSE {{.}}
{{- end }}
EXPOSE 8080

ADD . /app
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

func slug(s string) string {
	return strings.Trim(re.ReplaceAllString(strings.ToLower(s), "-"), "-")
}

func consoleURL(api, pid string) string {
	return fmt.Sprintf("https://console.developers.google.com/apis/api/%s/overview?project=%s", api, pid)
}

type appYAMLConfig struct {
	Runtime       string            `yaml:"runtime"`
	Env           string            `yaml:"env"`
	Entrypoint    string            `yaml:"entrypoint"`
	EnvVariables  map[string]string `yaml:"env_variables"`
	RuntimeConfig struct {
		PythonVersion string `yaml:"python_version"`
	} `yaml:"runtime_config"`
	Network struct {
		Ports []string `yaml:"forwarded_ports"`
	} `yaml:"network"`
	Resources struct {
		CPU      string `yaml:"cpu"`
		Memory   string `yaml:"memory_gb"`
		Disksize string `yaml:"disk_size_gb"`
	} `yaml:"resources"`
	HealthCheck struct {
		EnableCheck   bool  `yaml:"enable_health_check"`
		CheckInternal int32 `yaml:"check_interval_sec"`
		TimeoutTh     int32 `yaml:"timeout_sec"`
		HealthyTh     int32 `yaml:"healthy_threshold"`
		UnhealthyTh   int32 `yaml:"unhealthy_threshold"`
		RestartTh     int32 `yaml:"restart_threshold"`
	} `yaml:"health_check"`
	AutomaticScalinng struct {
		MinInstances int32 `yaml:"min_num_instances"`
		MaxInstances int32 `yaml:"max_num_instances"`
	} `yaml:"automatic_scaling"`
	ManualScaling struct {
		Instances int32 `yaml:"instances"`
	} `yaml:"manual_scaling"`
	RuntimeSteps []string
}

func parseAppYAML() (*appYAMLConfig, error) {
	data, err := ioutil.ReadFile("app.yaml")
	if err != nil {
		return nil, err
	}

	var appyaml appYAMLConfig
	if err := yaml.Unmarshal(data, &appyaml); err != nil {
		return nil, err
	}

	if !(appyaml.Env == "flex" || appyaml.Env == "custom") {
		fmt.Printf("\nignoring %s env", appyaml.Env)
		return &appyaml, nil
	}

	return nil, fmt.Errorf("invalid app.yaml file.")
}

func gnDockerFromGAE(filename string) error {
	appyaml, err := parseAppYAML()
	if err != nil {
		return err
	}

	if len(appyaml.Entrypoint) > 0 {
		appyaml.Entrypoint = entrypoint(appyaml.Entrypoint)
	} else {
		appyaml.Entrypoint = "/bin/sh -c"
	}

	appyaml.RuntimeSteps = getruntimeSteps(appyaml)

	log.Debugf(toJson(appyaml))
	tmpl, err := template.New("ct").Parse(dkrYAML)
	if err != nil {
		return err
	}

	var doc bytes.Buffer
	if err := tmpl.Execute(&doc, appyaml); err != nil {
		return err
	}

	err = ioutil.WriteFile(filename, doc.Bytes(), 0700)
	if err != nil {
		return err
	}

	return nil
}

func entrypoint(cmd string) string {
	parts := strings.Split(cmd, " ")
	return `["` + strings.Join(parts, `", "`) + `"]`
}

func getruntimeSteps(spec *appYAMLConfig) []string {
	steps := []string{}
	switch spec.Runtime {
	case "python":
		steps = append(steps, []string{
			"RUN virtualenv /env",
			"ENV VIRTUAL_ENV /env",
			"ENV PATH /env/bin:$PATH",
			"ADD requirements.txt /app/requirements.txt",
			"RUN pip install -r /app/requirements.txt",
		}...)
	default:
		log.Fatal(fmt.Errorf("unsupported runtime %s", spec.Runtime))
	}

	return steps
}

func toJson(object interface{}) string {
	dump, err := json.MarshalIndent(object, " ", "  ")
	if err != nil {
		log.Fatal(fmt.Errorf("dumping json: %v", err))
	}
	return string(dump)
}

func compileTmpl(content string, opts interface{}) string {
	tmpl, err := template.New("ct").Parse(content)
	if err != nil {
		log.Fatal(err)
	}

	var doc bytes.Buffer
	if err := tmpl.Execute(&doc, opts); err != nil {
		log.Fatal(err)
	}

	return doc.String()
}
