// Code generated by protoc-gen-go.
// source: types.proto
// DO NOT EDIT!

/*
Package models is a generated protocol buffer package.

It is generated from these files:
	types.proto

It has these top-level messages:
	App
	Build
	Release
	Resource
	ResourceVar
	EnvConfig
*/
package models

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "google.golang.org/genproto/googleapis/api/annotations"
import _ "github.com/golang/protobuf/ptypes/timestamp"
import _ "github.com/gogo/protobuf/gogoproto"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Status int32

const (
	Status_CREATED Status = 0
)

var Status_name = map[int32]string{
	0: "CREATED",
}
var Status_value = map[string]int32{
	"CREATED": 0,
}

func (x Status) String() string {
	return proto.EnumName(Status_name, int32(x))
}
func (Status) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type App struct {
	Name      string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	Status    Status `protobuf:"varint,2,opt,name=status,enum=models.Status" json:"status,omitempty"`
	ReleaseId string `protobuf:"bytes,3,opt,name=release_id,json=releaseId" json:"release_id,omitempty"`
	Endpoint  string `protobuf:"bytes,4,opt,name=endpoint" json:"endpoint,omitempty"`
}

func (m *App) Reset()                    { *m = App{} }
func (m *App) String() string            { return proto.CompactTextString(m) }
func (*App) ProtoMessage()               {}
func (*App) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *App) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *App) GetStatus() Status {
	if m != nil {
		return m.Status
	}
	return Status_CREATED
}

func (m *App) GetReleaseId() string {
	if m != nil {
		return m.ReleaseId
	}
	return ""
}

func (m *App) GetEndpoint() string {
	if m != nil {
		return m.Endpoint
	}
	return ""
}

type Build struct {
	Id        string `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	App       string `protobuf:"bytes,2,opt,name=app" json:"app,omitempty"`
	RemoteId  string `protobuf:"bytes,3,opt,name=remote_id,json=remoteId" json:"remote_id,omitempty"`
	Status    Status `protobuf:"varint,4,opt,name=status,enum=models.Status" json:"status,omitempty"`
	CreatedAt int32  `protobuf:"varint,5,opt,name=created_at,json=createdAt" json:"created_at,omitempty"`
}

func (m *Build) Reset()                    { *m = Build{} }
func (m *Build) String() string            { return proto.CompactTextString(m) }
func (*Build) ProtoMessage()               {}
func (*Build) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *Build) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *Build) GetApp() string {
	if m != nil {
		return m.App
	}
	return ""
}

func (m *Build) GetRemoteId() string {
	if m != nil {
		return m.RemoteId
	}
	return ""
}

func (m *Build) GetStatus() Status {
	if m != nil {
		return m.Status
	}
	return Status_CREATED
}

func (m *Build) GetCreatedAt() int32 {
	if m != nil {
		return m.CreatedAt
	}
	return 0
}

type Release struct {
	Id        string `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	App       string `protobuf:"bytes,2,opt,name=app" json:"app,omitempty"`
	BuildId   string `protobuf:"bytes,3,opt,name=build_id,json=buildId" json:"build_id,omitempty"`
	Status    Status `protobuf:"varint,4,opt,name=status,enum=models.Status" json:"status,omitempty"`
	CreatedAt int32  `protobuf:"varint,5,opt,name=created_at,json=createdAt" json:"created_at,omitempty"`
}

func (m *Release) Reset()                    { *m = Release{} }
func (m *Release) String() string            { return proto.CompactTextString(m) }
func (*Release) ProtoMessage()               {}
func (*Release) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *Release) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *Release) GetApp() string {
	if m != nil {
		return m.App
	}
	return ""
}

func (m *Release) GetBuildId() string {
	if m != nil {
		return m.BuildId
	}
	return ""
}

func (m *Release) GetStatus() Status {
	if m != nil {
		return m.Status
	}
	return Status_CREATED
}

func (m *Release) GetCreatedAt() int32 {
	if m != nil {
		return m.CreatedAt
	}
	return 0
}

