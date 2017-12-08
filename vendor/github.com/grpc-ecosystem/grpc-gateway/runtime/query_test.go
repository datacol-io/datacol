package runtime_test

import (
	"net/url"
	"testing"

	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/grpc-ecosystem/grpc-gateway/utilities"
)

func TestPopulateParameters(t *testing.T) {
	timeT := time.Date(2016, time.December, 15, 12, 23, 32, 49, time.UTC)
	timeStr := timeT.Format(time.RFC3339Nano)
	timePb, err := ptypes.TimestampProto(timeT)
	if err != nil {
		t.Fatalf("Couldn't setup timestamp in Protobuf format: %v", err)
	}

	for _, spec := range []struct {
		values url.Values
		filter *utilities.DoubleArray
		want   proto.Message
	}{
		{
			values: url.Values{
				"float_value":     {"1.5"},
				"double_value":    {"2.5"},
				"int64_value":     {"-1"},
				"int32_value":     {"-2"},
				"uint64_value":    {"3"},
				"uint32_value":    {"4"},
				"bool_value":      {"true"},
				"string_value":    {"str"},
				"repeated_value":  {"a", "b", "c"},
				"enum_value":      {"1"},
				"repeated_enum":   {"1", "2", "0"},
				"timestamp_value": {timeStr},
			},
			filter: utilities.NewDoubleArray(nil),
			want: &proto3Message{
				FloatValue:     1.5,
				DoubleValue:    2.5,
				Int64Value:     -1,
				Int32Value:     -2,
				Uint64Value:    3,
				Uint32Value:    4,
				BoolValue:      true,
				StringValue:    "str",
				RepeatedValue:  []string{"a", "b", "c"},
				EnumValue:      EnumValue_Y,
				RepeatedEnum:   []EnumValue{EnumValue_Y, EnumValue_Z, EnumValue_X},
				TimestampValue: timePb,
			},
		},
		{
			values: url.Values{
				"enum_value":    {"EnumValue_Z"},
				"repeated_enum": {"EnumValue_X", "2", "0"},
			},
			filter: utilities.NewDoubleArray(nil),
			want: &proto3Message{
				EnumValue:    EnumValue_Z,
				RepeatedEnum: []EnumValue{EnumValue_X, EnumValue_Z, EnumValue_X},
			},
		},
		{
			values: url.Values{
				"float_value":    {"1.5"},
				"double_value":   {"2.5"},
				"int64_value":    {"-1"},
				"int32_value":    {"-2"},
				"uint64_value":   {"3"},
				"uint32_value":   {"4"},
				"bool_value":     {"true"},
				"string_value":   {"str"},
				"repeated_value": {"a", "b", "c"},
				"enum_value":     {"1"},
				"repeated_enum":  {"1", "2", "0"},
			},
			filter: utilities.NewDoubleArray(nil),
			want: &proto2Message{
				FloatValue:    proto.Float32(1.5),
				DoubleValue:   proto.Float64(2.5),
				Int64Value:    proto.Int64(-1),
				Int32Value:    proto.Int32(-2),
				Uint64Value:   proto.Uint64(3),
				Uint32Value:   proto.Uint32(4),
				BoolValue:     proto.Bool(true),
				StringValue:   proto.String("str"),
				RepeatedValue: []string{"a", "b", "c"},
				EnumValue:     EnumValue_Y,
				RepeatedEnum:  []EnumValue{EnumValue_Y, EnumValue_Z, EnumValue_X},
			},
		},
		{
			values: url.Values{
				"nested.nested.nested.repeated_value": {"a", "b", "c"},
				"nested.nested.nested.string_value":   {"s"},
				"nested.nested.string_value":          {"t"},
				"nested.string_value":                 {"u"},
				"nested_non_null.string_value":        {"v"},
			},
			filter: utilities.NewDoubleArray(nil),
			want: &proto3Message{
				Nested: &proto2Message{
					Nested: &proto3Message{
						Nested: &proto2Message{
							RepeatedValue: []string{"a", "b", "c"},
							StringValue:   proto.String("s"),
						},
						StringValue: "t",
					},
					StringValue: proto.String("u"),
				},
				NestedNonNull: proto2Message{
					StringValue: proto.String("v"),
				},
			},
		},
		{
			values: url.Values{
				"uint64_value": {"1", "2", "3", "4", "5"},
			},
			filter: utilities.NewDoubleArray(nil),
			want: &proto3Message{
				Uint64Value: 1,
			},
		},
	} {
		msg := proto.Clone(spec.want)
		msg.Reset()
		err := runtime.PopulateQueryParameters(msg, spec.values, spec.filter)
		if err != nil {
			t.Errorf("runtime.PopulateQueryParameters(msg, %v, %v) failed with %v; want success", spec.values, spec.filter, err)
			continue
		}
		if got, want := msg, spec.want; !proto.Equal(got, want) {
			t.Errorf("runtime.PopulateQueryParameters(msg, %v, %v = %v; want %v", spec.values, spec.filter, got, want)
		}
	}
}

