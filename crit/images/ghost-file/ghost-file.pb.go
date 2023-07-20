// SPDX-License-Identifier: MIT

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        v4.23.4
// source: ghost-file.proto

package ghost_file

import (
	_ "github.com/checkpoint-restore/go-criu/v6/crit/images/opts"
	time "github.com/checkpoint-restore/go-criu/v6/crit/images/time"
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

type GhostFileEntry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Uid    *uint32       `protobuf:"varint,1,req,name=uid" json:"uid,omitempty"`
	Gid    *uint32       `protobuf:"varint,2,req,name=gid" json:"gid,omitempty"`
	Mode   *uint32       `protobuf:"varint,3,req,name=mode" json:"mode,omitempty"`
	Dev    *uint32       `protobuf:"varint,4,opt,name=dev" json:"dev,omitempty"`
	Ino    *uint64       `protobuf:"varint,5,opt,name=ino" json:"ino,omitempty"`
	Rdev   *uint32       `protobuf:"varint,6,opt,name=rdev" json:"rdev,omitempty"`
	Atim   *time.Timeval `protobuf:"bytes,7,opt,name=atim" json:"atim,omitempty"`
	Mtim   *time.Timeval `protobuf:"bytes,8,opt,name=mtim" json:"mtim,omitempty"`
	Chunks *bool         `protobuf:"varint,9,opt,name=chunks" json:"chunks,omitempty"`
	Size   *uint64       `protobuf:"varint,10,opt,name=size" json:"size,omitempty"`
	// this field makes sense only when S_ISLNK(mode)
	SymlnkTarget *string `protobuf:"bytes,11,opt,name=symlnk_target,json=symlnkTarget" json:"symlnk_target,omitempty"`
}

func (x *GhostFileEntry) Reset() {
	*x = GhostFileEntry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ghost_file_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GhostFileEntry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GhostFileEntry) ProtoMessage() {}

func (x *GhostFileEntry) ProtoReflect() protoreflect.Message {
	mi := &file_ghost_file_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GhostFileEntry.ProtoReflect.Descriptor instead.
func (*GhostFileEntry) Descriptor() ([]byte, []int) {
	return file_ghost_file_proto_rawDescGZIP(), []int{0}
}

func (x *GhostFileEntry) GetUid() uint32 {
	if x != nil && x.Uid != nil {
		return *x.Uid
	}
	return 0
}

func (x *GhostFileEntry) GetGid() uint32 {
	if x != nil && x.Gid != nil {
		return *x.Gid
	}
	return 0
}

func (x *GhostFileEntry) GetMode() uint32 {
	if x != nil && x.Mode != nil {
		return *x.Mode
	}
	return 0
}

func (x *GhostFileEntry) GetDev() uint32 {
	if x != nil && x.Dev != nil {
		return *x.Dev
	}
	return 0
}

func (x *GhostFileEntry) GetIno() uint64 {
	if x != nil && x.Ino != nil {
		return *x.Ino
	}
	return 0
}

func (x *GhostFileEntry) GetRdev() uint32 {
	if x != nil && x.Rdev != nil {
		return *x.Rdev
	}
	return 0
}

func (x *GhostFileEntry) GetAtim() *time.Timeval {
	if x != nil {
		return x.Atim
	}
	return nil
}

func (x *GhostFileEntry) GetMtim() *time.Timeval {
	if x != nil {
		return x.Mtim
	}
	return nil
}

func (x *GhostFileEntry) GetChunks() bool {
	if x != nil && x.Chunks != nil {
		return *x.Chunks
	}
	return false
}

func (x *GhostFileEntry) GetSize() uint64 {
	if x != nil && x.Size != nil {
		return *x.Size
	}
	return 0
}

func (x *GhostFileEntry) GetSymlnkTarget() string {
	if x != nil && x.SymlnkTarget != nil {
		return *x.SymlnkTarget
	}
	return ""
}

type GhostChunkEntry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Len *uint64 `protobuf:"varint,1,req,name=len" json:"len,omitempty"`
	Off *uint64 `protobuf:"varint,2,req,name=off" json:"off,omitempty"`
}

func (x *GhostChunkEntry) Reset() {
	*x = GhostChunkEntry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ghost_file_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GhostChunkEntry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GhostChunkEntry) ProtoMessage() {}

