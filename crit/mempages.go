package crit

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"

	"github.com/checkpoint-restore/go-criu/v8/crit/images/mm"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/pagemap"
)

var sysPageSize = os.Getpagesize()

// MemoryReader is a struct used to retrieve
// the content of memory associated with a specific process ID (pid).
// New instances should be created with NewMemoryReader()
type MemoryReader struct {
	checkpointDir  string
	pid            uint32
	pagesID        uint32
	pageSize       int
	pagemapEntries []*pagemap.PagemapEntry
	layer          *memoryLayer
}

func (mr *MemoryReader) GetPagesID() uint32 {
	return mr.pagesID
}

// NewMemoryReader creates a new instance of MemoryReader with all the fields populated
func NewMemoryReader(checkpointDir string, pid uint32, pageSize int) (*MemoryReader, error) {
	if pageSize == 0 {
		pageSize = sysPageSize
	}

	// Check if the given page size is a positive power of 2, otherwise return an error
	if pageSize <= 0 || (pageSize&(pageSize-1)) != 0 {
		return nil, errors.New("page size should be a positive power of 2")
	}

	pagemapName := fmt.Sprintf("pagemap-%d.img", pid)
	layer, err := loadMemoryLayer(checkpointDir, pagemapName, pageSize)
	if err != nil {
		return nil, err
	}

	return &MemoryReader{
		checkpointDir:  checkpointDir,
		pid:            pid,
		pageSize:       pageSize,
		pagesID:        layer.pagesID,
		pagemapEntries: layer.pagemapEntries,
		layer:          layer,
	}, nil
}

// GetMemPages retrieves the content of memory pages
// associated with a given process ID (pid).
// It retrieves the memory content within the
// specified range defined by the start and end addresses.
func (mr *MemoryReader) GetMemPages(start, end uint64) (*bytes.Buffer, error) {
	if end < start {
		return nil, fmt.Errorf("memory range end %#x precedes start %#x", end, start)
	}
	if end == start {
		return &bytes.Buffer{}, nil
	}

	session, err := mr.newReadSession()
	if err != nil {
		return nil, err
	}
	defer func() { _ = session.close() }()
	return mr.readMemRange(session, start, end)
}

// GetPsArgs retrieves process arguments from memory pages
func (mr *MemoryReader) GetPsArgs() (*bytes.Buffer, error) {
	mmImg, err := getImg(filepath.Join(mr.checkpointDir, fmt.Sprintf("mm-%d.img", mr.pid)), &mm.MmEntry{})
	if err != nil {
		return nil, err
	}
	mm := mmImg.Entries[0].Message.(*mm.MmEntry)

	return mr.GetMemPages(mm.GetMmArgStart(), mm.GetMmArgEnd())
}

// GetPsArgs retrieves process environment variables from memory pages.
func (mr *MemoryReader) GetPsEnvVars() (*bytes.Buffer, error) {
	mmImg, err := getImg(filepath.Join(mr.checkpointDir, fmt.Sprintf("mm-%d.img", mr.pid)), &mm.MmEntry{})
	if err != nil {
		return nil, err
	}
	mm := mmImg.Entries[0].Message.(*mm.MmEntry)

	return mr.GetMemPages(mm.GetMmEnvStart(), mm.GetMmEnvEnd())
}

func (mr *MemoryReader) GetPagemapEntries() []*pagemap.PagemapEntry {
	return mr.pagemapEntries
}

// GetShmemSize calculates and returns the size of shared memory used by the process.
func (mr *MemoryReader) GetShmemSize() (int64, error) {
	mmImg, err := getImg(filepath.Join(mr.checkpointDir, fmt.Sprintf("mm-%d.img", mr.pid)), &mm.MmEntry{})
	if err != nil {
		return 0, err
	}

	var size int64
	mm := mmImg.Entries[0].Message.(*mm.MmEntry)
	for _, vma := range mm.GetVmas() {
		// Check if VMA has the MAP_SHARED flag set in its flags
		if vma.GetFlags()&mapShared != 0 {
			size += int64(vma.GetEnd() - vma.GetStart())
		}
	}

	return size, nil
}

