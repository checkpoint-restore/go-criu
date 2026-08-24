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
	blockSize uint32,
) {
	t.Helper()
	entry := &inventory.InventoryEntry{
		ImgVersion: uint32Pointer(version),
		Compress:   uint32Pointer(uint32(mode)),
	}
	if blockSize != 0 {
		entry.CompressBlockSize = uint32Pointer(blockSize)
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
	writeTestInventory(t, directory, crtoolsImagesV1_2, CompressionBlock, 256*1024)

	info, err := ReadCompressionInfo(directory)
	if err != nil {
		t.Fatal(err)
	}
	if info.ImageVersion != crtoolsImagesV1_2 || info.Mode != CompressionBlock ||
		info.BlockSizeBytes != 256*1024 || !info.Compressed() {
		t.Fatalf("unexpected compression info: %+v", info)
	}
	if got := CompressionBlock.String(); got != "block" {
		t.Fatalf("unexpected mode string %q", got)
	}

	missing := filepath.Join(directory, "missing")
	if _, err := ReadCompressionInfo(missing); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing inventory error does not wrap os.ErrNotExist: %v", err)
	}

	invalid := t.TempDir()
	writeTestInventory(t, invalid, crtoolsImagesV1_1, CompressionBlock, 0)
	if _, err := ReadCompressionInfo(invalid); err == nil ||
		!strings.Contains(err.Error(), "cannot contain compressed") {
		t.Fatalf("expected version/mode error, got %v", err)
	}

	invalidBlock := t.TempDir()
	writeTestInventory(t, invalidBlock, crtoolsImagesV1_2, CompressionBlock, 123)
	if _, err := ReadCompressionInfo(invalidBlock); err == nil ||
		!strings.Contains(err.Error(), "invalid compression block size") {
		t.Fatalf("expected invalid block size error, got %v", err)
	}

	unexpectedBlock := t.TempDir()
	writeTestInventory(t, unexpectedBlock, crtoolsImagesV1_2, CompressionOff, 4096)
	if _, err := ReadCompressionInfo(unexpectedBlock); err == nil ||
		!strings.Contains(err.Error(), "while compression is off") {
		t.Fatalf("expected unexpected block size error, got %v", err)
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
		name      string
		version   uint32
		mode      CompressionMode
		blockSize uint32
		want      string
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
			name:      "block exceeds maximum",
			version:   crtoolsImagesV1_2,
			mode:      CompressionBlock,
			blockSize: uint32(maxCompressedBlockSize) + minimumCRIUPageSize,
			want:      "invalid compression block size",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			directory := t.TempDir()
			writeTestInventory(t, directory, test.version, test.mode, test.blockSize)
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
		{mode: CompressionBlock, want: "block"},
		{mode: CompressionMode(17), want: "unknown(17)"},
	}
	for _, test := range tests {
		if got := test.mode.String(); got != test.want {
			t.Errorf("CompressionMode(%d).String() = %q, want %q", test.mode, got, test.want)
		}
	}
}
