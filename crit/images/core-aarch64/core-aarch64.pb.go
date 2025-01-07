// SPDX-License-Identifier: MIT

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v5.28.3
// source: core-aarch64.proto

package core_aarch64

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

type UserAarch64RegsEntry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Regs   []uint64 `protobuf:"varint,1,rep,name=regs" json:"regs,omitempty"`
	Sp     *uint64  `protobuf:"varint,2,req,name=sp" json:"sp,omitempty"`
	Pc     *uint64  `protobuf:"varint,3,req,name=pc" json:"pc,omitempty"`
	Pstate *uint64  `protobuf:"varint,4,req,name=pstate" json:"pstate,omitempty"`
}

func (x *UserAarch64RegsEntry) Reset() {
	*x = UserAarch64RegsEntry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_core_aarch64_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UserAarch64RegsEntry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UserAarch64RegsEntry) ProtoMessage() {}

func (x *UserAarch64RegsEntry) ProtoReflect() protoreflect.Message {
	mi := &file_core_aarch64_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UserAarch64RegsEntry.ProtoReflect.Descriptor instead.
func (*UserAarch64RegsEntry) Descriptor() ([]byte, []int) {
	return file_core_aarch64_proto_rawDescGZIP(), []int{0}
}

func (x *UserAarch64RegsEntry) GetRegs() []uint64 {
	if x != nil {
		return x.Regs
	}
	return nil
}

func (x *UserAarch64RegsEntry) GetSp() uint64 {
	if x != nil && x.Sp != nil {
		return *x.Sp
	}
	return 0
}

func (x *UserAarch64RegsEntry) GetPc() uint64 {
	if x != nil && x.Pc != nil {
		return *x.Pc
	}
	return 0
}

func (x *UserAarch64RegsEntry) GetPstate() uint64 {
	if x != nil && x.Pstate != nil {
		return *x.Pstate
	}
	return 0
}

type UserAarch64FpsimdContextEntry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Vregs []uint64 `protobuf:"varint,1,rep,name=vregs" json:"vregs,omitempty"`
	Fpsr  *uint32  `protobuf:"varint,2,req,name=fpsr" json:"fpsr,omitempty"`
	Fpcr  *uint32  `protobuf:"varint,3,req,name=fpcr" json:"fpcr,omitempty"`
}

func (x *UserAarch64FpsimdContextEntry) Reset() {
	*x = UserAarch64FpsimdContextEntry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_core_aarch64_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UserAarch64FpsimdContextEntry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UserAarch64FpsimdContextEntry) ProtoMessage() {}

func (x *UserAarch64FpsimdContextEntry) ProtoReflect() protoreflect.Message {
	mi := &file_core_aarch64_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UserAarch64FpsimdContextEntry.ProtoReflect.Descriptor instead.
func (*UserAarch64FpsimdContextEntry) Descriptor() ([]byte, []int) {
	return file_core_aarch64_proto_rawDescGZIP(), []int{1}
}

func (x *UserAarch64FpsimdContextEntry) GetVregs() []uint64 {
	if x != nil {
		return x.Vregs
	}
	return nil
}

func (x *UserAarch64FpsimdContextEntry) GetFpsr() uint32 {
	if x != nil && x.Fpsr != nil {
		return *x.Fpsr
	}
	return 0
}

func (x *UserAarch64FpsimdContextEntry) GetFpcr() uint32 {
	if x != nil && x.Fpcr != nil {
		return *x.Fpcr
	}
	return 0
}

type ThreadInfoAarch64 struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ClearTidAddr *uint64                        `protobuf:"varint,1,req,name=clear_tid_addr,json=clearTidAddr" json:"clear_tid_addr,omitempty"`
	Tls          *uint64                        `protobuf:"varint,2,req,name=tls" json:"tls,omitempty"`
	Gpregs       *UserAarch64RegsEntry          `protobuf:"bytes,3,req,name=gpregs" json:"gpregs,omitempty"`
	Fpsimd       *UserAarch64FpsimdContextEntry `protobuf:"bytes,4,req,name=fpsimd" json:"fpsimd,omitempty"`
}

