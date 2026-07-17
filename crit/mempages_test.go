package crit

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/checkpoint-restore/go-criu/v8/crit/images/pagemap"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/pstree"
)

const (
	testImgsDir               = "test-imgs"
	searchPatternTestPageSize = 64
)

type readerAtFunc func([]byte, int64) (int, error)

func (f readerAtFunc) ReadAt(buff []byte, offset int64) (int, error) {
	return f(buff, offset)
}

type countingReaderAt struct {
	data       []byte
	bytesRead  int
	readStarts []int64
	readSizes  []int
}

func (r *countingReaderAt) ReadAt(buff []byte, offset int64) (int, error) {
	r.readStarts = append(r.readStarts, offset)
	r.readSizes = append(r.readSizes, len(buff))
	if offset < 0 || offset >= int64(len(r.data)) {
		return 0, io.EOF
	}
	n := copy(buff, r.data[offset:])
	r.bytesRead += n
	if n != len(buff) {
		return n, io.EOF
	}
	return n, nil
}

func TestNewMemoryReader(t *testing.T) {
	pid, err := getTestImgPID()
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		name          string
		dir           string
		pid           uint32
		pageSize      int
		expectedError error
	}{
		{
			name:          "Page size is 0",
			dir:           testImgsDir,
			pid:           pid,
			pageSize:      0,
			expectedError: nil,
		},
		{
			name:          "Invalid page size",
			dir:           testImgsDir,
			pid:           pid,
			pageSize:      4000,
			expectedError: errors.New("page size should be a positive power of 2"),
		},
		{
			name:          "Invalid test-imgs directory",
			dir:           "no test directory",
			pid:           pid,
			expectedError: errors.New("no such file or directory"),
		},
		{
			name:          "Valid test-imgs directory, pid and page size",
			dir:           testImgsDir,
			pid:           pid,
			pageSize:      4096,
			expectedError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mr, err := NewMemoryReader(tc.dir, tc.pid, tc.pageSize)
			if err != nil && !strings.Contains(err.Error(), tc.expectedError.Error()) {
				t.Errorf("Expected error: %v, got error: %v", tc.expectedError, err)
			}

			if mr == nil && tc.expectedError == nil {
				t.Errorf("MemoryReader creation failed for checkpoint directory: %s and pid: %d", tc.dir, tc.pid)
			}
		})
	}
}

