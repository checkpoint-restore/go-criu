//go:build linux

package crit

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/checkpoint-restore/go-criu/v8/crit/images/inventory"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/mm"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/pagemap"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/vma"
	"github.com/pierrec/lz4/v4"
	"golang.org/x/sys/unix"
	"google.golang.org/protobuf/proto"
)

// A transform discovers images, validates their layouts, stages every
// replacement, and commits the staged set. Originals remain untouched until
// all replacement files are ready.

const (
	compressionMapShared     = uint32(1 << 0)
	compressionMapHugetlb    = uint32(0x40000)
	compressionVMAFileShared = uint32(1 << 7)
	compressionVMAAnonShared = uint32(1 << 8)
	compressionVMASysVIPC    = uint32(1 << 10)
	compressionVMAExtPlugin  = uint32(1 << 27)

	compressionCopyBufferSize = 1024 * 1024
)

type compressionXattr struct {
	name  string
	value []byte
}

type compressionFileMetadata struct {
	path   string
	dev    uint64
	ino    uint64
	uid    uint32
	gid    uint32
	mode   uint32
	size   int64
	atime  time.Time
	mtime  time.Time
	xattrs []compressionXattr
}

func compressionUint64[T ~uint32 | ~uint64](value T) uint64 {
	return uint64(value)
}

type compressionPagemap struct {
	name      string
	pagesName string
	pagesID   uint32
	path      string
	pagesPath string
	image     *CriuImage
	entries   []*pagemap.PagemapEntry
}

type compressionInventory struct {
	path  string
	image *CriuImage
	entry *inventory.InventoryEntry
	info  CompressionInfo
}

// compressionRange is a half-open address interval [start, end).
type compressionRange struct {
	start uint64
	end   uint64
}

type compressionStagedFile struct {
	path     string
	tempPath string
	metadata *compressionFileMetadata
}

type compressionDecodeScratch struct {
	copyBuffer []byte
	zeroBuffer []byte
	encoded    []byte
	decoded    []byte
}

// compressCheckpoint converts raw pages to CRIU's per-page compression format.
func compressCheckpoint(ctx context.Context, checkpointDir string, opts CompressOptions) (result CheckpointCompressionResult, retErr error) {
	if ctx == nil {
		return result, errors.New("nil context")
	}
	if opts.Acceleration == 0 {
		opts.Acceleration = 1
	}
	if opts.Acceleration != 1 {
		return result, fmt.Errorf("%w: %d (only 1 is supported)", ErrUnsupportedAcceleration, opts.Acceleration)
	}

	pageSize, err := compressionPageSize(opts.PageSize)
	if err != nil {
		return result, err
	}
	checkpointDir, err = compressionPrepareDirectory(ctx, checkpointDir)
	if err != nil {
		return result, err
	}
	inventoryPath := filepath.Join(checkpointDir, "inventory.img")
	inventoryMetadata, err := compressionCaptureFileMetadata(inventoryPath)
	if err != nil {
		return result, err
	}
	metadata := map[string]*compressionFileMetadata{inventoryPath: inventoryMetadata}
	inv, err := compressionLoadInventory(checkpointDir, metadata)
	if err != nil {
		return result, err
	}
	result.Compression = inv.info
	metadata, err = compressionCaptureImageMetadata(checkpointDir)
	if err != nil {
		return result, err
	}
	metadata[inventoryPath] = inventoryMetadata

	pagemaps, err := compressionFindPagemaps(ctx, checkpointDir, metadata)
	if err != nil {
		return result, err
	}
	if len(pagemaps) == 0 {
		result.AlreadyInRequestedState = inv.info.Compressed()
		return result, nil
	}
	if err := compressionValidatePagemaps(pagemaps, metadata, pageSize, inv.info); err != nil {
		return result, err
	}
	if inv.info.Compressed() {
		result.AlreadyInRequestedState = true
		return result, nil
	}
	rawRanges, err := compressionExceptionalRanges(ctx, checkpointDir, pagemaps, metadata, pageSize)
	if err != nil {
		return result, err
	}

	staged := make([]compressionStagedFile, 0, len(pagemaps)*2+1)
	defer func() {
		if retErr != nil {
			retErr = errors.Join(retErr, compressionCleanupStaged(staged))
		}
	}()

	for _, pm := range pagemaps {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		pagesStage, stats, err := compressionStageCompressedPages(ctx, pm, metadata[pm.pagesPath], pageSize, rawRanges[pm.name])
		if err != nil {
			return result, err
		}
		staged = append(staged, pagesStage)
		pagemapStage, err := compressionStageImage(pm.path, pm.image, metadata[pm.path])
		if err != nil {
			return result, err
		}
		staged = append(staged, pagemapStage)
		result.Pagemaps = append(result.Pagemaps, stats)
	}

	inv.entry.Compress = proto.Uint32(uint32(CompressionPerPage))
	inv.entry.CompressRegionSize = nil
	inv.entry.ImgVersion = proto.Uint32(crtoolsImagesV1_2)
	invStage, err := compressionStageImage(inv.path, inv.image, metadata[inv.path])
	if err != nil {
		return result, err
	}
	staged = append(staged, invStage)

	commitErr := compressionCommit(ctx, staged, opts.InPlace)
	if commitErr == nil || errors.Is(commitErr, ErrCheckpointCompressionCleanup) {
		result.Changed = true
		result.Compression = CompressionInfo{ImageVersion: crtoolsImagesV1_2, Mode: CompressionPerPage}
	}
	return result, commitErr
}

