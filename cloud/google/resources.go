package google

import (
  "fmt"

  "gopkg.in/yaml.v2"
  log "github.com/Sirupsen/logrus"
  "github.com/dinesh/datacol/client/models"
)

type manifestConfig struct {
  Resources []struct{
    Name string `yaml: "name"`
    Type string `yaml: "type"`
    Properties map[string]interface{} `yaml: "properties"`
  } `yaml: "resources"`
}

func (g *GCPCloud) ResourceDelete(name string) (error) {
  service := g.deploymentmanager()
  dp, manifest, err := getManifest(service, g.Project, g.DeploymentName)
  if err != nil { return err }

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
  if err != nil { return err }

  log.Debugf("content: %+v", content)
  return g.updateDeployment(service, dp, manifest, content)
}

func (g *GCPCloud) ResourceList() (models.Resources, error) {
  resp := models.Resources{}
  service := g.deploymentmanager()
  _, manifest, err := getManifest(service, g.Project, g.DeploymentName)
  if err != nil { return resp, err }

  mc := manifestConfig{}
  if err := yaml.Unmarshal([]byte(manifest.ExpandedConfig), &mc); err != nil {
    return resp, err
  }

  for _, r := range mc.Resources {
    resp = append(resp, models.Resource{
      Name:  r.Name,
      Kind:  dpToResourceType(r.Type, r.Name),
    })
  }
  
  return resp, nil
}

func (g *GCPCloud) ResourceCreate(name, kind string, params map[string]string) (*models.Resource, error) {
  service := g.deploymentmanager()
  dp, manifest, err := getManifest(service, g.Project, g.DeploymentName)
  if err != nil { return nil, err }

  rs := &models.Resource{
    Name:       name,
    Kind:       kind,
  }

  switch kind {
  case "mysql", "postgres":
    params["region"] = getGcpRegion(g.Zone)
    params["zone"]   = g.Zone
  }

  sqlj2 := compileTmpl(mysqlInstanceYAML, params)
  content := manifest.ExpandedConfig + sqlj2

  log.Debugf("\nDM config: %+v", content)

  if err = g.updateDeployment(service, dp, manifest, content); err != nil {
    return nil, err
  }

  exports := make(map[string]string)

  switch kind {
  case "mysql", "postgres":
    user := g.DeploymentName
    passwd, err := generatePassword()
    if err != nil { return nil, err }
    if err := g.createSqlUser(user, passwd, name); err != nil {
      return nil, err
    }

    instName := fmt.Sprintf("%s:%s:%s", g.Project, params["region"], name)
    exports["INSTANCE_NAME"] = instName
    exports["INSTANCE_AUTH"] = fmt.Sprintf("%s:%s", user, passwd)
  }

  rs.Exports    = exports
  rs.Parameters = params

  return rs, nil
}
