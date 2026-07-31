package crit

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/checkpoint-restore/go-criu/v8/crit/images/pagemap"
)

func TestMemoryReaderParentChainAndCycle(t *testing.T) {
	const (
		pageSize = 4096
		pid      = 31
		base     = 0x30000
	)
	root := t.TempDir()
	parentDirectory := filepath.Join(root, "parent-checkpoint")
	childDirectory := filepath.Join(root, "child-checkpoint")
	if err := os.Mkdir(parentDirectory, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(childDirectory, 0o700); err != nil {
		t.Fatal(err)
	}
	writeTestInventory(t, parentDirectory, crtoolsImagesV1_2, CompressionPerPage, 0)
	// An uncompressed incremental child inherits V1.2 because its parent may
	// contain packed payloads even though the child's local mode is off.
	writeTestInventory(t, childDirectory, crtoolsImagesV1_2, CompressionOff, 0)

	parentPage := bytes.Repeat([]byte{'P'}, pageSize)
	encodedParentPage := compressTestBlock(t, parentPage)
	legacyPresent := &pagemap.PagemapEntry{
		Vaddr:               uint64Pointer(base),
		CompatNrPages:       uint32Pointer(1),
		CompressedSize:      []uint32{uint32(len(encodedParentPage))},
		TotalCompressedSize: uint64Pointer(uint64(len(encodedParentPage))),
	}
	writeTestPagemap(t, parentDirectory, pid, 1, legacyPresent)
	if err := os.WriteFile(filepath.Join(parentDirectory, "pages-1.img"), encodedParentPage, 0o600); err != nil {
		t.Fatal(err)
	}
	parentReference := &pagemap.PagemapEntry{
		Vaddr:         uint64Pointer(base),
		CompatNrPages: uint32Pointer(1),
		InParent:      boolPointer(true),
	}
	writeTestPagemap(t, childDirectory, pid, 2, parentReference)
	if err := os.Symlink("../parent-checkpoint", filepath.Join(childDirectory, "parent")); err != nil {
		t.Fatal(err)
	}

	reader, err := NewMemoryReader(childDirectory, pid, pageSize)
	if err != nil {
		t.Fatal(err)
	}
	memory, err := reader.GetMemPages(base, base+pageSize)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(memory.Bytes(), parentPage) {
		t.Fatal("parent-backed page does not match")
	}

	malformedChild := filepath.Join(root, "malformed-child")
	if err := os.Mkdir(malformedChild, 0o700); err != nil {
		t.Fatal(err)
	}
	writeTestInventory(t, malformedChild, crtoolsImagesV1_1, CompressionOff, 0)
	writeTestPagemap(t, malformedChild, pid, 3, parentReference)
	if err := os.Symlink("../parent-checkpoint", filepath.Join(malformedChild, "parent")); err != nil {
		t.Fatal(err)
	}
	malformedReader, err := NewMemoryReader(malformedChild, pid, pageSize)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := malformedReader.GetMemPages(base, base+pageSize); err == nil ||
		!strings.Contains(err.Error(), "does not propagate V1.2") {
		t.Fatalf("expected parent version propagation error, got %v", err)
	}

	missingInventoryChild := filepath.Join(root, "missing-inventory-child")
	if err := os.Mkdir(missingInventoryChild, 0o700); err != nil {
		t.Fatal(err)
	}
	writeTestPagemap(t, missingInventoryChild, pid, 4, parentReference)
	if err := os.Symlink("../parent-checkpoint", filepath.Join(missingInventoryChild, "parent")); err != nil {
		t.Fatal(err)
	}
	missingInventoryReader, err := NewMemoryReader(missingInventoryChild, pid, pageSize)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := missingInventoryReader.GetMemPages(base, base+pageSize); err == nil ||
		!strings.Contains(err.Error(), "image version missing does not propagate V1.2") {
		t.Fatalf("expected missing child inventory propagation error, got %v", err)
	}

	writeTestPagemap(t, parentDirectory, pid, 1, parentReference)
	if err := os.Symlink("../child-checkpoint", filepath.Join(parentDirectory, "parent")); err != nil {
		t.Fatal(err)
	}
	reader, err = NewMemoryReader(childDirectory, pid, pageSize)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := reader.GetMemPages(base, base+pageSize); err == nil ||
		!strings.Contains(err.Error(), "cycle") {
		t.Fatalf("expected parent cycle error, got %v", err)
	}
}

func TestMemoryReaderRejectsMissingPageInParent(t *testing.T) {
	const (
		pageSize = 4096
		pid      = 32
		base     = 0x38000
	)
	root := t.TempDir()
	parentDirectory := filepath.Join(root, "parent-checkpoint")
	childDirectory := filepath.Join(root, "child-checkpoint")
	for _, directory := range []string{parentDirectory, childDirectory} {
		if err := os.Mkdir(directory, 0o700); err != nil {
			t.Fatal(err)
		}
		writeTestInventory(t, directory, crtoolsImagesV1_2, CompressionOff, 0)
	}

	// A valid parent pagemap may still not contain the address delegated by
	// the child. That is corruption, not a zero-filled hole.
	writeTestPagemap(t, parentDirectory, pid, 1)
	parentReference := &pagemap.PagemapEntry{
		Vaddr:         uint64Pointer(base),
		CompatNrPages: uint32Pointer(1),
		InParent:      boolPointer(true),
	}
	writeTestPagemap(t, childDirectory, pid, 2, parentReference)
	if err := os.Symlink("../parent-checkpoint", filepath.Join(childDirectory, "parent")); err != nil {
		t.Fatal(err)
	}

	reader, err := NewMemoryReader(childDirectory, pid, pageSize)
	if err != nil {
		t.Fatal(err)
	}
	_, err = reader.GetMemPages(base, base+pageSize)
	if err == nil || !strings.Contains(err.Error(), "parent checkpoint has no page") {
		t.Fatalf("expected missing parent page error, got %v", err)
	}
}

func TestMemoryReaderMixedParentCompressionLayers(t *testing.T) {
	const (
		pageSize = 4096
		pid      = 33
		base     = 0x40000
	)
	root := t.TempDir()
	grandparentDirectory := filepath.Join(root, "grandparent-checkpoint")
	parentDirectory := filepath.Join(root, "parent-checkpoint")
	childDirectory := filepath.Join(root, "child-checkpoint")
	for _, directory := range []string{
		grandparentDirectory,
		parentDirectory,
		childDirectory,
	} {
		if err := os.Mkdir(directory, 0o700); err != nil {
			t.Fatal(err)
		}
	}

	grandparentPage := bytes.Repeat([]byte{'G'}, pageSize)
	encodedGrandparent := compressTestBlock(t, grandparentPage)
	writeTestInventory(
		t,
		grandparentDirectory,
		crtoolsImagesV1_2,
		CompressionRegion,
		pageSize,
	)
	grandparentEntry := presentTestEntry(base, 1)
	grandparentEntry.CompressedSize = []uint32{uint32(len(encodedGrandparent))}
	grandparentEntry.RegionPages = uint32Pointer(1)
	writeTestPagemap(t, grandparentDirectory, pid, 10, grandparentEntry)
	if err := os.WriteFile(
		filepath.Join(grandparentDirectory, "pages-10.img"),
		encodedGrandparent,
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	parentPage := bytes.Repeat([]byte{'P'}, pageSize)
	encodedParent := compressTestBlock(t, parentPage)
	writeTestInventory(t, parentDirectory, crtoolsImagesV1_2, CompressionPerPage, 0)
	grandparentReference := &pagemap.PagemapEntry{
		Vaddr:         uint64Pointer(base),
		CompatNrPages: uint32Pointer(1),
		InParent:      boolPointer(true),
	}
	parentEntry := presentTestEntry(base+pageSize, 1)
	parentEntry.CompressedSize = []uint32{uint32(len(encodedParent))}
	writeTestPagemap(
		t,
		parentDirectory,
		pid,
		11,
		grandparentReference,
		parentEntry,
	)
	if err := os.WriteFile(
		filepath.Join(parentDirectory, "pages-11.img"),
		encodedParent,
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(
		"../grandparent-checkpoint",
		filepath.Join(parentDirectory, "parent"),
	); err != nil {
		t.Fatal(err)
	}

	childPage := bytes.Repeat([]byte{'C'}, pageSize)
	writeTestInventory(t, childDirectory, crtoolsImagesV1_2, CompressionOff, 0)
	parentReference := &pagemap.PagemapEntry{
		Vaddr:         uint64Pointer(base),
		CompatNrPages: uint32Pointer(2),
		InParent:      boolPointer(true),
	}
	childEntry := presentTestEntry(base+2*pageSize, 1)
	writeTestPagemap(t, childDirectory, pid, 12, parentReference, childEntry)
	if err := os.WriteFile(
		filepath.Join(childDirectory, "pages-12.img"),
		childPage,
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(
		"../parent-checkpoint",
		filepath.Join(childDirectory, "parent"),
	); err != nil {
		t.Fatal(err)
	}

	reader, err := NewMemoryReader(childDirectory, pid, pageSize)
	if err != nil {
		t.Fatal(err)
	}
	memory, err := reader.GetMemPages(base, base+3*pageSize)
	if err != nil {
		t.Fatal(err)
	}
	expected := append(append([]byte(nil), grandparentPage...), parentPage...)
	expected = append(expected, childPage...)
	if !bytes.Equal(memory.Bytes(), expected) {
		t.Fatal("mixed parent compression layers do not match")
	}

	stats, err := InspectMemoryPages(childDirectory, pageSize)
	if err != nil {
		t.Fatal(err)
	}
	if stats.PresentPages != 1 || stats.StoredBytes != pageSize ||
		stats.UncompressedBytes != pageSize || !stats.HasParentReferences ||
		stats.HasCompressionMetadata {
		t.Fatalf("unexpected child-layer memory page stats: %+v", stats)
	}
}
