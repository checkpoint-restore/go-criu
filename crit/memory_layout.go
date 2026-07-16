package crit

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/checkpoint-restore/go-criu/v8/crit/images/pagemap"
)

const (
	peParent         uint32 = 1 << 0
	pePresent        uint32 = 1 << 2
	pePayloadAligned uint32 = 1 << 3

	memoryBlockOffsetStride uint64 = 64
)

var errEmptyPagemap = errors.New("pagemap image has no entries")

type memoryEntry struct {
	message                *pagemap.PagemapEntry
	vaddr                  uint64
	end                    uint64
	nrPages                uint64
	flags                  uint32
	payloadOffset          uint64
	payloadSize            uint64
	blockPages             uint64
	hasCompressionMetadata bool
	blockSizes             []uint32
	blockPayloadOffsets    []uint64
}

type memoryLayer struct {
	directory      string
	pagemapName    string
	pagesID        uint32
	pageSize       int
	entries        []memoryEntry
	pagemapEntries []*pagemap.PagemapEntry
	payloadSize    uint64
	presentPages   uint64
	hasPresent     bool
	hasParent      bool
	hasAligned     bool
	hasCompression bool
	hasInventory   bool
	compression    CompressionInfo
}

func effectivePagemapFlags(entry *pagemap.PagemapEntry) uint32 {
	flags := entry.GetFlags()
	if entry.GetInParent() {
		flags |= peParent
	} else if entry.Flags == nil {
		flags = pePresent
	}
	return flags
}

func pagemapPageCount(entry *pagemap.PagemapEntry) uint64 {
	if entry.NrPages != nil {
		return entry.GetNrPages()
	}
	return uint64(entry.GetCompatNrPages())
}

func hasPagemapCompressionMetadata(entry *pagemap.PagemapEntry) bool {
	return entry.GetBlocks() != nil
}

func loadMemoryLayer(directory, pagemapName string, pageSize int) (*memoryLayer, error) {
	path := filepath.Join(directory, pagemapName)
	img, err := getImg(path, &pagemap.PagemapHead{})
	if err != nil {
		return nil, err
	}
	if img.Magic != "PAGEMAP" {
		return nil, fmt.Errorf("%s is a %s image, expected PAGEMAP", path, img.Magic)
	}
	if len(img.Entries) == 0 {
		return nil, fmt.Errorf("%s: %w", path, errEmptyPagemap)
	}
	head, ok := img.Entries[0].Message.(*pagemap.PagemapHead)
	if !ok || head == nil || head.PagesId == nil {
		return nil, fmt.Errorf("%s has an invalid pagemap head", path)
	}

	entries := make([]*pagemap.PagemapEntry, 0, len(img.Entries)-1)
	for index, imageEntry := range img.Entries[1:] {
		entry, ok := imageEntry.Message.(*pagemap.PagemapEntry)
		if !ok || entry == nil {
			return nil, fmt.Errorf("%s entry %d is invalid", path, index+1)
		}
		entries = append(entries, entry)
	}

	layer, err := indexMemoryLayer(directory, pagemapName, head.GetPagesId(), pageSize, entries)
	if err != nil {
		return nil, err
	}
	if err := validateMemoryLayerInventory(layer); err != nil {
		return nil, err
	}
	return layer, nil
}

func validateMemoryLayerInventory(layer *memoryLayer) error {
	info, err := ReadCompressionInfo(layer.directory)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) && !layer.hasCompression && !layer.hasAligned {
			// Older callers commonly extract only pagemap/pages images. That is
			// unambiguous only for the legacy raw, unaligned representation.
			return nil
		}
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf(
				"%s requires inventory.img to validate compressed page layout: %w",
				layer.pagemapName,
				err,
			)
		}
		return err
	}
	layer.hasInventory = true
	layer.compression = info

	if (layer.hasCompression || layer.hasAligned) && !info.Mode.Compressed() {
		return fmt.Errorf(
			"%s contains compressed or aligned page metadata, but inventory compression is off",
			layer.pagemapName,
		)
	}
	for index := range layer.entries {
		entry := &layer.entries[index]
		if !entry.hasCompressionMetadata {
			continue
		}
		blockBytes, ok := checkedMultiply(entry.blockPages, uint64(layer.pageSize))
		if !ok || (info.BlockSizeBytes != 0 && blockBytes != uint64(info.BlockSizeBytes)) {
			return pagemapEntryError(
				layer.pagemapName,
				index+1,
				fmt.Sprintf(
					"block size %d conflicts with inventory block size %d",
					blockBytes,
					info.BlockSizeBytes,
				),
			)
		}
	}
	return nil
}