// decompressCheckpoint expands per-page or region-compressed pages.
func decompressCheckpoint(ctx context.Context, checkpointDir string, opts DecompressOptions) (result CheckpointCompressionResult, retErr error) {
	if ctx == nil {
		return result, errors.New("nil context")
	}
	pageSize, err := compressionPageSize(opts.PageSize)
	if err != nil {
		return result, err
	}
	checkpointDir, err = compressionPrepareDirectory(ctx, checkpointDir)
	if err != nil {
		return result, err
	}
	inventoryPath := filepath.Join(checkpointDir, "inventory.img")
	inventoryMetadata, err := compressionCaptureFileMetadata(inventoryPath)
	if err != nil {
		return result, err
	}
	metadata := map[string]*compressionFileMetadata{inventoryPath: inventoryMetadata}
	inv, err := compressionLoadInventory(checkpointDir, metadata)
	if err != nil {
		return result, err
	}
	result.Compression = inv.info
	metadata, err = compressionCaptureImageMetadata(checkpointDir)
	if err != nil {
		return result, err
	}
	metadata[inventoryPath] = inventoryMetadata

	pagemaps, err := compressionFindPagemaps(ctx, checkpointDir, metadata)
	if err != nil {
		return result, err
	}
	if len(pagemaps) == 0 {
		result.AlreadyInRequestedState = !inv.info.Compressed()
		return result, nil
	}
	if err := compressionValidatePagemaps(pagemaps, metadata, pageSize, inv.info); err != nil {
		return result, err
	}
	if !inv.info.Compressed() {
		result.AlreadyInRequestedState = true
		return result, nil
	}
	hasParent := compressionHasParentReference(checkpointDir, pagemaps)

	staged := make([]compressionStagedFile, 0, len(pagemaps)*2+1)
	defer func() {
		if retErr != nil {
			retErr = errors.Join(retErr, compressionCleanupStaged(staged))
		}
	}()

	for _, pm := range pagemaps {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		pagesStage, stats, err := compressionStageDecompressedPages(ctx, pm, metadata[pm.pagesPath], pageSize)
		if err != nil {
			return result, err
		}
		staged = append(staged, pagesStage)
		pagemapStage, err := compressionStageImage(pm.path, pm.image, metadata[pm.path])
		if err != nil {
			return result, err
		}
		staged = append(staged, pagemapStage)
		result.Pagemaps = append(result.Pagemaps, stats)
	}

	inv.entry.Compress = nil
	inv.entry.CompressRegionSize = nil
	if !hasParent && inv.entry.GetImgVersion() == crtoolsImagesV1_2 {
		inv.entry.ImgVersion = proto.Uint32(crtoolsImagesV1_1)
	}
	invStage, err := compressionStageImage(inv.path, inv.image, metadata[inv.path])
	if err != nil {
		return result, err
	}
	staged = append(staged, invStage)

	commitErr := compressionCommit(ctx, staged, opts.InPlace)
	if commitErr == nil || errors.Is(commitErr, ErrCheckpointCompressionCleanup) {
		result.Changed = true
		result.Compression = CompressionInfo{ImageVersion: inv.entry.GetImgVersion(), Mode: CompressionOff}
	}
	return result, commitErr
}

// Checkpoint discovery and validation.

func compressionPageSize(configured int) (uint64, error) {
	if configured == 0 {
		configured = os.Getpagesize()
	}
	if configured <= 0 || configured&(configured-1) != 0 {
		return 0, fmt.Errorf("invalid page size %d", configured)
	}
	if uint64(configured) > math.MaxUint32 {
		return 0, fmt.Errorf("page size %d does not fit CRIU block metadata", configured)
	}
	return uint64(configured), nil
}

func compressionPrepareDirectory(ctx context.Context, directory string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	absolute, err := filepath.Abs(directory)
	if err != nil {
		return "", fmt.Errorf("resolve checkpoint directory: %w", err)
	}
	info, err := os.Stat(absolute)
	if err != nil {
		return "", fmt.Errorf("stat checkpoint directory: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("checkpoint path is not a directory: %s", absolute)
	}
	return absolute, nil
}

func compressionCaptureImageMetadata(directory string) (map[string]*compressionFileMetadata, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("read checkpoint directory: %w", err)
	}
	metadata := make(map[string]*compressionFileMetadata)
	for _, entry := range entries {
		if !compressionIsImageMetadataFile(entry.Name()) {
			continue
		}
		path := filepath.Join(directory, entry.Name())
		captured, err := compressionCaptureFileMetadata(path)
		if err != nil {
			return nil, err
		}
		metadata[path] = captured
	}
	return metadata, nil
}

func compressionIsImageMetadataFile(name string) bool {
	if !strings.HasSuffix(name, ".img") {
		return false
	}
	return name == "inventory.img" || strings.HasPrefix(name, "mm-") ||
		strings.HasPrefix(name, "pagemap-") || strings.HasPrefix(name, "pages-")
}

// compressionCaptureFileMetadata records identity and attributes for later
// verification and replacement.
func compressionCaptureFileMetadata(path string) (*compressionFileMetadata, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, fmt.Errorf("stat image %s: %w", path, err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("image is not a regular file: %s", path)
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return nil, fmt.Errorf("unsupported file metadata for %s", path)
	}
	xattrs, err := compressionReadXattrs(path)
	if err != nil {
		return nil, fmt.Errorf("read extended attributes from %s: %w", path, err)
	}
	return &compressionFileMetadata{
		path: path, dev: compressionUint64(stat.Dev), ino: stat.Ino,
		uid: stat.Uid, gid: stat.Gid, mode: stat.Mode, size: stat.Size,
		atime: time.Unix(stat.Atim.Unix()),
		mtime: time.Unix(stat.Mtim.Unix()), xattrs: xattrs,
	}, nil
}

func compressionReadXattrs(path string) ([]compressionXattr, error) {
	size, err := unix.Llistxattr(path, nil)
	if errors.Is(err, unix.ENOTSUP) || errors.Is(err, unix.ENOSYS) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if size == 0 {
		return nil, nil
	}
	buf := make([]byte, size)
	n, err := unix.Llistxattr(path, buf)
	if err != nil {
		return nil, err
	}
	var attrs []compressionXattr
	for _, rawName := range bytes.Split(buf[:n], []byte{0}) {
		if len(rawName) == 0 {
			continue
		}
		name := string(rawName)
		valueSize, err := unix.Lgetxattr(path, name, nil)
		if err != nil {
			return nil, err
		}
		value := make([]byte, valueSize)
		if valueSize != 0 {
			n, err = unix.Lgetxattr(path, name, value)
			if err != nil {
				return nil, err
			}
			value = value[:n]
		}
		attrs = append(attrs, compressionXattr{name: name, value: value})
	}
	return attrs, nil
}

