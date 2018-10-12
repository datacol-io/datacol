package store

import (
	"context"
	"fmt"

	"cloud.google.com/go/datastore"
	"github.com/datacol-io/datacol/api/store"
)

var notImplemented = fmt.Errorf("Not Implemented in google/store package yet.")

type DSBackend struct {
	DeploymentName string
	*datastore.Client
}

var _ store.Store = &DSBackend{}

func (g *DSBackend) datastore() *datastore.Client {
	return g.Client
}

func (g *DSBackend) nestedKey(kind, key string) (context.Context, *datastore.Key) {
	return nameKey(kind, key, g.DeploymentName)
}
