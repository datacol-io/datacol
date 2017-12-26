package common

import (
	"bufio"
	"bytes"
	"strings"

	"github.com/appscode/go/crypto/rand"
	pb "github.com/dinesh/datacol/api/models"
)

func LoadEnvironment(data []byte) pb.Environment {
	e := pb.Environment{}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), "=", 2)

		if len(parts) == 2 {
			if key := strings.TrimSpace(parts[0]); key != "" {
				e[key] = parts[1]
			}
		}
	}

	return e
}

func GenerateId(prefix string, size int) string {
	return prefix + "-" + rand.Characters(size)
}
