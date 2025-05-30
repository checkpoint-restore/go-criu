// SPDX-License-Identifier: MIT

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.19.6
// source: img-streamer.proto

package img_streamer

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

// This message is sent from CRIU to the streamer.
//   - During dump, it communicates the name of the file that is about to be sent
//     to the streamer.
//   - During restore, CRIU requests image files from the streamer. The message is
//     used to communicate the name of the desired file.
type ImgStreamerRequestEntry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Filename *string `protobuf:"bytes,1,req,name=filename" json:"filename,omitempty"`
}

func (x *ImgStreamerRequestEntry) Reset() {
	*x = ImgStreamerRequestEntry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_img_streamer_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ImgStreamerRequestEntry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ImgStreamerRequestEntry) ProtoMessage() {}

func (x *ImgStreamerRequestEntry) ProtoReflect() protoreflect.Message {
	mi := &file_img_streamer_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ImgStreamerRequestEntry.ProtoReflect.Descriptor instead.
func (*ImgStreamerRequestEntry) Descriptor() ([]byte, []int) {
	return file_img_streamer_proto_rawDescGZIP(), []int{0}
}

func (x *ImgStreamerRequestEntry) GetFilename() string {
	if x != nil && x.Filename != nil {
		return *x.Filename
	}
	return ""
}

// This message is sent from the streamer to CRIU. It is only used during
// restore to report whether the requested file exists.
type ImgStreamerReplyEntry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Exists *bool `protobuf:"varint,1,req,name=exists" json:"exists,omitempty"`
}

func (x *ImgStreamerReplyEntry) Reset() {
	*x = ImgStreamerReplyEntry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_img_streamer_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ImgStreamerReplyEntry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ImgStreamerReplyEntry) ProtoMessage() {}

func (x *ImgStreamerReplyEntry) ProtoReflect() protoreflect.Message {
	mi := &file_img_streamer_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ImgStreamerReplyEntry.ProtoReflect.Descriptor instead.
func (*ImgStreamerReplyEntry) Descriptor() ([]byte, []int) {
	return file_img_streamer_proto_rawDescGZIP(), []int{1}
}

func (x *ImgStreamerReplyEntry) GetExists() bool {
	if x != nil && x.Exists != nil {
		return *x.Exists
	}
	return false
}

var File_img_streamer_proto protoreflect.FileDescriptor

var file_img_streamer_proto_rawDesc = []byte{
	0x0a, 0x12, 0x69, 0x6d, 0x67, 0x2d, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x65, 0x72, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0x38, 0x0a, 0x1a, 0x69, 0x6d, 0x67, 0x5f, 0x73, 0x74, 0x72, 0x65,
	0x61, 0x6d, 0x65, 0x72, 0x5f, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x5f, 0x65, 0x6e, 0x74,
	0x72, 0x79, 0x12, 0x1a, 0x0a, 0x08, 0x66, 0x69, 0x6c, 0x65, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01,
	0x20, 0x02, 0x28, 0x09, 0x52, 0x08, 0x66, 0x69, 0x6c, 0x65, 0x6e, 0x61, 0x6d, 0x65, 0x22, 0x32,
	0x0a, 0x18, 0x69, 0x6d, 0x67, 0x5f, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x65, 0x72, 0x5f, 0x72,
	0x65, 0x70, 0x6c, 0x79, 0x5f, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x16, 0x0a, 0x06, 0x65, 0x78,
	0x69, 0x73, 0x74, 0x73, 0x18, 0x01, 0x20, 0x02, 0x28, 0x08, 0x52, 0x06, 0x65, 0x78, 0x69, 0x73,
	0x74, 0x73,
}

var (
	file_img_streamer_proto_rawDescOnce sync.Once
	file_img_streamer_proto_rawDescData = file_img_streamer_proto_rawDesc
)

func file_img_streamer_proto_rawDescGZIP() []byte {
	file_img_streamer_proto_rawDescOnce.Do(func() {
		file_img_streamer_proto_rawDescData = protoimpl.X.CompressGZIP(file_img_streamer_proto_rawDescData)
	})
	return file_img_streamer_proto_rawDescData
}

var file_img_streamer_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_img_streamer_proto_goTypes = []interface{}{
	(*ImgStreamerRequestEntry)(nil), // 0: img_streamer_request_entry
	(*ImgStreamerReplyEntry)(nil),   // 1: img_streamer_reply_entry
}
var file_img_streamer_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_img_streamer_proto_init() }
func file_img_streamer_proto_init() {
	if File_img_streamer_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_img_streamer_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ImgStreamerRequestEntry); i {
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
		file_img_streamer_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ImgStreamerReplyEntry); i {
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
			RawDescriptor: file_img_streamer_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_img_streamer_proto_goTypes,
		DependencyIndexes: file_img_streamer_proto_depIdxs,
		MessageInfos:      file_img_streamer_proto_msgTypes,
	}.Build()
	File_img_streamer_proto = out.File
	file_img_streamer_proto_rawDesc = nil
	file_img_streamer_proto_goTypes = nil
	file_img_streamer_proto_depIdxs = nil
}