// TestNewMemoryReader test GetMempages method of MemoryReader.
func TestGetMemPages(t *testing.T) {
	type testcase struct {
		name          string
		mr            *MemoryReader
		start         uint64
		end           uint64
		expectedError error
	}

	pid, err := getTestImgPID()
	if err != nil {
		t.Fatal(err)
	}

	// Create a temporary empty memory pages file for testing
	tmpFilePath := filepath.Join(os.TempDir(), "pages-0.img")
	tmpFile, err := os.Create(tmpFilePath)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		_ = tmpFile.Close()
		if err := os.Remove(tmpFile.Name()); err != nil {
			t.Fatal(err)
		}
	}()

	mr, err := NewMemoryReader(testImgsDir, pid, sysPageSize)
	if err != nil {
		t.Fatal(err)
	}

	testCases := []testcase{
		{
			name: "Zero memory area size",
			mr: &MemoryReader{
				checkpointDir:  testImgsDir,
				pid:            pid,
				pageSize:       sysPageSize,
				pagesID:        mr.pagesID,
				pagemapEntries: mr.GetPagemapEntries(),
			},
			start:         0,
			end:           0,
			expectedError: nil,
		},
		{
			name: "Valid pagemap entry 1",
			mr: &MemoryReader{
				checkpointDir:  testImgsDir,
				pid:            pid,
				pageSize:       sysPageSize,
				pagesID:        mr.pagesID,
				pagemapEntries: mr.GetPagemapEntries(),
			},
			start:         mr.pagemapEntries[0].GetVaddr(),
			end:           mr.pagemapEntries[0].GetVaddr() + uint64(sysPageSize)*mr.pagemapEntries[0].GetNrPages(),
			expectedError: nil,
		},
		{
			name: "Valid pagemap entry 2",
			mr: &MemoryReader{
				checkpointDir:  testImgsDir,
				pid:            pid,
				pageSize:       sysPageSize,
				pagesID:        mr.pagesID,
				pagemapEntries: mr.GetPagemapEntries(),
			},
			start:         mr.pagemapEntries[1].GetVaddr(),
			end:           mr.pagemapEntries[1].GetVaddr() + uint64(sysPageSize)*mr.pagemapEntries[1].GetNrPages(),
			expectedError: nil,
		},
		{
			name: "Invalid pages file",
			mr: &MemoryReader{
				checkpointDir:  testImgsDir,
				pid:            pid,
				pageSize:       sysPageSize,
				pagesID:        mr.pagesID + 1,
				pagemapEntries: mr.GetPagemapEntries(),
			},
			start:         mr.pagemapEntries[0].GetVaddr(),
			end:           mr.pagemapEntries[0].GetVaddr() + uint64(sysPageSize)*mr.pagemapEntries[0].GetNrPages(),
			expectedError: errors.New("no such file or directory"),
		},
		{
			name: "Empty pages file",
			mr: &MemoryReader{
				checkpointDir:  os.TempDir(),
				pid:            pid,
				pageSize:       sysPageSize,
				pagesID:        0,
				pagemapEntries: mr.GetPagemapEntries(),
			},
			start:         mr.pagemapEntries[1].GetVaddr(),
			end:           mr.pagemapEntries[1].GetVaddr() + uint64(sysPageSize)*mr.pagemapEntries[1].GetNrPages(),
			expectedError: errors.New("EOF"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buff, err := tc.mr.GetMemPages(tc.start, tc.end)
			if err != nil && tc.expectedError != nil {
				if !strings.Contains(err.Error(), tc.expectedError.Error()) {
					t.Errorf("Expected error: %v, got error: %v", tc.expectedError, err)
				}
			}

			if tc.expectedError == nil && buff == nil {
				t.Errorf("Returned memory chunk is expected to be non-empty")
			}
		})
	}
}

func TestGetPsArgsAndEnvVars(t *testing.T) {
	pid, err := getTestImgPID()
	if err != nil {
		t.Fatal(err)
	}

	mr, err := NewMemoryReader(testImgsDir, pid, sysPageSize)
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		name          string
		expectedError error
		mr            *MemoryReader
	}{
		{
			name:          "wrong PID",
			expectedError: errors.New("no such file or directory"),
			mr: &MemoryReader{
				checkpointDir: testImgsDir,
				pid:           0,
			},
		},
		{
			name:          "valid arguments and environment variables",
			expectedError: nil,
			mr:            mr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			args, err := tc.mr.GetPsArgs()
			if err != nil && tc.expectedError != nil {
				if !strings.Contains(err.Error(), tc.expectedError.Error()) {
					t.Errorf("Expected error: %v, got error: %v", tc.expectedError, err)
				}
			}

			if tc.expectedError == nil && args == nil {
				t.Errorf("Expected non-nil arguments, got nil")
			}
		})

		t.Run(tc.name, func(t *testing.T) {
			envVars, err := tc.mr.GetPsEnvVars()
			if err != nil && tc.expectedError != nil {
				if !strings.Contains(err.Error(), tc.expectedError.Error()) {
					t.Errorf("Expected error: %v, got error: %v", tc.expectedError, err)
				}
			}

			if tc.expectedError == nil && envVars == nil {
				t.Errorf("Expected non-nil environment variables, got nil")
			}
		})
	}
}

