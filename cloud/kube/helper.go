package kube

import (
	"encoding/json"
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
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
