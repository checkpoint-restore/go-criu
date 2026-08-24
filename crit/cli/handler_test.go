package cli

import (
	"encoding/binary"
	"io"
	"os"
	"path/filepath"
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
