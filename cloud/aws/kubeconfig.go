package aws

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	log "github.com/Sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var kubeClient *kubernetes.Clientset
var cacheClientsetOnce sync.Once
var cacheConfigPathOnce sync.Once

var (
	rootPath      = "/opt/datacol"
	kcpath        = filepath.Join(rootPath, "kubeconfig")
	pemPathRE     = filepath.Join(rootPath, "%s.pem")
	privateIpAttr = "MasterPrivateIp"
	bastionIpAttr = "BastionHostPublicIp"
	scpCmd        = "scp -i %s -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null ubuntu@%s:~/kubeconfig %s"
)

func (p *AwsCloud) kubeClient() *kubernetes.Clientset {
	cacheConfigPathOnce.Do(func() {
		p.K8sConfigPath()
	})

	cacheClientsetOnce.Do(func() {
		kube, err := getKubeClientSet(p.DeploymentName)
		if err != nil {
			log.Fatal(err)
		}

		kubeClient = kube
	})

	return kubeClient
}

func (p *AwsCloud) K8sConfigPath() (string, error) {
	if _, err := os.Stat(kcpath); err != nil {
		if os.IsNotExist(err) {
			ipAddr, err := p.masterPrivateIp()
			if err != nil {
				return ipAddr, err
			}

			keyname := fmt.Sprintf(pemPathRE, os.Getenv("DATACOL_KEY_NAME"))
			cmd := fmt.Sprintf(scpCmd, keyname, ipAddr, kcpath)

			log.Debugf("Executing %s", cmd)
			if _, err := exec.Command("/bin/sh", "-c", cmd).Output(); err != nil {
				return "", err
			}
		} else {
			return kcpath, err
		}
	}

	return kcpath, nil
}

func (p *AwsCloud) bastionHostIp() (string, error) {
	return p.stackOutputValue(bastionIpAttr)
}

func (p *AwsCloud) stackOutputValue(attr string) (string, error) {
	s, err := p.describeStack("")
	if err != nil {
		return "", err
	}

	for _, o := range s.Outputs {
		if o.OutputKey != nil && attr == *o.OutputKey {
			return *o.OutputValue, nil
		}
	}

	return "", fmt.Errorf("unable to find %s from stack output", attr)
}

func (p *AwsCloud) masterPrivateIp() (string, error) {
	return p.stackOutputValue(privateIpAttr)
}

func getKubeClientSet(name string) (*kubernetes.Clientset, error) {
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
	return clientcmd.BuildConfigFromFlags("", kcpath)
}