func indexMemoryLayer(
	directory, pagemapName string,
	pagesID uint32,
	pageSize int,
	entries []*pagemap.PagemapEntry,
) (*memoryLayer, error) {
	if pageSize <= 0 || pageSize&(pageSize-1) != 0 {
		return nil, errors.New("page size should be a positive power of 2")
	}

	layer := &memoryLayer{
		directory:      directory,
		pagemapName:    pagemapName,
		pagesID:        pagesID,
		pageSize:       pageSize,
		pagemapEntries: entries,
	}
	pageBytes := uint64(pageSize)
	var previousEnd uint64
	for index, message := range entries {
		entryNumber := index + 1
		if message == nil {
			return nil, pagemapEntryError(pagemapName, entryNumber, "entry is nil")
		}

		nrPages := pagemapPageCount(message)
		if nrPages == 0 {
			return nil, pagemapEntryError(pagemapName, entryNumber, "page count is zero")
		}
		decodedBytes, ok := checkedMultiply(nrPages, pageBytes)
		if !ok {
			return nil, pagemapEntryError(pagemapName, entryNumber, "page count overflows")
		}
		if message.Vaddr == nil {
			return nil, pagemapEntryError(pagemapName, entryNumber, "address is missing")
		}
		vaddr := message.GetVaddr()
		if vaddr%pageBytes != 0 {
			return nil, pagemapEntryError(
				pagemapName,
				entryNumber,
				fmt.Sprintf("address %#x is not page-aligned", vaddr),
			)
		}
		end, ok := checkedAdd(vaddr, decodedBytes)
		if !ok {
			return nil, pagemapEntryError(pagemapName, entryNumber, "address range overflows")
		}
		if index != 0 && vaddr < previousEnd {
			return nil, pagemapEntryError(
				pagemapName,
				entryNumber,
				"entries are not sorted or overlap",
			)
		}
		previousEnd = end

		flags := effectivePagemapFlags(message)
		present := flags&pePresent != 0
		parent := flags&peParent != 0
		if present && parent {
			return nil, pagemapEntryError(
				pagemapName,
				entryNumber,
				"PE_PRESENT and PE_PARENT are mutually exclusive",
			)
		}
		if flags&pePayloadAligned != 0 && !present {
			return nil, pagemapEntryError(
				pagemapName,
				entryNumber,
				"PE_PAYLOAD_ALIGNED is set on a non-present entry",
			)
		}

		hasCompression := hasPagemapCompressionMetadata(message)
		if !present && hasCompression {
			return nil, pagemapEntryError(
				pagemapName,
				entryNumber,
				"compression metadata is set on a non-present entry",
			)
		}

		indexed := memoryEntry{
			message:                message,
			vaddr:                  vaddr,
			end:                    end,
			nrPages:                nrPages,
			flags:                  flags,
			hasCompressionMetadata: hasCompression,
		}
		if parent {
			layer.hasParent = true
		}
		if !present {
			layer.entries = append(layer.entries, indexed)
			continue
		}

		layer.hasPresent = true
		if layer.presentPages > ^uint64(0)-nrPages {
			return nil, pagemapEntryError(pagemapName, entryNumber, "present page count overflows")
		}
		layer.presentPages += nrPages
		if flags&pePayloadAligned != 0 {
			layer.hasAligned = true
			var ok bool
			layer.payloadSize, ok = checkedAlignUp(layer.payloadSize, pageBytes)
			if !ok {
				return nil, pagemapEntryError(pagemapName, entryNumber, "payload alignment overflows")
			}
		}
		indexed.payloadOffset = layer.payloadSize

		blocks := message.GetBlocks()
		if blocks == nil {
			indexed.payloadSize = decodedBytes
		} else {
			layer.hasCompression = true
			blockSizes := slices.Clone(blocks.GetBlockSizes())
			if len(blockSizes) == 0 || blocks.TotalPayloadSize == nil || blocks.PagesPerBlock == nil {
				return nil, pagemapEntryError(
					pagemapName,
					entryNumber,
					"compression metadata is incomplete",
				)
			}
			indexed.blockSizes = blockSizes
			blockPages := uint64(blocks.GetPagesPerBlock())
			if blockPages == 0 {
				return nil, pagemapEntryError(pagemapName, entryNumber, "pages_per_block is zero")
			}
			blockBytes, ok := checkedMultiply(blockPages, pageBytes)
			if !ok || blockBytes > maxCompressedBlockSize {
				return nil, pagemapEntryError(
					pagemapName,
					entryNumber,
					fmt.Sprintf("block size exceeds maximum %d bytes", maxCompressedBlockSize),
				)
			}
			expectedBlocks := nrPages / blockPages
			if nrPages%blockPages != 0 {
				expectedBlocks++
			}
			if uint64(len(blockSizes)) != expectedBlocks {
				return nil, pagemapEntryError(
					pagemapName,
					entryNumber,
					fmt.Sprintf("%d compressed blocks, expected %d", len(blockSizes), expectedBlocks),
				)
			}

			indexed.blockPages = blockPages
			blockGroups := (uint64(len(blockSizes)) + memoryBlockOffsetStride - 1) /
				memoryBlockOffsetStride
			indexed.blockPayloadOffsets = make([]uint64, 0, int(blockGroups))
			pagesDone := uint64(0)
			payloadOffset := layer.payloadSize
			for blockIndex, storedSize := range blockSizes {
				if uint64(blockIndex)%memoryBlockOffsetStride == 0 {
					indexed.blockPayloadOffsets = append(indexed.blockPayloadOffsets, payloadOffset)
				}
				pages := min(blockPages, nrPages-pagesDone)
				blockBytes, ok := checkedMultiply(pages, pageBytes)
				if !ok {
					return nil, pagemapEntryError(pagemapName, entryNumber, "block size overflows")
				}
				if uint64(storedSize) > blockBytes {
					return nil, pagemapEntryError(
						pagemapName,
						entryNumber,
						fmt.Sprintf(
							"block %d size %d exceeds its uncompressed size %d",
							blockIndex,
							storedSize,
							blockBytes,
						),
					)
				}
				payloadOffset, ok = checkedAdd(payloadOffset, uint64(storedSize))
				if !ok {
					return nil, pagemapEntryError(pagemapName, entryNumber, "compressed size sum overflows")
				}
				pagesDone += pages
			}
			indexed.payloadSize = payloadOffset - layer.payloadSize
			if blocks.GetTotalPayloadSize() != indexed.payloadSize {
				return nil, pagemapEntryError(
					pagemapName,
					entryNumber,
					fmt.Sprintf(
						"total_payload_size %d does not match block sum %d",
						blocks.GetTotalPayloadSize(),
						indexed.payloadSize,
					),
				)
			}
		}

		var okPayload bool
		layer.payloadSize, okPayload = checkedAdd(layer.payloadSize, indexed.payloadSize)
		if !okPayload {
			return nil, pagemapEntryError(pagemapName, entryNumber, "payload size overflows")
		}
		layer.entries = append(layer.entries, indexed)
	}

	return layer, nil
}

