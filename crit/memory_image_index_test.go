package crit

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/checkpoint-restore/go-criu/v8/crit/images/pagemap"
)

func TestMemoryBlockIndex(t *testing.T) {
	const (
		pageSize = uint64(4096)
		base     = uint64(0x100000)
		blocks   = 130
	)
	sizes := make([]uint32, blocks)
	for index := range sizes {
		sizes[index] = uint32(index%31 + 1)
	}
	for _, index := range []int{63, 64, 65} {
		sizes[index] = uint32(pageSize)
	}
	entry := presentTestEntry(base, blocks)
	entry.Blocks = testPagemapBlocks(sizes, 1)
	layer, err := indexMemoryLayer(".", "pagemap-1.img", 1, int(pageSize), []*pagemap.PagemapEntry{entry})
	if err != nil {
		t.Fatal(err)
	}
	indexed := &layer.entries[0]
	if len(indexed.blockPayloadOffsets) != 3 {
		t.Fatalf("sparse index entries = %d, want 3", len(indexed.blockPayloadOffsets))
	}

	prefix := make([]uint64, blocks+1)
	for index, size := range sizes {
		prefix[index+1] = prefix[index] + uint64(size)
	}
	for _, index := range []uint64{63, 64, 65, 127, 128} {
		if got := indexed.blockPayloadOffset(index); got != prefix[index] {
			t.Fatalf("block %d offset = %d, want %d", index, got, prefix[index])
		}
	}
	entry.Blocks.BlockSizes[64] = 1
	if _, _, storedSize, ok := indexed.blockLayout(64); !ok || storedSize != pageSize {
		t.Fatalf("validated block metadata changed through exposed protobuf: size=%d, ok=%v", storedSize, ok)
	}
}

func TestMemoryLayerInventoryCompatibilityAndDuplicateStats(t *testing.T) {
	const (
		pageSize = 4096
		base     = 0x28000
	)

	t.Run("raw legacy extraction without inventory", func(t *testing.T) {
		directory := t.TempDir()
		entry := presentTestEntry(base, 1)
		writeTestPagemap(t, directory, 1, 9, entry)
		if _, err := NewMemoryReader(directory, 1, pageSize); err != nil {
			t.Fatalf("raw pagemap without inventory was rejected: %v", err)
		}
	})

	t.Run("compressed metadata requires inventory", func(t *testing.T) {
		directory := t.TempDir()
		entry := presentTestEntry(base, 1)
		entry.Blocks = testPagemapBlocks([]uint32{0}, 1)
		writeTestPagemap(t, directory, 1, 9, entry)
		_, err := NewMemoryReader(directory, 1, pageSize)
		if err == nil || !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("expected missing inventory error, got %v", err)
		}
	})

	t.Run("inventory block size must match entry geometry", func(t *testing.T) {
		directory := t.TempDir()
		writeTestInventory(t, directory, crtoolsImagesV1_2, CompressionBlock, pageSize)
		entry := presentTestEntry(base, 2)
		entry.Blocks = testPagemapBlocks([]uint32{0}, 2)
		writeTestPagemap(t, directory, 1, 9, entry)
		if _, err := NewMemoryReader(directory, 1, pageSize); err == nil ||
			!strings.Contains(err.Error(), "conflicts") {
			t.Fatalf("expected compression block-size conflict, got %v", err)
		}
	})

	t.Run("missing informational block size accepts entry geometry", func(t *testing.T) {
		directory := t.TempDir()
		writeTestInventory(t, directory, crtoolsImagesV1_2, CompressionBlock, 0)
		entry := presentTestEntry(base, 2)
		entry.Blocks = testPagemapBlocks([]uint32{0}, 2)
		writeTestPagemap(t, directory, 1, 9, entry)

		if _, err := NewMemoryReader(directory, 1, pageSize); err != nil {
			t.Fatalf("missing informational block size was rejected: %v", err)
		}
	})

	t.Run("block inventory requires matching block size", func(t *testing.T) {
		directory := t.TempDir()
		writeTestInventory(t, directory, crtoolsImagesV1_2, CompressionBlock, 2*pageSize)
		entry := presentTestEntry(base, 2)
		entry.Blocks = testPagemapBlocks([]uint32{0, 0}, 1)
		writeTestPagemap(t, directory, 1, 9, entry)

		if _, err := NewMemoryReader(directory, 1, pageSize); err == nil ||
			!strings.Contains(err.Error(), "block size 4096 conflicts with inventory block size 8192") {
			t.Fatalf("expected block-size conflict, got %v", err)
		}
	})

	t.Run("aligned payload requires compression-enabled inventory", func(t *testing.T) {
		directory := t.TempDir()
		writeTestInventory(t, directory, crtoolsImagesV1_2, CompressionOff, 0)
		entry := presentTestEntry(base, 1)
		entry.Flags = uint32Pointer(pePresent | pePayloadAligned)
		writeTestPagemap(t, directory, 1, 9, entry)

		if _, err := NewMemoryReader(directory, 1, pageSize); err == nil ||
			!strings.Contains(err.Error(), "compressed or aligned page metadata, but inventory compression is off") {
			t.Fatalf("expected aligned/compression-off conflict, got %v", err)
		}
	})

	t.Run("equivalent shared pages id is counted once", func(t *testing.T) {
		directory := t.TempDir()
		writeTestInventory(t, directory, crtoolsImagesV1_1, CompressionOff, 0)
		entryA := presentTestEntry(base, 2)
		entryB := presentTestEntry(base, 2)
		writeTestPagemap(t, directory, 1, 9, entryA)
		writeTestPagemap(t, directory, 2, 9, entryB)

		stats, err := InspectMemoryPages(directory, pageSize)
		if err != nil {
			t.Fatal(err)
		}
		if stats.Pagemaps != 2 || stats.PresentPages != 2 ||
			stats.StoredBytes != 2*pageSize || stats.UncompressedBytes != 2*pageSize {
			t.Fatalf("shared pages image was double-counted: %+v", stats)
		}
	})

	t.Run("shared pages id requires identical block sizes", func(t *testing.T) {
		directory := t.TempDir()
		writeTestInventory(t, directory, crtoolsImagesV1_2, CompressionBlock, 0)
		entryA := presentTestEntry(base, 2)
		entryA.Blocks = testPagemapBlocks([]uint32{1, 3}, 1)
		entryB := presentTestEntry(base, 2)
		entryB.Blocks = testPagemapBlocks([]uint32{2, 2}, 1)
		writeTestPagemap(t, directory, 1, 9, entryA)
		writeTestPagemap(t, directory, 2, 9, entryB)

		if _, err := InspectMemoryPages(directory, pageSize); err == nil ||
			!strings.Contains(err.Error(), "different pagemap metadata") {
			t.Fatalf("expected shared pages ID mismatch, got %v", err)
		}
	})

	t.Run("pagemap head is required", func(t *testing.T) {
		directory := t.TempDir()
		writeTestInventory(t, directory, crtoolsImagesV1_1, CompressionOff, 0)
		writeTestImage(t, filepath.Join(directory, "pagemap-1.img"), &CriuImage{Magic: "PAGEMAP"})

		if _, err := InspectMemoryPages(directory, pageSize); err == nil ||
			!strings.Contains(err.Error(), "no entries") {
			t.Fatalf("expected missing pagemap head error, got %v", err)
		}
	})
}

