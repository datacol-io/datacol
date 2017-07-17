// Code generated by go-bindata.
// sources:
// cloud/aws/templates/app.tmpl
// cloud/aws/templates/mysql.tmpl
// cloud/aws/templates/postgres.tmpl
// cloud/aws/templates/redis.tmpl
// DO NOT EDIT!

package aws

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _cloudAwsTemplatesAppTmpl = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xbc\x58\x5b\x6f\xdb\xb8\x12\x7e\xf7\xaf\x20\x08\x3f\x9c\x13\xc4\x4e\x9c\xe2\xe0\x6c\xf5\xa6\xc8\x8a\x57\x0b\x5f\x04\xdb\x69\x17\xe8\x06\x2e\x23\x8d\x6d\xd6\xb2\xa8\x25\xa9\xa4\xa9\xa1\xff\xbe\x20\x25\x3b\xd6\xd5\x49\x9a\xad\x0b\x14\x11\x39\x33\x9c\xf9\xe6\xc2\x4f\xda\xed\x90\x0f\x4b\x1a\x02\xc2\x1c\x04\x8b\xb9\x07\x18\x25\x49\x0b\xa1\x5d\x0b\x21\x84\xb0\xf9\x79\x36\x87\x6d\x14\x10\x09\x37\x8c\x6f\x89\xfc\x04\x5c\x50\x16\x62\x03\xe1\xab\xcb\xde\x65\xe7\xf2\x63\xe7\xf2\x23\x3e\x4f\xa5\x5d\xc2\xc9\x16\x24\x70\x81\x8d\xcc\x82\xb2\x11\x45\x63\xb2\x85\xa3\x25\x84\xf0\xfc\x29\x52\x2b\x78\x26\x39\x0d\x57\x99\x01\xbd\xd3\x07\xe1\x71\x1a\xc9\xec\x14\xa5\x8a\xd8\x12\x91\x28\x0a\xa8\x47\xf4\x72\x26\x9c\xec\xb5\xf0\x75\xec\x6d\x40\xba\x1c\x96\xf4\xfb\x1b\xcf\x89\x88\x5c\xa3\x48\x5b\x50\xc7\xdd\x6b\x8b\x48\x32\x14\x30\xb6\x41\x4b\xc6\x51\x8a\x4f\xc7\x63\x3e\x94\x3d\x98\x49\xe2\x6d\xde\x21\xcc\x7b\x22\x00\x09\x65\x0c\x85\xca\x5a\xe9\xa0\x29\x44\x4c\x50\xc9\xf8\xd3\x2d\x0f\x7e\xf2\x30\x15\x8a\xc7\xb6\x5b\x2a\x11\x3f\x98\x3d\x1c\xd9\x3a\x3a\x18\x5b\x2c\xf4\xa9\xd2\xcf\xa5\xf6\x77\xf2\x00\x16\xf3\xc1\xd2\x46\xf2\xde\xdc\x84\x86\x31\x66\x6a\xf1\xcb\xf3\x6a\xb6\x6e\xff\x1d\x93\x40\xe8\x2d\x3c\x85\xa5\xf2\x29\x1f\x57\x72\x8e\xf0\x92\x04\x02\xf0\xdd\x41\x37\xb9\xab\xf4\x6c\x9a\x15\x6e\xce\x31\xdb\x9a\x3e\x1b\xac\x46\xc9\xfc\x3c\x33\x0c\xdb\x9a\x1a\xc6\x91\xe4\x11\x66\x2e\x67\x11\x70\x49\x73\x86\x0b\x29\x28\x25\xfc\x10\xe1\x1f\x8c\x2a\xa8\xbf\xe0\x0e\x3e\xcf\x03\x90\x99\x58\xa6\x99\xda\x57\x4d\x4e\x20\x39\x47\x75\x1a\xfb\x66\x2a\xca\x63\x95\x40\x7c\x77\xd7\x3a\x5e\xad\x76\xda\x65\x01\xf5\x9e\xe6\xf0\x5d\x96\x5d\xcf\xb5\xf7\xe5\x6f\x9d\xde\x65\xa7\xf7\x7f\x7c\x9e\x17\x9a\x49\x22\x61\x0b\x61\x59\x5f\x6d\x52\x5f\xfb\x19\x04\xec\xd1\x8d\x83\xc0\x8d\xc5\xba\x60\x40\xe5\x67\xb9\x04\x4f\x1e\x04\xcb\x02\x2e\xa7\xa1\x47\x23\xa2\x4a\x1c\x9f\x95\xf7\x4d\x2f\x2b\xe6\x2f\x18\x3c\x6e\x9c\xe1\xbb\x3c\x22\xad\xaa\xbf\x93\x52\x37\xa9\xe2\xbd\x8e\x69\xe0\x4f\x59\x50\xd3\xba\xba\x4e\x1c\x73\x64\x18\x5a\xe6\x45\x15\x62\x0a\x11\x6f\x41\xc9\xa7\x60\xf7\x99\x17\x57\x02\x96\x07\xbc\x77\x75\x1a\xf0\x72\x2d\xbd\x0a\xcb\xa2\xb6\xb2\x0e\xfc\x81\x7a\xa0\xb1\x54\x13\xe1\x5e\xe1\xd1\x25\x5b\xf2\x83\x85\xe4\x51\x74\x3d\xb6\x2d\xa0\x5b\xa8\xad\x62\x42\x84\x14\xc6\x33\x02\xc5\xcc\xd4\x97\xa8\xc6\x2a\x45\xb3\x10\x65\xba\x55\xdd\x6e\x2f\x6a\xb8\x93\x2d\x57\xd9\x74\x27\xda\x2e\x6d\xbc\x48\xbb\x96\x6f\xbd\x12\x40\xf8\x44\x1d\xbc\xb0\x12\x4e\xd5\x02\x7a\x49\x3d\xa0\x7c\xba\x2a\x76\x11\xc2\x01\x5b\x09\xc3\xe2\x40\x24\x0c\xd9\x6a\xc0\x59\x1c\x55\x1a\x2a\x89\xce\x24\x07\xb2\x6d\x94\x75\x63\x39\x64\x2b\xfb\x01\x42\x29\x8a\x88\xaa\xdf\x5d\xa5\xc7\xfb\x39\xaf\x02\x23\x3c\x34\xc8\xa3\x30\xb4\xb9\x33\xf5\xaf\x6c\xa7\x54\xa2\xa8\x22\xbf\xef\x05\x97\x9a\x41\x03\x90\x66\x2c\xd7\x8c\xd3\x1f\x9a\xa6\xcc\xd9\x06\xc2\x3a\x20\x9e\xaf\x5e\x63\x40\xa5\x1a\x94\x6f\x43\xe2\xd7\x05\x8e\xc5\x07\x15\xe2\xe4\xfe\x9b\xd2\x3a\xe9\x59\xd5\x91\x59\xab\xce\xe2\xfb\xe3\x24\x8a\x0f\x86\x61\xb4\x77\xc7\x1c\x2e\xb9\xa8\x88\x2b\x3f\xd5\xff\x85\x58\xdf\xe4\xbd\xca\x7c\x7b\xa7\x2f\x89\x29\xac\x28\x0b\x93\xfd\xa3\xe9\x79\x2c\x0e\xa5\xe3\x27\xc6\x33\xbb\xba\x68\xef\x72\xdc\xa4\x2e\xd2\x9f\xa9\xc2\x3e\x7b\x0c\x03\x46\xfc\x5b\x1e\xdc\x30\x3e\x24\x4f\xc0\xeb\xca\x50\x69\x5c\x13\xe9\xad\x07\x20\x9d\x2d\x59\xc1\x49\x41\x6b\x0d\xde\x46\xdb\x34\x1f\x08\x0d\xc8\x3d\x0d\xa8\x7c\x6a\x52\x73\xe3\xd3\xa6\x9d\x90\x4a\xaa\x06\x88\x32\x7c\x1b\x29\xef\x9b\xc4\x53\x09\x2d\xec\x12\x2e\x9b\x44\x2d\xb6\x8d\x02\xc8\x5b\xae\xea\xb4\x72\x69\xb5\x9a\x24\x72\x04\xe3\xae\x89\x61\x04\x2c\xf6\x35\xc5\x70\x39\xfb\x96\x16\x61\x1d\xcb\x38\xb0\x11\xc3\xd8\x0b\xbf\x88\x6e\xd4\xd3\xd0\xac\x58\xdb\xbb\xc3\xb5\x97\x74\xda\xbb\xec\x36\x4b\xf4\xbb\x4c\x47\xdf\xf7\xc0\x71\xed\xbd\x3c\xa7\x5b\x60\xb1\x74\xc2\x11\x0d\x63\xa9\x8f\xef\xfd\x2f\x27\x91\x51\x88\x12\x89\x3a\xf8\xa1\x66\xa3\xd4\xfd\x97\x63\x5c\x5d\x93\x87\xf5\xe7\xda\xe1\x03\xe5\x2c\xac\x26\x4e\x2a\xaf\xb1\x84\x3d\x7e\xd7\xb7\xce\xb0\xbf\x18\xd8\x63\x7b\x6a\x0e\x7b\x8b\x91\xdd\x77\x6e\x47\x45\x1e\x95\x96\xa1\x6a\xde\x47\x71\x71\xa0\x3a\x17\x3e\xf3\x36\xc0\x8d\x5e\xb7\x77\xd5\xed\x15\x75\xf6\x07\x0c\x9d\xf1\xed\x9f\x0b\x6b\x32\x9e\x9b\xce\xd8\x9e\x16\xc5\x8e\x7c\xfd\x44\x38\x25\xf7\x41\x15\x8f\x39\xce\x16\x76\x46\xe6\xc0\x5e\x4c\x6d\x77\xb2\x18\x9b\x23\xbb\xa2\x88\xf1\x27\x12\xc4\x75\xe3\x68\x4f\x4d\xf2\xaf\x3a\x27\xcb\xb8\x34\x59\x1a\x3c\x9c\xda\x03\x67\x32\x7e\xab\x63\x47\x33\xf1\x7d\xdd\x32\x2d\x6b\x72\x3b\x9e\x2f\x9c\xfe\x4f\xb9\x76\x98\xcf\xef\xeb\x5d\x9a\xd6\xb9\x39\x68\x72\x0e\x07\x44\x82\x90\x65\x3e\xf9\x0a\x14\x5c\xf7\xcd\xe1\x57\xb3\xd9\x8a\xb8\x73\xcf\xf5\xcc\xdd\xe4\x92\x2e\x89\x27\x8b\x93\xe9\xa8\x7f\xc6\x93\x85\x39\x9d\x3b\x37\xa6\x35\x9f\xd5\xb7\xfc\xac\xfa\x02\xde\x5b\xa9\x64\xfe\xce\x52\xd3\x93\xc2\xb7\x88\x73\x84\xad\x49\xdf\xb6\x26\xa3\x91\x33\x57\x4f\xb3\x0f\xc5\x37\x91\x42\x0b\x0f\x59\xf6\x79\xa9\xf1\x9c\x32\xe2\xa5\x83\x9b\xa0\xcf\x7f\xe6\xa8\xe0\x33\xd5\xda\xb9\x59\x9e\x67\x4a\x29\x67\xe9\xfe\xa0\xd1\xe9\x7c\x36\x03\xa0\xc7\xf2\x2c\x02\xef\xe4\x3b\xd6\x5f\xa1\x7a\xc9\x2a\x43\xf1\x90\xbe\xc3\x18\xe8\xb2\x7b\x55\x55\x9c\xd1\x9a\x08\x10\x46\xd5\x16\x8a\x38\x2c\xf4\x3c\xae\xde\x46\x48\xd1\x65\x12\xfa\x35\xea\xea\xd7\x41\xe0\xad\x19\x1a\xb2\xd5\x8a\x86\x2b\x44\x43\xc9\x90\xa9\xdf\x63\x91\x6d\x4d\xbb\xdd\x6e\x83\x66\xfb\x3f\xe4\x51\x20\xf0\x38\x5a\x81\xec\x04\x6c\x45\x43\xd4\xe9\x70\x3d\xbe\x50\x3b\x9d\x84\xff\xad\xd4\x7f\x1f\xa7\x35\xf8\x48\x48\xc2\x25\xf8\x88\x85\xe8\xab\x4f\x24\x7c\x7d\x91\x9a\x0a\x56\xae\x01\xf5\xf5\x45\x86\x68\x0d\xdb\xda\x2b\xa6\xf7\x5d\xea\x37\xea\x48\xd4\x2e\xdc\x44\x46\xfb\x30\xc3\x50\x13\x64\x99\x1d\x49\x56\x8d\x26\xda\xcf\xf3\xba\xeb\x6f\x78\x17\x3c\xde\xcd\x00\xcd\x7f\x65\xb8\x68\xb0\x72\x0a\x87\x33\xf5\xcb\x40\xf4\x32\xca\xe7\xa7\xab\xd5\xd5\xc6\x84\x7c\xa7\x72\x73\x63\xb1\x56\x19\x78\x1d\xfa\x51\x2c\xd6\xbf\x12\x1a\xcd\x7d\xb4\xaf\x07\x5c\x8a\xc3\xe1\x55\x9f\xd0\x5a\xfb\xff\x93\xd6\x6e\x87\x20\xf4\x51\x92\xfc\x13\x00\x00\xff\xff\xf1\x21\x2c\xac\x4f\x18\x00\x00")