// compressionLoadInventory validates the image version and compression mode.
func compressionLoadInventory(directory string, metadata map[string]*compressionFileMetadata) (*compressionInventory, error) {
	path := filepath.Join(directory, "inventory.img")
	meta, ok := metadata[path]
	if !ok {
		return nil, fmt.Errorf("missing inventory image: %s", path)
	}
	image, err := compressionLoadImage(path, &inventory.InventoryEntry{}, meta)
	if err != nil {
		return nil, err
	}
	if image.Magic != "INVENTORY" {
		return nil, fmt.Errorf("%s is a %s image, expected INVENTORY", path, image.Magic)
	}
	if len(image.Entries) == 0 {
		return nil, errors.New("inventory has no entries")
	}
	entry, ok := image.Entries[0].Message.(*inventory.InventoryEntry)
	if !ok {
		return nil, errors.New("inventory has an invalid first entry")
	}
	version := entry.GetImgVersion()
	if version != crtoolsImagesV1_1 && version != crtoolsImagesV1_2 {
		return nil, fmt.Errorf("inventory has unsupported image version %d", version)
	}
	mode := CompressionMode(entry.GetCompress())
	if mode != CompressionOff && mode != CompressionPerPage && mode != CompressionRegion {
		return nil, fmt.Errorf("inventory has invalid compression mode %d", mode)
	}
	if mode.Compressed() && version != crtoolsImagesV1_2 {
		return nil, fmt.Errorf("inventory image version %d cannot contain compressed pages", version)
	}
	if mode == CompressionRegion {
		regionSize := entry.GetCompressRegionSize()
		if entry.CompressRegionSize == nil || regionSize == 0 ||
			uint64(regionSize) > maxCompressedRegionSize ||
			regionSize%minimumCRIUPageSize != 0 {
			return nil, fmt.Errorf("inventory has invalid compression region size %d", regionSize)
		}
	} else if entry.CompressRegionSize != nil {
		return nil, errors.New("inventory records a compression region size outside region mode")
	}
	info := CompressionInfo{ImageVersion: version, Mode: mode, RegionSizeBytes: entry.GetCompressRegionSize()}
	return &compressionInventory{path: path, image: image, entry: entry, info: info}, nil
}

func compressionLoadImage(path string, entryType proto.Message, metadata *compressionFileMetadata) (image *CriuImage, retErr error) {
	file, restoreTimes, err := compressionOpenSource(path, metadata)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := compressionCloseSource(file, metadata, restoreTimes); err != nil {
			retErr = errors.Join(retErr, err)
		}
	}()
	image, retErr = decodeImg(file, entryType, false)
	if retErr != nil {
		retErr = fmt.Errorf("decode %s: %w", filepath.Base(path), retErr)
	}
	return image, retErr
}

// compressionOpenSource verifies the opened inode and avoids atime changes
// when the filesystem permits it.
func compressionOpenSource(path string, metadata *compressionFileMetadata) (*os.File, bool, error) {
	flags := unix.O_RDONLY | unix.O_CLOEXEC | unix.O_NOFOLLOW | unix.O_NOATIME
	fd, err := unix.Open(path, flags, 0)
	restoreTimes := false
	if errors.Is(err, unix.EPERM) || errors.Is(err, unix.EINVAL) || errors.Is(err, unix.EOPNOTSUPP) {
		fd, err = unix.Open(path, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
		restoreTimes = true
	}
	if err != nil {
		return nil, false, fmt.Errorf("open image %s: %w", path, err)
	}
	file := os.NewFile(uintptr(fd), path)
	if err := compressionVerifyOpenFile(file, metadata); err != nil {
		_ = file.Close()
		return nil, false, err
	}
	return file, restoreTimes, nil
}

func compressionVerifyOpenFile(file *os.File, metadata *compressionFileMetadata) error {
	var stat unix.Stat_t
	if err := unix.Fstat(int(file.Fd()), &stat); err != nil {
		return fmt.Errorf("stat opened image %s: %w", file.Name(), err)
	}
	if compressionUint64(stat.Dev) != metadata.dev || stat.Ino != metadata.ino {
		return fmt.Errorf("image changed while preparing to read: %s", file.Name())
	}
	return nil
}

func compressionCloseSource(file *os.File, metadata *compressionFileMetadata, restoreTimes bool) error {
	var errs []error
	if restoreTimes {
		times := []unix.Timespec{
			unix.NsecToTimespec(metadata.atime.UnixNano()),
			unix.NsecToTimespec(metadata.mtime.UnixNano()),
		}
		if err := unix.UtimesNanoAt(int(file.Fd()), "", times, unix.AT_EMPTY_PATH); err != nil {
			errs = append(errs, fmt.Errorf("restore image timestamps for %s: %w", file.Name(), err))
		}
	}
	if err := file.Close(); err != nil {
		errs = append(errs, fmt.Errorf("close image %s: %w", file.Name(), err))
	}
	return errors.Join(errs...)
}

func compressionNumericImageID(name, prefix string) (uint64, bool) {
	if !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, ".img") {
		return 0, false
	}
	raw := strings.TrimSuffix(strings.TrimPrefix(name, prefix), ".img")
	if raw == "" {
		return 0, false
	}
	for _, r := range raw {
		if r < '0' || r > '9' {
			return 0, false
		}
	}
	id, err := strconv.ParseUint(raw, 10, 64)
	return id, err == nil
}

func compressionIsPagemapName(name string) bool {
	if _, ok := compressionNumericImageID(name, "pagemap-shmem-"); ok {
		return true
	}
	_, ok := compressionNumericImageID(name, "pagemap-")
	return ok
}

// compressionFindPagemaps pairs each pagemap with a unique pages image.
func compressionFindPagemaps(ctx context.Context, directory string, metadata map[string]*compressionFileMetadata) ([]*compressionPagemap, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("read checkpoint directory: %w", err)
	}
	seenPages := make(map[uint32]string)
	var result []*compressionPagemap
	for _, dirEntry := range entries {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		name := dirEntry.Name()
		if !compressionIsPagemapName(name) {
			continue
		}
		path := filepath.Join(directory, name)
		meta, ok := metadata[path]
		if !ok {
			return nil, fmt.Errorf("missing metadata for %s", path)
		}
		image, err := compressionLoadImage(path, nil, meta)
		if err != nil {
			return nil, err
		}
		if image.Magic != "PAGEMAP" {
			return nil, fmt.Errorf("%s is a %s image, expected PAGEMAP", path, image.Magic)
		}
		if len(image.Entries) == 0 {
			return nil, fmt.Errorf("%s has no pagemap head", name)
		}
		head, ok := image.Entries[0].Message.(*pagemap.PagemapHead)
		if !ok || head.PagesId == nil {
			return nil, fmt.Errorf("%s has no pages_id", name)
		}
		pagesID := head.GetPagesId()
		if previous, exists := seenPages[pagesID]; exists {
			return nil, fmt.Errorf("%s reuses pages_id %d from %s", name, pagesID, previous)
		}
		seenPages[pagesID] = name
		pagesName := fmt.Sprintf("pages-%d.img", pagesID)
		pagesPath := filepath.Join(directory, pagesName)
		if _, ok := metadata[pagesPath]; !ok {
			return nil, fmt.Errorf("%s refers to missing %s", name, pagesName)
		}
		pm := &compressionPagemap{
			name: name, pagesName: pagesName, pagesID: pagesID,
			path: path, pagesPath: pagesPath, image: image,
		}
		for i, raw := range image.Entries[1:] {
			entry, ok := raw.Message.(*pagemap.PagemapEntry)
			if !ok {
				return nil, fmt.Errorf("%s entry %d has invalid type", name, i+1)
			}
			pm.entries = append(pm.entries, entry)
		}
		result = append(result, pm)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].name < result[j].name })
	return result, nil
}

