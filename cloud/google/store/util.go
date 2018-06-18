package store

import (
	"context"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/appscode/go/crypto/rand"
)

func nameKey(kind, name, ns string) (context.Context, *datastore.Key) {
	return context.Background(), &datastore.Key{
		Kind:      kind,
		Name:      name,
		Namespace: ns,
	}
}

func deleteFromQuery(dc *datastore.Client, ctx context.Context, q *datastore.Query) error {
	q = q.KeysOnly()
	keys, err := dc.GetAll(ctx, q, nil)
	if err != nil {
		return err
	}
	return dc.DeleteMulti(ctx, keys)
}

func generateId(prefix string, size int) string {
	return prefix + "-" + rand.Characters(size)
}

func timestampNow() int32 {
	return int32(time.Now().Unix())
}
