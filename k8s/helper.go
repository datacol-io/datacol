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

func ingressName(ns string) string {
	//Note: making name dependent on namespace i.e. stackName will only provision one load-balancer per stack
	// change this if you want to allocate individual load balanacer for each app and use Name = payload.Request.ServiceID
	// Also if you change this remember to chage in AppGet to fetch IP of load balancer code
	return fmt.Sprintf("%s-ing", ns)
}

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
	var resourceLimit, resourceRequest v1.ResourceList

	if resourceLimit == nil {
		resourceLimit = make(v1.ResourceList)
	}

	if resourceRequest == nil {
		resourceRequest = make(v1.ResourceList)
	}

	switch reqLimit {
	case "0": // API trying to unset the limits
		resourceLimit = container.Resources.Limits
		resourceRequest = container.Resources.Requests

		delete(resourceLimit, resourceType)
		delete(resourceRequest, resourceType)
	default:
		// API trying to set <limit> or <request>/<limit>
		if reqLimit != "" {
			parts := strings.Split(reqLimit, "/")
			var request, limit string
			if len(parts) == 2 {
				request = parts[0]
				limit = parts[1]
			} else {
				limit = parts[0]
			}

			// Set the non-empty request and limit values
			if request != "" {
				value, err := resource.ParseQuantity(request)
				if err != nil {
					return err
				}

				log.Debugf("setting request value: %+v", value)
				resourceRequest[resourceType] = value
			}

			if limit != "" {
				value, err := resource.ParseQuantity(limit)
				if err != nil {
					return err
				}

				log.Debugf("setting limit value: %+v", value)

				resourceLimit[resourceType] = value
			}
		}
	}

	container.Resources.Limits = resourceLimit
	container.Resources.Requests = resourceRequest

	log.Debugf("merged %s limits: %+v", resourceType, container.Resources)

	return nil
}

func getRequestLimit(rqrmts v1.ResourceRequirements, resource v1.ResourceName) (result string) {
	var req, limit string

	if qty, ok := rqrmts.Requests[resource]; ok {
		req = qty.String()
	}

	if qty, ok := rqrmts.Limits[resource]; ok {
		limit = qty.String()
	}

	if req != "" {
		result += limit
		if limit != "" {
			result += "/"
		}
	}

	if limit != "" {
		result += limit
	}

	return
}
