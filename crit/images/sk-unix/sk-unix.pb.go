// SPDX-License-Identifier: MIT

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.19.6
// source: sk-unix.proto

package sk_unix

import (
	fown "github.com/checkpoint-restore/go-criu/v7/crit/images/fown"
	_ "github.com/checkpoint-restore/go-criu/v7/crit/images/opts"
	sk_opts "github.com/checkpoint-restore/go-criu/v7/crit/images/sk-opts"
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

type FilePermsEntry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Mode *uint32 `protobuf:"varint,1,req,name=mode" json:"mode,omitempty"`
	Uid  *uint32 `protobuf:"varint,2,req,name=uid" json:"uid,omitempty"`
	Gid  *uint32 `protobuf:"varint,3,req,name=gid" json:"gid,omitempty"`
}

func (x *FilePermsEntry) Reset() {
	*x = FilePermsEntry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_sk_unix_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FilePermsEntry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FilePermsEntry) ProtoMessage() {}

func (x *FilePermsEntry) ProtoReflect() protoreflect.Message {
	mi := &file_sk_unix_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FilePermsEntry.ProtoReflect.Descriptor instead.
func (*FilePermsEntry) Descriptor() ([]byte, []int) {
	return file_sk_unix_proto_rawDescGZIP(), []int{0}
}

func (x *FilePermsEntry) GetMode() uint32 {
	if x != nil && x.Mode != nil {
		return *x.Mode
	}
	return 0
}

func (x *FilePermsEntry) GetUid() uint32 {
	if x != nil && x.Uid != nil {
		return *x.Uid
	}
	return 0
}

func (x *FilePermsEntry) GetGid() uint32 {
	if x != nil && x.Gid != nil {
		return *x.Gid
	}
	return 0
}

type UnixSkEntry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Few words about why we need both -- id and ino.
	//
	// The former one is used to link file descriptor from
	// fdinfo image with the unix_sk_entry that should be
	// opened under it.
	//
	// The latter one ties together unix peers -- the peer
	// member on this structure is the ino one of its peer
	// and simetimes vise-versa.
	Id      *uint32              `protobuf:"varint,1,req,name=id" json:"id,omitempty"`
	Ino     *uint32              `protobuf:"varint,2,req,name=ino" json:"ino,omitempty"`
	Type    *uint32              `protobuf:"varint,3,req,name=type" json:"type,omitempty"`
	State   *uint32              `protobuf:"varint,4,req,name=state" json:"state,omitempty"`
	Flags   *uint32              `protobuf:"varint,5,req,name=flags" json:"flags,omitempty"`
	Uflags  *uint32              `protobuf:"varint,6,req,name=uflags" json:"uflags,omitempty"`
	Backlog *uint32              `protobuf:"varint,7,req,name=backlog" json:"backlog,omitempty"`
	Peer    *uint32              `protobuf:"varint,8,req,name=peer" json:"peer,omitempty"`
	Fown    *fown.FownEntry      `protobuf:"bytes,9,req,name=fown" json:"fown,omitempty"`
	Opts    *sk_opts.SkOptsEntry `protobuf:"bytes,10,req,name=opts" json:"opts,omitempty"`
	// Abstract name may contain \0 at any point,
	// so we need to carry it as byte sequence...
	Name      []byte              `protobuf:"bytes,11,req,name=name" json:"name,omitempty"`
	Shutdown  *sk_opts.SkShutdown `protobuf:"varint,12,opt,name=shutdown,enum=SkShutdown" json:"shutdown,omitempty"`
	FilePerms *FilePermsEntry     `protobuf:"bytes,13,opt,name=file_perms,json=filePerms" json:"file_perms,omitempty"`
	// Relative socket name may have prefix.
	NameDir *string `protobuf:"bytes,14,opt,name=name_dir,json=nameDir" json:"name_dir,omitempty"`
	Deleted *bool   `protobuf:"varint,15,opt,name=deleted" json:"deleted,omitempty"`
	NsId    *uint32 `protobuf:"varint,16,opt,name=ns_id,json=nsId" json:"ns_id,omitempty"`
	MntId   *int32  `protobuf:"zigzag32,17,opt,name=mnt_id,json=mntId,def=-1" json:"mnt_id,omitempty"` // Please, don't use field with number 18.
}

