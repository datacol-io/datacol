package models

import (
  "time"
)

type Stack struct {
  Name      string
  ProjectId string
  Bucket    string
  Zone      string
  ServiceKey []byte
}

type App struct {
  Name    string    `json: "name"`
  Status  string    `json: "status"`
  Release string    `json: "release"`
  Stack   string    `json: "stack"`
}

type Apps []*App

type Build struct {
  Id  string            `json: "id"`
  App string            `json: "app"`
  Status string         `json: "status"`
  CreatedAt time.Time   `json: "created_at"`
}

type Builds []*Build

type Release struct {
  Id      string  `json: "id"`
  App     string  `json: "app"`
  BuildId string  `json: "buildId"`
  Status  string  `json: "status"`
  CreatedAt time.Time `json: "created_at"`
}

type Releases []*Release

type BuildOptions struct {
  Id  string
  Key string
}

type Environment map[string]string

type LogStreamOptions struct {
  Follow bool
  Since  time.Duration
}
