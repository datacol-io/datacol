package main

import (
	"github.com/datacol-io/datacol/cmd"
)

var (
	version string
	rbToken string
)

func main() {
	cmd.Initialize(version, rbToken)
}
