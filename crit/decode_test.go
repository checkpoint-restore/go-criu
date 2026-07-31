package crit

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
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

func TestGhostFileChunkRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ghost-file.img")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	first := []byte("first chunk payload is deliberately longer than its protobuf metadata")
	second := []byte("second chunk")
	image := &CriuImage{
		Magic: "GHOST_FILE",
		Entries: []*CriuEntry{
			{
				Message: &ghost_file.GhostFileEntry{
					Uid: proto.Uint32(1000), Gid: proto.Uint32(1000),
					Mode: proto.Uint32(0o100600), Chunks: proto.Bool(true),
					Size: proto.Uint64(4096 + uint64(len(second))),
				},
			},
			{
				Message: &ghost_file.GhostChunkEntry{Off: proto.Uint64(0), Len: proto.Uint64(uint64(len(first)))},
				Extra:   base64.StdEncoding.EncodeToString(first),
			},
			{
				Message: &ghost_file.GhostChunkEntry{Off: proto.Uint64(4096), Len: proto.Uint64(uint64(len(second)))},
				Extra:   base64.StdEncoding.EncodeToString(second),
			},
		},
	}
	jsonImage, err := json.Marshal(image)
	if err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	parsed := &CriuImage{}
	if err := json.Unmarshal(jsonImage, parsed); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	image = parsed
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
	if len(decoded.Entries) != len(image.Entries) {
		t.Fatalf("decoded entries = %d, want %d", len(decoded.Entries), len(image.Entries))
	}
	for index, want := range [][]byte{first, second} {
		got, err := base64.StdEncoding.DecodeString(decoded.Entries[index+1].Extra)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != string(want) {
			t.Fatalf("decoded chunk %d = %q, want %q", index, got, want)
		}
	}
}

func TestInfoCountsChunkedGhostEntries(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ghost-file.img")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	image := &CriuImage{
		Magic: "GHOST_FILE",
		Entries: []*CriuEntry{
			{
				Message: &ghost_file.GhostFileEntry{
					Uid: proto.Uint32(1000), Gid: proto.Uint32(1000),
					Mode: proto.Uint32(0o100600), Chunks: proto.Bool(true),
					Size: proto.Uint64(9),
				},
			},
			{
				Message: &ghost_file.GhostChunkEntry{Off: proto.Uint64(0), Len: proto.Uint64(4)},
				Extra:   base64.StdEncoding.EncodeToString([]byte("head")),
			},
			{
				Message: &ghost_file.GhostChunkEntry{Off: proto.Uint64(4), Len: proto.Uint64(5)},
				Extra:   base64.StdEncoding.EncodeToString([]byte("tail!")),
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
	if got := counted.Entries[0].Extra; got != "3" {
		t.Fatalf("Info() count = %s, want 3", got)
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
		t.Fatalf("Info() truncated chunk error = %v, want io.ErrUnexpectedEOF", countErr)
	}
	if closeErr != nil {
		t.Fatal(closeErr)
	}
}

func TestEncodeGhostFileRejectsInconsistentLayout(t *testing.T) {
	primary := func(chunks bool, size *uint64) *CriuEntry {
		return &CriuEntry{Message: &ghost_file.GhostFileEntry{
			Uid: proto.Uint32(1000), Gid: proto.Uint32(1000), Mode: proto.Uint32(0o100600),
			Chunks: proto.Bool(chunks), Size: size,
		}}
	}
	chunk := func(length uint64, data []byte) *CriuEntry {
		return &CriuEntry{
			Message: &ghost_file.GhostChunkEntry{Off: proto.Uint64(0), Len: proto.Uint64(length)},
			Extra:   base64.StdEncoding.EncodeToString(data),
		}
	}
	for _, test := range []struct {
		name    string
		entries []*CriuEntry
	}{
		{name: "missing primary"},
		{name: "invalid primary", entries: []*CriuEntry{chunk(0, nil)}},
		{name: "chunks without flag", entries: []*CriuEntry{primary(false, nil), chunk(1, []byte{1})}},
		{name: "missing file size", entries: []*CriuEntry{primary(true, nil)}},
		{
			name: "primary chunk payload",
			entries: []*CriuEntry{
				func() *CriuEntry {
					entry := primary(true, proto.Uint64(0))
					entry.Extra = base64.StdEncoding.EncodeToString([]byte{1})
					return entry
				}(),
			},
		},
		{name: "chunk length mismatch", entries: []*CriuEntry{primary(true, proto.Uint64(2)), chunk(2, []byte{1})}},
		{name: "chunk out of bounds", entries: []*CriuEntry{primary(true, proto.Uint64(1)), chunk(2, []byte{1, 2})}},
		{
			name: "chunk bounds overflow",
			entries: []*CriuEntry{
				primary(true, proto.Uint64(^uint64(0))),
				{
					Message: &ghost_file.GhostChunkEntry{
						Off: proto.Uint64(^uint64(0)), Len: proto.Uint64(1),
					},
					Extra: base64.StdEncoding.EncodeToString([]byte{1}),
				},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			file, err := os.Create(filepath.Join(t.TempDir(), "ghost-file.img"))
			if err != nil {
				t.Fatal(err)
			}
			encodeErr := New(nil, file, "", false, false).Encode(&CriuImage{Magic: "GHOST_FILE", Entries: test.entries})
			closeErr := file.Close()
			if encodeErr == nil {
				t.Fatal("Encode() accepted an inconsistent ghost image")
			}
			if closeErr != nil {
				t.Fatal(closeErr)
			}
		})
	}
}

func TestUnmarshalGhostFileRejectsMissingPrimary(t *testing.T) {
	var image CriuImage
	if err := json.Unmarshal([]byte(`{"magic":"GHOST_FILE","entries":[]}`), &image); err == nil {
		t.Fatal("UnmarshalJSON() accepted a ghost image without a primary entry")
	}
}
