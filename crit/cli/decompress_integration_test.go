//go:build linux

package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/checkpoint-restore/go-criu/v8/crit"
)

func TestDecompressCommandPrintsProgress(t *testing.T) {
	directory := t.TempDir()
	cliTestWriteCheckpoint(t, directory)
	if _, err := crit.CompressCheckpoint(context.Background(), directory, crit.CompressOptions{
		InPlace: true, Acceleration: 1,
	}); err != nil {
		t.Fatal(err)
	}
	oldInPlace := decompressInPlace
	decompressInPlace = true
	t.Cleanup(func() { decompressInPlace = oldInPlace })

	output, err := cliTestRunCommand(t, decompressCmd, directory)
	if err != nil {
		t.Fatal(err)
	}
	prefix := fmt.Sprintf("Decompressing checkpoint in %s\n  pagemap-1.img: 1 pages (", directory)
	suffix := fmt.Sprintf("K -> %dK)\nDone\n", os.Getpagesize()/1024)
	if !strings.HasPrefix(output, prefix) || !strings.HasSuffix(output, suffix) {
		t.Fatalf("decompress output = %q", output)
	}
}

func TestDecompressCommandPrintsNoOpState(t *testing.T) {
	tests := []struct {
		name string
		mode crit.CompressionMode
		want string
	}{
		{name: "already decompressed", mode: crit.CompressionOff, want: "Checkpoint in %s is already decompressed\n"},
		{name: "empty compressed checkpoint", mode: crit.CompressionBlock, want: "No pagemap files found in %s\n"},
	}
	oldInPlace := decompressInPlace
	decompressInPlace = true
	t.Cleanup(func() { decompressInPlace = oldInPlace })

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			directory := t.TempDir()
			cliTestWriteInventory(t, directory, test.mode)
			output, err := cliTestRunCommand(t, decompressCmd, directory)
			if err != nil {
				t.Fatal(err)
			}
			if want := fmt.Sprintf(test.want, directory); output != want {
				t.Fatalf("output = %q, want %q", output, want)
			}
		})
	}
}

func TestDecompressCommandWrapsFilesystemErrors(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing")
	output, err := cliTestRunCommand(t, decompressCmd, missing)
	if err == nil || !errors.Is(err, os.ErrNotExist) ||
		!strings.Contains(err.Error(), "failed to decompress checkpoint") {
		t.Fatalf("RunE() error = %v", err)
	}
	if output != "" {
		t.Fatalf("output after error = %q", output)
	}
}
