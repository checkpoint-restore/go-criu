package crit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/checkpoint-restore/go-criu/v8/crit/images/pagemap"
)

func TestEffectivePagemapFlags(t *testing.T) {
	for _, test := range []struct {
		name  string
		entry *pagemap.PagemapEntry
		want  PagemapFlags
	}{
		{name: "nil entry"},
		{name: "legacy present entry", entry: &pagemap.PagemapEntry{}, want: PagemapPresent},
		{name: "legacy parent entry", entry: &pagemap.PagemapEntry{InParent: boolPointer(true)}, want: PagemapParent},
		{name: "lazy entry", entry: &pagemap.PagemapEntry{Flags: uint32Pointer(uint32(PagemapLazy))}, want: PagemapLazy},
		{
			name:  "aligned entry",
			entry: &pagemap.PagemapEntry{Flags: uint32Pointer(uint32(PagemapPresent | PagemapPayloadAligned))},
			want:  PagemapPresent | PagemapPayloadAligned,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := EffectivePagemapFlags(test.entry); got != test.want {
				t.Fatalf("EffectivePagemapFlags() = %#x, want %#x", got, test.want)
			}
		})
	}
}

func TestPagemapPageCount(t *testing.T) {
	for _, test := range []struct {
		name  string
		entry *pagemap.PagemapEntry
		want  uint64
	}{
		{name: "nil entry"},
		{name: "nr_pages", entry: &pagemap.PagemapEntry{NrPages: uint64Pointer(7), CompatNrPages: uint32Pointer(3)}, want: 7},
		{name: "compat_nr_pages", entry: &pagemap.PagemapEntry{CompatNrPages: uint32Pointer(3)}, want: 3},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := PagemapPageCount(test.entry); got != test.want {
				t.Fatalf("PagemapPageCount() = %d, want %d", got, test.want)
			}
		})
	}
}

func TestMemoryReaderPageStats(t *testing.T) {
	const (
		pageSize = 4096
		pid      = 41
		base     = 0x70000
	)
	directory := t.TempDir()
	writeTestInventory(t, directory, crtoolsImagesV1_2, CompressionBlock, pageSize)

	// One zero block and one raw block leave a single stored page.
	compressed := presentTestEntry(base, 2)
	compressed.Blocks = testPagemapBlocks([]uint32{0, pageSize}, 1)
	lazy := presentTestEntry(base+2*pageSize, 1)
	lazy.Flags = uint32Pointer(uint32(PagemapLazy))
	parent := &pagemap.PagemapEntry{
		Vaddr:         uint64Pointer(base + 3*pageSize),
		CompatNrPages: uint32Pointer(3),
		InParent:      boolPointer(true),
	}
	writeTestPagemap(t, directory, pid, 1, compressed, lazy, parent)
	if err := os.WriteFile(filepath.Join(directory, "pages-1.img"), make([]byte, pageSize), 0o600); err != nil {
		t.Fatal(err)
	}

	reader, err := NewMemoryReader(directory, pid, pageSize)
	if err != nil {
		t.Fatal(err)
	}
	stats, err := reader.PageStats()
	if err != nil {
		t.Fatal(err)
	}
	want := MemoryPageStats{
		Pagemaps:               1,
		PresentPages:           2,
		LazyPages:              1,
		ParentPages:            3,
		StoredBytes:            pageSize,
		UncompressedBytes:      2 * pageSize,
		HasParentReferences:    true,
		HasCompressionMetadata: true,
	}
	if stats != want {
		t.Fatalf("PageStats() = %+v, want %+v", stats, want)
	}

	directoryStats, err := InspectMemoryPages(directory, pageSize)
	if err != nil {
		t.Fatal(err)
	}
	if directoryStats != want {
		t.Fatalf("InspectMemoryPages() = %+v, want %+v", directoryStats, want)
	}
}

func TestMemoryReaderPageStatsRequiresReader(t *testing.T) {
	if _, err := (&MemoryReader{}).PageStats(); err == nil {
		t.Fatal("expected an error from a MemoryReader without an index")
	}
}