// compressionValidatePagemaps checks metadata and exact payload lengths.
func compressionValidatePagemaps(pagemaps []*compressionPagemap, metadata map[string]*compressionFileMetadata, pageSize uint64, compression CompressionInfo) error {
	for _, pm := range pagemaps {
		layer, err := indexMemoryLayer(
			filepath.Dir(pm.path),
			pm.name,
			pm.pagesID,
			int(pageSize),
			pm.entries,
		)
		if err != nil {
			return err
		}
		if err := validateMemoryLayerCompression(layer, compression); err != nil {
			return err
		}
		meta := metadata[pm.pagesPath]
		if meta == nil {
			return fmt.Errorf("missing metadata for %s", pm.pagesPath)
		}
		if meta.size < 0 || layer.payloadSize != uint64(meta.size) {
			return fmt.Errorf("%s describes %d payload bytes, but %s contains %d", pm.name, layer.payloadSize, pm.pagesName, meta.size)
		}
	}
	return nil
}

// compressionExceptionalRanges returns the payload ranges that must remain raw.
// CRIU cannot restore compressed hugetlb or external-plugin pages through its
// generic memory restore path.
func compressionExceptionalRanges(ctx context.Context, directory string, pagemaps []*compressionPagemap, metadata map[string]*compressionFileMetadata, pageSize uint64) (map[string][]compressionRange, error) {
	ranges := make(map[string][]compressionRange, len(pagemaps))
	requiredMM := make(map[string]string)
	hasSharedPagemap := false
	for _, pm := range pagemaps {
		ranges[pm.name] = nil
		if taskID, ok := compressionNumericImageID(pm.name, "pagemap-"); ok {
			mmName := fmt.Sprintf("mm-%d.img", taskID)
			requiredMM[mmName] = pm.name
		} else {
			hasSharedPagemap = true
		}
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("read checkpoint directory: %w", err)
	}
	mmNames := make(map[string]struct{})
	for _, dirEntry := range entries {
		if _, ok := compressionNumericImageID(dirEntry.Name(), "mm-"); ok {
			mmNames[dirEntry.Name()] = struct{}{}
		}
	}
	for mmName, pagemapName := range requiredMM {
		if _, ok := mmNames[mmName]; !ok {
			return nil, fmt.Errorf("%s requires missing %s to identify hugetlb and external-plugin ranges", pagemapName, mmName)
		}
	}
	if hasSharedPagemap && len(mmNames) == 0 {
		return nil, errors.New("shared-memory pagemap requires mm images to identify exceptional mappings")
	}

	for _, dirEntry := range entries {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		taskID, ok := compressionNumericImageID(dirEntry.Name(), "mm-")
		if !ok {
			continue
		}
		path := filepath.Join(directory, dirEntry.Name())
		meta := metadata[path]
		if meta == nil {
			return nil, fmt.Errorf("missing metadata for %s", path)
		}
		image, err := compressionLoadImage(path, &mm.MmEntry{}, meta)
		if err != nil {
			return nil, err
		}
		if image.Magic != "MM" {
			return nil, fmt.Errorf("%s is a %s image, expected MM", path, image.Magic)
		}
		if len(image.Entries) == 0 {
			return nil, fmt.Errorf("%s has no entries", dirEntry.Name())
		}
		mmEntry, ok := image.Entries[0].Message.(*mm.MmEntry)
		if !ok {
			return nil, fmt.Errorf("%s has an invalid first entry", dirEntry.Name())
		}
		taskPagemap := fmt.Sprintf("pagemap-%d.img", taskID)
		for _, vmaEntry := range mmEntry.Vmas {
			target, start, end, exceptional, err := compressionExceptionalVMATarget(dirEntry.Name(), taskPagemap, vmaEntry, pageSize)
			if err != nil {
				return nil, err
			}
			if !exceptional {
				continue
			}
			if _, exists := ranges[target]; exists {
				ranges[target] = append(ranges[target], compressionRange{start: start, end: end})
			}
		}
	}
	for name, item := range ranges {
		ranges[name] = compressionMergeRanges(item)
	}
	return ranges, nil
}

// compressionExceptionalVMATarget maps a VMA to its pagemap address space.
func compressionExceptionalVMATarget(mmName, taskPagemap string, entry *vma.VmaEntry, pageSize uint64) (string, uint64, uint64, bool, error) {
	flags, status := entry.GetFlags(), entry.GetStatus()
	if flags&compressionMapHugetlb == 0 && status&compressionVMAExtPlugin == 0 {
		return "", 0, 0, false, nil
	}
	start, end := entry.GetStart(), entry.GetEnd()
	if entry.Start == nil || entry.End == nil || end <= start || start%pageSize != 0 || end%pageSize != 0 {
		return "", 0, 0, false, fmt.Errorf("%s has an invalid exceptional VMA range %#x-%#x", mmName, start, end)
	}
	sharedStatus := compressionVMAFileShared | compressionVMAAnonShared | compressionVMASysVIPC
	if flags&compressionMapShared == 0 && status&sharedStatus == 0 {
		// Private mappings use task virtual addresses in their task pagemap.
		return taskPagemap, start, end, true, nil
	}
	if entry.Shmid == nil || entry.Pgoff == nil || entry.GetPgoff()%pageSize != 0 || entry.GetPgoff() > math.MaxUint64-(end-start) {
		return "", 0, 0, false, fmt.Errorf("%s has invalid shared exceptional VMA metadata", mmName)
	}
	// Shared mappings use offsets in the corresponding shmem pagemap.
	return fmt.Sprintf("pagemap-shmem-%d.img", entry.GetShmid()), entry.GetPgoff(), entry.GetPgoff() + end - start, true, nil
}

