package google

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/appscode/go/crypto/rand"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"

	log "github.com/Sirupsen/logrus"
	pb "github.com/datacol-io/datacol/api/models"
)

func deleteFromQuery(dc *datastore.Client, ctx context.Context, q *datastore.Query) error {
	q = q.KeysOnly()
	keys, err := dc.GetAll(ctx, q, nil)
	if err != nil {
		return err
	}
	return dc.DeleteMulti(ctx, keys)
}

func nameKey(kind, name, ns string) (context.Context, *datastore.Key) {
	return context.Background(), &datastore.Key{
		Kind:      kind,
		Name:      name,
		Namespace: ns,
	}
}

func ditermineMachineType(nodes int) string {
	return "n1-standard-1"
}

func kubecfgPath(name string) string {
	return filepath.Join(pb.ConfigPath, name, "kubeconfig")
}

func timeNow() *timestamp.Timestamp {
	v, _ := ptypes.TimestampProto(time.Now())
	return v
}

func timestampNow() int32 {
	return int32(time.Now().Unix())
}

func getTokenFile(name string) string {
	return filepath.Join(pb.ConfigPath, name, pb.SvaFilename)
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

func buildTemplate(kind string, opts interface{}) (ret string, err error) {
	content, err := Asset(fmt.Sprintf("cloud/google/templates/%s.tmpl", kind))
	if err != nil {
		return ret, err
	}

	tmpl, err := template.New("ct").Parse(string(content))
	if err != nil {
		return ret, err
	}

	var doc bytes.Buffer
	if err := tmpl.Execute(&doc, opts); err != nil {
		return ret, err
	}

	return doc.String(), nil
}

func toJson(object interface{}) string {
	dump, err := json.MarshalIndent(object, " ", "  ")
	if err != nil {
		log.Fatal(fmt.Errorf("dumping json: %v", err))
	}
	return string(dump)
}

func getGcpRegion(zone string) string {
	return zone[0 : len(zone)-2]
}

func generateId(prefix string, size int) string {
	return prefix + "-" + rand.Characters(size)
}

func generatePassword() string {
	return rand.GeneratePassword()
}

func dpToResourceType(dpname, name string) string {
	kind := "unknown"

	switch dpname {
	case "storage.v1.bucket":
		kind = "gs"
	case "container.v1.cluster":
		kind = "cluster"
	case "sqladmin.v1beta4.instance":
		kind = strings.Split(name, "-")[0]
	}

	return kind
}

func rsVarToMap(source []*pb.ResourceVar) map[string]string {
	dst := make(map[string]string)
	for _, r := range source {
		dst[r.Key] = r.Value
	}

	return dst
}
