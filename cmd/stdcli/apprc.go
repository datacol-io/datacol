package stdcli

import (
	"fmt"
	"os"

	"github.com/appscode/go/io"
	term "github.com/appscode/go/term"
	pb "github.com/datacol-io/datacol/api/models"
	"github.com/urfave/cli"
)

var apprcPath string

func init() {
	apprcPath = pb.ApprcPath
}

type Auth struct {
	Name      string `json:"name,omitempty"`
	Provider  string `json:"provider,omitempty"`
	ApiServer string `json:"api_server,omitempty"`
	ApiKey    string `json:"api_key,omitempty"`
	Project   string `json:"project,omitempty"` // for gcp only
	Bucket    string `json:"bucket,omitempty"`
	Region    string `json:"region,omitempty"` // for aws only
}

type Apprc struct {
	Context string  `json:"context,omitempty"`
	Auths   []*Auth `json:"auths,omitempty"`
}

func (rc *Apprc) GetAuth(stack string) *Auth {
	if stack == "" {
		stack = rc.Context
	}
	if stack != "" {
		for _, a := range rc.Auths {
			if a.Name == stack {
				return a
			}
		}
	}
	return nil
}

func (rc *Apprc) SetAuth(a *Auth) error {
	for i, b := range rc.Auths {
		if b.Name == a.Name {
			rc.Auths = append(rc.Auths[:i], rc.Auths[i+1:]...)
			break
		}
	}
	rc.Context = a.Name
	rc.Auths = append(rc.Auths, a)
	return rc.Write()
}

func (rc *Apprc) DeleteAuth() error {
	if rc.Context != "" {
		for i, a := range rc.Auths {
			if a.Name == rc.Context {
				rc.Auths = append(rc.Auths[:i], rc.Auths[i+1:]...)
				rc.Context = ""
				break
			}
		}
	}
	return rc.Write()
}

func (rc *Apprc) Write() error {
	err := io.WriteJson(apprcPath, rc)
	if err != nil {
		return err
	}
	os.Chmod(apprcPath, 0600)
	return nil
}

func LoadApprc() (*Apprc, error) {
	if _, err := os.Stat(apprcPath); err != nil {
		return nil, err
	}

	os.Chmod(apprcPath, 0600)
	rc := &Apprc{}
	err := io.ReadFileAs(apprcPath, rc)
	if err != nil {
		return nil, err
	}
	return rc, nil
}

/* Exits if there is any error.*/
func GetAuthContextOrDie(stack string) (*Auth, *Apprc) {
	return loadAuthForStack(stack)
}

/* Exits if there is any error.*/
func GetAuthOrDie(c *cli.Context) (*Auth, *Apprc) {
	stack := c.String("stack")
	if stack == "" {
		stack = GetAppSetting("stack")
	}

	return loadAuthForStack(stack)
}

/* Exits if there is any error.*/
func GetAuthOrAnon() (*Auth, bool) {
	rc, err := LoadApprc()
	if err != nil {
		return NewAnonAUth(), false
	}
	a := rc.GetAuth("")

	if a == nil {
		return NewAnonAUth(), false
	}
	return a, true
}

func SetAuth(a *Auth) error {
	rc, err := LoadApprc()
	if err != nil {
		rc = &Apprc{}
		rc.Auths = make([]*Auth, 0)
	}
	return rc.SetAuth(a)
}

func NewAnonAUth() *Auth {
	a := &Auth{ApiServer: "localhost"}
	return a
}

func loadAuthForStack(stack string) (*Auth, *Apprc) {
	rc, err := LoadApprc()
	if err != nil {
		exitWithLoginError("failed to load config file.")
	}

	if stack == "" {
		term.Fatalln("No stack found. Please provide `STACK` environment variable or --stack flag")
	}

	a := rc.GetAuth(stack)

	if a == nil {
		exitWithLoginError(fmt.Sprintf("No stack found for `%s`.", stack))
	}
	return a, rc
}

func exitWithLoginError(msg ...string) {
	for _, m := range msg {
		term.Println(m)
	}
	term.Fatalln("Since the command requires authentication, please run `datacol login` if you haven't logged in.")
}
