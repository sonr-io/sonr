// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        (unknown)
// source: common/v1/uri.proto

package commonv1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type URI_URIProtocol int32

const (
	URI_HTTPS URI_URIProtocol = 0
	URI_IPFS  URI_URIProtocol = 1
	URI_IPNS  URI_URIProtocol = 2
	URI_DID   URI_URIProtocol = 3
)

// Enum value maps for URI_URIProtocol.
var (
	URI_URIProtocol_name = map[int32]string{
		0: "HTTPS",
		1: "IPFS",
		2: "IPNS",
		3: "DID",
	}
	URI_URIProtocol_value = map[string]int32{
		"HTTPS": 0,
		"IPFS":  1,
		"IPNS":  2,
		"DID":   3,
	}
)

func (x URI_URIProtocol) Enum() *URI_URIProtocol {
	p := new(URI_URIProtocol)
	*p = x
	return p
}

func (x URI_URIProtocol) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (URI_URIProtocol) Descriptor() protoreflect.EnumDescriptor {
	return file_common_v1_uri_proto_enumTypes[0].Descriptor()
}

func (URI_URIProtocol) Type() protoreflect.EnumType {
	return &file_common_v1_uri_proto_enumTypes[0]
}

func (x URI_URIProtocol) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use URI_URIProtocol.Descriptor instead.
func (URI_URIProtocol) EnumDescriptor() ([]byte, []int) {
	return file_common_v1_uri_proto_rawDescGZIP(), []int{0, 0}
}

type URI struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Protocol URI_URIProtocol `protobuf:"varint,1,opt,name=protocol,proto3,enum=common.v1.URI_URIProtocol" json:"protocol,omitempty"`
	Value    string          `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *URI) Reset() {
	*x = URI{}
	if protoimpl.UnsafeEnabled {
		mi := &file_common_v1_uri_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *URI) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*URI) ProtoMessage() {}

func (x *URI) ProtoReflect() protoreflect.Message {
	mi := &file_common_v1_uri_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use URI.ProtoReflect.Descriptor instead.
func (*URI) Descriptor() ([]byte, []int) {
	return file_common_v1_uri_proto_rawDescGZIP(), []int{0}
}

func (x *URI) GetProtocol() URI_URIProtocol {
	if x != nil {
		return x.Protocol
	}
	return URI_HTTPS
}

func (x *URI) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

var File_common_v1_uri_proto protoreflect.FileDescriptor

var file_common_v1_uri_proto_rawDesc = []byte{
	0x0a, 0x13, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x76, 0x31, 0x2f, 0x75, 0x72, 0x69, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x09, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31,
	0x22, 0x8a, 0x01, 0x0a, 0x03, 0x55, 0x52, 0x49, 0x12, 0x36, 0x0a, 0x08, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x63, 0x6f, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1a, 0x2e, 0x63, 0x6f, 0x6d,
	0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x55, 0x52, 0x49, 0x2e, 0x55, 0x52, 0x49, 0x50, 0x72,
	0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x52, 0x08, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c,
	0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x35, 0x0a, 0x0b, 0x55, 0x52, 0x49, 0x50, 0x72, 0x6f,
	0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x12, 0x09, 0x0a, 0x05, 0x48, 0x54, 0x54, 0x50, 0x53, 0x10, 0x00,
	0x12, 0x08, 0x0a, 0x04, 0x49, 0x50, 0x46, 0x53, 0x10, 0x01, 0x12, 0x08, 0x0a, 0x04, 0x49, 0x50,
	0x4e, 0x53, 0x10, 0x02, 0x12, 0x07, 0x0a, 0x03, 0x44, 0x49, 0x44, 0x10, 0x03, 0x42, 0x32, 0x5a,
	0x30, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6f, 0x6e, 0x73, 0x6f,
	0x6e, 0x72, 0x2f, 0x73, 0x6f, 0x6e, 0x72, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x63, 0x6f, 0x6d, 0x6d,
	0x6f, 0x6e, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x3b, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x76,
	0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_common_v1_uri_proto_rawDescOnce sync.Once
	file_common_v1_uri_proto_rawDescData = file_common_v1_uri_proto_rawDesc
)

func file_common_v1_uri_proto_rawDescGZIP() []byte {
	file_common_v1_uri_proto_rawDescOnce.Do(func() {
		file_common_v1_uri_proto_rawDescData = protoimpl.X.CompressGZIP(file_common_v1_uri_proto_rawDescData)
	})
	return file_common_v1_uri_proto_rawDescData
}

var file_common_v1_uri_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_common_v1_uri_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_common_v1_uri_proto_goTypes = []interface{}{
	(URI_URIProtocol)(0), // 0: common.v1.URI.URIProtocol
	(*URI)(nil),          // 1: common.v1.URI
}
var file_common_v1_uri_proto_depIdxs = []int32{
	0, // 0: common.v1.URI.protocol:type_name -> common.v1.URI.URIProtocol
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_common_v1_uri_proto_init() }
func file_common_v1_uri_proto_init() {
	if File_common_v1_uri_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_common_v1_uri_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*URI); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_common_v1_uri_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_common_v1_uri_proto_goTypes,
		DependencyIndexes: file_common_v1_uri_proto_depIdxs,
		EnumInfos:         file_common_v1_uri_proto_enumTypes,
		MessageInfos:      file_common_v1_uri_proto_msgTypes,
	}.Build()
	File_common_v1_uri_proto = out.File
	file_common_v1_uri_proto_rawDesc = nil
	file_common_v1_uri_proto_goTypes = nil
	file_common_v1_uri_proto_depIdxs = nil
}