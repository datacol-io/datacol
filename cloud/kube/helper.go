package kube

import (
	"encoding/json"
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"k8s.io/api/extensions/v1beta1"
)

func toJson(object interface{}) string {
	dump, err := json.MarshalIndent(object, " ", "  ")
	if err != nil {
		log.Fatal(fmt.Errorf("dumping json: %v", err))
	}
	return string(dump)
}

func MergeAppDomains(domains []string, item string) []string {
	if item == "" {
		return domains
	}

	itemIndex := -1
	dotted := strings.HasPrefix(item, ":")

	if dotted {
		item = item[1:]
	}

	for i, d := range domains {
		if d == item {
			itemIndex = i
			break
		}
	}

	if dotted && itemIndex >= 0 {
		return append(domains[0:itemIndex], domains[itemIndex+1:]...)
	}

	if !dotted && itemIndex == -1 {
		return append(domains, item)
	}

	return domains
}

func mergeIngressRules(dest *v1beta1.Ingress, source *v1beta1.Ingress) *v1beta1.Ingress {
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

	log.Debugf("ingress rules %s", toJson(dest.Spec))

	return dest
}