func cloudAwsTemplatesAppTmplBytes() ([]byte, error) {
	return bindataRead(
		_cloudAwsTemplatesAppTmpl,
		"cloud/aws/templates/app.tmpl",
	)
}

func cloudAwsTemplatesAppTmpl() (*asset, error) {
	bytes, err := cloudAwsTemplatesAppTmplBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloud/aws/templates/app.tmpl", size: 6223, mode: os.FileMode(420), modTime: time.Unix(1500318804, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudAwsTemplatesMysqlTmpl = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xb4\x57\x6d\x6b\x23\x37\x10\xfe\x9e\x5f\x31\xe8\x53\x0b\x3e\xe3\xcb\xd1\x2b\xb7\x94\x82\xb3\x76\xc2\xc2\x25\x75\xb3\xa9\x0f\x7a\xe4\x83\xbc\x3b\x36\xe2\x64\x49\x27\x69\x73\xf8\x8c\xff\x7b\xd1\xbe\x6b\x5f\x9c\x17\x68\x3e\x04\xa3\x99\x67\xe6\xd1\x68\x9e\x91\xf6\x78\x84\x14\xb7\x4c\x20\x10\x8d\x46\x66\x3a\x41\x02\xa7\xd3\x05\xc0\xf1\x02\x00\x80\xcc\xbf\xc4\x0f\xb8\x57\x9c\x5a\xbc\x96\x7a\x4f\xed\x1a\xb5\x61\x52\x10\x08\x80\x5c\xce\xde\xcf\xde\xcd\x3e\xbd\x9b\x7d\x22\x93\xc2\x3d\x94\x22\x65\x96\x49\x61\x48\x50\x86\x00\x20\x2b\xcd\x9e\xa8\x45\xb7\x04\xe4\x5a\x04\xc1\xf2\x7b\x46\xb9\x73\xf9\xea\x56\xee\x71\x4b\x82\xc6\x0b\x4e\x13\x20\x56\x67\x48\xe0\x11\x4e\x79\x8c\x53\x19\x7e\x45\x35\xdd\xa3\x45\xed\x85\x9f\x73\x2e\x13\x6a\x31\x8d\xad\xd4\x74\x87\x2d\x1b\x00\x79\x38\x28\xcc\xd9\xde\x65\xfb\x0d\xea\x92\x69\x6e\x5a\xe0\x96\x66\xdc\xe6\xd6\xf7\x33\xdf\x62\x12\xcd\x94\xad\x76\x5a\xa7\x00\x53\xe4\x00\xc3\x7e\x22\xfc\x72\x73\xf5\x2b\x29\x51\xa7\x0a\x4e\x16\xd4\xd2\x0d\x35\x63\x3c\x62\xab\x99\xd8\x8d\xf1\xa0\x4a\x9d\x23\x52\xba\x42\x5a\xe6\x00\x41\xf7\xd8\xa7\x10\x09\x63\xa9\x48\x30\x4f\xfa\x16\x1a\xe9\x66\x6a\x2f\xa7\x7b\x96\x68\x79\x8e\x4e\x95\x07\x12\x4e\x8d\x81\xad\xd4\x2d\x66\x32\x45\xd3\xa7\x76\x9b\x71\xcb\xe6\xff\xbe\x89\xd5\x96\x72\x83\xe7\xf8\xe4\xc1\x15\x47\xa0\x4f\x94\x71\xba\x61\x9c\xd9\x03\xfc\x94\x62\xa0\x46\x2b\x6a\xcc\x0f\xa9\xd3\x57\x30\xf1\x93\xc5\xa8\x9f\x50\x83\xaa\xe2\xf4\x33\x34\x7d\xdf\x49\xf0\x6c\xfc\x00\x48\xa8\x91\x5a\x04\x26\x40\x15\x71\xc0\x64\x1b\x81\xd6\x0c\x55\x67\xa8\x38\xae\x69\x7f\x60\xba\xa6\x3c\xc3\x42\x6c\x85\xac\x26\x95\x2f\x3c\xf6\x28\xc7\x65\x8a\x41\xca\x9f\x99\xb1\x7f\xcc\xbf\xc4\x41\xb0\x0c\x2f\x83\xa0\xf0\x0d\x82\x28\xfd\xf3\xcc\x36\xd6\xab\xb0\x26\x3e\x96\x6e\xbc\x50\xf0\x8a\xb4\x4d\x9f\x9c\x3b\x35\xc7\xa7\x5b\xd0\x1e\xaf\x7f\x0c\xea\x5c\x59\xff\x83\x84\xcb\xb6\xc9\xaa\x14\xbd\xe4\xd5\x88\xf5\x72\x77\xab\x5a\xf8\x80\xdc\xc2\xed\x21\xfe\xfb\xf3\x48\x4f\xfc\x36\xfd\x7d\xfa\xfe\x63\xdb\xd8\x69\xbf\x7e\x72\x95\x0c\x1f\x7e\x73\x00\xeb\x55\xe8\xaa\x7f\xfe\xcc\x07\x23\x87\x2c\xd5\xe7\xb7\xb5\x0a\x21\x8c\x16\xf7\x70\xc5\x65\xf2\xed\x05\xbc\xbd\xeb\xe1\xaf\xcc\xaa\xcc\xfa\x57\x8f\xd4\xf6\xc3\x87\xd9\xc7\x87\x44\xcd\xd3\x22\x39\x90\x5c\x11\xcd\x6d\x74\x83\x76\x6e\x6d\x21\x90\x6a\xa0\x39\x91\x2c\x45\xaa\x24\x13\x76\xea\x90\x68\x4c\x7e\x1f\xb5\xe5\xdd\xc4\x76\x3f\xdf\x16\x3b\x47\x76\x02\x2f\xc5\xd3\xed\xc1\x7c\xe7\xed\x8b\xc4\x8b\x5c\xde\x98\xb5\x7d\x10\xdd\x9e\x6f\x43\xe8\xda\x3e\x88\x6e\x2b\x60\x08\x5d\xdb\x1d\xda\x3b\x85\xfb\xf2\x21\xe1\x9d\x43\x8c\x49\xa6\x99\x3d\xdc\x68\x99\xa9\xe7\x1a\xcc\x77\x6e\x35\xc1\x4a\x4b\x85\xda\x32\xf4\x07\x14\x00\xc9\x5d\x3b\xcd\xb4\x77\xfb\x80\xfa\x5d\x33\x69\xfb\x7b\x29\x22\xb1\xcb\xcf\x37\x80\xaf\x2d\x1f\x70\xbb\x8d\xd4\x4a\x4b\x2b\x13\xc9\x5d\x44\x9b\x28\x77\x78\xd7\x5a\xee\xcb\x13\x27\xae\x01\xdc\xda\x83\xec\xae\xb8\x66\x8f\x94\x57\xb4\x4a\x02\x75\xcd\x8a\xbf\x47\x8f\xda\x5a\x25\x51\xda\x85\x91\x16\xe0\x34\x32\x46\x9f\xab\xed\xfd\x22\x0e\x82\xc5\x55\xdb\xf9\x45\xb5\xf5\x20\xaf\xa9\x71\x8e\x8a\x52\xd3\xc8\x21\xda\x16\x52\xa8\x06\xfe\xa4\x5b\xef\x72\xc7\x9d\x7b\xa1\xd9\xea\x88\x23\xf1\xeb\x79\xb6\x58\xb5\x0c\x9f\xad\x54\x23\xd8\x97\x94\x69\xe8\x19\x5a\xf3\xec\x19\xfd\x2d\x91\x26\x59\xe8\x5e\x52\x1e\xd6\x7b\xcc\x8d\xe2\xa2\x14\x85\x65\x5b\x86\xda\x4f\xec\xf6\x13\x5b\x9a\x7c\xbb\x2b\xd4\xda\x81\xdf\xd5\x1a\xef\x0f\x95\xc9\x68\x17\xf4\x50\xed\xa6\xea\x00\x97\x62\xc7\x04\xd6\xcd\x42\x06\x8c\xad\x1b\xaf\x69\xfa\xea\x43\xc3\x8f\x76\x4b\x8d\x45\xed\x4f\xa7\xfe\x48\x1a\x81\xf8\x23\xb1\x3f\x07\x7d\x58\xf3\x50\xad\x7d\xab\xb5\x8e\xab\x2f\x7d\xcf\x92\x6d\x38\x4b\xf8\x61\x9e\x24\x68\x0c\xdb\x70\x1c\x7a\xaa\x39\xb1\x14\x7d\x51\xb5\xe1\x4e\x5d\xfa\xf6\xf5\x2a\xf4\x66\x56\xf7\xeb\xc9\x9f\x99\x70\xaa\xdf\x76\x2d\x09\x5c\x54\xff\x4f\x17\xc7\x23\xa0\x48\xdd\xa7\xde\x7f\x01\x00\x00\xff\xff\x85\x93\xf3\xea\x03\x0e\x00\x00")

