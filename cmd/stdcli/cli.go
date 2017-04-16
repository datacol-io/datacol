package stdcli

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	rollbarAPI "github.com/stvp/rollbar"
	"gopkg.in/urfave/cli.v2"
)

var (
	Binary      string
	Version     string
	Commands    []cli.Command
	localappdir string
)

func init() {
	Version = "0.0.1-alpha.1"
	localappdir = ".dtcol"
	Binary = filepath.Base(os.Args[0])
	Commands = []cli.Command{}
}

func New() *cli.App {
	app := &cli.App{
		Name:     Binary,
		Commands: Commands,
		Version:  Version,
	}

	app.CommandNotFound = func(c *cli.Context, cmd string) {
		fmt.Fprintf(os.Stderr, "No such command \"%s\". Try `%s help`\n", cmd, Binary)
		os.Exit(1)
	}

	return app
}

func GetStack() string {
	stack := GetSetting("stack")
	if stack == "" {
		stack = os.Getenv("STACK")
	}

	if stack == "" {
		Error(fmt.Errorf("no stack found, Please run $] datacol init"))
	}
	return stack
}

func AddCommand(cmd cli.Command) {
	Commands = append(Commands, cmd)
}

func GetSetting(setting string) string {
	value, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", localappdir, setting))
	if err != nil {
		return ""
	}
	output := strings.TrimSpace(string(value))

	return output
}

func WriteSetting(setting, value string) error {
	if err := os.MkdirAll(localappdir, 0777); err != nil {
		return err
	}

	return ioutil.WriteFile(
		fmt.Sprintf(localappdir+"/%s", setting),
		[]byte(value),
		0777,
	)
}

func RmSettingDir(path string) error {
	return os.RemoveAll(filepath.Join(path, localappdir))
}

func CheckFlagsPresence(c *cli.Context, flags ...string) {
	for _, name := range flags {
		value := c.String(name)
		if value == "" {
			Error(fmt.Errorf("Missing required param %v", name))
		}
	}
}

func Error(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func HandlePanicErr(err error) {
	fmt.Println(err.Error())
	rollbar(err, "error")
}

// EnsureOnlyFlags ensures that every element in the args slice starts with --
func EnsureOnlyFlags(c *cli.Context, args []string) {
	for _, a := range args {
		if !strings.HasPrefix(a, "--") {
			Error(fmt.Errorf("got unexpected argument '%s'; please provide parameters in --flag or --flag=value format", a))
			Usage(c)
		}
	}
}

// FlagsToOptions converts a list of '--key=value'/'--bool' strings to 'key: value, bool: true'-style map
func FlagsToOptions(c *cli.Context, args []string) map[string]string {
	options := parseOpts(args)
	for key, value := range options {
		if value == "" {
			options[key] = "true"
		}
	}
	return options
}

func parseOpts(args []string) map[string]string {
	options := make(map[string]string)
	var key string

	for _, token := range args {
		isFlag := strings.HasPrefix(token, "-")
		if isFlag {
			key = strings.TrimLeft(token, "-")
			value := ""
			if strings.Contains(key, "=") {
				pivot := strings.Index(key, "=")
				value = key[pivot+1:]
				key = key[0:pivot]
			}

			key = strings.Replace(key, "-", "_", -1)
			options[key] = value
		} else {
			value := options[key]
			key = strings.Replace(key, "-", "_", -1)
			options[key] = strings.TrimSpace(value + " " + token)
		}
	}

	return options
}

// Usage prints help for the current command and exits
func Usage(c *cli.Context) {
	cli.ShowCommandHelp(c, c.Command.Name)
	os.Exit(129)
}

func rollbar(err error, level string) {
	if os.Getenv("TESTING") == "1" {
		panic(err)
		return
	}

	rollbarAPI.Platform = "client"
	rollbarAPI.Token = "915b990fdfee4bd4a8c280b3a838205d"
	var cmd string

	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	fields := []*rollbarAPI.Field{
		{"version", Version},
		{"os", runtime.GOOS},
		{"arch", runtime.GOARCH},
		{"command", cmd},
	}

	rollbarAPI.Error(level, err, fields...)
	rollbarAPI.Wait()
}
