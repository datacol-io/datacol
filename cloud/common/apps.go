package common

import (
	"strconv"

	pb "github.com/datacol-io/datacol/api/models"
	"github.com/datacol-io/datacol/cloud"
	"github.com/datacol-io/datacol/cloud/kube"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

func UpdateApp(c *kubernetes.Clientset, build *pb.Build,
	ns, image string, sqlProxy bool,
	domains []string, envVars map[string]string, provider cloud.CloudProvider) error {

	deployer, err := kube.NewDeployer(c)
	if err != nil {
		return err
	}

	port := cloud.DefaultProcPort
	if pv, ok := envVars["PORT"]; ok {
		p, err := strconv.Atoi(pv)
		if err != nil {
			return err
		}
		port = p
	}

	procesess, err := kube.ProcessList(c, ns, build.App)
	if err != nil {
		return err
	}

	if len(procesess) == 0 {
		procesess = append(procesess, &pb.Process{
			Proctype: GetDefaultProctype(build),
			Workers:  1,
		})
	}

	for _, proc := range procesess {
		proctype := proc.Proctype
		command, err := GetProcessCommand(proctype, build)
		if err != nil {
			return err
		}

		req := &kube.DeployRequest{
			App:                 build.App,
			Args:                command,
			Image:               image,
			Domains:             domains,
			EnvVars:             envVars,
			Namespace:           ns,
			Proctype:            proctype,
			Provider:            provider,
			ServiceID:           GetJobID(build.App, proctype),
			EnableCloudSqlProxy: sqlProxy,
		}

		if proctype == WebProcessKind || proctype == CmdProcessKind {
			req.ContainerPort = intstr.FromInt(port)
		}

		if _, err = deployer.Run(req); err != nil {
			return err
		}
	}

	return nil
}
