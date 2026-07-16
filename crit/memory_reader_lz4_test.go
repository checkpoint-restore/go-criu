package crit

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pierrec/lz4/v4"
)

func compressTestBlock(t *testing.T, data []byte) []byte {
	t.Helper()
	destination := make([]byte, lz4.CompressBlockBound(len(data)))
	var compressor lz4.Compressor
	n, err := compressor.CompressBlock(data, destination)
	if err != nil {
		t.Fatal(err)
	}
	if n == 0 || n >= len(data) {
		t.Fatalf("test data did not compress: got %d bytes from %d", n, len(data))
	}
	return destination[:n]
}

func TestMemoryReaderAcrossSparseBlockStride(t *testing.T) {
	const (
		pageSize = 4096
		blocks   = 130
		pid      = 41
		pagesID  = 17
		base     = uint64(0x200000)
	)
	directory := t.TempDir()
	writeTestInventory(t, directory, crtoolsImagesV1_2, CompressionPerPage, 0)

	entry := presentTestEntry(base, blocks)
	entry.CompressedSize = make([]uint32, blocks)
	var payload []byte
	expected := make([]byte, 0, blocks*pageSize)
	for index := 0; index < blocks; index++ {
		page := make([]byte, pageSize)
		switch index % 3 {
		case 0:
			for offset := range page {
				page[offset] = byte(index%251 + 1)
			}
			encoded := compressTestBlock(t, page)
			entry.CompressedSize[index] = uint32(len(encoded))
			payload = append(payload, encoded...)
		case 1:
			entry.CompressedSize[index] = 0
		case 2:
			for offset := range page {
				page[offset] = byte((index*17+offset*31)%251 + 1)
			}
			entry.CompressedSize[index] = pageSize
			payload = append(payload, page...)
		}
		expected = append(expected, page...)
	}
	entry.TotalCompressedSize = uint64Pointer(uint64(len(payload)))
	writeTestPagemap(t, directory, pid, pagesID, entry)
	if err := os.WriteFile(filepath.Join(directory, "pages-17.img"), payload, 0o600); err != nil {
		t.Fatal(err)
	}

	reader, err := NewMemoryReader(directory, pid, pageSize)
	if err != nil {
		t.Fatal(err)
	}
	const (
		startOffset = 62*pageSize + 17
		endOffset   = 67*pageSize - 23
	)
	memory, err := reader.GetMemPages(base+startOffset, base+endOffset)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(memory.Bytes(), expected[startOffset:endOffset]) {
		t.Fatal("decoded memory across the sparse block stride does not match")
	}
}

func TestMemoryReaderPerPageZeroRawAlignedAndSearch(t *testing.T) {
	const (
		pageSize = 4096
		pid      = 17
		pagesID  = 3
		base     = 0x10000
	)
	directory := t.TempDir()
	writeTestInventory(t, directory, crtoolsImagesV1_2, CompressionPerPage, 0)

	compressedPage := bytes.Repeat([]byte{'a'}, pageSize)
	needleOffset := 177
	copy(compressedPage[needleOffset:], "compressed-needle")
	encoded := compressTestBlock(t, compressedPage)
	first := presentTestEntry(base, 2)
	first.CompressedSize = []uint32{0, uint32(len(encoded))}
	first.TotalCompressedSize = uint64Pointer(uint64(len(encoded)))
	hole := presentTestEntry(base+2*pageSize, 1)
	hole.Flags = uint32Pointer(0)
	second := presentTestEntry(base+3*pageSize, 1)
	second.Flags = uint32Pointer(pePresent | pePayloadAligned)
	writeTestPagemap(t, directory, pid, pagesID, first, hole, second)

	// Construction and empty reads require only pagemap/inventory metadata.
	reader, err := NewMemoryReader(directory, pid, pageSize)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := reader.GetMemPages(base, base); err != nil {
		t.Fatalf("empty read unexpectedly opened pages image: %v", err)
	}

	rawPage := bytes.Repeat([]byte{'B'}, pageSize)
	payload := append([]byte(nil), encoded...)
	padding := (pageSize - len(payload)%pageSize) % pageSize
	payload = append(payload, make([]byte, padding)...)
	payload = append(payload, rawPage...)
	if err := os.WriteFile(filepath.Join(directory, "pages-3.img"), payload, 0o600); err != nil {
		t.Fatal(err)
	}

	memory, err := reader.GetMemPages(base+100, base+4*pageSize)
	if err != nil {
		t.Fatal(err)
	}
	expected := append(make([]byte, pageSize-100), compressedPage...)
	expected = append(expected, make([]byte, pageSize)...)
	expected = append(expected, rawPage...)
	if !bytes.Equal(memory.Bytes(), expected) {
		t.Fatal("decoded per-page memory does not match")
	}
	if _, err := reader.GetMemPages(base+1, base); err == nil {
		t.Fatal("reversed memory range was accepted")
	}

	matches, err := reader.SearchPattern("compressed-needle", true, 4, pageSize)
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 1 || matches[0].Vaddr != base+pageSize+uint64(needleOffset) {
		t.Fatalf("unexpected compressed search matches: %+v", matches)
	}

	stats, err := InspectMemoryPages(directory, pageSize)
	if err != nil {
		t.Fatal(err)
	}
	if stats.Pagemaps != 1 || stats.PresentPages != 3 ||
		stats.StoredBytes != uint64(len(payload)) ||
		stats.UncompressedBytes != 3*pageSize ||
		stats.HasParentReferences || !stats.HasCompressionMetadata {
		t.Fatalf("unexpected memory page stats: %+v", stats)
	}
}