// compressionMergeRanges sorts and merges half-open ranges so a membership
// lookup only needs to inspect the range with the greatest start <= address.
func compressionMergeRanges(ranges []compressionRange) []compressionRange {
	sort.Slice(ranges, func(i, j int) bool {
		if ranges[i].start == ranges[j].start {
			return ranges[i].end < ranges[j].end
		}
		return ranges[i].start < ranges[j].start
	})
	merged := make([]compressionRange, 0, len(ranges))
	for _, item := range ranges {
		if len(merged) != 0 && item.start <= merged[len(merged)-1].end {
			merged[len(merged)-1].end = max(merged[len(merged)-1].end, item.end)
			continue
		}
		merged = append(merged, item)
	}
	return merged
}

// compressionMergedRangesContain assumes ranges were normalized by
// compressionMergeRanges.
func compressionMergedRangesContain(address uint64, ranges []compressionRange) bool {
	firstRangeAfterAddress := sort.Search(len(ranges), func(i int) bool {
		return ranges[i].start > address
	})
	if firstRangeAfterAddress == 0 {
		return false
	}
	return address < ranges[firstRangeAfterAddress-1].end
}

// compressionAddressInRanges also handles sign-extended 32-bit x86 addresses.
func compressionAddressInRanges(address uint64, ranges []compressionRange) bool {
	if compressionMergedRangesContain(address, ranges) {
		return true
	}

	// Native 32-bit x86 encode_pointer() sign-extends pagemap addresses through
	// long, while mm.img stores VMA bounds as unsigned 32-bit values. Normalize
	// that sign-extended representation before comparing the two image formats.
	if address>>32 != math.MaxUint32 {
		return false
	}
	compatAddress := uint64(uint32(address))
	return compressionMergedRangesContain(compatAddress, ranges)
}

// compressionHasParentReference prevents downgrading an incremental chain to
// image version V1.1 while a parent may still contain compressed pages.
func compressionHasParentReference(directory string, pagemaps []*compressionPagemap) bool {
	if _, err := os.Lstat(filepath.Join(directory, "parent")); err == nil {
		return true
	}
	for _, pm := range pagemaps {
		for _, entry := range pm.entries {
			if effectivePagemapFlags(entry)&peParent != 0 {
				return true
			}
		}
	}
	return false
}

// Page payload transformation.

// compressionStageCompressedPages writes zero, raw, or LZ4 blocks and updates
// their pagemap metadata.
func compressionStageCompressedPages(ctx context.Context, pm *compressionPagemap, metadata *compressionFileMetadata, pageSize uint64, rawRanges []compressionRange) (compressionStagedFile, PagemapCompressionStats, error) {
	stats := PagemapCompressionStats{Pagemap: pm.name}
	stage, err := compressionStageFromSource(pm.pagesPath, metadata, func(source, output *os.File) error {
		reader := bufio.NewReaderSize(source, compressionCopyBufferSize)
		writer := bufio.NewWriterSize(output, compressionCopyBufferSize)
		zeroPage := make([]byte, int(pageSize))
		page := make([]byte, int(pageSize))
		compressed := make([]byte, lz4.CompressBlockBound(int(pageSize)))
		var compressor lz4.Compressor
		var outputOffset uint64
		for _, entry := range pm.entries {
			if effectivePagemapFlags(entry)&pePresent == 0 {
				continue
			}
			nrPages := pagemapPageCount(entry)
			compressedSizes := make([]uint32, 0, int(min(nrPages, uint64(1024*1024))))
			var totalCompressed uint64
			payloadStarted := false
			for pageIndex := uint64(0); pageIndex < nrPages; pageIndex++ {
				if err := ctx.Err(); err != nil {
					return err
				}
				if _, err := io.ReadFull(reader, page); err != nil {
					return fmt.Errorf("short read in %s: %w", pm.pagesName, err)
				}
				stats.Pages++
				stats.InputBytes += pageSize
				if bytes.Equal(page, zeroPage) {
					compressedSizes = append(compressedSizes, 0)
					continue
				}
				forceRaw := compressionAddressInRanges(entry.GetVaddr()+pageIndex*pageSize, rawRanges)
				payload := page
				storedSize := int(pageSize)
				if !forceRaw {
					n, err := compressor.CompressBlock(page, compressed)
					if err != nil {
						return fmt.Errorf("compress page in %s: %w", pm.pagesName, err)
					}
					if n != 0 && uint64(n) < pageSize*7/8 {
						payload = compressed[:n]
						storedSize = n
					}
				}
				if uint64(storedSize) == pageSize && !payloadStarted {
					aligned, ok := checkedAlignUp(outputOffset, pageSize)
					if !ok {
						return fmt.Errorf("output offset overflows in %s", pm.pagesName)
					}
					if err := compressionWriteZeros(ctx, writer, aligned-outputOffset, zeroPage); err != nil {
						return err
					}
					stats.OutputBytes += aligned - outputOffset
					outputOffset = aligned
					flags := effectivePagemapFlags(entry) | pePayloadAligned
					entry.Flags = proto.Uint32(flags)
				}
				payloadStarted = true
				if _, err := writer.Write(payload); err != nil {
					return fmt.Errorf("write %s: %w", pm.pagesName, err)
				}
				compressedSizes = append(compressedSizes, uint32(storedSize))
				totalCompressed += uint64(storedSize)
				outputOffset += uint64(storedSize)
				stats.OutputBytes += uint64(storedSize)
			}
			allRaw := len(compressedSizes) != 0
			for _, size := range compressedSizes {
				if uint64(size) != pageSize {
					allRaw = false
					break
				}
			}
			if allRaw {
				compressionRemoveCompressionMetadata(entry)
			} else {
				entry.CompressedSize = compressedSizes
				entry.TotalCompressedSize = proto.Uint64(totalCompressed)
				entry.RegionPages = nil
			}
		}
		extra := make([]byte, 1)
		n, err := reader.Read(extra)
		if n != 0 || (err != nil && !errors.Is(err, io.EOF)) {
			return fmt.Errorf("unexpected trailing data in %s", pm.pagesName)
		}
		if err := writer.Flush(); err != nil {
			return fmt.Errorf("flush %s: %w", pm.pagesName, err)
		}
		return nil
	})
	return stage, stats, err
}

