package crit

import (
	"encoding/base64"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	ipc_desc "github.com/checkpoint-restore/go-criu/v8/crit/images/ipc-desc"
	ipc_shm "github.com/checkpoint-restore/go-criu/v8/crit/images/ipc-shm"
	"github.com/checkpoint-restore/go-criu/v8/magic"
	"google.golang.org/protobuf/proto"
)

func TestDecodeIPCShmInPagemaps(t *testing.T) {
	desc := &ipc_desc.IpcDescEntry{
		Key: proto.Uint32(1), Uid: proto.Uint32(2), Gid: proto.Uint32(3),
		Cuid: proto.Uint32(4), Cgid: proto.Uint32(5), Mode: proto.Uint32(0o600),
		Id: proto.Uint32(6),
	}
	entries := []*ipc_shm.IpcShmEntry{
		{Desc: desc, Size: proto.Uint64(4096), InPagemaps: proto.Bool(true)},
		{Desc: desc, Size: proto.Uint64(3), InPagemaps: proto.Bool(false)},
		{Desc: desc, Size: proto.Uint64(8192), InPagemaps: proto.Bool(true)},
		{Desc: desc, Size: proto.Uint64(4)},
	}
	// Build the wire image independently of the extra-payload encoder. Only
	// the two legacy entries have inline data, padded to four bytes.
	payloads := [][]byte{nil, {'a', 'b', 'c', 0}, nil, {'d', 'e', 'f', 'g'}}
	data := binary.LittleEndian.AppendUint32(nil, uint32(magic.LoadMagic().ByName["IMG_COMMON"]))
	data = binary.LittleEndian.AppendUint32(data, uint32(magic.LoadMagic().ByName["IPCNS_SHM"]))
	for i, entry := range entries {
		encoded, err := proto.Marshal(entry)
		if err != nil {
			t.Fatal(err)
		}
		data = binary.LittleEndian.AppendUint32(data, uint32(len(encoded)))
		data = append(data, encoded...)
		data = append(data, payloads[i]...)
	}
	path := filepath.Join(t.TempDir(), "ipcns-shm.img")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	for _, noPayload := range []bool{false, true} {
		file, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		image, decodeErr := New(file, nil, "", false, noPayload).Decode(&ipc_shm.IpcShmEntry{})
		closeErr := file.Close()
		if decodeErr != nil {
			t.Fatalf("Decode(noPayload=%v): %v", noPayload, decodeErr)
		}
		if closeErr != nil {
			t.Fatal(closeErr)
		}
		if len(image.Entries) != len(entries) {
			t.Fatalf("decoded %d entries, want %d", len(image.Entries), len(entries))
		}
		for i, entry := range image.Entries {
			if !proto.Equal(entry.Message, entries[i]) {
				t.Errorf("entry %d = %v, want %v", i, entry.Message, entries[i])
			}
			if entries[i].GetInPagemaps() {
				if entry.Extra != "" {
					t.Errorf("entry %d has unexpected inline data %q", i, entry.Extra)
				}
			} else if !noPayload {
				want := base64.StdEncoding.EncodeToString(payloads[i][:entries[i].GetSize()])
				if entry.Extra != want {
					t.Errorf("entry %d inline data = %q, want %q", i, entry.Extra, want)
				}
			}
		}
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	image, countErr := New(file, nil, "", false, false).Info()
	closeErr := file.Close()
	if countErr != nil {
		t.Fatal(countErr)
	}
	if closeErr != nil {
		t.Fatal(closeErr)
	}
	if got := image.Entries[0].Extra; got != "4" {
		t.Fatalf("Info() count = %s, want 4", got)
	}
}
