// SPDX-License-Identifier: MIT

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v5.28.3
// source: sit.proto

package sit

import (
	_ "github.com/checkpoint-restore/go-criu/v7/crit/images/opts"
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

type SitEntry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Link           *uint32  `protobuf:"varint,1,opt,name=link" json:"link,omitempty"`
	Local          []uint32 `protobuf:"varint,2,rep,name=local" json:"local,omitempty"`
	Remote         []uint32 `protobuf:"varint,3,rep,name=remote" json:"remote,omitempty"`
	Ttl            *uint32  `protobuf:"varint,4,opt,name=ttl" json:"ttl,omitempty"`
	Tos            *uint32  `protobuf:"varint,5,opt,name=tos" json:"tos,omitempty"`
	Pmtudisc       *bool    `protobuf:"varint,6,opt,name=pmtudisc" json:"pmtudisc,omitempty"`
	Proto          *uint32  `protobuf:"varint,7,opt,name=proto" json:"proto,omitempty"`
	Flags          *uint32  `protobuf:"varint,8,opt,name=flags" json:"flags,omitempty"`
	EncapType      *uint32  `protobuf:"varint,9,opt,name=encap_type,json=encapType" json:"encap_type,omitempty"`
	EncapFlags     *uint32  `protobuf:"varint,10,opt,name=encap_flags,json=encapFlags" json:"encap_flags,omitempty"`
	EncapSport     *uint32  `protobuf:"varint,11,opt,name=encap_sport,json=encapSport" json:"encap_sport,omitempty"`
	EncapDport     *uint32  `protobuf:"varint,12,opt,name=encap_dport,json=encapDport" json:"encap_dport,omitempty"`
	RdPrefixlen    *uint32  `protobuf:"varint,13,opt,name=rd_prefixlen,json=rdPrefixlen" json:"rd_prefixlen,omitempty"`
	RdPrefix       []uint32 `protobuf:"varint,14,rep,name=rd_prefix,json=rdPrefix" json:"rd_prefix,omitempty"`
	RelayPrefixlen *uint32  `protobuf:"varint,15,opt,name=relay_prefixlen,json=relayPrefixlen" json:"relay_prefixlen,omitempty"`
	RelayPrefix    []uint32 `protobuf:"varint,16,rep,name=relay_prefix,json=relayPrefix" json:"relay_prefix,omitempty"`
}

func (x *SitEntry) Reset() {
	*x = SitEntry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_sit_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SitEntry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SitEntry) ProtoMessage() {}

