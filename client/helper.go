package client

import (
	"encoding/json"
	"fmt"
	"github.com/appscode/go/crypto/rand"
	"log"
)

func generateId(prefix string, size int) string {
	return prefix + "-" + rand.Characters(size)
}

func toJson(object interface{}) string {
	dump, err := json.MarshalIndent(object, " ", "  ")
	if err != nil {
		log.Fatal(fmt.Errorf("dumping json: %v", err))
	}
	return string(dump)
}

func intMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}
