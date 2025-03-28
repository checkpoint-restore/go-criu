// SPDX-License-Identifier: MIT

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.19.6
// source: cpuinfo.proto

package cpuinfo

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

type CpuinfoX86EntryVendor int32

const (
	CpuinfoX86Entry_UNKNOWN CpuinfoX86EntryVendor = 0
	CpuinfoX86Entry_INTEL   CpuinfoX86EntryVendor = 1
	CpuinfoX86Entry_AMD     CpuinfoX86EntryVendor = 2
)

// Enum value maps for CpuinfoX86EntryVendor.
var (
	CpuinfoX86EntryVendor_name = map[int32]string{
		0: "UNKNOWN",
		1: "INTEL",
		2: "AMD",
	}
	CpuinfoX86EntryVendor_value = map[string]int32{
		"UNKNOWN": 0,
		"INTEL":   1,
		"AMD":     2,
	}
)

func (x CpuinfoX86EntryVendor) Enum() *CpuinfoX86EntryVendor {
	p := new(CpuinfoX86EntryVendor)
	*p = x
	return p
}

func (x CpuinfoX86EntryVendor) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (CpuinfoX86EntryVendor) Descriptor() protoreflect.EnumDescriptor {
	return file_cpuinfo_proto_enumTypes[0].Descriptor()
}

func (CpuinfoX86EntryVendor) Type() protoreflect.EnumType {
	return &file_cpuinfo_proto_enumTypes[0]
}

func (x CpuinfoX86EntryVendor) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Do not use.
func (x *CpuinfoX86EntryVendor) UnmarshalJSON(b []byte) error {
	num, err := protoimpl.X.UnmarshalJSONEnum(x.Descriptor(), b)
	if err != nil {
		return err
	}
	*x = CpuinfoX86EntryVendor(num)
	return nil
}

// Deprecated: Use CpuinfoX86EntryVendor.Descriptor instead.
func (CpuinfoX86EntryVendor) EnumDescriptor() ([]byte, []int) {
	return file_cpuinfo_proto_rawDescGZIP(), []int{0, 0}
}

type CpuinfoPpc64EntryEndianness int32

const (
	CpuinfoPpc64Entry_BIGENDIAN    CpuinfoPpc64EntryEndianness = 0
	CpuinfoPpc64Entry_LITTLEENDIAN CpuinfoPpc64EntryEndianness = 1
)

// Enum value maps for CpuinfoPpc64EntryEndianness.
var (
	CpuinfoPpc64EntryEndianness_name = map[int32]string{
		0: "BIGENDIAN",
		1: "LITTLEENDIAN",
	}
	CpuinfoPpc64EntryEndianness_value = map[string]int32{
		"BIGENDIAN":    0,
		"LITTLEENDIAN": 1,
	}
)

func (x CpuinfoPpc64EntryEndianness) Enum() *CpuinfoPpc64EntryEndianness {
	p := new(CpuinfoPpc64EntryEndianness)
	*p = x
	return p
}

func (x CpuinfoPpc64EntryEndianness) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (CpuinfoPpc64EntryEndianness) Descriptor() protoreflect.EnumDescriptor {
	return file_cpuinfo_proto_enumTypes[1].Descriptor()
}

func (CpuinfoPpc64EntryEndianness) Type() protoreflect.EnumType {
	return &file_cpuinfo_proto_enumTypes[1]
}

func (x CpuinfoPpc64EntryEndianness) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Do not use.
func (x *CpuinfoPpc64EntryEndianness) UnmarshalJSON(b []byte) error {
	num, err := protoimpl.X.UnmarshalJSONEnum(x.Descriptor(), b)
	if err != nil {
		return err
	}
	*x = CpuinfoPpc64EntryEndianness(num)
	return nil
}

// Deprecated: Use CpuinfoPpc64EntryEndianness.Descriptor instead.
func (CpuinfoPpc64EntryEndianness) EnumDescriptor() ([]byte, []int) {
	return file_cpuinfo_proto_rawDescGZIP(), []int{1, 0}
}

