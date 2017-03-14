package client

import (
  "errors"
  "encoding/json"
  "github.com/joyrexus/buckets"
)

var (
  DB *buckets.DB
  stackBxName []byte
  ErrNotFound = errors.New("key not found")
)

func DBClose() {
  if DB != nil {
    DB.Close()
  }
}

func init(){
  stackBxName = []byte("stacks")
}

type Stack struct {
  Name      string
  ProjectId string
  Bucket    string
  Zone      string
  ServiceKey []byte
}

func (st *Stack) Persist(mk bool) error {
  encoded, err := json.Marshal(st)
  if err != nil { return err }

  sbx, _ := DB.New(stackBxName)
  return sbx.Put([]byte(st.Name), encoded)
}

func (st *Stack) Delete() error {
  sbx, _ := DB.New(stackBxName)
  return sbx.Delete([]byte(st.Name))
}

func FindStack(name string) (*Stack, error) {
  var instance Stack
  sbx, _ := DB.New(stackBxName)
  v, err := sbx.Get([]byte(name))
  if err != nil { return nil, err }

  if err = json.Unmarshal(v, &instance); err != nil { 
    return nil, err
  }

  return &instance, nil
}

