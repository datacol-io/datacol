package store

import (
	"strings"

	"github.com/appscode/go/crypto/rand"
)

// delimeter is used to join array fields while saving things into k8s secret
const delimeter = ","

// SecretValues provides accessor methods for secrets.
type SecretValues map[string][]byte

// Bytes returns the value in the map for the provided key.
func (sv SecretValues) Bytes(key string) []byte {
	return sv[key]
}

// Bytes returns the string value in the map for the provided key.
func (sv SecretValues) String(key string) string {
	return string(sv.Bytes(key))
}

// Bytes returns the string value in the map for the provided key.
func (sv SecretValues) Array(key string) []string {
	return strings.Split(string(sv.Bytes(key)), delimeter)
}

func generateId(prefix string, size int) string {
	return strings.ToLower(prefix + rand.Characters(size))
}
