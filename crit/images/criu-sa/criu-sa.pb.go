// SPDX-License-Identifier: MIT

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        v4.23.4
// source: criu-sa.proto

package criu_sa

import (
	_ "github.com/checkpoint-restore/go-criu/v6/crit/images/opts"
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

type SaEntry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Sigaction       *uint64 `protobuf:"varint,1,req,name=sigaction" json:"sigaction,omitempty"`
	Flags           *uint64 `protobuf:"varint,2,req,name=flags" json:"flags,omitempty"`
	Restorer        *uint64 `protobuf:"varint,3,req,name=restorer" json:"restorer,omitempty"`
	Mask            *uint64 `protobuf:"varint,4,req,name=mask" json:"mask,omitempty"`
	CompatSigaction *bool   `protobuf:"varint,5,opt,name=compat_sigaction,json=compatSigaction" json:"compat_sigaction,omitempty"`
	MaskExtended    *uint64 `protobuf:"varint,6,opt,name=mask_extended,json=maskExtended" json:"mask_extended,omitempty"`
}

func (x *SaEntry) Reset() {
	*x = SaEntry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_criu_sa_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SaEntry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SaEntry) ProtoMessage() {}

func (x *SaEntry) ProtoReflect() protoreflect.Message {
	mi := &file_criu_sa_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SaEntry.ProtoReflect.Descriptor instead.
func (*SaEntry) Descriptor() ([]byte, []int) {
	return file_criu_sa_proto_rawDescGZIP(), []int{0}
}

func (x *SaEntry) GetSigaction() uint64 {
	if x != nil && x.Sigaction != nil {
		return *x.Sigaction
	}
	return 0
}

func (x *SaEntry) GetFlags() uint64 {
	if x != nil && x.Flags != nil {
		return *x.Flags
	}
	return 0
}

func (x *SaEntry) GetRestorer() uint64 {
	if x != nil && x.Restorer != nil {
		return *x.Restorer
	}
	return 0
}

func (x *SaEntry) GetMask() uint64 {
	if x != nil && x.Mask != nil {
		return *x.Mask
	}
	return 0
}

func (x *SaEntry) GetCompatSigaction() bool {
	if x != nil && x.CompatSigaction != nil {
		return *x.CompatSigaction
	}
	return false
}

func (x *SaEntry) GetMaskExtended() uint64 {
	if x != nil && x.MaskExtended != nil {
		return *x.MaskExtended
	}
	return 0
}

var File_criu_sa_proto protoreflect.FileDescriptor

var file_criu_sa_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x63, 0x72, 0x69, 0x75, 0x2d, 0x73, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x0a, 0x6f, 0x70, 0x74, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xe1, 0x01, 0x0a, 0x08,
	0x73, 0x61, 0x5f, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x23, 0x0a, 0x09, 0x73, 0x69, 0x67, 0x61,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x02, 0x28, 0x04, 0x42, 0x05, 0xd2, 0x3f, 0x02,
	0x08, 0x01, 0x52, 0x09, 0x73, 0x69, 0x67, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x1b, 0x0a,
	0x05, 0x66, 0x6c, 0x61, 0x67, 0x73, 0x18, 0x02, 0x20, 0x02, 0x28, 0x04, 0x42, 0x05, 0xd2, 0x3f,
	0x02, 0x08, 0x01, 0x52, 0x05, 0x66, 0x6c, 0x61, 0x67, 0x73, 0x12, 0x21, 0x0a, 0x08, 0x72, 0x65,
	0x73, 0x74, 0x6f, 0x72, 0x65, 0x72, 0x18, 0x03, 0x20, 0x02, 0x28, 0x04, 0x42, 0x05, 0xd2, 0x3f,
	0x02, 0x08, 0x01, 0x52, 0x08, 0x72, 0x65, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x72, 0x12, 0x19, 0x0a,
	0x04, 0x6d, 0x61, 0x73, 0x6b, 0x18, 0x04, 0x20, 0x02, 0x28, 0x04, 0x42, 0x05, 0xd2, 0x3f, 0x02,
	0x08, 0x01, 0x52, 0x04, 0x6d, 0x61, 0x73, 0x6b, 0x12, 0x29, 0x0a, 0x10, 0x63, 0x6f, 0x6d, 0x70,
	0x61, 0x74, 0x5f, 0x73, 0x69, 0x67, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x05, 0x20, 0x01,
	0x28, 0x08, 0x52, 0x0f, 0x63, 0x6f, 0x6d, 0x70, 0x61, 0x74, 0x53, 0x69, 0x67, 0x61, 0x63, 0x74,
	0x69, 0x6f, 0x6e, 0x12, 0x2a, 0x0a, 0x0d, 0x6d, 0x61, 0x73, 0x6b, 0x5f, 0x65, 0x78, 0x74, 0x65,
	0x6e, 0x64, 0x65, 0x64, 0x18, 0x06, 0x20, 0x01, 0x28, 0x04, 0x42, 0x05, 0xd2, 0x3f, 0x02, 0x08,
	0x01, 0x52, 0x0c, 0x6d, 0x61, 0x73, 0x6b, 0x45, 0x78, 0x74, 0x65, 0x6e, 0x64, 0x65, 0x64,
}

var (
	file_criu_sa_proto_rawDescOnce sync.Once
	file_criu_sa_proto_rawDescData = file_criu_sa_proto_rawDesc
)

func file_criu_sa_proto_rawDescGZIP() []byte {
	file_criu_sa_proto_rawDescOnce.Do(func() {
		file_criu_sa_proto_rawDescData = protoimpl.X.CompressGZIP(file_criu_sa_proto_rawDescData)
	})
	return file_criu_sa_proto_rawDescData
}

var file_criu_sa_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_criu_sa_proto_goTypes = []interface{}{
	(*SaEntry)(nil), // 0: sa_entry
}
var file_criu_sa_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_criu_sa_proto_init() }
func file_criu_sa_proto_init() {
	if File_criu_sa_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_criu_sa_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SaEntry); i {
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
			RawDescriptor: file_criu_sa_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_criu_sa_proto_goTypes,
		DependencyIndexes: file_criu_sa_proto_depIdxs,
		MessageInfos:      file_criu_sa_proto_msgTypes,
	}.Build()
	File_criu_sa_proto = out.File
	file_criu_sa_proto_rawDesc = nil
	file_criu_sa_proto_goTypes = nil
	file_criu_sa_proto_depIdxs = nil
}
