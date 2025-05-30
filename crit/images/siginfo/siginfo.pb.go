// SPDX-License-Identifier: MIT

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.19.6
// source: siginfo.proto

package siginfo

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

type SiginfoEntry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Siginfo []byte `protobuf:"bytes,1,req,name=siginfo" json:"siginfo,omitempty"`
}

func (x *SiginfoEntry) Reset() {
	*x = SiginfoEntry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_siginfo_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SiginfoEntry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SiginfoEntry) ProtoMessage() {}

func (x *SiginfoEntry) ProtoReflect() protoreflect.Message {
	mi := &file_siginfo_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SiginfoEntry.ProtoReflect.Descriptor instead.
func (*SiginfoEntry) Descriptor() ([]byte, []int) {
	return file_siginfo_proto_rawDescGZIP(), []int{0}
}

func (x *SiginfoEntry) GetSiginfo() []byte {
	if x != nil {
		return x.Siginfo
	}
	return nil
}

type SignalQueueEntry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Signals []*SiginfoEntry `protobuf:"bytes,1,rep,name=signals" json:"signals,omitempty"`
}

func (x *SignalQueueEntry) Reset() {
	*x = SignalQueueEntry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_siginfo_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SignalQueueEntry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SignalQueueEntry) ProtoMessage() {}

func (x *SignalQueueEntry) ProtoReflect() protoreflect.Message {
	mi := &file_siginfo_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SignalQueueEntry.ProtoReflect.Descriptor instead.
func (*SignalQueueEntry) Descriptor() ([]byte, []int) {
	return file_siginfo_proto_rawDescGZIP(), []int{1}
}

func (x *SignalQueueEntry) GetSignals() []*SiginfoEntry {
	if x != nil {
		return x.Signals
	}
	return nil
}

var File_siginfo_proto protoreflect.FileDescriptor

var file_siginfo_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x73, 0x69, 0x67, 0x69, 0x6e, 0x66, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22,
	0x29, 0x0a, 0x0d, 0x73, 0x69, 0x67, 0x69, 0x6e, 0x66, 0x6f, 0x5f, 0x65, 0x6e, 0x74, 0x72, 0x79,
	0x12, 0x18, 0x0a, 0x07, 0x73, 0x69, 0x67, 0x69, 0x6e, 0x66, 0x6f, 0x18, 0x01, 0x20, 0x02, 0x28,
	0x0c, 0x52, 0x07, 0x73, 0x69, 0x67, 0x69, 0x6e, 0x66, 0x6f, 0x22, 0x3e, 0x0a, 0x12, 0x73, 0x69,
	0x67, 0x6e, 0x61, 0x6c, 0x5f, 0x71, 0x75, 0x65, 0x75, 0x65, 0x5f, 0x65, 0x6e, 0x74, 0x72, 0x79,
	0x12, 0x28, 0x0a, 0x07, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x6c, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x0e, 0x2e, 0x73, 0x69, 0x67, 0x69, 0x6e, 0x66, 0x6f, 0x5f, 0x65, 0x6e, 0x74, 0x72,
	0x79, 0x52, 0x07, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x6c, 0x73,
}

var (
	file_siginfo_proto_rawDescOnce sync.Once
	file_siginfo_proto_rawDescData = file_siginfo_proto_rawDesc
)

func file_siginfo_proto_rawDescGZIP() []byte {
	file_siginfo_proto_rawDescOnce.Do(func() {
		file_siginfo_proto_rawDescData = protoimpl.X.CompressGZIP(file_siginfo_proto_rawDescData)
	})
	return file_siginfo_proto_rawDescData
}

var file_siginfo_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_siginfo_proto_goTypes = []interface{}{
	(*SiginfoEntry)(nil),     // 0: siginfo_entry
	(*SignalQueueEntry)(nil), // 1: signal_queue_entry
}
var file_siginfo_proto_depIdxs = []int32{
	0, // 0: signal_queue_entry.signals:type_name -> siginfo_entry
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_siginfo_proto_init() }
func file_siginfo_proto_init() {
	if File_siginfo_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_siginfo_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SiginfoEntry); i {
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
		file_siginfo_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SignalQueueEntry); i {
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
			RawDescriptor: file_siginfo_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_siginfo_proto_goTypes,
		DependencyIndexes: file_siginfo_proto_depIdxs,
		MessageInfos:      file_siginfo_proto_msgTypes,
	}.Build()
	File_siginfo_proto = out.File
	file_siginfo_proto_rawDesc = nil
	file_siginfo_proto_goTypes = nil
	file_siginfo_proto_depIdxs = nil
}
