// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.19.4
// source: turbo/v1/relay.proto

package pbturbo

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	_ "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

var File_turbo_v1_relay_proto protoreflect.FileDescriptor

var file_turbo_v1_relay_proto_rawDesc = []byte{
	0x0a, 0x14, 0x74, 0x75, 0x72, 0x62, 0x6f, 0x2f, 0x76, 0x31, 0x2f, 0x72, 0x65, 0x6c, 0x61, 0x79,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x08, 0x74, 0x75, 0x72, 0x62, 0x6f, 0x2e, 0x76, 0x31,
	0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x32, 0x07, 0x0a, 0x05, 0x52, 0x65, 0x6c, 0x61, 0x79, 0x42, 0x36, 0x5a, 0x34, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6d, 0x61, 0x73, 0x73, 0x62, 0x69, 0x74,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2f, 0x74, 0x75, 0x72, 0x62, 0x6f, 0x2f, 0x70,
	0x62, 0x2f, 0x74, 0x75, 0x72, 0x62, 0x6f, 0x2f, 0x76, 0x31, 0x3b, 0x70, 0x62, 0x74, 0x75, 0x72,
	0x62, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var file_turbo_v1_relay_proto_goTypes = []interface{}{}
var file_turbo_v1_relay_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_turbo_v1_relay_proto_init() }
func file_turbo_v1_relay_proto_init() {
	if File_turbo_v1_relay_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_turbo_v1_relay_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_turbo_v1_relay_proto_goTypes,
		DependencyIndexes: file_turbo_v1_relay_proto_depIdxs,
	}.Build()
	File_turbo_v1_relay_proto = out.File
	file_turbo_v1_relay_proto_rawDesc = nil
	file_turbo_v1_relay_proto_goTypes = nil
	file_turbo_v1_relay_proto_depIdxs = nil
}
