package crit

import (
	"fmt"

	"github.com/pierrec/lz4/v4"
)

type cachedMemoryBlock struct {
	entry *memoryEntry
	index uint64
	data  []byte
}

type memoryLZ4State struct {
	cache   cachedMemoryBlock
	encoded []byte
	decoded []byte
}

func decodeMemoryBlock(encoded, decoded []byte) error {
	n, err := lz4.UncompressBlock(encoded, decoded)
	if err != nil {
		return err
	}
	if n != len(decoded) {
		return fmt.Errorf("produced %d bytes, expected %d", n, len(decoded))
	}
	return nil
}

func (session *memoryReadSession) readLZ4BlockInto(
	entry *memoryEntry,
	blockIndex, payloadOffset, pageInBlock, blockBytes, vaddr uint64,
	page []byte,
) error {
	state := &session.lz4
	if state.cache.entry != entry || state.cache.index != blockIndex {
		state.cache = cachedMemoryBlock{}
		storedSize := uint64(entry.blockSizes[blockIndex])
		if cap(state.encoded) < int(storedSize) {
			state.encoded = make([]byte, int(storedSize))
		}
		encoded := state.encoded[:int(storedSize)]
		if err := readAtFull(session.pagesFile, encoded, payloadOffset); err != nil {
			return fmt.Errorf("read compressed block at address %#x: %w", vaddr, err)
		}
		if cap(state.decoded) < int(blockBytes) {
			state.decoded = make([]byte, int(blockBytes))
		}
		decoded := state.decoded[:int(blockBytes)]
		if err := decodeMemoryBlock(encoded, decoded); err != nil {
			return fmt.Errorf("decompress block at address %#x: %w", vaddr, err)
		}
		state.cache = cachedMemoryBlock{entry: entry, index: blockIndex, data: decoded}
	}
	start := pageInBlock * uint64(session.layer.pageSize)
	copy(page, state.cache.data[start:start+uint64(session.layer.pageSize)])
	return nil
}