func TestSearchPattern(t *testing.T) {
	pid, err := getTestImgPID()
	if err != nil {
		t.Fatal(err)
	}

	mr, err := NewMemoryReader(testImgsDir, pid, sysPageSize)
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		name                   string
		pattern                string
		context                int
		escapeRegExpCharacters bool
		shouldMatch            bool
		expectedError          error
	}{
		{
			name:        "PATH environment variable",
			pattern:     "PATH=",
			shouldMatch: true,
		},
		{
			name:        "PATH environment variable regex",
			pattern:     `\bPATH=([^\s]+)\b`,
			shouldMatch: true,
		},
		{
			name:        "PATH environment variable regex with 10 bytes context",
			pattern:     `\bPATH=([^\s]+)\b`,
			context:     10,
			shouldMatch: true,
		},
		{
			name:          "PATH environment variable regex with a negative context",
			pattern:       `\bPATH=([^\s]+)\b`,
			context:       -1,
			expectedError: errors.New("context size cannot be negative"),
		},
		{
			name:        "PATH environment variable regex with a large context",
			pattern:     `\bPATH=([^\s]+)\b`,
			context:     100000,
			shouldMatch: true,
		},
		{
			name:    "Non-existent pattern",
			pattern: "NON_EXISTENT_PATTERN",
		},
		{
			name:        "PASSWORD environment variable value as regex",
			pattern:     "123 Hello.*?",
			shouldMatch: true,
		},
		{
			name:                   "PASSWORD environment variable value with regex metacharacters to escape",
			pattern:                `123 Hello.*?[^]@WORLD(|x)`,
			escapeRegExpCharacters: true,
			shouldMatch:            true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			matches, err := mr.SearchPattern(tc.pattern, tc.escapeRegExpCharacters, tc.context, 0)
			if err != nil && tc.expectedError == nil {
				t.Errorf("Unexpected error for pattern %s: %v", tc.pattern, err)
			} else if err == nil && tc.expectedError != nil {
				t.Errorf("Expected error for pattern %s: %v", tc.pattern, tc.expectedError)
			}

			if tc.shouldMatch && len(matches) == 0 {
				t.Errorf("Expected to find a match for pattern \"%s\"", tc.pattern)
			} else if !tc.shouldMatch && len(matches) > 0 {
				t.Errorf("Expected not to find any match for pattern \"%s\"", tc.pattern)
			}

			for _, match := range matches {
				content, err := mr.GetMemPages(match.Vaddr, match.Vaddr+uint64(match.Length))
				if err != nil {
					t.Fatalf("Failed to get memory pages: %v", err)
				}

				buff := content.Bytes()
				for i := range buff {
					if buff[i] < 32 || buff[i] >= 127 {
						buff[i] = 0x3F
					}
				}

				if !strings.Contains(match.Match, content.String()) {
					t.Errorf("Expected to find %s in matched pattern %s", content.String(), match.Match)
				}
			}
		})
	}
}

func newSearchPatternMemoryReader(t *testing.T, memory string, truncateAt int) *MemoryReader {
	t.Helper()
	if len(memory) > searchPatternTestPageSize {
		t.Fatalf("memory has %d bytes, page size is %d", len(memory), searchPatternTestPageSize)
	}
	memory += strings.Repeat("x", searchPatternTestPageSize-len(memory))
	fileMemory := memory
	if truncateAt >= 0 {
		fileMemory = memory[:truncateAt]
	}

	const pagesID = 1
	checkpointDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(checkpointDir, "pages-1.img"), []byte(fileMemory), 0o600); err != nil {
		t.Fatal(err)
	}
	vaddr := uint64(0x1000)
	compatNrPages := uint32(1)
	nrPages := uint64(1)

	return &MemoryReader{
		checkpointDir: checkpointDir,
		pagesID:       pagesID,
		pageSize:      searchPatternTestPageSize,
		pagemapEntries: []*pagemap.PagemapEntry{
			{
				Vaddr:         &vaddr,
				CompatNrPages: &compatNrPages,
				NrPages:       &nrPages,
			},
		},
	}
}

