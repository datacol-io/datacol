package common

import (
	"fmt"

	yaml "gopkg.in/yaml.v2"
)

const (
	StandardType = "standard"
	ExtentedType = "extended"

	WebProcessKind = "web"
	CmdProcessKind = "cmd"

	defaultShell = "sh"
)

type Procfile interface {
	Version() string
	HasProcessType(string) bool
	Command(proctype string) ([]string, error)
}

type Process struct {
	Command     interface{}       `yaml:"command"`
	Cron        *string           `yaml:"cron,omitempty"`
	Environment map[string]string `yaml:"environment,omitempty"`
}

type StdProcfile map[string]string

func (s StdProcfile) Version() string {
	return StandardType
}

func (s StdProcfile) HasProcessType(key string) bool {
	_, ok := s[key]
	return ok
}

func (s StdProcfile) Command(proctype string) ([]string, error) {
	if value, ok := s[proctype]; ok {
		return []string{defaultShell, "-c", value}, nil
	}

	return []string{}, nil
}

type ExtProcfile map[string]Process

func (s ExtProcfile) Version() string {
	return ExtentedType
}

func (ep ExtProcfile) Command(key string) (cmd []string, err error) {
	if _, ok := ep[key]; !ok {
		return cmd, nil
	}

	command := ep[key].Command
	switch command.(type) {
	case string:
		cmd = append(cmd, defaultShell, "-c", command.(string))
	case []interface{}:
		for _, c := range command.([]interface{}) {
			cmd = append(cmd, c.(string))
		}
	default:
		err = fmt.Errorf("unexpected command for %s=%v", key, command)
	}

	return
}

func (s ExtProcfile) HasProcessType(key string) bool {
	_, ok := s[key]
	return ok
}

func ParseProcfile(b []byte) (Procfile, error) {
	p, err := parseStdProcfile(b)
	if err != nil {
		p, err = parseExtProcfile(b)
	}
	return p, err
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
