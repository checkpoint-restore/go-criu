// SPDX-License-Identifier: MIT

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.19.6
// source: fdinfo.proto

package fdinfo

import (
	bpfmap_file "github.com/checkpoint-restore/go-criu/v7/crit/images/bpfmap-file"
	eventfd "github.com/checkpoint-restore/go-criu/v7/crit/images/eventfd"
	eventpoll "github.com/checkpoint-restore/go-criu/v7/crit/images/eventpoll"
	ext_file "github.com/checkpoint-restore/go-criu/v7/crit/images/ext-file"
	fifo "github.com/checkpoint-restore/go-criu/v7/crit/images/fifo"
	fsnotify "github.com/checkpoint-restore/go-criu/v7/crit/images/fsnotify"
	memfd "github.com/checkpoint-restore/go-criu/v7/crit/images/memfd"
	ns "github.com/checkpoint-restore/go-criu/v7/crit/images/ns"
	packet_sock "github.com/checkpoint-restore/go-criu/v7/crit/images/packet-sock"
	pidfd "github.com/checkpoint-restore/go-criu/v7/crit/images/pidfd"
	pipe "github.com/checkpoint-restore/go-criu/v7/crit/images/pipe"
	regfile "github.com/checkpoint-restore/go-criu/v7/crit/images/regfile"
	signalfd "github.com/checkpoint-restore/go-criu/v7/crit/images/signalfd"
	sk_inet "github.com/checkpoint-restore/go-criu/v7/crit/images/sk-inet"
	sk_netlink "github.com/checkpoint-restore/go-criu/v7/crit/images/sk-netlink"
	sk_unix "github.com/checkpoint-restore/go-criu/v7/crit/images/sk-unix"
	timerfd "github.com/checkpoint-restore/go-criu/v7/crit/images/timerfd"
	tty "github.com/checkpoint-restore/go-criu/v7/crit/images/tty"
	tun "github.com/checkpoint-restore/go-criu/v7/crit/images/tun"
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

type FdTypes int32

const (
	FdTypes_UND       FdTypes = 0
	FdTypes_REG       FdTypes = 1
	FdTypes_PIPE      FdTypes = 2
	FdTypes_FIFO      FdTypes = 3
	FdTypes_INETSK    FdTypes = 4
	FdTypes_UNIXSK    FdTypes = 5
	FdTypes_EVENTFD   FdTypes = 6
	FdTypes_EVENTPOLL FdTypes = 7
	FdTypes_INOTIFY   FdTypes = 8
	FdTypes_SIGNALFD  FdTypes = 9
	FdTypes_PACKETSK  FdTypes = 10
	FdTypes_TTY       FdTypes = 11
	FdTypes_FANOTIFY  FdTypes = 12
	FdTypes_NETLINKSK FdTypes = 13
	FdTypes_NS        FdTypes = 14
	FdTypes_TUNF      FdTypes = 15
	FdTypes_EXT       FdTypes = 16
	FdTypes_TIMERFD   FdTypes = 17
	FdTypes_MEMFD     FdTypes = 18
	FdTypes_BPFMAP    FdTypes = 19
	FdTypes_PIDFD     FdTypes = 20
	// Any number above the real used. Not stored to image
	FdTypes_CTL_TTY     FdTypes = 65534
	FdTypes_AUTOFS_PIPE FdTypes = 65535
)

// Enum value maps for FdTypes.
var (
	FdTypes_name = map[int32]string{
		0:     "UND",
		1:     "REG",
		2:     "PIPE",
		3:     "FIFO",
		4:     "INETSK",
		5:     "UNIXSK",
		6:     "EVENTFD",
		7:     "EVENTPOLL",
		8:     "INOTIFY",
		9:     "SIGNALFD",
		10:    "PACKETSK",
		11:    "TTY",
		12:    "FANOTIFY",
		13:    "NETLINKSK",
		14:    "NS",
		15:    "TUNF",
		16:    "EXT",
		17:    "TIMERFD",
		18:    "MEMFD",
		19:    "BPFMAP",
		20:    "PIDFD",
		65534: "CTL_TTY",
		65535: "AUTOFS_PIPE",
	}
	FdTypes_value = map[string]int32{
		"UND":         0,
		"REG":         1,
		"PIPE":        2,
		"FIFO":        3,
		"INETSK":      4,
		"UNIXSK":      5,
		"EVENTFD":     6,
		"EVENTPOLL":   7,
		"INOTIFY":     8,
		"SIGNALFD":    9,
		"PACKETSK":    10,
		"TTY":         11,
		"FANOTIFY":    12,
		"NETLINKSK":   13,
		"NS":          14,
		"TUNF":        15,
		"EXT":         16,
		"TIMERFD":     17,
		"MEMFD":       18,
		"BPFMAP":      19,
		"PIDFD":       20,
		"CTL_TTY":     65534,
		"AUTOFS_PIPE": 65535,
	}
)

func (x FdTypes) Enum() *FdTypes {
	p := new(FdTypes)
	*p = x
	return p
}

func (x FdTypes) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (FdTypes) Descriptor() protoreflect.EnumDescriptor {
	return file_fdinfo_proto_enumTypes[0].Descriptor()
}

func (FdTypes) Type() protoreflect.EnumType {
	return &file_fdinfo_proto_enumTypes[0]
}

func (x FdTypes) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Do not use.
func (x *FdTypes) UnmarshalJSON(b []byte) error {
	num, err := protoimpl.X.UnmarshalJSONEnum(x.Descriptor(), b)
	if err != nil {
		return err
	}
	*x = FdTypes(num)
	return nil
}

// Deprecated: Use FdTypes.Descriptor instead.
func (FdTypes) EnumDescriptor() ([]byte, []int) {
	return file_fdinfo_proto_rawDescGZIP(), []int{0}
}

type FdinfoEntry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id                   *uint32  `protobuf:"varint,1,req,name=id" json:"id,omitempty"`
	Flags                *uint32  `protobuf:"varint,2,req,name=flags" json:"flags,omitempty"`
	Type                 *FdTypes `protobuf:"varint,3,req,name=type,enum=FdTypes" json:"type,omitempty"`
	Fd                   *uint32  `protobuf:"varint,4,req,name=fd" json:"fd,omitempty"`
	XattrSecuritySelinux *string  `protobuf:"bytes,5,opt,name=xattr_security_selinux,json=xattrSecuritySelinux" json:"xattr_security_selinux,omitempty"`
}

