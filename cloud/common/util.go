package common

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/appscode/go/crypto/rand"
	pb "github.com/dinesh/datacol/api/models"
	sched "github.com/dinesh/datacol/cloud/kube"
	"k8s.io/client-go/kubernetes"
)

func LoadEnvironment(data []byte) pb.Environment {
	e := pb.Environment{}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), "=", 2)

		if len(parts) == 2 {
			if key := strings.TrimSpace(parts[0]); key != "" {
				e[key] = parts[1]
			}
		}
	}

	return e
}

func GenerateId(prefix string, size int) string {
	return prefix + "-" + rand.Characters(size)
}

func ScaleApp(c *kubernetes.Clientset, namespace, app, image string,
	procFileData []byte, structure map[string]int32) (err error) {

	var command []string

	log.Debugf("scaling request: %v", structure)

	var procfile Procfile
	if len(procFileData) > 0 {
		procfile, err = ParseProcfile(procFileData)
		if err != nil {
			return err
		}
	}

	scalePodFunc := func(jobID, image string, command []string, replicas int32) error {
		return sched.ScalePodReplicas(c, namespace, jobID, image, command, replicas)
	}

	for key, replicas := range structure {
		jobID := fmt.Sprintf("%s-%s", app, key)

		if procfile != nil {
			switch version := procfile.Version(); version {
			case StandardType:
				if rawCmd, ok := procfile.(StdProcfile)[key]; ok {
					command = strings.Split(rawCmd, " ")
					err = scalePodFunc(jobID, image, command, replicas)
				} else {
					err = fmt.Errorf("Unknown process type: %s", key)
				}
			case ExtentedType:
				extendProcfile := procfile.(ExtProcfile)
				if p, ok := extendProcfile[key]; ok {
					command = strings.Split(p.Command, " ")
					err = scalePodFunc(jobID, image, command, replicas)
				} else {
					err = fmt.Errorf("Unknown process type: %s", key)
				}
			}
		} else if key == "cmd" {
			err = scalePodFunc(jobID, image, command, replicas)
		} else {
			err = fmt.Errorf("Unknown process type: %s", key)
		}

		if err != nil {
			return err
		}
	}

	return
}