type Resource struct {
	Name       string            `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	Status     Status            `protobuf:"varint,2,opt,name=status,enum=models.Status" json:"status,omitempty"`
	Kind       string            `protobuf:"bytes,3,opt,name=kind" json:"kind,omitempty"`
	URL        string            `protobuf:"bytes,4,opt,name=URL" json:"URL,omitempty"`
	Apps       []string          `protobuf:"bytes,5,rep,name=apps" json:"apps,omitempty"`
	Exports    []*ResourceVar    `protobuf:"bytes,6,rep,name=exports" json:"exports,omitempty"`
	Parameters map[string]string `protobuf:"bytes,7,rep,name=parameters" json:"parameters,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *Resource) Reset()                    { *m = Resource{} }
func (m *Resource) String() string            { return proto.CompactTextString(m) }
func (*Resource) ProtoMessage()               {}
func (*Resource) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *Resource) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Resource) GetStatus() Status {
	if m != nil {
		return m.Status
	}
	return Status_CREATED
}

func (m *Resource) GetKind() string {
	if m != nil {
		return m.Kind
	}
	return ""
}

func (m *Resource) GetURL() string {
	if m != nil {
		return m.URL
	}
	return ""
}

func (m *Resource) GetApps() []string {
	if m != nil {
		return m.Apps
	}
	return nil
}

func (m *Resource) GetExports() []*ResourceVar {
	if m != nil {
		return m.Exports
	}
	return nil
}

func (m *Resource) GetParameters() map[string]string {
	if m != nil {
		return m.Parameters
	}
	return nil
}

type ResourceVar struct {
	Key   string `protobuf:"bytes,1,opt,name=key" json:"key,omitempty"`
	Value string `protobuf:"bytes,2,opt,name=value" json:"value,omitempty"`
}

func (m *ResourceVar) Reset()                    { *m = ResourceVar{} }
func (m *ResourceVar) String() string            { return proto.CompactTextString(m) }
func (*ResourceVar) ProtoMessage()               {}
func (*ResourceVar) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *ResourceVar) GetKey() string {
	if m != nil {
		return m.Key
	}
	return ""
}

func (m *ResourceVar) GetValue() string {
	if m != nil {
		return m.Value
	}
	return ""
}