func TestSearchPatternContextAcrossChunkBoundary(t *testing.T) {
	const (
		chunkSize = 8
		startAddr = 0x1000
	)
	testCases := []struct {
		name       string
		memory     string
		wantOffset uint64
	}{
		{name: "match starts at boundary", memory: "xxxxxxxxneedlexx", wantOffset: 8},
		{name: "match ends at boundary", memory: "xxneedlexxxxxxxx", wantOffset: 2},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mr := newSearchPatternMemoryReader(t, tc.memory, -1)
			matches, err := mr.SearchPattern("needle", true, 2, chunkSize)
			if err != nil {
				t.Fatal(err)
			}
			if len(matches) != 1 {
				t.Fatalf("got %d matches, want 1", len(matches))
			}
			if matches[0].Vaddr != startAddr+tc.wantOffset {
				t.Errorf("address = %#x, want %#x", matches[0].Vaddr, startAddr+tc.wantOffset)
			}
			if matches[0].Length != len("needle") {
				t.Errorf("length = %d, want %d", matches[0].Length, len("needle"))
			}
			if matches[0].Match != "xxneedlexx" {
				t.Errorf("match = %q, want %q", matches[0].Match, "xxneedlexx")
			}
		})
	}
}

func TestSearchPatternTruncatedContext(t *testing.T) {
	mr := newSearchPatternMemoryReader(t, "xxxxxxxx", 8)
	_, err := mr.SearchPattern("x", true, 8, 8)
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("got error %v, want %v", err, io.ErrUnexpectedEOF)
	}
}

func TestReadPatternMatchErrors(t *testing.T) {
	readErr := errors.New("read failed")
	testCases := []struct {
		name      string
		readerErr error
		wantErr   error
	}{
		{name: "EOF", readerErr: io.EOF, wantErr: io.ErrUnexpectedEOF},
		{name: "wrapped EOF", readerErr: fmt.Errorf("wrapped: %w", io.EOF), wantErr: io.ErrUnexpectedEOF},
		{name: "short read", wantErr: io.ErrUnexpectedEOF},
		{name: "other error", readerErr: readErr, wantErr: readErr},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reader := readerAtFunc(func(buff []byte, _ int64) (int, error) {
				buff[0] = 'x'
				return 1, tc.readerErr
			})
			_, err := readPatternMatch(reader, 0, 0x1000, 4, 1, 2, 2, 0, nil)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("got error %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestSearchLiteralAcrossChunkBoundary(t *testing.T) {
	const startAddr = 0x1000
	testCases := []struct {
		name        string
		memory      string
		pattern     string
		chunkSize   int
		context     int
		wantOffsets []uint64
		wantMatch   string
	}{
		{
			name:        "crosses boundary",
			memory:      "xxxxxxneedlexx",
			pattern:     "needle",
			chunkSize:   8,
			context:     2,
			wantOffsets: []uint64{6},
			wantMatch:   "xxneedlexx",
		},
		{
			name:        "longer than chunk",
			memory:      "xxlongneedlexx",
			pattern:     "longneedle",
			chunkSize:   4,
			wantOffsets: []uint64{2},
			wantMatch:   "longneedle",
		},
		{
			name:        "non-overlapping matches",
			memory:      "aaaa",
			pattern:     "aa",
			chunkSize:   1,
			wantOffsets: []uint64{0, 2},
			wantMatch:   "aa",
		},
		{
			name:        "sanitized boundary",
			memory:      "xxxxxxa\x00bxx",
			pattern:     "a?b",
			chunkSize:   8,
			wantOffsets: []uint64{6},
			wantMatch:   "a?b",
		},
		{
			name:        "prefix fallback across boundaries",
			memory:      "xxabababacaxx",
			pattern:     "ababaca",
			chunkSize:   3,
			wantOffsets: []uint64{4},
			wantMatch:   "ababaca",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mr := newSearchPatternMemoryReader(t, tc.memory, -1)
			matches, err := mr.SearchPattern(tc.pattern, true, tc.context, tc.chunkSize)
			if err != nil {
				t.Fatal(err)
			}
			if len(matches) != len(tc.wantOffsets) {
				t.Fatalf("got %d matches, want %d", len(matches), len(tc.wantOffsets))
			}
			for i, match := range matches {
				if match.Vaddr != startAddr+tc.wantOffsets[i] {
					t.Errorf("match %d address = %#x, want %#x", i, match.Vaddr, startAddr+tc.wantOffsets[i])
				}
				if match.Length != len(tc.pattern) {
					t.Errorf("match %d length = %d, want %d", i, match.Length, len(tc.pattern))
				}
				if match.Match != tc.wantMatch {
					t.Errorf("match %d = %q, want %q", i, match.Match, tc.wantMatch)
				}
			}
		})
	}
}

func TestSearchLiteralTruncatedInput(t *testing.T) {
	testCases := []struct {
		name    string
		pattern string
	}{
		{name: "partial search window", pattern: "needle"},
		{name: "at chunk boundary", pattern: "z"},
		{name: "literal longer than entry", pattern: strings.Repeat("x", searchPatternTestPageSize+1)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mr := newSearchPatternMemoryReader(t, "xxxxxxxx", 8)
			_, err := mr.SearchPattern(tc.pattern, true, 0, 8)
			if !errors.Is(err, io.ErrUnexpectedEOF) {
				t.Fatalf("got error %v, want %v", err, io.ErrUnexpectedEOF)
			}
		})
	}
}