// compressionStageDecompressedPages expands all block encodings and removes
// their compression metadata.
func compressionStageDecompressedPages(ctx context.Context, pm *compressionPagemap, metadata *compressionFileMetadata, pageSize uint64) (compressionStagedFile, PagemapCompressionStats, error) {
	stats := PagemapCompressionStats{Pagemap: pm.name}
	stage, err := compressionStageFromSource(pm.pagesPath, metadata, func(source, output *os.File) error {
		reader := bufio.NewReaderSize(source, compressionCopyBufferSize)
		writer := bufio.NewWriterSize(output, compressionCopyBufferSize)
		scratch := compressionDecodeScratch{
			copyBuffer: make([]byte, compressionCopyBufferSize),
			zeroBuffer: make([]byte, compressionCopyBufferSize),
		}
		var inputOffset uint64
		for _, entry := range pm.entries {
			if effectivePagemapFlags(entry)&pePresent == 0 {
				continue
			}
			nrPages := pagemapPageCount(entry)
			flags := effectivePagemapFlags(entry)
			if flags&pePayloadAligned != 0 {
				aligned, ok := checkedAlignUp(inputOffset, pageSize)
				if !ok {
					return fmt.Errorf("input offset overflows in %s", pm.pagesName)
				}
				if err := compressionDiscardExact(ctx, reader, aligned-inputOffset, pm.pagesName, scratch.copyBuffer); err != nil {
					return err
				}
				stats.InputBytes += aligned - inputOffset
				inputOffset = aligned
				entry.Flags = proto.Uint32(flags &^ pePayloadAligned)
			}
			if len(entry.CompressedSize) == 0 {
				bytes := nrPages * pageSize
				if err := compressionCopyExact(ctx, reader, writer, bytes, pm.pagesName, scratch.copyBuffer); err != nil {
					return err
				}
				inputOffset += bytes
				stats.Pages += nrPages
				stats.InputBytes += bytes
				stats.OutputBytes += bytes
				continue
			}
			regionPages := uint64(entry.GetRegionPages())
			if regionPages == 0 {
				regionPages = 1
			}
			remainingPages := nrPages
			for blockIndex := 0; blockIndex < len(entry.CompressedSize); {
				if err := ctx.Err(); err != nil {
					return err
				}
				encodedSize := uint64(entry.CompressedSize[blockIndex])
				blockPages := min(regionPages, remainingPages)
				blockBytes := blockPages * pageSize

				if encodedSize == 0 || encodedSize == blockBytes {
					rawRun := encodedSize != 0
					var runPages, runInputBytes, runOutputBytes uint64
					for blockIndex < len(entry.CompressedSize) {
						blockPages = min(regionPages, remainingPages)
						blockBytes = blockPages * pageSize
						encodedSize = uint64(entry.CompressedSize[blockIndex])
						if rawRun {
							if encodedSize != blockBytes {
								break
							}
						} else if encodedSize != 0 {
							break
						}

						var ok bool
						runPages, ok = checkedAdd(runPages, blockPages)
						if !ok {
							return fmt.Errorf("page count overflows in %s", pm.pagesName)
						}
						runInputBytes, ok = checkedAdd(runInputBytes, encodedSize)
						if !ok {
							return fmt.Errorf("input size overflows in %s", pm.pagesName)
						}
						runOutputBytes, ok = checkedAdd(runOutputBytes, blockBytes)
						if !ok {
							return fmt.Errorf("output size overflows in %s", pm.pagesName)
						}
						remainingPages -= blockPages
						blockIndex++
					}

					if rawRun {
						if err := compressionCopyExact(ctx, reader, writer, runInputBytes, pm.pagesName, scratch.copyBuffer); err != nil {
							return err
						}
					} else if err := compressionWriteZeros(ctx, writer, runOutputBytes, scratch.zeroBuffer); err != nil {
						return err
					}
					inputOffset += runInputBytes
					stats.Pages += runPages
					stats.InputBytes += runInputBytes
					stats.OutputBytes += runOutputBytes
					continue
				}

				if err := compressionDecodeBlock(ctx, reader, writer, encodedSize, blockBytes, pm.pagesName, &scratch); err != nil {
					return err
				}
				inputOffset += encodedSize
				stats.Pages += blockPages
				stats.InputBytes += encodedSize
				stats.OutputBytes += blockBytes
				remainingPages -= blockPages
				blockIndex++
			}
			if remainingPages != 0 {
				return fmt.Errorf("block page count mismatch in %s (%d pages unaccounted)", pm.pagesName, remainingPages)
			}
			compressionRemoveCompressionMetadata(entry)
		}
		extra := make([]byte, 1)
		n, err := reader.Read(extra)
		if n != 0 || (err != nil && !errors.Is(err, io.EOF)) {
			return fmt.Errorf("unexpected trailing data in %s", pm.pagesName)
		}
		if err := writer.Flush(); err != nil {
			return fmt.Errorf("flush %s: %w", pm.pagesName, err)
		}
		return nil
	})
	return stage, stats, err
}

func compressionRemoveCompressionMetadata(entry *pagemap.PagemapEntry) {
	entry.CompressedSize = nil
	entry.TotalCompressedSize = nil
	entry.RegionPages = nil
}

// compressionDecodeBlock expands a zero, raw, or LZ4 block.
func compressionDecodeBlock(ctx context.Context, source io.Reader, output io.Writer, encodedBytes, decodedBytes uint64, name string, scratch *compressionDecodeScratch) error {
	if encodedBytes == 0 {
		return compressionWriteZeros(ctx, output, decodedBytes, scratch.zeroBuffer)
	}
	if encodedBytes == decodedBytes {
		return compressionCopyExact(ctx, source, output, encodedBytes, name, scratch.copyBuffer)
	}
	if encodedBytes > uint64(math.MaxInt) || decodedBytes > uint64(math.MaxInt) {
		return fmt.Errorf("block in %s is too large", name)
	}
	if cap(scratch.encoded) < int(encodedBytes) {
		scratch.encoded = make([]byte, int(encodedBytes))
	}
	encoded := scratch.encoded[:int(encodedBytes)]
	if _, err := io.ReadFull(source, encoded); err != nil {
		return fmt.Errorf("short read in %s: %w", name, err)
	}
	if cap(scratch.decoded) < int(decodedBytes) {
		scratch.decoded = make([]byte, int(decodedBytes))
	}
	decoded := scratch.decoded[:int(decodedBytes)]
	n, err := lz4.UncompressBlock(encoded, decoded)
	if err != nil {
		return fmt.Errorf("decompression failed in %s: %w", name, err)
	}
	if n != len(decoded) {
		return fmt.Errorf("decompression in %s produced %d bytes, expected %d", name, n, len(decoded))
	}
	if _, err := output.Write(decoded); err != nil {
		return fmt.Errorf("write %s: %w", name, err)
	}
	return nil
}

