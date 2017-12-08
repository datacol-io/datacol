package google

import (
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	sqlSecretName         = "cloudsql-secret"
	cloudsqlContainerName = "cloudsql-proxy"
	cloudsqlImage         = "gcr.io/cloudsql-docker/gce-proxy:1.09"
	sqlCredVolName        = "cloudsql-instance-credentials"
)

func tearCloudProxy(c *kubernetes.Clientset, ns, name, process string) error {
	secretName := fmt.Sprintf("%s-%s", name, sqlSecretName)
	if err := c.Core().Secrets(ns).Delete(secretName, &metav1.DeleteOptions{}); err != nil {
		return nil
	}

	dp, err := c.Extensions().Deployments(ns).Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	found := false
	for i, ctnr := range dp.Spec.Template.Spec.Containers {
		for ctnr.Name == cloudsqlContainerName {
			containers := dp.Spec.Template.Spec.Containers
			dp.Spec.Template.Spec.Containers = append(containers[:i], containers[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("resource %s not liked with %s", process, name)
	}

	found = false
	vls := []v1.Volume{}
	for _, v := range dp.Spec.Template.Spec.Volumes {
		if v.Name != sqlCredVolName && v.Name != "cloudsql" {
			vls = append(vls, v)
		}
	}

	dp.Spec.Template.Spec.Volumes = vls
	if _, err = c.Extensions().Deployments(ns).Update(dp); err != nil {
		return err
	}
	return nil
}

func setupCloudProxy(c *kubernetes.Clientset, ns, project, name string, options map[string]string) error {
	dp, err := c.Extensions().Deployments(ns).Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	cred, err := svaPrivateKey(ns, project)
	if err != nil {
		return err
	}

	secretName := fmt.Sprintf("%s-%s", name, sqlSecretName)
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

	dp = mergeSqlManifest(dp, secretName, options)
	if _, err = c.Extensions().Deployments(ns).Update(dp); err != nil {
		return err
	}
	return nil
}

func mergeSqlManifest(dp *v1beta1.Deployment, secretName string, options map[string]string) *v1beta1.Deployment {
	containers := dp.Spec.Template.Spec.Containers
	parts := strings.Split(options["DATABASE_URL"], "://")
	port := getDefaultPort(parts[0])

	sqlContainer := v1.Container{
		Command: []string{"/cloud_sql_proxy", "--dir=/cloudsql",
			fmt.Sprintf("-instances=%s=tcp:%d", options["INSTANCE_NAME"], port),
			"-credential_file=/secrets/cloudsql/credentials.json"},
		Name:            cloudsqlContainerName,
		Image:           cloudsqlImage,
		ImagePullPolicy: "IfNotPresent",
		VolumeMounts: []v1.VolumeMount{
			v1.VolumeMount{
				Name:      sqlCredVolName,
				MountPath: "/secrets/cloudsql",
				ReadOnly:  true,
			},
			v1.VolumeMount{
				Name:      "cloudsql",
				MountPath: "/cloudsql",
			},
		},
	}

	found := false
	for i, c := range containers {
		if c.Name == cloudsqlContainerName {
			containers[i] = sqlContainer
			found = true
		}
	}

	if !found {
		containers = append(containers, sqlContainer)
	}

	dp.Spec.Template.Spec.Containers = containers

	if !found {
		volfound := false
		dpvolumes := dp.Spec.Template.Spec.Volumes

		for _, v := range dpvolumes {
			if v.Name == sqlCredVolName {
				volfound = true
			}
		}

		if !volfound {
			volumes := []v1.Volume{
				v1.Volume{
					Name: sqlCredVolName,
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: secretName,
						},
					},
				},
				v1.Volume{
					Name: "cloudsql",
					VolumeSource: v1.VolumeSource{
						EmptyDir: &v1.EmptyDirVolumeSource{},
					},
				},
			}

			dp.Spec.Template.Spec.Volumes = append(dpvolumes, volumes...)
		}
	}

	return dp
}

func getDefaultPort(kind string) int {
	var port int
	switch kind {
	case "mysql":
		port = 3306
	case "postgres":
		port = 5432
	default:
		log.Fatal(fmt.Errorf("No default port defined for %s", kind))
	}

	return port
}
