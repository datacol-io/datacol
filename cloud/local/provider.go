package local

import (
	"fmt"
	"os"
	"sync"

	log "github.com/Sirupsen/logrus"
	pb "github.com/dinesh/datacol/api/models"
	docker "github.com/fsouza/go-dockerclient"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type LocalCloud struct {
	Name            string
	Apps            pb.Apps
	Builds          pb.Builds
	Releases        pb.Releases
	EnvMap          map[string]pb.Environment
	RegistryAddress string
}

var cacheClientsetOnce sync.Once
var kubeClient *kubernetes.Clientset

func (g *LocalCloud) kubeClient() *kubernetes.Clientset {
	cacheClientsetOnce.Do(func() {
		kube, err := getKubeClientset(g.Name)
		if err != nil {
			log.Fatal(err)
		}

		kubeClient = kube
	})

	return kubeClient
}

var dkrOnce sync.Once
var dkrClient *docker.Client

func dockerClient() *docker.Client {
	dkrOnce.Do(func() {
		client, err := docker.NewClientFromEnv()
		if err != nil {
			log.Fatalf("failed to initiate docker client: %v", err)
		}

		if err := client.Ping(); err != nil {
			log.Errorf("Docker ping failed: %v", err)
		}
		dkrClient = client
	})

	return dkrClient
}

func getKubeClientset(name string) (*kubernetes.Clientset, error) {
	config, err := getKubeClientConfig(name)
	if err != nil {
		return nil, err
	}

	c, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("cluster connection %v", err)
	}

	return c, nil
}

func getKubeClientConfig(name string) (*rest.Config, error) {
	var config *rest.Config
	c, err := clientcmd.BuildConfigFromFlags("", os.Getenv("HOME")+"/.kube/config")
	if err != nil {
		return nil, err
	}
	config = c

	return config, nil
}
