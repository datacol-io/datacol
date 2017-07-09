package aws

import (
	"github.com/appscode/go/crypto/rand"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/docker/docker/pkg/archive"
	"github.com/jhoonb/archivex"
	"io/ioutil"
	"os"
	"time"
)

func generateId(prefix string, size int) string {
	return prefix + "-" + rand.Characters(size)
}

func coalesce(s *dynamodb.AttributeValue, def string) string {
	if s != nil {
		return *s.S
	} else {
		return def
	}
}

func timestampNow() int32 {
	return int32(time.Now().Unix())
}

func convertGzipToZip(src string) (string, error) {
	dir, err := ioutil.TempDir(os.Getenv("HOME"), "zip-")
	if err != nil {
		return dir, err
	}

	if err = untarPath(src, dir); err != nil {
		return dir, err
	}

	zipf := new(archivex.ZipFile)
	zipf.Create(dir)
	if err = zipf.AddAll(dir, false); err != nil {
		return dir, err
	}

	return zipf.Name, zipf.Close()
}

func untarPath(src, dst string) error {
	fd, err := os.Open(src)
	if err != nil {
		return err
	}
	defer fd.Close()

	defaultArchiver := archive.Archiver{Untar: archive.Untar, UIDMaps: nil, GIDMaps: nil}
	return defaultArchiver.Untar(fd, dst, &archive.TarOptions{NoLchown: true})
}
