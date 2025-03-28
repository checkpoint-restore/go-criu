// SPDX-License-Identifier: MIT

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.19.6
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

type PacAddressKeys struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ApiakeyLo     *uint64 `protobuf:"varint,1,req,name=apiakey_lo,json=apiakeyLo" json:"apiakey_lo,omitempty"`
	ApiakeyHi     *uint64 `protobuf:"varint,2,req,name=apiakey_hi,json=apiakeyHi" json:"apiakey_hi,omitempty"`
	ApibkeyLo     *uint64 `protobuf:"varint,3,req,name=apibkey_lo,json=apibkeyLo" json:"apibkey_lo,omitempty"`
	ApibkeyHi     *uint64 `protobuf:"varint,4,req,name=apibkey_hi,json=apibkeyHi" json:"apibkey_hi,omitempty"`
	ApdakeyLo     *uint64 `protobuf:"varint,5,req,name=apdakey_lo,json=apdakeyLo" json:"apdakey_lo,omitempty"`
	ApdakeyHi     *uint64 `protobuf:"varint,6,req,name=apdakey_hi,json=apdakeyHi" json:"apdakey_hi,omitempty"`
	ApdbkeyLo     *uint64 `protobuf:"varint,7,req,name=apdbkey_lo,json=apdbkeyLo" json:"apdbkey_lo,omitempty"`
	ApdbkeyHi     *uint64 `protobuf:"varint,8,req,name=apdbkey_hi,json=apdbkeyHi" json:"apdbkey_hi,omitempty"`
	PacEnabledKey *uint64 `protobuf:"varint,9,req,name=pac_enabled_key,json=pacEnabledKey" json:"pac_enabled_key,omitempty"`
}

func (x *PacAddressKeys) Reset() {
	*x = PacAddressKeys{}
	if protoimpl.UnsafeEnabled {
		mi := &file_core_aarch64_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PacAddressKeys) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PacAddressKeys) ProtoMessage() {}

func (x *PacAddressKeys) ProtoReflect() protoreflect.Message {
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

// Deprecated: Use PacAddressKeys.ProtoReflect.Descriptor instead.
func (*PacAddressKeys) Descriptor() ([]byte, []int) {
	return file_core_aarch64_proto_rawDescGZIP(), []int{2}
}

func (x *PacAddressKeys) GetApiakeyLo() uint64 {
	if x != nil && x.ApiakeyLo != nil {
		return *x.ApiakeyLo
	}
	return 0
}

func (x *PacAddressKeys) GetApiakeyHi() uint64 {
	if x != nil && x.ApiakeyHi != nil {
		return *x.ApiakeyHi
	}
	return 0
}

func (x *PacAddressKeys) GetApibkeyLo() uint64 {
	if x != nil && x.ApibkeyLo != nil {
		return *x.ApibkeyLo
	}
	return 0
}

func (x *PacAddressKeys) GetApibkeyHi() uint64 {
	if x != nil && x.ApibkeyHi != nil {
		return *x.ApibkeyHi
	}
	return 0
}

func (x *PacAddressKeys) GetApdakeyLo() uint64 {
	if x != nil && x.ApdakeyLo != nil {
		return *x.ApdakeyLo
	}
	return 0
}

func (x *PacAddressKeys) GetApdakeyHi() uint64 {
	if x != nil && x.ApdakeyHi != nil {
		return *x.ApdakeyHi
	}
	return 0
}

func (x *PacAddressKeys) GetApdbkeyLo() uint64 {
	if x != nil && x.ApdbkeyLo != nil {
		return *x.ApdbkeyLo
	}
	return 0
}

func (x *PacAddressKeys) GetApdbkeyHi() uint64 {
	if x != nil && x.ApdbkeyHi != nil {
		return *x.ApdbkeyHi
	}
	return 0
}

func (x *PacAddressKeys) GetPacEnabledKey() uint64 {
	if x != nil && x.PacEnabledKey != nil {
		return *x.PacEnabledKey
	}
	return 0
}

type PacGenericKeys struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ApgakeyLo *uint64 `protobuf:"varint,1,req,name=apgakey_lo,json=apgakeyLo" json:"apgakey_lo,omitempty"`
	ApgakeyHi *uint64 `protobuf:"varint,2,req,name=apgakey_hi,json=apgakeyHi" json:"apgakey_hi,omitempty"`
}

func (x *PacGenericKeys) Reset() {
	*x = PacGenericKeys{}
	if protoimpl.UnsafeEnabled {
		mi := &file_core_aarch64_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PacGenericKeys) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PacGenericKeys) ProtoMessage() {}

