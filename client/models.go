package client

import (
  "errors"
  "encoding/json"
)

var (
  stackBxName []byte
  ErrNotFound = errors.New("key not found")
)

func init(){
  stackBxName = []byte("stacks")
}

type Stack struct {
  Name      string
  Bucket    string
  Zone      string
  ServiceKey []byte
  ProjectId     string
}

func (st *Stack) Persist() error {
  return Persist(stackBxName, st.Name, st)
}

func FindStack(name string) (*Stack, error) {
  var instance Stack
  v, err := getV(stackBxName, []byte(name))
  if err != nil { return nil, err }

  if err = json.Unmarshal(v, &instance); err != nil { 
    return nil, err
  }

  return &instance, nil
}

