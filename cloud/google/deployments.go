package google

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"cloud.google.com/go/datastore"
	log "github.com/Sirupsen/logrus"
	pb "github.com/dinesh/datacol/api/models"
	dm "google.golang.org/api/deploymentmanager/v2"
	"google.golang.org/api/googleapi"
	sql "google.golang.org/api/sqladmin/v1beta4"
)

func (g *GCPCloud) updateDeployment(
	service *dm.Service,
	dp *dm.Deployment,
	manifest *dm.Manifest,
	content string,
) error {
	dp.Target = &dm.TargetConfiguration{
		Config:  &dm.ConfigFile{Content: content},
		Imports: manifest.Imports,
	}

	op, err := service.Deployments.Update(g.Project, g.DeploymentName, dp).Do()
	if err != nil {
		if gerr, ok := err.(*googleapi.Error); ok && gerr.Code == 403 {
			// TODO: better error message
			return err
		}
		return err
	}

	if err = waitForDpOp(service, op, g.Project, false, nil); err != nil {
		return err
	}
	return err
}

func waitForSqlOp(svc *sql.Service, op *sql.Operation, project string) error {
	log.Debugf("Waiting for %s [%v]", op.OperationType, op.Name)

	for {
		time.Sleep(2 * time.Second)
		op, err := svc.Operations.Get(project, op.Name).Do()
		if err != nil {
			return err
		}

		switch op.Status {
		case "PENDING", "RUNNING":
			fmt.Print(".")
			continue
		case "DONE":
			if op.Error != nil {
				var last error
				for _, operr := range op.Error.Errors {
					last = fmt.Errorf("%v", operr.Message)
				}
				// try to teardown if, just ignore error if any
				log.Errorf("sqlAdmin Operation failed: %v, Canceling ..", last)
				return last
			} else {
				return nil
			}
		default:
			return fmt.Errorf("Unknown status %q: %+v", op.Status, op)
		}
	}
}

func (g *GCPCloud) nestedKey(kind, key string) (context.Context, *datastore.Key) {
	return nameKey(kind, key, g.DeploymentName)
}

func waitForDpOp(svc *dm.Service, op *dm.Operation, project string, interrupt bool, teardown func() error) error {
	log.Infof("Waiting on %s [%v]", op.OperationType, op.Name)

	cancelCh := make(chan os.Signal, 1)
	signal.Notify(cancelCh, os.Interrupt, syscall.SIGTERM)

	for {
		time.Sleep(2 * time.Second)
		op, err := svc.Operations.Get(project, op.Name).Do()
		if err != nil {
			return err
		}

		select {
		case <-cancelCh:
			if interrupt {
				return teardown()
			} else {
				return nil
			}
		default:
		}

		switch op.Status {
		case "PENDING", "RUNNING":
			fmt.Print(".")
			continue
		case "DONE":
			if op.Error != nil {
				var last error
				for _, operr := range op.Error.Errors {
					last = fmt.Errorf("%v", operr.Message)
				}
				// try to teardown if, just ignore error if any
				log.Errorf("Deployment failed: %v, Canceling ..", last)

				if interrupt {
					if err := teardown(); err != nil {
						log.Debugf("deleting stack: %+v", err)
					}
				}
				return last
			}
			return nil
		default:
			return fmt.Errorf("Unknown status %q: %+v", op.Status, op)
		}
	}
}

func resetDatabase(name, project string) error {
	s, close := datastoreClient(name, project)
	defer close()

	ctx := datastore.WithNamespace(context.TODO(), name)

	if err := deleteFromQuery(s, ctx, datastore.NewQuery(appKind)); err != nil {
		return fmt.Errorf("deleting apps err: %v", err)
	}

	if err := deleteFromQuery(s, ctx, datastore.NewQuery(buildKind)); err != nil {
		return fmt.Errorf("deleting builds err: %v", err)
	}

	if err := deleteFromQuery(s, ctx, datastore.NewQuery(releaseKind)); err != nil {
		return fmt.Errorf("deleting releases err: %v", err)
	}

	if err := deleteFromQuery(s, ctx, datastore.NewQuery(resourceKind)); err != nil {
		return fmt.Errorf("deleting resources err: %v", err)
	}

	return nil
}

func (g *GCPCloud) resetDatabase() error {
	apps, err := g.AppList()
	if err != nil {
		return err
	}

	store := g.datastore()
	ctx := g.ctxNS()

	// delete apps, builds, releases
	for _, app := range apps {
		if err := g.deleteAppFromDatastore(app.Name); err != nil {
			return err
		}
	}

	// delete resources
	q := datastore.NewQuery(resourceKind).KeysOnly()
	return deleteFromQuery(store, ctx, q)
}

func getManifest(service *dm.Service, project, stack string) (*dm.Deployment, *dm.Manifest, error) {
	dp, err := service.Deployments.Get(project, stack).Do()
	if err != nil {
		return nil, nil, err
	}

	parts := strings.Split(dp.Manifest, "/")
	mname := parts[len(parts)-1]
	m, err := service.Manifests.Get(project, stack, mname).Do()

	return dp, m, err
}

func resourceFromStack(service *dm.Service, project, stack, name string) (*pb.Resource, error) {
	return &pb.Resource{Name: name}, nil
}