type CpuinfoX86Entry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	VendorId      *CpuinfoX86EntryVendor `protobuf:"varint,1,req,name=vendor_id,json=vendorId,enum=CpuinfoX86EntryVendor" json:"vendor_id,omitempty"`
	CpuFamily     *uint32                `protobuf:"varint,2,req,name=cpu_family,json=cpuFamily" json:"cpu_family,omitempty"`
	Model         *uint32                `protobuf:"varint,3,req,name=model" json:"model,omitempty"`
	Stepping      *uint32                `protobuf:"varint,4,req,name=stepping" json:"stepping,omitempty"`
	CapabilityVer *uint32                `protobuf:"varint,5,req,name=capability_ver,json=capabilityVer" json:"capability_ver,omitempty"`
	Capability    []uint32               `protobuf:"varint,6,rep,name=capability" json:"capability,omitempty"`
	ModelId       *string                `protobuf:"bytes,7,opt,name=model_id,json=modelId" json:"model_id,omitempty"`
	XfeaturesMask *uint64                `protobuf:"varint,8,opt,name=xfeatures_mask,json=xfeaturesMask" json:"xfeatures_mask,omitempty"`
	XsaveSize     *uint32                `protobuf:"varint,9,opt,name=xsave_size,json=xsaveSize" json:"xsave_size,omitempty"`
	XsaveSizeMax  *uint32                `protobuf:"varint,10,opt,name=xsave_size_max,json=xsaveSizeMax" json:"xsave_size_max,omitempty"`
}

func (x *CpuinfoX86Entry) Reset() {
	*x = CpuinfoX86Entry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cpuinfo_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CpuinfoX86Entry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CpuinfoX86Entry) ProtoMessage() {}

func (x *CpuinfoX86Entry) ProtoReflect() protoreflect.Message {
	mi := &file_cpuinfo_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CpuinfoX86Entry.ProtoReflect.Descriptor instead.
func (*CpuinfoX86Entry) Descriptor() ([]byte, []int) {
	return file_cpuinfo_proto_rawDescGZIP(), []int{0}
}

func (x *CpuinfoX86Entry) GetVendorId() CpuinfoX86EntryVendor {
	if x != nil && x.VendorId != nil {
		return *x.VendorId
	}
	return CpuinfoX86Entry_UNKNOWN
}

func (x *CpuinfoX86Entry) GetCpuFamily() uint32 {
	if x != nil && x.CpuFamily != nil {
		return *x.CpuFamily
	}
	return 0
}

func (x *CpuinfoX86Entry) GetModel() uint32 {
	if x != nil && x.Model != nil {
		return *x.Model
	}
	return 0
}

func (x *CpuinfoX86Entry) GetStepping() uint32 {
	if x != nil && x.Stepping != nil {
		return *x.Stepping
	}
	return 0
}

func (x *CpuinfoX86Entry) GetCapabilityVer() uint32 {
	if x != nil && x.CapabilityVer != nil {
		return *x.CapabilityVer
	}
	return 0
}

func (x *CpuinfoX86Entry) GetCapability() []uint32 {
	if x != nil {
		return x.Capability
	}
	return nil
}

func (x *CpuinfoX86Entry) GetModelId() string {
	if x != nil && x.ModelId != nil {
		return *x.ModelId
	}
	return ""
}

func (x *CpuinfoX86Entry) GetXfeaturesMask() uint64 {
	if x != nil && x.XfeaturesMask != nil {
		return *x.XfeaturesMask
	}
	return 0
}

func (x *CpuinfoX86Entry) GetXsaveSize() uint32 {
	if x != nil && x.XsaveSize != nil {
		return *x.XsaveSize
	}
	return 0
}

func (x *CpuinfoX86Entry) GetXsaveSizeMax() uint32 {
	if x != nil && x.XsaveSizeMax != nil {
		return *x.XsaveSizeMax
	}
	return 0
}

type CpuinfoPpc64Entry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Endian *CpuinfoPpc64EntryEndianness `protobuf:"varint,1,req,name=endian,enum=CpuinfoPpc64EntryEndianness" json:"endian,omitempty"`
	Hwcap  []uint64                     `protobuf:"varint,2,rep,name=hwcap" json:"hwcap,omitempty"`
}

func (x *CpuinfoPpc64Entry) Reset() {
	*x = CpuinfoPpc64Entry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cpuinfo_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CpuinfoPpc64Entry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CpuinfoPpc64Entry) ProtoMessage() {}

func (x *CpuinfoPpc64Entry) ProtoReflect() protoreflect.Message {
	mi := &file_cpuinfo_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CpuinfoPpc64Entry.ProtoReflect.Descriptor instead.
func (*CpuinfoPpc64Entry) Descriptor() ([]byte, []int) {
	return file_cpuinfo_proto_rawDescGZIP(), []int{1}
}

func (x *CpuinfoPpc64Entry) GetEndian() CpuinfoPpc64EntryEndianness {
	if x != nil && x.Endian != nil {
		return *x.Endian
	}
	return CpuinfoPpc64Entry_BIGENDIAN
}

func (x *CpuinfoPpc64Entry) GetHwcap() []uint64 {
	if x != nil {
		return x.Hwcap
	}
	return nil
}

type CpuinfoS390Entry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Hwcap []uint64 `protobuf:"varint,2,rep,name=hwcap" json:"hwcap,omitempty"`
}

