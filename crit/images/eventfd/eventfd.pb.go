// SPDX-License-Identifier: MIT

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.19.6
// source: eventfd.proto

package eventfd

import (
	fown "github.com/checkpoint-restore/go-criu/v7/crit/images/fown"
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

type EventfdFileEntry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id      *uint32         `protobuf:"varint,1,req,name=id" json:"id,omitempty"`
	Flags   *uint32         `protobuf:"varint,2,req,name=flags" json:"flags,omitempty"`
	Fown    *fown.FownEntry `protobuf:"bytes,3,req,name=fown" json:"fown,omitempty"`
	Counter *uint64         `protobuf:"varint,4,req,name=counter" json:"counter,omitempty"`
}

func (x *EventfdFileEntry) Reset() {
	*x = EventfdFileEntry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_eventfd_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EventfdFileEntry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EventfdFileEntry) ProtoMessage() {}

func (x *EventfdFileEntry) ProtoReflect() protoreflect.Message {
	mi := &file_eventfd_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EventfdFileEntry.ProtoReflect.Descriptor instead.
func (*EventfdFileEntry) Descriptor() ([]byte, []int) {
	return file_eventfd_proto_rawDescGZIP(), []int{0}
}

func (x *EventfdFileEntry) GetId() uint32 {
	if x != nil && x.Id != nil {
		return *x.Id
	}
	return 0
}

func (x *EventfdFileEntry) GetFlags() uint32 {
	if x != nil && x.Flags != nil {
		return *x.Flags
	}
	return 0
}

func (x *EventfdFileEntry) GetFown() *fown.FownEntry {
	if x != nil {
		return x.Fown
	}
	return nil
}

func (x *EventfdFileEntry) GetCounter() uint64 {
	if x != nil && x.Counter != nil {
		return *x.Counter
	}
	return 0
}

var File_eventfd_proto protoreflect.FileDescriptor

var file_eventfd_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x66, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x0a, 0x66, 0x6f, 0x77, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x75, 0x0a, 0x12, 0x65,
	0x76, 0x65, 0x6e, 0x74, 0x66, 0x64, 0x5f, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x65, 0x6e, 0x74, 0x72,
	0x79, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x02, 0x69,
	0x64, 0x12, 0x14, 0x0a, 0x05, 0x66, 0x6c, 0x61, 0x67, 0x73, 0x18, 0x02, 0x20, 0x02, 0x28, 0x0d,
	0x52, 0x05, 0x66, 0x6c, 0x61, 0x67, 0x73, 0x12, 0x1f, 0x0a, 0x04, 0x66, 0x6f, 0x77, 0x6e, 0x18,
	0x03, 0x20, 0x02, 0x28, 0x0b, 0x32, 0x0b, 0x2e, 0x66, 0x6f, 0x77, 0x6e, 0x5f, 0x65, 0x6e, 0x74,
	0x72, 0x79, 0x52, 0x04, 0x66, 0x6f, 0x77, 0x6e, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x6f, 0x75, 0x6e,
	0x74, 0x65, 0x72, 0x18, 0x04, 0x20, 0x02, 0x28, 0x04, 0x52, 0x07, 0x63, 0x6f, 0x75, 0x6e, 0x74,
	0x65, 0x72,
}

var (
	file_eventfd_proto_rawDescOnce sync.Once
	file_eventfd_proto_rawDescData = file_eventfd_proto_rawDesc
)

func file_eventfd_proto_rawDescGZIP() []byte {
	file_eventfd_proto_rawDescOnce.Do(func() {
		file_eventfd_proto_rawDescData = protoimpl.X.CompressGZIP(file_eventfd_proto_rawDescData)
	})
	return file_eventfd_proto_rawDescData
}

var file_eventfd_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_eventfd_proto_goTypes = []interface{}{
	(*EventfdFileEntry)(nil), // 0: eventfd_file_entry
	(*fown.FownEntry)(nil),   // 1: fown_entry
}
var file_eventfd_proto_depIdxs = []int32{
	1, // 0: eventfd_file_entry.fown:type_name -> fown_entry
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_eventfd_proto_init() }
func file_eventfd_proto_init() {
	if File_eventfd_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_eventfd_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EventfdFileEntry); i {
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
			RawDescriptor: file_eventfd_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_eventfd_proto_goTypes,
		DependencyIndexes: file_eventfd_proto_depIdxs,
		MessageInfos:      file_eventfd_proto_msgTypes,
	}.Build()
	File_eventfd_proto = out.File
	file_eventfd_proto_rawDesc = nil
	file_eventfd_proto_goTypes = nil
	file_eventfd_proto_depIdxs = nil
}
