package kube

import "k8s.io/client-go/kubernetes"

func UpdateTLSCertificates(c *kubernetes.Clientset, app, domain string, crt, key []byte) error {

}