func (x *PacGenericKeys) ProtoReflect() protoreflect.Message {
	mi := &file_core_aarch64_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PacGenericKeys.ProtoReflect.Descriptor instead.
func (*PacGenericKeys) Descriptor() ([]byte, []int) {
	return file_core_aarch64_proto_rawDescGZIP(), []int{3}
}

func (x *PacGenericKeys) GetApgakeyLo() uint64 {
	if x != nil && x.ApgakeyLo != nil {
		return *x.ApgakeyLo
	}
	return 0
}

func (x *PacGenericKeys) GetApgakeyHi() uint64 {
	if x != nil && x.ApgakeyHi != nil {
		return *x.ApgakeyHi
	}
	return 0
}

type PacKeys struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PacAddressKeys *PacAddressKeys `protobuf:"bytes,6,opt,name=pac_address_keys,json=pacAddressKeys" json:"pac_address_keys,omitempty"`
	PacGenericKeys *PacGenericKeys `protobuf:"bytes,7,opt,name=pac_generic_keys,json=pacGenericKeys" json:"pac_generic_keys,omitempty"`
}

func (x *PacKeys) Reset() {
	*x = PacKeys{}
	if protoimpl.UnsafeEnabled {
		mi := &file_core_aarch64_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PacKeys) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PacKeys) ProtoMessage() {}

func (x *PacKeys) ProtoReflect() protoreflect.Message {
	mi := &file_core_aarch64_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PacKeys.ProtoReflect.Descriptor instead.
func (*PacKeys) Descriptor() ([]byte, []int) {
	return file_core_aarch64_proto_rawDescGZIP(), []int{4}
}

func (x *PacKeys) GetPacAddressKeys() *PacAddressKeys {
	if x != nil {
		return x.PacAddressKeys
	}
	return nil
}

func (x *PacKeys) GetPacGenericKeys() *PacGenericKeys {
	if x != nil {
		return x.PacGenericKeys
	}
	return nil
}

type ThreadInfoAarch64 struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ClearTidAddr *uint64                        `protobuf:"varint,1,req,name=clear_tid_addr,json=clearTidAddr" json:"clear_tid_addr,omitempty"`
	Tls          *uint64                        `protobuf:"varint,2,req,name=tls" json:"tls,omitempty"`
	Gpregs       *UserAarch64RegsEntry          `protobuf:"bytes,3,req,name=gpregs" json:"gpregs,omitempty"`
	Fpsimd       *UserAarch64FpsimdContextEntry `protobuf:"bytes,4,req,name=fpsimd" json:"fpsimd,omitempty"`
	PacKeys      *PacKeys                       `protobuf:"bytes,5,opt,name=pac_keys,json=pacKeys" json:"pac_keys,omitempty"`
}

func (x *ThreadInfoAarch64) Reset() {
	*x = ThreadInfoAarch64{}
	if protoimpl.UnsafeEnabled {
		mi := &file_core_aarch64_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ThreadInfoAarch64) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ThreadInfoAarch64) ProtoMessage() {}