func TestSearchLiteralReadsInputOnce(t *testing.T) {
	const (
		entrySize = 4096
		chunkSize = 16
	)
	reader := &countingReaderAt{data: bytes.Repeat([]byte{'a'}, entrySize)}
	literalPattern := strings.Repeat("a", 256) + "b"
	patternTable := literalFailureTable(literalPattern)

	matches, err := searchLiteralPattern(
		reader,
		literalPattern,
		patternTable,
		0,
		0x1000,
		entrySize,
		0,
		chunkSize,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 0 {
		t.Fatalf("got %d matches, want 0", len(matches))
	}
	if reader.bytesRead != entrySize {
		t.Fatalf("read %d bytes, want %d", reader.bytesRead, entrySize)
	}
	for i, offset := range reader.readStarts {
		wantOffset := int64(i * chunkSize)
		if offset != wantOffset {
			t.Fatalf("read %d starts at %d, want %d", i, offset, wantOffset)
		}
		wantSize := min(chunkSize, entrySize-i*chunkSize)
		if reader.readSizes[i] != wantSize {
			t.Fatalf("read %d has size %d, want %d", i, reader.readSizes[i], wantSize)
		}
	}
}

func BenchmarkSearchLiteralPattern(b *testing.B) {
	const (
		entrySize = 1 << 20
		chunkSize = 256
	)
	memory := bytes.Repeat([]byte{'a'}, entrySize)
	literalPattern := strings.Repeat("a", 4096) + "b"
	patternTable := literalFailureTable(literalPattern)
	reader := bytes.NewReader(memory)
	b.ReportAllocs()
	b.SetBytes(entrySize)

	for b.Loop() {
		matches, err := searchLiteralPattern(
			reader,
			literalPattern,
			patternTable,
			0,
			0x1000,
			entrySize,
			0,
			chunkSize,
		)
		if err != nil {
			b.Fatal(err)
		}
		if len(matches) != 0 {
			b.Fatalf("got %d matches, want 0", len(matches))
		}
	}
}

func TestMemoryRuneReader(t *testing.T) {
	t.Run("buffers and sanitizes bytes", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "pages.img")
		if err := os.WriteFile(path, []byte{'s', 'k', 'i', 'p', 'a', 'b', 0, 0xff}, 0o600); err != nil {
			t.Fatal(err)
		}
		file, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = file.Close() }()

		reader := newMemoryRuneReader(file, 2)
		reader.setEntry(4, 4)
		var got []rune
		for {
			r, size, err := reader.ReadRune()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
			if size != 1 {
				t.Fatalf("rune size = %d, want 1", size)
			}
			got = append(got, r)
		}
		if string(got) != "ab??" {
			t.Fatalf("got %q, want %q", got, "ab??")
		}

		reader.reset(1)
		r, size, err := reader.ReadRune()
		if err != nil {
			t.Fatal(err)
		}
		if r != 'b' || size != 1 {
			t.Fatalf("got (%q, %d), want (%q, 1)", r, size, 'b')
		}
	})

	t.Run("rejects truncated entry", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "pages.img")
		if err := os.WriteFile(path, []byte("abc"), 0o600); err != nil {
			t.Fatal(err)
		}
		file, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = file.Close() }()

		reader := newMemoryRuneReader(file, 2)
		reader.setEntry(0, 4)
		if _, _, err := reader.ReadRune(); err != nil {
			t.Fatal(err)
		}
		if _, _, err := reader.ReadRune(); err != nil {
			t.Fatal(err)
		}
		if _, _, err := reader.ReadRune(); !errors.Is(err, io.ErrUnexpectedEOF) {
			t.Fatalf("got error %v, want %v", err, io.ErrUnexpectedEOF)
		}
	})

	t.Run("propagates read errors", func(t *testing.T) {
		readErr := errors.New("read failed")
		testCases := []struct {
			name      string
			readerErr error
			wantErr   error
		}{
			{name: "EOF", readerErr: io.EOF, wantErr: io.ErrUnexpectedEOF},
			{name: "wrapped EOF", readerErr: fmt.Errorf("wrapped: %w", io.EOF), wantErr: io.ErrUnexpectedEOF},
			{name: "short read", wantErr: io.ErrUnexpectedEOF},
			{name: "other error", readerErr: readErr, wantErr: readErr},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				reader := newMemoryRuneReader(readerAtFunc(func(buff []byte, _ int64) (int, error) {
					buff[0] = 'x'
					return 1, tc.readerErr
				}), 2)
				reader.setEntry(0, 2)
				if _, _, err := reader.ReadRune(); !errors.Is(err, tc.wantErr) {
					t.Fatalf("got error %v, want %v", err, tc.wantErr)
				}
			})
		}
	})
}

