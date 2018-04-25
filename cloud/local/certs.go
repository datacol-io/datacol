package local

import (
	"github.com/datacol-io/datacol/cloud"
	kube "github.com/datacol-io/datacol/k8s"
)

func (l *LocalCloud) CertificateCreate(name, domain, cert, key string) error {
	return kube.UpdateTLSCertificates(l.kubeClient(), l.Name, name, domain, cert, key, cloud.LocalProvider)
}

func (l *LocalCloud) CertificateDelete(name, domain string) error {
	return kube.DeleteTLSCertificates(l.kubeClient(), l.Name, name, domain, cloud.LocalProvider)
}