func (x *CpuinfoS390Entry) Reset() {
	*x = CpuinfoS390Entry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cpuinfo_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CpuinfoS390Entry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CpuinfoS390Entry) ProtoMessage() {}

func (x *CpuinfoS390Entry) ProtoReflect() protoreflect.Message {
	mi := &file_cpuinfo_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CpuinfoS390Entry.ProtoReflect.Descriptor instead.
func (*CpuinfoS390Entry) Descriptor() ([]byte, []int) {
	return file_cpuinfo_proto_rawDescGZIP(), []int{2}
}

func (x *CpuinfoS390Entry) GetHwcap() []uint64 {
	if x != nil {
		return x.Hwcap
	}
	return nil
}

type CpuinfoEntry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Usually on SMP system there should be same CPUs
	// installed, but it might happen that system carries
	// various CPUs so @repeated used.
	X86Entry   []*CpuinfoX86Entry   `protobuf:"bytes,1,rep,name=x86_entry,json=x86Entry" json:"x86_entry,omitempty"`
	Ppc64Entry []*CpuinfoPpc64Entry `protobuf:"bytes,2,rep,name=ppc64_entry,json=ppc64Entry" json:"ppc64_entry,omitempty"`
	S390Entry  []*CpuinfoS390Entry  `protobuf:"bytes,3,rep,name=s390_entry,json=s390Entry" json:"s390_entry,omitempty"`
}

func (x *CpuinfoEntry) Reset() {
	*x = CpuinfoEntry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cpuinfo_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CpuinfoEntry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CpuinfoEntry) ProtoMessage() {}

