package crit

import (
	"testing"

	"github.com/checkpoint-restore/go-criu/v8/crit/images/pagemap"
)

func TestMemoryEntryRawPayloadRuns(t *testing.T) {
	const (
		pageSize = uint64(4096)
		base     = uint64(0x80000)
	)
	entry := memoryEntry{
		vaddr:                  base,
		end:                    base + 6*pageSize,
		nrPages:                6,
		hasCompressionMetadata: true,
		compressedSizes:        []uint32{uint32(pageSize), uint32(pageSize), 0, uint32(pageSize), uint32(pageSize), 17},
		message: &pagemap.PagemapEntry{
			CompressedSize: []uint32{uint32(pageSize), uint32(pageSize), 0, uint32(pageSize), uint32(pageSize), 17},
		},
		blockPayloadOffsets: []uint64{0},
	}

	tests := []struct {
		name       string
		start      uint64
		wantOffset uint64
		wantEnd    uint64
		wantRaw    bool
	}{
		{name: "leading raw run", start: base + 19, wantOffset: 19, wantEnd: base + 2*pageSize, wantRaw: true},
		{name: "zero block", start: base + 2*pageSize, wantEnd: base + 2*pageSize},
		{name: "later raw run", start: base + 3*pageSize + 23, wantOffset: 2*pageSize + 23, wantEnd: base + 5*pageSize, wantRaw: true},
		{name: "compressed block", start: base + 5*pageSize, wantEnd: base + 5*pageSize},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			offset, end, raw := entry.rawPayloadRun(test.start, entry.end, pageSize)
			if offset != test.wantOffset || end != test.wantEnd || raw != test.wantRaw {
				t.Fatalf("rawPayloadRun() = (%d, %#x, %v), want (%d, %#x, %v)",
					offset, end, raw, test.wantOffset, test.wantEnd, test.wantRaw)
			}
		})
	}
}

func TestMemoryEntryRawPayloadRunAcrossSparseStride(t *testing.T) {
	const (
		pageSize = uint64(4096)
		base     = uint64(0x100000)
		blocks   = 130
	)
	sizes := make([]uint32, blocks)
	prefix := make([]uint64, blocks+1)
	for index := range sizes {
		sizes[index] = uint32(index%31 + 1)
	}
	for _, index := range []int{63, 64, 65} {
		sizes[index] = uint32(pageSize)
	}
	for index, size := range sizes {
		prefix[index+1] = prefix[index] + uint64(size)
	}
	entry := presentTestEntry(base, blocks)
	entry.CompressedSize = sizes
	layer, err := indexMemoryLayer(".", "pagemap-1.img", 1, int(pageSize), []*pagemap.PagemapEntry{entry})
	if err != nil {
		t.Fatal(err)
	}
	indexed := &layer.entries[0]

	start := base + 63*pageSize + 17
	end := base + 66*pageSize
	offset, runEnd, raw := indexed.rawPayloadRun(start, end, pageSize)
	if !raw || offset != prefix[63]+17 || runEnd != end {
		t.Fatalf(
			"raw run = (%d, %#x, %v), want (%d, %#x, true)",
			offset,
			runEnd,
			raw,
			prefix[63]+17,
			end,
		)
	}
}
