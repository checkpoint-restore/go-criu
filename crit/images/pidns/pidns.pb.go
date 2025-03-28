// SPDX-License-Identifier: MIT

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.19.6
// source: pidns.proto

package pidns

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

type PidnsEntry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ExtKey *string `protobuf:"bytes,1,opt,name=ext_key,json=extKey" json:"ext_key,omitempty"`
}

func (x *PidnsEntry) Reset() {
	*x = PidnsEntry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pidns_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PidnsEntry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PidnsEntry) ProtoMessage() {}

func (x *PidnsEntry) ProtoReflect() protoreflect.Message {
	mi := &file_pidns_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PidnsEntry.ProtoReflect.Descriptor instead.
func (*PidnsEntry) Descriptor() ([]byte, []int) {
	return file_pidns_proto_rawDescGZIP(), []int{0}
}

func (x *PidnsEntry) GetExtKey() string {
	if x != nil && x.ExtKey != nil {
		return *x.ExtKey
	}
	return ""
}

var File_pidns_proto protoreflect.FileDescriptor

var file_pidns_proto_rawDesc = []byte{
	0x0a, 0x0b, 0x70, 0x69, 0x64, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x26, 0x0a,
	0x0b, 0x70, 0x69, 0x64, 0x6e, 0x73, 0x5f, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x17, 0x0a, 0x07,
	0x65, 0x78, 0x74, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x65,
	0x78, 0x74, 0x4b, 0x65, 0x79,
}

var (
	file_pidns_proto_rawDescOnce sync.Once
	file_pidns_proto_rawDescData = file_pidns_proto_rawDesc
)

func file_pidns_proto_rawDescGZIP() []byte {
	file_pidns_proto_rawDescOnce.Do(func() {
		file_pidns_proto_rawDescData = protoimpl.X.CompressGZIP(file_pidns_proto_rawDescData)
	})
	return file_pidns_proto_rawDescData
}

var file_pidns_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_pidns_proto_goTypes = []interface{}{
	(*PidnsEntry)(nil), // 0: pidns_entry
}
var file_pidns_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_pidns_proto_init() }
func file_pidns_proto_init() {
	if File_pidns_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pidns_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PidnsEntry); i {
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
			RawDescriptor: file_pidns_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_pidns_proto_goTypes,
		DependencyIndexes: file_pidns_proto_depIdxs,
		MessageInfos:      file_pidns_proto_msgTypes,
	}.Build()
	File_pidns_proto = out.File
	file_pidns_proto_rawDesc = nil
	file_pidns_proto_goTypes = nil
	file_pidns_proto_depIdxs = nil
}
