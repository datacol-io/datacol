package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"os/exec"
	"path/filepath"
)

var (
	rootPath      = "/opt/datacol"
	kcpath        = filepath.Join(rootPath, "kubeconfig")
	pemPathRE     = filepath.Join(rootPath, "datacol-%s-key.pem")
	privateIpAttr = "MasterPrivateIp"
)

func (p *AwsCloud) K8sConfigPath() (string, error) {
	if _, err := os.Stat(kcpath); err != nil {
		if os.IsNotExist(err) {
			ipAddr, err := p.masterPrivateIp()
			if err != nil {
				return ipAddr, err
			}
			cmd := fmt.Sprintf("scp -i %s -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null ubuntu@%s:~/kubeconfig %s", fmt.Sprintf(pemPathRE, p.DeploymentName), ipAddr, kcpath)
			if _, err := exec.Command("/bin/sh", "-c", cmd).Output(); err != nil {
				return "", err
			}
		} else {
			return kcpath, err
		}
	}

	return kcpath, nil
}

func (p *AwsCloud) masterPrivateIp() (string, error) {
	out, err := p.cloudformation().DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: &p.DeploymentName,
	})

	if err != nil {
		return "", err
	}

	for _, s := range out.Stacks {
		for _, o := range s.Outputs {
			if o.OutputKey != nil && privateIpAttr == *o.OutputKey {
				return *o.OutputValue, nil
			}
		}
	}

	return "", fmt.Errorf("unable to find MasterPrivateIp from stack output")
}

func getKubeClientset(name string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kcpath)
	if err != nil {
		return nil, err
	}

	c, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("cluster connection %v", err)
	}

	return c, nil
}
