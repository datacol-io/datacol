package google

import (
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	pb "github.com/datacol-io/datacol/api/models"
	dm "google.golang.org/api/deploymentmanager/v2"
	"gopkg.in/yaml.v2"
)

const (
	databaseName     = "app"
	systemName       = "datacol"
	stackLabelKey    = "stack"
	resourceLabelKey = "resource"
)

type dmResource struct {
	Name       string                 `yaml:"name"`
	Type       string                 `yaml:"type"`
	Properties map[string]interface{} `yaml:"properties"`
}

type dmOutput struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type manifestConfig struct {
	Resources []dmResource `yaml:"resources"`
	Outputs   []dmOutput   `yaml:"outputs"`
}

func (g *GCPCloud) ResourceGet(name string) (*pb.Resource, error) {
	dpName := fmt.Sprintf("%s-%s", g.DeploymentName, name)
	dp, manifest, err := g.describeDeployment(dpName)
	if err != nil {
		return nil, err
	}

	rs, err := g.resourceFromDeployment(dp, manifest, g.DeploymentName)
	if err != nil {
		return nil, err
	}

	if rs.Tags[stackLabelKey] != "" && rs.Tags[stackLabelKey] != g.DeploymentName {
		return nil, fmt.Errorf("no such deployment in this stack: %s", name)
	}

	switch rs.Kind {
	case "mysql":
		rs.Exports["DATABASE_URL"] = fmt.Sprintf("mysql://%s:%s@%s:%s/%s", rs.Outputs["EnvMysqlUsername"], rs.Outputs["EnvMysqlPassword"], "127.0.0.1", "3306", rs.Outputs["EnvMysqlDatabase"])
		rs.Exports["INSTANCE_NAME"] = fmt.Sprintf("%s:%s:%s", g.Project, g.Region, name)
	case "postgres":
		rs.Exports["DATABASE_URL"] = fmt.Sprintf("postgres://%s:%s@%s:%s/%s", rs.Outputs["EnvPostgresUsername"], rs.Outputs["EnvPostgresPassword"], "127.0.0.1", "5432", rs.Outputs["EnvPostgresDatabase"])
		rs.Exports["INSTANCE_NAME"] = fmt.Sprintf("%s:%s:%s", g.Project, g.Region, name)
	}

	return rs, nil
}

func (g *GCPCloud) ResourceDelete(name string) error {
	dpName := fmt.Sprintf("%s-%s", g.DeploymentName, name)
	_, err := g.deploymentmanager().Deployments.Delete(g.Project, dpName).Do()
	return err
}

func (g *GCPCloud) ResourceList() (pb.Resources, error) {
	return g.resourceListFromStack()
}

func (g *GCPCloud) resourceListFromStack() (pb.Resources, error) {
	resp := pb.Resources{}
	service := g.deploymentmanager()

	out, err := service.Deployments.List(g.Project).Do()
	if err != nil {
		return resp, err
	}

	for _, dp := range out.Deployments {
		tags := deploymentLabels(dp)

		if tags[stackLabelKey] != g.DeploymentName || tags["system"] != systemName {
			continue
		}

		mc, err := fetchManifest(service, g.Project, dp.Name, dp.Manifest)
		if err != nil {
			return resp, err
		}
		rs, err := g.resourceFromDeployment(dp, mc, g.DeploymentName)
		if err != nil {
			return resp, err
		}
		resp = append(resp, rs)
	}

	return resp, nil
}

func (g *GCPCloud) ResourceCreate(name, kind string, params map[string]string) (*pb.Resource, error) {
	rs := &pb.Resource{Name: name, Kind: kind}
	params["project"] = g.Project

	if kind == "postgres" {
		if _, ok := params["tier"]; !ok {
			c, k := params["cpu"]
			if !k {
				return nil, fmt.Errorf("missing param `--cpu`")
			}

			m, k := params["memory"]
			if !k {
				return nil, fmt.Errorf("missing param `--memory`")
			}

			params["tier"] = fmt.Sprintf("db-custom-%s-%s", c, m)
		}
	}

	switch kind {
	case "mysql", "postgres":
		params["region"] = g.Region
		params["zone"] = g.DefaultZone
		params["database"] = databaseName
		params["password"] = generatePassword()
		params["username"] = kind
	default:
		return nil, fmt.Errorf("%s is not supported yet.", kind)
	}

	sqlj2, err := buildTemplate(kind, params)
	if err != nil {
		return nil, fmt.Errorf("Parsing template err: %v", err)
	}

	log.Debugf("creating resource with: %s", sqlj2)

	_, err = g.deploymentmanager().Deployments.Insert(g.Project, &dm.Deployment{
		Name: fmt.Sprintf("%s-%s", g.DeploymentName, name),
		Target: &dm.TargetConfiguration{
			Config: &dm.ConfigFile{Content: sqlj2},
		},
		Labels: []*dm.DeploymentLabelEntry{
			&dm.DeploymentLabelEntry{Key: "name", Value: name},
			&dm.DeploymentLabelEntry{Key: resourceLabelKey, Value: kind},
			&dm.DeploymentLabelEntry{Key: stackLabelKey, Value: g.DeploymentName},
			&dm.DeploymentLabelEntry{Key: "system", Value: systemName},
		},
	}).Do()

	return rs, err
}