func (x *ThreadInfoAarch64) ProtoReflect() protoreflect.Message {
	mi := &file_core_aarch64_proto_msgTypes[5]
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
	return file_core_aarch64_proto_rawDescGZIP(), []int{5}
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

func (x *ThreadInfoAarch64) GetPacKeys() *PacKeys {
	if x != nil {
		return x.PacKeys
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
	0x20, 0x02, 0x28, 0x0d, 0x52, 0x04, 0x66, 0x70, 0x63, 0x72, 0x22, 0xb2, 0x02, 0x0a, 0x10, 0x70,
	0x61, 0x63, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x5f, 0x6b, 0x65, 0x79, 0x73, 0x12,
	0x1d, 0x0a, 0x0a, 0x61, 0x70, 0x69, 0x61, 0x6b, 0x65, 0x79, 0x5f, 0x6c, 0x6f, 0x18, 0x01, 0x20,
	0x02, 0x28, 0x04, 0x52, 0x09, 0x61, 0x70, 0x69, 0x61, 0x6b, 0x65, 0x79, 0x4c, 0x6f, 0x12, 0x1d,
	0x0a, 0x0a, 0x61, 0x70, 0x69, 0x61, 0x6b, 0x65, 0x79, 0x5f, 0x68, 0x69, 0x18, 0x02, 0x20, 0x02,
	0x28, 0x04, 0x52, 0x09, 0x61, 0x70, 0x69, 0x61, 0x6b, 0x65, 0x79, 0x48, 0x69, 0x12, 0x1d, 0x0a,
	0x0a, 0x61, 0x70, 0x69, 0x62, 0x6b, 0x65, 0x79, 0x5f, 0x6c, 0x6f, 0x18, 0x03, 0x20, 0x02, 0x28,
	0x04, 0x52, 0x09, 0x61, 0x70, 0x69, 0x62, 0x6b, 0x65, 0x79, 0x4c, 0x6f, 0x12, 0x1d, 0x0a, 0x0a,
	0x61, 0x70, 0x69, 0x62, 0x6b, 0x65, 0x79, 0x5f, 0x68, 0x69, 0x18, 0x04, 0x20, 0x02, 0x28, 0x04,
	0x52, 0x09, 0x61, 0x70, 0x69, 0x62, 0x6b, 0x65, 0x79, 0x48, 0x69, 0x12, 0x1d, 0x0a, 0x0a, 0x61,
	0x70, 0x64, 0x61, 0x6b, 0x65, 0x79, 0x5f, 0x6c, 0x6f, 0x18, 0x05, 0x20, 0x02, 0x28, 0x04, 0x52,
	0x09, 0x61, 0x70, 0x64, 0x61, 0x6b, 0x65, 0x79, 0x4c, 0x6f, 0x12, 0x1d, 0x0a, 0x0a, 0x61, 0x70,
	0x64, 0x61, 0x6b, 0x65, 0x79, 0x5f, 0x68, 0x69, 0x18, 0x06, 0x20, 0x02, 0x28, 0x04, 0x52, 0x09,
	0x61, 0x70, 0x64, 0x61, 0x6b, 0x65, 0x79, 0x48, 0x69, 0x12, 0x1d, 0x0a, 0x0a, 0x61, 0x70, 0x64,
	0x62, 0x6b, 0x65, 0x79, 0x5f, 0x6c, 0x6f, 0x18, 0x07, 0x20, 0x02, 0x28, 0x04, 0x52, 0x09, 0x61,
	0x70, 0x64, 0x62, 0x6b, 0x65, 0x79, 0x4c, 0x6f, 0x12, 0x1d, 0x0a, 0x0a, 0x61, 0x70, 0x64, 0x62,
	0x6b, 0x65, 0x79, 0x5f, 0x68, 0x69, 0x18, 0x08, 0x20, 0x02, 0x28, 0x04, 0x52, 0x09, 0x61, 0x70,
	0x64, 0x62, 0x6b, 0x65, 0x79, 0x48, 0x69, 0x12, 0x26, 0x0a, 0x0f, 0x70, 0x61, 0x63, 0x5f, 0x65,
	0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x09, 0x20, 0x02, 0x28, 0x04,
	0x52, 0x0d, 0x70, 0x61, 0x63, 0x45, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, 0x4b, 0x65, 0x79, 0x22,
	0x50, 0x0a, 0x10, 0x70, 0x61, 0x63, 0x5f, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x69, 0x63, 0x5f, 0x6b,
	0x65, 0x79, 0x73, 0x12, 0x1d, 0x0a, 0x0a, 0x61, 0x70, 0x67, 0x61, 0x6b, 0x65, 0x79, 0x5f, 0x6c,
	0x6f, 0x18, 0x01, 0x20, 0x02, 0x28, 0x04, 0x52, 0x09, 0x61, 0x70, 0x67, 0x61, 0x6b, 0x65, 0x79,
	0x4c, 0x6f, 0x12, 0x1d, 0x0a, 0x0a, 0x61, 0x70, 0x67, 0x61, 0x6b, 0x65, 0x79, 0x5f, 0x68, 0x69,
	0x18, 0x02, 0x20, 0x02, 0x28, 0x04, 0x52, 0x09, 0x61, 0x70, 0x67, 0x61, 0x6b, 0x65, 0x79, 0x48,
	0x69, 0x22, 0x84, 0x01, 0x0a, 0x08, 0x70, 0x61, 0x63, 0x5f, 0x6b, 0x65, 0x79, 0x73, 0x12, 0x3b,
	0x0a, 0x10, 0x70, 0x61, 0x63, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x5f, 0x6b, 0x65,
	0x79, 0x73, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x70, 0x61, 0x63, 0x5f, 0x61,
	0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x5f, 0x6b, 0x65, 0x79, 0x73, 0x52, 0x0e, 0x70, 0x61, 0x63,
	0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x4b, 0x65, 0x79, 0x73, 0x12, 0x3b, 0x0a, 0x10, 0x70,
	0x61, 0x63, 0x5f, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x69, 0x63, 0x5f, 0x6b, 0x65, 0x79, 0x73, 0x18,
	0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x70, 0x61, 0x63, 0x5f, 0x67, 0x65, 0x6e, 0x65,
	0x72, 0x69, 0x63, 0x5f, 0x6b, 0x65, 0x79, 0x73, 0x52, 0x0e, 0x70, 0x61, 0x63, 0x47, 0x65, 0x6e,
	0x65, 0x72, 0x69, 0x63, 0x4b, 0x65, 0x79, 0x73, 0x22, 0xef, 0x01, 0x0a, 0x13, 0x74, 0x68, 0x72,
	0x65, 0x61, 0x64, 0x5f, 0x69, 0x6e, 0x66, 0x6f, 0x5f, 0x61, 0x61, 0x72, 0x63, 0x68, 0x36, 0x34,
	0x12, 0x2b, 0x0a, 0x0e, 0x63, 0x6c, 0x65, 0x61, 0x72, 0x5f, 0x74, 0x69, 0x64, 0x5f, 0x61, 0x64,
	0x64, 0x72, 0x18, 0x01, 0x20, 0x02, 0x28, 0x04, 0x42, 0x05, 0xd2, 0x3f, 0x02, 0x08, 0x01, 0x52,
	0x0c, 0x63, 0x6c, 0x65, 0x61, 0x72, 0x54, 0x69, 0x64, 0x41, 0x64, 0x64, 0x72, 0x12, 0x10, 0x0a,
	0x03, 0x74, 0x6c, 0x73, 0x18, 0x02, 0x20, 0x02, 0x28, 0x04, 0x52, 0x03, 0x74, 0x6c, 0x73, 0x12,
	0x37, 0x0a, 0x06, 0x67, 0x70, 0x72, 0x65, 0x67, 0x73, 0x18, 0x03, 0x20, 0x02, 0x28, 0x0b, 0x32,
	0x18, 0x2e, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x61, 0x61, 0x72, 0x63, 0x68, 0x36, 0x34, 0x5f, 0x72,
	0x65, 0x67, 0x73, 0x5f, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x42, 0x05, 0xd2, 0x3f, 0x02, 0x08, 0x01,
	0x52, 0x06, 0x67, 0x70, 0x72, 0x65, 0x67, 0x73, 0x12, 0x3a, 0x0a, 0x06, 0x66, 0x70, 0x73, 0x69,
	0x6d, 0x64, 0x18, 0x04, 0x20, 0x02, 0x28, 0x0b, 0x32, 0x22, 0x2e, 0x75, 0x73, 0x65, 0x72, 0x5f,
	0x61, 0x61, 0x72, 0x63, 0x68, 0x36, 0x34, 0x5f, 0x66, 0x70, 0x73, 0x69, 0x6d, 0x64, 0x5f, 0x63,
	0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x5f, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x06, 0x66, 0x70,
	0x73, 0x69, 0x6d, 0x64, 0x12, 0x24, 0x0a, 0x08, 0x70, 0x61, 0x63, 0x5f, 0x6b, 0x65, 0x79, 0x73,
	0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x09, 0x2e, 0x70, 0x61, 0x63, 0x5f, 0x6b, 0x65, 0x79,
	0x73, 0x52, 0x07, 0x70, 0x61, 0x63, 0x4b, 0x65, 0x79, 0x73,
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

var file_core_aarch64_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_core_aarch64_proto_goTypes = []interface{}{
	(*UserAarch64RegsEntry)(nil),          // 0: user_aarch64_regs_entry
	(*UserAarch64FpsimdContextEntry)(nil), // 1: user_aarch64_fpsimd_context_entry
	(*PacAddressKeys)(nil),                // 2: pac_address_keys
	(*PacGenericKeys)(nil),                // 3: pac_generic_keys
	(*PacKeys)(nil),                       // 4: pac_keys
	(*ThreadInfoAarch64)(nil),             // 5: thread_info_aarch64
}
var file_core_aarch64_proto_depIdxs = []int32{
	2, // 0: pac_keys.pac_address_keys:type_name -> pac_address_keys
	3, // 1: pac_keys.pac_generic_keys:type_name -> pac_generic_keys
	0, // 2: thread_info_aarch64.gpregs:type_name -> user_aarch64_regs_entry
	1, // 3: thread_info_aarch64.fpsimd:type_name -> user_aarch64_fpsimd_context_entry
	4, // 4: thread_info_aarch64.pac_keys:type_name -> pac_keys
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
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
			switch v := v.(*PacAddressKeys); i {
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
		file_core_aarch64_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PacGenericKeys); i {
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
		file_core_aarch64_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PacKeys); i {
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
		file_core_aarch64_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
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
			NumMessages:   6,
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
