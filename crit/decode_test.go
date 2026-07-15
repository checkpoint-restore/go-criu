package crit

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	ghost_file "github.com/checkpoint-restore/go-criu/v8/crit/images/ghost-file"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/inventory"
	pipe_data "github.com/checkpoint-restore/go-criu/v8/crit/images/pipe-data"
	"github.com/checkpoint-restore/go-criu/v8/magic"
	"google.golang.org/protobuf/proto"
)

func TestDecodeRejectsTruncatedDefaultEntry(t *testing.T) {
	magicValue := uint32(magic.LoadMagic().ByName["INVENTORY"])
	magicBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(magicBytes, magicValue)

	tests := []struct {
		name string
		data []byte
	}{
		{name: "entry header", data: append(append([]byte{}, magicBytes...), 1, 0)},
		{name: "entry payload", data: append(append(append([]byte{}, magicBytes...), 4, 0, 0, 0), 8, 1)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "inventory.img")
			if err := os.WriteFile(path, test.data, 0o600); err != nil {
				t.Fatal(err)
			}
			file, err := os.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = file.Close() }()

			if _, err := New(file, nil, "", false, false).Decode(&inventory.InventoryEntry{}); !errors.Is(err, io.ErrUnexpectedEOF) {
				t.Fatalf("Decode() error = %v, want io.ErrUnexpectedEOF", err)
			}
		})
	}
}

func TestInfoRejectsTruncatedEntryPayload(t *testing.T) {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint32(data[:4], uint32(magic.LoadMagic().ByName["INVENTORY"]))
	binary.LittleEndian.PutUint32(data[4:], 1)
	path := filepath.Join(t.TempDir(), "inventory.img")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = file.Close() }()

	if _, err := New(file, nil, "", false, false).Info(); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("Info() error = %v, want io.ErrUnexpectedEOF", err)
	}
}

func TestInfoCountsExtraPayloadEntries(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pipes-data.img")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	image := &CriuImage{
		Magic: "PIPES_DATA",
		Entries: []*CriuEntry{
			{
				Message: &pipe_data.PipeDataEntry{PipeId: proto.Uint32(1), Bytes: proto.Uint32(3)},
				Extra:   base64.StdEncoding.EncodeToString([]byte("one")),
			},
			{
				Message: &pipe_data.PipeDataEntry{PipeId: proto.Uint32(2), Bytes: proto.Uint32(3)},
				Extra:   base64.StdEncoding.EncodeToString([]byte("two")),
			},
		},
	}
	if err := New(nil, file, "", false, false).Encode(image); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	file, err = os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	counted, countErr := New(file, nil, "", false, false).Info()
	closeErr := file.Close()
	if countErr != nil {
		t.Fatal(countErr)
	}
	if closeErr != nil {
		t.Fatal(closeErr)
	}
	if got := counted.Entries[0].Extra; got != "2" {
		t.Fatalf("Info() count = %s, want 2", got)
	}

	if err := os.Truncate(path, info.Size()-1); err != nil {
		t.Fatal(err)
	}
	file, err = os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	_, countErr = New(file, nil, "", false, false).Info()
	closeErr = file.Close()
	if !errors.Is(countErr, io.ErrUnexpectedEOF) {
		t.Fatalf("Info() truncated extra error = %v, want io.ErrUnexpectedEOF", countErr)
	}
	if closeErr != nil {
		t.Fatal(closeErr)
	}
}

func TestDecodeRejectsTruncatedExtraPayload(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pipes-data.img")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	image := &CriuImage{
		Magic: "PIPES_DATA",
		Entries: []*CriuEntry{{
			Message: &pipe_data.PipeDataEntry{PipeId: proto.Uint32(1), Bytes: proto.Uint32(4)},
			Extra:   base64.StdEncoding.EncodeToString([]byte{1, 2, 3, 4}),
		}},
	}
	if err := New(nil, file, "", false, false).Encode(image); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.Truncate(path, info.Size()-1); err != nil {
		t.Fatal(err)
	}

	for _, noPayload := range []bool{false, true} {
		file, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		_, decodeErr := New(file, nil, "", false, noPayload).Decode(&pipe_data.PipeDataEntry{})
		closeErr := file.Close()
		if !errors.Is(decodeErr, io.ErrUnexpectedEOF) {
			t.Fatalf("Decode(noPayload=%v) error = %v, want io.ErrUnexpectedEOF", noPayload, decodeErr)
		}
		if closeErr != nil {
			t.Fatal(closeErr)
		}
	}
}

func TestDecodeGhostFilePayload(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ghost-file.img")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	want := []byte("ghost payload")
	image := &CriuImage{
		Magic: "GHOST_FILE",
		Entries: []*CriuEntry{{
			Message: &ghost_file.GhostFileEntry{
				Uid: proto.Uint32(1000), Gid: proto.Uint32(1000), Mode: proto.Uint32(0o100600),
			},
			Extra: base64.StdEncoding.EncodeToString(want),
		}},
	}
	if err := New(nil, file, "", false, false).Encode(image); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	file, err = os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = file.Close() }()

	decoded, err := New(file, nil, "", false, false).Decode(&ghost_file.GhostFileEntry{})
	if err != nil {
		t.Fatal(err)
	}
	if len(decoded.Entries) != 1 {
		t.Fatalf("decoded entries = %d, want 1", len(decoded.Entries))
	}
	got, err := base64.StdEncoding.DecodeString(decoded.Entries[0].Extra)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatalf("decoded ghost payload = %q, want %q", got, want)
	}
}
