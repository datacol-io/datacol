package main

import (
  "os"
  "fmt"

  "gopkg.in/urfave/cli.v2"
)


func main(){
  builder := func(c *cli.Context) {
    var appDir string
    if c.NArg() > 0 {
      appDir = c.Args().Get(0)
    } else {
      dir, err := os.Getwd()
      if err != nil {
        fmt.Printf(err.Error())
        os.Exit(1)
      }
      appDir = dir
    }


    fmt.Printf("building the %s\n", appDir)
  }

  app := &cli.App {
    Commands: []cli.Command{
      {
        Name: "build",
        UsageText: "build an app",
        Action: builder,
      },
    },
  }

  app.Run(os.Args)
}