func (x *SitEntry) ProtoReflect() protoreflect.Message {
	mi := &file_sit_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SitEntry.ProtoReflect.Descriptor instead.
func (*SitEntry) Descriptor() ([]byte, []int) {
	return file_sit_proto_rawDescGZIP(), []int{0}
}

func (x *SitEntry) GetLink() uint32 {
	if x != nil && x.Link != nil {
		return *x.Link
	}
	return 0
}

func (x *SitEntry) GetLocal() []uint32 {
	if x != nil {
		return x.Local
	}
	return nil
}

func (x *SitEntry) GetRemote() []uint32 {
	if x != nil {
		return x.Remote
	}
	return nil
}

func (x *SitEntry) GetTtl() uint32 {
	if x != nil && x.Ttl != nil {
		return *x.Ttl
	}
	return 0
}

func (x *SitEntry) GetTos() uint32 {
	if x != nil && x.Tos != nil {
		return *x.Tos
	}
	return 0
}

func (x *SitEntry) GetPmtudisc() bool {
	if x != nil && x.Pmtudisc != nil {
		return *x.Pmtudisc
	}
	return false
}

func (x *SitEntry) GetProto() uint32 {
	if x != nil && x.Proto != nil {
		return *x.Proto
	}
	return 0
}

func (x *SitEntry) GetFlags() uint32 {
	if x != nil && x.Flags != nil {
		return *x.Flags
	}
	return 0
}

func (x *SitEntry) GetEncapType() uint32 {
	if x != nil && x.EncapType != nil {
		return *x.EncapType
	}
	return 0
}

func (x *SitEntry) GetEncapFlags() uint32 {
	if x != nil && x.EncapFlags != nil {
		return *x.EncapFlags
	}
	return 0
}

func (x *SitEntry) GetEncapSport() uint32 {
	if x != nil && x.EncapSport != nil {
		return *x.EncapSport
	}
	return 0
}

func (x *SitEntry) GetEncapDport() uint32 {
	if x != nil && x.EncapDport != nil {
		return *x.EncapDport
	}
	return 0
}

func (x *SitEntry) GetRdPrefixlen() uint32 {
	if x != nil && x.RdPrefixlen != nil {
		return *x.RdPrefixlen
	}
	return 0
}

func (x *SitEntry) GetRdPrefix() []uint32 {
	if x != nil {
		return x.RdPrefix
	}
	return nil
}

func (x *SitEntry) GetRelayPrefixlen() uint32 {
	if x != nil && x.RelayPrefixlen != nil {
		return *x.RelayPrefixlen
	}
	return 0
}

func (x *SitEntry) GetRelayPrefix() []uint32 {
	if x != nil {
		return x.RelayPrefix
	}
	return nil
}

var File_sit_proto protoreflect.FileDescriptor

var file_sit_proto_rawDesc = []byte{
	0x0a, 0x09, 0x73, 0x69, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0a, 0x6f, 0x70, 0x74,
	0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xe3, 0x03, 0x0a, 0x09, 0x73, 0x69, 0x74, 0x5f,
	0x65, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x12, 0x0a, 0x04, 0x6c, 0x69, 0x6e, 0x6b, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0d, 0x52, 0x04, 0x6c, 0x69, 0x6e, 0x6b, 0x12, 0x1b, 0x0a, 0x05, 0x6c, 0x6f, 0x63,
	0x61, 0x6c, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0d, 0x42, 0x05, 0xd2, 0x3f, 0x02, 0x10, 0x01, 0x52,
	0x05, 0x6c, 0x6f, 0x63, 0x61, 0x6c, 0x12, 0x1d, 0x0a, 0x06, 0x72, 0x65, 0x6d, 0x6f, 0x74, 0x65,
	0x18, 0x03, 0x20, 0x03, 0x28, 0x0d, 0x42, 0x05, 0xd2, 0x3f, 0x02, 0x10, 0x01, 0x52, 0x06, 0x72,
	0x65, 0x6d, 0x6f, 0x74, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x74, 0x74, 0x6c, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x0d, 0x52, 0x03, 0x74, 0x74, 0x6c, 0x12, 0x10, 0x0a, 0x03, 0x74, 0x6f, 0x73, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x0d, 0x52, 0x03, 0x74, 0x6f, 0x73, 0x12, 0x1a, 0x0a, 0x08, 0x70, 0x6d, 0x74,
	0x75, 0x64, 0x69, 0x73, 0x63, 0x18, 0x06, 0x20, 0x01, 0x28, 0x08, 0x52, 0x08, 0x70, 0x6d, 0x74,
	0x75, 0x64, 0x69, 0x73, 0x63, 0x12, 0x14, 0x0a, 0x05, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x18, 0x07,
	0x20, 0x01, 0x28, 0x0d, 0x52, 0x05, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x14, 0x0a, 0x05, 0x66,
	0x6c, 0x61, 0x67, 0x73, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x05, 0x66, 0x6c, 0x61, 0x67,
	0x73, 0x12, 0x1d, 0x0a, 0x0a, 0x65, 0x6e, 0x63, 0x61, 0x70, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18,
	0x09, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x09, 0x65, 0x6e, 0x63, 0x61, 0x70, 0x54, 0x79, 0x70, 0x65,
	0x12, 0x1f, 0x0a, 0x0b, 0x65, 0x6e, 0x63, 0x61, 0x70, 0x5f, 0x66, 0x6c, 0x61, 0x67, 0x73, 0x18,
	0x0a, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x0a, 0x65, 0x6e, 0x63, 0x61, 0x70, 0x46, 0x6c, 0x61, 0x67,
	0x73, 0x12, 0x1f, 0x0a, 0x0b, 0x65, 0x6e, 0x63, 0x61, 0x70, 0x5f, 0x73, 0x70, 0x6f, 0x72, 0x74,
	0x18, 0x0b, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x0a, 0x65, 0x6e, 0x63, 0x61, 0x70, 0x53, 0x70, 0x6f,
	0x72, 0x74, 0x12, 0x1f, 0x0a, 0x0b, 0x65, 0x6e, 0x63, 0x61, 0x70, 0x5f, 0x64, 0x70, 0x6f, 0x72,
	0x74, 0x18, 0x0c, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x0a, 0x65, 0x6e, 0x63, 0x61, 0x70, 0x44, 0x70,
	0x6f, 0x72, 0x74, 0x12, 0x21, 0x0a, 0x0c, 0x72, 0x64, 0x5f, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78,
	0x6c, 0x65, 0x6e, 0x18, 0x0d, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x0b, 0x72, 0x64, 0x50, 0x72, 0x65,
	0x66, 0x69, 0x78, 0x6c, 0x65, 0x6e, 0x12, 0x22, 0x0a, 0x09, 0x72, 0x64, 0x5f, 0x70, 0x72, 0x65,
	0x66, 0x69, 0x78, 0x18, 0x0e, 0x20, 0x03, 0x28, 0x0d, 0x42, 0x05, 0xd2, 0x3f, 0x02, 0x10, 0x01,
	0x52, 0x08, 0x72, 0x64, 0x50, 0x72, 0x65, 0x66, 0x69, 0x78, 0x12, 0x27, 0x0a, 0x0f, 0x72, 0x65,
	0x6c, 0x61, 0x79, 0x5f, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78, 0x6c, 0x65, 0x6e, 0x18, 0x0f, 0x20,
	0x01, 0x28, 0x0d, 0x52, 0x0e, 0x72, 0x65, 0x6c, 0x61, 0x79, 0x50, 0x72, 0x65, 0x66, 0x69, 0x78,
	0x6c, 0x65, 0x6e, 0x12, 0x28, 0x0a, 0x0c, 0x72, 0x65, 0x6c, 0x61, 0x79, 0x5f, 0x70, 0x72, 0x65,
	0x66, 0x69, 0x78, 0x18, 0x10, 0x20, 0x03, 0x28, 0x0d, 0x42, 0x05, 0xd2, 0x3f, 0x02, 0x10, 0x01,
	0x52, 0x0b, 0x72, 0x65, 0x6c, 0x61, 0x79, 0x50, 0x72, 0x65, 0x66, 0x69, 0x78,
}

var (
	file_sit_proto_rawDescOnce sync.Once
	file_sit_proto_rawDescData = file_sit_proto_rawDesc
)

func file_sit_proto_rawDescGZIP() []byte {
	file_sit_proto_rawDescOnce.Do(func() {
		file_sit_proto_rawDescData = protoimpl.X.CompressGZIP(file_sit_proto_rawDescData)
	})
	return file_sit_proto_rawDescData
}

var file_sit_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_sit_proto_goTypes = []interface{}{
	(*SitEntry)(nil), // 0: sit_entry
}
var file_sit_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_sit_proto_init() }
func file_sit_proto_init() {
	if File_sit_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_sit_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SitEntry); i {
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
			RawDescriptor: file_sit_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_sit_proto_goTypes,
		DependencyIndexes: file_sit_proto_depIdxs,
		MessageInfos:      file_sit_proto_msgTypes,
	}.Build()
	File_sit_proto = out.File
	file_sit_proto_rawDesc = nil
	file_sit_proto_goTypes = nil
	file_sit_proto_depIdxs = nil
}
