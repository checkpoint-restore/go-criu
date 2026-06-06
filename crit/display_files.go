package crit

import (
	"encoding/json"
	"fmt"

	"github.com/checkpoint-restore/go-criu/v8/crit/images/fdinfo"
	sk_inet "github.com/checkpoint-restore/go-criu/v8/crit/images/sk-inet"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// fileEntryProtojson is used for files.img entries before optional field patches.
var fileEntryProtojson = protojson.MarshalOptions{
	UseProtoNames: true,
	AllowPartial:  true,
}

// fileEntryPayloadHumanizer humanizes one FileEntry payload sub-message (reg, isk, …).
type fileEntryPayloadHumanizer func(proto.Message) ([]byte, error)

type fileEntryHumanSpec struct {
	jsonField string
	payload   func(*fdinfo.FileEntry) proto.Message
	humanize  fileEntryPayloadHumanizer
}

// fileEntryHumanizers maps fd_types to payload humanizers. Add an entry here when
// implementing humanized output for a new files.img fd type.
var fileEntryHumanizers = map[fdinfo.FdTypes]fileEntryHumanSpec{
	fdinfo.FdTypes_INETSK: {
		jsonField: "isk",
		payload:   func(e *fdinfo.FileEntry) proto.Message { return e.GetIsk() },
		humanize:  humanizeInetSkPayload,
	},
	// fdinfo.FdTypes_REG: { jsonField: "reg", payload: ..., humanize: humanizeRegPayload },
}

var (
	inetSkFamilyHumanMap = map[uint32]string{
		1:  "UNIX",
		2:  "INET",
		10: "INET6",
		16: "NETLINK",
		17: "PACKET",
	}
	inetSkTypeHumanMap = map[uint32]string{
		1:  "STREAM",
		2:  "DGRAM",
		3:  "RAW",
		5:  "SEQPACKET",
		10: "PACKET",
	}
	inetSkProtocolHumanMap = map[uint32]string{
		0:   "IP",
		6:   "TCP",
		17:  "UDP",
		136: "UDPLITE",
	}
	inetSkStateHumanMap = map[uint32]string{
		1:  "ESTABLISHED",
		2:  "SYN_SENT",
		3:  "SYN_RECV",
		4:  "FIN_WAIT1",
		5:  "FIN_WAIT2",
		6:  "TIME_WAIT",
		7:  "CLOSE",
		8:  "CLOSE_WAIT",
		9:  "LAST_ACK",
		10: "LISTEN",
	}
)

// fileEntryJSON mirrors file_entry field order from fdinfo.proto.
type fileEntryJSON struct {
	Type  string          `json:"type"`
	Id    uint32          `json:"id"`
	Reg   json.RawMessage `json:"reg,omitempty"`
	Isk   json.RawMessage `json:"isk,omitempty"`
	Nsf   json.RawMessage `json:"nsf,omitempty"`
	Psk   json.RawMessage `json:"psk,omitempty"`
	Nlsk  json.RawMessage `json:"nlsk,omitempty"`
	Efd   json.RawMessage `json:"efd,omitempty"`
	Epfd  json.RawMessage `json:"epfd,omitempty"`
	Sgfd  json.RawMessage `json:"sgfd,omitempty"`
	Tunf  json.RawMessage `json:"tunf,omitempty"`
	Tfd   json.RawMessage `json:"tfd,omitempty"`
	Ify   json.RawMessage `json:"ify,omitempty"`
	Ffy   json.RawMessage `json:"ffy,omitempty"`
	Ext   json.RawMessage `json:"ext,omitempty"`
	Usk   json.RawMessage `json:"usk,omitempty"`
	Fifo  json.RawMessage `json:"fifo,omitempty"`
	Pipe  json.RawMessage `json:"pipe,omitempty"`
	Tty   json.RawMessage `json:"tty,omitempty"`
	Memfd json.RawMessage `json:"memfd,omitempty"`
	Bpf   json.RawMessage `json:"bpf,omitempty"`
	Pidfd json.RawMessage `json:"pidfd,omitempty"`
}

// inetSkJSON mirrors inet_sk_entry field order from sk-inet.proto.
type inetSkJSON struct {
	Id       json.RawMessage `json:"id"`
	Ino      json.RawMessage `json:"ino"`
	Family   json.RawMessage `json:"family,omitempty"`
	Type     json.RawMessage `json:"type,omitempty"`
	Proto    json.RawMessage `json:"proto,omitempty"`
	State    json.RawMessage `json:"state,omitempty"`
	SrcPort  json.RawMessage `json:"src_port"`
	DstPort  json.RawMessage `json:"dst_port"`
	Flags    json.RawMessage `json:"flags,omitempty"`
	Backlog  json.RawMessage `json:"backlog"`
	SrcAddr  json.RawMessage `json:"src_addr,omitempty"`
	DstAddr  json.RawMessage `json:"dst_addr,omitempty"`
	Fown     json.RawMessage `json:"fown,omitempty"`
	Opts     json.RawMessage `json:"opts,omitempty"`
	V6only   json.RawMessage `json:"v6only,omitempty"`
	IpOpts   json.RawMessage `json:"ip_opts,omitempty"`
	Ifname   json.RawMessage `json:"ifname,omitempty"`
	NsId     json.RawMessage `json:"ns_id,omitempty"`
	Shutdown json.RawMessage `json:"shutdown,omitempty"`
	TcpOpts  json.RawMessage `json:"tcp_opts,omitempty"`
}

func marshalFileEntryHuman(msg proto.Message) ([]byte, error) {
	entry, ok := msg.(*fdinfo.FileEntry)
	if !ok {
		return nil, fmt.Errorf("unexpected message type %T", msg)
	}

	data, err := fileEntryProtojson.Marshal(entry)
	if err != nil {
		return nil, err
	}

	spec, ok := fileEntryHumanizers[entry.GetType()]
	if !ok {
		return data, nil
	}

	payload := spec.payload(entry)
	if payload == nil {
		return data, nil
	}

	humanPayload, err := spec.humanize(payload)
	if err != nil {
		return nil, err
	}

	var fe fileEntryJSON
	if err := json.Unmarshal(data, &fe); err != nil {
		return nil, err
	}
	if err := fe.setPayloadField(spec.jsonField, humanPayload); err != nil {
		return nil, err
	}
	return json.Marshal(fe)
}

func (fe *fileEntryJSON) setPayloadField(field string, payload json.RawMessage) error {
	switch field {
	case "reg":
		fe.Reg = payload
	case "isk":
		fe.Isk = payload
	case "nsf":
		fe.Nsf = payload
	case "psk":
		fe.Psk = payload
	case "nlsk":
		fe.Nlsk = payload
	case "efd":
		fe.Efd = payload
	case "epfd":
		fe.Epfd = payload
	case "sgfd":
		fe.Sgfd = payload
	case "tunf":
		fe.Tunf = payload
	case "tfd":
		fe.Tfd = payload
	case "ify":
		fe.Ify = payload
	case "ffy":
		fe.Ffy = payload
	case "ext":
		fe.Ext = payload
	case "usk":
		fe.Usk = payload
	case "fifo":
		fe.Fifo = payload
	case "pipe":
		fe.Pipe = payload
	case "tty":
		fe.Tty = payload
	case "memfd":
		fe.Memfd = payload
	case "bpf":
		fe.Bpf = payload
	case "pidfd":
		fe.Pidfd = payload
	default:
		return fmt.Errorf("unsupported payload field %q", field)
	}
	return nil
}

func humanizeInetSkPayload(msg proto.Message) ([]byte, error) {
	isk, ok := msg.(*sk_inet.InetSkEntry)
	if !ok {
		return nil, fmt.Errorf("humanizeInetSkPayload: unexpected type %T", msg)
	}

	data, err := fileEntryProtojson.Marshal(isk)
	if err != nil {
		return nil, err
	}

	var sk inetSkJSON
	if err := json.Unmarshal(data, &sk); err != nil {
		return nil, err
	}

	if isk.Family != nil {
		sk.Family = rawJSONHumanMap(inetSkFamilyHumanMap, isk.GetFamily())
	}
	if isk.Type != nil {
		sk.Type = rawJSONHumanMap(inetSkTypeHumanMap, isk.GetType())
	}
	if isk.Proto != nil {
		sk.Proto = rawJSONHumanMap(inetSkProtocolHumanMap, isk.GetProto())
	}
	if isk.State != nil {
		sk.State = rawJSONHumanMap(inetSkStateHumanMap, isk.GetState())
	}
	if isk.Flags != nil {
		sk.Flags = rawJSONString(fmt.Sprintf("0x%x", isk.GetFlags()))
	}

	if addrs := ipAddrsHuman(isk.GetSrcAddr()); addrs != nil {
		sk.SrcAddr = rawJSONValue(addrs)
	}
	if addrs := ipAddrsHuman(isk.GetDstAddr()); addrs != nil {
		sk.DstAddr = rawJSONValue(addrs)
	}

	return json.Marshal(sk)
}

func rawJSONString(s string) json.RawMessage {
	b, _ := json.Marshal(s)
	return b
}

func rawJSONValue(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

func rawJSONHumanMap(m map[uint32]string, value uint32) json.RawMessage {
	if text, ok := m[value]; ok {
		return rawJSONString(text)
	}
	return rawJSONValue(value)
}

func ipAddrsHuman(parts []uint32) []string {
	if len(parts) == 0 {
		return nil
	}
	addr := processIP(parts)
	if addr == "" {
		return nil
	}
	return []string{addr}
}
