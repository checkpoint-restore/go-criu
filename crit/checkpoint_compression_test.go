//go:build linux

package crit

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/checkpoint-restore/go-criu/v8/crit/images/inventory"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/mm"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/pagemap"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/vma"
	"golang.org/x/sys/unix"
	"google.golang.org/protobuf/proto"
)

const compressionTestPageSize = 4096

func TestCompressAndDecompressCheckpoint(t *testing.T) {
	directory := t.TempDir()
	randomPage := compressionTestRandomPage(t)
	originalPages := append(make([]byte, compressionTestPageSize), bytes.Repeat([]byte{'A'}, compressionTestPageSize)...)
	originalPages = append(originalPages, randomPage...)
	compressionTestWriteCheckpoint(t, directory, CompressionOff, []*pagemap.PagemapEntry{
		compressionTestPagemapEntry(0x1000, 3),
	}, originalPages)

	originals := make(map[string][]byte)
	for _, name := range []string{"inventory.img", "pagemap-1.img", "pages-1.img"} {
		data, err := os.ReadFile(filepath.Join(directory, name))
		if err != nil {
			t.Fatal(err)
		}
		originals[name] = data
	}
	mode := os.FileMode(0o640)
	if err := os.Chmod(filepath.Join(directory, "pages-1.img"), mode); err != nil {
		t.Fatal(err)
	}
	wantTime := time.Unix(1_700_000_000, 123_456_789)
	if err := os.Chtimes(filepath.Join(directory, "pages-1.img"), wantTime, wantTime); err != nil {
		t.Fatal(err)
	}

	compressed, err := CompressCheckpoint(context.Background(), directory, CompressOptions{PageSize: compressionTestPageSize})
	if err != nil {
		t.Fatalf("CompressCheckpoint() error = %v", err)
	}
	if !compressed.Changed || compressed.Compression.Mode != CompressionBlock {
		t.Fatalf("CompressCheckpoint() = %+v", compressed)
	}
	if len(compressed.Pagemaps) != 1 || compressed.Pagemaps[0].Pages != 3 {
		t.Fatalf("unexpected compression statistics: %+v", compressed.Pagemaps)
	}

	for name, original := range originals {
		backup, err := os.ReadFile(filepath.Join(directory, name+".bak"))
		if err != nil {
			t.Fatalf("read %s backup: %v", name, err)
		}
		if !bytes.Equal(backup, original) {
			t.Errorf("%s backup differs from original", name)
		}
	}
	info, err := ReadCompressionInfo(directory)
	if err != nil || info.Mode != CompressionBlock || info.ImageVersion != crtoolsImagesV1_2 ||
		info.BlockSizeBytes != compressionTestPageSize {
		t.Fatalf("compression info = %+v, %v", info, err)
	}
	entries := compressionTestReadPagemap(t, directory)
	if got := entries[0].GetBlocks().GetBlockSizes(); len(got) != 3 || got[0] != 0 || got[1] >= compressionTestPageSize || got[2] != compressionTestPageSize {
		t.Fatalf("compressed sizes = %v", got)
	}
	stat, err := os.Stat(filepath.Join(directory, "pages-1.img"))
	if err != nil {
		t.Fatal(err)
	}
	if stat.Mode().Perm() != mode {
		t.Errorf("pages mode = %v, want %v", stat.Mode().Perm(), mode)
	}
	if !stat.ModTime().Equal(wantTime) {
		t.Errorf("pages mtime = %v, want %v", stat.ModTime(), wantTime)
	}

	decompressed, err := DecompressCheckpoint(context.Background(), directory, DecompressOptions{
		PageSize: compressionTestPageSize,
		InPlace:  true,
	})
	if err != nil {
		t.Fatalf("DecompressCheckpoint() error = %v", err)
	}
	if !decompressed.Changed || decompressed.Compression.Mode != CompressionOff {
		t.Fatalf("DecompressCheckpoint() = %+v", decompressed)
	}
	gotPages, err := os.ReadFile(filepath.Join(directory, "pages-1.img"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotPages, originalPages) {
		t.Fatal("decompressed pages differ from original")
	}
	info, err = ReadCompressionInfo(directory)
	if err != nil || info.Mode != CompressionOff || info.ImageVersion != crtoolsImagesV1_1 {
		t.Fatalf("decompressed info = %+v, %v", info, err)
	}
	for _, entry := range compressionTestReadPagemap(t, directory) {
		if hasPagemapCompressionMetadata(entry) || effectivePagemapFlags(entry)&pePayloadAligned != 0 {
			t.Fatalf("compression metadata remains after decompression: %+v", entry)
		}
	}
	if matches, err := filepath.Glob(filepath.Join(directory, ".*.crit-rollback-*")); err != nil || len(matches) != 0 {
		t.Fatalf("rollback files remain: %v, %v", matches, err)
	}
}

func TestCheckpointCompressionBackupPolicy(t *testing.T) {
	for _, operation := range []string{"compress", "decompress"} {
		for _, inPlace := range []bool{false, true} {
			name := operation + map[bool]string{false: "/backups", true: "/in-place"}[inPlace]
			t.Run(name, func(t *testing.T) {
				directory := t.TempDir()
				compressionTestWriteCheckpoint(t, directory, CompressionOff, []*pagemap.PagemapEntry{
					compressionTestPagemapEntry(0x1000, 1),
				}, bytes.Repeat([]byte{'A'}, compressionTestPageSize))
				if operation == "decompress" {
					if _, err := CompressCheckpoint(context.Background(), directory, CompressOptions{
						PageSize: compressionTestPageSize,
						InPlace:  true,
					}); err != nil {
						t.Fatal(err)
					}
				}

				imageNames := []string{"inventory.img", "pagemap-1.img", "pages-1.img"}
				before := compressionTestReadFiles(t, directory, imageNames)
				var result CheckpointCompressionResult
				var err error
				if operation == "compress" {
					result, err = CompressCheckpoint(context.Background(), directory, CompressOptions{
						PageSize: compressionTestPageSize,
						InPlace:  inPlace,
					})
				} else {
					result, err = DecompressCheckpoint(context.Background(), directory, DecompressOptions{
						PageSize: compressionTestPageSize,
						InPlace:  inPlace,
					})
				}
				if err != nil {
					t.Fatal(err)
				}
				if !result.Changed {
					t.Fatalf("%s result = %+v", operation, result)
				}

				for _, imageName := range imageNames {
					backupPath := filepath.Join(directory, imageName+".bak")
					backup, readErr := os.ReadFile(backupPath)
					if inPlace {
						if !errors.Is(readErr, os.ErrNotExist) {
							t.Errorf("in-place %s left %s: %v", operation, backupPath, readErr)
						}
						continue
					}
					if readErr != nil {
						t.Errorf("read %s: %v", backupPath, readErr)
						continue
					}
					if !bytes.Equal(backup, before[imageName]) {
						t.Errorf("%s does not contain the original image", backupPath)
					}
				}
				compressionTestAssertNoTemporaryFiles(t, directory)
			})
		}
	}
}

func TestCheckpointCompressionPreservesMetadata(t *testing.T) {
	for _, operation := range []string{"compress", "decompress"} {
		t.Run(operation, func(t *testing.T) {
			directory := t.TempDir()
			compressionTestWriteCheckpoint(t, directory, CompressionOff, []*pagemap.PagemapEntry{
				compressionTestPagemapEntry(0x1000, 1),
			}, bytes.Repeat([]byte{'A'}, compressionTestPageSize))
			if operation == "decompress" {
				if _, err := CompressCheckpoint(context.Background(), directory, CompressOptions{
					PageSize: compressionTestPageSize,
					InPlace:  true,
				}); err != nil {
					t.Fatal(err)
				}
			}

			imageNames := []string{"inventory.img", "pagemap-1.img", "pages-1.img"}
			expected := make(map[string]compressionTestMetadata, len(imageNames))
			for index, imageName := range imageNames {
				path := filepath.Join(directory, imageName)
				expected[imageName] = compressionTestApplyMetadata(t, path, index)
			}

			var err error
			if operation == "compress" {
				_, err = CompressCheckpoint(context.Background(), directory, CompressOptions{
					PageSize: compressionTestPageSize,
					InPlace:  true,
				})
			} else {
				_, err = DecompressCheckpoint(context.Background(), directory, DecompressOptions{
					PageSize: compressionTestPageSize,
					InPlace:  true,
				})
			}
			if err != nil {
				t.Fatal(err)
			}
			for _, imageName := range imageNames {
				compressionTestCheckMetadata(t, filepath.Join(directory, imageName), expected[imageName])
			}
		})
	}
}

func TestCheckpointCompressionIdempotence(t *testing.T) {
	for _, operation := range []string{"compress", "decompress"} {
		t.Run(operation, func(t *testing.T) {
			directory := t.TempDir()
			compressionTestWriteCheckpoint(t, directory, CompressionOff, []*pagemap.PagemapEntry{
				compressionTestPagemapEntry(0x1000, 1),
			}, bytes.Repeat([]byte{'A'}, compressionTestPageSize))
			if operation == "compress" {
				if _, err := CompressCheckpoint(context.Background(), directory, CompressOptions{
					PageSize: compressionTestPageSize,
					InPlace:  true,
				}); err != nil {
					t.Fatal(err)
				}
			}

			imageNames := []string{"inventory.img", "pagemap-1.img", "pages-1.img"}
			before := compressionTestReadFiles(t, directory, imageNames)
			var result CheckpointCompressionResult
			var err error
			if operation == "compress" {
				result, err = CompressCheckpoint(context.Background(), directory, CompressOptions{
					PageSize: compressionTestPageSize,
				})
			} else {
				result, err = DecompressCheckpoint(context.Background(), directory, DecompressOptions{
					PageSize: compressionTestPageSize,
				})
			}
			if err != nil {
				t.Fatal(err)
			}
			if result.Changed || !result.AlreadyInRequestedState || len(result.Pagemaps) != 0 {
				t.Fatalf("idempotent %s result = %+v", operation, result)
			}
			if operation == "compress" && result.Compression.Mode != CompressionBlock {
				t.Fatalf("compression result = %+v", result.Compression)
			}
			if operation == "decompress" && result.Compression.Mode != CompressionOff {
				t.Fatalf("compression result = %+v", result.Compression)
			}
			for imageName, want := range before {
				got, readErr := os.ReadFile(filepath.Join(directory, imageName))
				if readErr != nil {
					t.Fatal(readErr)
				}
				if !bytes.Equal(got, want) {
					t.Errorf("idempotent %s changed %s", operation, imageName)
				}
				if _, statErr := os.Lstat(filepath.Join(directory, imageName+".bak")); !errors.Is(statErr, os.ErrNotExist) {
					t.Errorf("idempotent %s created %s.bak: %v", operation, imageName, statErr)
				}
			}
			compressionTestAssertNoTemporaryFiles(t, directory)
		})
	}
}

func TestCompressAllRawElidesMetadataAndAlignsPayload(t *testing.T) {
	directory := t.TempDir()
	original := compressionTestRandomPage(t)
	compressionTestWriteCheckpoint(t, directory, CompressionOff, []*pagemap.PagemapEntry{
		compressionTestPagemapEntry(0x1000, 1),
	}, original)

	if _, err := CompressCheckpoint(context.Background(), directory, CompressOptions{
		PageSize: compressionTestPageSize,
		InPlace:  true,
	}); err != nil {
		t.Fatal(err)
	}
	entry := compressionTestReadPagemap(t, directory)[0]
	if hasPagemapCompressionMetadata(entry) {
		t.Fatalf("all-raw entry retained compression metadata: %+v", entry)
	}
	if effectivePagemapFlags(entry)&pePayloadAligned == 0 {
		t.Fatal("all-raw entry does not carry PE_PAYLOAD_ALIGNED")
	}
	got, err := os.ReadFile(filepath.Join(directory, "pages-1.img"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, original) {
		t.Fatal("all-raw payload changed")
	}

	if _, err := DecompressCheckpoint(context.Background(), directory, DecompressOptions{
		PageSize: compressionTestPageSize,
		InPlace:  true,
	}); err != nil {
		t.Fatal(err)
	}
	entry = compressionTestReadPagemap(t, directory)[0]
	if effectivePagemapFlags(entry)&pePayloadAligned != 0 {
		t.Fatal("decompression did not clear PE_PAYLOAD_ALIGNED")
	}
}

func TestCompressionAlignsFirstRawPayloadOfLaterEntry(t *testing.T) {
	directory := t.TempDir()
	compressedPage := bytes.Repeat([]byte{'A'}, compressionTestPageSize)
	rawPage := compressionTestRandomPage(t)
	original := append(append([]byte(nil), compressedPage...), rawPage...)
	compressionTestWriteCheckpoint(t, directory, CompressionOff, []*pagemap.PagemapEntry{
		compressionTestPagemapEntry(0x1000, 1),
		compressionTestPagemapEntry(0x2000, 1),
	}, original)

	if _, err := CompressCheckpoint(context.Background(), directory, CompressOptions{
		PageSize: compressionTestPageSize,
		InPlace:  true,
	}); err != nil {
		t.Fatal(err)
	}
	entries := compressionTestReadPagemap(t, directory)
	if len(entries[0].GetBlocks().GetBlockSizes()) != 1 {
		t.Fatalf("first entry was not compressed: %+v", entries[0])
	}
	if hasPagemapCompressionMetadata(entries[1]) {
		t.Fatalf("raw entry retained compression metadata: %+v", entries[1])
	}
	if effectivePagemapFlags(entries[1])&pePayloadAligned == 0 {
		t.Fatal("later raw entry is not payload-aligned")
	}
	stored, err := os.ReadFile(filepath.Join(directory, "pages-1.img"))
	if err != nil {
		t.Fatal(err)
	}
	if len(stored) != 2*compressionTestPageSize {
		t.Fatalf("aligned pages image size = %d, want %d", len(stored), 2*compressionTestPageSize)
	}
	if !bytes.Equal(stored[compressionTestPageSize:], rawPage) {
		t.Fatal("raw payload does not begin at the aligned offset")
	}

	if _, err := DecompressCheckpoint(context.Background(), directory, DecompressOptions{
		PageSize: compressionTestPageSize,
		InPlace:  true,
	}); err != nil {
		t.Fatal(err)
	}
	decoded, err := os.ReadFile(filepath.Join(directory, "pages-1.img"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(decoded, original) {
		t.Fatal("aligned round trip changed page data")
	}
}

func TestDecompressCheckpointMultiPageBlocks(t *testing.T) {
	directory := t.TempDir()
	decoded := append(bytes.Repeat([]byte{'x'}, compressionTestPageSize), bytes.Repeat([]byte{'y'}, compressionTestPageSize)...)
	// Generated by python-lz4's raw block API with store_size=False. Keeping
	// this as a fixed vector verifies compatibility with Python CRIT rather
	// than round-tripping through the same Go encoder.
	encoded, err := hex.DecodeString("1f780100fffffffffffffffffffffffffffffffb1f790100fffffffffffffffffffffffffffffff6507979797979")
	if err != nil {
		t.Fatal(err)
	}
	entry := compressionTestPagemapEntry(0x1000, 3)
	entry.Blocks = testPagemapBlocks([]uint32{uint32(len(encoded)), 0}, 2)
	compressionTestWriteCheckpoint(t, directory, CompressionBlock, []*pagemap.PagemapEntry{entry}, encoded)

	result, err := DecompressCheckpoint(context.Background(), directory, DecompressOptions{
		PageSize: compressionTestPageSize,
		InPlace:  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Pagemaps) != 1 || result.Pagemaps[0].Pages != 3 {
		t.Fatalf("unexpected statistics: %+v", result.Pagemaps)
	}
	want := append(decoded, make([]byte, compressionTestPageSize)...)
	got, err := os.ReadFile(filepath.Join(directory, "pages-1.img"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatal("multi-page block decompression produced incorrect pages")
	}
}

func TestDecompressCheckpointFinalShortCompressedBlock(t *testing.T) {
	directory := t.TempDir()
	firstBlock := append(bytes.Repeat([]byte{'A'}, compressionTestPageSize), bytes.Repeat([]byte{'B'}, compressionTestPageSize)...)
	finalPage := bytes.Repeat([]byte{'C'}, compressionTestPageSize)
	encodedFinalPage := compressTestBlock(t, finalPage)
	entry := compressionTestPagemapEntry(0x1000, 3)
	entry.Blocks = testPagemapBlocks([]uint32{uint32(len(firstBlock)), uint32(len(encodedFinalPage))}, 2)
	stored := append(append([]byte(nil), firstBlock...), encodedFinalPage...)
	compressionTestWriteCheckpoint(t, directory, CompressionBlock, []*pagemap.PagemapEntry{entry}, stored)

	result, err := DecompressCheckpoint(context.Background(), directory, DecompressOptions{
		PageSize: compressionTestPageSize,
		InPlace:  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Pagemaps) != 1 || result.Pagemaps[0].InputBytes != uint64(len(stored)) ||
		result.Pagemaps[0].OutputBytes != 3*compressionTestPageSize {
		t.Fatalf("unexpected statistics: %+v", result.Pagemaps)
	}
	want := append(append([]byte(nil), firstBlock...), finalPage...)
	got, err := os.ReadFile(filepath.Join(directory, "pages-1.img"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatal("decompression mishandled the short final block")
	}
}

func TestDecompressCheckpointRawAndZeroRuns(t *testing.T) {
	directory := t.TempDir()
	rawOne := compressionTestRandomPage(t)
	rawTwo := append([]byte(nil), rawOne...)
	rawTwo[0]++
	compressedPage := bytes.Repeat([]byte{'C'}, compressionTestPageSize)
	encodedPage := compressTestBlock(t, compressedPage)
	rawThree := append([]byte(nil), rawTwo...)
	rawThree[0]++

	entry := compressionTestPagemapEntry(0x1000, 6)
	entry.Blocks = testPagemapBlocks([]uint32{
		compressionTestPageSize,
		compressionTestPageSize,
		0,
		0,
		uint32(len(encodedPage)),
		compressionTestPageSize,
	}, 1)
	stored := append(append(append([]byte(nil), rawOne...), rawTwo...), encodedPage...)
	stored = append(stored, rawThree...)
	compressionTestWriteCheckpoint(t, directory, CompressionBlock, []*pagemap.PagemapEntry{entry}, stored)

	result, err := DecompressCheckpoint(context.Background(), directory, DecompressOptions{
		PageSize: compressionTestPageSize,
		InPlace:  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Pagemaps) != 1 || result.Pagemaps[0].Pages != 6 ||
		result.Pagemaps[0].InputBytes != uint64(len(stored)) ||
		result.Pagemaps[0].OutputBytes != 6*compressionTestPageSize {
		t.Fatalf("unexpected statistics: %+v", result.Pagemaps)
	}

	want := append(append([]byte(nil), rawOne...), rawTwo...)
	want = append(want, make([]byte, 2*compressionTestPageSize)...)
	want = append(want, compressedPage...)
	want = append(want, rawThree...)
	got, err := os.ReadFile(filepath.Join(directory, "pages-1.img"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatal("decompression changed a raw, zero, or compressed run")
	}
}

func TestCompressionAddressInRanges(t *testing.T) {
	ranges := []compressionRange{
		{start: 0x1000, end: 0x2000},
		{start: 0x90000000, end: 0x90001000},
	}
	tests := []struct {
		name    string
		address uint64
		want    bool
	}{
		{name: "first range start", address: 0x1000, want: true},
		{name: "first range interior", address: 0x1800, want: true},
		{name: "half-open range end", address: 0x2000, want: false},
		{name: "before first range", address: 0x0fff, want: false},
		{name: "between ranges", address: 0x80000000, want: false},
		{name: "unsigned 32-bit address", address: 0x90000000, want: true},
		{name: "sign-extended 32-bit address", address: 0xffffffff90000000, want: true},
		{name: "sign-extended range interior", address: 0xffffffff90000800, want: true},
		{name: "sign-extended range end", address: 0xffffffff90001000, want: false},
		{name: "different high bits", address: 0xfffffffe90000000, want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := compressionAddressInRanges(test.address, ranges); got != test.want {
				t.Fatalf("compressionAddressInRanges(%#x) = %t, want %t", test.address, got, test.want)
			}
		})
	}
}

func TestCompressCheckpointKeepsExceptionalSharedVMABytesRaw(t *testing.T) {
	directory := t.TempDir()
	payload := bytes.Repeat([]byte{'A'}, compressionTestPageSize)
	compressionTestWriteCheckpointNamed(t, directory, "pagemap-shmem-42.img", 1, CompressionOff, []*pagemap.PagemapEntry{
		compressionTestPagemapEntry(0, 1),
	}, payload)
	compressionTestWriteMM(t, filepath.Join(directory, "mm-1.img"), &vma.VmaEntry{
		Start: proto.Uint64(0x1000), End: proto.Uint64(0x2000),
		Pgoff: proto.Uint64(0), Shmid: proto.Uint64(42),
		Prot: proto.Uint32(0), Flags: proto.Uint32(compressionMapShared | compressionMapHugetlb),
		Status: proto.Uint32(compressionVMAAnonShared), Fd: proto.Int64(-1),
	})

	if _, err := CompressCheckpoint(context.Background(), directory, CompressOptions{
		PageSize: compressionTestPageSize,
		InPlace:  true,
	}); err != nil {
		t.Fatal(err)
	}
	image, err := getImg(filepath.Join(directory, "pagemap-shmem-42.img"), nil)
	if err != nil {
		t.Fatal(err)
	}
	entry := image.Entries[1].Message.(*pagemap.PagemapEntry)
	if hasPagemapCompressionMetadata(entry) {
		t.Fatalf("exceptional VMA was compressed: %+v", entry)
	}
	got, err := os.ReadFile(filepath.Join(directory, "pages-1.img"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatal("exceptional VMA payload changed")
	}
}

func TestCompressCheckpointRejectsInvalidPayloadWithoutChanges(t *testing.T) {
	directory := t.TempDir()
	compressionTestWriteCheckpoint(t, directory, CompressionOff, []*pagemap.PagemapEntry{
		compressionTestPagemapEntry(0x1000, 1),
	}, make([]byte, compressionTestPageSize-1))
	originalInventory, err := os.ReadFile(filepath.Join(directory, "inventory.img"))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := CompressCheckpoint(context.Background(), directory, CompressOptions{PageSize: compressionTestPageSize}); err == nil || !strings.Contains(err.Error(), "describes 4096 payload bytes") {
		t.Fatalf("CompressCheckpoint() error = %v", err)
	}
	gotInventory, err := os.ReadFile(filepath.Join(directory, "inventory.img"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotInventory, originalInventory) {
		t.Fatal("inventory changed after validation failure")
	}
	compressionTestAssertNoTemporaryFiles(t, directory)
}

func TestCheckpointCompressionNoOpValidatesCheckpoint(t *testing.T) {
	for _, operation := range []string{"compress", "decompress"} {
		t.Run(operation, func(t *testing.T) {
			directory := t.TempDir()
			compressionTestWriteCheckpoint(t, directory, CompressionOff, []*pagemap.PagemapEntry{
				compressionTestPagemapEntry(0x1000, 1),
			}, bytes.Repeat([]byte{'A'}, compressionTestPageSize))
			if operation == "compress" {
				if _, err := CompressCheckpoint(context.Background(), directory, CompressOptions{
					PageSize: compressionTestPageSize,
					InPlace:  true,
				}); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.Truncate(filepath.Join(directory, "pages-1.img"), 0); err != nil {
				t.Fatal(err)
			}

			var result CheckpointCompressionResult
			var err error
			if operation == "compress" {
				result, err = CompressCheckpoint(context.Background(), directory, CompressOptions{PageSize: compressionTestPageSize})
			} else {
				result, err = DecompressCheckpoint(context.Background(), directory, DecompressOptions{PageSize: compressionTestPageSize})
			}
			if err == nil || !strings.Contains(err.Error(), "describes") {
				t.Fatalf("%s no-op error = %v, want payload-size validation error", operation, err)
			}
			if result.Changed || result.AlreadyInRequestedState {
				t.Fatalf("%s no-op result = %+v", operation, result)
			}
			compressionTestAssertNoTemporaryFiles(t, directory)
		})
	}
}

func TestCompressCheckpointRejectsMissingPagemapHead(t *testing.T) {
	directory := t.TempDir()
	compressionTestWriteCheckpoint(t, directory, CompressionOff, []*pagemap.PagemapEntry{
		compressionTestPagemapEntry(0x1000, 1),
	}, bytes.Repeat([]byte{'A'}, compressionTestPageSize))
	compressionTestWriteImage(t, filepath.Join(directory, "pagemap-1.img"), &CriuImage{Magic: "PAGEMAP"})

	result, err := CompressCheckpoint(context.Background(), directory, CompressOptions{PageSize: compressionTestPageSize})
	if err == nil || !strings.Contains(err.Error(), "no pagemap head") {
		t.Fatalf("CompressCheckpoint() error = %v, want missing-head error", err)
	}
	if result.Changed {
		t.Fatalf("CompressCheckpoint() result = %+v", result)
	}
	compressionTestAssertNoTemporaryFiles(t, directory)
}

func TestCompressCheckpointRejectsUnorderedPagemapEntries(t *testing.T) {
	tests := []struct {
		name    string
		entries []*pagemap.PagemapEntry
	}{
		{
			name: "overlap",
			entries: []*pagemap.PagemapEntry{
				compressionTestPagemapEntry(0x1000, 2),
				compressionTestPagemapEntry(0x2000, 1),
			},
		},
		{
			name: "unsorted",
			entries: []*pagemap.PagemapEntry{
				compressionTestPagemapEntry(0x2000, 1),
				compressionTestPagemapEntry(0x1000, 1),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			directory := t.TempDir()
			payload := bytes.Repeat([]byte{'A'}, len(test.entries)*compressionTestPageSize)
			if test.name == "overlap" {
				payload = append(payload, bytes.Repeat([]byte{'A'}, compressionTestPageSize)...)
			}
			compressionTestWriteCheckpoint(t, directory, CompressionOff, test.entries, payload)
			imageNames := []string{"inventory.img", "pagemap-1.img", "pages-1.img"}
			originals := compressionTestReadFiles(t, directory, imageNames)

			_, err := CompressCheckpoint(context.Background(), directory, CompressOptions{
				PageSize: compressionTestPageSize,
				InPlace:  true,
			})
			if err == nil || !strings.Contains(err.Error(), "entries are not sorted or overlap") {
				t.Fatalf("CompressCheckpoint() error = %v", err)
			}
			for imageName, original := range originals {
				got, readErr := os.ReadFile(filepath.Join(directory, imageName))
				if readErr != nil {
					t.Fatal(readErr)
				}
				if !bytes.Equal(got, original) {
					t.Errorf("%s changed after validation failure", imageName)
				}
			}
			compressionTestAssertNoTemporaryFiles(t, directory)
		})
	}
}

func TestCompressCheckpointRejectsMissingMMImage(t *testing.T) {
	directory := t.TempDir()
	compressionTestWriteCheckpoint(t, directory, CompressionOff, []*pagemap.PagemapEntry{
		compressionTestPagemapEntry(0x1000, 1),
	}, bytes.Repeat([]byte{'A'}, compressionTestPageSize))
	if err := os.Remove(filepath.Join(directory, "mm-1.img")); err != nil {
		t.Fatal(err)
	}
	original, err := os.ReadFile(filepath.Join(directory, "pages-1.img"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := CompressCheckpoint(context.Background(), directory, CompressOptions{
		PageSize: compressionTestPageSize,
		InPlace:  true,
	}); err == nil || !strings.Contains(err.Error(), "requires missing mm-1.img") {
		t.Fatalf("CompressCheckpoint() error = %v", err)
	}
	got, err := os.ReadFile(filepath.Join(directory, "pages-1.img"))
	if err != nil || !bytes.Equal(got, original) {
		t.Fatalf("pages changed after missing-mm rejection: %v", err)
	}
	compressionTestAssertNoTemporaryFiles(t, directory)
}

func TestCompressCheckpointBackupCollisionRollsBackPreflight(t *testing.T) {
	directory := t.TempDir()
	compressionTestWriteCheckpoint(t, directory, CompressionOff, []*pagemap.PagemapEntry{
		compressionTestPagemapEntry(0x1000, 1),
	}, bytes.Repeat([]byte{'A'}, compressionTestPageSize))
	marker := []byte("existing backup")
	if err := os.WriteFile(filepath.Join(directory, "inventory.img.bak"), marker, 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := CompressCheckpoint(context.Background(), directory, CompressOptions{PageSize: compressionTestPageSize}); err == nil || !strings.Contains(err.Error(), "refusing to overwrite existing backup") {
		t.Fatalf("CompressCheckpoint() error = %v", err)
	}
	for _, name := range []string{"pages-1.img.bak", "pagemap-1.img.bak"} {
		if _, err := os.Lstat(filepath.Join(directory, name)); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("temporary backup %s remains: %v", name, err)
		}
	}
	got, err := os.ReadFile(filepath.Join(directory, "inventory.img.bak"))
	if err != nil || !bytes.Equal(got, marker) {
		t.Fatalf("existing backup changed: %q, %v", got, err)
	}
	compressionTestAssertNoTemporaryFiles(t, directory)
}

func TestCompressCheckpointRejectsUnsupportedAcceleration(t *testing.T) {
	_, err := CompressCheckpoint(context.Background(), t.TempDir(), CompressOptions{Acceleration: 2})
	if !errors.Is(err, ErrUnsupportedAcceleration) {
		t.Fatalf("CompressCheckpoint() error = %v", err)
	}
}

func TestCheckpointCompressionRejectsInvalidInventoryBlockMetadata(t *testing.T) {
	tests := []struct {
		name  string
		entry *inventory.InventoryEntry
	}{
		{
			name: "outside-block-mode",
			entry: &inventory.InventoryEntry{
				ImgVersion: proto.Uint32(crtoolsImagesV1_1), CompressBlockSize: proto.Uint32(compressionTestPageSize),
			},
		},
		{
			name: "unaligned",
			entry: &inventory.InventoryEntry{
				ImgVersion: proto.Uint32(crtoolsImagesV1_2), Compress: proto.Uint32(uint32(CompressionBlock)), CompressBlockSize: proto.Uint32(compressionTestPageSize + 1),
			},
		},
		{
			name: "too-large",
			entry: &inventory.InventoryEntry{
				ImgVersion: proto.Uint32(crtoolsImagesV1_2), Compress: proto.Uint32(uint32(CompressionBlock)), CompressBlockSize: proto.Uint32(uint32(maxCompressedBlockSize) + compressionTestPageSize),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			directory := t.TempDir()
			compressionTestWriteImage(t, filepath.Join(directory, "inventory.img"), &CriuImage{
				Magic: "INVENTORY", Entries: []*CriuEntry{{Message: test.entry}},
			})
			if _, err := DecompressCheckpoint(context.Background(), directory, DecompressOptions{InPlace: true}); err == nil {
				t.Fatal("DecompressCheckpoint() unexpectedly accepted invalid inventory")
			}
			compressionTestAssertNoTemporaryFiles(t, directory)
		})
	}
}

func TestCheckpointCompressionRejectsWrongInventoryMagicWithoutChanges(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "inventory.img")
	compressionTestWriteImage(t, path, &CriuImage{
		Magic: "APPARMOR",
		Entries: []*CriuEntry{{Message: &inventory.InventoryEntry{
			ImgVersion: proto.Uint32(crtoolsImagesV1_1),
		}}},
	})
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := CompressCheckpoint(context.Background(), directory, CompressOptions{InPlace: true}); err == nil || !strings.Contains(err.Error(), "expected INVENTORY") {
		t.Fatalf("CompressCheckpoint() error = %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, original) {
		t.Fatal("wrong-magic inventory changed after rejection")
	}
	compressionTestAssertNoTemporaryFiles(t, directory)
}

func TestDecompressCheckpointKeepsV12WithParentReference(t *testing.T) {
	directory := t.TempDir()
	entry := compressionTestPagemapEntry(0x1000, 1)
	entry.Blocks = testPagemapBlocks([]uint32{0}, 1)
	compressionTestWriteCheckpoint(t, directory, CompressionBlock, []*pagemap.PagemapEntry{entry}, nil)
	if err := os.WriteFile(filepath.Join(directory, "parent"), nil, 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := DecompressCheckpoint(context.Background(), directory, DecompressOptions{
		PageSize: compressionTestPageSize,
		InPlace:  true,
	}); err != nil {
		t.Fatal(err)
	}
	info, err := ReadCompressionInfo(directory)
	if err != nil {
		t.Fatal(err)
	}
	if info.ImageVersion != crtoolsImagesV1_2 || info.Mode != CompressionOff {
		t.Fatalf("compression info = %+v", info)
	}
}

func TestCompressCheckpointMultiplePagemapsCommitRollback(t *testing.T) {
	directory := t.TempDir()
	compressionTestWriteCheckpointNamed(t, directory, "pagemap-1.img", 1, CompressionOff, []*pagemap.PagemapEntry{
		compressionTestPagemapEntry(0x1000, 1),
	}, bytes.Repeat([]byte{'A'}, compressionTestPageSize))
	compressionTestWriteCheckpointNamed(t, directory, "pagemap-2.img", 2, CompressionOff, []*pagemap.PagemapEntry{
		compressionTestPagemapEntry(0x2000, 1),
	}, bytes.Repeat([]byte{'B'}, compressionTestPageSize))
	imageNames := []string{
		"inventory.img", "mm-1.img", "mm-2.img",
		"pagemap-1.img", "pages-1.img", "pagemap-2.img", "pages-2.img",
	}
	originals := compressionTestReadFiles(t, directory, imageNames)
	ctx := &compressionTestCancelAfterFileChangesContext{
		path:     filepath.Join(directory, "pages-1.img"),
		original: originals["pages-1.img"],
	}

	result, err := CompressCheckpoint(ctx, directory, CompressOptions{PageSize: compressionTestPageSize})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("CompressCheckpoint() error = %v", err)
	}
	if result.Changed {
		t.Fatalf("failed compression result = %+v", result)
	}
	for imageName, original := range originals {
		path := filepath.Join(directory, imageName)
		got, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Fatal(readErr)
		}
		if !bytes.Equal(got, original) {
			t.Errorf("%s was not rolled back", imageName)
		}
		if _, statErr := os.Lstat(path + ".bak"); !errors.Is(statErr, os.ErrNotExist) {
			t.Errorf("failed transaction left %s.bak: %v", imageName, statErr)
		}
	}
	compressionTestAssertNoTemporaryFiles(t, directory)
}

func TestCheckpointCompressionCommitRollsBackCancellation(t *testing.T) {
	for _, inPlace := range []bool{false, true} {
		t.Run(map[bool]string{false: "backups", true: "in-place"}[inPlace], func(t *testing.T) {
			directory := t.TempDir()
			originals := map[string][]byte{
				filepath.Join(directory, "one.img"): []byte("original one"),
				filepath.Join(directory, "two.img"): []byte("original two"),
			}
			staged := make([]compressionStagedFile, 0, len(originals))
			for path, original := range originals {
				if err := os.WriteFile(path, original, 0o600); err != nil {
					t.Fatal(err)
				}
				metadata, err := compressionCaptureFileMetadata(path)
				if err != nil {
					t.Fatal(err)
				}
				stage, err := compressionStageFile(path, metadata, func(output *os.File) error {
					_, err := output.Write([]byte("replacement"))
					return err
				})
				if err != nil {
					t.Fatal(err)
				}
				staged = append(staged, stage)
			}
			defer func() {
				if err := compressionCleanupStaged(staged); err != nil {
					t.Errorf("clean up staged images: %v", err)
				}
			}()

			// Err becomes context.Canceled immediately before the second
			// replacement, after the first rename has completed.
			ctx := &compressionTestCancelContext{cancelAt: 3}
			err := compressionCommit(ctx, staged, inPlace)
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("compressionCommit() error = %v", err)
			}
			for path, original := range originals {
				got, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(got, original) {
					t.Fatalf("%s was not rolled back: %q", filepath.Base(path), got)
				}
				if _, err := os.Lstat(path + ".bak"); !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("failed transaction left backup %s: %v", path+".bak", err)
				}
			}
			if err := compressionCleanupStaged(staged); err != nil {
				t.Fatalf("clean up staged images: %v", err)
			}
			compressionTestAssertNoTemporaryFiles(t, directory)
		})
	}
}

func TestCheckpointCompressionCommitTreatsFinalSyncAsCommitPoint(t *testing.T) {
	for _, inPlace := range []bool{false, true} {
		t.Run(map[bool]string{false: "backups", true: "in-place"}[inPlace], func(t *testing.T) {
			directory := t.TempDir()
			path := filepath.Join(directory, "image.img")
			original := []byte("original")
			replacement := []byte("replacement")
			if err := os.WriteFile(path, original, 0o600); err != nil {
				t.Fatal(err)
			}
			metadata, err := compressionCaptureFileMetadata(path)
			if err != nil {
				t.Fatal(err)
			}
			stage, err := compressionStageFile(path, metadata, func(output *os.File) error {
				_, err := output.Write(replacement)
				return err
			})
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := compressionCleanupStaged([]compressionStagedFile{stage}); err != nil {
					t.Errorf("clean up staged image: %v", err)
				}
			}()

			// The context reports cancellation as soon as the destination changes.
			// With only one destination, that is after the last rename, so a
			// successful final directory sync must still commit the replacement.
			ctx := &compressionTestCancelAfterFileChangesContext{
				path:     path,
				original: original,
			}
			if err := compressionCommit(ctx, []compressionStagedFile{stage}, inPlace); err != nil {
				t.Fatalf("compressionCommit() error = %v", err)
			}
			if !errors.Is(ctx.Err(), context.Canceled) {
				t.Fatal("test context did not observe the committed replacement")
			}

			got, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, replacement) {
				t.Fatalf("committed image = %q, want %q", got, replacement)
			}
			backup, err := os.ReadFile(path + ".bak")
			if inPlace {
				if !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("in-place backup error = %v, want file not found", err)
				}
			} else if err != nil {
				t.Fatalf("read backup: %v", err)
			} else if !bytes.Equal(backup, original) {
				t.Fatalf("backup = %q, want %q", backup, original)
			}
			compressionTestAssertNoTemporaryFiles(t, directory)
		})
	}
}

func TestCheckpointCompressionCommitRollsBackSecondRenameFailure(t *testing.T) {
	for _, inPlace := range []bool{false, true} {
		t.Run(map[bool]string{false: "backups", true: "in-place"}[inPlace], func(t *testing.T) {
			directory := t.TempDir()
			paths := []string{
				filepath.Join(directory, "one.img"),
				filepath.Join(directory, "two.img"),
			}
			originals := [][]byte{[]byte("original one"), []byte("original two")}
			staged := make([]compressionStagedFile, 0, len(paths))
			for index, path := range paths {
				if err := os.WriteFile(path, originals[index], 0o600); err != nil {
					t.Fatal(err)
				}
				metadata, err := compressionCaptureFileMetadata(path)
				if err != nil {
					t.Fatal(err)
				}
				stage, err := compressionStageFile(path, metadata, func(output *os.File) error {
					_, err := output.Write([]byte("replacement"))
					return err
				})
				if err != nil {
					t.Fatal(err)
				}
				staged = append(staged, stage)
			}
			defer func() {
				if err := compressionCleanupStaged(staged); err != nil {
					t.Errorf("clean up staged images: %v", err)
				}
			}()

			// The first replacement succeeds. Removing only the second staged
			// file makes its real os.Rename fail inside the commit window.
			if err := os.Remove(staged[1].tempPath); err != nil {
				t.Fatal(err)
			}
			err := compressionCommit(context.Background(), staged, inPlace)
			if !errors.Is(err, os.ErrNotExist) ||
				!strings.Contains(err.Error(), "commit failed; original images were restored") ||
				!strings.Contains(err.Error(), "replace image "+paths[1]) {
				t.Fatalf("compressionCommit() error = %v", err)
			}

			for index, path := range paths {
				got, readErr := os.ReadFile(path)
				if readErr != nil {
					t.Fatal(readErr)
				}
				if !bytes.Equal(got, originals[index]) {
					t.Fatalf("%s was not rolled back: %q", filepath.Base(path), got)
				}
				if _, statErr := os.Lstat(path + ".bak"); !errors.Is(statErr, os.ErrNotExist) {
					t.Fatalf("failed transaction left backup %s: %v", path+".bak", statErr)
				}
			}
			if err := compressionCleanupStaged(staged); err != nil {
				t.Fatalf("clean up staged images: %v", err)
			}
			compressionTestAssertNoTemporaryFiles(t, directory)
		})
	}
}

func TestCheckpointCompressionCommitReportsCleanupFailure(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "image.img")
	if err := os.WriteFile(path, []byte("original"), 0o600); err != nil {
		t.Fatal(err)
	}
	metadata, err := compressionCaptureFileMetadata(path)
	if err != nil {
		t.Fatal(err)
	}
	stage, err := compressionStageFile(path, metadata, func(output *os.File) error {
		_, err := output.Write([]byte("replacement"))
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := compressionCleanupStaged([]compressionStagedFile{stage}); err != nil {
			t.Errorf("clean up staged image: %v", err)
		}
	}()

	cleanupError := errors.New("injected cleanup failure")
	err = compressionCommitWithCleanup(
		context.Background(),
		[]compressionStagedFile{stage},
		true,
		func(string) error { return cleanupError },
	)
	if !errors.Is(err, ErrCheckpointCompressionCleanup) || !errors.Is(err, cleanupError) {
		t.Fatalf("compressionCommitWithCleanup() error = %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "replacement" {
		t.Fatalf("committed image = %q, want replacement", got)
	}
	matches, err := filepath.Glob(filepath.Join(directory, ".image.img.crit-rollback-*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 1 {
		t.Fatalf("rollback links = %v, want one", matches)
	}
}

type compressionTestCancelContext struct {
	mu       sync.Mutex
	calls    int
	cancelAt int
}

func (*compressionTestCancelContext) Deadline() (time.Time, bool) { return time.Time{}, false }
func (*compressionTestCancelContext) Done() <-chan struct{}       { return nil }
func (*compressionTestCancelContext) Value(any) any               { return nil }

func (ctx *compressionTestCancelContext) Err() error {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.calls++
	if ctx.calls >= ctx.cancelAt {
		return context.Canceled
	}
	return nil
}

type compressionTestCancelAfterFileChangesContext struct {
	path     string
	original []byte
}

func (*compressionTestCancelAfterFileChangesContext) Deadline() (time.Time, bool) {
	return time.Time{}, false
}
func (*compressionTestCancelAfterFileChangesContext) Done() <-chan struct{} { return nil }
func (*compressionTestCancelAfterFileChangesContext) Value(any) any         { return nil }

func (ctx *compressionTestCancelAfterFileChangesContext) Err() error {
	current, err := os.ReadFile(ctx.path)
	if err != nil || !bytes.Equal(current, ctx.original) {
		return context.Canceled
	}
	return nil
}

type compressionTestMetadata struct {
	permissions os.FileMode
	uid         uint32
	gid         uint32
	mtime       time.Time
	xattrName   string
	xattrValue  []byte
}

func compressionTestApplyMetadata(t *testing.T, path string, index int) compressionTestMetadata {
	t.Helper()
	permissions := []os.FileMode{0o600, 0o640, 0o660}[index]
	if err := os.Chmod(path, permissions); err != nil {
		t.Fatal(err)
	}
	mtime := time.Unix(1_700_001_000+int64(index), 123_456_789)
	if err := os.Chtimes(path, mtime, mtime); err != nil {
		t.Fatal(err)
	}

	xattrName := "user.go-criu-compression-test"
	xattrValue := []byte(fmt.Sprintf("image-%d", index))
	if err := unix.Lsetxattr(path, xattrName, xattrValue, 0); err != nil {
		if errors.Is(err, unix.ENOTSUP) || errors.Is(err, unix.ENOSYS) ||
			errors.Is(err, unix.EPERM) || errors.Is(err, unix.EACCES) {
			xattrName = ""
			xattrValue = nil
		} else {
			t.Fatalf("set xattr on %s: %v", path, err)
		}
	}

	info, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		t.Fatalf("unexpected stat type for %s", path)
	}
	return compressionTestMetadata{
		permissions: permissions,
		uid:         stat.Uid,
		gid:         stat.Gid,
		mtime:       info.ModTime(),
		xattrName:   xattrName,
		xattrValue:  xattrValue,
	}
}

func compressionTestCheckMetadata(t *testing.T, path string, want compressionTestMetadata) {
	t.Helper()
	info, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		t.Fatalf("unexpected stat type for %s", path)
	}
	if info.Mode().Perm() != want.permissions {
		t.Errorf("%s permissions = %v, want %v", path, info.Mode().Perm(), want.permissions)
	}
	if stat.Uid != want.uid || stat.Gid != want.gid {
		t.Errorf("%s owner = %d:%d, want %d:%d", path, stat.Uid, stat.Gid, want.uid, want.gid)
	}
	if !info.ModTime().Equal(want.mtime) {
		t.Errorf("%s mtime = %v, want %v", path, info.ModTime(), want.mtime)
	}
	if want.xattrName == "" {
		return
	}
	value := make([]byte, len(want.xattrValue))
	n, err := unix.Lgetxattr(path, want.xattrName, value)
	if err != nil {
		t.Errorf("read xattr from %s: %v", path, err)
		return
	}
	if !bytes.Equal(value[:n], want.xattrValue) {
		t.Errorf("%s xattr = %q, want %q", path, value[:n], want.xattrValue)
	}
}

func compressionTestWriteCheckpoint(t *testing.T, directory string, mode CompressionMode, entries []*pagemap.PagemapEntry, pages []byte) {
	t.Helper()
	compressionTestWriteCheckpointNamed(t, directory, "pagemap-1.img", 1, mode, entries, pages)
}

func compressionTestWriteCheckpointNamed(t *testing.T, directory, pagemapName string, pagesID uint32, mode CompressionMode, entries []*pagemap.PagemapEntry, pages []byte) {
	t.Helper()
	version := crtoolsImagesV1_1
	invEntry := &inventory.InventoryEntry{ImgVersion: proto.Uint32(version)}
	if mode.Compressed() {
		invEntry.ImgVersion = proto.Uint32(crtoolsImagesV1_2)
		invEntry.Compress = proto.Uint32(uint32(mode))
		blockSize := uint32(compressionTestPageSize)
		for _, entry := range entries {
			if blocks := entry.GetBlocks(); blocks != nil && blocks.GetPagesPerBlock() != 0 {
				blockSize = blocks.GetPagesPerBlock() * compressionTestPageSize
				break
			}
		}
		invEntry.CompressBlockSize = proto.Uint32(blockSize)
	}
	compressionTestWriteImage(t, filepath.Join(directory, "inventory.img"), &CriuImage{
		Magic: "INVENTORY", Entries: []*CriuEntry{{Message: invEntry}},
	})
	imageEntries := make([]*CriuEntry, 1, 1+len(entries))
	imageEntries[0] = &CriuEntry{Message: &pagemap.PagemapHead{PagesId: proto.Uint32(pagesID)}}
	for _, entry := range entries {
		imageEntries = append(imageEntries, &CriuEntry{Message: entry})
	}
	compressionTestWriteImage(t, filepath.Join(directory, pagemapName), &CriuImage{Magic: "PAGEMAP", Entries: imageEntries})
	if err := os.WriteFile(filepath.Join(directory, "pages-"+strconv.FormatUint(uint64(pagesID), 10)+".img"), pages, 0o600); err != nil {
		t.Fatal(err)
	}
	if taskID, ok := compressionNumericImageID(pagemapName, "pagemap-"); ok {
		compressionTestWriteMM(t, filepath.Join(directory, fmt.Sprintf("mm-%d.img", taskID)))
	}
}

func compressionTestPagemapEntry(vaddr uint64, pages uint32) *pagemap.PagemapEntry {
	return &pagemap.PagemapEntry{
		Vaddr: proto.Uint64(vaddr), CompatNrPages: proto.Uint32(pages), Flags: proto.Uint32(pePresent),
	}
}

func compressionTestWriteMM(t *testing.T, path string, exceptional ...*vma.VmaEntry) {
	t.Helper()
	zero64 := func() *uint64 { return proto.Uint64(0) }
	entry := &mm.MmEntry{
		MmStartCode: zero64(), MmEndCode: zero64(), MmStartData: zero64(), MmEndData: zero64(),
		MmStartStack: zero64(), MmStartBrk: zero64(), MmBrk: zero64(), MmArgStart: zero64(),
		MmArgEnd: zero64(), MmEnvStart: zero64(), MmEnvEnd: zero64(), ExeFileId: proto.Uint32(0),
		Vmas: exceptional,
	}
	compressionTestWriteImage(t, path, &CriuImage{Magic: "MM", Entries: []*CriuEntry{{Message: entry}}})
}

func compressionTestWriteImage(t *testing.T, path string, image *CriuImage) {
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

func compressionTestReadPagemap(t *testing.T, directory string) []*pagemap.PagemapEntry {
	t.Helper()
	image, err := getImg(filepath.Join(directory, "pagemap-1.img"), nil)
	if err != nil {
		t.Fatal(err)
	}
	entries := make([]*pagemap.PagemapEntry, 0, len(image.Entries)-1)
	for _, raw := range image.Entries[1:] {
		entries = append(entries, raw.Message.(*pagemap.PagemapEntry))
	}
	return entries
}

func compressionTestReadFiles(t *testing.T, directory string, names []string) map[string][]byte {
	t.Helper()
	files := make(map[string][]byte, len(names))
	for _, name := range names {
		data, err := os.ReadFile(filepath.Join(directory, name))
		if err != nil {
			t.Fatal(err)
		}
		files[name] = data
	}
	return files
}

func compressionTestRandomPage(t *testing.T) []byte {
	t.Helper()
	page := make([]byte, compressionTestPageSize)
	// A deterministic non-cryptographic stream keeps this compression test reproducible.
	//nolint:gosec
	if _, err := rand.New(rand.NewSource(1)).Read(page); err != nil {
		t.Fatal(err)
	}
	return page
}

func compressionTestAssertNoTemporaryFiles(t *testing.T, directory string) {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(directory, ".*.crit-*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 0 {
		t.Fatalf("staged compression files remain: %v", matches)
	}
}
