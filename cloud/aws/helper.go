package aws

import (
	"github.com/appscode/go/crypto/rand"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"time"
)

func generateId(prefix string, size int) string {
	return prefix + "-" + rand.Characters(size)
}

func coalesce(s *dynamodb.AttributeValue, def string) string {
	if s != nil {
		return *s.S
	} else {
		return def
	}
}

func timestampNow() int32 {
	return int32(time.Now().Unix())
}
