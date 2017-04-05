package client

import (
  "encoding/json"
  "github.com/dinesh/datacol/client/models"
)

var rs_bucket = []byte("resources")

func (c *Client) GetResource(name string) (*models.Resource, error) {
  item, err := getV(rs_bucket, []byte(name))
  if err != nil {
    return nil, err
  }

  var r models.Resource
  if err := json.Unmarshal(item, &r); err != nil { 
    return nil, err
  }

  return &r, nil
}

func (c *Client) CreateResource(kind string, options map[string]string) (*models.Resource, error) {
  rs, err := c.Provider().ResourceCreate(options["name"], kind, options)
  if err != nil {
    return nil, err
  }

  if err := Persist(rs_bucket, rs.Name, rs); err != nil {
    return nil, err
  }

  return rs, nil
}
