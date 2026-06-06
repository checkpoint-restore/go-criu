package crit

import (
	"encoding/json"
	"fmt"

	"github.com/checkpoint-restore/go-criu/v8/crit/images/fdinfo"
	sk_inet "github.com/checkpoint-restore/go-criu/v8/crit/images/sk-inet"
	"google.golang.org/protobuf/proto"
)

// fileEntryJSON is a human-readable view of files.img: to be used with criu show/decode
type fileEntryJSON struct {
	ID   uint32 `json:"id,omitempty"`
	Type string `json:"type,omitempty"`
	Reg  any    `json:"reg,omitempty"`
	Pipe any    `json:"pipe,omitempty"`
	Isk  any    `json:"isk,omitempty"`
	Usk  any    `json:"usk,omitempty"`
	Psk  any    `json:"psk,omitempty"`
	Nlsk any    `json:"nlsk,omitempty"`
	Ext  any    `json:"ext,omitempty"`
}

type inetSkJSON struct {
	ID      uint32   `json:"id"`
	Ino     uint32   `json:"ino"`
	Family  string   `json:"family,omitempty"`
	Type    string   `json:"type,omitempty"`
	Proto   string   `json:"proto,omitempty"`
	State   string   `json:"state,omitempty"`
	SrcPort uint32   `json:"src_port"`
	DstPort uint32   `json:"dst_port"`
	Flags   string   `json:"flags,omitempty"`
	Backlog uint32   `json:"backlog"`
	SrcAddr []string `json:"src_addr,omitempty"`
	DstAddr []string `json:"dst_addr,omitempty"`
}

func marshalFileEntryHuman(msg proto.Message) ([]byte, error) {
	entry, ok := msg.(*fdinfo.FileEntry)
	if !ok {
		return nil, fmt.Errorf("unexpected message type %T", msg)
	}

	out := fileEntryJSON{
		ID:   entry.GetId(),
		Type: entry.GetType().String(),
	}

	switch entry.GetType() {
	case fdinfo.FdTypes_REG:
		if reg := entry.GetReg(); reg != nil {
			out.Reg = reg
		}
	case fdinfo.FdTypes_PIPE:
		if pipe := entry.GetPipe(); pipe != nil {
			out.Pipe = pipe
		}
	case fdinfo.FdTypes_INETSK:
		if isk := entry.GetIsk(); isk != nil {
			out.Isk = inetSkEntryHuman(isk)
		}
	case fdinfo.FdTypes_UNIXSK:
		if usk := entry.GetUsk(); usk != nil {
			out.Usk = usk
		}
	case fdinfo.FdTypes_PACKETSK:
		if psk := entry.GetPsk(); psk != nil {
			out.Psk = psk
		}
	case fdinfo.FdTypes_NETLINKSK:
		if nlsk := entry.GetNlsk(); nlsk != nil {
			out.Nlsk = nlsk
		}
	case fdinfo.FdTypes_EXT:
		if ext := entry.GetExt(); ext != nil {
			out.Ext = ext
		}
	default:
		// Fall back to embedding the raw proto for uncommon types.
		out.Type = entry.GetType().String()
	}

	return json.Marshal(out)
}

func inetSkEntryHuman(isk *sk_inet.InetSkEntry) inetSkJSON {
	out := inetSkJSON{
		ID:      isk.GetId(),
		Ino:     isk.GetIno(),
		Family:  displayInetFamily(isk.GetFamily()),
		Type:    getSkType(isk.GetType()),
		Proto:   getSkProtocol(isk.GetProto()),
		State:   getSkState(tcpState(isk.GetState())),
		SrcPort: isk.GetSrcPort(),
		DstPort: isk.GetDstPort(),
		Flags:   fmt.Sprintf("0x%x", isk.GetFlags()),
		Backlog: isk.GetBacklog(),
		SrcAddr: ipAddrsHuman(isk.GetSrcAddr()),
		DstAddr: ipAddrsHuman(isk.GetDstAddr()),
	}
	return out
}

// displayInetFamily maps AF_INET/AF_INET6 names to match Python CRIT output.
func displayInetFamily(family uint32) string {
	switch getAddressFamily(family) {
	case "IPV4":
		return "INET"
	case "IPV6":
		return "INET6"
	default:
		return getAddressFamily(family)
	}
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
