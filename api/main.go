package main

import (
  "net/http"
  "log"
  "os"
  "fmt"

  "github.com/dinesh/rz/api/controller"
)

func main() {
  port := os.Getenv("PORT")
  if port == "" {
    port = "8000"
  }

  router := controller.NewRouter()
  fmt.Printf("Starting api at port %s ...\n", port)
  
  log.Fatal(http.ListenAndServe(":" + port, router))
}