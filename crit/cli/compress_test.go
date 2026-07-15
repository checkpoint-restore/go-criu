package cli

import (
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/checkpoint-restore/go-criu/v8/crit"
)

var cliTestInitOnce sync.Once

func TestCompressionCommandsAreRegistered(t *testing.T) {
	cliTestInitOnce.Do(Init)
	compress, _, err := rootCmd.Find([]string{"compress"})
	if err != nil || compress != compressCmd {
		t.Fatalf("find compress command = %v, %v", compress, err)
	}
	if compress.Flags().Lookup("in-place") == nil || compress.Flags().Lookup("acceleration") == nil {
		t.Fatal("compress flags are not registered")
	}
	decompress, _, err := rootCmd.Find([]string{"decompress"})
	if err != nil || decompress != decompressCmd {
		t.Fatalf("find decompress command = %v, %v", decompress, err)
	}
	if decompress.Flags().Lookup("in-place") == nil {
		t.Fatal("decompress --in-place flag is not registered")
	}
}

func TestCompressCommandDocumentsIncrementalRestriction(t *testing.T) {
	for _, phrase := range []string{"newest layer", "parent"} {
		if !strings.Contains(compressCmd.Long, phrase) {
			t.Errorf("compress help does not mention %q", phrase)
		}
	}
}

func TestCompressCommandAcceptsOnlyAccelerationOne(t *testing.T) {
	original := compressAcceleration
	defer func() { compressAcceleration = original }()
	for _, acceleration := range []int{-1, 0, 2, 100} {
		compressAcceleration = acceleration
		err := compressCmd.RunE(compressCmd, []string{t.TempDir()})
		if !errors.Is(err, crit.ErrUnsupportedAcceleration) {
			t.Errorf("acceleration %d error = %v", acceleration, err)
		}
	}
}

func TestCompressCommandRequiresCheckpointDirectory(t *testing.T) {
	for _, args := range [][]string{nil, {"one", "two"}} {
		err := compressCmd.Args(compressCmd, args)
		if err == nil || !strings.Contains(err.Error(), "accepts 1 arg(s)") {
			t.Errorf("compress Args(%q) error = %v", args, err)
		}
	}
}
