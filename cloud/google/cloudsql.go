package google

import (
	"fmt"

	"github.com/dinesh/datacol/cloud"
	"github.com/dinesh/datacol/cloud/kube"
	"k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	klabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

const (
	managedBy = "managed-by"
	heritage  = "datacol"
	appLabel  = "app"
)

func tearCloudProxy(c *kubernetes.Clientset, ns, app, process string) error {
	secretName := fmt.Sprintf("%s-%s", app, cloud.CloudsqlSecretName)
	if err := c.Core().Secrets(ns).Delete(secretName, &metav1.DeleteOptions{}); err != nil {
		return nil
	}

	listSelector := metav1.ListOptions{
		LabelSelector: klabels.Set(map[string]string{appLabel: app, managedBy: heritage}).AsSelector().String(),
	}

	resp, err := c.Extensions().Deployments(ns).List(listSelector)
	if err != nil {
		return err
	}

	for _, dp := range resp.Items {
		// Delete the cloudsql related sidecar container if present
		kube.PruneCloudSQLManifest(&dp.Spec.Template.Spec)

		if _, err = c.Extensions().Deployments(ns).Update(&dp); err != nil {
			return err
		}
	}

	return nil
}

func setupCloudProxy(c *kubernetes.Clientset, ns, project, app string, options map[string]string) error {
	cred, err := svaPrivateKey(ns, project)
	if err != nil {
		return err
	}

	// if you change the secretName, make sure to change it inside kube.MergeCloudSQLManifest
	secretName := fmt.Sprintf("%s-%s", app, cloud.CloudsqlSecretName)

	if _, err := c.Core().Secrets(ns).Create(&v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
		},
		Data: map[string][]byte{
			"credentials.json": cred,
		},
	}); err != nil {
		if !kerrors.IsAlreadyExists(err) {
			return err
		}
	}

	listSelector := metav1.ListOptions{
		LabelSelector: klabels.Set(map[string]string{appLabel: app, managedBy: heritage}).AsSelector().String(),
	}

	resp, err := c.Extensions().Deployments(ns).List(listSelector)
	if err != nil {
		return err
	}

	for _, dp := range resp.Items {
		// merge clousql sidecar container manifest
		kube.MergeCloudSQLManifest(&dp.Spec.Template.Spec, app, options)

		if _, err = c.Extensions().Deployments(ns).Update(&dp); err != nil {
			return err
		}
	}

	return nil
}
