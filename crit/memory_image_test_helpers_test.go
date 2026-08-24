package crit

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/checkpoint-restore/go-criu/v8/crit/images/pagemap"
)

func uint64Pointer(value uint64) *uint64 { return &value }
func boolPointer(value bool) *bool       { return &value }

func writeTestPagemap(
	t *testing.T,
	directory string,
	pid, pagesID uint32,
	entries ...*pagemap.PagemapEntry,
) {
	t.Helper()
	imageEntries := make([]*CriuEntry, 1, 1+len(entries))
	imageEntries[0] = &CriuEntry{
		Message: &pagemap.PagemapHead{PagesId: uint32Pointer(pagesID)},
	}
	for _, entry := range entries {
		imageEntries = append(imageEntries, &CriuEntry{Message: entry})
	}
	writeTestImage(t, filepath.Join(directory, "pagemap-"+stringID(pid)+".img"), &CriuImage{
		Magic:   "PAGEMAP",
		Entries: imageEntries,
	})
}

func stringID(id uint32) string {
	return fmt.Sprintf("%d", id)
}

func presentTestEntry(vaddr, pages uint64) *pagemap.PagemapEntry {
	return &pagemap.PagemapEntry{
		Vaddr:         uint64Pointer(vaddr),
		CompatNrPages: uint32Pointer(uint32(pages)),
		Flags:         uint32Pointer(pePresent),
		NrPages:       uint64Pointer(pages),
	}
}

func testPagemapBlocks(blockSizes []uint32, pagesPerBlock uint32) *pagemap.PagemapBlocks {
	var total uint64
	for _, size := range blockSizes {
		total += uint64(size)
	}
	return &pagemap.PagemapBlocks{
		BlockSizes:       blockSizes,
		TotalPayloadSize: uint64Pointer(total),
		PagesPerBlock:    uint32Pointer(pagesPerBlock),
	}
}