func cloudAwsTemplatesMysqlTmplBytes() ([]byte, error) {
	return bindataRead(
		_cloudAwsTemplatesMysqlTmpl,
		"cloud/aws/templates/mysql.tmpl",
	)
}

func cloudAwsTemplatesMysqlTmpl() (*asset, error) {
	bytes, err := cloudAwsTemplatesMysqlTmplBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloud/aws/templates/mysql.tmpl", size: 3587, mode: os.FileMode(420), modTime: time.Unix(1500220644, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudAwsTemplatesPostgresTmpl = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x9c\x58\x6d\x6b\x23\x37\x10\xfe\x9e\x5f\x31\xe8\xcb\xb5\xe0\x73\x9d\xb4\x77\x90\xa5\x14\x1c\x3b\x09\x86\x26\x5d\xe2\xd4\x07\x3d\x42\x91\x77\xc7\x46\x64\x2d\x6d\x25\xad\xef\x7c\x66\xff\x7b\xd1\xbe\x4b\xfb\x62\x27\xf9\x70\x5c\x34\x33\xcf\x3c\x7a\x34\x33\xd2\xe6\x78\x84\x10\x37\x8c\x23\x10\x89\x4a\x24\x32\x40\x02\x69\x7a\x01\x70\xbc\x00\x00\x20\xd3\x2f\xcb\x67\xdc\xc5\x11\xd5\x78\x27\xe4\x8e\xea\x15\x4a\xc5\x04\x27\x1e\x90\xab\xc9\xe5\xe4\xe3\xe4\xfa\xe3\xe4\x9a\x8c\x72\xef\x99\xe0\x21\xd3\x4c\x70\x45\xbc\x02\x01\x80\xdc\x44\x94\xbf\x3e\xd0\xef\x33\xc1\x39\x06\x95\x19\xc8\x1d\xf7\xbc\xdb\xff\x12\x1a\x99\xdf\xbf\x9a\x95\x27\xdc\x18\x64\xc7\x19\xd2\x11\x10\x02\x2f\x90\x8e\x4a\x4c\x5f\xb2\x3d\xd5\x78\x02\xa7\xf4\xca\x00\xb4\x4c\x30\x03\xc9\x30\x0a\x28\xe2\x53\x49\x77\xa8\x51\x5a\x94\xa7\x51\x24\x02\xaa\x31\x5c\x6a\x21\xe9\x16\x1b\x36\x00\xf2\x7c\x88\xcd\x0a\x79\x4c\x76\x6b\x94\x64\x54\x5b\xe6\xb8\xa1\x49\xa4\x8d\xf1\x72\x62\x1b\x54\x20\x59\xac\x0b\xe9\x2a\x7c\x50\x79\x02\x50\xec\x07\xc2\x4f\xf7\x37\x3f\x93\x22\xa8\xde\xeb\x9c\x6a\xba\xa6\xaa\x87\xc4\x52\x4b\xc6\xb7\x3d\x24\x68\x1c\x0f\xb0\x28\x1c\x21\x2c\x12\x00\xa7\x3b\x6c\xe7\x5f\x70\xa5\x29\x0f\xb0\xc8\xf8\x56\x0e\xe1\x7a\xac\xaf\xc6\x3b\x16\x48\x31\xc0\xa5\x4c\x02\x41\x44\x95\x82\x8d\x90\x0d\x5a\x22\x44\xd5\xe6\x75\x47\x77\x2c\x3a\xbc\x83\x51\x2c\x94\xde\x4a\x54\xd7\xe3\xcf\x03\x8c\xfc\xc2\x0b\xf6\x79\xc9\xc3\x26\xcf\xd7\xe2\xd1\x2e\xed\x37\xf2\x19\x22\x51\x96\xe7\xbd\x14\x49\x0c\x3b\xfa\xfd\xdf\xa0\xce\x05\x7b\x1a\x25\x38\x02\x36\xc6\x31\x7c\x38\xce\x6f\x4a\x15\x67\x46\xc4\x07\xdc\x09\x79\xf8\xe5\xf2\xd3\x24\xfb\x49\x3f\x74\x50\x4f\x22\xcd\xa6\xff\xbc\x83\xf3\x86\x46\x0a\x07\x88\x67\xc8\x71\x84\x40\xf7\x94\x45\x74\xcd\x22\xa6\x0f\xf0\x43\xf0\x8e\x02\xf3\xa9\x52\xdf\x84\x0c\xcf\xa7\x61\xa5\x5a\xa2\xdc\xa3\x84\xb8\x44\x69\xe3\xd7\xc3\xe2\x1d\xf0\x33\x89\x54\x23\x30\x0e\x71\x8e\x03\x2a\x59\x73\xd4\xea\x5c\x65\x4c\xb3\x7f\xc3\x70\x65\xce\x2a\x9f\x50\xf9\x2c\x1a\x95\xbe\xf0\xd2\xa2\xbc\x2c\x52\x74\x52\xfe\x93\x29\xfd\xfb\xf4\xcb\xd2\xf3\x6e\x67\x57\x9e\x97\xfb\x7a\xde\x22\xfc\x63\x60\x1b\x2b\x7f\x56\x11\xef\x4b\x37\x28\xd4\xb9\x59\xcf\x29\x6b\x43\xc6\x55\xb3\x45\xea\x6f\x85\x32\x9b\x48\xef\xef\xf0\xd3\x55\x93\x94\x49\x5a\xe9\x57\x71\xd0\x9d\xb9\xd6\x60\xe5\xcf\x8c\x00\xc3\x1b\xed\x44\x9e\xb1\x50\xda\xe8\x1d\x0a\xcd\x16\xf3\x27\xb8\x89\x44\xf0\xda\xcc\xe0\x08\xd0\x46\xaf\x6e\xe8\x01\xf4\x62\xa4\x89\x0d\xf8\x9d\x52\x55\x4a\x5e\x8f\x3f\x8f\x2f\xcf\x48\x6f\x5d\xaa\x7f\x25\x3a\x4e\xac\xea\x25\xbe\x90\xfa\xd3\x6f\xbf\x5e\x3d\x07\xf1\x34\xcc\xf7\x0e\x24\x6b\x89\xfa\x0e\xbf\x47\x3d\xd5\x3a\xef\x90\x72\x92\x99\x2e\xb9\xe5\x61\x2c\x18\xd7\x63\x13\x89\x4a\x65\xb7\x78\xb3\xbf\x6b\x6c\xf3\xdf\xf7\x61\x67\x91\x0e\xf0\x2d\xdf\x97\xf2\x34\x2f\x61\x0b\xbc\x78\x6a\x54\xf6\x3e\x80\xe6\x90\xeb\x02\xa8\xec\x7d\x00\xcd\x76\xe8\x02\xa8\xec\x06\xc0\x3a\x8e\xa7\xe2\x69\x67\x1d\xc8\x12\x83\x44\x32\x7d\xc8\x2e\x96\x53\x85\x6e\x3b\x37\xaa\xc1\x97\x22\x46\xa9\x19\xda\xa3\x0a\x80\x64\xae\x4e\xd9\x95\x6d\x09\xd5\x63\x73\xd4\x0c\xb1\xb2\x2c\xf8\x36\x3b\x6b\x0f\xbe\x36\x7c\xc0\x6c\x78\x11\xfb\x52\x68\x11\x88\xc8\x80\xea\x20\x36\x07\x79\x27\xc5\xae\x38\x7d\x62\x8a\xc1\xac\x3d\x0b\x77\xc5\xf4\xdd\x22\xb6\x74\x2b\xbb\xb1\x92\x2d\xff\x79\xb1\xa8\xad\xe2\x60\x11\xba\x61\xa4\x11\x90\xf6\xcc\xd4\x53\xf2\x3e\xcd\x97\x9e\x37\xbf\x69\x3a\x9f\x25\xaf\x15\xf2\x46\x99\xb3\xc0\x45\xd8\x78\x85\x2f\x36\x79\x67\x94\x17\xc0\xc8\x95\xbc\xd8\xb4\x73\x4f\xd4\xbb\xed\x71\x24\xb6\xa4\x83\x7a\x55\x5d\x79\x52\xac\xba\x7f\xcf\x51\xaa\xeb\x2d\x5f\xf1\x6c\x19\xed\x2d\x11\xe7\x49\x65\xc5\x5a\xef\xe2\xde\xb8\x45\x88\x5c\xb3\x0d\x43\x69\x27\x36\xfb\x59\x6a\x1a\xbc\x3e\xe6\x3d\xeb\x84\x3f\x56\x9d\xde\x1e\x30\x8e\xab\xfd\x46\x6c\x05\xda\xe6\x76\x78\xa3\x8e\x5a\xb1\xcd\xb2\x74\x02\x6f\xf9\x96\x71\xec\xb9\x6c\x2b\x7b\xe3\x32\xaa\x3b\xa7\x58\x73\x00\x1f\xa8\xd2\x28\xed\x29\xd7\x1e\x6d\x3d\x21\xf6\x74\x6d\x8f\x54\x3b\xac\x7e\xf5\xd6\x9f\x9c\xc5\x9a\xe3\x6a\xcf\x0f\xcb\x92\xac\x23\x16\x44\x87\x69\x10\xa0\x52\x6c\x1d\x61\xd7\xe3\xcf\xb4\x5b\x5e\x59\x65\x21\x6f\x63\x07\x68\xe5\xcf\xac\xc1\xe7\x7e\xc4\xda\xb3\x17\xd2\xea\xb5\xd8\xd5\x44\xce\x61\x9f\x6c\x25\xc7\xff\xbc\xd1\x63\x4d\x9b\xd3\x15\x5a\x7f\xa7\x55\xae\xc5\x92\x2b\x76\xd7\xb7\x78\x61\x73\xbe\x7c\xda\xa3\xab\xeb\xaf\x0c\x23\xb7\xe1\x1e\x45\x7e\x6f\x42\x3a\x1a\xfa\x73\x83\x35\xa9\xa0\x6b\x6a\x5d\x94\xff\xa6\x17\xc7\x23\x20\x0f\x21\x4d\x2f\xfe\x0f\x00\x00\xff\xff\xe6\x5d\x23\xc0\x4e\x11\x00\x00")