type EnvConfig struct {
	Data map[string]string `protobuf:"bytes,1,rep,name=data" json:"data,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *EnvConfig) Reset()                    { *m = EnvConfig{} }
func (m *EnvConfig) String() string            { return proto.CompactTextString(m) }
func (*EnvConfig) ProtoMessage()               {}
func (*EnvConfig) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *EnvConfig) GetData() map[string]string {
	if m != nil {
		return m.Data
	}
	return nil
}

func init() {
	proto.RegisterType((*App)(nil), "models.App")
	proto.RegisterType((*Build)(nil), "models.Build")
	proto.RegisterType((*Release)(nil), "models.Release")
	proto.RegisterType((*Resource)(nil), "models.Resource")
	proto.RegisterType((*ResourceVar)(nil), "models.ResourceVar")
	proto.RegisterType((*EnvConfig)(nil), "models.EnvConfig")
	proto.RegisterEnum("models.Status", Status_name, Status_value)
}

func init() { proto.RegisterFile("types.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 634 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xc4, 0x54, 0xdd, 0x6a, 0xd4, 0x40,
	0x14, 0x36, 0x9b, 0xfd, 0xcb, 0x59, 0x68, 0xcb, 0x74, 0x5b, 0xe3, 0x2a, 0xcd, 0x32, 0x20, 0x2c,
	0x82, 0xbb, 0xd0, 0xfa, 0x5b, 0x50, 0xec, 0xb6, 0x2b, 0x14, 0x7a, 0x21, 0xa3, 0xf5, 0xc2, 0x9b,
	0x32, 0xdb, 0x4c, 0x63, 0xe8, 0x26, 0x33, 0x24, 0x93, 0x62, 0xef, 0xf5, 0x9d, 0x7c, 0x04, 0x2f,
	0x7d, 0x82, 0x50, 0xfa, 0x08, 0x79, 0x02, 0x99, 0x49, 0xb2, 0x1b, 0x57, 0x2f, 0x54, 0x44, 0xef,
	0xce, 0xc9, 0xf7, 0x7d, 0x67, 0xce, 0xf9, 0xe6, 0x64, 0xa0, 0x23, 0x2f, 0x05, 0x8b, 0x87, 0x22,
	0xe2, 0x92, 0xa3, 0x66, 0xc0, 0x5d, 0x36, 0x8b, 0x7b, 0x77, 0x3c, 0xce, 0xbd, 0x19, 0x1b, 0x51,
	0xe1, 0x8f, 0x68, 0x18, 0x72, 0x49, 0xa5, 0xcf, 0xc3, 0x82, 0xd5, 0x73, 0x0a, 0x54, 0x67, 0xd3,
	0xe4, 0x6c, 0x24, 0xfd, 0x80, 0xc5, 0x92, 0x06, 0xa2, 0x20, 0xdc, 0xf7, 0x7c, 0xf9, 0x3e, 0x99,
	0x0e, 0x4f, 0x79, 0x30, 0xf2, 0xb8, 0xc7, 0x17, 0x4c, 0x95, 0xe9, 0x44, 0x47, 0x39, 0x1d, 0x5f,
	0x19, 0x60, 0xee, 0x09, 0x81, 0x06, 0x50, 0x0f, 0x69, 0xc0, 0x6c, 0xa3, 0x6f, 0x0c, 0xac, 0x71,
	0x37, 0x4b, 0x9d, 0x35, 0x97, 0x4a, 0x1a, 0x4b, 0x1e, 0xb1, 0x5d, 0xac, 0x20, 0x4c, 0x34, 0x03,
	0x3d, 0x87, 0x66, 0x2c, 0xa9, 0x4c, 0x62, 0xbb, 0xd6, 0x37, 0x06, 0x2b, 0xdb, 0x2b, 0xc3, 0xbc,
	0xf1, 0xe1, 0x6b, 0xfd, 0x75, 0xbc, 0x99, 0xa5, 0x0e, 0xaa, 0x68, 0x73, 0x32, 0x26, 0x85, 0x0a,
	0x3d, 0x05, 0x88, 0xd8, 0x8c, 0xd1, 0x98, 0x9d, 0xf8, 0xae, 0x6d, 0xea, 0xf3, 0x7a, 0x59, 0xea,
	0x6c, 0x56, 0x34, 0x0b, 0x02, 0x26, 0x56, 0x91, 0x1c, 0xba, 0xe8, 0x01, 0xb4, 0x59, 0xe8, 0x0a,
	0xee, 0x87, 0xd2, 0xae, 0x6b, 0xa1, 0x9d, 0xa5, 0x4e, 0xb7, 0x22, 0x2c, 0x61, 0x4c, 0xe6, 0x4c,
	0xfc, 0xa9, 0x06, 0x8d, 0x71, 0xe2, 0xcf, 0x5c, 0x84, 0xa1, 0xe6, 0xbb, 0xc5, 0x88, 0x28, 0x4b,
	0x9d, 0x95, 0x8a, 0x52, 0x1d, 0x55, 0xf3, 0x5d, 0x74, 0x17, 0x4c, 0x2a, 0x84, 0x9e, 0xcd, 0x1a,
	0xaf, 0x67, 0xa9, 0xb3, 0x5a, 0x21, 0x51, 0x21, 0x30, 0x51, 0x38, 0x7a, 0x04, 0x56, 0xc4, 0x02,
	0x2e, 0x2b, 0x43, 0xdc, 0xca, 0x52, 0x67, 0xe3, 0xbb, 0x21, 0x0a, 0x1c, 0x93, 0x76, 0x1e, 0x1f,
	0xba, 0x15, 0xf7, 0xea, 0x7f, 0xea, 0xde, 0x69, 0xc4, 0xa8, 0x64, 0xee, 0x09, 0x95, 0x76, 0xa3,
	0x6f, 0x0c, 0x1a, 0x3f, 0xb8, 0xb7, 0x20, 0x60, 0x62, 0x15, 0xc9, 0x9e, 0xc4, 0x1f, 0x6b, 0xd0,
	0x22, 0xb9, 0x97, 0x7f, 0xd3, 0x89, 0x1d, 0x68, 0x4f, 0x95, 0xbb, 0x0b, 0x23, 0x96, 0x2f, 0xa5,
	0x84, 0x31, 0x69, 0xe9, 0xf0, 0xff, 0xda, 0xf0, 0xd9, 0x84, 0x36, 0x61, 0x31, 0x4f, 0xa2, 0x53,
	0xf6, 0x0f, 0xd7, 0x7e, 0x00, 0xf5, 0x73, 0x3f, 0x2c, 0x2d, 0x5a, 0x3e, 0x49, 0x41, 0x98, 0x68,
	0x86, 0xf2, 0xfd, 0x98, 0x1c, 0x15, 0x0b, 0xbe, 0xec, 0xfb, 0x31, 0x39, 0xc2, 0x44, 0xe1, 0xaa,
	0x20, 0x15, 0x22, 0xb6, 0x1b, 0x7d, 0xf3, 0x27, 0x05, 0x15, 0x84, 0x89, 0x66, 0xa0, 0x97, 0xd0,
	0x62, 0x1f, 0x04, 0x8f, 0x64, 0x6c, 0x37, 0xfb, 0xe6, 0xa0, 0xb3, 0xbd, 0x5e, 0xf6, 0x5e, 0xfa,
	0xf0, 0x96, 0x46, 0xe3, 0x9b, 0x59, 0xea, 0xac, 0x57, 0x7f, 0xa5, 0x5c, 0x82, 0x49, 0x29, 0x46,
	0x2f, 0x00, 0x04, 0x8d, 0x68, 0xc0, 0x24, 0x8b, 0x62, 0xbb, 0xa5, 0x4b, 0xf5, 0x97, 0x4b, 0x0d,
	0x5f, 0xcd, 0x29, 0x93, 0x50, 0x46, 0x97, 0xa4, 0xa2, 0xe9, 0x3d, 0x83, 0xd5, 0x25, 0x18, 0xad,
	0x81, 0x79, 0xce, 0x2e, 0xf3, 0x0b, 0x20, 0x2a, 0x44, 0x5d, 0x68, 0x5c, 0xd0, 0x59, 0xc2, 0xf2,
	0xcd, 0x23, 0x79, 0xb2, 0x5b, 0x7b, 0x62, 0xe0, 0x87, 0xd0, 0xa9, 0x74, 0xfc, 0xab, 0x52, 0x9c,
	0x80, 0x35, 0x09, 0x2f, 0xf6, 0x79, 0x78, 0xe6, 0x7b, 0x68, 0x04, 0x75, 0x35, 0xa4, 0x6d, 0xe8,
	0xf6, 0x6f, 0x97, 0xed, 0xcf, 0x09, 0xc3, 0x03, 0x2a, 0x69, 0xde, 0xb9, 0x26, 0xf6, 0x1e, 0x83,
	0x35, 0xff, 0xf4, 0x3b, 0xdd, 0xde, 0xdb, 0x80, 0x66, 0xbe, 0x1b, 0xa8, 0x03, 0xad, 0x7d, 0x32,
	0xd9, 0x7b, 0x33, 0x39, 0x58, 0xbb, 0x31, 0xee, 0x7e, 0xb9, 0xde, 0x32, 0xbe, 0x5e, 0x6f, 0x19,
	0x57, 0xd7, 0x5b, 0xc6, 0xbb, 0xe2, 0xd5, 0x9f, 0x36, 0xf5, 0x73, 0xbc, 0xf3, 0x2d, 0x00, 0x00,
	0xff, 0xff, 0x2b, 0x58, 0xd3, 0x78, 0x13, 0x06, 0x00, 0x00,
}