func TestPopulateParametersWithFilters(t *testing.T) {
	for _, spec := range []struct {
		values url.Values
		filter *utilities.DoubleArray
		want   proto.Message
	}{
		{
			values: url.Values{
				"bool_value":     {"true"},
				"string_value":   {"str"},
				"repeated_value": {"a", "b", "c"},
			},
			filter: utilities.NewDoubleArray([][]string{
				{"bool_value"}, {"repeated_value"},
			}),
			want: &proto3Message{
				StringValue: "str",
			},
		},
		{
			values: url.Values{
				"nested.nested.bool_value":   {"true"},
				"nested.nested.string_value": {"str"},
				"nested.string_value":        {"str"},
				"string_value":               {"str"},
			},
			filter: utilities.NewDoubleArray([][]string{
				{"nested"},
			}),
			want: &proto3Message{
				StringValue: "str",
			},
		},
		{
			values: url.Values{
				"nested.nested.bool_value":   {"true"},
				"nested.nested.string_value": {"str"},
				"nested.string_value":        {"str"},
				"string_value":               {"str"},
			},
			filter: utilities.NewDoubleArray([][]string{
				{"nested", "nested"},
			}),
			want: &proto3Message{
				Nested: &proto2Message{
					StringValue: proto.String("str"),
				},
				StringValue: "str",
			},
		},
		{
			values: url.Values{
				"nested.nested.bool_value":   {"true"},
				"nested.nested.string_value": {"str"},
				"nested.string_value":        {"str"},
				"string_value":               {"str"},
			},
			filter: utilities.NewDoubleArray([][]string{
				{"nested", "nested", "string_value"},
			}),
			want: &proto3Message{
				Nested: &proto2Message{
					StringValue: proto.String("str"),
					Nested: &proto3Message{
						BoolValue: true,
					},
				},
				StringValue: "str",
			},
		},
	} {
		msg := proto.Clone(spec.want)
		msg.Reset()
		err := runtime.PopulateQueryParameters(msg, spec.values, spec.filter)
		if err != nil {
			t.Errorf("runtime.PoplateQueryParameters(msg, %v, %v) failed with %v; want success", spec.values, spec.filter, err)
			continue
		}
		if got, want := msg, spec.want; !proto.Equal(got, want) {
			t.Errorf("runtime.PopulateQueryParameters(msg, %v, %v = %v; want %v", spec.values, spec.filter, got, want)
		}
	}
}

