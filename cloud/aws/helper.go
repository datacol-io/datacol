package aws

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/appscode/go/crypto/rand"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/docker/docker/pkg/archive"
	"github.com/mholt/archiver"
)

func generateId(prefix string, size int) string {
	return prefix + "-" + rand.Characters(size)
}

func stackParameters(stack *cloudformation.Stack) map[string]string {
	parameters := make(map[string]string)

	for _, parameter := range stack.Parameters {
		parameters[*parameter.ParameterKey] = *parameter.ParameterValue
	}

	return parameters
}

func stackOutputs(stack *cloudformation.Stack) map[string]string {
	outputs := make(map[string]string)

	for _, output := range stack.Outputs {
		outputs[*output.OutputKey] = *output.OutputValue
	}

	return outputs
}

func coalesce(s *dynamodb.AttributeValue, def string) string {
	if s != nil {
		return *s.S
	} else {
		return def
	}
}

func coalesceInt(s *dynamodb.AttributeValue, def int) int {
	if s != nil {
		num, _ := strconv.Atoi(*s.N)
		return num
	} else {
		return def
	}
}

func camelize(dasherized string) string {
	tokens := strings.Split(dasherized, "-")

	for i, token := range tokens {
		switch token {
		case "az":
			tokens[i] = "AZ"
		default:
			tokens[i] = strings.Title(token)
		}
	}

	return strings.Join(tokens, "")
}

func cfParams(source map[string]string) map[string]string {
	params := make(map[string]string)

	for key, value := range source {
		var val string
		switch value {
		case "":
			val = "false"
		case "true":
			val = "true"
		default:
			val = value
		}
		params[camelize(key)] = val
	}

	return params
}

func timestampNow() int32 {
	return int32(time.Now().Unix())
}

func convertGzipToZip(app, src string) (string, error) {
	dir, err := ioutil.TempDir("", app+"-")
	if err != nil {
		return dir, err
	}

	log.Debugf("untar %s to %s", src, dir)
	if err = untarPath(src, dir); err != nil {
		return dir, err
	}

	fileInfos, err := ioutil.ReadDir(dir)
	if err != nil {
		return dir, fmt.Errorf("listing files: %v", err)
	}

	files := make([]string, len(fileInfos))
	for i, f := range fileInfos {
		files[i] = filepath.Join(dir, f.Name())
	}

	zipPath := dir + ".zip"
	if err := archiver.Zip.Make(zipPath, files); err != nil {
		return dir, fmt.Errorf("creating a zip archive err: %v", err)
	}
	return zipPath, nil
}

func buildTemplate(name, section string, data interface{}) (string, error) {
	d, err := Asset(fmt.Sprintf("cloud/aws/templates/%s.tmpl", name))
	if err != nil {
		return "", err
	}

	tmpl, err := template.New(section).Funcs(templateHelpers()).Parse(string(d))
	if err != nil {
		return "", err
	}

	var formation bytes.Buffer

	err = tmpl.Execute(&formation, data)
	if err != nil {
		return "", err
	}

	return formation.String(), nil
}

func untarPath(src, dst string) error {
	fd, err := os.Open(src)
	if err != nil {
		return err
	}
	defer fd.Close()

	defaultArchiver := archive.Archiver{Untar: archive.Untar, UIDMaps: nil, GIDMaps: nil}
	return defaultArchiver.Untar(fd, dst, &archive.TarOptions{NoLchown: true})
}

func toJson(object interface{}) string {
	dump, err := json.MarshalIndent(object, " ", "  ")
	if err != nil {
		log.Fatal(fmt.Errorf("dumping json: %v", err))
	}
	return string(dump)
}

func templateHelpers() template.FuncMap {
	return template.FuncMap{
		"env": func(s string) string {
			return os.Getenv(s)
		},
		"upper": func(s string) string {
			return upperName(s)
		},
		"value": func(s string) template.HTML {
			return template.HTML(fmt.Sprintf("%q", s))
		},
	}
}

func upperName(name string) string {
	// myapp -> Myapp; my-app -> MyApp
	us := strings.ToUpper(name[0:1]) + name[1:]

	for {
		i := strings.Index(us, "-")

		if i == -1 {
			break
		}

		s := us[0:i]

		if len(us) > i+1 {
			s += strings.ToUpper(us[i+1 : i+2])
		}

		if len(us) > i+2 {
			s += us[i+2:]
		}

		us = s
	}

	return us
}

func generatePassword() string {
	return rand.GeneratePassword()
}
