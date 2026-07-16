package crit

import "github.com/checkpoint-restore/go-criu/v8/crit/images/pagemap"

// PagemapFlags holds the PE_* flags of a CRIU pagemap entry.
type PagemapFlags uint32

const (
	// PagemapParent means the pages are stored in a parent checkpoint.
	PagemapParent PagemapFlags = 1 << 0
	// PagemapLazy means the pages may be restored on demand by the lazy
	// pages daemon. Without PagemapPresent, no image stores them.
	PagemapLazy PagemapFlags = 1 << 1
	// PagemapPresent means the pages are stored in this checkpoint.
	PagemapPresent PagemapFlags = 1 << 2
	// PagemapPayloadAligned means the entry payload starts at a
	// page-aligned offset in the pages image.
	PagemapPayloadAligned PagemapFlags = 1 << 3
)

// EffectivePagemapFlags returns the flags of entry, including the flags
// that older images imply: in_parent marks parent pages, and an entry
// without flags is present.
func EffectivePagemapFlags(entry *pagemap.PagemapEntry) PagemapFlags {
	if entry == nil {
		return 0
	}
	return PagemapFlags(effectivePagemapFlags(entry))
}

// PagemapPageCount returns the number of pages that entry covers. Older
// images record it only in compat_nr_pages.
func PagemapPageCount(entry *pagemap.PagemapEntry) uint64 {
	if entry == nil {
		return 0
	}
	return pagemapPageCount(entry)
}