func TestPopulateQueryParametersWithInvalidNestedParameters(t *testing.T) {
	for _, spec := range []struct {
		msg    proto.Message
		values url.Values
		filter *utilities.DoubleArray
	}{
		{
			msg: &proto3Message{},
			values: url.Values{
				"float_value.nested": {"test"},
			},
			filter: utilities.NewDoubleArray(nil),
		},
		{
			msg: &proto3Message{},
			values: url.Values{
				"double_value.nested": {"test"},
			},
			filter: utilities.NewDoubleArray(nil),
		},
		{
			msg: &proto3Message{},
			values: url.Values{
				"int64_value.nested": {"test"},
			},
			filter: utilities.NewDoubleArray(nil),
		},
		{
			msg: &proto3Message{},
			values: url.Values{
				"int32_value.nested": {"test"},
			},
			filter: utilities.NewDoubleArray(nil),
		},
		{
			msg: &proto3Message{},
			values: url.Values{
				"uint64_value.nested": {"test"},
			},
			filter: utilities.NewDoubleArray(nil),
		},
		{
			msg: &proto3Message{},
			values: url.Values{
				"uint32_value.nested": {"test"},
			},
			filter: utilities.NewDoubleArray(nil),
		},
		{
			msg: &proto3Message{},
			values: url.Values{
				"bool_value.nested": {"test"},
			},
			filter: utilities.NewDoubleArray(nil),
		},
		{
			msg: &proto3Message{},
			values: url.Values{
				"string_value.nested": {"test"},
			},
			filter: utilities.NewDoubleArray(nil),
		},
		{
			msg: &proto3Message{},
			values: url.Values{
				"repeated_value.nested": {"test"},
			},
			filter: utilities.NewDoubleArray(nil),
		},
		{
			msg: &proto3Message{},
			values: url.Values{
				"enum_value.nested": {"test"},
			},
			filter: utilities.NewDoubleArray(nil),
		},
		{
			msg: &proto3Message{},
			values: url.Values{
				"enum_value.nested": {"test"},
			},
			filter: utilities.NewDoubleArray(nil),
		},
		{
			msg: &proto3Message{},
			values: url.Values{
				"repeated_enum.nested": {"test"},
			},
			filter: utilities.NewDoubleArray(nil),
		},
	} {
		spec.msg.Reset()
		err := runtime.PopulateQueryParameters(spec.msg, spec.values, spec.filter)
		if err == nil {
			t.Errorf("runtime.PopulateQueryParameters(msg, %v, %v) did not fail; want error", spec.values, spec.filter)
		}
	}
}

type proto3Message struct {
	Nested         *proto2Message       `protobuf:"bytes,1,opt,name=nested" json:"nested,omitempty"`
	NestedNonNull  proto2Message        `protobuf:"bytes,11,opt,name=nested_non_null" json:"nested_non_null,omitempty"`
	FloatValue     float32              `protobuf:"fixed32,2,opt,name=float_value" json:"float_value,omitempty"`
	DoubleValue    float64              `protobuf:"fixed64,3,opt,name=double_value" json:"double_value,omitempty"`
	Int64Value     int64                `protobuf:"varint,4,opt,name=int64_value" json:"int64_value,omitempty"`
	Int32Value     int32                `protobuf:"varint,5,opt,name=int32_value" json:"int32_value,omitempty"`
	Uint64Value    uint64               `protobuf:"varint,6,opt,name=uint64_value" json:"uint64_value,omitempty"`
	Uint32Value    uint32               `protobuf:"varint,7,opt,name=uint32_value" json:"uint32_value,omitempty"`
	BoolValue      bool                 `protobuf:"varint,8,opt,name=bool_value" json:"bool_value,omitempty"`
	StringValue    string               `protobuf:"bytes,9,opt,name=string_value" json:"string_value,omitempty"`
	RepeatedValue  []string             `protobuf:"bytes,10,rep,name=repeated_value" json:"repeated_value,omitempty"`
	EnumValue      EnumValue            `protobuf:"varint,11,opt,name=enum_value,json=enumValue,enum=runtime_test_api.EnumValue" json:"enum_value,omitempty"`
	RepeatedEnum   []EnumValue          `protobuf:"varint,12,rep,packed,name=repeated_enum,json=repeated_enum,enum=runtime_test_api.EnumValue" json:"repeated_enum,omitempty"`
	TimestampValue *timestamp.Timestamp `protobuf:"bytes,11,opt,name=timestamp_value" json:"timestamp_value,omitempty"`
}

func (m *proto3Message) Reset()         { *m = proto3Message{} }
func (m *proto3Message) String() string { return proto.CompactTextString(m) }
func (*proto3Message) ProtoMessage()    {}

func (m *proto3Message) GetNested() *proto2Message {
	if m != nil {
		return m.Nested
	}
	return nil
}