func (g *GCPCloud) ResourceLink(app, name string) (*pb.Resource, error) {
	rs, err := g.ResourceGet(name)
	if err != nil {
		return nil, err
	}

	switch rs.Kind {
	case "mysql", "postgres":
		// setup cloud-sql proxy
		ns := g.DeploymentName
		kube, err := getKubeClientset(ns)
		if err != nil {
			return nil, err
		}

		rsvars := rs.Exports

		if err = setupCloudProxy(kube, ns, g.Project, app, rsvars); err != nil {
			return nil, err
		}

		// todo refactor env setting
		env, err := g.EnvironmentGet(app)
		if err != nil {
			return nil, err
		}

		env["DATABASE_URL"] = rsvars["DATABASE_URL"]
		env["INSTANCE_NAME"] = rsvars["INSTANCE_NAME"]

		data := ""
		for key, value := range env {
			data += fmt.Sprintf("%s=%s\n", key, value)
		}

		if err = g.EnvironmentSet(app, strings.NewReader(data)); err != nil {
			return nil, err
		}

		if err = g.AppRestart(app); err != nil {
			log.Debugf("error: %+v", err)
		}
	default:
		return nil, fmt.Errorf("link is not necessary for %s", rs.Name)
	}

	dbApp, err := g.AppGet(app)
	if err != nil {
		return nil, err
	}

	appendApp(dbApp, rs)

	if err := g.saveApp(dbApp); err != nil {
		log.Warnf("saving app %v", err)
	}

	return rs, nil
}

func (g *GCPCloud) ResourceUnlink(app, name string) (*pb.Resource, error) {
	rs, err := g.ResourceGet(name)
	if err != nil {
		return nil, err
	}

	switch rs.Kind {
	case "mysql", "postgres":
		// setup cloud-sql proxy
		ns := g.DeploymentName
		kube, err := getKubeClientset(ns)
		if err != nil {
			return nil, err
		}

		if err = tearCloudProxy(kube, ns, app, rs.Name); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("link is not necessary for %s", rs.Name)
	}

	dbApp, err := g.AppGet(app)
	if err != nil {
		return nil, err
	}

	removeApp(dbApp, rs)

	if err := g.saveApp(dbApp); err != nil {
		log.Warnf("saving app %v", err)
	}

	return rs, nil
}

func (g *GCPCloud) resourceFromDeployment(dp *dm.Deployment, manifest *dm.Manifest, stack string) (*pb.Resource, error) {
	tags := deploymentLabels(dp)
	outputs := make(map[string]string)
	exports := make(map[string]string)

	var mc manifestConfig
	if manifest.Config != nil {
		if err := yaml.Unmarshal([]byte(manifest.Config.Content), &mc); err != nil {
			return nil, err
		}

		for _, out := range mc.Outputs {
			outputs[out.Name] = out.Value
		}
	}

	rs := &pb.Resource{
		Name:    tags["name"],
		Stack:   stack,
		Kind:    tags[resourceLabelKey],
		Tags:    tags,
		Outputs: outputs,
		Exports: exports,
	}

	if dp.Operation != nil {
		rs.Status = dp.Operation.Status
		rs.StatusReason = dp.Operation.StatusMessage

		if dp.Operation.Error != nil {
			rs.Status = "FAILED"
			if errText, err := dp.Operation.Error.MarshalJSON(); err == nil {
				rs.StatusReason = string(errText)
			}
		}
	}

	return rs, nil
}

func removeApp(app *pb.App, rs *pb.Resource) {
	for i, a := range rs.Apps {
		if app.Name == a {
			rs.Apps = append(rs.Apps[:i], rs.Apps[i+1:]...)
			break
		}
	}

	for i, r := range app.Resources {
		if rs.Name == r {
			app.Resources = append(app.Resources[:i], app.Resources[i+1:]...)
			break
		}
	}
}

func appendApp(app *pb.App, rs *pb.Resource) {
	found := false
	for _, a := range rs.Apps {
		if app.Name == a {
			found = true
		}
	}

	if !found {
		rs.Apps = append(rs.Apps, app.Name)
	}

	found = false
	for _, r := range app.Resources {
		if rs.Name == r {
			found = true
		}
	}

	if !found {
		app.Resources = append(app.Resources, rs.Name)
	}

}

func deploymentLabels(dp *dm.Deployment) map[string]string {
	tags := map[string]string{}

	for _, label := range dp.Labels {
		tags[label.Key] = label.Value
	}

	return tags
}
