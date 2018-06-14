package store

import (
	"fmt"
	"strconv"

	pb "github.com/datacol-io/datacol/api/models"
	"k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	releaseComponent = "releases"
)

func (s *SecretStore) ReleaseSave(r *pb.Release) error {
	if r.Id == "" {
		r.Id = generateId("R", 8)
	}

	name := releasekey(r.Id)

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				componentKey: releaseComponent,
				stackLabel:   s.Stack,
				managedBy:    heritage,
				appLabelKey:  r.App,
			},
		},
		StringData: map[string]string{
			"id":       r.Id,
			"app":      r.App,
			"status":   r.Status,
			"build_id": r.BuildId,
		},
	}

	if r.Version > 0 {
		secret.StringData["version"] = string(r.Version)
	}

	_, err := s.Client.Core().Secrets(s.Namespace).Update(secret)
	if err != nil {
		if kerrors.IsNotFound(err) {
			_, err = s.Client.Core().Secrets(s.Namespace).Create(secret)
		}
	}

	return err
}

func (s *SecretStore) ReleaseGet(id string) (*pb.Release, error) {
	name := releasekey(id)
	secret, err := s.Client.Core().Secrets(s.Namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return newReleaseFromSecret(*secret), nil
}

func (s *SecretStore) ReleaseList(app string, limit int64) (pb.Releases, error) {
	scope := s.scoped(buildComponent)
	scope = fmt.Sprintf("%s,%s=%s", scope, appLabelKey, app)

	lo := metav1.ListOptions{LabelSelector: scope}
	if limit > 0 {
		lo.Limit = limit
	}

	secretList, err := s.Client.CoreV1().Secrets(s.Namespace).List(lo)
	if err != nil {
		return nil, err
	}

	items := make(pb.Releases, 0, len(secretList.Items))

	for _, secret := range secretList.Items {
		items = append(items, newReleaseFromSecret(secret))
	}

	return items, nil
}

func (s *SecretStore) ReleaseCount(name string) int64 {
	items, _ := s.ReleaseList(name, -1)
	return int64(len(items))
}

func (s *SecretStore) ReleaseDelete(app, id string) error {
	name := releasekey(id)
	err := s.Client.Core().Secrets(s.Namespace).Delete(name, &metav1.DeleteOptions{})
	return err
}

func newReleaseFromSecret(secret v1.Secret) *pb.Release {
	sv := SecretValues(secret.Data)

	r := &pb.Release{
		Id:        sv.String("id"),
		Status:    sv.String("status"),
		App:       sv.String("app"),
		BuildId:   sv.String("build_id"),
		CreatedAt: toTimestamp(secret.CreationTimestamp),
	}

	if version := sv.String("version"); version != "" {
		value, _ := strconv.ParseInt(version, 10, 0)
		r.Version = value
	}

	return r
}

func releasekey(key string) string {
	return fmt.Sprintf("release-%s", key)
}
