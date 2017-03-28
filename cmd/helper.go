package main

import (
  "os"
  "regexp"
  "errors"
  "strings"
  "github.com/dinesh/datacol/cmd/stdcli"
  "github.com/dinesh/datacol/client"
)

var (
  crashing = false
  re = regexp.MustCompile("[^a-z0-9]+")
)

func handlePanic(){
  if crashing { return }
  crashing = true

  if rec := recover(); rec != nil {
    err, ok := rec.(error)
    if !ok {
      err = errors.New(rec.(string))
    }

    stdcli.HandlePanicErr(err)
    os.Exit(1)
  }
}

func closeDb(){
  if client.DB != nil {
    client.DB.Close()
  }
}

func slug(s string) string {
  return strings.Trim(re.ReplaceAllString(strings.ToLower(s), "-"), "-")
}