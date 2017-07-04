package google

import (
	"cloud.google.com/go/datastore"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"cloud.google.com/go/compute/metadata"
	log "github.com/Sirupsen/logrus"
	oauth2_google "golang.org/x/oauth2/google"
	"google.golang.org/api/cloudbuild/v1"
	csm "google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/container/v1"
	"google.golang.org/api/deploymentmanager/v2"
	iam "google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
	"google.golang.org/api/sqladmin/v1beta4"
	"google.golang.org/api/storage/v1"

	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	pb "github.com/dinesh/datacol/api/models"
	gcp "github.com/dinesh/datacol/cmd/provider/gcp"
	"google.golang.org/grpc"
)

var _jwtClient *http.Client

type GCPCloud struct {
	Project        string
	ProjectNumber  string
	DeploymentName string
	BucketName     string
	Zone           string
}


func (g *GCPCloud) storage() *storage.Service {
	return storageService(g.DeploymentName)
}

func storageService(name string) *storage.Service {
	svc, err := storage.New(httpClient(name))
	if err != nil {
		log.Fatal(fmt.Errorf("storage client %s", err))
	}

	return svc
}

func computeService(name string) *compute.Service {
	svc, err := compute.New(httpClient(name))
	if err != nil {
		log.Fatal(fmt.Errorf("compute client %s", err))
	}

	return svc
}

func (g *GCPCloud) cloudbuilder() *cloudbuild.Service {
	svc, err := cloudbuild.New(httpClient(g.DeploymentName))
	if err != nil {
		log.Fatal(fmt.Errorf("cloudbuilder client %s", err))
	}

	return svc
}

func (g *GCPCloud) csmanager() *csm.Service {
	svc, err := csm.New(httpClient(g.DeploymentName))
	if err != nil {
		log.Fatal(fmt.Errorf("cloudresourcemanager client %s", err))
	}

	return svc
}

func (g *GCPCloud) deploymentmanager() *deploymentmanager.Service {
	return dmService(g.DeploymentName)
}

func dmService(name string) *deploymentmanager.Service {
	svc, err := deploymentmanager.New(httpClient(name))
	if err != nil {
		log.Fatal(fmt.Errorf("deploymentmanager client %s", err))
	}

	return svc
}

func (g *GCPCloud) container() *container.Service {
	svc, err := container.New(httpClient(g.DeploymentName))
	if err != nil {
		log.Fatal(fmt.Errorf("container client %s", err))
	}

	return svc
}

func (g *GCPCloud) sqlAdmin() *sqladmin.Service {
	svc, err := sqladmin.New(httpClient(g.DeploymentName))
	if err != nil {
		log.Fatal(fmt.Errorf("sqlAdmin client %s", err))
	}

	return svc
}

func (g *GCPCloud) iam() *iam.Service {
	svc, err := iam.New(httpClient(g.DeploymentName))
	if err != nil {
		log.Fatal(fmt.Errorf("iam client %s", err))
	}

	return svc
}

func svaPrivateKey(name, project string) ([]byte, error) {
	iamClient, err := iam.New(httpClient(name))
	if err != nil {
		return []byte(""), err
	}

	_, c, err := gcp.NewServiceAccountPrivateKey(iamClient, name, project)
	return c, err
}

func (g *GCPCloud) datastore() *datastore.Client {
	dc, _ := datastoreClient(g.DeploymentName, g.Project)
	return dc
}

func datastoreClient(name, project string) (*datastore.Client, func()) {
	opts := []option.ClientOption{
		option.WithGRPCDialOption(grpc.WithBackoffMaxDelay(5 * time.Second)),
		option.WithGRPCDialOption(grpc.WithTimeout(30 * time.Second)),
	}

	if !metadata.OnGCE() {
		opts = append(opts, option.WithServiceAccountFile(service_key_path(name)))
	}

	client, err := datastore.NewClient(context.TODO(), project, opts...)
	if err != nil {
		log.Fatal(fmt.Errorf("datastore client %s", err))
	}

	return client, client.Close
}

func (g *GCPCloud) getCluster(name string) (*container.Cluster, error) {
	service := g.container()
	return service.Projects.Zones.Clusters.Get(g.Project, g.Zone, name).Do()
}

func (g *GCPCloud) ctxNS() context.Context {
	return datastore.WithNamespace(context.TODO(), g.DeploymentName)
}

func httpClient(name string) *http.Client {
	if !metadata.OnGCE() {
		return jwtClient(service_key(name))
	}

	htx, err := oauth2_google.DefaultClient(
		context.TODO(),
		csm.CloudPlatformScope,
		sqladmin.SqlserviceAdminScope,
	)

	if err != nil {
		log.Fatal(fmt.Errorf("failed to create http client err:%v", err))
	}

	return htx
}

func jwtClient(sva []byte) *http.Client {
	if _jwtClient != nil {
		return _jwtClient
	}

	jwtConfig, err := oauth2_google.JWTConfigFromJSON(sva, csm.CloudPlatformScope, sqladmin.SqlserviceAdminScope)
	if err != nil {
		log.Fatal(fmt.Errorf("JWT client %s", err))
	}

	_jwtClient = jwtConfig.Client(context.TODO())
	return _jwtClient
}

func (g *GCPCloud) gsGet(bucket, key string) ([]byte, error) {
	service := g.storage()
	resp, err := service.Objects.Get(bucket, key).Download()
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func (g *GCPCloud) gsPut(bucket, key string, body io.Reader) error {
	service := g.storage()
	_, err := service.Objects.Insert(bucket, &storage.Object{Name: key}).Media(body).Do()
	return err
}

func getKubeClientset(name string) (*kubernetes.Clientset, error) {
	var config *rest.Config
	if metadata.OnGCE() {
		c, err := clientcmd.BuildConfigFromFlags("", "/opt/datacol/kubeconfig")
		if err != nil {
			return nil, err
		}
		config = c
	} else {
		if err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", service_key_path(name)); err != nil {
			return nil, err
		}

		c, err := clientcmd.BuildConfigFromFlags("", kubecfgPath(name))
		if err != nil {
			return nil, err
		}
		config = c
	}

	c, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("cluster connection %v", err)
	}

	return c, nil
}

func externalIp(obj *compute.Instance) string {
	log.Debugf(toJson(obj))

	if len(obj.NetworkInterfaces) > 0 {
		intf := obj.NetworkInterfaces[0]
		if len(intf.AccessConfigs) > 0 {
			return intf.AccessConfigs[0].NatIP
		}
		return intf.NetworkIP
	}

	return ""
}

func service_key_path(name string) string {
	return filepath.Join(pb.ConfigPath, name, pb.SvaFilename)
}

var svkeys = map[string][]byte{}

func service_key(name string) []byte {
	if value, ok := svkeys[name]; ok {
		return value
	}

	value, err := ioutil.ReadFile(service_key_path(name))
	if err != nil {
		log.Fatal(err)
	}
	svkeys[name] = value

	return value
}
