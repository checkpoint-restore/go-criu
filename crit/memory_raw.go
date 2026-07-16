package crit

import "fmt"

func (entry *memoryEntry) rawPayloadRun(start, end, pageSize uint64) (uint64, uint64, bool) {
	if start < entry.vaddr || end < start || end > entry.end {
		return 0, start, false
	}
	if !entry.hasCompressionMetadata {
		return entry.payloadOffset + start - entry.vaddr, end, true
	}

	pageIndex := (start - entry.vaddr) / pageSize
	blockPages := entry.blockPages
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

func (session *memoryReadSession) copyRawRange(output []byte, entry *memoryEntry, start, end uint64) (uint64, bool, error) {
	payloadOffset, runEnd, ok := entry.rawPayloadRun(start, end, uint64(session.layer.pageSize))
	if !ok {
		return start, false, nil
	}
	if err := session.ensurePagesFile(); err != nil {
		return start, false, err
	}
	if err := readAtFull(session.pagesFile, output[:runEnd-start], payloadOffset); err != nil {
		return start, false, fmt.Errorf("read raw memory range %#x-%#x: %w", start, runEnd, err)
	}
	return runEnd, true, nil
}
