package google

import (
	"fmt"
	"k8s.io/client-go/kubernetes"
	kerrors "k8s.io/client-go/pkg/api/errors"
	kapi "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"strings"

	log "github.com/Sirupsen/logrus"
	sql "google.golang.org/api/sqladmin/v1beta4"
)

var (
	sqlSecretName         = "cloudsql-secret"
	cloudsqlContainerName = "cloudsql-proxy"
	cloudsqlImage         = "gcr.io/cloudsql-docker/gce-proxy:1.09"
	sqlCredVolName        = "cloudsql-instance-credentials"
)

func (g *GCPCloud) createSqlUser(user, password, instance string) error {
	service := g.sqlAdmin()
	log.Debugf("\ncreating user with %s:%s on database instance %s", user, password, instance)

	op, err := service.Users.Insert(g.Project, instance, &sql.User{
		Name:     user,
		Password: password,
	}).Do()

	if err != nil {
		return err
	}

	if err = waitForSqlOp(service, op, g.Project); err != nil {
		return fmt.Errorf("Error, failure waiting %s:[%s] err:%v", op.OperationType, op.Name, err)
	}

	return nil
}

func (g *GCPCloud) getSqlInstance(name string) (*sql.DatabaseInstance, error) {
	return g.sqlAdmin().Instances.Get(g.Project, name).Do()
}

func (g *GCPCloud) createSqlDatabase(name, instance string) error {
	if _, err := g.sqlAdmin().Databases.Insert(g.Project, instance, &sql.Database{
		Name:     name,
		Project:  g.Project,
		Instance: instance,
	}).Do(); err != nil {
		return err
	}

	return nil
}

func tearCloudProxy(c *kubernetes.Clientset, ns, name, process string) error {
	secretName := fmt.Sprintf("%s-%s", name, sqlSecretName)
	if err := c.Core().Secrets(ns).Delete(secretName, &kapi.DeleteOptions{}); err != nil {
		return nil
	}

	dp, err := c.Extensions().Deployments(ns).Get(name)
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
	vls := []kapi.Volume{}
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
	dp, err := c.Extensions().Deployments(ns).Get(name)
	if err != nil {
		return err
	}

	cred, err := svaPrivateKey(name, project)
	if err != nil {
		return err
	}

	secretName := fmt.Sprintf("%s-%s", name, sqlSecretName)
	if _, err := c.Core().Secrets(ns).Create(&kapi.Secret{
		ObjectMeta: kapi.ObjectMeta{
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

	sqlContainer := kapi.Container{
		Command: []string{"/cloud_sql_proxy", "--dir=/cloudsql",
			fmt.Sprintf("-instances=%s=tcp:%d", options["INSTANCE_NAME"], port),
			"-credential_file=/secrets/cloudsql/credentials.json"},
		Name:            cloudsqlContainerName,
		Image:           cloudsqlImage,
		ImagePullPolicy: "IfNotPresent",
		VolumeMounts: []kapi.VolumeMount{
			kapi.VolumeMount{
				Name:      sqlCredVolName,
				MountPath: "/secrets/cloudsql",
				ReadOnly:  true,
			},
			kapi.VolumeMount{
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
			volumes := []kapi.Volume{
				kapi.Volume{
					Name: sqlCredVolName,
					VolumeSource: kapi.VolumeSource{
						Secret: &kapi.SecretVolumeSource{
							SecretName: secretName,
						},
					},
				},
				kapi.Volume{
					Name: "cloudsql",
					VolumeSource: kapi.VolumeSource{
						EmptyDir: &kapi.EmptyDirVolumeSource{},
					},
				},
			}

			dp.Spec.Template.Spec.Volumes = append(dpvolumes, volumes...)
		}
	}

	return dp
}
