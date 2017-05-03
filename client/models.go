package client

import (
	"errors"
)

var (
	stackBxName []byte
	ErrNotFound = errors.New("key not found")
)

func init() {
	stackBxName = []byte("stacks")
}