func cloudAwsTemplatesPostgresTmplBytes() ([]byte, error) {
	return bindataRead(
		_cloudAwsTemplatesPostgresTmpl,
		"cloud/aws/templates/postgres.tmpl",
	)
}

func cloudAwsTemplatesPostgresTmpl() (*asset, error) {
	bytes, err := cloudAwsTemplatesPostgresTmplBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloud/aws/templates/postgres.tmpl", size: 4430, mode: os.FileMode(420), modTime: time.Unix(1500220644, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _cloudAwsTemplatesRedisTmpl = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xac\x57\x51\x6f\xe3\x36\x0c\x7e\xef\xaf\x20\xf4\x9c\x06\xb9\xdc\xb0\xad\xc6\x6e\x40\x96\xa6\x07\x03\x6b\x17\x24\x59\x0e\xd8\xa1\x0f\x8a\xc4\x24\xc2\xd9\x92\x27\xd1\xed\x8a\x20\xff\x7d\x90\xed\x38\x96\x9d\xa4\xbd\xe2\xfa\x50\x14\x26\x45\x7e\xfc\xf8\x91\x52\x77\x3b\x90\xb8\x56\x1a\x81\x59\x74\x26\xb7\x02\x19\xec\xf7\x57\x00\xbb\x2b\x00\x00\x36\xfa\x32\x5f\x60\x9a\x25\x9c\xf0\xce\xd8\x94\xd3\x12\xad\x53\x46\x33\x88\x80\x0d\x07\x1f\x06\xd7\x83\x9b\xeb\xc1\x0d\xeb\x95\xee\x63\xa3\xa5\x22\x65\xb4\x63\x51\x15\x02\x80\x4d\xad\x7a\xe2\x84\xfe\x13\xb0\x3b\x1d\x45\x93\x7f\x73\x9e\x78\x97\xaf\xfe\xcb\x0c\xd7\x2c\x3a\x7a\xc1\xbe\x07\x8c\x6c\x8e\x0c\x1e\x61\x5f\xc4\xd8\x57\xe1\xa7\xdc\xf2\x14\x09\x6d\x10\x7e\x94\x93\x49\x39\x29\x71\xc7\x55\x62\x9e\xd0\x4e\x34\x5f\x25\x28\x1b\x3e\x00\x6c\xf1\x92\x79\x04\x6c\x4e\x56\xe9\x4d\x05\xb8\xb0\xdc\xe2\x9a\xe7\x09\x79\xe3\x9a\x27\x0e\x43\x9b\x13\x56\x65\xbe\x24\x6f\x8f\xb5\x54\x82\x13\x3a\x78\xde\x22\x6d\xd1\xc2\x7d\x9e\x90\xba\x1e\xfd\x03\xca\x01\x96\x79\xfb\x70\x9f\x3b\x82\x15\x02\x17\xc2\xa4\x19\xd7\x0a\x25\x3c\x2b\xda\x42\xac\x1d\x71\x2d\xd0\x83\xf9\x24\xb8\xd8\x62\x3f\xfd\xd8\x4f\x51\xaa\x3c\x05\x63\x61\xab\x36\x3e\x28\xd7\x12\x1e\xf2\x74\xec\x1d\xc6\x49\xee\x08\xed\xa7\xe1\xd1\xde\x67\x15\xc0\xfd\x01\x29\xbb\xe5\xc4\x57\xdc\xe1\x89\x9a\xe1\x62\xd1\xde\x3a\x38\x57\xb1\x37\x56\x8e\x20\xab\x0c\xa0\xb4\xc4\xff\xba\x08\x9a\xa5\xbd\x83\xf9\x92\x0c\x1a\xf6\x53\x25\xac\xb9\xd0\x82\xc5\x16\x81\x5e\x32\x04\xb3\x06\x55\xe5\x04\x32\x90\x3b\xec\x82\x6a\x28\xef\x6d\x78\x82\x54\x63\x8b\x9c\x7c\xc1\x90\x95\x71\xc0\xe5\x2b\x8d\xe4\xde\xaa\x9e\x51\x92\x98\x67\x94\x4b\x9e\xe4\x58\xca\xbd\x14\x76\xef\xe0\x0b\x8f\x1d\xc8\xad\xbe\xbb\x77\x70\xf9\xe1\x15\xfa\x74\x9e\xae\xd0\x7a\x02\x0b\xd6\x41\x54\xa9\x60\x6d\x2c\xd0\x56\x39\xb0\x98\x25\x5e\xe8\xca\x68\xd8\x58\x93\x67\x5d\x6a\xe7\x15\x15\x27\xe1\xfd\xa9\x1c\xfd\x36\xfa\x32\x8f\xa2\xc9\x78\x18\x45\xa5\x6f\x14\xc5\xf2\xf7\x0b\xd0\x96\xd3\x71\x4d\xf0\xb9\x74\xe7\x1b\x0a\xdf\x91\xf6\x28\xfd\x4b\xca\xf7\x78\xda\x8d\xef\xe0\x5a\x66\xe2\x34\x05\x47\x18\xcb\xe9\xd8\x63\xb8\x5c\xf9\xc9\xc8\x63\x25\x6d\x18\xfd\x04\x65\xe3\xf8\x76\x06\x7f\x24\x46\x7c\x6b\x66\x68\x09\xe5\x10\x3d\x58\xa7\x7f\xe5\x94\xe5\x14\xae\x6a\x63\xe9\xe7\x8f\xbf\xdc\x2c\x44\x36\x92\x65\x72\x60\x85\x7e\x8f\xdb\xfb\x33\xd2\x88\xa8\x94\xf3\xec\x28\x94\xcf\x85\x4e\x7a\xc5\xd0\xa5\xdc\xbe\x4c\xb4\x9c\x1a\xa5\xa9\xef\x03\xa1\x73\xc5\x3a\x6f\xce\xe6\x31\x95\xff\xf3\x87\xa4\x2a\x02\xb5\xf2\x4c\xf4\xd3\x0c\xa5\x72\xcd\x15\x19\x24\xaa\xee\x9f\xda\xee\x4f\x07\x3c\xcd\xaa\xab\x31\x60\x6a\x8e\x22\xb7\x8a\x5e\x4a\x2c\xaf\x48\x20\x74\x6e\xb4\x69\x6a\x4d\x86\x96\x14\x86\x83\x04\xc0\x0a\xd7\x56\xbb\xad\xaf\x03\xea\x9b\xba\xd7\xf4\x0f\x52\xc4\x7a\x53\x50\x1e\xc1\xd7\x86\x0f\xf8\x6a\xe3\x6c\x6a\x0d\x19\x61\x12\x1f\x91\x44\xc1\xe3\x9d\x35\x69\xd5\x04\xe6\x7b\xe2\xbf\x2d\x4c\xfb\x8b\x97\x63\x9c\x05\xa4\x1d\x44\x5a\x73\x56\xfe\x3c\x06\xd0\x96\x99\x88\x65\xfb\x18\x6b\x1c\xd8\x77\xc4\x5f\xac\xc0\x72\x7e\x5f\x25\x38\xe1\x8e\x54\x71\xe0\x30\xf2\xdf\x41\x73\x8b\xe1\x42\x29\xd5\xb4\x57\x9b\x2f\x64\xb9\xb0\xc4\xd2\x1d\x35\x1a\xaf\x4b\x7d\x1e\x56\x53\xaf\xcd\x78\x55\x73\x6b\x83\x1d\x8b\x3d\xe3\xc8\x42\x46\x2f\xd2\xd5\x99\x8d\x37\xd2\xd5\x9d\xa9\xb7\x70\x76\xe9\xc1\x55\x57\x71\xd6\x29\x2c\xbc\xf0\xbb\x57\xda\xd8\xea\x65\xf9\x77\xb6\xb1\x5c\x7a\xbc\xfe\xaa\x0c\x5c\x0b\xd0\x0f\x46\xd6\x6f\x8c\x3a\x59\xf0\xf8\x68\x25\x68\x4b\xe9\x81\xa7\xe1\xe1\x8e\xd6\x5a\x01\x26\x7a\xa3\x34\xd6\xf3\xc7\x4e\x18\x0f\xaf\x62\xff\x28\xee\xff\xda\x1f\xfe\x14\x3a\x9d\xba\xd3\xeb\xf4\x1d\x63\x2b\x7d\x38\x85\x4d\x4b\xbb\x7b\xa1\x96\x1b\xad\xf0\x7d\x9f\x13\x17\xdf\x8a\xda\x5b\xf1\xc3\xc5\x21\xdb\xcf\xf2\x70\x75\xc1\xbe\x7e\xb2\x34\x74\x78\x75\xf8\xbd\xbf\xda\xed\x00\xb5\xf4\xff\x43\xfc\x1f\x00\x00\xff\xff\xc0\xd8\x0d\xc3\x5c\x0c\x00\x00")

