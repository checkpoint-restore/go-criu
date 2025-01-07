// SPDX-License-Identifier: MIT

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v5.28.3
// source: file-lock.proto

package file_lock

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

type FileLockEntry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Flag  *uint32 `protobuf:"varint,1,req,name=flag" json:"flag,omitempty"`
	Type  *uint32 `protobuf:"varint,2,req,name=type" json:"type,omitempty"`
	Pid   *int32  `protobuf:"varint,3,req,name=pid" json:"pid,omitempty"`
	Fd    *int32  `protobuf:"varint,4,req,name=fd" json:"fd,omitempty"`
	Start *int64  `protobuf:"varint,5,req,name=start" json:"start,omitempty"`
	Len   *int64  `protobuf:"varint,6,req,name=len" json:"len,omitempty"`
}

func (x *FileLockEntry) Reset() {
	*x = FileLockEntry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_file_lock_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FileLockEntry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FileLockEntry) ProtoMessage() {}

func (x *FileLockEntry) ProtoReflect() protoreflect.Message {
	mi := &file_file_lock_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FileLockEntry.ProtoReflect.Descriptor instead.
func (*FileLockEntry) Descriptor() ([]byte, []int) {
	return file_file_lock_proto_rawDescGZIP(), []int{0}
}

func (x *FileLockEntry) GetFlag() uint32 {
	if x != nil && x.Flag != nil {
		return *x.Flag
	}
	return 0
}

func (x *FileLockEntry) GetType() uint32 {
	if x != nil && x.Type != nil {
		return *x.Type
	}
	return 0
}

func (x *FileLockEntry) GetPid() int32 {
	if x != nil && x.Pid != nil {
		return *x.Pid
	}
	return 0
}

func (x *FileLockEntry) GetFd() int32 {
	if x != nil && x.Fd != nil {
		return *x.Fd
	}
	return 0
}

func (x *FileLockEntry) GetStart() int64 {
	if x != nil && x.Start != nil {
		return *x.Start
	}
	return 0
}

func (x *FileLockEntry) GetLen() int64 {
	if x != nil && x.Len != nil {
		return *x.Len
	}
	return 0
}

var File_file_lock_proto protoreflect.FileDescriptor

var file_file_lock_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x66, 0x69, 0x6c, 0x65, 0x2d, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x22, 0x83, 0x01, 0x0a, 0x0f, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x6c, 0x6f, 0x63, 0x6b, 0x5f,
	0x65, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x12, 0x0a, 0x04, 0x66, 0x6c, 0x61, 0x67, 0x18, 0x01, 0x20,
	0x02, 0x28, 0x0d, 0x52, 0x04, 0x66, 0x6c, 0x61, 0x67, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x79, 0x70,
	0x65, 0x18, 0x02, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x10, 0x0a,
	0x03, 0x70, 0x69, 0x64, 0x18, 0x03, 0x20, 0x02, 0x28, 0x05, 0x52, 0x03, 0x70, 0x69, 0x64, 0x12,
	0x0e, 0x0a, 0x02, 0x66, 0x64, 0x18, 0x04, 0x20, 0x02, 0x28, 0x05, 0x52, 0x02, 0x66, 0x64, 0x12,
	0x14, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x72, 0x74, 0x18, 0x05, 0x20, 0x02, 0x28, 0x03, 0x52, 0x05,
	0x73, 0x74, 0x61, 0x72, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x6c, 0x65, 0x6e, 0x18, 0x06, 0x20, 0x02,
	0x28, 0x03, 0x52, 0x03, 0x6c, 0x65, 0x6e,
}

var (
	file_file_lock_proto_rawDescOnce sync.Once
	file_file_lock_proto_rawDescData = file_file_lock_proto_rawDesc
)

func file_file_lock_proto_rawDescGZIP() []byte {
	file_file_lock_proto_rawDescOnce.Do(func() {
		file_file_lock_proto_rawDescData = protoimpl.X.CompressGZIP(file_file_lock_proto_rawDescData)
	})
	return file_file_lock_proto_rawDescData
}

var file_file_lock_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_file_lock_proto_goTypes = []interface{}{
	(*FileLockEntry)(nil), // 0: file_lock_entry
}
var file_file_lock_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_file_lock_proto_init() }
func file_file_lock_proto_init() {
	if File_file_lock_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_file_lock_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FileLockEntry); i {
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
			RawDescriptor: file_file_lock_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_file_lock_proto_goTypes,
		DependencyIndexes: file_file_lock_proto_depIdxs,
		MessageInfos:      file_file_lock_proto_msgTypes,
	}.Build()
	File_file_lock_proto = out.File
	file_file_lock_proto_rawDesc = nil
	file_file_lock_proto_goTypes = nil
	file_file_lock_proto_depIdxs = nil
}
