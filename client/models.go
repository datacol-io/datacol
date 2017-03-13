package client

import (
  "fmt"
  "encoding/json"
  "github.com/boltdb/bolt"
)

var (
  DB *Database
  bucketName []byte
)

type Database struct {
  store *bolt.DB
}

func (db *Database) Close() {
  if db.store != nil {
    db.store.Close()
  }
}

func init(){
  bucketName = []byte("stacks")
}

type Stack struct {
  Name      string
  ProjectId string
  Bucket    string
  Zone      string
  ServiceKey []byte
}

func (st *Stack) Persist(mk bool) error {
  err := DB.store.Update(func(tx *bolt.Tx) error {
    var bucket *bolt.Bucket
    if mk {
      b, err := tx.CreateBucketIfNotExists(bucketName)
      if err != nil { return err }
      bucket = b
    } else {
      bucket = tx.Bucket(bucketName)
    }

    encoded, err := json.Marshal(st)
    if err != nil {
      return err
    }
    return bucket.Put([]byte(st.Name), encoded)
  })

  return err
}

func (st *Stack) Delete() error {
  return DB.store.Update(func(tx *bolt.Tx) error {
    bucket := tx.Bucket(bucketName)
    return bucket.Delete([]byte(st.Name))
  })
}

func findStack(name string) (*Stack, error) {
  var instance Stack

  err := DB.store.View(func(tx *bolt.Tx) error {
    bucket := tx.Bucket(bucketName)
    v := bucket.Get([]byte(name))
    err := json.Unmarshal(v, &instance)
    if err != nil {
      return fmt.Errorf("find stack: %v", err)
    }
    return nil
  })

  if err != nil {
    return nil, err
  }

  return &instance, nil
}


