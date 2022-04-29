// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.19.4
// source: common/internal_code.proto

package common

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

type InternalCode int32

const (
	//general
	InternalCode_Ok1              InternalCode = 0
	InternalCode_InternalError1   InternalCode = 1
	InternalCode_InvalidParams1   InternalCode = 2
	InternalCode_OperationFailed1 InternalCode = 3
	InternalCode_NotFindData1     InternalCode = 4
)

// Enum value maps for InternalCode.
var (
	InternalCode_name = map[int32]string{
		0: "Ok1",
		1: "InternalError1",
		2: "InvalidParams1",
		3: "OperationFailed1",
		4: "NotFindData1",
	}
	InternalCode_value = map[string]int32{
		"Ok1":              0,
		"InternalError1":   1,
		"InvalidParams1":   2,
		"OperationFailed1": 3,
		"NotFindData1":     4,
	}
)

func (x InternalCode) Enum() *InternalCode {
	p := new(InternalCode)
	*p = x
	return p
}

func (x InternalCode) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (InternalCode) Descriptor() protoreflect.EnumDescriptor {
	return file_common_internal_code_proto_enumTypes[0].Descriptor()
}

func (InternalCode) Type() protoreflect.EnumType {
	return &file_common_internal_code_proto_enumTypes[0]
}

func (x InternalCode) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use InternalCode.Descriptor instead.
func (InternalCode) EnumDescriptor() ([]byte, []int) {
	return file_common_internal_code_proto_rawDescGZIP(), []int{0}
}

var File_common_internal_code_proto protoreflect.FileDescriptor

var file_common_internal_code_proto_rawDesc = []byte{
	0x0a, 0x1a, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61,
	0x6c, 0x5f, 0x63, 0x6f, 0x64, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06, 0x63, 0x6f,
	0x6d, 0x6d, 0x6f, 0x6e, 0x2a, 0x67, 0x0a, 0x0c, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c,
	0x43, 0x6f, 0x64, 0x65, 0x12, 0x07, 0x0a, 0x03, 0x4f, 0x6b, 0x31, 0x10, 0x00, 0x12, 0x12, 0x0a,
	0x0e, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x31, 0x10,
	0x01, 0x12, 0x12, 0x0a, 0x0e, 0x49, 0x6e, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x50, 0x61, 0x72, 0x61,
	0x6d, 0x73, 0x31, 0x10, 0x02, 0x12, 0x14, 0x0a, 0x10, 0x4f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x46, 0x61, 0x69, 0x6c, 0x65, 0x64, 0x31, 0x10, 0x03, 0x12, 0x10, 0x0a, 0x0c, 0x4e,
	0x6f, 0x74, 0x46, 0x69, 0x6e, 0x64, 0x44, 0x61, 0x74, 0x61, 0x31, 0x10, 0x04, 0x42, 0x10, 0x5a,
	0x0e, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x3b, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_common_internal_code_proto_rawDescOnce sync.Once
	file_common_internal_code_proto_rawDescData = file_common_internal_code_proto_rawDesc
)

func file_common_internal_code_proto_rawDescGZIP() []byte {
	file_common_internal_code_proto_rawDescOnce.Do(func() {
		file_common_internal_code_proto_rawDescData = protoimpl.X.CompressGZIP(file_common_internal_code_proto_rawDescData)
	})
	return file_common_internal_code_proto_rawDescData
}

var file_common_internal_code_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_common_internal_code_proto_goTypes = []interface{}{
	(InternalCode)(0), // 0: common.internalCode
}
var file_common_internal_code_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_common_internal_code_proto_init() }
func file_common_internal_code_proto_init() {
	if File_common_internal_code_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_common_internal_code_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_common_internal_code_proto_goTypes,
		DependencyIndexes: file_common_internal_code_proto_depIdxs,
		EnumInfos:         file_common_internal_code_proto_enumTypes,
	}.Build()
	File_common_internal_code_proto = out.File
	file_common_internal_code_proto_rawDesc = nil
	file_common_internal_code_proto_goTypes = nil
	file_common_internal_code_proto_depIdxs = nil
}
