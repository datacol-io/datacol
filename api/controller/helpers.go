package controller

import (
  "net/http"
  "encoding/json"
  "crypto/rand"
  "math/big"
)

var idAlphabet = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")

func generateId(prefix string, size int) string {
  b := make([]rune, size)
  for i := range b {
    idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(idAlphabet))))
    if err != nil {
      panic(err)
    }
    b[i] = idAlphabet[idx.Int64()]
  }
  return prefix + string(b)
}

func RenderJson(rw http.ResponseWriter, model interface{}){
  if err := json.NewEncoder(rw).Encode(model); err != nil {
    http.Error(rw, err.Error(), http.StatusInternalServerError)
  }
  rw.Header().Set("Content-Type", "application/json")
}

