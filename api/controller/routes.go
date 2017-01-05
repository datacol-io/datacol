package controller

import (
  "net/http"

  "github.com/gorilla/mux"
)

func NewRouter() http.Handler {
  router := mux.NewRouter()

  router.HandleFunc("/apps", AppIndex).Methods("GET")
  router.HandleFunc("/app/build", AppBuildCreate).Methods("POST")

  return router
}
