package stdcli

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	term "github.com/appscode/go/term"
	pb "github.com/datacol-io/datacol/api/models"
	rollbarAPI "github.com/stvp/rollbar"
	"github.com/urfave/cli"
	"google.golang.org/grpc/status"
)

var (
	Binary      string
	Commands    []cli.Command
	LocalAppDir string
	Stack404    error
)

func init() {
	LocalAppDir = ".datacol"
	Binary = filepath.Base(os.Args[0])
	Commands = []cli.Command{}
	Stack404 = errors.New("stack not found. To create a new stack run `datacol init` or set `STACK` environment variable.")
}

func New() *cli.App {
	app := cli.NewApp()

	app.Name = Binary
	app.Commands = Commands
	app.CommandNotFound = func(c *cli.Context, cmd string) {
		fmt.Fprintf(os.Stderr, "No such command \"%s\". Try `%s help`\n", cmd, Binary)
		os.Exit(1)
	}

	return app
}

func GetAppStack() string {
	stack := GetAppSetting("stack")
	if stack == "" {
		stack = os.Getenv("STACK")
	}

	if stack == "" {
		ExitOnError(Stack404)
	}
	return stack
}

func AddCommand(cmd cli.Command) {
	Commands = append(Commands, cmd)
}

func GetAppSetting(setting string) string {
	value, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", LocalAppDir, setting))
	if err != nil {
		return ""
	}
	output := strings.TrimSpace(string(value))

	return output
}

func WriteAppSetting(setting, value string) error {
	if err := os.MkdirAll(LocalAppDir, 0777); err != nil {
		return err
	}

	return ioutil.WriteFile(
		fmt.Sprintf(LocalAppDir+"/%s", setting),
		[]byte(value),
		0777,
	)
}

func RmSettingDir(path string) error {
	return os.RemoveAll(filepath.Join(path, LocalAppDir))
}

func CurrentStack() string {
	name := os.Getenv("STACK")
	name = strings.Split(name, "@")[0]

	if len(name) == 0 {
		v, err := ioutil.ReadFile(filepath.Join(pb.ConfigPath, "stack"))
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Printf("no active stack found.")
				os.Exit(0)
			} else {
				log.Fatal(err)
			}
		}
		name = strings.TrimSpace(string(v))
	}

	return name
}

func WriteSetting(name, setting, value string) error {
	path := filepath.Join(pb.ConfigPath, name, setting)
	return ioutil.WriteFile(path, []byte(value), 0700)
}

func ReadSetting(name, setting string) string {
	path := filepath.Join(pb.ConfigPath, name, setting)
	value, err := ioutil.ReadFile(path)
	if err != nil {
		return ""
	}
	output := strings.TrimSpace(string(value))

	return output
}

func CheckFlagsPresence(c *cli.Context, flags ...string) {
	for _, name := range flags {
		value := c.String(name)
		if value == "" {
			term.Errorln(fmt.Errorf("Missing required param --%v", name))
			cli.ShowCommandHelp(c, c.Command.Name)
			os.Exit(1)
		}
	}
}

func ExitOnError(err error) {
	if err != nil {
		if gerr, ok := status.FromError(err); ok {
			term.Fatalln(gerr.Message())
		}
		term.Fatalln(err)
	}
}

func ExitOnErrorf(msg string, args ...interface{}) {
	ExitOnError(fmt.Errorf(msg, args...))
}

func HandlePanicErr(err error, token, version string) {
	term.Errorln(err.Error())

	rollbarAPI.Platform = "client"
	rollbarAPI.Token = token
	var cmd string

	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	fields := []*rollbarAPI.Field{
		{"version", version},
		{"os", runtime.GOOS},
		{"arch", runtime.GOARCH},
		{"command", cmd},
	}

	rollbarAPI.Error("error", err, fields...)
	rollbarAPI.Wait()
}

// EnsureOnlyFlags ensures that every element in the args slice starts with --
func EnsureOnlyFlags(c *cli.Context, args []string) {
	for _, a := range args {
		if !strings.HasPrefix(a, "--") {
			ExitOnErrorf("got unexpected argument '%s'; please provide parameters in --flag or --flag=value format", a)
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
		} else if parts := strings.Split(token, "="); len(parts) == 2 {
			key, value := parts[0], parts[1]
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

func rollbar(err error, level, token, version string) {
	rollbarAPI.Platform = "client"
	rollbarAPI.Token = token
	var cmd string

	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	fields := []*rollbarAPI.Field{
		{"version", version},
		{"os", runtime.GOOS},
		{"arch", runtime.GOARCH},
		{"command", cmd},
	}

	rollbarAPI.Error(level, err, fields...)
	rollbarAPI.Wait()
}
