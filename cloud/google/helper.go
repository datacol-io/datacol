package google

import (
	"bufio"
	"bytes"
	"cloud.google.com/go/datastore"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"math/big"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	pb "github.com/dinesh/datacol/api/models"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
)

func deleteFromQuery(dc *datastore.Client, ctx context.Context, q *datastore.Query) error {
	q = q.KeysOnly()
	keys, err := dc.GetAll(ctx, q, nil)
	if err != nil {
		return err
	}
	return dc.DeleteMulti(ctx, keys)
}

func nameKey(kind, id, ns string) (context.Context, *datastore.Key) {
	ctx := datastore.WithNamespace(context.TODO(), ns)
	return ctx, datastore.NewKey(ctx, kind, id, 0, nil)
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

func loadEnv(data []byte) pb.Environment {
	e := pb.Environment{}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), "=", 2)

		if len(parts) == 2 {
			if key := strings.TrimSpace(parts[0]); key != "" {
				e[key] = parts[1]
			}
		}
	}

	return e
}

func getGcpRegion(zone string) string {
	return zone[0 : len(zone)-2]
}

var idAlphabet = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")

func generateId(prefix string, size int) string {
	b := make([]rune, size)
	for i := range b {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(idAlphabet))))
		if err != nil {
			panic(err)
		}
		b[i] = idAlphabet[idx.Int64()]
	}
	return prefix + string(b)
}

func generatePassword() (string, error) {
	data := make([]byte, 1024)

	if _, err := rand.Read(data); err != nil {
		return "", err
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])[0:30], nil
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
