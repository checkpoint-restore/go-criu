package crit

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/checkpoint-restore/go-criu/v8/crit/images/fdinfo"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/fown"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/memfd"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/regfile"
	sk_inet "github.com/checkpoint-restore/go-criu/v8/crit/images/sk-inet"
	sk_opts "github.com/checkpoint-restore/go-criu/v8/crit/images/sk-opts"
	"google.golang.org/protobuf/proto"
)

func TestMarshalFileEntryHumanINETSK(t *testing.T) {
	entry := &fdinfo.FileEntry{
		Id:   proto.Uint32(36),
		Type: fdinfo.FdTypes_INETSK.Enum(),
		Isk: &sk_inet.InetSkEntry{
			Id:      proto.Uint32(36),
			Ino:     proto.Uint32(106886),
			Family:  proto.Uint32(2), // AF_INET
			Type:    proto.Uint32(1), // SOCK_STREAM
			Proto:   proto.Uint32(6), // TCP
			State:   proto.Uint32(7), // TCP_CLOSE
			SrcPort: proto.Uint32(0),
			DstPort: proto.Uint32(0),
			Flags:   proto.Uint32(2),
			Backlog: proto.Uint32(0),
			SrcAddr: []uint32{0},
			DstAddr: []uint32{0},
		},
	}

	data, err := marshalFileEntryHuman(entry)
	if err != nil {
		t.Fatal(err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got["type"] != "INETSK" {
		t.Fatalf("expected INETSK entry, got %#v", got["type"])
	}

	isk, ok := got["isk"].(map[string]any)
	if !ok {
		t.Fatalf("expected isk object, got %#v", got["isk"])
	}

	checks := map[string]any{
		"family":   "INET",
		"type":     "STREAM",
		"proto":    "TCP",
		"state":    "CLOSE",
		"flags":    "0x2",
		"src_port": float64(0),
		"dst_port": float64(0),
	}
	for key, want := range checks {
		if isk[key] != want {
			t.Errorf("isk.%s: want %v, got %v", key, want, isk[key])
		}
	}
	if _, ok := isk["fown"]; ok {
		t.Error("did not expect fown in minimal test fixture")
	}

	srcAddr, ok := isk["src_addr"].([]any)
	if !ok || len(srcAddr) != 1 || srcAddr[0] != "0.0.0.0" {
		t.Errorf("isk.src_addr: want [0.0.0.0], got %v", isk["src_addr"])
	}
	dstAddr, ok := isk["dst_addr"].([]any)
	if !ok || len(dstAddr) != 1 || dstAddr[0] != "0.0.0.0" {
		t.Errorf("isk.dst_addr: want [0.0.0.0], got %v", isk["dst_addr"])
	}
}

func TestMarshalFileEntryHumanINETSKUsesPythonSkDictNames(t *testing.T) {
	isk := marshalHumanInetSk(t, &sk_inet.InetSkEntry{
		Family: proto.Uint32(2),
		Type:   proto.Uint32(1),
		Proto:  proto.Uint32(0),
		State:  proto.Uint32(3),
	})

	checks := map[string]any{
		"family": "INET",
		"type":   "STREAM",
		"proto":  "IP",
		"state":  "SYN_RECV",
	}
	for key, want := range checks {
		if isk[key] != want {
			t.Errorf("isk.%s: want %v, got %v", key, want, isk[key])
		}
	}
}

func TestMarshalFileEntryHumanINETSKSkDictNumericFallback(t *testing.T) {
	isk := marshalHumanInetSk(t, &sk_inet.InetSkEntry{
		Family: proto.Uint32(42),
		Type:   proto.Uint32(4),
		Proto:  proto.Uint32(1),
		State:  proto.Uint32(11),
	})

	checks := map[string]float64{
		"family": 42,
		"type":   4,
		"proto":  1,
		"state":  11,
	}
	for key, want := range checks {
		got, ok := isk[key].(float64)
		if !ok || got != want {
			t.Errorf("isk.%s: want numeric %v, got %#v", key, want, isk[key])
		}
	}
}

func TestMarshalFileEntryHumanINETSKOptsUint64Numbers(t *testing.T) {
	isk := marshalHumanInetSk(t, &sk_inet.InetSkEntry{
		Family: proto.Uint32(2),
		Type:   proto.Uint32(1),
		Proto:  proto.Uint32(6),
		State:  proto.Uint32(7),
		Opts: &sk_opts.SkOptsEntry{
			SoSndbuf:     proto.Uint32(4096),
			SoRcvbuf:     proto.Uint32(8192),
			SoSndTmoSec:  proto.Uint64(1),
			SoSndTmoUsec: proto.Uint64(2),
			SoRcvTmoSec:  proto.Uint64(3),
			SoRcvTmoUsec: proto.Uint64(4),
			SoFilter:     []uint64{5},
		},
	})

	opts, ok := isk["opts"].(map[string]any)
	if !ok {
		t.Fatalf("expected opts object, got %#v", isk["opts"])
	}

	checks := map[string]float64{
		"so_snd_tmo_sec":  1,
		"so_snd_tmo_usec": 2,
		"so_rcv_tmo_sec":  3,
		"so_rcv_tmo_usec": 4,
	}
	for key, want := range checks {
		got, ok := opts[key].(float64)
		if !ok || got != want {
			t.Errorf("opts.%s: want numeric %v, got %#v", key, want, opts[key])
		}
	}

	filter, ok := opts["so_filter"].([]any)
	if !ok || len(filter) != 1 || filter[0] != float64(5) {
		t.Errorf("opts.so_filter: want numeric [5], got %#v", opts["so_filter"])
	}
}

func TestMarshalFileEntryHumanINETSKPreservesFieldOrder(t *testing.T) {
	entry := &fdinfo.FileEntry{
		Id:   proto.Uint32(7),
		Type: fdinfo.FdTypes_INETSK.Enum(),
		Isk: &sk_inet.InetSkEntry{
			Id:      proto.Uint32(7),
			Ino:     proto.Uint32(106886),
			Family:  proto.Uint32(2),
			Type:    proto.Uint32(1),
			Proto:   proto.Uint32(6),
			State:   proto.Uint32(7),
			SrcPort: proto.Uint32(8880),
			DstPort: proto.Uint32(0),
			Flags:   proto.Uint32(2),
			Backlog: proto.Uint32(1),
			SrcAddr: []uint32{0},
			DstAddr: []uint32{0},
		},
	}

	data, err := marshalFileEntryHuman(entry)
	if err != nil {
		t.Fatal(err)
	}

	typeIdx := bytes.Index(data, []byte(`"type"`))
	idIdx := bytes.Index(data, []byte(`"id"`))
	if typeIdx < 0 || idIdx < 0 || typeIdx >= idIdx {
		t.Fatalf("expected entry type before id, got %s", data)
	}

	iskIdx := bytes.Index(data, []byte(`"isk"`))
	iskSection := data[iskIdx:]
	idInIsk := bytes.Index(iskSection, []byte(`"id"`))
	inoInIsk := bytes.Index(iskSection, []byte(`"ino"`))
	familyInIsk := bytes.Index(iskSection, []byte(`"family"`))
	backlogInIsk := bytes.Index(iskSection, []byte(`"backlog"`))
	if idInIsk < 0 || inoInIsk < 0 || familyInIsk < 0 || backlogInIsk < 0 {
		t.Fatalf("missing expected isk fields in %s", iskSection)
	}
	if idInIsk >= inoInIsk || inoInIsk >= familyInIsk || familyInIsk >= backlogInIsk {
		t.Fatalf("expected proto field order in isk, got %s", iskSection)
	}
}

func TestMarshalFileEntryHumanRegUsesProtojson(t *testing.T) {
	entry := &fdinfo.FileEntry{
		Id:   proto.Uint32(1),
		Type: fdinfo.FdTypes_REG.Enum(),
	}

	data, err := marshalFileEntryHuman(entry)
	if err != nil {
		t.Fatal(err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got["type"] != "REG" {
		t.Fatalf("expected REG entry, got %#v", got["type"])
	}
	if _, ok := got["reg"]; ok {
		t.Fatalf("did not expect reg field without reg payload, got %#v", got["reg"])
	}
}

func TestMarshalFileEntryHumanRegMatchesPythonPrettyFields(t *testing.T) {
	entry := &fdinfo.FileEntry{
		Id:   proto.Uint32(1),
		Type: fdinfo.FdTypes_REG.Enum(),
		Reg: &regfile.RegFileEntry{
			Id:    proto.Uint32(1),
			Flags: proto.Uint32(0o00100002),
			Pos:   proto.Uint64(9),
			Fown: &fown.FownEntry{
				Uid:     proto.Uint32(0),
				Euid:    proto.Uint32(0),
				Signum:  proto.Uint32(0),
				PidType: proto.Uint32(0),
				Pid:     proto.Uint32(0),
			},
			Name: proto.String("/tmp/file"),
			Size: proto.Uint64(123),
		},
	}

	data, err := marshalFileEntryHuman(entry)
	if err != nil {
		t.Fatal(err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}

	reg := got["reg"].(map[string]any)
	if reg["flags"] != "O_RDWR | O_LARGEFILE" {
		t.Errorf("reg.flags: want O_RDWR | O_LARGEFILE, got %#v", reg["flags"])
	}
	if reg["pos"] != float64(9) {
		t.Errorf("reg.pos: want numeric 9, got %#v", reg["pos"])
	}
	if reg["size"] != float64(123) {
		t.Errorf("reg.size: want numeric 123, got %#v", reg["size"])
	}
}

func TestMarshalFileEntryHumanMemfdMatchesPythonPrettyFields(t *testing.T) {
	entry := &fdinfo.FileEntry{
		Id:   proto.Uint32(1),
		Type: fdinfo.FdTypes_MEMFD.Enum(),
		Memfd: &memfd.MemfdFileEntry{
			Id:    proto.Uint32(1),
			Flags: proto.Uint32(0),
			Pos:   proto.Uint64(11),
			Fown: &fown.FownEntry{
				Uid:     proto.Uint32(0),
				Euid:    proto.Uint32(0),
				Signum:  proto.Uint32(0),
				PidType: proto.Uint32(0),
				Pid:     proto.Uint32(0),
			},
			InodeId: proto.Uint32(7),
		},
	}

	data, err := marshalFileEntryHuman(entry)
	if err != nil {
		t.Fatal(err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}

	memfd := got["memfd"].(map[string]any)
	if memfd["flags"] != "" {
		t.Errorf("memfd.flags: want empty string, got %#v", memfd["flags"])
	}
	if memfd["pos"] != float64(11) {
		t.Errorf("memfd.pos: want numeric 11, got %#v", memfd["pos"])
	}
}

func TestCriuEntryMarshalJSONHumanizeRegUsesProtojson(t *testing.T) {
	entry := &CriuEntry{
		Humanize: true,
		Message: &fdinfo.FileEntry{
			Id:   proto.Uint32(1),
			Type: fdinfo.FdTypes_REG.Enum(),
		},
	}

	data, err := entry.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if _, ok := got["reg"]; ok {
		t.Fatalf("did not expect reg field for entry without reg payload, got %#v", got["reg"])
	}
}

func TestCriuEntryMarshalJSONHumanizeFileEntry(t *testing.T) {
	entry := &CriuEntry{
		Humanize: true,
		Message: &fdinfo.FileEntry{
			Id:   proto.Uint32(1),
			Type: fdinfo.FdTypes_INETSK.Enum(),
			Isk: &sk_inet.InetSkEntry{
				Family: proto.Uint32(2),
				Type:   proto.Uint32(1),
				Proto:  proto.Uint32(6),
				State:  proto.Uint32(7),
			},
		},
	}

	data, err := entry.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}

	isk := got["isk"].(map[string]any)
	if isk["proto"] != "TCP" {
		t.Errorf("expected humanized proto, got %v", isk["proto"])
	}
}

func TestCriuImageUnmarshalHumanizedFilesEntry(t *testing.T) {
	entries := []*fdinfo.FileEntry{
		{
			Id:   proto.Uint32(1),
			Type: fdinfo.FdTypes_REG.Enum(),
			Reg: &regfile.RegFileEntry{
				Id:    proto.Uint32(1),
				Flags: proto.Uint32(0),
				Pos:   proto.Uint64(9),
				Fown: &fown.FownEntry{
					Uid:     proto.Uint32(0),
					Euid:    proto.Uint32(0),
					Signum:  proto.Uint32(0),
					PidType: proto.Uint32(0),
					Pid:     proto.Uint32(0),
				},
				Name: proto.String("/tmp/file"),
				Size: proto.Uint64(123),
			},
		},
		{
			Id:   proto.Uint32(2),
			Type: fdinfo.FdTypes_REG.Enum(),
			Reg: &regfile.RegFileEntry{
				Id:    proto.Uint32(2),
				Flags: proto.Uint32(0o00100002),
				Pos:   proto.Uint64(10),
				Fown: &fown.FownEntry{
					Uid:     proto.Uint32(0),
					Euid:    proto.Uint32(0),
					Signum:  proto.Uint32(0),
					PidType: proto.Uint32(0),
					Pid:     proto.Uint32(0),
				},
				Name: proto.String("/tmp/file2"),
			},
		},
		{
			Id:   proto.Uint32(3),
			Type: fdinfo.FdTypes_INETSK.Enum(),
			Isk: &sk_inet.InetSkEntry{
				Id:      proto.Uint32(3),
				Ino:     proto.Uint32(106886),
				Family:  proto.Uint32(2),
				Type:    proto.Uint32(1),
				Proto:   proto.Uint32(6),
				State:   proto.Uint32(7),
				SrcPort: proto.Uint32(8880),
				DstPort: proto.Uint32(0),
				Flags:   proto.Uint32(2),
				Backlog: proto.Uint32(1),
				SrcAddr: []uint32{0x0100007f},
				DstAddr: []uint32{0},
				Fown: &fown.FownEntry{
					Uid:     proto.Uint32(0),
					Euid:    proto.Uint32(0),
					Signum:  proto.Uint32(0),
					PidType: proto.Uint32(0),
					Pid:     proto.Uint32(0),
				},
				Opts: &sk_opts.SkOptsEntry{
					SoSndbuf:     proto.Uint32(4096),
					SoRcvbuf:     proto.Uint32(8192),
					SoSndTmoSec:  proto.Uint64(1),
					SoSndTmoUsec: proto.Uint64(2),
					SoRcvTmoSec:  proto.Uint64(3),
					SoRcvTmoUsec: proto.Uint64(4),
					SoFilter:     []uint64{5},
				},
			},
		},
	}

	rawEntries := make([]string, 0, len(entries))
	for _, entry := range entries {
		data, err := marshalFileEntryHuman(entry)
		if err != nil {
			t.Fatal(err)
		}
		rawEntries = append(rawEntries, string(data))
	}

	data := []byte(`{"magic":"FILES","entries":[` + strings.Join(rawEntries, ",") + `]}`)
	img := CriuImage{EntryType: &fdinfo.FileEntry{}}
	if err := json.Unmarshal(data, &img); err != nil {
		t.Fatal(err)
	}

	regZero := img.Entries[0].Message.(*fdinfo.FileEntry).GetReg()
	if regZero.GetFlags() != 0 || regZero.GetPos() != 9 || regZero.GetSize() != 123 {
		t.Fatalf("unexpected zero-flag reg entry: flags=%#o pos=%d size=%d", regZero.GetFlags(), regZero.GetPos(), regZero.GetSize())
	}

	regNames := img.Entries[1].Message.(*fdinfo.FileEntry).GetReg()
	if regNames.GetFlags() != 0o00100002 {
		t.Fatalf("reg flags: want %#o, got %#o", 0o00100002, regNames.GetFlags())
	}

	isk := img.Entries[2].Message.(*fdinfo.FileEntry).GetIsk()
	checks := map[string]uint32{
		"family": isk.GetFamily(),
		"type":   isk.GetType(),
		"proto":  isk.GetProto(),
		"state":  isk.GetState(),
		"flags":  isk.GetFlags(),
	}
	wants := map[string]uint32{
		"family": 2,
		"type":   1,
		"proto":  6,
		"state":  7,
		"flags":  2,
	}
	for field, got := range checks {
		if got != wants[field] {
			t.Fatalf("isk.%s: want %d, got %d", field, wants[field], got)
		}
	}
	if got := processIP(isk.GetSrcAddr()); got != "127.0.0.1" {
		t.Fatalf("isk.src_addr: want 127.0.0.1, got %s", got)
	}
	if got := isk.GetOpts().GetSoFilter(); len(got) != 1 || got[0] != 5 {
		t.Fatalf("opts.so_filter: want [5], got %#v", got)
	}
}

func marshalHumanInetSk(t *testing.T, isk *sk_inet.InetSkEntry) map[string]any {
	t.Helper()

	entry := &fdinfo.FileEntry{
		Id:   proto.Uint32(1),
		Type: fdinfo.FdTypes_INETSK.Enum(),
		Isk:  isk,
	}

	data, err := marshalFileEntryHuman(entry)
	if err != nil {
		t.Fatal(err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}

	iskJSON, ok := got["isk"].(map[string]any)
	if !ok {
		t.Fatalf("expected isk object, got %#v", got["isk"])
	}
	return iskJSON
}