func pagemapEntryError(name string, index int, message string) error {
	return fmt.Errorf("%s entry %d: %s", name, index, message)
}

func checkedAdd(a, b uint64) (uint64, bool) {
	if a > ^uint64(0)-b {
		return 0, false
	}
	return a + b, true
}

func checkedMultiply(a, b uint64) (uint64, bool) {
	if a != 0 && b > ^uint64(0)/a {
		return 0, false
	}
	return a * b, true
}

func checkedAlignUp(value, alignment uint64) (uint64, bool) {
	remainder := value % alignment
	if remainder == 0 {
		return value, true
	}
	return checkedAdd(value, alignment-remainder)
}

func (entry *memoryEntry) blockLayout(blockIndex uint64) (uint64, uint64, uint64, bool) {
	blockSizes := entry.blockSizes
	if blockIndex >= uint64(len(blockSizes)) {
		return 0, 0, 0, false
	}
	blockPages := entry.blockPages
	firstPage := blockIndex * blockPages
	pageCount := min(blockPages, entry.nrPages-firstPage)
	return firstPage, pageCount, uint64(blockSizes[blockIndex]), true
}

func (entry *memoryEntry) blockPayloadOffset(blockIndex uint64) uint64 {
	group := blockIndex / memoryBlockOffsetStride
	offset := entry.blockPayloadOffsets[group]
	first := group * memoryBlockOffsetStride
	for index := first; index < blockIndex; index++ {
		offset += uint64(entry.blockSizes[index])
	}
	return offset
}

func (layer *memoryLayer) findEntry(vaddr uint64) *memoryEntry {
	index := sort.Search(len(layer.entries), func(index int) bool {
		return layer.entries[index].end > vaddr
	})
	if index == len(layer.entries) || vaddr < layer.entries[index].vaddr {
		return nil
	}
	return &layer.entries[index]
}

// MemoryPageStats summarizes local memory-page image metadata in a
// checkpoint directory. Pages delegated to parents are not counted as local.
type MemoryPageStats struct {
	Pagemaps               int
	PresentPages           uint64
	StoredBytes            uint64
	UncompressedBytes      uint64
	HasParentReferences    bool
	HasCompressionMetadata bool
}