func TestMemoryReaderRegionAndCorruptBlock(t *testing.T) {
	const (
		pageSize = 4096
		pid      = 23
		base     = 0x20000
	)
	directory := t.TempDir()
	writeTestInventory(t, directory, crtoolsImagesV1_2, CompressionRegion, 2*pageSize)
	pageA := bytes.Repeat([]byte{'A'}, pageSize)
	pageB := bytes.Repeat([]byte{'C'}, pageSize)
	pageC := bytes.Repeat([]byte{'Z'}, pageSize)
	encoded := compressTestBlock(t, append(append([]byte(nil), pageA...), pageB...))
	entry := presentTestEntry(base, 3)
	entry.CompressedSize = []uint32{uint32(len(encoded)), pageSize}
	entry.TotalCompressedSize = uint64Pointer(uint64(len(encoded) + pageSize))
	entry.RegionPages = uint32Pointer(2)
	writeTestPagemap(t, directory, pid, 5, entry)
	if err := os.WriteFile(
		filepath.Join(directory, "pages-5.img"),
		append(encoded, pageC...),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	reader, err := NewMemoryReader(directory, pid, pageSize)
	if err != nil {
		t.Fatal(err)
	}
	// Both pages are stored in the same compressed region. Reading them in one
	// session exercises the decoded-region cache for the second page.
	cachedRegion, err := reader.GetMemPages(base+19, base+2*pageSize-7)
	if err != nil {
		t.Fatal(err)
	}
	wantCachedRegion := append(append([]byte(nil), pageA[19:]...), pageB[:pageSize-7]...)
	if !bytes.Equal(cachedRegion.Bytes(), wantCachedRegion) {
		t.Fatal("decoded pages from a cached region do not match")
	}

	memory, err := reader.GetMemPages(base+pageSize+19, base+3*pageSize-7)
	if err != nil {
		t.Fatal(err)
	}
	expected := append(append([]byte(nil), pageB[19:]...), pageC[:pageSize-7]...)
	if !bytes.Equal(memory.Bytes(), expected) {
		t.Fatal("decoded region memory does not match")
	}

	corrupt := t.TempDir()
	writeTestInventory(t, corrupt, crtoolsImagesV1_2, CompressionPerPage, 0)
	bad := presentTestEntry(base, 1)
	bad.CompressedSize = []uint32{1}
	bad.TotalCompressedSize = uint64Pointer(1)
	writeTestPagemap(t, corrupt, pid, 6, bad)
	if err := os.WriteFile(filepath.Join(corrupt, "pages-6.img"), []byte{0xff}, 0o600); err != nil {
		t.Fatal(err)
	}
	badReader, err := NewMemoryReader(corrupt, pid, pageSize)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := badReader.GetMemPages(base, base+pageSize); err == nil ||
		!strings.Contains(err.Error(), "decompress") {
		t.Fatalf("expected corrupt LZ4 error, got %v", err)
	}
}

func TestMemoryReaderRejectsCompressedBlockDecodedSizeMismatch(t *testing.T) {
	const (
		pageSize = 4096
		pid      = 24
		pagesID  = 6
		base     = 0x22000
	)
	directory := t.TempDir()
	writeTestInventory(t, directory, crtoolsImagesV1_2, CompressionRegion, 2*pageSize)

	// The payload is valid LZ4 for one page, but the pagemap declares it as a
	// two-page region. The decoder must reject a successful short decode.
	encoded := compressTestBlock(t, bytes.Repeat([]byte{'M'}, pageSize))
	entry := presentTestEntry(base, 2)
	entry.CompressedSize = []uint32{uint32(len(encoded))}
	entry.TotalCompressedSize = uint64Pointer(uint64(len(encoded)))
	entry.RegionPages = uint32Pointer(2)
	writeTestPagemap(t, directory, pid, pagesID, entry)
	if err := os.WriteFile(filepath.Join(directory, "pages-6.img"), encoded, 0o600); err != nil {
		t.Fatal(err)
	}

	reader, err := NewMemoryReader(directory, pid, pageSize)
	if err != nil {
		t.Fatal(err)
	}
	_, err = reader.GetMemPages(base, base+2*pageSize)
	if err == nil || !strings.Contains(err.Error(), "produced 4096 bytes, expected 8192") {
		t.Fatalf("expected decoded-size mismatch, got %v", err)
	}
}

func TestMemoryReaderRegionZeroRawAndFinalShortBlockWithoutTotal(t *testing.T) {
	const (
		pageSize = 4096
		pid      = 29
		pagesID  = 7
		base     = 0x24000
	)
	directory := t.TempDir()
	writeTestInventory(t, directory, crtoolsImagesV1_2, CompressionRegion, 2*pageSize)

	rawRegion := append(
		bytes.Repeat([]byte{'R'}, pageSize),
		bytes.Repeat([]byte{'S'}, pageSize)...,
	)
	finalPage := bytes.Repeat([]byte{'F'}, pageSize)
	encodedFinalPage := compressTestBlock(t, finalPage)
	entry := presentTestEntry(base, 5)
	entry.CompressedSize = []uint32{
		0,
		uint32(len(rawRegion)),
		uint32(len(encodedFinalPage)),
	}
	entry.RegionPages = uint32Pointer(2)
	// total_compressed_size is optional. Readers must derive the payload
	// length from compressed_size when it is absent.
	writeTestPagemap(t, directory, pid, pagesID, entry)
	payload := append(append([]byte(nil), rawRegion...), encodedFinalPage...)
	if err := os.WriteFile(
		filepath.Join(directory, "pages-7.img"),
		payload,
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	reader, err := NewMemoryReader(directory, pid, pageSize)
	if err != nil {
		t.Fatal(err)
	}
	memory, err := reader.GetMemPages(base, base+5*pageSize)
	if err != nil {
		t.Fatal(err)
	}
	expected := append(make([]byte, 2*pageSize), rawRegion...)
	expected = append(expected, finalPage...)
	if !bytes.Equal(memory.Bytes(), expected) {
		t.Fatal("decoded zero, raw, and final short region blocks do not match")
	}

	stats, err := InspectMemoryPages(directory, pageSize)
	if err != nil {
		t.Fatal(err)
	}
	if stats.PresentPages != 5 || stats.StoredBytes != uint64(len(payload)) ||
		stats.UncompressedBytes != 5*pageSize || !stats.HasCompressionMetadata {
		t.Fatalf("unexpected memory page stats: %+v", stats)
	}
}

func TestMemoryReaderCompressedPayloadAlignment(t *testing.T) {
	const (
		pageSize = 4096
		pid      = 30
		pagesID  = 8
		base     = 0x2a000
	)
	directory := t.TempDir()
	writeTestInventory(t, directory, crtoolsImagesV1_2, CompressionPerPage, 0)

	firstPage := bytes.Repeat([]byte{'A'}, pageSize)
	secondPage := bytes.Repeat([]byte{'B'}, pageSize)
	encodedFirst := compressTestBlock(t, firstPage)
	encodedSecond := compressTestBlock(t, secondPage)
	first := presentTestEntry(base, 1)
	first.CompressedSize = []uint32{uint32(len(encodedFirst))}
	second := presentTestEntry(base+pageSize, 1)
	second.Flags = uint32Pointer(pePresent | pePayloadAligned)
	second.CompressedSize = []uint32{uint32(len(encodedSecond))}
	writeTestPagemap(t, directory, pid, pagesID, first, second)

	payload := append([]byte(nil), encodedFirst...)
	padding := (pageSize - len(payload)%pageSize) % pageSize
	payload = append(payload, make([]byte, padding)...)
	payload = append(payload, encodedSecond...)
	if err := os.WriteFile(
		filepath.Join(directory, "pages-8.img"),
		payload,
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	reader, err := NewMemoryReader(directory, pid, pageSize)
	if err != nil {
		t.Fatal(err)
	}
	memory, err := reader.GetMemPages(base, base+2*pageSize)
	if err != nil {
		t.Fatal(err)
	}
	expected := append(append([]byte(nil), firstPage...), secondPage...)
	if !bytes.Equal(memory.Bytes(), expected) {
		t.Fatal("aligned compressed payload was read at the wrong offset")
	}

	stats, err := InspectMemoryPages(directory, pageSize)
	if err != nil {
		t.Fatal(err)
	}
	if stats.StoredBytes != uint64(len(payload)) {
		t.Fatalf("stored bytes = %d, want %d", stats.StoredBytes, len(payload))
	}
}

func TestMemoryReaderRejectsTruncatedOrOversizedPayload(t *testing.T) {
	const (
		pageSize = 4096
		pid      = 32
		base     = 0x36000
	)
	page := bytes.Repeat([]byte{'T'}, pageSize)
	encoded := compressTestBlock(t, page)

	for _, test := range []struct {
		name    string
		payload []byte
	}{
		{name: "truncated", payload: encoded[:len(encoded)-1]},
		{name: "oversized", payload: append(append([]byte(nil), encoded...), 0)},
	} {
		t.Run(test.name, func(t *testing.T) {
			directory := t.TempDir()
			writeTestInventory(t, directory, crtoolsImagesV1_2, CompressionPerPage, 0)
			entry := presentTestEntry(base, 1)
			entry.CompressedSize = []uint32{uint32(len(encoded))}
			entry.TotalCompressedSize = uint64Pointer(uint64(len(encoded)))
			writeTestPagemap(t, directory, pid, 9, entry)
			if err := os.WriteFile(
				filepath.Join(directory, "pages-9.img"),
				test.payload,
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