func (x *FdinfoEntry) Reset() {
	*x = FdinfoEntry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_fdinfo_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FdinfoEntry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FdinfoEntry) ProtoMessage() {}

func (x *FdinfoEntry) ProtoReflect() protoreflect.Message {
	mi := &file_fdinfo_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FdinfoEntry.ProtoReflect.Descriptor instead.
func (*FdinfoEntry) Descriptor() ([]byte, []int) {
	return file_fdinfo_proto_rawDescGZIP(), []int{0}
}

func (x *FdinfoEntry) GetId() uint32 {
	if x != nil && x.Id != nil {
		return *x.Id
	}
	return 0
}

func (x *FdinfoEntry) GetFlags() uint32 {
	if x != nil && x.Flags != nil {
		return *x.Flags
	}
	return 0
}

func (x *FdinfoEntry) GetType() FdTypes {
	if x != nil && x.Type != nil {
		return *x.Type
	}
	return FdTypes_UND
}

func (x *FdinfoEntry) GetFd() uint32 {
	if x != nil && x.Fd != nil {
		return *x.Fd
	}
	return 0
}

func (x *FdinfoEntry) GetXattrSecuritySelinux() string {
	if x != nil && x.XattrSecuritySelinux != nil {
		return *x.XattrSecuritySelinux
	}
	return ""
}

type FileEntry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type  *FdTypes                      `protobuf:"varint,1,req,name=type,enum=FdTypes" json:"type,omitempty"`
	Id    *uint32                       `protobuf:"varint,2,req,name=id" json:"id,omitempty"`
	Reg   *regfile.RegFileEntry         `protobuf:"bytes,3,opt,name=reg" json:"reg,omitempty"`
	Isk   *sk_inet.InetSkEntry          `protobuf:"bytes,4,opt,name=isk" json:"isk,omitempty"`
	Nsf   *ns.NsFileEntry               `protobuf:"bytes,5,opt,name=nsf" json:"nsf,omitempty"`
	Psk   *packet_sock.PacketSockEntry  `protobuf:"bytes,6,opt,name=psk" json:"psk,omitempty"`
	Nlsk  *sk_netlink.NetlinkSkEntry    `protobuf:"bytes,7,opt,name=nlsk" json:"nlsk,omitempty"`
	Efd   *eventfd.EventfdFileEntry     `protobuf:"bytes,8,opt,name=efd" json:"efd,omitempty"`
	Epfd  *eventpoll.EventpollFileEntry `protobuf:"bytes,9,opt,name=epfd" json:"epfd,omitempty"`
	Sgfd  *signalfd.SignalfdEntry       `protobuf:"bytes,10,opt,name=sgfd" json:"sgfd,omitempty"`
	Tunf  *tun.TunfileEntry             `protobuf:"bytes,11,opt,name=tunf" json:"tunf,omitempty"`
	Tfd   *timerfd.TimerfdEntry         `protobuf:"bytes,12,opt,name=tfd" json:"tfd,omitempty"`
	Ify   *fsnotify.InotifyFileEntry    `protobuf:"bytes,13,opt,name=ify" json:"ify,omitempty"`
	Ffy   *fsnotify.FanotifyFileEntry   `protobuf:"bytes,14,opt,name=ffy" json:"ffy,omitempty"`
	Ext   *ext_file.ExtFileEntry        `protobuf:"bytes,15,opt,name=ext" json:"ext,omitempty"`
	Usk   *sk_unix.UnixSkEntry          `protobuf:"bytes,16,opt,name=usk" json:"usk,omitempty"`
	Fifo  *fifo.FifoEntry               `protobuf:"bytes,17,opt,name=fifo" json:"fifo,omitempty"`
	Pipe  *pipe.PipeEntry               `protobuf:"bytes,18,opt,name=pipe" json:"pipe,omitempty"`
	Tty   *tty.TtyFileEntry             `protobuf:"bytes,19,opt,name=tty" json:"tty,omitempty"`
	Memfd *memfd.MemfdFileEntry         `protobuf:"bytes,20,opt,name=memfd" json:"memfd,omitempty"`
	Bpf   *bpfmap_file.BpfmapFileEntry  `protobuf:"bytes,21,opt,name=bpf" json:"bpf,omitempty"`
	Pidfd *pidfd.PidfdEntry             `protobuf:"bytes,22,opt,name=pidfd" json:"pidfd,omitempty"`
}

