package store

import (
	"fmt"
	"strings"

	pb "github.com/datacol-io/datacol/api/models"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	appComponent = "apps"
)

func (s *SecretStore) AppCreate(app *pb.App, req *pb.AppCreateOptions) error {
	name := appkey(app.Name)

	if _, err := s.Client.Core().Secrets(s.Namespace).Get(name, metav1.GetOptions{}); err == nil {
		return fmt.Errorf("App:%s already exists. Please choose another name.", name)
	}

	secret := appToSecret(name, s.Stack, app)
	fmt.Printf("creating app %v in %s\n", toJson(secret), s.Namespace)
	_, err := s.Client.Core().Secrets(s.Namespace).Create(secret)

	return err
}

func (s *SecretStore) AppUpdate(app *pb.App) error {
	name := appkey(app.Name)
	_, err := s.Client.Core().Secrets(s.Namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	secret := appToSecret(name, s.Stack, app)
	_, err = s.Client.Core().Secrets(s.Namespace).Update(secret)

	return err
}

func (s *SecretStore) AppGet(name string) (*pb.App, error) {
	name = appkey(name)
	secret, err := s.Client.Core().Secrets(s.Namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return newAppFromSecret(*secret), nil
}

func (s *SecretStore) AppList() (pb.Apps, error) {
	lo := metav1.ListOptions{LabelSelector: s.scoped(appComponent)}

	secretList, err := s.Client.CoreV1().Secrets(s.Namespace).List(lo)
	if err != nil {
		return nil, err
	}

	apps := make(pb.Apps, 0, len(secretList.Items))

	for _, secret := range secretList.Items {
		apps = append(apps, newAppFromSecret(secret))
	}

	return apps, nil
}

func (s *SecretStore) AppDelete(name string) error {
	secretName := appkey(name)
	if err := s.Client.Core().Secrets(s.Namespace).Delete(secretName, &metav1.DeleteOptions{}); err != nil {
		return err
	}

	builds, _ := s.BuildList(name, -1)
	for _, b := range builds {
		s.BuildDelete(b.App, b.Id)
	}

	releases, _ := s.BuildList(name, -1)
	for _, r := range releases {
		s.BuildDelete(r.App, r.Id)
	}

	return nil
}

func newAppFromSecret(secret v1.Secret) *pb.App {
	sv := SecretValues(secret.Data)

	return &pb.App{
		Name:      sv.String("name"),
		Status:    sv.String("status"),
		ReleaseId: sv.String("release_id"),
		RepoUrl:   sv.String("repo_url"),
		BuildId:   sv.String("build_id"),
		Domains:   sv.Array("domains"),
		Resources: sv.Array("resources"),
	}
}

func appToSecret(name, stack string, app *pb.App) *v1.Secret {
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				componentKey: appComponent,
				stackLabel:   stack,
				managedBy:    heritage,
			},
		},
		StringData: map[string]string{
			"name":       app.Name,
			"status":     app.Status,
			"release_id": app.ReleaseId,
			"build_id":   app.BuildId,
			"repo_url":   app.RepoUrl,
			"domains":    strings.Join(app.Domains, delimeter),
			"resources":  strings.Join(app.Resources, delimeter),
		},
	}

	return secret
}

func appkey(key string) string {
	return fmt.Sprintf("app-%s", key)
}
