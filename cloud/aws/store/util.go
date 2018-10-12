package store

import (
	"fmt"
	"strconv"
	"time"

	"github.com/appscode/go/crypto/rand"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func generateId(prefix string, size int) string {
	return prefix + "-" + rand.Characters(size)
}

func timestampNow() int32 {
	return int32(time.Now().Unix())
}

func coalesce(s *dynamodb.AttributeValue, def string) string {
	if s != nil {
		return *s.S
	} else {
		return def
	}
}

func coalesceInt(s *dynamodb.AttributeValue, def int) int {
	if s != nil {
		num, _ := strconv.Atoi(*s.N)
		return num
	} else {
		return def
	}
}

func coalesceBytes(s *dynamodb.AttributeValue) (data []byte) {
	if s != nil {
		return s.B
	} else {
		return data
	}
}

func stackNameForApp(a string) string {
	return fmt.Sprintf("app-%s", a)
}

func cfNameForApp(a, b string) string {
	return fmt.Sprintf("%s-app-%s", a, b)
}