func compressionDiscardExact(ctx context.Context, source io.Reader, size uint64, name string, buffer []byte) error {
	if len(buffer) == 0 {
		return errors.New("empty compression discard buffer")
	}
	remaining := size
	for remaining != 0 {
		if err := ctx.Err(); err != nil {
			return err
		}
		chunk := min(remaining, uint64(len(buffer)))
		if _, err := io.ReadFull(source, buffer[:chunk]); err != nil {
			return fmt.Errorf("short read in %s: %w", name, err)
		}
		remaining -= chunk
	}
	return nil
}

func compressionCopyExact(ctx context.Context, source io.Reader, output io.Writer, size uint64, name string, buffer []byte) error {
	if len(buffer) == 0 {
		return errors.New("empty compression copy buffer")
	}
	remaining := size
	for remaining != 0 {
		if err := ctx.Err(); err != nil {
			return err
		}
		chunk := min(remaining, uint64(len(buffer)))
		if _, err := io.ReadFull(source, buffer[:chunk]); err != nil {
			return fmt.Errorf("short read in %s: %w", name, err)
		}
		if _, err := output.Write(buffer[:chunk]); err != nil {
			return fmt.Errorf("write %s: %w", name, err)
		}
		remaining -= chunk
	}
	return nil
}

func compressionWriteZeros(ctx context.Context, output io.Writer, size uint64, zeroBuffer []byte) error {
	if len(zeroBuffer) == 0 {
		return errors.New("empty compression zero buffer")
	}
	remaining := size
	for remaining != 0 {
		if err := ctx.Err(); err != nil {
			return err
		}
		chunk := min(remaining, uint64(len(zeroBuffer)))
		if _, err := output.Write(zeroBuffer[:chunk]); err != nil {
			return fmt.Errorf("write zero payload: %w", err)
		}
		remaining -= chunk
	}
	return nil
}

// Transactional staging and commit.

// compressionStageFromSource reads a verified original while staging its
// replacement.
func compressionStageFromSource(path string, metadata *compressionFileMetadata, write func(*os.File, *os.File) error) (stage compressionStagedFile, retErr error) {
	source, restoreTimes, err := compressionOpenSource(path, metadata)
	if err != nil {
		return stage, err
	}
	defer func() {
		if closeErr := compressionCloseSource(source, metadata, restoreTimes); closeErr != nil {
			retErr = errors.Join(retErr, closeErr)
			if stage.tempPath != "" {
				if removeErr := os.Remove(stage.tempPath); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
					retErr = errors.Join(retErr, fmt.Errorf("remove staged image %s: %w", stage.tempPath, removeErr))
				}
			}
		}
	}()
	stage, retErr = compressionStageFile(path, metadata, func(output *os.File) error { return write(source, output) })
	return stage, retErr
}

func compressionStageImage(path string, image *CriuImage, metadata *compressionFileMetadata) (compressionStagedFile, error) {
	return compressionStageFile(path, metadata, func(output *os.File) error {
		if err := encodeImg(image, output); err != nil {
			return fmt.Errorf("encode %s: %w", filepath.Base(path), err)
		}
		return nil
	})
}

// compressionStageFile writes and syncs a same-directory temporary file while
// preserving the original's metadata.
func compressionStageFile(path string, metadata *compressionFileMetadata, write func(*os.File) error) (stage compressionStagedFile, retErr error) {
	if metadata == nil {
		return stage, fmt.Errorf("missing metadata for %s", path)
	}
	output, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".crit-")
	if err != nil {
		return stage, fmt.Errorf("create staged image for %s: %w", path, err)
	}
	stage = compressionStagedFile{path: path, tempPath: output.Name(), metadata: metadata}
	keep := false
	defer func() {
		if output != nil {
			if err := output.Close(); err != nil {
				retErr = errors.Join(retErr, fmt.Errorf("close staged image %s: %w", output.Name(), err))
			}
		}
		if !keep {
			if err := os.Remove(stage.tempPath); err != nil && !errors.Is(err, os.ErrNotExist) {
				retErr = errors.Join(retErr, fmt.Errorf("remove staged image %s: %w", stage.tempPath, err))
			}
		}
	}()
	if err := write(output); err != nil {
		return stage, err
	}
	if err := unix.Fchown(int(output.Fd()), int(metadata.uid), int(metadata.gid)); err != nil {
		return stage, fmt.Errorf("restore owner on staged image %s: %w", path, err)
	}
	if err := unix.Fchmod(int(output.Fd()), metadata.mode&0o7777); err != nil {
		return stage, fmt.Errorf("restore mode on staged image %s: %w", path, err)
	}
	for _, attr := range metadata.xattrs {
		if err := unix.Fsetxattr(int(output.Fd()), attr.name, attr.value, 0); err != nil {
			return stage, fmt.Errorf("restore extended attribute %s on staged image %s: %w", attr.name, path, err)
		}
	}
	if err := os.Chtimes(stage.tempPath, metadata.atime, metadata.mtime); err != nil {
		return stage, fmt.Errorf("restore timestamps on staged image %s: %w", path, err)
	}
	if err := output.Sync(); err != nil {
		return stage, fmt.Errorf("sync staged image %s: %w", path, err)
	}
	if err := output.Close(); err != nil {
		output = nil
		return stage, fmt.Errorf("close staged image %s: %w", path, err)
	}
	output = nil
	keep = true
	return stage, nil
}

func compressionCommit(ctx context.Context, staged []compressionStagedFile, inPlace bool) error {
	return compressionCommitWithCleanup(ctx, staged, inPlace, os.Remove)
}

