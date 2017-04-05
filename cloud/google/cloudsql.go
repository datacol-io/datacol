package google

import (
  "fmt"
  sql "google.golang.org/api/sqladmin/v1beta4"
  log "github.com/Sirupsen/logrus"
)

func (g *GCPCloud) createSqlUser(user, password, instance string) error {
  service := g.sqlAdmin()
  log.Debugf("\nSQL %s:%s on %s", user, password, instance)

  op, err := service.Users.Insert(g.Project, instance, &sql.User{
    Name:     user,
    Password: password,
  }).Do()

  if err != nil { return err }

  if err = waitForSqlOp(service, op, g.Project); err != nil {
    return fmt.Errorf("Error, failure waiting for insertion of %s "+
      "into %s: %s", user, instance, err)
  }

  return nil
}

func (g *GCPCloud) getSqlInstance(name string) (*sql.DatabaseInstance, error) {
  return g.sqlAdmin().Instances.Get(g.Project, name).Do()
}