func TestPagemapValidation(t *testing.T) {
	const pageSize = 4096
	baseEntry := func() *pagemap.PagemapEntry {
		return presentTestEntry(0x1000, 1)
	}
	tests := []struct {
		name    string
		entries []*pagemap.PagemapEntry
		want    string
	}{
		{
			name: "present and parent",
			entries: func() []*pagemap.PagemapEntry {
				entry := baseEntry()
				entry.Flags = uint32Pointer(pePresent | peParent)
				return []*pagemap.PagemapEntry{entry}
			}(),
			want: "mutually exclusive",
		},
		{
			name: "incomplete compression metadata",
			entries: func() []*pagemap.PagemapEntry {
				entry := baseEntry()
				entry.Blocks = &pagemap.PagemapBlocks{TotalPayloadSize: uint64Pointer(0)}
				return []*pagemap.PagemapEntry{entry}
			}(),
			want: "incomplete",
		},
		{
			name: "wrong block count",
			entries: func() []*pagemap.PagemapEntry {
				entry := presentTestEntry(0x1000, 2)
				entry.Blocks = testPagemapBlocks([]uint32{1}, 1)
				return []*pagemap.PagemapEntry{entry}
			}(),
			want: "compressed blocks",
		},
		{
			name: "mismatched total payload size",
			entries: func() []*pagemap.PagemapEntry {
				entry := baseEntry()
				entry.Blocks = testPagemapBlocks([]uint32{1}, 1)
				entry.Blocks.TotalPayloadSize = uint64Pointer(2)
				return []*pagemap.PagemapEntry{entry}
			}(),
			want: "does not match block sum",
		},
		{
			name: "oversized final block",
			entries: func() []*pagemap.PagemapEntry {
				entry := presentTestEntry(0x1000, 3)
				entry.Blocks = testPagemapBlocks([]uint32{1, pageSize + 1}, 2)
				return []*pagemap.PagemapEntry{entry}
			}(),
			want: "exceeds its uncompressed size",
		},
		{
			name: "page count overflow",
			entries: func() []*pagemap.PagemapEntry {
				entry := baseEntry()
				entry.NrPages = uint64Pointer(^uint64(0)/pageSize + 1)
				return []*pagemap.PagemapEntry{entry}
			}(),
			want: "page count overflows",
		},
		{
			name: "address range overflow",
			entries: []*pagemap.PagemapEntry{
				presentTestEntry(^uint64(0)-(pageSize-1), 1),
			},
			want: "address range overflows",
		},
		{
			name: "overlap",
			entries: []*pagemap.PagemapEntry{
				presentTestEntry(0x1000, 2),
				presentTestEntry(0x2000, 1),
			},
			want: "overlap",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := indexMemoryLayer(".", "pagemap-1.img", 1, pageSize, test.entries)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("expected error containing %q, got %v", test.want, err)
			}
		})
	}
}