func (x *FileEntry) Reset() {
	*x = FileEntry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_fdinfo_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FileEntry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FileEntry) ProtoMessage() {}

func (x *FileEntry) ProtoReflect() protoreflect.Message {
	mi := &file_fdinfo_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FileEntry.ProtoReflect.Descriptor instead.
func (*FileEntry) Descriptor() ([]byte, []int) {
	return file_fdinfo_proto_rawDescGZIP(), []int{1}
}

func (x *FileEntry) GetType() FdTypes {
	if x != nil && x.Type != nil {
		return *x.Type
	}
	return FdTypes_UND
}

func (x *FileEntry) GetId() uint32 {
	if x != nil && x.Id != nil {
		return *x.Id
	}
	return 0
}

func (x *FileEntry) GetReg() *regfile.RegFileEntry {
	if x != nil {
		return x.Reg
	}
	return nil
}

func (x *FileEntry) GetIsk() *sk_inet.InetSkEntry {
	if x != nil {
		return x.Isk
	}
	return nil
}

func (x *FileEntry) GetNsf() *ns.NsFileEntry {
	if x != nil {
		return x.Nsf
	}
	return nil
}

func (x *FileEntry) GetPsk() *packet_sock.PacketSockEntry {
	if x != nil {
		return x.Psk
	}
	return nil
}

func (x *FileEntry) GetNlsk() *sk_netlink.NetlinkSkEntry {
	if x != nil {
		return x.Nlsk
	}
	return nil
}

func (x *FileEntry) GetEfd() *eventfd.EventfdFileEntry {
	if x != nil {
		return x.Efd
	}
	return nil
}

func (x *FileEntry) GetEpfd() *eventpoll.EventpollFileEntry {
	if x != nil {
		return x.Epfd
	}
	return nil
}

func (x *FileEntry) GetSgfd() *signalfd.SignalfdEntry {
	if x != nil {
		return x.Sgfd
	}
	return nil
}

func (x *FileEntry) GetTunf() *tun.TunfileEntry {
	if x != nil {
		return x.Tunf
	}
	return nil
}

func (x *FileEntry) GetTfd() *timerfd.TimerfdEntry {
	if x != nil {
		return x.Tfd
	}
	return nil
}

func (x *FileEntry) GetIfy() *fsnotify.InotifyFileEntry {
	if x != nil {
		return x.Ify
	}
	return nil
}

func (x *FileEntry) GetFfy() *fsnotify.FanotifyFileEntry {
	if x != nil {
		return x.Ffy
	}
	return nil
}

func (x *FileEntry) GetExt() *ext_file.ExtFileEntry {
	if x != nil {
		return x.Ext
	}
	return nil
}

func (x *FileEntry) GetUsk() *sk_unix.UnixSkEntry {
	if x != nil {
		return x.Usk
	}
	return nil
}

func (x *FileEntry) GetFifo() *fifo.FifoEntry {
	if x != nil {
		return x.Fifo
	}
	return nil
}

func (x *FileEntry) GetPipe() *pipe.PipeEntry {
	if x != nil {
		return x.Pipe
	}
	return nil
}

func (x *FileEntry) GetTty() *tty.TtyFileEntry {
	if x != nil {
		return x.Tty
	}
	return nil
}

func (x *FileEntry) GetMemfd() *memfd.MemfdFileEntry {
	if x != nil {
		return x.Memfd
	}
	return nil
}

func (x *FileEntry) GetBpf() *bpfmap_file.BpfmapFileEntry {
	if x != nil {
		return x.Bpf
	}
	return nil
}

func (x *FileEntry) GetPidfd() *pidfd.PidfdEntry {
	if x != nil {
		return x.Pidfd
	}
	return nil
}

var File_fdinfo_proto protoreflect.FileDescriptor

var file_fdinfo_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x66, 0x64, 0x69, 0x6e, 0x66, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0d,
	0x72, 0x65, 0x67, 0x66, 0x69, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0d, 0x73,
	0x6b, 0x2d, 0x69, 0x6e, 0x65, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x08, 0x6e, 0x73,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x11, 0x70, 0x61, 0x63, 0x6b, 0x65, 0x74, 0x2d, 0x73,
	0x6f, 0x63, 0x6b, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x10, 0x73, 0x6b, 0x2d, 0x6e, 0x65,
	0x74, 0x6c, 0x69, 0x6e, 0x6b, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0d, 0x65, 0x76, 0x65,
	0x6e, 0x74, 0x66, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0f, 0x65, 0x76, 0x65, 0x6e,
	0x74, 0x70, 0x6f, 0x6c, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0e, 0x73, 0x69, 0x67,
	0x6e, 0x61, 0x6c, 0x66, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x09, 0x74, 0x75, 0x6e,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0d, 0x74, 0x69, 0x6d, 0x65, 0x72, 0x66, 0x64, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0e, 0x66, 0x73, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0e, 0x65, 0x78, 0x74, 0x2d, 0x66, 0x69, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0d, 0x73, 0x6b, 0x2d, 0x75, 0x6e, 0x69, 0x78, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0a, 0x66, 0x69, 0x66, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x0a, 0x70, 0x69, 0x70, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0b, 0x70, 0x69,
	0x64, 0x66, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x09, 0x74, 0x74, 0x79, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0b, 0x6d, 0x65, 0x6d, 0x66, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x1a, 0x11, 0x62, 0x70, 0x66, 0x6d, 0x61, 0x70, 0x2d, 0x66, 0x69, 0x6c, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0x99, 0x01, 0x0a, 0x0c, 0x66, 0x64, 0x69, 0x6e, 0x66, 0x6f, 0x5f,
	0x65, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x02, 0x28,
	0x0d, 0x52, 0x02, 0x69, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x66, 0x6c, 0x61, 0x67, 0x73, 0x18, 0x02,
	0x20, 0x02, 0x28, 0x0d, 0x52, 0x05, 0x66, 0x6c, 0x61, 0x67, 0x73, 0x12, 0x1d, 0x0a, 0x04, 0x74,
	0x79, 0x70, 0x65, 0x18, 0x03, 0x20, 0x02, 0x28, 0x0e, 0x32, 0x09, 0x2e, 0x66, 0x64, 0x5f, 0x74,
	0x79, 0x70, 0x65, 0x73, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x66, 0x64,
	0x18, 0x04, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x02, 0x66, 0x64, 0x12, 0x34, 0x0a, 0x16, 0x78, 0x61,
	0x74, 0x74, 0x72, 0x5f, 0x73, 0x65, 0x63, 0x75, 0x72, 0x69, 0x74, 0x79, 0x5f, 0x73, 0x65, 0x6c,
	0x69, 0x6e, 0x75, 0x78, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x14, 0x78, 0x61, 0x74, 0x74,
	0x72, 0x53, 0x65, 0x63, 0x75, 0x72, 0x69, 0x74, 0x79, 0x53, 0x65, 0x6c, 0x69, 0x6e, 0x75, 0x78,
	0x22, 0x98, 0x06, 0x0a, 0x0a, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x12,
	0x1d, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x02, 0x28, 0x0e, 0x32, 0x09, 0x2e,
	0x66, 0x64, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x0e,
	0x0a, 0x02, 0x69, 0x64, 0x18, 0x02, 0x20, 0x02, 0x28, 0x0d, 0x52, 0x02, 0x69, 0x64, 0x12, 0x21,
	0x0a, 0x03, 0x72, 0x65, 0x67, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x72, 0x65,
	0x67, 0x5f, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x03, 0x72, 0x65,
	0x67, 0x12, 0x20, 0x0a, 0x03, 0x69, 0x73, 0x6b, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e,
	0x2e, 0x69, 0x6e, 0x65, 0x74, 0x5f, 0x73, 0x6b, 0x5f, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x03,
	0x69, 0x73, 0x6b, 0x12, 0x20, 0x0a, 0x03, 0x6e, 0x73, 0x66, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x0e, 0x2e, 0x6e, 0x73, 0x5f, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x65, 0x6e, 0x74, 0x72, 0x79,
	0x52, 0x03, 0x6e, 0x73, 0x66, 0x12, 0x24, 0x0a, 0x03, 0x70, 0x73, 0x6b, 0x18, 0x06, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x12, 0x2e, 0x70, 0x61, 0x63, 0x6b, 0x65, 0x74, 0x5f, 0x73, 0x6f, 0x63, 0x6b,
	0x5f, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x03, 0x70, 0x73, 0x6b, 0x12, 0x25, 0x0a, 0x04, 0x6e,
	0x6c, 0x73, 0x6b, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x6e, 0x65, 0x74, 0x6c,
	0x69, 0x6e, 0x6b, 0x5f, 0x73, 0x6b, 0x5f, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x04, 0x6e, 0x6c,
	0x73, 0x6b, 0x12, 0x25, 0x0a, 0x03, 0x65, 0x66, 0x64, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x13, 0x2e, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x66, 0x64, 0x5f, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x65,
	0x6e, 0x74, 0x72, 0x79, 0x52, 0x03, 0x65, 0x66, 0x64, 0x12, 0x29, 0x0a, 0x04, 0x65, 0x70, 0x66,
	0x64, 0x18, 0x09, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x70,
	0x6f, 0x6c, 0x6c, 0x5f, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x04,
	0x65, 0x70, 0x66, 0x64, 0x12, 0x23, 0x0a, 0x04, 0x73, 0x67, 0x66, 0x64, 0x18, 0x0a, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x6c, 0x66, 0x64, 0x5f, 0x65, 0x6e,
	0x74, 0x72, 0x79, 0x52, 0x04, 0x73, 0x67, 0x66, 0x64, 0x12, 0x22, 0x0a, 0x04, 0x74, 0x75, 0x6e,
	0x66, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x74, 0x75, 0x6e, 0x66, 0x69, 0x6c,
	0x65, 0x5f, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x04, 0x74, 0x75, 0x6e, 0x66, 0x12, 0x20, 0x0a,
	0x03, 0x74, 0x66, 0x64, 0x18, 0x0c, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x74, 0x69, 0x6d,
	0x65, 0x72, 0x66, 0x64, 0x5f, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x03, 0x74, 0x66, 0x64, 0x12,
	0x25, 0x0a, 0x03, 0x69, 0x66, 0x79, 0x18, 0x0d, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x69,
	0x6e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x5f, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x65, 0x6e, 0x74, 0x72,
	0x79, 0x52, 0x03, 0x69, 0x66, 0x79, 0x12, 0x26, 0x0a, 0x03, 0x66, 0x66, 0x79, 0x18, 0x0e, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x66, 0x61, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x5f, 0x66,
	0x69, 0x6c, 0x65, 0x5f, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x03, 0x66, 0x66, 0x79, 0x12, 0x21,
	0x0a, 0x03, 0x65, 0x78, 0x74, 0x18, 0x0f, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x65, 0x78,
	0x74, 0x5f, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x03, 0x65, 0x78,
	0x74, 0x12, 0x20, 0x0a, 0x03, 0x75, 0x73, 0x6b, 0x18, 0x10, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e,
	0x2e, 0x75, 0x6e, 0x69, 0x78, 0x5f, 0x73, 0x6b, 0x5f, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x03,
	0x75, 0x73, 0x6b, 0x12, 0x1f, 0x0a, 0x04, 0x66, 0x69, 0x66, 0x6f, 0x18, 0x11, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x0b, 0x2e, 0x66, 0x69, 0x66, 0x6f, 0x5f, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x04,
	0x66, 0x69, 0x66, 0x6f, 0x12, 0x1f, 0x0a, 0x04, 0x70, 0x69, 0x70, 0x65, 0x18, 0x12, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x0b, 0x2e, 0x70, 0x69, 0x70, 0x65, 0x5f, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x52,
	0x04, 0x70, 0x69, 0x70, 0x65, 0x12, 0x21, 0x0a, 0x03, 0x74, 0x74, 0x79, 0x18, 0x13, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x74, 0x74, 0x79, 0x5f, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x65, 0x6e,
	0x74, 0x72, 0x79, 0x52, 0x03, 0x74, 0x74, 0x79, 0x12, 0x27, 0x0a, 0x05, 0x6d, 0x65, 0x6d, 0x66,
	0x64, 0x18, 0x14, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x6d, 0x65, 0x6d, 0x66, 0x64, 0x5f,
	0x66, 0x69, 0x6c, 0x65, 0x5f, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x05, 0x6d, 0x65, 0x6d, 0x66,
	0x64, 0x12, 0x24, 0x0a, 0x03, 0x62, 0x70, 0x66, 0x18, 0x15, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12,
	0x2e, 0x62, 0x70, 0x66, 0x6d, 0x61, 0x70, 0x5f, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x65, 0x6e, 0x74,
	0x72, 0x79, 0x52, 0x03, 0x62, 0x70, 0x66, 0x12, 0x22, 0x0a, 0x05, 0x70, 0x69, 0x64, 0x66, 0x64,
	0x18, 0x16, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0c, 0x2e, 0x70, 0x69, 0x64, 0x66, 0x64, 0x5f, 0x65,
	0x6e, 0x74, 0x72, 0x79, 0x52, 0x05, 0x70, 0x69, 0x64, 0x66, 0x64, 0x2a, 0x9f, 0x02, 0x0a, 0x08,
	0x66, 0x64, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x12, 0x07, 0x0a, 0x03, 0x55, 0x4e, 0x44, 0x10,
	0x00, 0x12, 0x07, 0x0a, 0x03, 0x52, 0x45, 0x47, 0x10, 0x01, 0x12, 0x08, 0x0a, 0x04, 0x50, 0x49,
	0x50, 0x45, 0x10, 0x02, 0x12, 0x08, 0x0a, 0x04, 0x46, 0x49, 0x46, 0x4f, 0x10, 0x03, 0x12, 0x0a,
	0x0a, 0x06, 0x49, 0x4e, 0x45, 0x54, 0x53, 0x4b, 0x10, 0x04, 0x12, 0x0a, 0x0a, 0x06, 0x55, 0x4e,
	0x49, 0x58, 0x53, 0x4b, 0x10, 0x05, 0x12, 0x0b, 0x0a, 0x07, 0x45, 0x56, 0x45, 0x4e, 0x54, 0x46,
	0x44, 0x10, 0x06, 0x12, 0x0d, 0x0a, 0x09, 0x45, 0x56, 0x45, 0x4e, 0x54, 0x50, 0x4f, 0x4c, 0x4c,
	0x10, 0x07, 0x12, 0x0b, 0x0a, 0x07, 0x49, 0x4e, 0x4f, 0x54, 0x49, 0x46, 0x59, 0x10, 0x08, 0x12,
	0x0c, 0x0a, 0x08, 0x53, 0x49, 0x47, 0x4e, 0x41, 0x4c, 0x46, 0x44, 0x10, 0x09, 0x12, 0x0c, 0x0a,
	0x08, 0x50, 0x41, 0x43, 0x4b, 0x45, 0x54, 0x53, 0x4b, 0x10, 0x0a, 0x12, 0x07, 0x0a, 0x03, 0x54,
	0x54, 0x59, 0x10, 0x0b, 0x12, 0x0c, 0x0a, 0x08, 0x46, 0x41, 0x4e, 0x4f, 0x54, 0x49, 0x46, 0x59,
	0x10, 0x0c, 0x12, 0x0d, 0x0a, 0x09, 0x4e, 0x45, 0x54, 0x4c, 0x49, 0x4e, 0x4b, 0x53, 0x4b, 0x10,
	0x0d, 0x12, 0x06, 0x0a, 0x02, 0x4e, 0x53, 0x10, 0x0e, 0x12, 0x08, 0x0a, 0x04, 0x54, 0x55, 0x4e,
	0x46, 0x10, 0x0f, 0x12, 0x07, 0x0a, 0x03, 0x45, 0x58, 0x54, 0x10, 0x10, 0x12, 0x0b, 0x0a, 0x07,
	0x54, 0x49, 0x4d, 0x45, 0x52, 0x46, 0x44, 0x10, 0x11, 0x12, 0x09, 0x0a, 0x05, 0x4d, 0x45, 0x4d,
	0x46, 0x44, 0x10, 0x12, 0x12, 0x0a, 0x0a, 0x06, 0x42, 0x50, 0x46, 0x4d, 0x41, 0x50, 0x10, 0x13,
	0x12, 0x09, 0x0a, 0x05, 0x50, 0x49, 0x44, 0x46, 0x44, 0x10, 0x14, 0x12, 0x0d, 0x0a, 0x07, 0x43,
	0x54, 0x4c, 0x5f, 0x54, 0x54, 0x59, 0x10, 0xfe, 0xff, 0x03, 0x12, 0x11, 0x0a, 0x0b, 0x41, 0x55,
	0x54, 0x4f, 0x46, 0x53, 0x5f, 0x50, 0x49, 0x50, 0x45, 0x10, 0xff, 0xff, 0x03,
}

