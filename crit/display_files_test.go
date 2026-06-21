package crit

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/checkpoint-restore/go-criu/v8/crit/images/fdinfo"
	sk_inet "github.com/checkpoint-restore/go-criu/v8/crit/images/sk-inet"
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
