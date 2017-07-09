package kube

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
)

func toJson(object interface{}) string {
	dump, err := json.MarshalIndent(object, " ", "  ")
	if err != nil {
		log.Fatal(fmt.Errorf("dumping json: %v", err))
	}
	return string(dump)
}
