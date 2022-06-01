// SPDX-License-Identifier: MIT

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.21.1
// source: ipc-shm.proto

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

type IpcShmEntry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Desc        *IpcDescEntry `protobuf:"bytes,1,req,name=desc" json:"desc,omitempty"`
	Size        *uint64       `protobuf:"varint,2,req,name=size" json:"size,omitempty"`
	InPagemaps  *bool         `protobuf:"varint,3,opt,name=in_pagemaps,json=inPagemaps" json:"in_pagemaps,omitempty"`
	HugetlbFlag *uint32       `protobuf:"varint,4,opt,name=hugetlb_flag,json=hugetlbFlag" json:"hugetlb_flag,omitempty"`
}

func (x *IpcShmEntry) Reset() {
	*x = IpcShmEntry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ipc_shm_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *IpcShmEntry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IpcShmEntry) ProtoMessage() {}

func (x *IpcShmEntry) ProtoReflect() protoreflect.Message {
	mi := &file_ipc_shm_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IpcShmEntry.ProtoReflect.Descriptor instead.
func (*IpcShmEntry) Descriptor() ([]byte, []int) {
	return file_ipc_shm_proto_rawDescGZIP(), []int{0}
}

func (x *IpcShmEntry) GetDesc() *IpcDescEntry {
	if x != nil {
		return x.Desc
	}
	return nil
}

func (x *IpcShmEntry) GetSize() uint64 {
	if x != nil && x.Size != nil {
		return *x.Size
	}
	return 0
}

func (x *IpcShmEntry) GetInPagemaps() bool {
	if x != nil && x.InPagemaps != nil {
		return *x.InPagemaps
	}
	return false
}

func (x *IpcShmEntry) GetHugetlbFlag() uint32 {
	if x != nil && x.HugetlbFlag != nil {
		return *x.HugetlbFlag
	}
	return 0
}

var File_ipc_shm_proto protoreflect.FileDescriptor

var file_ipc_shm_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x69, 0x70, 0x63, 0x2d, 0x73, 0x68, 0x6d, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x0e, 0x69, 0x70, 0x63, 0x2d, 0x64, 0x65, 0x73, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22,
	0x8c, 0x01, 0x0a, 0x0d, 0x69, 0x70, 0x63, 0x5f, 0x73, 0x68, 0x6d, 0x5f, 0x65, 0x6e, 0x74, 0x72,
	0x79, 0x12, 0x23, 0x0a, 0x04, 0x64, 0x65, 0x73, 0x63, 0x18, 0x01, 0x20, 0x02, 0x28, 0x0b, 0x32,
	0x0f, 0x2e, 0x69, 0x70, 0x63, 0x5f, 0x64, 0x65, 0x73, 0x63, 0x5f, 0x65, 0x6e, 0x74, 0x72, 0x79,
	0x52, 0x04, 0x64, 0x65, 0x73, 0x63, 0x12, 0x12, 0x0a, 0x04, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x02,
	0x20, 0x02, 0x28, 0x04, 0x52, 0x04, 0x73, 0x69, 0x7a, 0x65, 0x12, 0x1f, 0x0a, 0x0b, 0x69, 0x6e,
	0x5f, 0x70, 0x61, 0x67, 0x65, 0x6d, 0x61, 0x70, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52,
	0x0a, 0x69, 0x6e, 0x50, 0x61, 0x67, 0x65, 0x6d, 0x61, 0x70, 0x73, 0x12, 0x21, 0x0a, 0x0c, 0x68,
	0x75, 0x67, 0x65, 0x74, 0x6c, 0x62, 0x5f, 0x66, 0x6c, 0x61, 0x67, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x0d, 0x52, 0x0b, 0x68, 0x75, 0x67, 0x65, 0x74, 0x6c, 0x62, 0x46, 0x6c, 0x61, 0x67,
}

var (
	file_ipc_shm_proto_rawDescOnce sync.Once
	file_ipc_shm_proto_rawDescData = file_ipc_shm_proto_rawDesc
)

func file_ipc_shm_proto_rawDescGZIP() []byte {
	file_ipc_shm_proto_rawDescOnce.Do(func() {
		file_ipc_shm_proto_rawDescData = protoimpl.X.CompressGZIP(file_ipc_shm_proto_rawDescData)
	})
	return file_ipc_shm_proto_rawDescData
}

var file_ipc_shm_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_ipc_shm_proto_goTypes = []interface{}{
	(*IpcShmEntry)(nil),  // 0: ipc_shm_entry
	(*IpcDescEntry)(nil), // 1: ipc_desc_entry
}
var file_ipc_shm_proto_depIdxs = []int32{
	1, // 0: ipc_shm_entry.desc:type_name -> ipc_desc_entry
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_ipc_shm_proto_init() }
func file_ipc_shm_proto_init() {
	if File_ipc_shm_proto != nil {
		return
	}
	file_ipc_desc_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_ipc_shm_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*IpcShmEntry); i {
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
			RawDescriptor: file_ipc_shm_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_ipc_shm_proto_goTypes,
		DependencyIndexes: file_ipc_shm_proto_depIdxs,
		MessageInfos:      file_ipc_shm_proto_msgTypes,
	}.Build()
	File_ipc_shm_proto = out.File
	file_ipc_shm_proto_rawDesc = nil
	file_ipc_shm_proto_goTypes = nil
	file_ipc_shm_proto_depIdxs = nil
}