// PatternMatch represents a match when searching for a pattern in memory.
type PatternMatch struct {
	Vaddr   uint64
	Length  int
	Context int
	Match   string
}

// memoryEntryReaderAt exposes one decoded pagemap entry as an io.ReaderAt.
// The backing session resolves raw, zero, compressed, and parent-backed pages;
// callers therefore see logical memory rather than the packed pages image.
type memoryEntryReaderAt struct {
	mr      *MemoryReader
	session *memoryReadSession
	entry   *memoryEntry
}

func (r *memoryEntryReaderAt) ReadAt(buffer []byte, offset int64) (int, error) {
	if len(buffer) == 0 {
		return 0, nil
	}
	if offset < 0 {
		return 0, fmt.Errorf("memory entry offset cannot be negative: %d", offset)
	}
	if r.entry == nil {
		return 0, errors.New("memory entry is not selected")
	}

	entrySize := r.entry.end - r.entry.vaddr
	entryOffset := uint64(offset)
	if entryOffset >= entrySize {
		return 0, io.EOF
	}
	readSize := min(uint64(len(buffer)), entrySize-entryOffset)
	start := r.entry.vaddr + entryOffset
	memory, err := r.mr.readMemRange(r.session, start, start+readSize)
	if err != nil {
		return 0, err
	}
	n := copy(buffer, memory.Bytes())
	if n != len(buffer) {
		return n, io.EOF
	}
	return n, nil
}

func readMemoryAt(reader io.ReaderAt, buff []byte, initialOffset, offset uint64) error {
	if len(buff) == 0 {
		return nil
	}
	if initialOffset > uint64(math.MaxInt64) || offset > uint64(math.MaxInt64)-initialOffset {
		return fmt.Errorf("memory image offset is too large: %d + %d", initialOffset, offset)
	}

	n, err := reader.ReadAt(buff, int64(initialOffset+offset))
	if n == len(buff) {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}
	if err == nil || errors.Is(err, io.EOF) {
		return io.ErrUnexpectedEOF
	}
	return err
}

type memoryRuneReader struct {
	reader        io.ReaderAt
	initialOffset uint64
	entrySize     uint64
	position      uint64
	buffer        []byte
	bufferStart   uint64
	bufferEnd     uint64
	err           error
}

func newMemoryRuneReader(reader io.ReaderAt, chunkSize int) *memoryRuneReader {
	return &memoryRuneReader{
		reader: reader,
		buffer: make([]byte, chunkSize),
	}
}

func (r *memoryRuneReader) setEntry(initialOffset, entrySize uint64) {
	r.initialOffset = initialOffset
	r.entrySize = entrySize
	r.bufferStart = 0
	r.bufferEnd = 0
}

func (r *memoryRuneReader) reset(position uint64) {
	r.position = position
	r.err = nil
}

func (r *memoryRuneReader) ReadRune() (rune, int, error) {
	if r.position >= r.entrySize {
		return 0, 0, io.EOF
	}
	if r.position < r.bufferStart || r.position >= r.bufferEnd {
		readSize := min(uint64(len(r.buffer)), r.entrySize-r.position)
		if err := readMemoryAt(r.reader, r.buffer[:int(readSize)], r.initialOffset, r.position); err != nil {
			r.err = err
			return 0, 0, err
		}
		r.bufferStart = r.position
		r.bufferEnd = r.position + readSize
	}
	b := r.buffer[r.position-r.bufferStart]
	r.position++
	if b < 32 || b >= 127 {
		b = '?'
	}
	return rune(b), 1, nil
}

type streamingRegexps struct {
	initial      *regexp.Regexp
	continuation *regexp.Regexp
}

func compileStreamingRegexps(pattern string) (*streamingRegexps, error) {
	// Anchor a lazy prefix and capture the earliest match of pattern. A
	// continuation consumes the byte preceding the next search position so
	// boundary assertions observe the same context as they do in one buffer.
	initial, err := regexp.Compile(`\A(?s:.*?)((?:` + pattern + `))`)
	if err != nil {
		return nil, err
	}
	continuation, err := regexp.Compile(`\A(?s:.)(?s:.*?)((?:` + pattern + `))`)
	if err != nil {
		return nil, err
	}
	return &streamingRegexps{
		initial:      initial,
		continuation: continuation,
	}, nil
}

