// SPDX-License-Identifier: MIT

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.21.1
// source: ipc-sem.proto

package images

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

type IpcSemEntry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Desc  *IpcDescEntry `protobuf:"bytes,1,req,name=desc" json:"desc,omitempty"`
	Nsems *uint32       `protobuf:"varint,2,req,name=nsems" json:"nsems,omitempty"`
}

func (x *IpcSemEntry) Reset() {
	*x = IpcSemEntry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ipc_sem_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *IpcSemEntry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IpcSemEntry) ProtoMessage() {}

func (x *IpcSemEntry) ProtoReflect() protoreflect.Message {
	mi := &file_ipc_sem_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IpcSemEntry.ProtoReflect.Descriptor instead.
func (*IpcSemEntry) Descriptor() ([]byte, []int) {
	return file_ipc_sem_proto_rawDescGZIP(), []int{0}
}

func (x *IpcSemEntry) GetDesc() *IpcDescEntry {
	if x != nil {
		return x.Desc
	}
	return nil
}

func (x *IpcSemEntry) GetNsems() uint32 {
	if x != nil && x.Nsems != nil {
		return *x.Nsems
	}
	return 0
}

var File_ipc_sem_proto protoreflect.FileDescriptor

var file_ipc_sem_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x69, 0x70, 0x63, 0x2d, 0x73, 0x65, 0x6d, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x0e, 0x69, 0x70, 0x63, 0x2d, 0x64, 0x65, 0x73, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22,
	0x4a, 0x0a, 0x0d, 0x69, 0x70, 0x63, 0x5f, 0x73, 0x65, 0x6d, 0x5f, 0x65, 0x6e, 0x74, 0x72, 0x79,
	0x12, 0x23, 0x0a, 0x04, 0x64, 0x65, 0x73, 0x63, 0x18, 0x01, 0x20, 0x02, 0x28, 0x0b, 0x32, 0x0f,
	0x2e, 0x69, 0x70, 0x63, 0x5f, 0x64, 0x65, 0x73, 0x63, 0x5f, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x52,
	0x04, 0x64, 0x65, 0x73, 0x63, 0x12, 0x14, 0x0a, 0x05, 0x6e, 0x73, 0x65, 0x6d, 0x73, 0x18, 0x02,
	0x20, 0x02, 0x28, 0x0d, 0x52, 0x05, 0x6e, 0x73, 0x65, 0x6d, 0x73,
}

var (
	file_ipc_sem_proto_rawDescOnce sync.Once
	file_ipc_sem_proto_rawDescData = file_ipc_sem_proto_rawDesc
)

func file_ipc_sem_proto_rawDescGZIP() []byte {
	file_ipc_sem_proto_rawDescOnce.Do(func() {
		file_ipc_sem_proto_rawDescData = protoimpl.X.CompressGZIP(file_ipc_sem_proto_rawDescData)
	})
	return file_ipc_sem_proto_rawDescData
}

var file_ipc_sem_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_ipc_sem_proto_goTypes = []interface{}{
	(*IpcSemEntry)(nil),  // 0: ipc_sem_entry
	(*IpcDescEntry)(nil), // 1: ipc_desc_entry
}
var file_ipc_sem_proto_depIdxs = []int32{
	1, // 0: ipc_sem_entry.desc:type_name -> ipc_desc_entry
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_ipc_sem_proto_init() }
func file_ipc_sem_proto_init() {
	if File_ipc_sem_proto != nil {
		return
	}
	file_ipc_desc_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_ipc_sem_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*IpcSemEntry); i {
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
			RawDescriptor: file_ipc_sem_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_ipc_sem_proto_goTypes,
		DependencyIndexes: file_ipc_sem_proto_depIdxs,
		MessageInfos:      file_ipc_sem_proto_msgTypes,
	}.Build()
	File_ipc_sem_proto = out.File
	file_ipc_sem_proto_rawDesc = nil
	file_ipc_sem_proto_goTypes = nil
	file_ipc_sem_proto_depIdxs = nil
}
