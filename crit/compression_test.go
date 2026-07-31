package crit

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/checkpoint-restore/go-criu/v8/crit/images/inventory"
)

func uint32Pointer(value uint32) *uint32 { return &value }

func writeTestImage(t *testing.T, path string, image *CriuImage) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := encodeImg(image, file); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}

func writeTestInventory(
	t *testing.T,
	directory string,
	version uint32,
	mode CompressionMode,
	regionSize uint32,
) {
	t.Helper()
	entry := &inventory.InventoryEntry{
		ImgVersion: uint32Pointer(version),
		Compress:   uint32Pointer(uint32(mode)),
	}
	if regionSize != 0 {
		entry.CompressRegionSize = uint32Pointer(regionSize)
	}
	writeTestImage(t, filepath.Join(directory, "inventory.img"), &CriuImage{
		Magic: "INVENTORY",
		Entries: []*CriuEntry{
			{Message: entry},
		},
	})
}

func TestReadCompressionInfo(t *testing.T) {
	directory := t.TempDir()
	writeTestInventory(t, directory, crtoolsImagesV1_2, CompressionRegion, 256*1024)

	info, err := ReadCompressionInfo(directory)
	if err != nil {
		t.Fatal(err)
	}
	if info.ImageVersion != crtoolsImagesV1_2 || info.Mode != CompressionRegion ||
		info.RegionSizeBytes != 256*1024 || !info.Compressed() {
		t.Fatalf("unexpected compression info: %+v", info)
	}
	if got := CompressionPerPage.String(); got != "per-page" {
		t.Fatalf("unexpected mode string %q", got)
	}

	missing := filepath.Join(directory, "missing")
	if _, err := ReadCompressionInfo(missing); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing inventory error does not wrap os.ErrNotExist: %v", err)
	}

	invalid := t.TempDir()
	writeTestInventory(t, invalid, crtoolsImagesV1_1, CompressionPerPage, 0)
	if _, err := ReadCompressionInfo(invalid); err == nil ||
		!strings.Contains(err.Error(), "cannot contain compressed") {
		t.Fatalf("expected version/mode error, got %v", err)
	}

	invalidRegion := t.TempDir()
	writeTestInventory(t, invalidRegion, crtoolsImagesV1_2, CompressionRegion, 123)
	if _, err := ReadCompressionInfo(invalidRegion); err == nil ||
		!strings.Contains(err.Error(), "invalid compression region size") {
		t.Fatalf("expected invalid region size error, got %v", err)
	}

	unexpectedRegion := t.TempDir()
	writeTestInventory(t, unexpectedRegion, crtoolsImagesV1_2, CompressionPerPage, 4096)
	if _, err := ReadCompressionInfo(unexpectedRegion); err == nil ||
		!strings.Contains(err.Error(), "outside region mode") {
		t.Fatalf("expected unexpected region size error, got %v", err)
	}

	wrongMagic := t.TempDir()
	writeTestImage(t, filepath.Join(wrongMagic, "inventory.img"), &CriuImage{
		Magic: "STATS",
		Entries: []*CriuEntry{{Message: &inventory.InventoryEntry{
			ImgVersion: uint32Pointer(crtoolsImagesV1_2),
		}}},
	})
	if _, err := ReadCompressionInfo(wrongMagic); err == nil ||
		!strings.Contains(err.Error(), "expected INVENTORY") {
		t.Fatalf("expected inventory magic error, got %v", err)
	}
}

func TestReadCompressionInfoRejectsMalformedInventory(t *testing.T) {
	tests := []struct {
		name       string
		version    uint32
		mode       CompressionMode
		regionSize uint32
		want       string
	}{
		{
			name:    "unsupported version",
			version: crtoolsImagesV1_2 + 1,
			mode:    CompressionOff,
			want:    "unsupported image version",
		},
		{
			name:    "unknown compression mode",
			version: crtoolsImagesV1_2,
			mode:    CompressionMode(99),
			want:    "invalid compression mode",
		},
		{
			name:       "region exceeds maximum",
			version:    crtoolsImagesV1_2,
			mode:       CompressionRegion,
			regionSize: uint32(maxCompressedRegionSize) + minimumCRIUPageSize,
			want:       "invalid compression region size",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			directory := t.TempDir()
			writeTestInventory(t, directory, test.version, test.mode, test.regionSize)
			if _, err := ReadCompressionInfo(directory); err == nil ||
				!strings.Contains(err.Error(), test.want) {
				t.Fatalf("ReadCompressionInfo() error = %v, want %q", err, test.want)
			}
		})
	}

	empty := t.TempDir()
	writeTestImage(t, filepath.Join(empty, "inventory.img"), &CriuImage{Magic: "INVENTORY"})
	if _, err := ReadCompressionInfo(empty); err == nil ||
		!strings.Contains(err.Error(), "has no entries") {
		t.Fatalf("ReadCompressionInfo() empty inventory error = %v", err)
	}
}

func TestCompressionModeString(t *testing.T) {
	tests := []struct {
		mode CompressionMode
		want string
	}{
		{mode: CompressionOff, want: "off"},
		{mode: CompressionPerPage, want: "per-page"},
		{mode: CompressionRegion, want: "region"},
		{mode: CompressionMode(17), want: "unknown(17)"},
	}
	for _, test := range tests {
		if got := test.mode.String(); got != test.want {
			t.Errorf("CompressionMode(%d).String() = %q, want %q", test.mode, got, test.want)
		}
	}
}
