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
	procfile map[string]string, structure map[string]int32) error {

	var command []string

	log.Debugf("scaling request: %v", structure)

	for key, replicas := range structure {
		jobID := fmt.Sprintf("%s-%s", app, key)

		if rawCmd, ok := procfile[key]; ok {
			command = strings.Split(rawCmd, " ")
			if err := sched.ScalePodReplicas(c, namespace, jobID, image, command, replicas); err != nil {
				return err
			}
		} else if key == "cmd" {
			if err := sched.ScalePodReplicas(c, namespace, jobID, image, command, replicas); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("Unknown process type: %s", key)
		}
	}

	return nil
}