func sanitizeMemory(buff []byte) {
	for i := range buff {
		if buff[i] < 32 || buff[i] >= 127 {
			buff[i] = '?'
		}
	}
}

func readPatternMatch(
	reader io.ReaderAt,
	initialOffset, startAddr, entrySize, matchStart, matchEnd uint64,
	context int,
	cachedStart uint64,
	cached []byte,
) (PatternMatch, error) {
	contextSize := uint64(context)
	contextStart := uint64(0)
	if contextSize < matchStart {
		contextStart = matchStart - contextSize
	}

	contextEnd := entrySize
	if contextSize < entrySize-matchEnd {
		contextEnd = matchEnd + contextSize
	}

	readSize := contextEnd - contextStart
	if readSize > uint64(math.MaxInt) {
		return PatternMatch{}, fmt.Errorf("memory match is too large: %d bytes", readSize)
	}

	var buff []byte
	cachedEnd := cachedStart + uint64(len(cached))
	if contextStart >= cachedStart && contextEnd <= cachedEnd {
		buff = cached[contextStart-cachedStart : contextEnd-cachedStart]
	} else {
		buff = make([]byte, int(readSize))
		if err := readMemoryAt(reader, buff, initialOffset, contextStart); err != nil {
			return PatternMatch{}, err
		}
		sanitizeMemory(buff)
	}

	return PatternMatch{
		Vaddr:   startAddr + matchStart,
		Length:  int(matchEnd - matchStart),
		Context: context,
		Match:   string(buff),
	}, nil
}

func literalFailureTable(pattern string) []int {
	table := make([]int, len(pattern))
	matched := 0
	for i := 1; i < len(pattern); i++ {
		for matched > 0 && pattern[i] != pattern[matched] {
			matched = table[matched-1]
		}
		if pattern[i] == pattern[matched] {
			matched++
		}
		table[i] = matched
	}
	return table
}

func searchLiteralPattern(
	reader io.ReaderAt,
	literalPattern string,
	patternTable []int,
	initialOffset, startAddr, entrySize uint64,
	context, chunkSize int,
) ([]PatternMatch, error) {
	if len(literalPattern) == 0 {
		return nil, errors.New("literal pattern cannot be empty")
	}
	canMatch := uint64(len(literalPattern)) <= entrySize
	if canMatch && len(patternTable) != len(literalPattern) {
		return nil, errors.New("literal pattern table has an invalid size")
	}

	var results []PatternMatch
	bufferSize := min(uint64(chunkSize), entrySize)
	buff := make([]byte, int(bufferSize))
	matched := 0

	for offset := uint64(0); offset < entrySize; offset += uint64(chunkSize) {
		readSize := int(min(uint64(chunkSize), entrySize-offset))
		window := buff[:readSize]
		if err := readMemoryAt(reader, window, initialOffset, offset); err != nil {
			return nil, err
		}
		if !canMatch {
			continue
		}
		sanitizeMemory(window)

		for i, b := range window {
			for matched > 0 && b != literalPattern[matched] {
				matched = patternTable[matched-1]
			}
			if b == literalPattern[matched] {
				matched++
			}
			if matched != len(literalPattern) {
				continue
			}

			matchEnd := offset + uint64(i) + 1
			matchStart := matchEnd - uint64(len(literalPattern))

			match, err := readPatternMatch(
				reader,
				initialOffset,
				startAddr,
				entrySize,
				matchStart,
				matchEnd,
				context,
				offset,
				window,
			)
			if err != nil {
				return nil, err
			}
			results = append(results, match)

			// regexp.FindAllIndex reports non-overlapping matches.
			matched = 0
		}
	}

	return results, nil
}