func TestSearchPatternStream(t *testing.T) {
	const startAddr = 0x1000
	testCases := []struct {
		name       string
		memory     string
		pattern    string
		chunkSize  int
		context    int
		truncateAt int
		wantError  error
	}{
		{name: "bounded", memory: "xxxxxxneedlexx", pattern: `n[e]{2}dle`, chunkSize: 8, truncateAt: -1},
		{
			name:       "unbounded",
			memory:     "xxxxxxbegin" + strings.Repeat("x", 20) + "endxx",
			pattern:    `begin.*end`,
			chunkSize:  8,
			context:    2,
			truncateAt: -1,
		},
		{name: "multiple", memory: "begin1endxxbegin2end", pattern: `begin.*?end`, chunkSize: 4, truncateAt: -1},
		{name: "alternation", memory: "xxabax", pattern: `(?:ab|a)`, chunkSize: 2, truncateAt: -1},
		{name: "assertion", memory: "xxxxxx needle xx", pattern: `\bneedle\b`, chunkSize: 8, truncateAt: -1},
		{name: "empty", memory: "axxxxxxx", pattern: `a?`, chunkSize: 1, truncateAt: -1},
		{name: "sanitized", memory: "xxa\x00bxx", pattern: `a[?]b`, chunkSize: 2, truncateAt: -1},
		{
			name:       "truncated",
			memory:     "xxxxxxxx",
			pattern:    `missing.*pattern`,
			chunkSize:  8,
			truncateAt: 8,
			wantError:  io.ErrUnexpectedEOF,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mr := newSearchPatternMemoryReader(t, tc.memory, tc.truncateAt)
			memory := tc.memory + strings.Repeat("x", searchPatternTestPageSize-len(tc.memory))
			file, err := os.Open(filepath.Join(mr.checkpointDir, "pages-1.img"))
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = file.Close() }()
			patterns, err := compileStreamingRegexps(tc.pattern)
			if err != nil {
				t.Fatal(err)
			}
			reader := newMemoryRuneReader(file, tc.chunkSize)
			matches, err := searchPatternStream(file, patterns, reader, 0, startAddr, searchPatternTestPageSize, tc.context)
			if tc.wantError != nil {
				if !errors.Is(err, tc.wantError) {
					t.Fatalf("got error %v, want %v", err, tc.wantError)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}

			transformed := []byte(memory)
			for i := range transformed {
				if transformed[i] < 32 || transformed[i] >= 127 {
					transformed[i] = '?'
				}
			}
			want := regexp.MustCompile(tc.pattern).FindAllIndex(transformed, -1)
			if len(matches) != len(want) {
				t.Fatalf("got %d matches, want %d", len(matches), len(want))
			}
			for i, index := range want {
				contextStart := max(index[0]-tc.context, 0)
				contextEnd := min(index[1]+tc.context, len(transformed))
				if matches[i].Vaddr != startAddr+uint64(index[0]) {
					t.Errorf("match %d address = %#x, want %#x", i, matches[i].Vaddr, startAddr+uint64(index[0]))
				}
				if matches[i].Length != index[1]-index[0] {
					t.Errorf("match %d length = %d, want %d", i, matches[i].Length, index[1]-index[0])
				}
				if matches[i].Match != string(transformed[contextStart:contextEnd]) {
					t.Errorf("match %d = %q, want %q", i, matches[i].Match, transformed[contextStart:contextEnd])
				}
			}
		})
	}
}

