package kube

import (
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/datacol-io/datacol/cloud"
	core_v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func UpdateTLSCertificates(c *kubernetes.Clientset, ns, app, domain, cert, key string, provider cloud.CloudProvider) (err error) {
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
		Type: core_v1.SecretTypeTLS,
		StringData: map[string]string{
			core_v1.TLSCertKey:       cert,
			core_v1.TLSPrivateKeyKey: key,
		},
	}

	if _, err = c.Core().Secrets(ns).Create(secret); err != nil {
		if !kerrors.IsAlreadyExists(err) {
			return err
		}

		log.Infof("Will update the %s secret", secretName)
		log.Debugln(toJson(secret))

		_, err = c.Core().Secrets(ns).Update(secret)
		if err != nil {
			return fmt.Errorf("updating TLS certificates: %v", err)
		}
	}

	_, err = c.Extensions().Ingresses(ns).Update(ing)

	return err
}

func DeleteTLSCertificates(c *kubernetes.Clientset, ns, app, domain string, provider cloud.CloudProvider) error {
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
