package crit

import (
	"fmt"
	"path/filepath"

	"github.com/checkpoint-restore/go-criu/v8/crit/images/inventory"
)

const (
	crtoolsImagesV1         uint32 = 1
	crtoolsImagesV1_1       uint32 = 2
	crtoolsImagesV1_2       uint32 = 3
	minimumCRIUPageSize     uint32 = 4096
	maxCompressedRegionSize uint64 = 4 * 1024 * 1024
)

// CompressionMode is the memory page compression mode recorded in a CRIU
// inventory image.
type CompressionMode uint32

const (
	// CompressionOff means memory pages use the ordinary raw layout.
	CompressionOff CompressionMode = iota
	// CompressionPerPage means each compressed_size value describes one page.
	CompressionPerPage
	// CompressionRegion means each compressed_size value describes a region.
	CompressionRegion
)

// String returns the stable, machine-readable name of a compression mode.
func (m CompressionMode) String() string {
	switch m {
	case CompressionOff:
		return "off"
	case CompressionPerPage:
		return "per-page"
	case CompressionRegion:
		return "region"
	default:
		return fmt.Sprintf("unknown(%d)", uint32(m))
	}
}

// Compressed reports whether the mode stores compressed memory pages.
func (m CompressionMode) Compressed() bool {
	return m == CompressionPerPage || m == CompressionRegion
}

// CompressionInfo describes the memory page compression recorded by a
// checkpoint's inventory image. It describes the local checkpoint layer;
// parent layers may use a different representation.
type CompressionInfo struct {
	ImageVersion    uint32
	Mode            CompressionMode
	RegionSizeBytes uint32
}

// Compressed reports whether the local checkpoint layer uses compression.
func (i CompressionInfo) Compressed() bool {
	return i.Mode.Compressed()
}

// ReadCompressionInfo reads and validates the compression fields in a CRIU
// checkpoint inventory.
func ReadCompressionInfo(checkpointDir string) (CompressionInfo, error) {
	path := filepath.Join(checkpointDir, "inventory.img")
	img, err := getImg(path, &inventory.InventoryEntry{})
	if err != nil {
		return CompressionInfo{}, err
	}
	if img.Magic != "INVENTORY" {
		return CompressionInfo{}, fmt.Errorf(
			"%s is a %s image, expected INVENTORY",
			path,
			img.Magic,
		)
	}
	if len(img.Entries) == 0 {
		return CompressionInfo{}, fmt.Errorf("%s has no entries", path)
	}

	entry, ok := img.Entries[0].Message.(*inventory.InventoryEntry)
	if !ok || entry == nil {
		return CompressionInfo{}, fmt.Errorf("%s has an invalid inventory entry", path)
	}
	version := entry.GetImgVersion()
	switch version {
	case crtoolsImagesV1, crtoolsImagesV1_1, crtoolsImagesV1_2:
	default:
		return CompressionInfo{}, fmt.Errorf("%s has unsupported image version %d", path, version)
	}

	mode := CompressionMode(entry.GetCompress())
	switch mode {
	case CompressionOff, CompressionPerPage, CompressionRegion:
	default:
		return CompressionInfo{}, fmt.Errorf("%s has invalid compression mode %d", path, uint32(mode))
	}
	if mode.Compressed() && version != crtoolsImagesV1_2 {
		return CompressionInfo{}, fmt.Errorf(
			"%s image version %d cannot contain compressed memory pages",
			path,
			version,
		)
	}
	if mode == CompressionRegion {
		regionSize := entry.GetCompressRegionSize()
		if entry.CompressRegionSize == nil || regionSize == 0 ||
			uint64(regionSize) > maxCompressedRegionSize ||
			regionSize%minimumCRIUPageSize != 0 {
			return CompressionInfo{}, fmt.Errorf(
				"%s has invalid compression region size %d",
				path,
				regionSize,
			)
		}
	} else if entry.CompressRegionSize != nil {
		return CompressionInfo{}, fmt.Errorf(
			"%s records a compression region size outside region mode",
			path,
		)
	}
	return CompressionInfo{
		ImageVersion:    version,
		Mode:            mode,
		RegionSizeBytes: entry.GetCompressRegionSize(),
	}, nil
}
