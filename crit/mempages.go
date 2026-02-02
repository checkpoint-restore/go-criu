package crit

import (
	"bytes"
	"errors"
	"fmt"
	"io"
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

	var results []PatternMatch

	f, err := os.Open(filepath.Join(mr.checkpointDir, fmt.Sprintf("pages-%d.img", mr.pagesID)))
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	for _, entry := range mr.pagemapEntries {
		startAddr := entry.GetVaddr()
		endAddr := startAddr + entry.GetNrPages()*uint64(mr.pageSize)

		initialOffset := uint64(0)
		for _, e := range mr.pagemapEntries {
			if e == entry {
				break
			}
			initialOffset += e.GetNrPages() * uint64(mr.pageSize)
		}

		for offset := uint64(0); offset < endAddr-startAddr; offset += uint64(chunkSize) {
			readSize := chunkSize
			if endAddr-startAddr-offset < uint64(chunkSize) {
				readSize = int(endAddr - startAddr - offset)
			}

			buff := make([]byte, readSize)
			if _, err := f.ReadAt(buff, int64(initialOffset+offset)); err != nil {
				if err == io.EOF {
					break
				}
				return nil, err
			}

			// Replace non-printable ASCII characters in the buffer with a question mark (0x3f) to prevent unexpected behavior
			// during regex matching. Non-printable characters might cause incorrect interpretation or premature
			// termination of strings, leading to inaccuracies in pattern matching.
			for i := range buff {
				if buff[i] < 32 || buff[i] >= 127 {
					buff[i] = 0x3F
				}
			}

			indexes := regexPattern.FindAllIndex(buff, -1)
			for _, index := range indexes {
				startContext := index[0] - context
				if startContext < 0 {
					startContext = 0
				}

				endContext := index[1] + context
				if endContext > len(buff) {
					endContext = len(buff)
				}

				results = append(results, PatternMatch{
					Vaddr:   startAddr + offset + uint64(index[0]),
					Length:  index[1] - index[0],
					Context: context,
					Match:   string(buff[startContext:endContext]),
				})
			}
		}
	}

	return results, nil
}