// InspectMemoryPages validates local pagemap metadata and summarizes each
// unique pages image. It does not open pages-N.img payloads.
func InspectMemoryPages(checkpointDir string, pageSize int) (MemoryPageStats, error) {
	if pageSize == 0 {
		pageSize = sysPageSize
	}
	if pageSize <= 0 || pageSize&(pageSize-1) != 0 {
		return MemoryPageStats{}, errors.New("page size should be a positive power of 2")
	}

	names, err := findPagemapNames(checkpointDir)
	if err != nil {
		return MemoryPageStats{}, err
	}
	stats := MemoryPageStats{}
	seenPagesIDs := make(map[uint32]*memoryLayer)
	for _, name := range names {
		layer, err := loadMemoryLayer(checkpointDir, name, pageSize)
		if err != nil {
			return MemoryPageStats{}, err
		}
		stats.Pagemaps++
		if previous, exists := seenPagesIDs[layer.pagesID]; exists {
			if !equivalentMemoryLayerMetadata(previous, layer) {
				return MemoryPageStats{}, fmt.Errorf(
					"%s reuses pages_id %d with different pagemap metadata",
					name,
					layer.pagesID,
				)
			}
			continue
		}
		seenPagesIDs[layer.pagesID] = layer
		if stats.PresentPages > ^uint64(0)-layer.presentPages {
			return MemoryPageStats{}, errors.New("present page count overflows")
		}
		stats.PresentPages += layer.presentPages
		if stats.StoredBytes > ^uint64(0)-layer.payloadSize {
			return MemoryPageStats{}, errors.New("stored page size overflows")
		}
		stats.StoredBytes += layer.payloadSize
		uncompressed, ok := checkedMultiply(layer.presentPages, uint64(pageSize))
		if !ok || stats.UncompressedBytes > ^uint64(0)-uncompressed {
			return MemoryPageStats{}, errors.New("uncompressed page size overflows")
		}
		stats.UncompressedBytes += uncompressed
		stats.HasParentReferences = stats.HasParentReferences || layer.hasParent
		stats.HasCompressionMetadata = stats.HasCompressionMetadata || layer.hasCompression
	}
	return stats, nil
}

func equivalentMemoryLayerMetadata(a, b *memoryLayer) bool {
	if a.payloadSize != b.payloadSize || a.presentPages != b.presentPages ||
		a.hasPresent != b.hasPresent || a.hasParent != b.hasParent ||
		a.hasAligned != b.hasAligned || a.hasCompression != b.hasCompression ||
		a.hasInventory != b.hasInventory || a.compression != b.compression ||
		len(a.entries) != len(b.entries) {
		return false
	}
	for index := range a.entries {
		aEntry := &a.entries[index]
		bEntry := &b.entries[index]
		if aEntry.vaddr != bEntry.vaddr || aEntry.end != bEntry.end ||
			aEntry.nrPages != bEntry.nrPages || aEntry.flags != bEntry.flags ||
			aEntry.payloadOffset != bEntry.payloadOffset ||
			aEntry.payloadSize != bEntry.payloadSize ||
			aEntry.blockPages != bEntry.blockPages ||
			aEntry.hasCompressionMetadata != bEntry.hasCompressionMetadata ||
			!slices.Equal(aEntry.blockSizes, bEntry.blockSizes) ||
			len(aEntry.blockPayloadOffsets) != len(bEntry.blockPayloadOffsets) {
			return false
		}
		for blockIndex := range aEntry.blockPayloadOffsets {
			if aEntry.blockPayloadOffsets[blockIndex] != bEntry.blockPayloadOffsets[blockIndex] {
				return false
			}
		}
	}
	return true
}

func findPagemapNames(directory string) ([]string, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("read checkpoint directory %s: %w", directory, err)
	}
	var names []string
	for _, entry := range entries {
		if entry.IsDir() || !isPagemapName(entry.Name()) {
			continue
		}
		names = append(names, entry.Name())
	}
	sort.Strings(names)
	return names, nil
}

func isPagemapName(name string) bool {
	if !strings.HasSuffix(name, ".img") {
		return false
	}
	id := strings.TrimSuffix(strings.TrimPrefix(name, "pagemap-"), ".img")
	id = strings.TrimPrefix(id, "shmem-")
	if id == "" || !strings.HasPrefix(name, "pagemap-") {
		return false
	}
	_, err := strconv.ParseUint(id, 10, 64)
	return err == nil
}
