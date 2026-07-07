package crit

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/netip"
	"strconv"
	"strings"

	"github.com/checkpoint-restore/go-criu/v8/crit/images/fdinfo"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/memfd"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/regfile"
	sk_inet "github.com/checkpoint-restore/go-criu/v8/crit/images/sk-inet"
	sk_opts "github.com/checkpoint-restore/go-criu/v8/crit/images/sk-opts"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// fileEntryProtojson is used for files.img entries before optional field patches.
var fileEntryProtojson = protojson.MarshalOptions{
	UseProtoNames: true,
	AllowPartial:  true,
}

// fileEntryPayloadHumanizer humanizes one FileEntry payload sub-message (reg, isk, ...).
type fileEntryPayloadHumanizer func(proto.Message) ([]byte, error)

type fileEntryHumanSpec struct {
	jsonField string
	payload   func(*fdinfo.FileEntry) proto.Message
	humanize  fileEntryPayloadHumanizer
}

// fileEntryHumanizers maps fd_types to payload humanizers. Add an entry here when
// implementing humanized output for a new files.img fd type.
var fileEntryHumanizers = map[fdinfo.FdTypes]fileEntryHumanSpec{
	fdinfo.FdTypes_REG: {
		jsonField: "reg",
		payload:   regFilePayload,
		humanize:  humanizeRegPayload,
	},
	fdinfo.FdTypes_INETSK: {
		jsonField: "isk",
		payload:   inetSkPayload,
		humanize:  humanizeInetSkPayload,
	},
	fdinfo.FdTypes_MEMFD: {
		jsonField: "memfd",
		payload:   memfdFilePayload,
		humanize:  humanizeMemfdPayload,
	},
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
	rfileFlagHumanMap = []flagHumanEntry{
		{name: "O_WRONLY", value: 0o00000001},
		{name: "O_RDWR", value: 0o00000002},
		{name: "O_CREAT", value: 0o00000100},
		{name: "O_EXCL", value: 0o00000200},
		{name: "O_NOCTTY", value: 0o00000400},
		{name: "O_TRUNC", value: 0o00001000},
		{name: "O_APPEND", value: 0o00002000},
		{name: "O_NONBLOCK", value: 0o00004000},
		{name: "O_DSYNC", value: 0o00010000},
		{name: "FASYNC", value: 0o00020000},
		{name: "O_DIRECT", value: 0o00040000},
		{name: "O_LARGEFILE", value: 0o00100000},
		{name: "O_DIRECTORY", value: 0o00200000},
		{name: "O_NOFOLLOW", value: 0o00400000},
		{name: "O_NOATIME", value: 0o01000000},
		{name: "O_CLOEXEC", value: 0o02000000},
	}
)

type flagHumanEntry struct {
	name  string
	value uint64
}

func regFilePayload(entry *fdinfo.FileEntry) proto.Message {
	if payload := entry.GetReg(); payload != nil {
		return payload
	}
	return nil
}

func inetSkPayload(entry *fdinfo.FileEntry) proto.Message {
	if payload := entry.GetIsk(); payload != nil {
		return payload
	}
	return nil
}

func memfdFilePayload(entry *fdinfo.FileEntry) proto.Message {
	if payload := entry.GetMemfd(); payload != nil {
		return payload
	}
	return nil
}

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

