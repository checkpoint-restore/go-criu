package crit

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/checkpoint-restore/go-criu/v6/crit/images/mm"
	"github.com/checkpoint-restore/go-criu/v6/crit/images/pagemap"
	"golang.org/x/sys/unix"
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

		if pageNumber == startPage {
			nSkip = start - pageNumber*uint64(mr.pageSize)
			if startPage == endPage {
				nRead = size
			} else {
				nRead = uint64(mr.pageSize) - nSkip
			}
		} else if pageNumber == endPage {
			nSkip = 0
			nRead = end - pageNumber*uint64(mr.pageSize)
		} else {
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
		for i := 0; i < int(*m.NrPages); i++ {
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

		defer f.Close()

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

	return mr.GetMemPages(*mm.MmArgStart, *mm.MmArgEnd)
}

// GetPsArgs retrieves process environment variables from memory pages.
func (mr *MemoryReader) GetPsEnvVars() (*bytes.Buffer, error) {
	mmImg, err := getImg(filepath.Join(mr.checkpointDir, fmt.Sprintf("mm-%d.img", mr.pid)), &mm.MmEntry{})
	if err != nil {
		return nil, err
	}
	mm := mmImg.Entries[0].Message.(*mm.MmEntry)

	return mr.GetMemPages(*mm.MmEnvStart, *mm.MmEnvEnd)
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
	for _, vma := range mm.Vmas {
		// Check if VMA has the MAP_SHARED flag set in its flags
		if vma.GetFlags()&unix.MAP_SHARED != 0 {
			size += int64(vma.GetEnd() - vma.GetStart())
		}
	}

	return size, nil
}
