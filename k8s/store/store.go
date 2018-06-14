package store

import (
	"github.com/datacol-io/datacol/api/store"
	klabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

var _ store.Store = &SecretStore{}

const (
	// Synonym of table for secret backend
	componentKey = "datacol.io/component"

	// The namespace of Datacol Stack where the object belongs
	// Will be used in ObjectMeta.Labels
	stackLabel = "datacol.io/stack"

	// A identifier for secret origin
	// Will be used in ObjectMeta.Labels
	managedBy = "heritage"

	// Value of managedBy in Labels
	heritage = "datacol.io"

	// The key in k8s secret labels for objects belonging to an app
	appLabelKey = "app"
)

type SecretStore struct {
	Client *kubernetes.Clientset

	// Kubernetes Namespace where secrets will be stored
	Namespace string

	// Datacol Stack name
	Stack string
}

func (s *SecretStore) scoped(component string) string {
	return labelScope(s.Stack, component, map[string]string{
		managedBy:    heritage,
		componentKey: component,
	})
}

func labelScope(stack, component string, extra map[string]string) string {
	labels := map[string]string{
		managedBy:    heritage,
		componentKey: component,
	}

	for key, value := range extra {
		labels[key] = value
	}

	return klabels.Set(labels).AsSelector().String()
}