// regFileJSON mirrors reg_file_entry field order from regfile.proto.
type regFileJSON struct {
	Id                json.RawMessage `json:"id"`
	Flags             json.RawMessage `json:"flags"`
	Pos               json.RawMessage `json:"pos"`
	Fown              json.RawMessage `json:"fown"`
	Name              json.RawMessage `json:"name"`
	MntId             json.RawMessage `json:"mnt_id,omitempty"`
	Size              json.RawMessage `json:"size,omitempty"`
	Ext               json.RawMessage `json:"ext,omitempty"`
	Mode              json.RawMessage `json:"mode,omitempty"`
	BuildId           json.RawMessage `json:"build_id,omitempty"`
	Checksum          json.RawMessage `json:"checksum,omitempty"`
	ChecksumConfig    json.RawMessage `json:"checksum_config,omitempty"`
	ChecksumParameter json.RawMessage `json:"checksum_parameter,omitempty"`
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

// skOptsJSON mirrors sk_opts_entry field order from sk-opts.proto.
type skOptsJSON struct {
	SoSndbuf     json.RawMessage `json:"so_sndbuf"`
	SoRcvbuf     json.RawMessage `json:"so_rcvbuf"`
	SoSndTmoSec  json.RawMessage `json:"so_snd_tmo_sec"`
	SoSndTmoUsec json.RawMessage `json:"so_snd_tmo_usec"`
	SoRcvTmoSec  json.RawMessage `json:"so_rcv_tmo_sec"`
	SoRcvTmoUsec json.RawMessage `json:"so_rcv_tmo_usec"`
	Reuseaddr    json.RawMessage `json:"reuseaddr,omitempty"`
	SoPriority   json.RawMessage `json:"so_priority,omitempty"`
	SoRcvlowat   json.RawMessage `json:"so_rcvlowat,omitempty"`
	SoMark       json.RawMessage `json:"so_mark,omitempty"`
	SoPasscred   json.RawMessage `json:"so_passcred,omitempty"`
	SoPasssec    json.RawMessage `json:"so_passsec,omitempty"`
	SoDontroute  json.RawMessage `json:"so_dontroute,omitempty"`
	SoNoCheck    json.RawMessage `json:"so_no_check,omitempty"`
	SoBoundDev   json.RawMessage `json:"so_bound_dev,omitempty"`
	SoFilter     json.RawMessage `json:"so_filter,omitempty"`
	SoReuseport  json.RawMessage `json:"so_reuseport,omitempty"`
	SoBroadcast  json.RawMessage `json:"so_broadcast,omitempty"`
	SoKeepalive  json.RawMessage `json:"so_keepalive,omitempty"`
	TcpKeepcnt   json.RawMessage `json:"tcp_keepcnt,omitempty"`
	TcpKeepidle  json.RawMessage `json:"tcp_keepidle,omitempty"`
	TcpKeepintvl json.RawMessage `json:"tcp_keepintvl,omitempty"`
	SoOobinline  json.RawMessage `json:"so_oobinline,omitempty"`
	SoLinger     json.RawMessage `json:"so_linger,omitempty"`
	SoBufLock    json.RawMessage `json:"so_buf_lock,omitempty"`
}

// memfdFileJSON mirrors memfd_file_entry field order from memfd.proto.
type memfdFileJSON struct {
	Id      json.RawMessage `json:"id"`
	Flags   json.RawMessage `json:"flags"`
	Pos     json.RawMessage `json:"pos"`
	Fown    json.RawMessage `json:"fown"`
	InodeId json.RawMessage `json:"inode_id"`
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

func humanizeRegPayload(msg proto.Message) ([]byte, error) {
	reg, ok := msg.(*regfile.RegFileEntry)
	if !ok {
		return nil, fmt.Errorf("humanizeRegPayload: unexpected type %T", msg)
	}

	data, err := fileEntryProtojson.Marshal(reg)
	if err != nil {
		return nil, err
	}

	var rf regFileJSON
	if err := json.Unmarshal(data, &rf); err != nil {
		return nil, err
	}

	if reg.Flags != nil {
		rf.Flags = rawJSONString(formatFlags(uint64(reg.GetFlags()), rfileFlagHumanMap))
	}
	if reg.Pos != nil {
		rf.Pos = rawJSONValue(reg.GetPos())
	}
	if reg.Size != nil {
		rf.Size = rawJSONValue(reg.GetSize())
	}

	return json.Marshal(rf)
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
	if opts := isk.GetOpts(); opts != nil {
		humanOpts, err := humanizeSkOpts(opts)
		if err != nil {
			return nil, err
		}
		sk.Opts = humanOpts
	}

	return json.Marshal(sk)
}

func humanizeSkOpts(opts *sk_opts.SkOptsEntry) (json.RawMessage, error) {
	data, err := fileEntryProtojson.Marshal(opts)
	if err != nil {
		return nil, err
	}

	var so skOptsJSON
	if err := json.Unmarshal(data, &so); err != nil {
		return nil, err
	}

	if opts.SoSndTmoSec != nil {
		so.SoSndTmoSec = rawJSONValue(opts.GetSoSndTmoSec())
	}
	if opts.SoSndTmoUsec != nil {
		so.SoSndTmoUsec = rawJSONValue(opts.GetSoSndTmoUsec())
	}
	if opts.SoRcvTmoSec != nil {
		so.SoRcvTmoSec = rawJSONValue(opts.GetSoRcvTmoSec())
	}
	if opts.SoRcvTmoUsec != nil {
		so.SoRcvTmoUsec = rawJSONValue(opts.GetSoRcvTmoUsec())
	}
	if len(opts.GetSoFilter()) > 0 {
		so.SoFilter = rawJSONValue(opts.GetSoFilter())
	}

	return json.Marshal(so)
}

func humanizeMemfdPayload(msg proto.Message) ([]byte, error) {
	m, ok := msg.(*memfd.MemfdFileEntry)
	if !ok {
		return nil, fmt.Errorf("humanizeMemfdPayload: unexpected type %T", msg)
	}

	data, err := fileEntryProtojson.Marshal(m)
	if err != nil {
		return nil, err
	}

	var mf memfdFileJSON
	if err := json.Unmarshal(data, &mf); err != nil {
		return nil, err
	}

	if m.Flags != nil {
		mf.Flags = rawJSONString(formatFlags(uint64(m.GetFlags()), rfileFlagHumanMap))
	}
	if m.Pos != nil {
		mf.Pos = rawJSONValue(m.GetPos())
	}

	return json.Marshal(mf)
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

func formatFlags(value uint64, flags []flagHumanEntry) string {
	var names []string
	for _, flag := range flags {
		if value&flag.value != 0 {
			names = append(names, flag.name)
			value &^= flag.value
		}
	}
	if value != 0 {
		names = append(names, fmt.Sprintf("0x%x", value))
	}
	return strings.Join(names, " | ")
}

func normalizeFileEntryJSON(data []byte) ([]byte, error) {
	var entry map[string]json.RawMessage
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}

	for _, field := range []string{"reg", "memfd"} {
		if err := normalizeFileFlagsPayload(entry, field); err != nil {
			return nil, err
		}
	}
	if err := normalizeInetSkPayload(entry); err != nil {
		return nil, err
	}

	return json.Marshal(entry)
}

func normalizeFileFlagsPayload(entry map[string]json.RawMessage, field string) error {
	payload, ok, err := jsonObjectField(entry, field)
	if err != nil || !ok {
		return err
	}

	if err := normalizeFlagField(payload, "flags", rfileFlagHumanMap); err != nil {
		return err
	}
	if err := normalizeUint64Fields(payload, "pos", "size"); err != nil {
		return err
	}

	entry[field], err = json.Marshal(payload)
	return err
}

func normalizeInetSkPayload(entry map[string]json.RawMessage) error {
	isk, ok, err := jsonObjectField(entry, "isk")
	if err != nil || !ok {
		return err
	}

	if err := normalizeHumanMapField(isk, "family", inetSkFamilyHumanMap); err != nil {
		return err
	}
	if err := normalizeHumanMapField(isk, "type", inetSkTypeHumanMap); err != nil {
		return err
	}
	if err := normalizeHumanMapField(isk, "proto", inetSkProtocolHumanMap); err != nil {
		return err
	}
	if err := normalizeHumanMapField(isk, "state", inetSkStateHumanMap); err != nil {
		return err
	}
	if err := normalizeUint32StringField(isk, "flags"); err != nil {
		return err
	}
	if err := normalizeIPAddrField(isk, "src_addr"); err != nil {
		return err
	}
	if err := normalizeIPAddrField(isk, "dst_addr"); err != nil {
		return err
	}

	opts, ok, err := jsonObjectField(isk, "opts")
	if err != nil {
		return err
	}
	if ok {
		if err := normalizeUint64Fields(opts,
			"so_snd_tmo_sec",
			"so_snd_tmo_usec",
			"so_rcv_tmo_sec",
			"so_rcv_tmo_usec",
			"so_filter",
		); err != nil {
			return err
		}
		isk["opts"], err = json.Marshal(opts)
		if err != nil {
			return err
		}
	}

	entry["isk"], err = json.Marshal(isk)
	return err
}

func jsonObjectField(obj map[string]json.RawMessage, field string) (map[string]json.RawMessage, bool, error) {
	data, ok := obj[field]
	if !ok {
		return nil, false, nil
	}

	var child map[string]json.RawMessage
	if err := json.Unmarshal(data, &child); err != nil {
		return nil, false, err
	}
	return child, true, nil
}

func normalizeFlagField(obj map[string]json.RawMessage, field string, flags []flagHumanEntry) error {
	text, ok := jsonStringField(obj, field)
	if !ok {
		return nil
	}

	value, err := parseFlags(text, flags)
	if err != nil {
		return err
	}
	obj[field] = rawJSONValue(value)
	return nil
}

func normalizeHumanMapField(obj map[string]json.RawMessage, field string, values map[uint32]string) error {
	text, ok := jsonStringField(obj, field)
	if !ok {
		return nil
	}

	value, err := parseHumanMapValue(text, values)
	if err != nil {
		return err
	}
	obj[field] = rawJSONValue(value)
	return nil
}

func normalizeUint32StringField(obj map[string]json.RawMessage, field string) error {
	text, ok := jsonStringField(obj, field)
	if !ok {
		return nil
	}

	value, err := strconv.ParseUint(text, 0, 32)
	if err != nil {
		return err
	}
	obj[field] = rawJSONValue(uint32(value))
	return nil
}

func normalizeIPAddrField(obj map[string]json.RawMessage, field string) error {
	data, ok := obj[field]
	if !ok {
		return nil
	}

	if !isJSONStringArray(data) {
		return nil
	}

	var addrs []string
	if err := json.Unmarshal(data, &addrs); err != nil {
		return err
	}
	if len(addrs) != 1 {
		return fmt.Errorf("%s: expected one humanized IP address, got %d", field, len(addrs))
	}

	parts, err := ipParts(addrs[0])
	if err != nil {
		return err
	}
	obj[field] = rawJSONValue(parts)
	return nil
}

func isJSONStringArray(data json.RawMessage) bool {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) < 2 || trimmed[0] != '[' {
		return false
	}
	trimmed = bytes.TrimSpace(trimmed[1:])
	return len(trimmed) > 0 && trimmed[0] == '"'
}

func normalizeUint64Fields(obj map[string]json.RawMessage, fields ...string) error {
	for _, field := range fields {
		data, ok := obj[field]
		if !ok {
			continue
		}
		normalized, changed, err := normalizeUint64JSON(data)
		if err != nil {
			return err
		}
		if changed {
			obj[field] = normalized
		}
	}
	return nil
}

func normalizeUint64JSON(data json.RawMessage) (json.RawMessage, bool, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()

	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, false, err
	}

	switch typed := value.(type) {
	case json.Number:
		text, err := uint64Text(typed)
		if err != nil {
			return nil, false, err
		}
		return rawJSONString(text), true, nil
	case []any:
		values := make([]string, 0, len(typed))
		changed := false
		for _, item := range typed {
			switch item := item.(type) {
			case json.Number:
				text, err := uint64Text(item)
				if err != nil {
					return nil, false, err
				}
				values = append(values, text)
				changed = true
			case string:
				values = append(values, item)
			default:
				return nil, false, fmt.Errorf("unexpected uint64 JSON value %T", item)
			}
		}
		if !changed {
			return nil, false, nil
		}
		return rawJSONValue(values), true, nil
	default:
		return nil, false, nil
	}
}