// Default values for UnixSkEntry fields.
const (
	Default_UnixSkEntry_MntId = int32(-1)
)

func (x *UnixSkEntry) Reset() {
	*x = UnixSkEntry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_sk_unix_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UnixSkEntry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UnixSkEntry) ProtoMessage() {}

func (x *UnixSkEntry) ProtoReflect() protoreflect.Message {
	mi := &file_sk_unix_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UnixSkEntry.ProtoReflect.Descriptor instead.
func (*UnixSkEntry) Descriptor() ([]byte, []int) {
	return file_sk_unix_proto_rawDescGZIP(), []int{1}
}

func (x *UnixSkEntry) GetId() uint32 {
	if x != nil && x.Id != nil {
		return *x.Id
	}
	return 0
}

func (x *UnixSkEntry) GetIno() uint32 {
	if x != nil && x.Ino != nil {
		return *x.Ino
	}
	return 0
}

func (x *UnixSkEntry) GetType() uint32 {
	if x != nil && x.Type != nil {
		return *x.Type
	}
	return 0
}

func (x *UnixSkEntry) GetState() uint32 {
	if x != nil && x.State != nil {
		return *x.State
	}
	return 0
}

func (x *UnixSkEntry) GetFlags() uint32 {
	if x != nil && x.Flags != nil {
		return *x.Flags
	}
	return 0
}

func (x *UnixSkEntry) GetUflags() uint32 {
	if x != nil && x.Uflags != nil {
		return *x.Uflags
	}
	return 0
}

func (x *UnixSkEntry) GetBacklog() uint32 {
	if x != nil && x.Backlog != nil {
		return *x.Backlog
	}
	return 0
}

func (x *UnixSkEntry) GetPeer() uint32 {
	if x != nil && x.Peer != nil {
		return *x.Peer
	}
	return 0
}

func (x *UnixSkEntry) GetFown() *fown.FownEntry {
	if x != nil {
		return x.Fown
	}
	return nil
}

func (x *UnixSkEntry) GetOpts() *sk_opts.SkOptsEntry {
	if x != nil {
		return x.Opts
	}
	return nil
}

func (x *UnixSkEntry) GetName() []byte {
	if x != nil {
		return x.Name
	}
	return nil
}

func (x *UnixSkEntry) GetShutdown() sk_opts.SkShutdown {
	if x != nil && x.Shutdown != nil {
		return *x.Shutdown
	}
	return sk_opts.SkShutdown(0)
}

func (x *UnixSkEntry) GetFilePerms() *FilePermsEntry {
	if x != nil {
		return x.FilePerms
	}
	return nil
}

func (x *UnixSkEntry) GetNameDir() string {
	if x != nil && x.NameDir != nil {
		return *x.NameDir
	}
	return ""
}

func (x *UnixSkEntry) GetDeleted() bool {
	if x != nil && x.Deleted != nil {
		return *x.Deleted
	}
	return false
}

func (x *UnixSkEntry) GetNsId() uint32 {
	if x != nil && x.NsId != nil {
		return *x.NsId
	}
	return 0
}

func (x *UnixSkEntry) GetMntId() int32 {
	if x != nil && x.MntId != nil {
		return *x.MntId
	}
	return Default_UnixSkEntry_MntId
}

var File_sk_unix_proto protoreflect.FileDescriptor

var file_sk_unix_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x73, 0x6b, 0x2d, 0x75, 0x6e, 0x69, 0x78, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x0a, 0x6f, 0x70, 0x74, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0a, 0x66, 0x6f, 0x77,
	0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0d, 0x73, 0x6b, 0x2d, 0x6f, 0x70, 0x74, 0x73,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x4a, 0x0a, 0x10, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x70,
	0x65, 0x72, 0x6d, 0x73, 0x5f, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x12, 0x0a, 0x04, 0x6d, 0x6f,
	0x64, 0x65, 0x18, 0x01, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x04, 0x6d, 0x6f, 0x64, 0x65, 0x12, 0x10,
	0x0a, 0x03, 0x75, 0x69, 0x64, 0x18, 0x02, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x03, 0x75, 0x69, 0x64,
	0x12, 0x10, 0x0a, 0x03, 0x67, 0x69, 0x64, 0x18, 0x03, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x03, 0x67,
	0x69, 0x64, 0x22, 0x81, 0x04, 0x0a, 0x0d, 0x75, 0x6e, 0x69, 0x78, 0x5f, 0x73, 0x6b, 0x5f, 0x65,
	0x6e, 0x74, 0x72, 0x79, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x02, 0x28, 0x0d,
	0x52, 0x02, 0x69, 0x64, 0x12, 0x10, 0x0a, 0x03, 0x69, 0x6e, 0x6f, 0x18, 0x02, 0x20, 0x02, 0x28,
	0x0d, 0x52, 0x03, 0x69, 0x6e, 0x6f, 0x12, 0x1b, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x03,
	0x20, 0x02, 0x28, 0x0d, 0x42, 0x07, 0xd2, 0x3f, 0x04, 0x32, 0x02, 0x73, 0x6b, 0x52, 0x04, 0x74,
	0x79, 0x70, 0x65, 0x12, 0x1d, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x18, 0x04, 0x20, 0x02,
	0x28, 0x0d, 0x42, 0x07, 0xd2, 0x3f, 0x04, 0x32, 0x02, 0x73, 0x6b, 0x52, 0x05, 0x73, 0x74, 0x61,
	0x74, 0x65, 0x12, 0x1b, 0x0a, 0x05, 0x66, 0x6c, 0x61, 0x67, 0x73, 0x18, 0x05, 0x20, 0x02, 0x28,
	0x0d, 0x42, 0x05, 0xd2, 0x3f, 0x02, 0x08, 0x01, 0x52, 0x05, 0x66, 0x6c, 0x61, 0x67, 0x73, 0x12,
	0x1d, 0x0a, 0x06, 0x75, 0x66, 0x6c, 0x61, 0x67, 0x73, 0x18, 0x06, 0x20, 0x02, 0x28, 0x0d, 0x42,
	0x05, 0xd2, 0x3f, 0x02, 0x08, 0x01, 0x52, 0x06, 0x75, 0x66, 0x6c, 0x61, 0x67, 0x73, 0x12, 0x18,
	0x0a, 0x07, 0x62, 0x61, 0x63, 0x6b, 0x6c, 0x6f, 0x67, 0x18, 0x07, 0x20, 0x02, 0x28, 0x0d, 0x52,
	0x07, 0x62, 0x61, 0x63, 0x6b, 0x6c, 0x6f, 0x67, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x65, 0x65, 0x72,
	0x18, 0x08, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x04, 0x70, 0x65, 0x65, 0x72, 0x12, 0x1f, 0x0a, 0x04,
	0x66, 0x6f, 0x77, 0x6e, 0x18, 0x09, 0x20, 0x02, 0x28, 0x0b, 0x32, 0x0b, 0x2e, 0x66, 0x6f, 0x77,
	0x6e, 0x5f, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x04, 0x66, 0x6f, 0x77, 0x6e, 0x12, 0x22, 0x0a,
	0x04, 0x6f, 0x70, 0x74, 0x73, 0x18, 0x0a, 0x20, 0x02, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x73, 0x6b,
	0x5f, 0x6f, 0x70, 0x74, 0x73, 0x5f, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x04, 0x6f, 0x70, 0x74,
	0x73, 0x12, 0x22, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x0b, 0x20, 0x02, 0x28, 0x0c, 0x42,
	0x0e, 0xd2, 0x3f, 0x0b, 0x3a, 0x09, 0x75, 0x6e, 0x69, 0x78, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x52,
	0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x28, 0x0a, 0x08, 0x73, 0x68, 0x75, 0x74, 0x64, 0x6f, 0x77,
	0x6e, 0x18, 0x0c, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0c, 0x2e, 0x73, 0x6b, 0x5f, 0x73, 0x68, 0x75,
	0x74, 0x64, 0x6f, 0x77, 0x6e, 0x52, 0x08, 0x73, 0x68, 0x75, 0x74, 0x64, 0x6f, 0x77, 0x6e, 0x12,
	0x30, 0x0a, 0x0a, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x70, 0x65, 0x72, 0x6d, 0x73, 0x18, 0x0d, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x70, 0x65, 0x72, 0x6d, 0x73,
	0x5f, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x09, 0x66, 0x69, 0x6c, 0x65, 0x50, 0x65, 0x72, 0x6d,
	0x73, 0x12, 0x19, 0x0a, 0x08, 0x6e, 0x61, 0x6d, 0x65, 0x5f, 0x64, 0x69, 0x72, 0x18, 0x0e, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x07, 0x6e, 0x61, 0x6d, 0x65, 0x44, 0x69, 0x72, 0x12, 0x18, 0x0a, 0x07,
	0x64, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x64, 0x18, 0x0f, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x64,
	0x65, 0x6c, 0x65, 0x74, 0x65, 0x64, 0x12, 0x13, 0x0a, 0x05, 0x6e, 0x73, 0x5f, 0x69, 0x64, 0x18,
	0x10, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x04, 0x6e, 0x73, 0x49, 0x64, 0x12, 0x19, 0x0a, 0x06, 0x6d,
	0x6e, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x11, 0x20, 0x01, 0x28, 0x11, 0x3a, 0x02, 0x2d, 0x31, 0x52,
	0x05, 0x6d, 0x6e, 0x74, 0x49, 0x64,
}

