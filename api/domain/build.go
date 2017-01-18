package domain

type Build struct {
  Id string         `json: "id"`
  ProjectId string  `json: "project_id"`
  Release string    `json: "release"`
  Status  string    `json: "status"`
  AppName string
}