var (
	file_fdinfo_proto_rawDescOnce sync.Once
	file_fdinfo_proto_rawDescData = file_fdinfo_proto_rawDesc
)

func file_fdinfo_proto_rawDescGZIP() []byte {
	file_fdinfo_proto_rawDescOnce.Do(func() {
		file_fdinfo_proto_rawDescData = protoimpl.X.CompressGZIP(file_fdinfo_proto_rawDescData)
	})
	return file_fdinfo_proto_rawDescData
}

var file_fdinfo_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_fdinfo_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_fdinfo_proto_goTypes = []interface{}{
	(FdTypes)(0),                         // 0: fd_types
	(*FdinfoEntry)(nil),                  // 1: fdinfo_entry
	(*FileEntry)(nil),                    // 2: file_entry
	(*regfile.RegFileEntry)(nil),         // 3: reg_file_entry
	(*sk_inet.InetSkEntry)(nil),          // 4: inet_sk_entry
	(*ns.NsFileEntry)(nil),               // 5: ns_file_entry
	(*packet_sock.PacketSockEntry)(nil),  // 6: packet_sock_entry
	(*sk_netlink.NetlinkSkEntry)(nil),    // 7: netlink_sk_entry
	(*eventfd.EventfdFileEntry)(nil),     // 8: eventfd_file_entry
	(*eventpoll.EventpollFileEntry)(nil), // 9: eventpoll_file_entry
	(*signalfd.SignalfdEntry)(nil),       // 10: signalfd_entry
	(*tun.TunfileEntry)(nil),             // 11: tunfile_entry
	(*timerfd.TimerfdEntry)(nil),         // 12: timerfd_entry
	(*fsnotify.InotifyFileEntry)(nil),    // 13: inotify_file_entry
	(*fsnotify.FanotifyFileEntry)(nil),   // 14: fanotify_file_entry
	(*ext_file.ExtFileEntry)(nil),        // 15: ext_file_entry
	(*sk_unix.UnixSkEntry)(nil),          // 16: unix_sk_entry
	(*fifo.FifoEntry)(nil),               // 17: fifo_entry
	(*pipe.PipeEntry)(nil),               // 18: pipe_entry
	(*tty.TtyFileEntry)(nil),             // 19: tty_file_entry
	(*memfd.MemfdFileEntry)(nil),         // 20: memfd_file_entry
	(*bpfmap_file.BpfmapFileEntry)(nil),  // 21: bpfmap_file_entry
	(*pidfd.PidfdEntry)(nil),             // 22: pidfd_entry
}
var file_fdinfo_proto_depIdxs = []int32{
	0,  // 0: fdinfo_entry.type:type_name -> fd_types
	0,  // 1: file_entry.type:type_name -> fd_types
	3,  // 2: file_entry.reg:type_name -> reg_file_entry
	4,  // 3: file_entry.isk:type_name -> inet_sk_entry
	5,  // 4: file_entry.nsf:type_name -> ns_file_entry
	6,  // 5: file_entry.psk:type_name -> packet_sock_entry
	7,  // 6: file_entry.nlsk:type_name -> netlink_sk_entry
	8,  // 7: file_entry.efd:type_name -> eventfd_file_entry
	9,  // 8: file_entry.epfd:type_name -> eventpoll_file_entry
	10, // 9: file_entry.sgfd:type_name -> signalfd_entry
	11, // 10: file_entry.tunf:type_name -> tunfile_entry
	12, // 11: file_entry.tfd:type_name -> timerfd_entry
	13, // 12: file_entry.ify:type_name -> inotify_file_entry
	14, // 13: file_entry.ffy:type_name -> fanotify_file_entry
	15, // 14: file_entry.ext:type_name -> ext_file_entry
	16, // 15: file_entry.usk:type_name -> unix_sk_entry
	17, // 16: file_entry.fifo:type_name -> fifo_entry
	18, // 17: file_entry.pipe:type_name -> pipe_entry
	19, // 18: file_entry.tty:type_name -> tty_file_entry
	20, // 19: file_entry.memfd:type_name -> memfd_file_entry
	21, // 20: file_entry.bpf:type_name -> bpfmap_file_entry
	22, // 21: file_entry.pidfd:type_name -> pidfd_entry
	22, // [22:22] is the sub-list for method output_type
	22, // [22:22] is the sub-list for method input_type
	22, // [22:22] is the sub-list for extension type_name
	22, // [22:22] is the sub-list for extension extendee
	0,  // [0:22] is the sub-list for field type_name
}

func init() { file_fdinfo_proto_init() }
func file_fdinfo_proto_init() {
	if File_fdinfo_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_fdinfo_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FdinfoEntry); i {
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
		file_fdinfo_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FileEntry); i {
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
			RawDescriptor: file_fdinfo_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_fdinfo_proto_goTypes,
		DependencyIndexes: file_fdinfo_proto_depIdxs,
		EnumInfos:         file_fdinfo_proto_enumTypes,
		MessageInfos:      file_fdinfo_proto_msgTypes,
	}.Build()
	File_fdinfo_proto = out.File
	file_fdinfo_proto_rawDesc = nil
	file_fdinfo_proto_goTypes = nil
	file_fdinfo_proto_depIdxs = nil
}