var (
	file_sk_unix_proto_rawDescOnce sync.Once
	file_sk_unix_proto_rawDescData = file_sk_unix_proto_rawDesc
)

func file_sk_unix_proto_rawDescGZIP() []byte {
	file_sk_unix_proto_rawDescOnce.Do(func() {
		file_sk_unix_proto_rawDescData = protoimpl.X.CompressGZIP(file_sk_unix_proto_rawDescData)
	})
	return file_sk_unix_proto_rawDescData
}

var file_sk_unix_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_sk_unix_proto_goTypes = []interface{}{
	(*FilePermsEntry)(nil),      // 0: file_perms_entry
	(*UnixSkEntry)(nil),         // 1: unix_sk_entry
	(*fown.FownEntry)(nil),      // 2: fown_entry
	(*sk_opts.SkOptsEntry)(nil), // 3: sk_opts_entry
	(sk_opts.SkShutdown)(0),     // 4: sk_shutdown
}
var file_sk_unix_proto_depIdxs = []int32{
	2, // 0: unix_sk_entry.fown:type_name -> fown_entry
	3, // 1: unix_sk_entry.opts:type_name -> sk_opts_entry
	4, // 2: unix_sk_entry.shutdown:type_name -> sk_shutdown
	0, // 3: unix_sk_entry.file_perms:type_name -> file_perms_entry
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_sk_unix_proto_init() }
func file_sk_unix_proto_init() {
	if File_sk_unix_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_sk_unix_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FilePermsEntry); i {
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
		file_sk_unix_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UnixSkEntry); i {
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
			RawDescriptor: file_sk_unix_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_sk_unix_proto_goTypes,
		DependencyIndexes: file_sk_unix_proto_depIdxs,
		MessageInfos:      file_sk_unix_proto_msgTypes,
	}.Build()
	File_sk_unix_proto = out.File
	file_sk_unix_proto_rawDesc = nil
	file_sk_unix_proto_goTypes = nil
	file_sk_unix_proto_depIdxs = nil
}
