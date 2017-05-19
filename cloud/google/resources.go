package google

import (
	"encoding/json"
	"fmt"
	"strings"

	"cloud.google.com/go/datastore"
	log "github.com/Sirupsen/logrus"
	pb "github.com/dinesh/datacol/api/models"
	"gopkg.in/yaml.v2"
)

const (
	databaseName = "app"
	resourceKind = "Resource"
)

type dmResource struct {
	Name       string                 `yaml:"name"`
	Type       string                 `yaml:"type"`
	Properties map[string]interface{} `yaml:"properties"`
}

type manifestConfig struct {
	Resources []dmResource `yaml:"resources"`
}

func (g *GCPCloud) ResourceGet(name string) (*pb.Resource, error) {
	rs := new(pb.Resource)
	ctx, key := g.nestedKey(resourceKind, name)
	err := g.datastore().Get(ctx, key, rs)
	return rs, err
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
	rs_db := fmt.Sprintf("%s-%s", name, databaseName)

	resources := []dmResource{}

	for _, r := range mc.Resources {
		if r.Name != name && r.Name != rs_db {
			resources = append(resources, r)
		} else {
			found = true
		}
	}

	mc.Resources = resources

	if !found {
		return fmt.Errorf("%s not found", name)
	}

	c, err := yaml.Marshal(mc)
	content := string(c)
	if err != nil {
		return err
	}

	log.Debugf("content: %+v", content)
	if err := g.updateDeployment(service, dp, manifest, content); err != nil {
		return err
	}

	ctx, key := g.nestedKey(resourceKind, name)
	return g.datastore().Delete(ctx, key)
}

func (g *GCPCloud) ResourceList() (pb.Resources, error) {
	var rs pb.Resources

	q := datastore.NewQuery(resourceKind)
	if _, err := g.datastore().GetAll(g.ctxNS(), q, &rs); err != nil {
		return nil, err
	}

	return rs, nil
}

func (g *GCPCloud) resourceListFromStack() (pb.Resources, error) {
	resp := pb.Resources{}
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
		resp = append(resp, &pb.Resource{
			Name: r.Name,
			Kind: dpToResourceType(r.Type, r.Name),
		})
	}

	return resp, nil
}

func (g *GCPCloud) ResourceCreate(name, kind string, params map[string]string) (*pb.Resource, error) {
	service := g.deploymentmanager()
	dp, manifest, err := getManifest(service, g.Project, g.DeploymentName)
	if err != nil {
		return nil, err
	}

	rs := &pb.Resource{Name: name, Kind: kind}

	var sqlj2 string
	switch kind {
	case "mysql":
		params["region"] = getGcpRegion(g.Zone)
		params["zone"] = g.Zone
		params["database"] = databaseName
		sqlj2 = compileTmpl(mysqlInstanceYAML, params)
	case "postgres":
		params["region"] = getGcpRegion(g.Zone)
		params["zone"] = g.Zone
		params["database"] = databaseName
		
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
		passwd := generatePassword()
		if err := g.createSqlUser(kind, passwd, name); err != nil {
			return nil, err
		}

		instName := fmt.Sprintf("%s:%s:%s", g.Project, params["region"], name)
		exports["INSTANCE_NAME"] = instName
		hostName := fmt.Sprintf("127.0.0.1:%d", getDefaultPort(kind))
		exports["DATABASE_URL"] = fmt.Sprintf("%s://%s:%s@%s/%s", kind, kind, passwd, hostName, databaseName)
	}

	rs.Exports = jsonEncode(exports)
	rs.Parameters = jsonEncode(params)

	log.Debugf("storing %s details into datastore.", toJson(rs))

	store := g.datastore()
	ctx, key := g.nestedKey(resourceKind, name)
	if _, err := store.Put(ctx, key, rs); err != nil {
		return nil, err
	}

	defer store.Close()
	return rs, nil
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

		rsvars := jsonDecode(rs.Exports)
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

	append_app(app, rs)

	ctx, key := g.nestedKey(resourceKind, rs.Name)
	if _, err := g.datastore().Put(ctx, key, rs); err != nil {
		return nil, err
	}

	return rs, nil
}

func append_app(app string, rs *pb.Resource) {
	found := false
	for _, a := range rs.Apps {
		if app == a {
			found = true
		}
	}

	if !found {
		rs.Apps = append(rs.Apps, app)
	}
}

func remove_app(app string, rs *pb.Resource) {
	for i, a := range rs.Apps {
		if app == a {
			rs.Apps = append(rs.Apps[:i], rs.Apps[i+1:]...)
			break
		}
	}
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

	remove_app(app, rs)

	ctx, key := g.nestedKey(buildKind, rs.Name)
	if _, err := g.datastore().Put(ctx, key, rs); err != nil {
		return nil, err
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

func jsonEncode(opts map[string]string) []byte {
	b, err := json.Marshal(opts)
	if err != nil {
		log.Fatal(fmt.Errorf("marshaling %+v err:%v", opts, err))
	}
	return b
}

func jsonDecode(b []byte) map[string]string {
	var opts map[string]string
	if err := json.Unmarshal(b, &opts); err != nil {
		log.Fatal(fmt.Errorf("unmarshaling %+v err:%v", opts, err))
	}
	return opts
}