// compressionCommitWithCleanup installs the staged images as one recoverable
// set. Cancellation is checked before each rename; replacement or final-sync
// failures trigger rollback. A successful final sync is the commit point.
func compressionCommitWithCleanup(ctx context.Context, staged []compressionStagedFile, inPlace bool, removeCommittedRecovery func(string) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// Prepare recovery names before changing any destination and persist them so
	// every original remains reachable throughout the replacement phase.
	recovery := make(map[string]string, len(staged))
	created := make([]string, 0, len(staged))
	removeCreatedRecovery := func() error {
		var cleanupErrs []error
		for _, path := range created {
			if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
				cleanupErrs = append(cleanupErrs, fmt.Errorf("remove recovery link %s: %w", path, err))
			}
		}
		return errors.Join(cleanupErrs...)
	}
	for _, item := range staged {
		if err := compressionVerifyPath(item.path, item.metadata); err != nil {
			return errors.Join(err, removeCreatedRecovery())
		}
		var recoveryPath string
		if inPlace {
			file, err := os.CreateTemp(filepath.Dir(item.path), "."+filepath.Base(item.path)+".crit-rollback-")
			if err != nil {
				return errors.Join(fmt.Errorf("create rollback name for %s: %w", item.path, err), removeCreatedRecovery())
			}
			recoveryPath = file.Name()
			if err := file.Close(); err != nil {
				removeErr := os.Remove(recoveryPath)
				if errors.Is(removeErr, os.ErrNotExist) {
					removeErr = nil
				} else if removeErr != nil {
					removeErr = fmt.Errorf("remove rollback placeholder %s: %w", recoveryPath, removeErr)
				}
				return errors.Join(
					fmt.Errorf("close rollback placeholder for %s: %w", item.path, err),
					removeErr,
					removeCreatedRecovery(),
				)
			}
			if err := os.Remove(recoveryPath); err != nil {
				return errors.Join(fmt.Errorf("prepare rollback link for %s: %w", item.path, err), removeCreatedRecovery())
			}
		} else {
			recoveryPath = item.path + ".bak"
		}
		if err := os.Link(item.path, recoveryPath); err != nil {
			if errors.Is(err, os.ErrExist) && !inPlace {
				return errors.Join(fmt.Errorf("refusing to overwrite existing backup: %s", recoveryPath), removeCreatedRecovery())
			}
			return errors.Join(fmt.Errorf("create recovery link for %s: %w", item.path, err), removeCreatedRecovery())
		}
		recovery[item.path] = recoveryPath
		created = append(created, recoveryPath)
		if err := compressionVerifyPath(recoveryPath, item.metadata); err != nil {
			return errors.Join(err, removeCreatedRecovery())
		}
	}
	for _, item := range staged {
		if err := compressionVerifyPath(item.path, item.metadata); err != nil {
			return errors.Join(err, removeCreatedRecovery())
		}
	}
	if err := compressionSyncDirectories(staged); err != nil {
		return errors.Join(err, removeCreatedRecovery())
	}

	// Replace the staged images. Until the replacements have been synced, any
	// handled failure restores every destination that has already changed.
	replaced := make([]compressionStagedFile, 0, len(staged))
	commitErr := error(nil)
	for _, item := range staged {
		if err := ctx.Err(); err != nil {
			commitErr = err
			break
		}
		if err := compressionVerifyPath(item.path, item.metadata); err != nil {
			commitErr = err
			break
		}
		if err := os.Rename(item.tempPath, item.path); err != nil {
			commitErr = fmt.Errorf("replace image %s: %w", item.path, err)
			break
		}
		replaced = append(replaced, item)
	}
	if commitErr == nil {
		commitErr = compressionSyncDirectories(staged)
	}
	if commitErr != nil {
		var rollbackErrs []error
		replacedSet := make(map[string]bool, len(replaced))
		for i := len(replaced) - 1; i >= 0; i-- {
			item := replaced[i]
			replacedSet[item.path] = true
			if err := os.Rename(recovery[item.path], item.path); err != nil {
				rollbackErrs = append(rollbackErrs, fmt.Errorf("restore %s from %s: %w", item.path, recovery[item.path], err))
			}
		}
		for _, item := range staged {
			if replacedSet[item.path] {
				continue
			}
			if err := os.Remove(recovery[item.path]); err != nil && !errors.Is(err, os.ErrNotExist) {
				rollbackErrs = append(rollbackErrs, fmt.Errorf("remove recovery link %s: %w", recovery[item.path], err))
			}
		}
		if err := compressionSyncDirectories(staged); err != nil {
			rollbackErrs = append(rollbackErrs, err)
		}
		if len(rollbackErrs) != 0 {
			return fmt.Errorf("commit failed: %w; rollback also failed: %w", commitErr, errors.Join(rollbackErrs...))
		}
		return fmt.Errorf("commit failed; original images were restored: %w", commitErr)
	}

	// The replacements are committed. In-place mode no longer needs its private
	// recovery links; backup mode intentionally leaves the .bak links in place.
	if inPlace {
		var cleanupErrs []error
		for _, path := range created {
			if err := removeCommittedRecovery(path); err != nil && !errors.Is(err, os.ErrNotExist) {
				cleanupErrs = append(cleanupErrs, fmt.Errorf("remove recovery link %s: %w", path, err))
			}
		}
		if err := compressionSyncDirectories(staged); err != nil {
			cleanupErrs = append(cleanupErrs, err)
		}
		if len(cleanupErrs) != 0 {
			return fmt.Errorf("%w: %w", ErrCheckpointCompressionCleanup, errors.Join(cleanupErrs...))
		}
	}
	return nil
}

func compressionVerifyPath(path string, metadata *compressionFileMetadata) error {
	var stat unix.Stat_t
	if err := unix.Lstat(path, &stat); err != nil {
		return fmt.Errorf("stat image %s: %w", path, err)
	}
	if compressionUint64(stat.Dev) != metadata.dev || stat.Ino != metadata.ino {
		return fmt.Errorf("image changed while preparing to replace: %s", path)
	}
	return nil
}

// compressionSyncDirectories persists namespace changes once per directory.
func compressionSyncDirectories(staged []compressionStagedFile) error {
	directories := make(map[string]struct{})
	for _, item := range staged {
		directories[filepath.Dir(item.path)] = struct{}{}
	}
	for directory := range directories {
		fd, err := unix.Open(directory, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
		if err != nil {
			return fmt.Errorf("open image directory %s: %w", directory, err)
		}
		syncErr := unix.Fsync(fd)
		closeErr := unix.Close(fd)
		if syncErr != nil {
			return fmt.Errorf("sync image directory %s: %w", directory, syncErr)
		}
		if closeErr != nil {
			return fmt.Errorf("close image directory %s: %w", directory, closeErr)
		}
	}
	return nil
}

func compressionCleanupStaged(staged []compressionStagedFile) error {
	var cleanupErrs []error
	for _, item := range staged {
		if err := os.Remove(item.tempPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			cleanupErrs = append(cleanupErrs, fmt.Errorf("remove staged image %s: %w", item.tempPath, err))
		}
	}
	return errors.Join(cleanupErrs...)
}