func uint64Text(value json.Number) (string, error) {
	if _, err := strconv.ParseUint(value.String(), 10, 64); err != nil {
		return "", err
	}
	return value.String(), nil
}

func jsonStringField(obj map[string]json.RawMessage, field string) (string, bool) {
	data, ok := obj[field]
	if !ok {
		return "", false
	}

	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return "", false
	}
	return value, true
}

func parseFlags(text string, flags []flagHumanEntry) (uint64, error) {
	if text == "" {
		return 0, nil
	}

	names := make(map[string]uint64, len(flags))
	for _, flag := range flags {
		names[flag.name] = flag.value
	}

	var value uint64
	for _, part := range strings.Split(text, "|") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if flag, ok := names[part]; ok {
			value |= flag
			continue
		}
		unknown, err := strconv.ParseUint(part, 0, 64)
		if err != nil {
			return 0, fmt.Errorf("unknown flag %q", part)
		}
		value |= unknown
	}
	return value, nil
}

func parseHumanMapValue(text string, values map[uint32]string) (uint32, error) {
	for value, name := range values {
		if text == name {
			return value, nil
		}
	}

	value, err := strconv.ParseUint(text, 0, 32)
	if err != nil {
		return 0, fmt.Errorf("unknown humanized value %q", text)
	}
	return uint32(value), nil
}

func ipParts(text string) ([]uint32, error) {
	addr, err := netip.ParseAddr(text)
	if err != nil {
		return nil, err
	}

	if addr.Is4() {
		ip := addr.As4()
		return []uint32{binary.LittleEndian.Uint32(ip[:])}, nil
	}

	ip := addr.As16()
	parts := make([]uint32, 4)
	for i := range parts {
		parts[i] = binary.LittleEndian.Uint32(ip[i*4:])
	}
	return parts, nil
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
