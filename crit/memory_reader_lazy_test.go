package crit

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMemoryReaderLazyPages(t *testing.T) {
	const (
		pageSize = 4096
		pid      = 39
		base     = 0x50000
		lazyFlag = uint32(1 << 1)
	)
	page := bytes.Repeat([]byte{'A'}, pageSize)
	copy(page[91:], "lazy-page-needle")

	for _, test := range []struct {
		name    string
		flags   uint32
		payload []byte
		wantErr bool
	}{
		{name: "unavailable lazy page", flags: lazyFlag, wantErr: true},
		{name: "present lazy page", flags: pePresent | lazyFlag, payload: page},
		{name: "hole", flags: 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			directory := t.TempDir()
			writeTestInventory(t, directory, crtoolsImagesV1_1, CompressionOff, 0)
			entry := presentTestEntry(base, 1)
			entry.Flags = uint32Pointer(test.flags)
			writeTestPagemap(t, directory, pid, 1, entry)
			if err := os.WriteFile(filepath.Join(directory, "pages-1.img"), test.payload, 0o600); err != nil {
				t.Fatal(err)
			}
			reader, err := NewMemoryReader(directory, pid, pageSize)
			if err != nil {
				t.Fatal(err)
			}

			t.Run("read", func(t *testing.T) {
				memory, err := reader.GetMemPages(base, base+pageSize)
				if test.wantErr {
					if err == nil || !strings.Contains(err.Error(), "lazy page") {
						t.Fatalf("expected unavailable lazy page error, got %v", err)
					}
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				want := test.payload
				if want == nil {
					want = make([]byte, pageSize)
				}
				if !bytes.Equal(memory.Bytes(), want) {
					t.Fatal("memory contents do not match")
				}
			})

			for _, search := range []struct {
				name    string
				pattern string
				escape  bool
			}{
				{name: "literal search", pattern: "lazy-page-needle", escape: true},
				{name: "regexp search", pattern: "lazy-page-need.e"},
			} {
				t.Run(search.name, func(t *testing.T) {
					// Searches skip lazy pages instead of failing on them.
					matches, err := reader.SearchPattern(search.pattern, search.escape, 0, 64)
					if err != nil {
						t.Fatal(err)
					}
					if test.payload == nil {
						if len(matches) != 0 {
							t.Fatalf("unexpected matches without local pages: %+v", matches)
						}
					} else if len(matches) != 1 || matches[0].Vaddr != base+91 {
						t.Fatalf("unexpected matches in a present lazy page: %+v", matches)
					}
				})
			}
		})
	}
}

func TestSearchPatternSkipsLazyPages(t *testing.T) {
	const (
		pageSize = 4096
		pid      = 40
		base     = 0x60000
		lazyFlag = uint32(1 << 1)
	)
	page := bytes.Repeat([]byte{'A'}, pageSize)
	copy(page[17:], "present-page-needle")

	directory := t.TempDir()
	writeTestInventory(t, directory, crtoolsImagesV1_1, CompressionOff, 0)
	lazy := presentTestEntry(base, 1)
	lazy.Flags = uint32Pointer(lazyFlag)
	present := presentTestEntry(base+pageSize, 1)
	writeTestPagemap(t, directory, pid, 1, lazy, present)
	if err := os.WriteFile(filepath.Join(directory, "pages-1.img"), page, 0o600); err != nil {
		t.Fatal(err)
	}
	reader, err := NewMemoryReader(directory, pid, pageSize)
	if err != nil {
		t.Fatal(err)
	}

	matches, err := reader.SearchPattern("present-page-needle", true, 0, 64)
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 1 || matches[0].Vaddr != base+pageSize+17 {
		t.Fatalf("unexpected matches next to a lazy page: %+v", matches)
	}
}
