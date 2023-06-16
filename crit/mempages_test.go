package crit

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/checkpoint-restore/go-criu/v6/crit/images/pstree"
)

const (
	testImgsDir = "test-imgs"
)

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
		tmpFile.Close()
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
			end:           mr.pagemapEntries[0].GetVaddr() + uint64(uint32(sysPageSize)*mr.pagemapEntries[0].GetNrPages()),
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
			end:           mr.pagemapEntries[1].GetVaddr() + uint64(uint32(sysPageSize)*mr.pagemapEntries[1].GetNrPages()),
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
			end:           mr.pagemapEntries[0].GetVaddr() + uint64(uint32(sysPageSize)*mr.pagemapEntries[0].GetNrPages()),
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
			end:           mr.pagemapEntries[1].GetVaddr() + uint64(uint32(sysPageSize)*mr.pagemapEntries[1].GetNrPages()),
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

// helper function to get the PID from the test-imgs directory
func getTestImgPID() (uint32, error) {
	psTreeImg, err := getImg(filepath.Join(testImgsDir, "pstree.img"), &pstree.PstreeEntry{})
	if err != nil {
		return 0, err
	}
	process := psTreeImg.Entries[0].Message.(*pstree.PstreeEntry)

	return process.GetPgid(), nil
}
