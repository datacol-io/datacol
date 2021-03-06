package aws

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	pb "github.com/datacol-io/datacol/api/models"
	"github.com/datacol-io/datacol/common"
)

func (a *AwsCloud) s3KeyForEnv(name string) string {
	return fmt.Sprintf("%s/.env", name)
}

func (a *AwsCloud) EnvironmentGet(name string) (pb.Environment, error) {
	s3key := a.s3KeyForEnv(name)
	data, err := a.s3Get(a.SettingBucket, s3key)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == s3.ErrCodeNoSuchKey {
			return pb.Environment{}, nil
		}
		return nil, err
	}

	return common.LoadEnvironment(data), nil
}

func (a *AwsCloud) EnvironmentSet(name string, r io.Reader) error {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	s3key := a.s3KeyForEnv(name)
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