func runSearchPatternStream(memory []byte, pattern string, chunkSize, context int) ([]PatternMatch, error) {
	const (
		initialOffset = 3
		startAddr     = 0x2000
	)
	backing := make([]byte, initialOffset+len(memory))
	copy(backing, "pad")
	copy(backing[initialOffset:], memory)
	readerAt := bytes.NewReader(backing)
	patterns, err := compileStreamingRegexps(pattern)
	if err != nil {
		return nil, err
	}
	reader := newMemoryRuneReader(readerAt, chunkSize)
	return searchPatternStream(
		readerAt,
		patterns,
		reader,
		initialOffset,
		startAddr,
		uint64(len(memory)),
		context,
	)
}

func expectedSearchPatternMatches(memory []byte, pattern string, context int) []PatternMatch {
	const startAddr = 0x2000
	transformed := append([]byte(nil), memory...)
	for i := range transformed {
		if transformed[i] < 32 || transformed[i] >= 127 {
			transformed[i] = '?'
		}
	}

	indexes := regexp.MustCompile(pattern).FindAllIndex(transformed, -1)
	matches := make([]PatternMatch, 0, len(indexes))
	for _, index := range indexes {
		contextStart := max(index[0]-context, 0)
		contextEnd := min(index[1]+context, len(transformed))
		matches = append(matches, PatternMatch{
			Vaddr:   startAddr + uint64(index[0]),
			Length:  index[1] - index[0],
			Context: context,
			Match:   string(transformed[contextStart:contextEnd]),
		})
	}
	return matches
}

func searchPatternStreamChunkSizes(memoryLength int, matches []PatternMatch) []int {
	candidates := []int{1, 2, memoryLength + 1}
	for _, match := range matches {
		if match.Length > 1 {
			candidates = append(candidates, match.Length-1)
		}
		if match.Length > 0 {
			candidates = append(candidates, match.Length, match.Length+1)
		}
	}

	seen := make(map[int]struct{}, len(candidates))
	chunkSizes := make([]int, 0, len(candidates))
	for _, chunkSize := range candidates {
		if chunkSize <= 0 {
			continue
		}
		if _, ok := seen[chunkSize]; ok {
			continue
		}
		seen[chunkSize] = struct{}{}
		chunkSizes = append(chunkSizes, chunkSize)
	}
	return chunkSizes
}