func searchPatternStream(
	readerAt io.ReaderAt,
	patterns *streamingRegexps,
	reader *memoryRuneReader,
	initialOffset, startAddr, entrySize uint64,
	context int,
) ([]PatternMatch, error) {
	var results []PatternMatch
	searchOffset := uint64(0)
	previousMatchEnd := uint64(0)
	havePreviousMatch := false
	reader.setEntry(initialOffset, entrySize)

	for {
		readerOffset := searchOffset
		regexPattern := patterns.initial
		if searchOffset > 0 {
			readerOffset--
			regexPattern = patterns.continuation
		}

		reader.reset(readerOffset)
		indexes := regexPattern.FindReaderSubmatchIndex(reader)
		if reader.err != nil {
			return nil, reader.err
		}
		if indexes == nil {
			break
		}
		if len(indexes) < 4 || indexes[2] < 0 || indexes[3] < 0 {
			return nil, errors.New("streaming regexp did not capture its match")
		}

		matchStart := readerOffset + uint64(indexes[2])
		matchEnd := readerOffset + uint64(indexes[3])
		if matchStart < searchOffset || matchEnd < matchStart || matchEnd > entrySize {
			return nil, errors.New("streaming regexp returned an invalid match range")
		}

		acceptMatch := true
		done := false
		if matchEnd == searchOffset {
			// Mirror regexp.FindAllIndex: ignore an empty match immediately
			// after the previous match and advance by one input byte.
			if havePreviousMatch && matchStart == previousMatchEnd {
				acceptMatch = false
			}
			if searchOffset == entrySize {
				done = true
			} else {
				searchOffset++
			}
		} else {
			searchOffset = matchEnd
		}
		previousMatchEnd = matchEnd
		havePreviousMatch = true

		if acceptMatch {
			match, err := readPatternMatch(
				readerAt,
				initialOffset,
				startAddr,
				entrySize,
				matchStart,
				matchEnd,
				context,
				0,
				nil,
			)
			if err != nil {
				return nil, err
			}
			results = append(results, match)
		}
		if done {
			break
		}
	}

	return results, nil
}

// SearchPattern searches for a pattern in the process memory pages.
func (mr *MemoryReader) SearchPattern(pattern string, escapeRegExpCharacters bool, context, chunkSize int) ([]PatternMatch, error) {
	if context < 0 {
		return nil, errors.New("context size cannot be negative")
	}

	// Set a default chunk size of 10MB to be read at a time
	if chunkSize <= 0 {
		chunkSize = 10 * 1024 * 1024
	}

	// Escape regular expression characters in the pattern
	if escapeRegExpCharacters {
		pattern = regexp.QuoteMeta(pattern)
	}

	regexPattern, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	literalPattern, literalOnly := regexPattern.LiteralPrefix()
	needsStreaming := !literalOnly || len(literalPattern) == 0
	var patternTable []int
	var streamPatterns *streamingRegexps
	if needsStreaming {
		streamPatterns, err = compileStreamingRegexps(pattern)
		if err != nil {
			return nil, err
		}
	}

	var results []PatternMatch
	layer, err := mr.memoryLayer()
	if err != nil {
		return nil, err
	}
	session, err := newMemoryReadSession(layer)
	if err != nil {
		return nil, err
	}
	defer func() { _ = session.close() }()

	entryReader := &memoryEntryReaderAt{mr: mr, session: session}
	var streamReader *memoryRuneReader
	if needsStreaming {
		streamReader = newMemoryRuneReader(entryReader, chunkSize)
	}

	for index := range layer.entries {
		entry := &layer.entries[index]
		entryReader.entry = entry
		startAddr := entry.vaddr
		entrySize := entry.end - entry.vaddr

		if needsStreaming {
			matches, err := searchPatternStream(
				entryReader,
				streamPatterns,
				streamReader,
				0,
				startAddr,
				entrySize,
				context,
			)
			if err != nil {
				return nil, err
			}
			results = append(results, matches...)
			continue
		}

		if patternTable == nil && uint64(len(literalPattern)) <= entrySize {
			patternTable = literalFailureTable(literalPattern)
		}
		matches, err := searchLiteralPattern(
			entryReader,
			literalPattern,
			patternTable,
			0,
			startAddr,
			entrySize,
			context,
			chunkSize,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, matches...)
	}

	return results, nil
}
