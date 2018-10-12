package common

import (
	"fmt"
	"strconv"

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

		if proc.CronExpr != "" {
			req.CronExpr = &proc.CronExpr
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

func ScaleApp(c *kubernetes.Clientset,
	namespace, app, image string,
	envVars map[string]string,
	enableSQLproxy bool,
	procFileData []byte,
	structure map[string]int32,
	provider cloud.CloudProvider,
) (err error) {

	var command []string

	log.Debugf("scaling request: %v", structure)

	var procfile Procfile
	if len(procFileData) > 0 {
		procfile, err = ParseProcfile(procFileData)
		if err != nil {
			return err
		}
	}

	scalePodFunc := func(proctype, image string, command []string, replicas int32, cronexpr *string) error {
		return kube.ScalePodReplicas(c, namespace, app, proctype, image, command, replicas, enableSQLproxy, cronexpr, envVars, provider)
	}

	for key, replicas := range structure {
		if procfile != nil {
			cmd, cerr := procfile.Command(key)
			if cerr != nil {
				return cerr
			}

			if procfile.HasProcessType(key) {
				switch proc := procfile.(type) {
				case ExtProcfile:
					err = scalePodFunc(key, image, cmd, replicas, proc[key].Cron)
				default:
					err = scalePodFunc(key, image, cmd, replicas, nil)
				}
			} else {
				err = fmt.Errorf("Unknown process type: %s", key)
			}
		} else if key == "cmd" {
			err = scalePodFunc(key, image, command, replicas, nil)
		} else {
			err = fmt.Errorf("Unknown process type: %s", key)
		}

		if err != nil {
			return err
		}
	}

	return
}