func (x *CpuinfoEntry) ProtoReflect() protoreflect.Message {
	mi := &file_cpuinfo_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CpuinfoEntry.ProtoReflect.Descriptor instead.
func (*CpuinfoEntry) Descriptor() ([]byte, []int) {
	return file_cpuinfo_proto_rawDescGZIP(), []int{3}
}

func (x *CpuinfoEntry) GetX86Entry() []*CpuinfoX86Entry {
	if x != nil {
		return x.X86Entry
	}
	return nil
}

func (x *CpuinfoEntry) GetPpc64Entry() []*CpuinfoPpc64Entry {
	if x != nil {
		return x.Ppc64Entry
	}
	return nil
}

func (x *CpuinfoEntry) GetS390Entry() []*CpuinfoS390Entry {
	if x != nil {
		return x.S390Entry
	}
	return nil
}

var File_cpuinfo_proto protoreflect.FileDescriptor

var file_cpuinfo_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x63, 0x70, 0x75, 0x69, 0x6e, 0x66, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22,
	0x95, 0x03, 0x0a, 0x11, 0x63, 0x70, 0x75, 0x69, 0x6e, 0x66, 0x6f, 0x5f, 0x78, 0x38, 0x36, 0x5f,
	0x65, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x36, 0x0a, 0x09, 0x76, 0x65, 0x6e, 0x64, 0x6f, 0x72, 0x5f,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x02, 0x28, 0x0e, 0x32, 0x19, 0x2e, 0x63, 0x70, 0x75, 0x69, 0x6e,
	0x66, 0x6f, 0x5f, 0x78, 0x38, 0x36, 0x5f, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x2e, 0x76, 0x65, 0x6e,
	0x64, 0x6f, 0x72, 0x52, 0x08, 0x76, 0x65, 0x6e, 0x64, 0x6f, 0x72, 0x49, 0x64, 0x12, 0x1d, 0x0a,
	0x0a, 0x63, 0x70, 0x75, 0x5f, 0x66, 0x61, 0x6d, 0x69, 0x6c, 0x79, 0x18, 0x02, 0x20, 0x02, 0x28,
	0x0d, 0x52, 0x09, 0x63, 0x70, 0x75, 0x46, 0x61, 0x6d, 0x69, 0x6c, 0x79, 0x12, 0x14, 0x0a, 0x05,
	0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x18, 0x03, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x05, 0x6d, 0x6f, 0x64,
	0x65, 0x6c, 0x12, 0x1a, 0x0a, 0x08, 0x73, 0x74, 0x65, 0x70, 0x70, 0x69, 0x6e, 0x67, 0x18, 0x04,
	0x20, 0x02, 0x28, 0x0d, 0x52, 0x08, 0x73, 0x74, 0x65, 0x70, 0x70, 0x69, 0x6e, 0x67, 0x12, 0x25,
	0x0a, 0x0e, 0x63, 0x61, 0x70, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x79, 0x5f, 0x76, 0x65, 0x72,
	0x18, 0x05, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x0d, 0x63, 0x61, 0x70, 0x61, 0x62, 0x69, 0x6c, 0x69,
	0x74, 0x79, 0x56, 0x65, 0x72, 0x12, 0x1e, 0x0a, 0x0a, 0x63, 0x61, 0x70, 0x61, 0x62, 0x69, 0x6c,
	0x69, 0x74, 0x79, 0x18, 0x06, 0x20, 0x03, 0x28, 0x0d, 0x52, 0x0a, 0x63, 0x61, 0x70, 0x61, 0x62,
	0x69, 0x6c, 0x69, 0x74, 0x79, 0x12, 0x19, 0x0a, 0x08, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x5f, 0x69,
	0x64, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x49, 0x64,
	0x12, 0x25, 0x0a, 0x0e, 0x78, 0x66, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x73, 0x5f, 0x6d, 0x61,
	0x73, 0x6b, 0x18, 0x08, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0d, 0x78, 0x66, 0x65, 0x61, 0x74, 0x75,
	0x72, 0x65, 0x73, 0x4d, 0x61, 0x73, 0x6b, 0x12, 0x1d, 0x0a, 0x0a, 0x78, 0x73, 0x61, 0x76, 0x65,
	0x5f, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x09, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x09, 0x78, 0x73, 0x61,
	0x76, 0x65, 0x53, 0x69, 0x7a, 0x65, 0x12, 0x24, 0x0a, 0x0e, 0x78, 0x73, 0x61, 0x76, 0x65, 0x5f,
	0x73, 0x69, 0x7a, 0x65, 0x5f, 0x6d, 0x61, 0x78, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x0c,
	0x78, 0x73, 0x61, 0x76, 0x65, 0x53, 0x69, 0x7a, 0x65, 0x4d, 0x61, 0x78, 0x22, 0x29, 0x0a, 0x06,
	0x76, 0x65, 0x6e, 0x64, 0x6f, 0x72, 0x12, 0x0b, 0x0a, 0x07, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57,
	0x4e, 0x10, 0x00, 0x12, 0x09, 0x0a, 0x05, 0x49, 0x4e, 0x54, 0x45, 0x4c, 0x10, 0x01, 0x12, 0x07,
	0x0a, 0x03, 0x41, 0x4d, 0x44, 0x10, 0x02, 0x22, 0x93, 0x01, 0x0a, 0x13, 0x63, 0x70, 0x75, 0x69,
	0x6e, 0x66, 0x6f, 0x5f, 0x70, 0x70, 0x63, 0x36, 0x34, 0x5f, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x12,
	0x37, 0x0a, 0x06, 0x65, 0x6e, 0x64, 0x69, 0x61, 0x6e, 0x18, 0x01, 0x20, 0x02, 0x28, 0x0e, 0x32,
	0x1f, 0x2e, 0x63, 0x70, 0x75, 0x69, 0x6e, 0x66, 0x6f, 0x5f, 0x70, 0x70, 0x63, 0x36, 0x34, 0x5f,
	0x65, 0x6e, 0x74, 0x72, 0x79, 0x2e, 0x65, 0x6e, 0x64, 0x69, 0x61, 0x6e, 0x6e, 0x65, 0x73, 0x73,
	0x52, 0x06, 0x65, 0x6e, 0x64, 0x69, 0x61, 0x6e, 0x12, 0x14, 0x0a, 0x05, 0x68, 0x77, 0x63, 0x61,
	0x70, 0x18, 0x02, 0x20, 0x03, 0x28, 0x04, 0x52, 0x05, 0x68, 0x77, 0x63, 0x61, 0x70, 0x22, 0x2d,
	0x0a, 0x0a, 0x65, 0x6e, 0x64, 0x69, 0x61, 0x6e, 0x6e, 0x65, 0x73, 0x73, 0x12, 0x0d, 0x0a, 0x09,
	0x42, 0x49, 0x47, 0x45, 0x4e, 0x44, 0x49, 0x41, 0x4e, 0x10, 0x00, 0x12, 0x10, 0x0a, 0x0c, 0x4c,
	0x49, 0x54, 0x54, 0x4c, 0x45, 0x45, 0x4e, 0x44, 0x49, 0x41, 0x4e, 0x10, 0x01, 0x22, 0x2a, 0x0a,
	0x12, 0x63, 0x70, 0x75, 0x69, 0x6e, 0x66, 0x6f, 0x5f, 0x73, 0x33, 0x39, 0x30, 0x5f, 0x65, 0x6e,
	0x74, 0x72, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x68, 0x77, 0x63, 0x61, 0x70, 0x18, 0x02, 0x20, 0x03,
	0x28, 0x04, 0x52, 0x05, 0x68, 0x77, 0x63, 0x61, 0x70, 0x22, 0xab, 0x01, 0x0a, 0x0d, 0x63, 0x70,
	0x75, 0x69, 0x6e, 0x66, 0x6f, 0x5f, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x2f, 0x0a, 0x09, 0x78,
	0x38, 0x36, 0x5f, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x12,
	0x2e, 0x63, 0x70, 0x75, 0x69, 0x6e, 0x66, 0x6f, 0x5f, 0x78, 0x38, 0x36, 0x5f, 0x65, 0x6e, 0x74,
	0x72, 0x79, 0x52, 0x08, 0x78, 0x38, 0x36, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x35, 0x0a, 0x0b,
	0x70, 0x70, 0x63, 0x36, 0x34, 0x5f, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x18, 0x02, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x14, 0x2e, 0x63, 0x70, 0x75, 0x69, 0x6e, 0x66, 0x6f, 0x5f, 0x70, 0x70, 0x63, 0x36,
	0x34, 0x5f, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x0a, 0x70, 0x70, 0x63, 0x36, 0x34, 0x45, 0x6e,
	0x74, 0x72, 0x79, 0x12, 0x32, 0x0a, 0x0a, 0x73, 0x33, 0x39, 0x30, 0x5f, 0x65, 0x6e, 0x74, 0x72,
	0x79, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x63, 0x70, 0x75, 0x69, 0x6e, 0x66,
	0x6f, 0x5f, 0x73, 0x33, 0x39, 0x30, 0x5f, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x09, 0x73, 0x33,
	0x39, 0x30, 0x45, 0x6e, 0x74, 0x72, 0x79,
}

