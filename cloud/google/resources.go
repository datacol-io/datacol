package google

import (
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/dinesh/datacol/client/models"
	"gopkg.in/yaml.v2"
)

type manifestConfig struct {
	Resources []struct {
		Name       string                 `yaml:"name"`
		Type       string                 `yaml:"type"`
		Properties map[string]interface{} `yaml:"properties"`
	} `yaml:"resources"`
}

func (g *GCPCloud) ResourceDelete(name string) error {
	service := g.deploymentmanager()
	dp, manifest, err := getManifest(service, g.Project, g.DeploymentName)
	if err != nil {
		return err
	}

	mc := manifestConfig{}
	if err := yaml.Unmarshal([]byte(manifest.ExpandedConfig), &mc); err != nil {
		return err
	}

	found := false
	for i, r := range mc.Resources {
		if r.Name == name {
			mc.Resources = append(mc.Resources[:i], mc.Resources[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("%s not found", name)
	}

	c, err := yaml.Marshal(mc)
	content := string(c)
	if err != nil {
		return err
	}

	log.Debugf("content: %+v", content)
	return g.updateDeployment(service, dp, manifest, content)
}

func (g *GCPCloud) ResourceList() (models.Resources, error) {
	resp := models.Resources{}
	service := g.deploymentmanager()
	_, manifest, err := getManifest(service, g.Project, g.DeploymentName)
	if err != nil {
		return resp, err
	}

	mc := manifestConfig{}
	if err := yaml.Unmarshal([]byte(manifest.ExpandedConfig), &mc); err != nil {
		return resp, err
	}

	for _, r := range mc.Resources {
		resp = append(resp, models.Resource{
			Name: r.Name,
			Kind: dpToResourceType(r.Type, r.Name),
		})
	}

	return resp, nil
}

func (g *GCPCloud) ResourceCreate(name, kind string, params map[string]string) (*models.Resource, error) {
	service := g.deploymentmanager()
	dp, manifest, err := getManifest(service, g.Project, g.DeploymentName)
	if err != nil {
		return nil, err
	}

	rs := &models.Resource{
		Name: name,
		Kind: kind,
	}

	var sqlj2 string
	switch kind {
	case "mysql":
		params["region"] = getGcpRegion(g.Zone)
		params["zone"] = g.Zone
		params["database"] = "app"
		sqlj2 = compileTmpl(mysqlInstanceYAML, params)
	case "postgres":
		params["region"] = getGcpRegion(g.Zone)
		params["zone"] = g.Zone
		params["database"] = "app"
		sqlj2 = compileTmpl(pgsqlInstanceYAML, params)
	default:
		log.Fatal(fmt.Errorf("%s is not supported yet.", kind))
	}

	content := manifest.ExpandedConfig + sqlj2
	log.Debugf("\nDM config: %+v", content)

	if err = g.updateDeployment(service, dp, manifest, content); err != nil {
		return nil, err
	}

	exports := make(map[string]string)

	switch kind {
	case "mysql", "postgres":
		passwd, err := generatePassword()
		if err != nil {
			return nil, err
		}
		if err := g.createSqlUser(kind, passwd, name); err != nil {
			return nil, err
		}

		instName := fmt.Sprintf("%s:%s:%s", g.Project, params["region"], name)
		exports["INSTANCE_NAME"] = instName
		hostName := fmt.Sprintf("127.0.0.1:%d", getDefaultPort(kind))
		exports["DATABASE_URL"] = fmt.Sprintf("%s://%s:%s@%s/%s", kind, kind, passwd, hostName, databaseName)
	}

	rs.Exports = exports
	rs.Parameters = params

	return rs, nil
}

func (g *GCPCloud) ResourceLink(app string, rs *models.Resource) (*models.Resource, error) {
	switch rs.Kind {
	case "postgres", "mysql":
		// setup cloud-sql proxy
		ns := g.DeploymentName
		kube, err := getKubeClientset(ns)
		if err != nil {
			return nil, err
		}

		if err = setupCloudProxy(kube, ns, app, rs.Exports, g.ServiceKey); err != nil {
			return nil, err
		}

		// todo refactor env setting
		env, err := g.EnvironmentGet(app)
		if err != nil {
			return nil, err
		}

		env["DATABASE_URL"] = rs.Exports["DATABASE_URL"]
		env["INSTANCE_NAME"] = rs.Exports["INSTANCE_NAME"]

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

	return rs, nil
}

func (g *GCPCloud) ResourceUnlink(app string, rs *models.Resource) (*models.Resource, error) {
	switch rs.Kind {
	case "postgres", "mysql":
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

	return rs, nil
}

func getDefaultPort(kind string) int {
	var port int
	switch kind {
	case "mysql":
		port = 3306
	case "postgres":
		port = 5432
	default:
		log.Fatal(fmt.Errorf("No default port defined for %s", kind))
	}

	return port
}
