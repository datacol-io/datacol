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

	scalePodFunc := func(proctype, image string, command []string, replicas int32) error {
		return sched.ScalePodReplicas(c, namespace, app, proctype, image, command, replicas)
	}

	for key, replicas := range structure {
		if procfile != nil {
			if procfile.HasProcessType(key) {
				if cmd, err := procfile.Command(key); err == nil {
					err = scalePodFunc(key, image, cmd, replicas)
				}
			} else {
				err = fmt.Errorf("Unknown process type: %s", key)
			}
		} else if key == "cmd" {
			err = scalePodFunc(key, image, command, replicas)
		} else {
			err = fmt.Errorf("Unknown process type: %s", key)
		}

		if err != nil {
			return err
		}
	}

	return
}

func GetJobID(name, process_type string) string {
	if process_type == "" {
		process_type = "cmd"
	}

	return fmt.Sprintf("%s-%s", name, process_type)
}

func GetContainerCommand(b *pb.Build) (command []string, proctype string, err error) {
	proctype = "cmd"
	if len(b.Procfile) > 0 {
		proctype = "web"
		procfile, err := ParseProcfile(b.Procfile)
		if err != nil {
			return nil, proctype, err
		}

		command, err = procfile.Command(proctype)
	}

	return
}