var (
	file_cpuinfo_proto_rawDescOnce sync.Once
	file_cpuinfo_proto_rawDescData = file_cpuinfo_proto_rawDesc
)

func file_cpuinfo_proto_rawDescGZIP() []byte {
	file_cpuinfo_proto_rawDescOnce.Do(func() {
		file_cpuinfo_proto_rawDescData = protoimpl.X.CompressGZIP(file_cpuinfo_proto_rawDescData)
	})
	return file_cpuinfo_proto_rawDescData
}

var file_cpuinfo_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_cpuinfo_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_cpuinfo_proto_goTypes = []interface{}{
	(CpuinfoX86EntryVendor)(0),       // 0: cpuinfo_x86_entry.vendor
	(CpuinfoPpc64EntryEndianness)(0), // 1: cpuinfo_ppc64_entry.endianness
	(*CpuinfoX86Entry)(nil),          // 2: cpuinfo_x86_entry
	(*CpuinfoPpc64Entry)(nil),        // 3: cpuinfo_ppc64_entry
	(*CpuinfoS390Entry)(nil),         // 4: cpuinfo_s390_entry
	(*CpuinfoEntry)(nil),             // 5: cpuinfo_entry
}
var file_cpuinfo_proto_depIdxs = []int32{
	0, // 0: cpuinfo_x86_entry.vendor_id:type_name -> cpuinfo_x86_entry.vendor
	1, // 1: cpuinfo_ppc64_entry.endian:type_name -> cpuinfo_ppc64_entry.endianness
	2, // 2: cpuinfo_entry.x86_entry:type_name -> cpuinfo_x86_entry
	3, // 3: cpuinfo_entry.ppc64_entry:type_name -> cpuinfo_ppc64_entry
	4, // 4: cpuinfo_entry.s390_entry:type_name -> cpuinfo_s390_entry
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_cpuinfo_proto_init() }
func file_cpuinfo_proto_init() {
	if File_cpuinfo_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_cpuinfo_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CpuinfoX86Entry); i {
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
		file_cpuinfo_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CpuinfoPpc64Entry); i {
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
		file_cpuinfo_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CpuinfoS390Entry); i {
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
		file_cpuinfo_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CpuinfoEntry); i {
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
			RawDescriptor: file_cpuinfo_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_cpuinfo_proto_goTypes,
		DependencyIndexes: file_cpuinfo_proto_depIdxs,
		EnumInfos:         file_cpuinfo_proto_enumTypes,
		MessageInfos:      file_cpuinfo_proto_msgTypes,
	}.Build()
	File_cpuinfo_proto = out.File
	file_cpuinfo_proto_rawDesc = nil
	file_cpuinfo_proto_goTypes = nil
	file_cpuinfo_proto_depIdxs = nil
}
