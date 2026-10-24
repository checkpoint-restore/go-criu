package cli

import (
	"encoding/binary"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/checkpoint-restore/go-criu/v8/crit/images/pidfd"
	"github.com/checkpoint-restore/go-criu/v8/magic"
)

func TestGetEntryTypeFromPIDFDImage(t *testing.T) {
	magicMap := magic.LoadMagic()
	header := make([]byte, 8)
	binary.LittleEndian.PutUint32(header[:4], uint32(magicMap.ByName["IMG_COMMON"]))
	binary.LittleEndian.PutUint32(header[4:], uint32(magicMap.ByName["PIDFD"]))

	path := filepath.Join(t.TempDir(), "pidfd.img")
	if err := os.WriteFile(path, header, 0o600); err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = file.Close() }()

	entry, err := GetEntryTypeFromImg(file)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := entry.(*pidfd.PidfdEntry); !ok {
		t.Fatalf("GetEntryTypeFromImg() returned %T, want *pidfd.PidfdEntry", entry)
	}
	if offset, err := file.Seek(0, io.SeekCurrent); err != nil || offset != 0 {
		t.Fatalf("image offset = %d, %v; want 0", offset, err)
	}
}

func TestGetEntryTypeFromJSON(t *testing.T) {
	testCases := []struct {
		name        string
		jsonContent string
		wantErr     bool
		errContains string
		validate    func(t *testing.T, entry any)
	}{
		{
			name:        "valid magic",
			jsonContent: `{"magic": "PIDFD"}`,
			wantErr:     false,
			validate: func(t *testing.T, entry any) {
				if _, ok := entry.(*pidfd.PidfdEntry); !ok {
					t.Fatalf("returned %T, want *pidfd.PidfdEntry", entry)
				}
			},
		},
		{
			name:        "missing magic field",
			jsonContent: `{"entries": []}`,
			wantErr:     true,
			errContains: "missing or invalid 'magic' field",
		},
		{
			name:        "non-string magic field (number)",
			jsonContent: `{"magic": 12345}`,
			wantErr:     true,
			errContains: "missing or invalid 'magic' field",
		},
		{
			name:        "empty magic field",
			jsonContent: `{"magic": ""}`,
			wantErr:     true,
			errContains: "missing or invalid 'magic' field",
		},
		{
			name:        "unknown magic string",
			jsonContent: `{"magic": "NONEXISTENT_MAGIC"}`,
			wantErr:     true,
			errContains: `no protobuf binding found for magic "NONEXISTENT_MAGIC"`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "test.json")
			if err := os.WriteFile(path, []byte(tc.jsonContent), 0o600); err != nil {
				t.Fatal(err)
			}
			file, err := os.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = file.Close() }()

			entry, err := GetEntryTypeFromJSON(file)
			if (err != nil) != tc.wantErr {
				t.Fatalf("GetEntryTypeFromJSON() error = %v, wantErr %v", err, tc.wantErr)
			}
			if tc.wantErr {
				if tc.errContains != "" && (err == nil || !strings.Contains(err.Error(), tc.errContains)) {
					t.Fatalf("GetEntryTypeFromJSON() error = %v, want containing %q", err, tc.errContains)
				}
				return
			}
			if tc.validate != nil {
				tc.validate(t, entry)
			}
			if offset, err := file.Seek(0, io.SeekCurrent); err != nil || offset != 0 {
				t.Fatalf("file offset = %d, %v; want 0", offset, err)
			}
		})
	}
}