func TestSearchPatternStreamDifferential(t *testing.T) {
	testCases := []struct {
		name    string
		memory  []byte
		pattern string
		context int
	}{
		{name: "start anchor", memory: []byte("abc"), pattern: `^a`, context: 1},
		{name: "end anchor", memory: []byte("abc"), pattern: `c$`, context: 1},
		{name: "absolute start", memory: []byte("abc"), pattern: `\Aa`},
		{name: "absolute end", memory: []byte("abc"), pattern: `c\z`},
		{name: "continuation start anchor", memory: []byte("ab"), pattern: `a|^b`},
		{name: "continuation absolute start", memory: []byte("ab"), pattern: `a|\Ab`},
		{name: "continuation word boundary", memory: []byte("ab"), pattern: `a|\bb`},
		{name: "continuation non-word boundary", memory: []byte("ab"), pattern: `a|\Bb`},
		{name: "continuation end anchor", memory: []byte("ab"), pattern: `a|b$`},
		{name: "continuation absolute end", memory: []byte("ab"), pattern: `a|b\z`},
		{name: "word boundaries", memory: []byte(" a "), pattern: `\ba\b`},
		{name: "shared prefix long first", memory: []byte("ab"), pattern: `(?:ab|a)`},
		{name: "shared prefix short first", memory: []byte("ab"), pattern: `(?:a|ab)`},
		{name: "greedy", memory: []byte("a1b2b"), pattern: `a.*b`},
		{name: "non-greedy", memory: []byte("a1b2b"), pattern: `a.*?b`},
		{name: "captures", memory: []byte("xxabbx"), pattern: `(a(b)?)`, context: 1},
		{name: "named capture", memory: []byte("xxaaax"), pattern: `(?P<word>a+)`},
		{name: "inline flags", memory: []byte("xxAaax"), pattern: `(?i:a+)`},
		{name: "empty repetition", memory: []byte("baa"), pattern: `a*`},
		{name: "empty expression", memory: []byte("ab"), pattern: `(?:)`},
		{name: "empty record", pattern: `(?:)`},
		{name: "empty record start", pattern: `^`},
		{name: "empty record end", pattern: `$`},
		{name: "no match", memory: []byte("abc"), pattern: `z+`},
		{name: "sanitized NUL", memory: []byte{'a', 0, 'b'}, pattern: `a[?]b`, context: 2},
		{name: "sanitized invalid UTF-8", memory: []byte{0xff, 'a'}, pattern: `[?]a`},
		{name: "sanitized non-ASCII", memory: []byte("é"), pattern: `[?]+`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			want := expectedSearchPatternMatches(tc.memory, tc.pattern, tc.context)
			for _, chunkSize := range searchPatternStreamChunkSizes(len(tc.memory), want) {
				t.Run(fmt.Sprintf("chunk_%d", chunkSize), func(t *testing.T) {
					got, err := runSearchPatternStream(tc.memory, tc.pattern, chunkSize, tc.context)
					if err != nil {
						t.Fatal(err)
					}
					if !slices.Equal(got, want) {
						t.Fatalf("memory %q pattern %q: got %#v, want %#v", tc.memory, tc.pattern, got, want)
					}
				})
			}
		})
	}
}

func FuzzSearchPatternStream(f *testing.F) {
	patterns := []string{
		`a+`, `a*`, `(?:)`, `^`, `$`, `\Aa`, `a\z`, `\ba\b`, `\B`,
		`a.*b`, `a.*?b`, `(?:ab|a)`, `(?:a|ab)`, `(a(b)?)`,
		`(?P<word>a+)`, `(?i:a+)`, `[?]{1,4}`,
	}
	for i := range patterns {
		f.Add([]byte("ab\x00é"), uint8(i), uint8(i+1), uint8(i))
	}

	f.Fuzz(func(t *testing.T, memory []byte, patternIndex, chunkSeed, contextSeed uint8) {
		if len(memory) > 256 {
			memory = memory[:256]
		}
		pattern := patterns[int(patternIndex)%len(patterns)]
		chunkSize := int(chunkSeed%32) + 1
		context := int(contextSeed % 33)

		got, err := runSearchPatternStream(memory, pattern, chunkSize, context)
		if err != nil {
			t.Fatal(err)
		}
		want := expectedSearchPatternMatches(memory, pattern, context)
		if !slices.Equal(got, want) {
			t.Fatalf(
				"memory %q pattern %q chunk %d context %d: got %#v, want %#v",
				memory,
				pattern,
				chunkSize,
				context,
				got,
				want,
			)
		}
	})
}

// helper function to get the PID from the test-imgs directory
func getTestImgPID() (uint32, error) {
	psTreeImg, err := getImg(filepath.Join(testImgsDir, "pstree.img"), &pstree.PstreeEntry{})
	if err != nil {
		return 0, err
	}
	process := psTreeImg.Entries[0].Message.(*pstree.PstreeEntry)

	return process.GetPgid(), nil
}
