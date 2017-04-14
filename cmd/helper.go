package main

import (
  "os"
  "log"
  "fmt"
  "bufio"
  "regexp"
  "errors"
  "strings"
  "github.com/dinesh/datacol/cmd/stdcli"
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

func confirm(s string, tries int) bool {
  r := bufio.NewReader(os.Stdin)

  for ; tries > 0; tries-- {
    fmt.Printf("%s [y/n]: ", s)

    res, err := r.ReadString('\n')
    if err != nil {
      log.Fatal(err)
    }

    // Empty input (i.e. "\n")
    if len(res) < 2 {
      continue
    }

    return strings.ToLower(strings.TrimSpace(res))[0] == 'y'
  }

  return false
}

func slug(s string) string {
  return strings.Trim(re.ReplaceAllString(strings.ToLower(s), "-"), "-")
}

func consoleURL(api, pid string) string {
  return fmt.Sprintf("https://console.developers.google.com/apis/api/%s/overview?project=%s", api, pid)
}