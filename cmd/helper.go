package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"time"

	log "github.com/Sirupsen/logrus"
	pb "github.com/datacol-io/datacol/api/models"
	"github.com/datacol-io/datacol/cmd/stdcli"
	"gopkg.in/yaml.v2"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func withIntSuffix(seed string) string {
	return fmt.Sprintf("%s-%d", seed, (rand.Intn(89999) + 1000))
}

var (
	crashing = false
	re       = regexp.MustCompile("[^a-z0-9]+")
)

func handlePanic(version, token string) {
	if crashing {
		return
	}
	crashing = true

	if rec := recover(); rec != nil {
		err, ok := rec.(error)
		if !ok {
			err = errors.New(rec.(string))
		}

		stdcli.HandlePanicErr(err, token, version)
		os.Exit(1)
	}
}

func slug(s string) string {
	return strings.Trim(re.ReplaceAllString(strings.ToLower(s), "-"), "-")
}

func consoleURL(api, pid string) string {
	return fmt.Sprintf("https://console.developers.google.com/apis/api/%s/overview?project=%s", api, pid)
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

func parseProcfile() (data []byte, err error) {
	data, err = ioutil.ReadFile("Procfile")
	return
}

func unmarshalProcfile(procfile []byte) (map[string]string, error) {
	procfileMap := make(map[string]string)
	return procfileMap, yaml.Unmarshal(procfile, &procfileMap)
}

func elaspedDuration(t time.Time) string {
	duration := time.Since(t)
	days := int64(duration.Hours() / 24)
	hours := int64(math.Mod(duration.Hours(), 24))
	minutes := int64(math.Mod(duration.Minutes(), 60))

	chunks := []struct {
		singularName string
		amount       int64
	}{
		{"d", days},
		{"h", hours},
		{"m", minutes},
	}

	parts := []string{}

	for _, chunk := range chunks {
		switch chunk.amount {
		case 0:
			continue
		default:
			parts = append(parts, fmt.Sprintf("%d%s", chunk.amount, chunk.singularName))
		}
	}

	return strings.Join(parts, " ")
}

func sortEnvKeys(current pb.Environment) []string {
	keys := make([]string, 0, len(current))
	for key := range current {
		keys = append(keys, key)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	return keys
}
