package cli

import (
	"strings"
	"testing"
)

func TestDecompressCommandRequiresCheckpointDirectory(t *testing.T) {
	for _, args := range [][]string{nil, {"one", "two"}} {
		err := decompressCmd.Args(decompressCmd, args)
		if err == nil || !strings.Contains(err.Error(), "accepts 1 arg(s)") {
			t.Errorf("decompress Args(%q) error = %v", args, err)
		}
	}
}
