package crit

import (
	"context"
	"errors"
)

var (
	// ErrUnsupportedAcceleration reports an LZ4 acceleration value that the
	// pure-Go encoder cannot reproduce. Acceleration 0 selects the default (1).
	ErrUnsupportedAcceleration = errors.New("unsupported LZ4 acceleration")
	// ErrCheckpointCompressionCleanup reports that updated checkpoint images
	// were committed but temporary rollback links could not be removed or synced.
	ErrCheckpointCompressionCleanup = errors.New("checkpoint compression committed but cleanup failed")
)

// CompressOptions configures CompressCheckpoint.
type CompressOptions struct {
	// PageSize is the checkpoint page size. Zero uses the host page size.
	PageSize int
	// InPlace skips persistent .bak files. Handled commit failures are rolled
	// back, but the operation is not atomic across process or system crashes.
	InPlace bool
	// Acceleration is the LZ4 fast-compressor acceleration. Zero defaults to 1.
	// The pure-Go implementation currently supports only 1.
	Acceleration int
}

// DecompressOptions configures DecompressCheckpoint.
type DecompressOptions struct {
	// PageSize is the checkpoint page size. Zero uses the host page size.
	PageSize int
	// InPlace skips persistent .bak files. Handled commit failures are rolled
	// back, but the operation is not atomic across process or system crashes.
	InPlace bool
}

// PagemapCompressionStats describes one updated pagemap/pages image pair.
type PagemapCompressionStats struct {
	Pagemap     string
	Pages       uint64
	InputBytes  uint64
	OutputBytes uint64
}

// CheckpointCompressionResult describes a compression or decompression result.
type CheckpointCompressionResult struct {
	// Changed is false when the checkpoint is already in the requested state or
	// the directory has no pagemap images.
	Changed bool
	// AlreadyInRequestedState distinguishes an idempotent no-op from an empty
	// checkpoint directory.
	AlreadyInRequestedState bool
	Compression             CompressionInfo
	Pagemaps                []PagemapCompressionStats
}

// CompressCheckpoint converts local, uncompressed pages images to CRIU's
// page-sized LZ4 block representation. For an incremental chain, checkpointDir must be
// its newest layer because a single-directory operation cannot propagate
// image version V1.2 into newer descendants. By default each replaced image
// gets a .bak hard link containing the original. The checkpoint directory must
// not be modified concurrently. If the returned error matches
// ErrCheckpointCompressionCleanup, the result describes a compression that
// was committed successfully. On non-Linux systems the operation returns an
// error matching [errors.ErrUnsupported]. Reading compressed images remains
// supported on every platform on which CRIT builds.
func CompressCheckpoint(ctx context.Context, checkpointDir string, opts CompressOptions) (CheckpointCompressionResult, error) {
	return compressCheckpoint(ctx, checkpointDir, opts)
}

// DecompressCheckpoint converts local page-sized or multi-page block-compressed pages
// images to ordinary raw pages images. By default each replaced image gets a
// .bak hard link containing the original. The checkpoint directory must not
// be modified concurrently. If the returned error matches
// ErrCheckpointCompressionCleanup, the result describes a decompression that
// was committed successfully. On non-Linux systems the operation returns an
// error matching [errors.ErrUnsupported]. Reading compressed images remains
// supported on every platform on which CRIT builds.
func DecompressCheckpoint(ctx context.Context, checkpointDir string, opts DecompressOptions) (CheckpointCompressionResult, error) {
	return decompressCheckpoint(ctx, checkpointDir, opts)
}
