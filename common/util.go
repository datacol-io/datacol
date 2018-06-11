package common

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/appscode/go/crypto/rand"
	pb "github.com/datacol-io/datacol/api/models"
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

func GetJobID(name, process_type string) string {
	if process_type == "" {
		process_type = CmdProcessKind
	}

	return fmt.Sprintf("%s-%s", name, process_type)
}

func GetDefaultProctype(b *pb.Build) string {
	proctype := CmdProcessKind

	if len(b.Procfile) > 0 {
		log.Debugf("build %s have non-empty procfile", b.Id)
		proctype = WebProcessKind
	}

	return proctype
}

func GetProcessCommand(proctype string, b *pb.Build) (command []string, err error) {
	if proctype == "" {
		err = errors.New("Empty process type")
	}

	if len(b.Procfile) > 0 {
		procfile, err := ParseProcfile(b.Procfile)
		if err != nil {
			return nil, err
		}

		command, err = procfile.Command(proctype)
	}

	return
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
