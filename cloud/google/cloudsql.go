package google

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	sql "google.golang.org/api/sqladmin/v1beta4"
)

func (g *GCPCloud) createSqlUser(user, password, instance string) error {
	service := g.sqlAdmin()
	log.Debugf("\ncreating user with %s:%s on database instance %s", user, password, instance)

	op, err := service.Users.Insert(g.Project, instance, &sql.User{
		Name:     user,
		Password: password,
	}).Do()

	if err != nil {
		return err
	}

	if err = waitForSqlOp(service, op, g.Project); err != nil {
		return fmt.Errorf("Error, failure waiting %s:[%s] err:%v", op.OperationType, op.Name, err)
	}

	return nil
}

func (g *GCPCloud) getSqlInstance(name string) (*sql.DatabaseInstance, error) {
	return g.sqlAdmin().Instances.Get(g.Project, name).Do()
}

func (g *GCPCloud) createSqlDatabase(name, instance string) error {
	if _, err := g.sqlAdmin().Databases.Insert(g.Project, instance, &sql.Database{
		Name:     name,
		Project:  g.Project,
		Instance: instance,
	}).Do(); err != nil {
		return err
	}

	return nil
}