type proto2Message struct {
	Nested           *proto3Message `protobuf:"bytes,1,opt,name=nested" json:"nested,omitempty"`
	FloatValue       *float32       `protobuf:"fixed32,2,opt,name=float_value" json:"float_value,omitempty"`
	DoubleValue      *float64       `protobuf:"fixed64,3,opt,name=double_value" json:"double_value,omitempty"`
	Int64Value       *int64         `protobuf:"varint,4,opt,name=int64_value" json:"int64_value,omitempty"`
	Int32Value       *int32         `protobuf:"varint,5,opt,name=int32_value" json:"int32_value,omitempty"`
	Uint64Value      *uint64        `protobuf:"varint,6,opt,name=uint64_value" json:"uint64_value,omitempty"`
	Uint32Value      *uint32        `protobuf:"varint,7,opt,name=uint32_value" json:"uint32_value,omitempty"`
	BoolValue        *bool          `protobuf:"varint,8,opt,name=bool_value" json:"bool_value,omitempty"`
	StringValue      *string        `protobuf:"bytes,9,opt,name=string_value" json:"string_value,omitempty"`
	RepeatedValue    []string       `protobuf:"bytes,10,rep,name=repeated_value" json:"repeated_value,omitempty"`
	EnumValue        EnumValue      `protobuf:"varint,11,opt,name=enum_value,json=enumValue,enum=runtime_test_api.EnumValue" json:"enum_value,omitempty"`
	RepeatedEnum     []EnumValue    `protobuf:"varint,12,rep,packed,name=repeated_enum,json=repeated_enum,enum=runtime_test_api.EnumValue" json:"repeated_enum,omitempty"`
	XXX_unrecognized []byte         `json:"-"`
}

func (m *proto2Message) Reset()         { *m = proto2Message{} }
func (m *proto2Message) String() string { return proto.CompactTextString(m) }
func (*proto2Message) ProtoMessage()    {}

func (m *proto2Message) GetNested() *proto3Message {
	if m != nil {
		return m.Nested
	}
	return nil
}

func (m *proto2Message) GetFloatValue() float32 {
	if m != nil && m.FloatValue != nil {
		return *m.FloatValue
	}
	return 0
}

func (m *proto2Message) GetDoubleValue() float64 {
	if m != nil && m.DoubleValue != nil {
		return *m.DoubleValue
	}
	return 0
}

func (m *proto2Message) GetInt64Value() int64 {
	if m != nil && m.Int64Value != nil {
		return *m.Int64Value
	}
	return 0
}

func (m *proto2Message) GetInt32Value() int32 {
	if m != nil && m.Int32Value != nil {
		return *m.Int32Value
	}
	return 0
}

func (m *proto2Message) GetUint64Value() uint64 {
	if m != nil && m.Uint64Value != nil {
		return *m.Uint64Value
	}
	return 0
}

func (m *proto2Message) GetUint32Value() uint32 {
	if m != nil && m.Uint32Value != nil {
		return *m.Uint32Value
	}
	return 0
}

func (m *proto2Message) GetBoolValue() bool {
	if m != nil && m.BoolValue != nil {
		return *m.BoolValue
	}
	return false
}

func (m *proto2Message) GetStringValue() string {
	if m != nil && m.StringValue != nil {
		return *m.StringValue
	}
	return ""
}

func (m *proto2Message) GetRepeatedValue() []string {
	if m != nil {
		return m.RepeatedValue
	}
	return nil
}

type EnumValue int32

const (
	EnumValue_X EnumValue = 0
	EnumValue_Y EnumValue = 1
	EnumValue_Z EnumValue = 2
)

var EnumValue_name = map[int32]string{
	0: "EnumValue_X",
	1: "EnumValue_Y",
	2: "EnumValue_Z",
}
var EnumValue_value = map[string]int32{
	"EnumValue_X": 0,
	"EnumValue_Y": 1,
	"EnumValue_Z": 2,
}

func init() {
	proto.RegisterEnum("runtime_test_api.EnumValue", EnumValue_name, EnumValue_value)
}
