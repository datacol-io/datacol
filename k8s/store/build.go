package store

import (
	"fmt"
	"sort"

	pb "github.com/datacol-io/datacol/api/models"
	v1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	buildComponent = "builds"
)

func (s *SecretStore) BuildSave(b *pb.Build) error {
	if b.Id == "" {
		b.Id = generateId("B", 8)
	}
	name := buildkey(b.Id)

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				componentKey: buildComponent,
				stackLabel:   s.Stack,
				managedBy:    heritage,
				appLabelKey:  b.App,
			},
		},
		Data: map[string][]byte{
			"procfile": b.Procfile,
		},
		StringData: map[string]string{
			"id":        b.Id,
			"app":       b.App,
			"status":    b.Status,
			"remote_id": b.RemoteId,
			"version":   b.Version,
		},
	}

	_, err := s.Client.Core().Secrets(s.Namespace).Update(secret)
	if err != nil {
		if kerrors.IsNotFound(err) {
			_, err = s.Client.Core().Secrets(s.Namespace).Create(secret)
		}

		return err
	}

	return nil
}

func (s *SecretStore) BuildGet(app, id string) (*pb.Build, error) {
	name := buildkey(id)
	secret, err := s.Client.Core().Secrets(s.Namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return newBuildFromSecret(*secret), nil
}

func (s *SecretStore) BuildList(app string, limit int64) (pb.Builds, error) {
	scope := s.scoped(buildComponent)
	scope = fmt.Sprintf("%s,%s=%s", scope, appLabelKey, app)

	lo := metav1.ListOptions{LabelSelector: scope}

	if limit > 0 {
		// Ignore the limit para for now as I don't know to how k8s order things
		// lo.Limit = limit
	}

	secretList, err := s.Client.CoreV1().Secrets(s.Namespace).List(lo)
	if err != nil {
		return nil, err
	}

	builds := make(pb.Builds, 0, len(secretList.Items))

	for _, secret := range secretList.Items {
		builds = append(builds, newBuildFromSecret(secret))
	}

	sort.Slice(builds, func(i, j int) bool {
		return builds[i].CreatedAt > builds[j].CreatedAt
	})

	if len(builds) < int(limit) {
		limit = int64(len(builds))
	}

	if len(builds) == 0 {
		return builds, nil
	}

	return builds[0:limit], nil
}

func (s *SecretStore) BuildDelete(app, id string) error {
	name := buildkey(id)
	err := s.Client.Core().Secrets(s.Namespace).Delete(name, &metav1.DeleteOptions{})
	return err
}

func newBuildFromSecret(secret v1.Secret) *pb.Build {
	sv := SecretValues(secret.Data)

	return &pb.Build{
		Id:        sv.String("id"),
		Status:    sv.String("status"),
		App:       sv.String("app"),
		RemoteId:  sv.String("remote_id"),
		Version:   sv.String("version"),
		Procfile:  sv.Bytes("procfile"),
		CreatedAt: toTimestamp(secret.ObjectMeta.CreationTimestamp),
	}
}

func buildkey(key string) string {
	return fmt.Sprintf("build-%s", key)
}
