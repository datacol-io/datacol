package aws

import (
	"bufio"
	"bytes"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"io/ioutil"
	"strings"

	pb "github.com/dinesh/datacol/api/models"
	"io"
)

func (a *AwsCloud) EnvironmentGet(name string) (pb.Environment, error) {
	s3key := fmt.Sprintf("%s.env", name)
	data, err := a.s3Get(a.SettingBucket, s3key)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == s3.ErrCodeNoSuchKey {
			return pb.Environment{}, nil
		}
		return nil, err
	}

	return loadEnv(data), nil
}

func (a *AwsCloud) EnvironmentSet(name string, r io.Reader) error {
	s3key := fmt.Sprintf("%s.env", name)
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	return a.s3Put(a.SettingBucket, s3key, data)
}

func (a *AwsCloud) s3Get(bucket, key string) ([]byte, error) {
	log.Debugf("Fetching data from s3://%s/%s", bucket, key)

	res, err := a.s3().GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(res.Body)
}

func (a *AwsCloud) s3Put(bucket, key string, data []byte) error {
	req := &s3.PutObjectInput{
		Body:          bytes.NewReader(data),
		Bucket:        aws.String(bucket),
		ContentLength: aws.Int64(int64(len(data))),
		Key:           aws.String(key),
	}

	_, err := a.s3().PutObject(req)

	return err
}

func loadEnv(data []byte) pb.Environment {
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