func (x *ThreadInfoAarch64) Reset() {
	*x = ThreadInfoAarch64{}
	if protoimpl.UnsafeEnabled {
		mi := &file_core_aarch64_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ThreadInfoAarch64) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ThreadInfoAarch64) ProtoMessage() {}

func (x *ThreadInfoAarch64) ProtoReflect() protoreflect.Message {
	mi := &file_core_aarch64_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ThreadInfoAarch64.ProtoReflect.Descriptor instead.
func (*ThreadInfoAarch64) Descriptor() ([]byte, []int) {
	return file_core_aarch64_proto_rawDescGZIP(), []int{2}
}

func (x *ThreadInfoAarch64) GetClearTidAddr() uint64 {
	if x != nil && x.ClearTidAddr != nil {
		return *x.ClearTidAddr
	}
	return 0
}

func (x *ThreadInfoAarch64) GetTls() uint64 {
	if x != nil && x.Tls != nil {
		return *x.Tls
	}
	return 0
}

func (x *ThreadInfoAarch64) GetGpregs() *UserAarch64RegsEntry {
	if x != nil {
		return x.Gpregs
	}
	return nil
}

func (x *ThreadInfoAarch64) GetFpsimd() *UserAarch64FpsimdContextEntry {
	if x != nil {
		return x.Fpsimd
	}
	return nil
}

var File_core_aarch64_proto protoreflect.FileDescriptor

var file_core_aarch64_proto_rawDesc = []byte{
	0x0a, 0x12, 0x63, 0x6f, 0x72, 0x65, 0x2d, 0x61, 0x61, 0x72, 0x63, 0x68, 0x36, 0x34, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0a, 0x6f, 0x70, 0x74, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0x65, 0x0a, 0x17, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x61, 0x61, 0x72, 0x63, 0x68, 0x36, 0x34,
	0x5f, 0x72, 0x65, 0x67, 0x73, 0x5f, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x12, 0x0a, 0x04, 0x72,
	0x65, 0x67, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x04, 0x52, 0x04, 0x72, 0x65, 0x67, 0x73, 0x12,
	0x0e, 0x0a, 0x02, 0x73, 0x70, 0x18, 0x02, 0x20, 0x02, 0x28, 0x04, 0x52, 0x02, 0x73, 0x70, 0x12,
	0x0e, 0x0a, 0x02, 0x70, 0x63, 0x18, 0x03, 0x20, 0x02, 0x28, 0x04, 0x52, 0x02, 0x70, 0x63, 0x12,
	0x16, 0x0a, 0x06, 0x70, 0x73, 0x74, 0x61, 0x74, 0x65, 0x18, 0x04, 0x20, 0x02, 0x28, 0x04, 0x52,
	0x06, 0x70, 0x73, 0x74, 0x61, 0x74, 0x65, 0x22, 0x61, 0x0a, 0x21, 0x75, 0x73, 0x65, 0x72, 0x5f,
	0x61, 0x61, 0x72, 0x63, 0x68, 0x36, 0x34, 0x5f, 0x66, 0x70, 0x73, 0x69, 0x6d, 0x64, 0x5f, 0x63,
	0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x5f, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x14, 0x0a, 0x05,
	0x76, 0x72, 0x65, 0x67, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x04, 0x52, 0x05, 0x76, 0x72, 0x65,
	0x67, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x66, 0x70, 0x73, 0x72, 0x18, 0x02, 0x20, 0x02, 0x28, 0x0d,
	0x52, 0x04, 0x66, 0x70, 0x73, 0x72, 0x12, 0x12, 0x0a, 0x04, 0x66, 0x70, 0x63, 0x72, 0x18, 0x03,
	0x20, 0x02, 0x28, 0x0d, 0x52, 0x04, 0x66, 0x70, 0x63, 0x72, 0x22, 0xc9, 0x01, 0x0a, 0x13, 0x74,
	0x68, 0x72, 0x65, 0x61, 0x64, 0x5f, 0x69, 0x6e, 0x66, 0x6f, 0x5f, 0x61, 0x61, 0x72, 0x63, 0x68,
	0x36, 0x34, 0x12, 0x2b, 0x0a, 0x0e, 0x63, 0x6c, 0x65, 0x61, 0x72, 0x5f, 0x74, 0x69, 0x64, 0x5f,
	0x61, 0x64, 0x64, 0x72, 0x18, 0x01, 0x20, 0x02, 0x28, 0x04, 0x42, 0x05, 0xd2, 0x3f, 0x02, 0x08,
	0x01, 0x52, 0x0c, 0x63, 0x6c, 0x65, 0x61, 0x72, 0x54, 0x69, 0x64, 0x41, 0x64, 0x64, 0x72, 0x12,
	0x10, 0x0a, 0x03, 0x74, 0x6c, 0x73, 0x18, 0x02, 0x20, 0x02, 0x28, 0x04, 0x52, 0x03, 0x74, 0x6c,
	0x73, 0x12, 0x37, 0x0a, 0x06, 0x67, 0x70, 0x72, 0x65, 0x67, 0x73, 0x18, 0x03, 0x20, 0x02, 0x28,
	0x0b, 0x32, 0x18, 0x2e, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x61, 0x61, 0x72, 0x63, 0x68, 0x36, 0x34,
	0x5f, 0x72, 0x65, 0x67, 0x73, 0x5f, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x42, 0x05, 0xd2, 0x3f, 0x02,
	0x08, 0x01, 0x52, 0x06, 0x67, 0x70, 0x72, 0x65, 0x67, 0x73, 0x12, 0x3a, 0x0a, 0x06, 0x66, 0x70,
	0x73, 0x69, 0x6d, 0x64, 0x18, 0x04, 0x20, 0x02, 0x28, 0x0b, 0x32, 0x22, 0x2e, 0x75, 0x73, 0x65,
	0x72, 0x5f, 0x61, 0x61, 0x72, 0x63, 0x68, 0x36, 0x34, 0x5f, 0x66, 0x70, 0x73, 0x69, 0x6d, 0x64,
	0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x5f, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x06,
	0x66, 0x70, 0x73, 0x69, 0x6d, 0x64,
}

var (
	file_core_aarch64_proto_rawDescOnce sync.Once
	file_core_aarch64_proto_rawDescData = file_core_aarch64_proto_rawDesc
)

func file_core_aarch64_proto_rawDescGZIP() []byte {
	file_core_aarch64_proto_rawDescOnce.Do(func() {
		file_core_aarch64_proto_rawDescData = protoimpl.X.CompressGZIP(file_core_aarch64_proto_rawDescData)
	})
	return file_core_aarch64_proto_rawDescData
}

var file_core_aarch64_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_core_aarch64_proto_goTypes = []interface{}{
	(*UserAarch64RegsEntry)(nil),          // 0: user_aarch64_regs_entry
	(*UserAarch64FpsimdContextEntry)(nil), // 1: user_aarch64_fpsimd_context_entry
	(*ThreadInfoAarch64)(nil),             // 2: thread_info_aarch64
}
var file_core_aarch64_proto_depIdxs = []int32{
	0, // 0: thread_info_aarch64.gpregs:type_name -> user_aarch64_regs_entry
	1, // 1: thread_info_aarch64.fpsimd:type_name -> user_aarch64_fpsimd_context_entry
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_core_aarch64_proto_init() }
func file_core_aarch64_proto_init() {
	if File_core_aarch64_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_core_aarch64_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UserAarch64RegsEntry); i {
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
		file_core_aarch64_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UserAarch64FpsimdContextEntry); i {
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
		file_core_aarch64_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ThreadInfoAarch64); i {
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
			RawDescriptor: file_core_aarch64_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_core_aarch64_proto_goTypes,
		DependencyIndexes: file_core_aarch64_proto_depIdxs,
		MessageInfos:      file_core_aarch64_proto_msgTypes,
	}.Build()
	File_core_aarch64_proto = out.File
	file_core_aarch64_proto_rawDesc = nil
	file_core_aarch64_proto_goTypes = nil
	file_core_aarch64_proto_depIdxs = nil
}
