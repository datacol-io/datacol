package kube

import (
	"encoding/json"
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func ingressNamespace(parentNS string) string {
	// return ingressDefaultNamespace
	return parentNS
}

func toJson(object interface{}) string {
	dump, err := json.MarshalIndent(object, " ", "  ")
	if err != nil {
		log.Warnf("dumping json: %v", err)
	}

	return string(dump)
}

// Merge ingress rules from different applications since we share an ingress controller among
// different services. `source` will represent the rules for only an app
func mergeIngressRules(dest *v1beta1.Ingress, source *v1beta1.Ingress) *v1beta1.Ingress {
	// check if the app domains can be added
	for _, r := range source.Spec.Rules {
		foundAt := -1
		for i, rr := range dest.Spec.Rules {
			if rr.Host == r.Host {
				foundAt = i
				break
			}
		}

		if foundAt >= 0 {
			dest.Spec.Rules[foundAt] = r
		} else {
			dest.Spec.Rules = append(dest.Spec.Rules, r)
		}
	}

	// check if the app domains should be removed
	for i, r := range dest.Spec.Rules {
		backend := r.IngressRuleValue.HTTP.Paths[0].Backend
		serviceID := fmt.Sprintf("%s:%v", backend.ServiceName, backend.ServicePort)
		related, found := false, false

		for _, rr := range source.Spec.Rules {
			bk := rr.IngressRuleValue.HTTP.Paths[0].Backend
			sid := fmt.Sprintf("%s:%v", bk.ServiceName, bk.ServicePort)

			if serviceID == sid {
				related = true
			}

			if rr.Host == r.Host {
				found = true
				break
			}
		}

		if related && !found {
			dest.Spec.Rules = append(dest.Spec.Rules[:i], dest.Spec.Rules[i+1:]...)
		}
	}

	log.Debugf("final ingress rules %s", toJson(dest.Spec.Rules))

	return dest
}

// mergeResourceConstraints sets the limit and requests values for cpu, memory for a container
func mergeResourceConstraints(resourceType v1.ResourceName, container *v1.Container, reqLimit string) error {
	resourceLimit, resourceRequest := container.Resources.Limits, container.Resources.Requests

	if reqLimit != "" {
		parts := strings.Split(reqLimit, "/")
		var request, limit string
		if len(parts) == 2 {
			request = parts[0]
			limit = parts[1]
		} else {
			limit = parts[0]
		}

		if request != "" {
			value, err := resource.ParseQuantity(request)
			if err != nil {
				return err
			}

			resourceRequest[resourceType] = value
		}

		if limit != "" {
			value, err := resource.ParseQuantity(limit)
			if err != nil {
				return err
			}

			resourceLimit[resourceType] = value
		}
	}

	return nil
}
