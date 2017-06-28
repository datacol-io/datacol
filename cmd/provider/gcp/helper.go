package gcp

import (
	"bufio"
	"bytes"
	"cloud.google.com/go/datastore"
	"context"
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	pb "github.com/dinesh/datacol/api/models"
	"google.golang.org/api/compute/v1"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"text/template"
)

func deleteFromQuery(dc *datastore.Client, ctx context.Context, q *datastore.Query) error {
	q = q.KeysOnly()
	keys, err := dc.GetAll(ctx, q, nil)
	if err != nil {
		return err
	}
	return dc.DeleteMulti(ctx, keys)
}

func externalIp(obj *compute.Instance) string {
	if len(obj.NetworkInterfaces) > 0 {
		intf := obj.NetworkInterfaces[0]
		if len(intf.AccessConfigs) > 0 {
			return intf.AccessConfigs[0].NatIP
		}
		return intf.NetworkIP
	}

	return ""
}

func getGcpRegion(zone string) string {
	return zone[0 : len(zone)-2]
}

func ditermineMachineType(nodes int) string {
	return "n1-standard-1"
}

func loadTmpl(name string) string {
	_, filename, _, _ := runtime.Caller(1)
	dir := path.Join(path.Dir(filename), "templates")

	content, err := ioutil.ReadFile(dir + "/" + name)
	if err != nil {
		log.Fatal(err)
	}

	return string(content)
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

func toJson(object interface{}) string {
	dump, err := json.MarshalIndent(object, " ", "  ")
	if err != nil {
		log.Fatal(fmt.Errorf("dumping json: %v", err))
	}
	return string(dump)
}

func serviceKeyPath(name string) string {
	return filepath.Join(pb.ConfigPath, name, pb.SvaFilename)
}

func serviceKey(path string) []byte {
	value, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(fmt.Errorf("getting service account err: %v", err))
	}

	return value
}

func prompt(s string) {
	r := bufio.NewReader(os.Stdin)
	fmt.Printf("%s\n\nPlease press [ENTER] or Ctrl-C to cancel", s)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		if line == "\n" {
			break
		}
	}
}
