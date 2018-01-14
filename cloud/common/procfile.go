package common

import (
	yaml "gopkg.in/yaml.v2"
)

const (
	StandardType = "standard"
	ExtentedType = "extended"
)

type Process struct {
	Command     string            `yaml:"command"`
	Cron        *string           `yaml:"cron,omitempty"`
	Environment map[string]string `yaml:"environment,omitempty"`
}

type StdProcfile map[string]string

func (s StdProcfile) Version() string {
	return StandardType
}

type ExtProcfile map[string]Process

func (s ExtProcfile) Version() string {
	return ExtentedType
}

func ParseProcfile(b []byte) (Procfile, error) {
	p, err := parseStdProcfile(b)
	if err != nil {
		p, err = parseExtProcfile(b)
	}
	return p, err
}

type Procfile interface {
	Version() string
}

func parseExtProcfile(b []byte) (Procfile, error) {
	y := make(ExtProcfile)
	err := yaml.Unmarshal(b, &y)
	return y, err
}

func parseStdProcfile(b []byte) (Procfile, error) {
	y := make(StdProcfile)
	err := yaml.Unmarshal(b, &y)
	return y, err
}
