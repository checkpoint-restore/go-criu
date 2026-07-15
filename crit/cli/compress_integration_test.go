//go:build linux

package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/checkpoint-restore/go-criu/v8/crit"
)

func TestCompressCommandPrintsProgress(t *testing.T) {
	directory := t.TempDir()
	cliTestWriteCheckpoint(t, directory)
	oldInPlace, oldAcceleration := compressInPlace, compressAcceleration
	t.Cleanup(func() {
		compressInPlace, compressAcceleration = oldInPlace, oldAcceleration
	})
	compressInPlace, compressAcceleration = true, 1

	output, err := cliTestRunCommand(t, compressCmd, directory)
	if err != nil {
		t.Fatal(err)
	}
	prefix := fmt.Sprintf("Compressing checkpoint in %s\n  pagemap-1.img: 1 pages (%dK -> ",
		directory, os.Getpagesize()/1024)
	if !strings.HasPrefix(output, prefix) || !strings.Contains(output, "% saved)\n") ||
		!strings.HasSuffix(output, "Done\n") {
		t.Fatalf("compress output = %q", output)
	}
}

func TestCompressCommandPrintsNoOpState(t *testing.T) {
	tests := []struct {
		name string
		mode crit.CompressionMode
		want string
	}{
		{name: "empty", mode: crit.CompressionOff, want: "No pagemap files found in %s\n"},
		{name: "already compressed", mode: crit.CompressionPerPage, want: "Checkpoint in %s is already compressed\n"},
	}
	oldInPlace, oldAcceleration := compressInPlace, compressAcceleration
	t.Cleanup(func() {
		compressInPlace, compressAcceleration = oldInPlace, oldAcceleration
	})
	compressInPlace, compressAcceleration = true, 1

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			directory := t.TempDir()
			cliTestWriteInventory(t, directory, test.mode)
			output, err := cliTestRunCommand(t, compressCmd, directory)
			if err != nil {
				t.Fatal(err)
			}
			if want := fmt.Sprintf(test.want, directory); output != want {
				t.Fatalf("output = %q, want %q", output, want)
			}
		})
	}
}

func TestCompressCommandWrapsFilesystemErrors(t *testing.T) {
	oldAcceleration := compressAcceleration
	compressAcceleration = 1
	t.Cleanup(func() { compressAcceleration = oldAcceleration })
	missing := filepath.Join(t.TempDir(), "missing")
	output, err := cliTestRunCommand(t, compressCmd, missing)
	if err == nil || !errors.Is(err, os.ErrNotExist) ||
		!strings.Contains(err.Error(), "failed to compress checkpoint") {
		t.Fatalf("RunE() error = %v", err)
	}
	if output != "" {
		t.Fatalf("output after error = %q", output)
	}
}
