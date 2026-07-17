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
	if (pageSize & (pageSize - 1)) != 0 {
		return nil, errors.New("page size should be a positive power of 2")
	}

	pagemapImg, err := getImg(filepath.Join(checkpointDir, fmt.Sprintf("pagemap-%d.img", pid)), &pagemap.PagemapHead{})
	if err != nil {
		return nil, err
	}

	pagesID := pagemapImg.Entries[0].Message.(*pagemap.PagemapHead).GetPagesId()

	pagemapEntries := make([]*pagemap.PagemapEntry, 0)

	for _, entry := range pagemapImg.Entries[1:] {
		pagemapEntries = append(pagemapEntries, entry.Message.(*pagemap.PagemapEntry))
	}

	return &MemoryReader{
		checkpointDir:  checkpointDir,
		pid:            pid,
		pageSize:       pageSize,
		pagesID:        pagesID,
		pagemapEntries: pagemapEntries,
	}, nil
}

// GetMemPages retrieves the content of memory pages
// associated with a given process ID (pid).
// It retrieves the memory content within the
// specified range defined by the start and end addresses.
func (mr *MemoryReader) GetMemPages(start, end uint64) (*bytes.Buffer, error) {
	size := end - start

	startPage := start / uint64(mr.pageSize)
	endPage := end / uint64(mr.pageSize)

	var buffer bytes.Buffer

	for pageNumber := startPage; pageNumber <= endPage; pageNumber++ {
		var page []byte = nil

		pageMem, err := mr.getPage(pageNumber)
		if err != nil {
			return nil, err
		}

		if pageMem != nil {
			page = pageMem
		} else {
			page = bytes.Repeat([]byte("\x00"), mr.pageSize)
		}

		var nSkip, nRead uint64

		switch pageNumber {
		case startPage:
			nSkip = start - pageNumber*uint64(mr.pageSize)
			if startPage == endPage {
				nRead = size
			} else {
				nRead = uint64(mr.pageSize) - nSkip
			}
		case endPage:
			nSkip = 0
			nRead = end - pageNumber*uint64(mr.pageSize)
		default:
			nSkip = 0
			nRead = uint64(mr.pageSize)
		}

		if _, err := buffer.Write(page[nSkip : nSkip+nRead]); err != nil {
			return nil, err
		}
	}

	return &buffer, nil
}

// getPage retrieves a memory page from the pages.img file.
func (mr *MemoryReader) getPage(pageNo uint64) ([]byte, error) {
	var offset uint64 = 0

	// Iterate over pagemap entries to find the corresponding page
	for _, m := range mr.pagemapEntries {
		found := false
		for i := 0; i < int(m.GetNrPages()); i++ {
			if m.GetVaddr()+uint64(i)*uint64(mr.pageSize) == pageNo*uint64(mr.pageSize) {
				found = true
				break
			}
			offset += uint64(mr.pageSize)
		}

		if !found {
			continue
		}
		f, err := os.Open(filepath.Join(mr.checkpointDir, fmt.Sprintf("pages-%d.img", mr.pagesID)))
		if err != nil {
			return nil, err
		}

		defer func() { _ = f.Close() }()

		buff := make([]byte, mr.pageSize)

		if _, err := f.ReadAt(buff, int64(offset)); err != nil {
			return nil, err
		}

		return buff, nil
	}
	return nil, nil
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
	var patternTable []int

	var results []PatternMatch

	f, err := os.Open(filepath.Join(mr.checkpointDir, fmt.Sprintf("pages-%d.img", mr.pagesID)))
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	for _, entry := range mr.pagemapEntries {
		startAddr := entry.GetVaddr()
		endAddr := startAddr + entry.GetNrPages()*uint64(mr.pageSize)
		entrySize := endAddr - startAddr

		initialOffset := uint64(0)
		for _, e := range mr.pagemapEntries {
			if e == entry {
				break
			}
			initialOffset += e.GetNrPages() * uint64(mr.pageSize)
		}

		if literalOnly && len(literalPattern) > 0 {
			if patternTable == nil && uint64(len(literalPattern)) <= entrySize {
				patternTable = literalFailureTable(literalPattern)
			}
			matches, err := searchLiteralPattern(
				f,
				literalPattern,
				patternTable,
				initialOffset,
				startAddr,
				entrySize,
				context,
				chunkSize,
			)
			if err != nil {
				return nil, err
			}
			results = append(results, matches...)
			continue
		}
		for offset := uint64(0); offset < entrySize; offset += uint64(chunkSize) {
			readSize := chunkSize
			if entrySize-offset < uint64(chunkSize) {
				readSize = int(entrySize - offset)
			}

			buff := make([]byte, readSize)
			if err := readMemoryAt(f, buff, initialOffset, offset); err != nil {
				return nil, err
			}

			// Replace non-printable ASCII characters in the buffer with a question mark (0x3f) to prevent unexpected behavior
			// during regex matching. Non-printable characters might cause incorrect interpretation or premature
			// termination of strings, leading to inaccuracies in pattern matching.
			sanitizeMemory(buff)

			indexes := regexPattern.FindAllIndex(buff, -1)
			for _, index := range indexes {
				matchStart := offset + uint64(index[0])
				matchEnd := offset + uint64(index[1])
				match, err := readPatternMatch(
					f,
					initialOffset,
					startAddr,
					entrySize,
					matchStart,
					matchEnd,
					context,
					offset,
					buff,
				)
				if err != nil {
					return nil, err
				}
				results = append(results, match)
			}
		}
	}

	return results, nil
}
