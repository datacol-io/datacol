// Code generated by go-bindata.
// sources:
// cmd/provider/aws/templates/formation.yaml
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

var _cmdProviderAwsTemplatesFormationYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xec\x7c\x7b\x73\x1b\x37\x92\xf8\xff\xfc\x14\x1d\x46\x1b\xc9\x59\x0f\x5f\x92\xbd\x32\x7f\xeb\xd4\x8f\xa6\x64\x9b\x27\x59\xe2\x8a\xb2\x52\xc9\x25\xe7\x02\x67\x9a\x24\x56\x43\x60\x02\x60\x24\x31\x3e\xdd\x67\xbf\xc2\x6b\x1e\x9c\xa1\x1e\x4e\x6c\xdf\x6d\x5d\x5c\x95\x12\x81\x46\xa3\x5f\xe8\x6e\x34\x80\x69\x04\x41\xd0\x18\xfc\x38\x39\xc7\x65\x12\x13\x85\xaf\xb9\x58\x12\x75\x81\x42\x52\xce\xfa\xb0\xdd\xeb\x74\x3b\x41\xe7\x45\xd0\x79\xb1\xdd\x38\x40\x19\x0a\x9a\x28\xdb\x73\x94\x4e\x51\x30\x54\x28\x61\xf0\xe3\x04\x86\x31\x4f\x23\x3b\x9a\x72\x06\x1e\x5f\x1f\x86\x02\x89\x42\x20\x90\x0f\x68\x00\x84\x71\x2a\x15\x0a\xa0\x0c\x08\x30\xbc\x86\x8b\xf1\xb0\x05\xe7\x0b\x84\x25\x31\x1d\x8c\x47\x08\x54\x02\x61\x40\x52\xc5\x03\x81\x21\xbf\x42\x41\xd9\x1c\x06\x4b\xf2\x3b\x67\x70\x38\xec\x35\x00\x28\x93\x8a\xb0\x10\x5b\xd0\x0d\x7a\x1d\x20\x51\x44\x35\x01\x24\xd6\xfd\x59\xaf\x34\x13\x31\x18\xa4\x8a\x4f\x42\x12\x53\x36\x7f\x23\x78\x9a\xc0\x3f\x39\x65\xa0\x16\xd8\x80\x02\x7d\x19\x75\x44\x1a\x3a\x64\x0b\x06\x0c\x0e\x8f\x5f\x41\x22\xf8\x15\x8d\x34\x04\x67\x33\x3a\x4f\x05\x99\xc6\x08\x78\xa3\x50\xe8\x29\x49\x18\xa2\xd4\xec\x29\xae\x91\x16\x51\x0e\xc6\x23\xcb\x9f\x63\x16\x28\x0b\xe3\x54\xa3\x22\x30\x25\xd2\x08\x6d\xc1\xa5\xd2\x43\xe7\x82\x30\xd5\x00\x98\x4c\xde\x3a\x94\x1e\x61\x22\xe8\x95\x96\xa6\x4c\xa7\x0c\x15\xcc\xb8\x30\xcd\x8e\x5e\x3d\x01\x95\xa0\x9c\xec\x21\x34\xa2\x37\xf4\x5c\x73\x90\x8a\x84\x97\xb2\x0f\x9c\x61\x36\xd0\x13\x43\x58\x54\x6a\x2f\x20\x44\x3b\x10\xa8\xc6\x23\x53\xaa\x0c\xcb\x1a\x30\xc2\x2b\x8c\x79\xb2\x44\xa6\x0c\x02\xb9\x24\x71\x0c\x92\xb2\x79\x8c\x81\x42\xb2\xf4\x68\x64\x0b\xbe\xff\xfe\xc7\xc1\xd9\xc9\xe8\xe4\xcd\xf7\xdf\x1b\x22\x35\x4d\x6b\x64\xc2\x8c\xa7\xa2\xa0\xdc\x82\xf2\xae\xa9\x5a\x40\x84\x33\x92\xc6\x0a\x24\x2a\x45\xd9\x5c\xb6\xe0\x27\x9e\xc2\x35\x8d\xe3\x06\xc0\x14\x61\x4a\xe3\x18\xa3\x8c\x05\x6d\x94\x02\x25\x4f\x85\xc6\x90\x4a\x8c\x80\xce\x60\xc5\x53\x37\x1f\x10\xc7\xd8\x4c\xf0\x25\xa8\xa2\xe0\x5a\xdb\x8d\xc6\xb7\x86\xf5\x77\xa8\x48\x44\x14\x01\x85\x71\x6c\x0d\x7d\xc1\xaf\xb5\x36\x22\x2a\x93\x98\xac\xac\x56\x88\x20\x4b\xd4\x8c\x42\x94\x1a\x13\xb5\x98\xcd\x44\x94\xb3\x86\x47\xd3\x6f\x80\xc6\xd1\xef\x97\x57\x4b\xbf\x3f\x62\x0a\xc5\x8c\x84\xa8\x21\x00\xc6\x1e\xa1\x31\x52\x69\x1b\x03\x38\x26\x53\x8c\xed\x0f\xfd\x9f\x13\x48\x1f\xce\xf0\xb7\x94\x0a\x8c\x5c\x4f\x36\x5a\x7a\xd8\x00\x7e\xe6\x0c\x3b\xa5\x5f\xdd\xec\xd7\x20\x5a\x52\x36\x62\x73\x81\x52\x1e\xf3\xd0\x92\xec\x3b\x8f\x70\x75\x42\x96\x78\x0f\x05\x83\xe8\x4a\x6b\xea\x2e\x0a\x4e\x50\x5d\x73\x71\x49\xd9\x7c\x6c\x17\x91\xc8\xe7\xd8\x97\x27\x3c\xc2\x21\x49\x48\x48\xd5\x2a\x6b\x1f\x39\x03\x38\x5f\x25\x98\x35\x1e\x50\x79\x39\xa1\xbf\xe3\x9b\x69\xd6\xf4\xca\xae\x9f\x5a\xf0\x7f\x4c\x26\xbb\xaf\xd2\xf0\x12\x55\xc6\x86\x6f\x3e\xc2\xd5\x58\xe0\x8c\xde\x34\xca\x42\x37\x4c\x66\x94\x3b\x01\xd4\x30\xad\x97\xe7\x11\x7a\x72\x8d\x84\xeb\x44\x73\x45\x68\x4c\xa6\x34\xa6\x6a\x65\x80\x0a\xf0\xdd\x3a\xac\x18\x72\x16\x6d\x1c\x56\xa7\xac\x5a\x85\x2c\x29\x03\x07\x07\x6b\x5a\x2d\x0a\xaa\x66\xac\xef\x86\x82\x20\x73\xa9\xd7\x0c\xd0\x9d\xa0\x7b\x61\xe7\x0d\x7d\xf5\xc4\x01\xd4\x68\xe5\xbe\xc9\x60\xc7\x0d\x82\xb7\x5c\x2a\x8f\x68\xcd\x3c\x6a\x90\xe8\x6e\x58\x33\x9f\xb2\xe2\xeb\x04\xbd\x0b\xb6\xbf\x30\x20\x33\x89\x7a\xf8\x23\x5c\x81\x33\x19\xdb\x59\x35\xea\x3a\xea\x32\x20\xc8\x4c\xdf\xf9\x97\x7c\xa1\x00\x89\x63\xed\x59\x16\xa8\x7d\x95\xd0\x2e\x26\x21\x52\x42\x98\x4a\xc5\x97\x99\xd3\xf3\x81\xc0\x3a\x98\x29\xce\xb8\xc0\xdc\xcf\x94\xd7\x5d\xc9\x72\x4b\xc1\xfb\xf0\x86\x4a\x8d\xce\xb8\x59\xcd\x33\xa1\xc2\x38\xce\x3c\xe2\xb4\xcc\x28\xa3\x34\xeb\xb2\x0e\x87\xbd\x7e\xdf\xc1\x9a\x3f\xb2\xf5\x34\xe4\x4c\x2a\x41\x28\x53\xa5\x49\x96\xa9\x54\xda\x2f\x9b\x38\x43\x96\x08\x7c\xa6\x43\x30\xd6\xcc\xdd\x6a\x58\x72\xdf\x11\x85\x82\x92\xf8\xa1\x24\x7b\xf8\x3b\x68\x9f\x28\xed\x90\x1b\x00\xda\xf5\x26\xf4\x08\x57\x35\xc8\xc7\x44\xca\x6b\x2e\x6c\xf0\x38\x20\x8a\x84\x3c\xd6\x6c\x29\xc1\xe3\x18\x85\x0e\xde\x50\xc5\xd9\x00\x0f\xeb\xf3\xa5\x2a\xe6\x2b\xdb\xa3\x79\xaf\xc7\x5b\x8b\x76\x20\x14\x9d\x91\x50\x59\xfb\xac\x41\x7b\xe2\xe4\x39\xe7\x7c\x1e\x6b\x6b\xe0\x82\xcc\x11\xa6\x06\xde\x84\x27\x7e\xcd\x62\x4e\x22\x88\xdc\xac\x53\xca\x88\x58\xd5\xce\x36\xb1\xb6\xb5\x71\x32\x87\x55\xcb\x46\x4f\xa4\x95\x10\xea\x05\xa7\x63\x3e\xb2\x2b\x2a\x38\x33\x39\xc0\x15\x11\x54\x27\x07\xb2\x76\x96\xaa\x1f\x28\xeb\xb7\x10\xf0\x41\x69\x5f\xb0\x9e\x8b\xd4\x29\x55\x23\x71\xab\x6c\xb9\xd7\x8a\x89\x98\xa3\x9e\xeb\x5b\x98\x24\x18\xd2\x19\x45\x69\x57\x8b\xf6\x4d\x7c\x66\xfe\x16\x9c\x2b\x1d\xbd\x2f\xcd\x04\x3a\x63\x29\xe5\x1a\x4f\x5d\x62\xa6\xb9\xb4\xa9\xa8\xc1\xa7\x79\xb5\xa9\x60\xa3\xea\x0c\xcb\x89\xf1\x64\xe3\x64\xba\xa5\x32\x19\xbc\xa1\xaf\x5a\x05\x3e\xf6\x3a\xdb\x65\xc6\xf6\x3a\x05\xce\x4f\xd2\xe5\xd4\x45\xce\x77\x94\x5d\x90\x38\xc5\x3e\xec\xdb\xdf\xe4\xc6\xfd\xee\x76\x7a\x7b\x5a\x0c\x1b\x3d\xf0\x03\x25\x5f\xca\x4b\x77\x4c\x0a\xa5\xbb\x92\x74\x1a\xd3\xb0\xb0\xde\x9e\xdc\xa7\x1b\xd5\x6b\x2d\x69\x28\xb8\x69\x1d\x68\x47\x87\x91\xa1\xb5\x90\x1f\xa8\x5e\x8b\x11\xc6\x8b\xbf\xf3\x31\xae\xc1\x64\x98\x25\x08\x8c\x68\xba\x2c\xb6\x58\x1b\xf0\x0d\xcb\xdd\x75\x90\xe5\x6e\x15\xe4\xa6\xd2\xd2\x5b\x6f\xda\x5b\x1f\xb5\x57\x1c\x75\x8f\x07\x24\x70\x45\x62\x1a\x55\x25\x6d\xfc\x5e\x21\x71\x28\x8d\xd6\xf1\xc1\xe7\xbc\x95\x6c\xc0\xe9\x88\x4a\x1b\x09\x5a\x99\x77\xd1\xfb\xa4\xe5\x12\x59\x24\x1d\xad\x6a\x41\x94\x49\x7b\x45\xca\x4c\x8e\xef\x37\x37\x89\x76\x40\x3f\x1b\xcb\x4e\x25\x82\xe2\x5c\x6f\x8b\xb4\xf3\x08\x39\x17\x11\x65\x26\x49\x0e\x05\x97\x12\x06\x3f\x6f\x08\x08\x45\xc2\x34\x5d\xfd\xfe\x27\x86\x05\x9d\x5a\x57\x73\x9e\x46\x29\x4f\xaa\x88\x47\x6e\xc8\x95\xd6\xa5\xf3\x95\x48\xdf\x9c\xab\x95\xd0\x0e\x47\x07\x67\x30\x8d\x79\x78\x09\x3b\xa3\xb1\xde\xc2\x9a\xa4\x4d\x10\x36\xc7\x27\x5a\x1f\x36\x35\xa8\xee\x06\x9d\x86\x4b\xcb\x54\xab\xf3\xed\xf9\xf9\x78\xb2\x06\xbb\xbe\x15\x7d\x2f\x11\x3a\x2d\xf3\xaf\xed\x37\x06\xd9\x54\x6e\xa8\xd9\x19\x69\x0f\x19\x3b\xe2\xeb\x23\xab\x71\x45\xc7\xc8\xe6\x6a\xd1\x87\xed\x17\xdb\xde\x1b\x65\x4d\xdd\xfd\xed\xe2\xca\x1f\x13\xa5\xf7\xcb\x7d\x68\xee\xfc\xf2\x4b\xf4\xb1\xfb\x74\xf7\xf6\xc9\x2f\xbf\xb4\x1e\xf2\xa3\xed\xfe\xec\xdd\x3e\x69\x3e\x66\xe5\x8d\xc6\x56\xca\x46\xa6\xde\x39\xcf\xb8\x58\xc2\x4d\xcb\xfc\x6b\xdf\xd8\x1c\xa4\x2e\xd5\xcc\xdc\xd8\x76\x6f\xbb\xaa\xbd\x11\xa3\x4a\xa7\x20\xcc\x78\x66\x8d\xbb\x20\x6a\x13\x32\x60\xa7\x1b\xf4\x3a\x25\x27\x59\xeb\xc6\xb7\xbb\xdb\x6b\x8e\x7c\xbb\xe7\x82\xc1\x3d\x6c\x4e\x51\x5d\x23\x32\xe8\x1a\xf5\xf7\x3a\xe5\x28\xd3\x72\x21\xd1\xe7\xbb\x79\xf1\xc2\x62\x31\x3a\x37\x5b\x64\x61\xcc\xc5\x78\x83\x05\x52\x01\xfc\x9a\x99\x4c\x42\x2a\xa1\xb7\xf3\x92\x91\x44\x2e\xb8\x92\x06\x9f\x93\xe2\x6f\x29\x0d\x2f\xa5\x22\x42\x05\xe4\x5a\x06\x57\x49\x68\x88\x28\x34\xc7\x94\xa5\x37\x81\xb7\x52\xbf\xc9\xd6\x48\xea\x72\xf4\x8a\x91\xfc\xc7\xbf\x77\x82\x17\x24\xf8\x7d\x10\xfc\xfc\xeb\x5f\x77\xf2\x1f\xc1\xaf\xdf\x17\x7a\x9e\x7c\xbf\x75\xa7\x45\xfc\x43\xd3\x03\x13\x4d\x90\x4f\x68\xcc\x12\x0e\x09\xf3\xa5\x18\xa7\x42\xf9\x14\x34\x05\x22\x24\xd2\xfb\xfb\x18\x35\x31\xf2\x29\xa4\x49\x62\x3b\xf2\x26\xcd\xec\x62\x95\x2c\x90\x49\xd8\x09\x9e\xb4\x60\xa4\x34\x52\xc6\x15\x18\xf6\x81\x0b\x40\x16\xd9\x1a\x06\x71\xa0\x0e\xaf\x86\x6f\x94\x6d\xac\x20\x37\x81\x33\x14\xc8\x42\xac\x1a\xdd\x29\x8b\x57\x10\x2e\x8c\x35\x1b\x3f\xe7\x4a\x1b\x0b\x72\xa5\xbd\xa2\x82\x34\x01\x22\x25\x2a\xcd\x0c\xbd\x44\xdd\x69\xd5\xc9\xb2\xcd\x88\xa3\xa1\x64\x0c\x4f\x5d\x9d\x6c\xb2\xeb\x84\xe4\x4a\x4a\x9a\x49\xad\xec\xd2\x1e\x28\xaf\x7b\x68\x69\x98\xf9\xb3\x02\x98\xfe\xcf\xd2\x2b\x7d\x81\x05\x7d\x43\xdb\xa2\xf3\x76\xd0\x86\x88\x0a\x0c\x75\x66\x89\x52\x5b\x55\x46\xeb\x8c\x8b\x4b\x6f\x66\x6f\x51\xb3\xee\x5d\x5e\x51\x9b\x9e\xcf\x34\xd1\x09\x2f\x46\xda\x84\x27\xbb\xb6\x22\xa5\xb8\xc0\x08\x88\xdf\xe2\x6d\x7d\xb4\x5c\x69\xcd\xdf\xb6\xe4\x6e\x8b\x98\x6a\x13\xb9\x96\xad\x90\x2f\xdb\x5b\x1f\x13\xc3\xd8\x6d\xdb\x53\x2a\xf9\x12\x67\x34\xc6\x96\xba\x51\xad\xc9\xae\xad\x87\x15\xcc\xc7\xcb\xf0\x4e\x23\x7a\x94\xf9\xf8\x2d\xfa\x03\x8c\x28\x87\xaf\x6e\x75\x6a\x36\xb3\x77\x2d\xac\xe0\xd7\xbf\xee\xb4\x4b\x3f\x1f\xb3\x9e\x2e\x71\x05\x56\x72\x0f\x95\x84\xe3\xb2\x46\x1e\x05\x59\x58\xe1\xcc\xb8\xb8\x26\x22\x02\x19\x13\xb9\x80\x9d\xf6\xc6\x15\xe6\x70\x1a\x11\x55\x06\xc1\x14\x43\xe2\x1c\xdb\x0a\x88\x40\x53\x55\x5e\x12\x45\x43\x12\xc7\x2b\x20\x49\x82\x2c\xc2\xa8\x55\x5e\x8d\x0b\x63\x75\x6d\x63\xa7\xea\x8f\x2c\xc3\xf5\x75\xf5\x14\x88\x04\xbc\x49\x62\x42\x59\x56\x36\xb3\xa5\xe8\x82\xa3\xce\x16\x98\x35\xbc\xc9\xee\x17\x14\xb5\x8e\x5b\x9b\x85\x6d\xc4\xec\x50\xd7\x0c\x7b\x9c\xb8\xd7\xb7\x8a\x9b\xaa\x2a\x35\x3b\x87\x00\x42\x12\xd3\x90\xbb\x1f\xd7\x48\xae\xee\x4c\xe1\xb6\x87\xa9\x10\xc8\x54\xbc\x02\x99\x26\x09\x17\x0a\x23\x9d\x23\xa4\x28\x0d\x99\x4d\x8b\xae\x69\xe4\xd1\x34\xe8\x9a\x6b\xfb\xb1\xc2\x84\xe5\x34\x6e\xc1\xb9\x74\xc9\x61\x5e\xf1\x71\x27\x06\xc2\x07\x56\x9d\x99\xea\x0c\x3d\x65\xd4\xe6\x54\x3e\x78\x3b\x61\x26\x3c\x92\xde\x10\xaa\x27\x12\x2d\x98\x64\x54\x97\x1c\xb7\xa5\xde\x92\xe6\x9d\xe4\xce\x42\xa9\xa4\xdf\x6e\x47\x3c\x94\xad\x44\xf0\x7f\x62\xa8\x2c\x44\x8b\x8b\x79\xfb\xaa\xd7\xea\xb4\xe7\x76\xef\x1f\x18\xdd\x62\xd4\xbe\xcc\xa6\x6c\x9b\xfc\x21\x8e\x0d\xfa\xb6\xce\x2d\x5d\x37\x89\x96\x6d\x5f\x9a\xd3\x52\x32\x42\xb2\x93\xc9\x7e\xbb\x3d\xa7\x6a\x91\x4e\x8d\x3f\x35\x3d\x5a\x10\xd2\xfe\xd9\x9e\xc6\x7c\xda\xb6\xbb\xea\xb6\xa4\x0a\x0d\xbe\x80\x44\x11\x67\xad\x65\x54\xe7\xcc\x6c\xc1\x7c\x60\xc2\x4a\x56\xbb\xa9\x81\x98\x60\x28\x50\xdd\x03\x67\x8f\xa2\x8e\xf6\xe5\xd0\x4a\x73\x2d\xbd\x53\x22\xad\x09\xb5\xd9\xf9\x55\x2e\x1a\x57\x7d\xe3\x02\xf4\xb2\xd8\x59\x72\xa9\xed\x29\xdb\x1f\x17\xcf\x45\x92\x54\x24\x5c\x62\xeb\xc9\x9d\x5b\x5f\x3f\xb1\xfe\x31\x23\xb1\xcb\x3b\xd6\xc9\xbf\x18\x0f\x75\x12\x5b\xc3\x5b\x89\x8d\x66\xd7\xe7\xf5\xdd\xe7\xcd\x2a\x3f\xcd\x8b\xf1\xb0\xb0\xe5\x68\xda\x78\x31\x31\x07\x4a\x9d\x87\xe3\xef\xb6\x3a\xed\xde\x5e\x1d\xfe\xb1\x2d\x0f\xb8\x23\xaa\x4e\x69\xae\x6c\xa6\xee\xc3\x67\xea\x3d\x70\xa6\x6e\xfd\x4c\x63\x7b\x64\xf6\x08\xd6\x76\x37\x4f\x58\x3e\x7e\xdb\xc0\x9b\x83\x7a\x04\x8b\x7b\x0f\x9d\x71\x8d\xc7\x77\x24\x49\x28\x9b\x1b\x53\x3a\xc3\x39\xe5\xec\x1d\x49\xec\x94\x24\x09\x18\x17\x6a\x81\x44\xaa\x20\x3b\x67\xd8\x7e\xbe\xb7\xdd\x07\xb2\xa4\x41\x77\x9f\xcc\xc2\xbd\xbf\xcd\xaa\xc0\xbd\x2a\xf0\x8b\xdd\xe8\x79\xa7\x33\x8b\x3c\xb0\xe4\xa9\x5a\xd4\x61\xdd\x7f\x86\xb3\x17\x5d\x24\x25\xc0\x4d\x24\x4c\x7b\xfb\xbd\xee\xf3\xa8\x5b\x05\xae\x21\x01\x9f\x4f\x9f\xed\xe3\xfe\xb3\x86\xcd\xb7\x82\x10\x99\x12\x24\xae\xc3\xfb\xb7\x6e\xb7\x47\x3a\x5d\x0b\x8a\xe9\x5d\xa0\x33\xdc\xeb\xec\x77\x5e\x74\x3d\xe8\x35\xd6\x93\x1a\x92\xfd\x0e\xe9\x4c\x5f\x94\xe0\xea\xa8\x8c\xb0\x87\xfb\xfb\x16\x4e\x92\x60\x13\xeb\xd8\xf9\xdb\x33\x8c\xf6\x43\xd3\x91\xca\x8d\x70\xcf\x3b\xd3\x59\x67\xd6\x9d\x95\xe0\x6a\xe6\x9d\x85\x61\xf7\xc5\xf4\xc5\x0b\x0f\xb7\x89\x8f\x69\xe7\x59\xaf\xb3\x1b\x75\x4a\x70\x35\xf8\xa6\xbd\x68\xef\xf9\x6e\xd4\x6b\x34\x86\x9c\xd9\x13\x75\x63\x64\xef\xe5\x21\x91\xaa\x9b\x35\xda\x91\xaf\x59\xbf\x7f\xf8\x5b\x4a\xe2\x2c\x28\x7f\x73\x86\x33\x5b\x6a\xb1\x66\xe9\x9a\x9b\x19\xab\xda\xd6\x27\x0b\x9e\xc6\x51\xe6\x94\xfb\xb0\x01\x59\xe6\x1e\x1d\xde\x75\x37\xde\x68\x9c\xf9\x73\xde\xbe\xd9\x99\x66\x3f\x8d\x3b\x76\x67\xdc\xd6\x81\xf6\x6b\x4b\x41\xb6\x1b\x60\x2c\x78\x82\x42\xd1\xdc\x39\x0f\x69\x24\x5e\xe9\xd5\xa6\xb7\xe7\xb9\x5f\xdd\x76\xdd\x87\x8c\x4c\x63\x3c\x60\xd2\x85\xe5\x3e\x6c\x6b\x6a\x2b\xdd\x6f\xb9\x34\x9b\x0e\xb9\x06\x70\x4e\xe6\x85\x38\xa0\x03\x17\x14\x8e\x29\x01\x5c\x1d\x20\x17\xe7\x44\x47\x1e\x03\xd2\x00\x38\x78\x3b\x1c\x9f\x26\x99\x72\xaa\x7c\x15\x00\x36\xf0\x77\xc0\x97\x84\xb2\xf2\x09\xd9\xb7\xb9\x45\x02\x43\x8c\x24\xb4\x30\xec\xb5\x28\xb3\xb7\x1c\x9e\xda\xca\x36\x4a\x95\x55\xb9\x8d\x8e\x25\xcc\x51\xc1\xdf\xed\x8f\x1f\x74\x02\x90\xa4\x0a\xb3\x61\xad\x02\xfe\x09\x22\x14\x93\x13\xbd\x01\xb3\x5b\x31\x93\x37\xd8\x3b\x00\x17\xe3\xa1\x4b\xbc\xdb\xef\x25\x8a\x37\x29\x8d\xb0\x7d\x31\x1e\x7e\xd0\x5c\x7d\x70\x6c\xb5\x16\x6a\x19\x67\x98\xb5\xed\x8c\x66\x39\x27\x41\xc5\x60\x0b\x5d\xcd\x22\x53\xcd\x42\xc7\x37\x93\x74\x0a\xcd\xad\x8f\x05\xfb\xbd\xad\xb0\xd3\xac\x08\x70\x82\xe2\xaa\x74\xe2\x6d\xd9\x70\x29\x6c\x74\x70\x32\x71\x41\xbc\xa0\x95\x81\x94\x3c\xa4\x85\xc2\x60\x8d\x65\xd6\x83\x6f\xd0\xe7\x45\x12\x8e\x22\x67\x30\xde\xaa\x01\x0e\x16\x61\xe2\x50\x64\xbd\x45\xdb\x30\xc7\x33\xca\x64\x37\x6f\x88\xc2\x6b\xb2\xaa\xa7\x66\x0d\x68\x03\x0d\x75\x46\x6d\x13\xe2\x75\xbb\xb6\x01\xdc\x89\xc5\x21\x1d\x28\x45\xc2\x85\x4e\x9e\x36\x4a\xa4\x02\xf9\x28\x61\xac\x31\x91\x41\xac\x33\xa7\x5d\xc9\xb7\x59\x96\x6d\x72\x79\x9b\x6f\xe8\xa4\xb7\x7c\x11\x47\x67\xec\x92\x46\x08\xbf\x73\x86\x81\x8e\x7e\x2e\x6e\xbb\xb4\xaa\x9e\x13\xdb\xf9\x28\xda\x0b\xee\xc8\xf4\x54\x53\x1c\x07\x58\x29\x6e\x5b\xf8\xe2\x35\x90\x07\xfa\x9e\x72\x06\xf2\x30\xb5\xda\x21\x5a\x82\x56\xc3\x5f\x46\x0c\x9f\x89\xff\x62\x92\xf9\x18\xab\x2e\x81\xe6\x5b\x39\x1f\xb4\xd6\x06\xd5\xb9\x78\x0b\xf0\x8e\x24\x16\xe3\x28\x39\x65\xc7\x24\x65\xe1\xc2\xed\x55\x4c\xb0\x3b\x5f\x20\x9c\x0c\xce\x61\x34\xce\xce\xed\xca\xc6\x69\x4a\x0d\x12\x91\xd9\x42\x9c\xde\xb8\xbb\xfd\xa5\xb3\x67\x7f\x60\x92\x19\xef\xc9\xe0\xfc\x70\x34\xce\xce\xa4\xf4\x56\x5d\x9e\xb2\x7e\xed\x22\xad\x55\xe9\xe1\x68\x7c\x67\xbc\xe9\xc3\x55\x12\x96\xc9\x9f\x5b\xc4\x1b\x78\xa8\xd0\xe7\xe8\xf8\x03\x34\x9e\x90\x7b\xdc\x98\xde\xac\xd9\x2d\xba\xb1\xc0\x37\xa8\x06\x4a\x79\xd9\xb4\x8a\xbd\x6e\x80\xb5\xc2\xcc\x5c\x4b\xa6\xdf\xa8\xf8\x84\x33\x9e\x2a\x3c\xd7\xf9\x41\xfd\xb2\xc8\xfb\x1f\xb5\x34\x3e\x69\x55\xcb\x47\x2f\xeb\x2a\x2b\x9f\xae\x09\x33\x7c\x03\x93\xb9\x14\x72\xb9\x6e\x10\xa3\xb7\x30\x94\x8a\x32\xa3\x99\x82\x9b\x58\x3f\xea\xca\xb5\x9f\xe1\x2d\x58\xd5\x5d\xda\xba\x37\x64\xdb\x11\xb5\x03\x36\x30\xb9\x6e\x38\xa5\x99\x3f\x41\x10\xeb\x7e\xf7\x2b\x9b\x9a\xbb\x2e\xf0\x18\x4b\xcb\xf2\x82\x2a\x1f\x5f\xce\xce\xea\x45\xf8\x28\x33\x7b\x40\x8e\xb1\x61\x9a\x2f\x60\x66\x25\xff\xf4\x68\x31\x7c\x7a\x7e\x34\x5d\x5f\x5d\xdd\x2f\x93\x1f\x75\x1f\x96\x1f\x74\xbf\x56\x7e\xf4\xb9\xc5\xf0\x99\xf8\xff\x63\xf9\xd1\xd7\xca\x6e\xba\xff\x0b\xb2\x9b\x3f\x40\xe3\x1f\xca\x6e\xba\x8f\xcc\x6e\xba\x95\x78\xd9\xfd\xd7\xc9\x6e\xba\x5f\x31\xbb\xe9\x7e\x8e\xec\xe6\x4e\x6d\x7d\xe1\xec\xa6\xfb\x09\x82\x58\xf7\x9a\xff\x22\xd9\xcd\x97\xb6\xb3\x7a\x11\x7e\xce\xec\xe6\xcb\x9a\x59\xc9\x3f\x7d\x82\x18\x1a\x00\xfe\x72\x52\x85\xc4\xc9\x6e\xbf\x5f\x78\x3b\x50\x25\xa6\x70\xa9\xc9\x45\xe0\xe2\x1d\x6b\x17\x1b\xc8\xa5\x8f\x5a\x85\x33\x53\x72\x2d\x83\xfc\x0e\x50\x7b\xd3\x35\xaa\x16\xc0\x8f\x08\x11\x67\xdb\xca\x20\x0b\x49\x1c\x03\x55\xee\x2a\x4d\xbc\x2a\x1c\xc0\x13\x55\xb8\x54\xa4\x43\x4d\x88\x12\x7a\xa5\x5b\x83\xd2\x1d\xe4\x02\x67\xf1\xca\xe0\xbb\x26\x4c\x81\x7d\x75\xe2\x6e\x11\xbf\xe5\x72\x43\x45\xce\xdf\x2f\xde\x20\x8b\xd1\x92\xcc\xb5\xc0\xb3\x65\xf0\x9a\xf5\xfb\xaf\x29\x8b\x46\xf9\x69\x95\x5d\x34\xd9\x11\x56\xa9\x6d\xd6\xaf\x9c\x24\xd8\xae\xed\xe7\x7b\xbe\x8e\x5e\xba\xe2\x6c\x25\xbe\xf9\x51\x90\x5b\x96\xd9\x9b\xab\x62\xc9\xd6\x99\x19\xfa\x9c\x64\x60\x2f\x63\xf6\x8b\x67\xb4\x7a\x79\xc4\xa8\xf0\x94\x9d\xa3\x58\xba\x65\x52\x81\xb8\xa2\x21\x8e\x58\x84\x37\x7d\xe8\x64\xcd\x0f\x48\xc1\x6d\xda\x40\x65\x76\x0f\x94\x4a\x08\x17\x5c\x22\xcb\x74\xca\xd3\xec\x12\xb6\x4b\x20\xa6\x38\xa7\x4c\x02\x51\x50\x38\x9c\xcd\x10\x3a\x57\x5a\xe0\x66\xdb\x81\x3d\xdb\xce\x80\xcc\x53\xb3\x89\x37\xf6\x82\xf0\x9d\x20\x27\x18\xa6\x82\xaa\x95\x81\x7b\x9c\x8f\x74\xa6\x16\x68\x53\x73\x7d\xfe\x6d\x8a\x95\x43\xf1\x35\x09\xc0\x7b\x89\xe2\xc0\xbd\x96\xcb\x2d\xe6\x15\x91\xf8\x7c\x2f\x6f\xb3\xad\x93\x74\xda\x87\xff\x2c\x34\x02\x7c\xfb\x4d\x7b\x4a\x59\x7b\x4a\xe4\xa2\xd4\x4e\x12\x15\xcc\x51\x81\xbb\xd7\x00\x29\xfb\x9d\x26\x10\x04\x2b\x94\x25\xb8\xd2\x8f\xe5\x65\x44\x05\x04\x09\xb4\x79\xa2\xda\xfe\x31\xc5\x77\xdf\x95\x80\x00\xc2\x54\xc4\x10\x1c\x4b\x68\xab\x65\x02\xfe\x2a\x84\x7b\x96\xd1\xb2\xaf\x34\x48\x42\xfd\x35\xb3\xf2\xfb\x8e\xdb\xb6\x79\x9c\x41\x51\xb6\xb7\x3e\x96\x5f\x94\xdc\xb6\x49\x42\x43\x15\xb7\x34\xa9\x3f\x18\xec\xc5\x96\x0a\x1d\x96\xa7\x75\xb0\x20\xba\x8f\xfc\xc5\x92\x47\xf0\xd7\x9b\x12\x98\xc3\xb0\x59\x36\x21\x51\xf0\xf7\xbf\x1f\x9e\xbe\x86\x1f\x7e\x28\x8f\xdc\xfa\xe8\x34\x7a\xdb\x4a\x70\x59\x1a\x64\xba\xfc\xbb\x9d\xdb\x52\xd7\xe1\xe9\xeb\x46\xa3\x4a\xd6\xf3\x4e\xe7\xc1\xd8\xc3\x05\xbf\x66\x10\x9c\x41\x3a\x4d\x99\x4a\x4b\xe3\x4a\x80\x72\x25\x15\x2e\x43\x15\x43\x44\x70\xc9\x59\x20\xd0\x3c\x97\xf9\xee\xbb\x42\x17\x9a\xe3\x3b\x70\x72\xb8\x83\x77\x54\x61\xbb\xf0\x12\xa6\x04\x79\x30\x38\x1f\x0c\x4f\x8f\x3f\x8c\xcf\x4e\x2f\x46\x07\x87\x67\x2f\xc9\xb5\xac\x05\x98\x9c\x0f\x86\x47\x2f\xdd\xb1\x53\x56\x04\xbe\xad\x85\xbd\x38\x3c\x9b\x8c\x4e\x4f\x5e\x56\xec\xa5\x16\xfa\xd5\xfb\xe1\xd1\xe1\xf9\xcb\xad\x8f\xa5\x28\x54\x0f\x3b\x18\x8f\x3e\x1c\x1d\xfe\xa4\xe9\x30\xaf\xa5\xca\x50\x83\x1f\x27\x1f\x2e\xc6\xc3\x0f\xa3\x83\x97\x5b\x1f\x4b\xde\xab\x65\xb2\xa8\x7a\x9c\x47\x87\x3f\x7d\x38\x19\xbc\x3b\x7c\x99\x6b\xae\x82\x75\xf2\xfe\xd5\xc9\xe1\xf9\x64\x1d\xed\xed\xd3\x72\x43\x77\xe3\xc8\x0f\xe3\xb3\xd1\xc5\xe0\x5c\x4f\x52\x2e\x8c\x19\x14\xa5\x64\xb2\x06\xc7\xe1\xf0\xfd\xd9\xe8\xfc\xa7\x0f\x6f\xce\x4e\xdf\x8f\x5f\x6e\x7d\xac\x73\x7a\x2d\xf3\xff\x75\x26\xbd\x48\xf4\xe6\xfa\xe5\xd6\xc7\x8b\xf1\xb0\x95\xa5\x4c\x55\xc8\xc1\x70\x78\xfa\xfe\xe4\xdc\x0a\xd0\x68\x7a\x10\x86\x3c\x65\xaa\x0e\xed\xd9\xe1\x1b\xab\xe4\xe2\x49\x64\x1d\xca\xc3\xc9\xc4\x08\xd9\x63\xcd\xee\x41\xd5\x72\x7a\x76\x78\x5e\x18\x64\x47\xac\xdd\x9f\xba\x67\x61\xe6\x4b\xc4\x26\x15\xf9\x0a\xf9\x16\x4e\x13\x64\x90\x26\x90\x70\xa1\xa0\xd7\xcb\x9e\xe6\xd5\x3d\x2a\xca\xb3\x8b\x92\xa0\x37\xa4\x84\x95\x00\x54\xcd\x35\x4c\x5f\xf9\x75\x93\x5d\xc3\x85\xc7\x13\x57\x94\x78\xe2\x9e\xc2\x20\xa1\x79\xc3\x7e\x67\xbf\x63\x32\xa1\x37\x22\x09\xf3\xe6\x6e\xa7\xd3\xe9\xdc\xb9\x57\x28\x91\xe6\x1e\x7c\xe4\x71\x71\x94\x8c\x05\x57\x3c\xe4\x71\x1f\x54\x98\x27\x37\xaf\x05\x5f\x8e\xed\xa5\x81\x5e\x2f\x0f\xc4\xe7\xbc\xa6\x51\x9b\xd4\x28\xf1\xc7\x44\x77\x3d\xd8\x7e\xc0\x6c\x9a\xcf\x9a\xf9\xca\xcd\x7f\xea\x8c\x46\x84\x35\x53\xae\xb5\xdf\x3f\xa7\x31\x31\x7f\x25\x2f\x49\x24\x44\x2b\x46\x96\x3c\x9a\x82\x72\x79\xfb\x81\x69\x18\x24\x49\xfd\x96\xd0\x76\x1f\xbc\xea\xf7\xef\xda\x12\x9a\x3e\x9b\xa0\x94\x92\x90\x7f\xe3\x94\x15\x53\x90\x00\x9a\x41\x13\x4a\x0d\xc1\x1d\x47\x79\xd9\x20\x4d\xba\xbf\x41\x30\x50\x4a\xd0\x69\xaa\xf0\x00\x67\x94\xd1\xc2\x75\x0e\x0b\x9c\xf5\x5b\x7a\x58\x19\x5d\xd6\x6b\x79\x6c\x4e\x9a\x79\x86\x35\x09\x17\xb8\x24\x77\xe0\x6a\x6a\x64\xcd\x02\xb6\x23\x5c\x39\x3c\x6f\x07\x93\xb7\xbe\xc7\x5c\x61\xd0\xe1\x05\xa3\xf3\x85\xe0\xe9\x7c\x91\xa4\x85\x3c\xf1\x0c\x49\xe4\x5f\xcc\xbc\x67\x54\xc9\x3e\x34\x9f\xe5\x58\x7f\x14\x54\x61\x4d\xbf\xa9\x1f\x67\xca\x9c\xa6\x34\x8e\x36\x69\xf3\x95\xee\xfc\x73\xf4\xf9\x59\xd4\x69\x88\xff\x54\x85\x36\x69\xd4\xfc\xd3\x14\x5a\xc6\xf5\xf5\xd4\x29\x30\x46\x22\x71\xe3\xfa\x3c\xb3\xfd\xff\x83\x75\xea\x39\xf8\x3f\xad\xc2\x50\x6f\x97\x0a\x4f\xb1\xf3\x4f\xb7\x98\x47\x36\x69\x92\xc4\x2b\xa0\x4a\x16\x3e\x84\xd2\xc8\xb6\xb1\xfe\xeb\x2e\xc4\xbf\xd6\xb4\x97\xa7\x4d\x5d\xc2\xf7\x69\xdc\x24\x54\x29\x89\x6b\xee\xbc\xdb\x7a\x44\x5e\xd7\xcf\xef\xf3\x1d\xed\x4b\xa3\xbe\x8a\x01\xad\x7f\x5e\xc5\x40\x19\xa0\xfc\xe6\xe2\xfa\xf5\xc3\x4d\x56\xe6\x58\x7d\x7f\x76\xdc\x77\x37\xc3\xfc\xf6\x6e\xeb\x63\xf9\xfd\x5a\xed\x8b\xa2\xd2\x43\x9c\xdb\x76\xfe\xea\x29\xbf\x4e\x1e\xf8\xcb\xfd\xbe\x33\x53\x69\xe5\x33\x2a\xe6\xa2\xd4\xe8\xa0\x92\x80\x3c\xe8\xc2\x4b\x6d\x99\xa4\xa6\x3e\x52\x7a\xe0\xee\x2e\x8b\xad\x7f\x74\x05\xc0\xdd\x63\x79\xc8\xd9\xb5\x56\xe0\x81\x29\x4c\x15\xf3\x30\xce\xe2\x55\xfe\x36\xac\xf4\x88\x95\x2a\x89\xf1\x2c\xaf\x9a\x4c\xde\x66\xcf\x67\xb3\xdb\x79\x85\xb2\x54\x2b\x2b\x70\xdc\xb6\x77\x7b\xb9\x51\x0f\x12\x7a\x3c\x2d\x8c\xbc\x27\x93\xb9\xbb\x2a\x51\x7d\x1c\xea\xa0\x6a\x3f\x5e\x53\x79\xdb\x68\x81\x6b\xbf\x46\xb3\xfe\x5a\x2b\x07\xcd\xbf\x50\xb3\x26\xf5\x62\xfd\xf4\x4e\x67\x56\xf3\x94\xc6\x9d\x07\x6c\xfa\x1c\x0f\xc0\x31\x27\xd1\x2b\x12\x6b\xab\xa8\xa8\xb7\x7c\xa7\xe5\x34\x55\x49\xaa\xdc\xad\x5b\xf7\xc3\xaa\xf4\x62\x3c\xcc\x3f\x4b\xe2\xad\xd6\xcc\x50\x7e\x3c\x7b\xe0\xef\x90\x32\xbc\x8e\x57\x81\x75\x08\xf6\xa5\xbc\xde\x49\x99\x11\xc5\x1b\x52\xd6\xe4\xb3\xab\xf6\xb2\x06\x67\xf3\x98\xda\xab\xa9\xa5\x12\x99\xf3\xe3\x0e\x97\xe3\xb5\x1a\x2f\x02\x68\x3e\x2d\x44\x8b\x2c\x56\xd4\x57\xea\xa0\xae\xbb\x9b\xbf\x04\x90\xce\x30\xef\xa4\xb2\x7c\x56\xf5\x07\xc9\xac\x5f\x7e\x50\xdb\xdf\x6d\x40\xb9\xbc\xeb\x8b\x9e\x75\x8a\x1a\x83\x2b\x1e\x7a\x85\x95\x16\x6c\xe1\x7b\x62\xf7\x69\xd1\x9d\x34\x96\x96\xaf\x9b\xb7\x51\x47\xce\xc1\xc9\xa4\x86\x1e\x77\xec\x72\x70\x32\x81\xd7\xff\x38\x38\xf9\x6c\x44\x1d\x30\xe9\x6f\x5b\x4f\x26\x6f\xc7\x82\xdf\xac\x86\x7c\xb9\x24\xcc\xd5\xb3\xef\x89\x27\x25\x9a\xcf\x52\x66\x5e\xd2\xc7\xf1\x0a\x02\xe3\x09\x43\x8b\xca\x7c\xec\x47\xa3\xf6\x8f\xf5\xdd\x37\xf0\x68\xb1\xb0\x0e\xa0\x6c\x58\xaf\x30\xfa\xd4\x3c\xdc\xb7\x4e\x35\xdf\xc7\xee\x14\x90\xeb\xb9\xca\xa8\xd7\xde\xa8\x3f\x29\x8b\x43\xbb\xd8\x1f\x82\x46\xe6\x7d\x4d\x9d\xa0\x99\x10\xb5\x68\x2b\xbe\x56\x7f\x6b\xfe\x3f\x07\x27\xb3\x7a\x6b\x40\x61\xcb\x8d\xf2\x2d\x03\x08\x8e\x35\x59\x7d\x23\x00\x4d\x75\x5f\xff\xf4\xdd\x1c\x8a\xa2\x7d\xd9\x94\x72\xa1\xb1\xfc\xd2\xdc\xfa\xf8\x8d\xc3\x74\xfb\x4b\xd3\x55\xf5\xfe\xff\x9a\xf3\x77\xd6\x73\x0b\x2c\x84\xbf\x2c\xe0\x2f\x89\x8f\x00\x19\xb8\x4f\x14\x5a\xce\x47\xb5\xde\x19\x29\xe4\x61\x43\xeb\xf7\x0d\x2a\x2d\x95\xa1\xb9\xd0\xf2\x67\x29\x79\x38\x2e\x2a\x39\xfb\x2a\xcf\xfa\xeb\xbe\xe2\x53\x3e\x47\xfc\x8c\xba\xcf\xec\x59\xcd\x9a\x4f\x73\xb8\xec\xc7\xa7\x61\x57\xd4\xbe\x4a\x0b\x55\xfc\xb4\xf4\x81\xc5\x6c\xce\x98\x32\xfb\x65\x0f\x7f\xfb\x7e\x98\x65\x63\x4d\x33\xd2\x4c\xdc\xb4\x93\xb9\xfc\x2a\xb4\x2f\x24\xb3\x57\xd9\x2b\xf3\xf8\x94\x3d\xb5\xdf\xce\x23\x0c\x44\xea\x69\x6c\xe2\x8d\x31\xb8\xa3\xf7\xaf\x0e\x87\xa7\x27\xaf\x47\x6f\x5e\x6e\xed\x24\xd7\xd1\x93\x76\x11\xb9\xe2\x80\x4c\xa6\x02\x3d\xb1\x90\x4a\x93\xf4\xd1\x35\xce\x0d\x19\x9e\xd2\xc1\x94\xa7\x2a\x1b\x11\x64\x45\xf5\x3c\x6f\x6a\x51\x6e\x9e\x0f\xb4\x53\x89\x22\x98\x9b\xf7\x01\x89\x40\x81\xbf\xc9\x36\xfc\x09\xf6\x9c\x95\x32\x6a\xec\xf9\x6b\x1b\x6c\xff\xbf\x0a\x22\x86\x56\xe1\x47\xa3\x1c\x8c\xed\xdb\x5c\x8c\xf2\x4c\xeb\x72\x3f\xff\xd2\x61\x03\xc0\x62\xf6\x69\xe0\xe8\x13\x4c\x3e\x1f\xeb\xdd\xb0\x73\x32\xc5\xaf\x50\xd4\xfa\xdb\x0d\x6c\xe6\x08\x1b\x19\x81\x19\xeb\x8f\xa7\xcf\xdf\x06\x29\x7c\x5b\xa5\x44\xe7\x63\x48\xcb\xc8\x30\xef\x94\x79\x84\xae\xe4\xf7\x67\x4a\xaf\x1c\xaf\xb2\x49\x1e\x46\x66\x0d\x4d\x9a\x54\x9d\x3e\xe8\x2e\xf9\x78\x02\x87\xb9\xff\x32\x5f\x4b\x5d\x72\x81\xee\xc3\x26\x26\xa6\x50\x59\xfe\x44\xd7\x7d\x14\x66\xa4\x14\x3e\xa0\xb6\xf1\xcb\x6c\x83\x84\x9a\x37\xef\xe6\x5d\xad\xff\x80\xda\xf1\xa8\x92\x17\x5a\x04\xff\x1d\x00\x00\xff\xff\x01\xb3\x4d\x70\xce\x56\x00\x00")

func cmdProviderAwsTemplatesFormationYamlBytes() ([]byte, error) {
	return bindataRead(
		_cmdProviderAwsTemplatesFormationYaml,
		"cmd/provider/aws/templates/formation.yaml",
	)
}

func cmdProviderAwsTemplatesFormationYaml() (*asset, error) {
	bytes, err := cmdProviderAwsTemplatesFormationYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "cmd/provider/aws/templates/formation.yaml", size: 22222, mode: os.FileMode(420), modTime: time.Unix(1524790131, 0)}
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
	"cmd/provider/aws/templates/formation.yaml": cmdProviderAwsTemplatesFormationYaml,
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
	"cmd": &bintree{nil, map[string]*bintree{
		"provider": &bintree{nil, map[string]*bintree{
			"aws": &bintree{nil, map[string]*bintree{
				"templates": &bintree{nil, map[string]*bintree{
					"formation.yaml": &bintree{cmdProviderAwsTemplatesFormationYaml, map[string]*bintree{}},
				}},
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