func cloudAwsTemplatesRedisTmplBytes() ([]byte, error) {
	return bindataRead(
		_cloudAwsTemplatesRedisTmpl,
		"cloud/aws/templates/redis.tmpl",
	)
}

func cloudAwsTemplatesRedisTmpl() (*asset, error) {
	bytes, err := cloudAwsTemplatesRedisTmplBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cloud/aws/templates/redis.tmpl", size: 3164, mode: os.FileMode(420), modTime: time.Unix(1500220644, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"cloud/aws/templates/app.tmpl": cloudAwsTemplatesAppTmpl,
	"cloud/aws/templates/mysql.tmpl": cloudAwsTemplatesMysqlTmpl,
	"cloud/aws/templates/postgres.tmpl": cloudAwsTemplatesPostgresTmpl,
	"cloud/aws/templates/redis.tmpl": cloudAwsTemplatesRedisTmpl,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}
var _bintree = &bintree{nil, map[string]*bintree{
	"cloud": &bintree{nil, map[string]*bintree{
		"aws": &bintree{nil, map[string]*bintree{
			"templates": &bintree{nil, map[string]*bintree{
				"app.tmpl": &bintree{cloudAwsTemplatesAppTmpl, map[string]*bintree{}},
				"mysql.tmpl": &bintree{cloudAwsTemplatesMysqlTmpl, map[string]*bintree{}},
				"postgres.tmpl": &bintree{cloudAwsTemplatesPostgresTmpl, map[string]*bintree{}},
				"redis.tmpl": &bintree{cloudAwsTemplatesRedisTmpl, map[string]*bintree{}},
			}},
		}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}

