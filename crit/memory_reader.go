package crit

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
)

type cachedMemoryBlockOffset struct {
	entry  *memoryEntry
	index  uint64
	offset uint64
}

type memoryReadSession struct {
	layer       *memoryLayer
	pagesFile   *os.File
	offsetCache cachedMemoryBlockOffset
}

func newMemoryReadSession(layer *memoryLayer) (*memoryReadSession, error) {
	return &memoryReadSession{layer: layer}, nil
}

func (session *memoryReadSession) close() error {
	if session.pagesFile == nil {
		return nil
	}
	return session.pagesFile.Close()
}

func (session *memoryReadSession) ensurePagesFile() error {
	if session.pagesFile != nil {
		return nil
	}
	path := filepath.Join(
		session.layer.directory,
		fmt.Sprintf("pages-%d.img", session.layer.pagesID),
	)
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open pages image %s: %w", path, err)
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return fmt.Errorf("stat pages image %s: %w", path, err)
	}
	if !info.Mode().IsRegular() {
		_ = file.Close()
		return fmt.Errorf("pages image %s is not a regular file", path)
	}
	if session.layer.payloadSize > math.MaxInt64 {
		_ = file.Close()
		return fmt.Errorf(
			"%s describes %d payload bytes, but %s contains %d",
			session.layer.pagemapName,
			session.layer.payloadSize,
			filepath.Base(path),
			info.Size(),
		)
	}
	actualSize := uint64(info.Size())
	if actualSize != session.layer.payloadSize {
		_ = file.Close()
		err := fmt.Errorf(
			"%s describes %d payload bytes, but %s contains %d",
			session.layer.pagemapName,
			session.layer.payloadSize,
			filepath.Base(path),
			info.Size(),
		)
		if actualSize < session.layer.payloadSize {
			err = fmt.Errorf("%w: %w", io.ErrUnexpectedEOF, err)
		}
		return err
	}
	session.pagesFile = file
	return nil
}

func (session *memoryReadSession) blockPayloadOffset(entry *memoryEntry, blockIndex uint64) uint64 {
	if session.offsetCache.entry == entry {
		switch blockIndex {
		case session.offsetCache.index:
			return session.offsetCache.offset
		case session.offsetCache.index + 1:
			offset := session.offsetCache.offset + uint64(entry.blockSizes[session.offsetCache.index])
			session.offsetCache = cachedMemoryBlockOffset{entry: entry, index: blockIndex, offset: offset}
			return offset
		}
	}
	offset := entry.blockPayloadOffset(blockIndex)
	session.offsetCache = cachedMemoryBlockOffset{entry: entry, index: blockIndex, offset: offset}
	return offset
}

func (session *memoryReadSession) readPageInto(vaddr uint64, page []byte) (bool, error) {
	if len(page) != session.layer.pageSize {
		return false, fmt.Errorf("page buffer has %d bytes, expected %d", len(page), session.layer.pageSize)
	}
	entry := session.layer.findEntry(vaddr)
	if entry == nil {
		return false, nil
	}
	if entry.flags&peParent != 0 {
		return false, fmt.Errorf("page at address %#x is stored in a parent image", vaddr)
	}
	if entry.flags&pePresent == 0 {
		return false, nil
	}
	if err := session.ensurePagesFile(); err != nil {
		return false, err
	}

	pageIndex := (vaddr - entry.vaddr) / uint64(session.layer.pageSize)
	if !entry.hasCompressionMetadata {
		offset := entry.payloadOffset + pageIndex*uint64(session.layer.pageSize)
		if err := readAtFull(session.pagesFile, page, offset); err != nil {
			return false, fmt.Errorf("read page at address %#x: %w", vaddr, err)
		}
		return true, nil
	}

	blockPages := entry.blockPages
	blockIndex := pageIndex / blockPages
	firstPage, pageCount, storedSize, ok := entry.blockLayout(blockIndex)
	if !ok {
		return false, fmt.Errorf("compressed block index %d is outside entry at address %#x", blockIndex, entry.vaddr)
	}
	payloadOffset := session.blockPayloadOffset(entry, blockIndex)
	pageInBlock := pageIndex - firstPage
	blockBytes := pageCount * uint64(session.layer.pageSize)
	if storedSize == 0 {
		clear(page)
		return true, nil
	}
	if storedSize == blockBytes {
		offset := payloadOffset + pageInBlock*uint64(session.layer.pageSize)
		if err := readAtFull(session.pagesFile, page, offset); err != nil {
			return false, fmt.Errorf("read raw block at address %#x: %w", vaddr, err)
		}
		return true, nil
	}
	return false, fmt.Errorf("LZ4 block at address %#x is not supported", vaddr)
}

func readAtFull(file *os.File, buffer []byte, offset uint64) error {
	if offset > math.MaxInt64 {
		return fmt.Errorf("file offset %d exceeds supported range", offset)
	}
	n, err := file.ReadAt(buffer, int64(offset))
	if err != nil {
		return err
	}
	if n != len(buffer) {
		return io.ErrUnexpectedEOF
	}
	return nil
}

func (mr *MemoryReader) memoryLayer() (*memoryLayer, error) {
	if mr.layer != nil {
		return mr.layer, nil
	}
	return indexMemoryLayer(
		mr.checkpointDir,
		fmt.Sprintf("pagemap-%d.img", mr.pid),
		mr.pagesID,
		mr.pageSize,
		mr.pagemapEntries,
	)
}

func (mr *MemoryReader) newReadSession() (*memoryReadSession, error) {
	layer, err := mr.memoryLayer()
	if err != nil {
		return nil, err
	}
	return newMemoryReadSession(layer)
}

func (mr *MemoryReader) readMemRange(
	session *memoryReadSession,
	start, end uint64,
) (*bytes.Buffer, error) {
	var buffer bytes.Buffer
	if end-start > uint64(math.MaxInt) {
		return nil, fmt.Errorf("memory range %#x-%#x is too large", start, end)
	}
	buffer.Grow(int(end - start))
	pageSize := uint64(mr.pageSize)
	page := make([]byte, mr.pageSize)
	for cursor := start; cursor < end; {
		pageAddress := cursor - cursor%pageSize
		found, err := session.readPageInto(pageAddress, page)
		if err != nil {
			return nil, err
		}
		if !found {
			clear(page)
		}
		pageOffset := cursor - pageAddress
		readSize := min(pageSize-pageOffset, end-cursor)
		_, _ = buffer.Write(page[pageOffset : pageOffset+readSize])
		cursor += readSize
	}
	return &buffer, nil
}