func (x *GhostChunkEntry) ProtoReflect() protoreflect.Message {
	mi := &file_ghost_file_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GhostChunkEntry.ProtoReflect.Descriptor instead.
func (*GhostChunkEntry) Descriptor() ([]byte, []int) {
	return file_ghost_file_proto_rawDescGZIP(), []int{1}
}

func (x *GhostChunkEntry) GetLen() uint64 {
	if x != nil && x.Len != nil {
		return *x.Len
	}
	return 0
}

func (x *GhostChunkEntry) GetOff() uint64 {
	if x != nil && x.Off != nil {
		return *x.Off
	}
	return 0
}

var File_ghost_file_proto protoreflect.FileDescriptor

var file_ghost_file_proto_rawDesc = []byte{
	0x0a, 0x10, 0x67, 0x68, 0x6f, 0x73, 0x74, 0x2d, 0x66, 0x69, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x0a, 0x6f, 0x70, 0x74, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0a,
	0x74, 0x69, 0x6d, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x9f, 0x02, 0x0a, 0x10, 0x67,
	0x68, 0x6f, 0x73, 0x74, 0x5f, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x12,
	0x10, 0x0a, 0x03, 0x75, 0x69, 0x64, 0x18, 0x01, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x03, 0x75, 0x69,
	0x64, 0x12, 0x10, 0x0a, 0x03, 0x67, 0x69, 0x64, 0x18, 0x02, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x03,
	0x67, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6d, 0x6f, 0x64, 0x65, 0x18, 0x03, 0x20, 0x02, 0x28,
	0x0d, 0x52, 0x04, 0x6d, 0x6f, 0x64, 0x65, 0x12, 0x17, 0x0a, 0x03, 0x64, 0x65, 0x76, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x0d, 0x42, 0x05, 0xd2, 0x3f, 0x02, 0x20, 0x01, 0x52, 0x03, 0x64, 0x65, 0x76,
	0x12, 0x10, 0x0a, 0x03, 0x69, 0x6e, 0x6f, 0x18, 0x05, 0x20, 0x01, 0x28, 0x04, 0x52, 0x03, 0x69,
	0x6e, 0x6f, 0x12, 0x1b, 0x0a, 0x04, 0x72, 0x64, 0x65, 0x76, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0d,
	0x42, 0x07, 0xd2, 0x3f, 0x04, 0x20, 0x01, 0x28, 0x01, 0x52, 0x04, 0x72, 0x64, 0x65, 0x76, 0x12,
	0x1c, 0x0a, 0x04, 0x61, 0x74, 0x69, 0x6d, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x08, 0x2e,
	0x74, 0x69, 0x6d, 0x65, 0x76, 0x61, 0x6c, 0x52, 0x04, 0x61, 0x74, 0x69, 0x6d, 0x12, 0x1c, 0x0a,
	0x04, 0x6d, 0x74, 0x69, 0x6d, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x08, 0x2e, 0x74, 0x69,
	0x6d, 0x65, 0x76, 0x61, 0x6c, 0x52, 0x04, 0x6d, 0x74, 0x69, 0x6d, 0x12, 0x16, 0x0a, 0x06, 0x63,
	0x68, 0x75, 0x6e, 0x6b, 0x73, 0x18, 0x09, 0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x63, 0x68, 0x75,
	0x6e, 0x6b, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x0a, 0x20, 0x01, 0x28,
	0x04, 0x52, 0x04, 0x73, 0x69, 0x7a, 0x65, 0x12, 0x23, 0x0a, 0x0d, 0x73, 0x79, 0x6d, 0x6c, 0x6e,
	0x6b, 0x5f, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c,
	0x73, 0x79, 0x6d, 0x6c, 0x6e, 0x6b, 0x54, 0x61, 0x72, 0x67, 0x65, 0x74, 0x22, 0x37, 0x0a, 0x11,
	0x67, 0x68, 0x6f, 0x73, 0x74, 0x5f, 0x63, 0x68, 0x75, 0x6e, 0x6b, 0x5f, 0x65, 0x6e, 0x74, 0x72,
	0x79, 0x12, 0x10, 0x0a, 0x03, 0x6c, 0x65, 0x6e, 0x18, 0x01, 0x20, 0x02, 0x28, 0x04, 0x52, 0x03,
	0x6c, 0x65, 0x6e, 0x12, 0x10, 0x0a, 0x03, 0x6f, 0x66, 0x66, 0x18, 0x02, 0x20, 0x02, 0x28, 0x04,
	0x52, 0x03, 0x6f, 0x66, 0x66,
}

var (
	file_ghost_file_proto_rawDescOnce sync.Once
	file_ghost_file_proto_rawDescData = file_ghost_file_proto_rawDesc
)

func file_ghost_file_proto_rawDescGZIP() []byte {
	file_ghost_file_proto_rawDescOnce.Do(func() {
		file_ghost_file_proto_rawDescData = protoimpl.X.CompressGZIP(file_ghost_file_proto_rawDescData)
	})
	return file_ghost_file_proto_rawDescData
}

var file_ghost_file_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_ghost_file_proto_goTypes = []interface{}{
	(*GhostFileEntry)(nil),  // 0: ghost_file_entry
	(*GhostChunkEntry)(nil), // 1: ghost_chunk_entry
	(*time.Timeval)(nil),    // 2: timeval
}
var file_ghost_file_proto_depIdxs = []int32{
	2, // 0: ghost_file_entry.atim:type_name -> timeval
	2, // 1: ghost_file_entry.mtim:type_name -> timeval
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_ghost_file_proto_init() }
func file_ghost_file_proto_init() {
	if File_ghost_file_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_ghost_file_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GhostFileEntry); i {
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
		file_ghost_file_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GhostChunkEntry); i {
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
			RawDescriptor: file_ghost_file_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_ghost_file_proto_goTypes,
		DependencyIndexes: file_ghost_file_proto_depIdxs,
		MessageInfos:      file_ghost_file_proto_msgTypes,
	}.Build()
	File_ghost_file_proto = out.File
	file_ghost_file_proto_rawDesc = nil
	file_ghost_file_proto_goTypes = nil
	file_ghost_file_proto_depIdxs = nil
}
