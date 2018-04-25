package kube

import (
	"errors"
	"fmt"
	"strings"

	"github.com/datacol-io/datacol/cloud"
	core_v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func UpdateTLSCertificates(c *kubernetes.Clientset, ns, app, domain, cert, key string, provider cloud.CloudProvider) (err error) {
	if provider == cloud.GCPProvider {
		return errors.New("TLS certificates are not implemented for GCP yet.")
	}

	ingName := ingressName(ns)
	secretName := fmt.Sprintf("%s-%s", app, strings.Replace(domain, ".", "-", -1))

	ing, err := c.Extensions().Ingresses(ns).Get(ingName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	foundAt, tlsSpec := -1, ing.Spec.TLS
	for i, tls := range tlsSpec {
		for _, host := range tls.Hosts {
			if host == domain {
				tls.SecretName = secretName
				foundAt = i
			}
		}
	}

	// It won't support multiple domains with common certificates
	if foundAt >= 0 && len(tlsSpec[foundAt].Hosts) > 1 {
		return fmt.Errorf("We only support single certificate per domain. We found multiple domains: %v", tlsSpec[foundAt].Hosts)
	}

	if foundAt < 0 {
		tlsSpec = append(tlsSpec, v1beta1.IngressTLS{
			Hosts:      []string{domain},
			SecretName: secretName,
		})
	}

	ing.Spec.TLS = tlsSpec

	secret := &core_v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
			Labels: map[string]string{
				appLabel:  app,
				managedBy: heritage,
			},
		},
		Type: core_v1.SecretTypeOpaque,
		Data: map[string][]byte{
			"tls.crt": []byte(cert),
			"tls.key": []byte(key),
		},
	}

	if _, err = c.Core().Secrets(ns).Create(secret); err != nil {
		if !kerrors.IsAlreadyExists(err) {
			return err
		}

		_, err = c.Core().Secrets(ns).Update(secret)
	}

	_, err = c.Extensions().Ingresses(ns).Update(ing)

	return err
}

func DeleteTLSCertificates(c *kubernetes.Clientset, ns, app, domain string, provider cloud.CloudProvider) error {
	if provider == cloud.GCPProvider {
		return errors.New("TLS certificates are not implemented for GCP yet.")
	}

	ingName := ingressName(ns)
	secretName := fmt.Sprintf("%s-%s", app, strings.Replace(domain, ".", "-", -1))

	ing, err := c.Extensions().Ingresses(ns).Get(ingName, metav1.GetOptions{})
	if ing != nil {
		tlsSpec, foundAt := ing.Spec.TLS, -1
		for i, t := range tlsSpec {
			if t.SecretName == secretName {
				foundAt = i
			}
		}

		if foundAt >= 0 {
			tlsSpec = append(tlsSpec[:foundAt], tlsSpec[foundAt+1:]...)
		}

		ing.Spec.TLS = tlsSpec
		if _, err = c.Extensions().Ingresses(ns).Update(ing); err != nil {
			return err
		}
	}

	err = c.Core().Secrets(ns).Delete(secretName, &metav1.DeleteOptions{})

	return err
}
