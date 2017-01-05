package domain

type App struct {
  Name      string `json: "name"`
  ProjectId string "json: `project_id`"
}