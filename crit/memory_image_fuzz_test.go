package crit

import (
	"encoding/binary"
	"testing"

	"github.com/checkpoint-restore/go-criu/v8/crit/images/pagemap"
)

// FuzzIndexMemoryLayer exercises pagemap metadata accounting without opening
// image files. Malformed inputs may be rejected, but accepted inputs must
// describe a self-consistent logical range and payload layout.
func FuzzIndexMemoryLayer(f *testing.F) {
	f.Add([]byte{0, 0, 0, 0, 0, 0, 0, 0, 1, byte(pePresent), 0, 0})
	f.Add([]byte{0, 0x10, 0, 0, 0, 0, 0, 0, 1, byte(pePresent), 0, 2, 0, 0})
	f.Add([]byte{0, 0x20, 0, 0, 0, 0, 0, 0, 3, byte(pePresent), 2, 2, 0, 0, 0, 0})
	f.Add([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 16, 0xff, 0xff, 0xff, 0xff, 0xff})

	f.Fuzz(func(t *testing.T, data []byte) {
		const pageSize = 4096
		if len(data) < 12 {
			return
		}

		vaddr := binary.LittleEndian.Uint64(data[:8]) &^ uint64(pageSize-1)
		nrPages := uint64(data[8] % 17)
		flags := uint32(data[9]) & (peParent | pePresent | pePayloadAligned)
		entry := &pagemap.PagemapEntry{
			Vaddr:         uint64Pointer(vaddr),
			CompatNrPages: uint32Pointer(uint32(nrPages)),
		}
		if data[11]&0x10 == 0 {
			entry.Flags = uint32Pointer(flags)
		}
		if data[11]&0x04 != 0 {
			entry.NrPages = uint64Pointer(nrPages)
		}
		if data[11]&0x08 != 0 {
			entry.InParent = boolPointer(true)
		}
		blocks := &pagemap.PagemapBlocks{}
		hasBlocks := false
		if data[10]&0x80 != 0 {
			blocks.PagesPerBlock = uint32Pointer(uint32(data[10] & 0x0f))
			hasBlocks = true
		}

		encodedSum := uint64(0)
		encoded := data[12:]
		if len(encoded) > 32 {
			encoded = encoded[:32]
		}
		for len(encoded) >= 2 {
			size := uint32(binary.LittleEndian.Uint16(encoded[:2]))
			blocks.BlockSizes = append(blocks.BlockSizes, size)
			hasBlocks = true
			encodedSum += uint64(size)
			encoded = encoded[2:]
		}
		if data[11]&0x01 != 0 {
			total := binary.LittleEndian.Uint64(data[:8])
			if data[11]&0x02 != 0 {
				total = encodedSum
			}
			blocks.TotalPayloadSize = uint64Pointer(total)
			hasBlocks = true
		}
		if hasBlocks {
			entry.Blocks = blocks
		}

		layer, err := indexMemoryLayer(".", "pagemap-fuzz.img", 1, pageSize, []*pagemap.PagemapEntry{entry})
		if err != nil {
			return
		}
		if len(layer.entries) != 1 {
			t.Fatalf("accepted layer has %d entries, want 1", len(layer.entries))
		}

		indexed := layer.entries[0]
		if indexed.vaddr != vaddr || indexed.nrPages != nrPages {
			t.Fatalf("indexed range changed: %#x/%d, want %#x/%d", indexed.vaddr, indexed.nrPages, vaddr, nrPages)
		}
		if indexed.end-indexed.vaddr != nrPages*pageSize {
			t.Fatalf("indexed range length = %d, want %d", indexed.end-indexed.vaddr, nrPages*pageSize)
		}

		present := indexed.flags&pePresent != 0
		if !present {
			if indexed.payloadSize != 0 || layer.payloadSize != 0 || layer.presentPages != 0 {
				t.Fatalf("non-present entry consumes payload: entry=%+v layer=%+v", indexed, layer)
			}
			return
		}
		if layer.presentPages != nrPages || !layer.hasPresent {
			t.Fatalf("present-page accounting = %d/%v, want %d/true", layer.presentPages, layer.hasPresent, nrPages)
		}
		if indexed.payloadOffset+indexed.payloadSize != layer.payloadSize {
			t.Fatalf("payload accounting = %d+%d, layer total %d", indexed.payloadOffset, indexed.payloadSize, layer.payloadSize)
		}

		if entry.Blocks == nil {
			if indexed.payloadSize != nrPages*pageSize {
				t.Fatalf("raw payload size = %d, want %d", indexed.payloadSize, nrPages*pageSize)
			}
			return
		}
		expectedOffsets := (len(entry.Blocks.BlockSizes) + int(memoryBlockOffsetStride) - 1) / int(memoryBlockOffsetStride)
		if len(indexed.blockPayloadOffsets) != expectedOffsets {
			t.Fatalf("block offset index has %d entries, want %d", len(indexed.blockPayloadOffsets), expectedOffsets)
		}

		pages := uint64(0)
		stored := uint64(0)
		nextOffset := indexed.payloadOffset
		for blockIndex := range entry.Blocks.BlockSizes {
			firstPage, pageCount, storedSize, ok := indexed.blockLayout(uint64(blockIndex))
			if !ok || firstPage != pages || indexed.blockPayloadOffset(uint64(blockIndex)) != nextOffset {
				t.Fatalf("block %d is discontinuous", blockIndex)
			}
			if storedSize > pageCount*pageSize {
				t.Fatalf("block %d stores %d bytes for %d pages", blockIndex, storedSize, pageCount)
			}
			pages += pageCount
			stored += storedSize
			nextOffset += storedSize
		}
		if pages != nrPages || stored != indexed.payloadSize {
			t.Fatalf("block totals pages=%d bytes=%d, want pages=%d bytes=%d", pages, stored, nrPages, indexed.payloadSize)
		}
	})
}
