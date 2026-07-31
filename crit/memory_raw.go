package crit

import (
	"fmt"
	"io"
	"math"
)

func (entry *memoryEntry) rawPayloadRun(start, end, pageSize uint64) (uint64, uint64, bool) {
	if start < entry.vaddr || end < start || end > entry.end {
		return 0, start, false
	}
	if !entry.hasCompressionMetadata {
		return entry.payloadOffset + start - entry.vaddr, end, true
	}

	pageIndex := (start - entry.vaddr) / pageSize
	blockPages := entry.regionPages
	if blockPages == 0 {
		blockPages = 1
	}
	blockIndex := pageIndex / blockPages
	firstPage, pageCount, storedSize, ok := entry.blockLayout(blockIndex)
	if !ok {
		return 0, start, false
	}
	blockBytes := pageCount * pageSize
	if storedSize != blockBytes {
		return 0, start, false
	}
	blockStart := entry.vaddr + firstPage*pageSize
	blockPayloadOffset := entry.blockPayloadOffset(blockIndex)
	payloadOffset := blockPayloadOffset + start - blockStart
	runEnd := min(end, blockStart+blockBytes)
	expectedPayloadOffset := blockPayloadOffset + blockBytes

	for runEnd < end {
		blockIndex++
		firstPage, pageCount, storedSize, ok = entry.blockLayout(blockIndex)
		if !ok {
			break
		}
		blockBytes = pageCount * pageSize
		blockStart = entry.vaddr + firstPage*pageSize
		blockPayloadOffset = expectedPayloadOffset
		if blockStart != runEnd || storedSize != blockBytes {
			break
		}
		runEnd = min(end, blockStart+blockBytes)
		expectedPayloadOffset = blockPayloadOffset + blockBytes
	}
	return payloadOffset, runEnd, true
}

func (session *memoryReadSession) copyRawRange(output io.Writer, entry *memoryEntry, start, end uint64) (uint64, bool, error) {
	payloadOffset, runEnd, ok := entry.rawPayloadRun(start, end, uint64(session.layer.pageSize))
	if !ok {
		return start, false, nil
	}
	if err := session.ensurePagesFile(); err != nil {
		return start, false, err
	}
	length := runEnd - start
	if payloadOffset > math.MaxInt64 || length > math.MaxInt64 {
		return start, false, fmt.Errorf("raw memory range %#x-%#x exceeds supported file offsets", start, runEnd)
	}
	reader := io.NewSectionReader(session.pagesFile, int64(payloadOffset), int64(length))
	written, err := io.CopyN(output, reader, int64(length))
	if err != nil {
		return start, false, fmt.Errorf("read raw memory range %#x-%#x: %w", start, runEnd, err)
	}
	if written != int64(length) {
		return start, false, fmt.Errorf("read raw memory range %#x-%#x: %w", start, runEnd, io.ErrUnexpectedEOF)
	}
	return runEnd, true, nil
}
