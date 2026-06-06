package crit

import (
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

	srcAddr, ok := isk["src_addr"].([]any)
	if !ok || len(srcAddr) != 1 || srcAddr[0] != "0.0.0.0" {
		t.Errorf("isk.src_addr: want [0.0.0.0], got %v", isk["src_addr"])
	}
	dstAddr, ok := isk["dst_addr"].([]any)
	if !ok || len(dstAddr) != 1 || dstAddr[0] != "0.0.0.0" {
		t.Errorf("isk.dst_addr: want [0.0.0.0], got %v", isk["dst_addr"])
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
