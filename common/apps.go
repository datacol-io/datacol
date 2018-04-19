package common

import (
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	pb "github.com/datacol-io/datacol/api/models"
	"github.com/datacol-io/datacol/cloud"
	kube "github.com/datacol-io/datacol/k8s"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

func UpdateApp(c *kubernetes.Clientset, build *pb.Build,
	ns, image string, sqlProxy bool,
	domains []string, envVars map[string]string,
	provider cloud.CloudProvider,
	version string) error {

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

	var procesess []*pb.Process

	defaultProctype := GetDefaultProctype(build)
	procesess = append(procesess, &pb.Process{
		Proctype: defaultProctype,
		Count:    1,
	})

	runningProcesses, err := kube.ProcessList(c, ns, build.App)
	if err != nil {
		return err
	}

	for _, rp := range runningProcesses {
		if rp.Proctype == defaultProctype {
			procesess[0].Count = rp.Count // set the current worker similar to whatever running currently
		}

		// Only append non-default proceses
		if rp.Proctype != WebProcessKind && rp.Proctype != CmdProcessKind {
			procesess = append(procesess, rp)
		}
	}

	log.Debugf("defaultProctype:%s updating processes: %+v", defaultProctype, procesess)

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
			Version:             version,
		}

		if proctype == WebProcessKind || proctype == CmdProcessKind {
			req.ContainerPort = intstr.FromInt(port)
		}

		if _, err = deployer.Run(req); err != nil {
			return err
		}
	}

	// TODO: cleanup old resource based on req.Version
	return nil
}

func MergeAppDomains(domains []string, item string) []string {
	if item == "" {
		return domains
	}

	itemIndex := -1
	dotted := strings.HasPrefix(item, ":")

	if dotted {
		item = item[1:]
	}

	for i, d := range domains {
		if d == item {
			itemIndex = i
			break
		}
	}

	if dotted && itemIndex >= 0 {
		return append(domains[0:itemIndex], domains[itemIndex+1:]...)
	}

	if !dotted && itemIndex == -1 {
		return append(domains, item)
	}

	return domains
}
