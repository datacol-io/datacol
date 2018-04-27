package aws

import (
	"github.com/datacol-io/datacol/cloud"
	kube "github.com/datacol-io/datacol/k8s"
)

func (l *AwsCloud) CertificateCreate(name, domain, cert, key string) error {
	return kube.UpdateTLSCertificates(l.kubeClient(), l.DeploymentName, name, domain, cert, key, cloud.AwsProvider)
}

func (l *AwsCloud) CertificateDelete(name, domain string) error {
	return kube.DeleteTLSCertificates(l.kubeClient(), l.DeploymentName, name, domain, cloud.AwsProvider)
}