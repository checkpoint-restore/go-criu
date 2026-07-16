package crit

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/checkpoint-restore/go-criu/v8/crit/images/pagemap"
)

func TestMemoryReaderRawAndZeroBlocks(t *testing.T) {
	const (
		pageSize = 4096
		pid      = 21
		pagesID  = 4
		base     = 0x20000
	)
	directory := t.TempDir()
	writeTestInventory(t, directory, crtoolsImagesV1_2, CompressionBlock, 2*pageSize)

	entry := presentTestEntry(base, 4)
	entry.Blocks = testPagemapBlocks([]uint32{0, 2 * pageSize}, 2)
	writeTestPagemap(t, directory, pid, pagesID, entry)

	raw := make([]byte, 2*pageSize)
	for index := range raw {
		raw[index] = byte(index%251 + 1)
	}
	copy(raw[173:], "raw-block-needle")
	if err := os.WriteFile(filepath.Join(directory, "pages-4.img"), raw, 0o600); err != nil {
		t.Fatal(err)
	}

	reader, err := NewMemoryReader(directory, pid, pageSize)
	if err != nil {
		t.Fatal(err)
	}
	memory, err := reader.GetMemPages(base, base+4*pageSize)
	if err != nil {
		t.Fatal(err)
	}
	expected := append(make([]byte, 2*pageSize), raw...)
	if !bytes.Equal(memory.Bytes(), expected) {
		t.Fatal("raw and zero block contents do not match")
	}

	matches, err := reader.SearchPattern("raw-block-needle", true, 0, pageSize)
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 1 || matches[0].Vaddr != base+2*pageSize+173 {
		t.Fatalf("unexpected memory search matches: %+v", matches)
	}
}

func TestMemoryReadSessionCachesBlockPayloadOffsets(t *testing.T) {
	const (
		pageSize = 4096
		blocks   = 130
	)
	sizes := make([]uint32, blocks)
	prefix := make([]uint64, blocks+1)
	for index := range sizes {
		sizes[index] = uint32(index%31 + 1)
		prefix[index+1] = prefix[index] + uint64(sizes[index])
	}
	entry := presentTestEntry(0x100000, blocks)
	entry.Blocks = testPagemapBlocks(sizes, 1)
	layer, err := indexMemoryLayer(".", "pagemap-1.img", 1, pageSize, []*pagemap.PagemapEntry{entry})
	if err != nil {
		t.Fatal(err)
	}
	indexed := &layer.entries[0]
	session := &memoryReadSession{}
	for _, index := range []uint64{63, 64, 65, 127, 128} {
		if got := session.blockPayloadOffset(indexed, index); got != prefix[index] {
			t.Fatalf("block %d offset = %d, want %d", index, got, prefix[index])
		}
	}
}

func TestMemoryReaderRejectsRawPayloadLength(t *testing.T) {
	const (
		pageSize = 4096
		pid      = 23
		pagesID  = 6
		base     = 0x40000
	)

	for _, size := range []int{pageSize - 1, pageSize + 1} {
		t.Run(stringID(uint32(size)), func(t *testing.T) {
			directory := t.TempDir()
			writeTestInventory(t, directory, crtoolsImagesV1_1, CompressionOff, 0)
			writeTestPagemap(t, directory, pid, pagesID, presentTestEntry(base, 1))
			if err := os.WriteFile(
				filepath.Join(directory, "pages-6.img"),
				make([]byte, size),
				0o600,
			); err != nil {
				t.Fatal(err)
			}

			reader, err := NewMemoryReader(directory, pid, pageSize)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := reader.GetMemPages(base, base+pageSize); err == nil ||
				!strings.Contains(err.Error(), "describes") {
				t.Fatalf("expected payload length error, got %v", err)
			}
		})
	}
}
