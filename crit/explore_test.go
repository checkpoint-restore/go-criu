package crit

import (
	"path/filepath"
	"strings"
	"testing"

	criu_core "github.com/checkpoint-restore/go-criu/v8/crit/images/criu-core"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/fs"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/pstree"
)

func TestGetFirstEntryEmptyImage(t *testing.T) {
	path := filepath.Join(t.TempDir(), "core-1.img")
	writeTestImage(t, path, &CriuImage{Magic: "CORE"})

	_, err := getFirstEntry(path, &criu_core.CoreEntry{})
	if err == nil || !strings.Contains(err.Error(), "has no entries") {
		t.Fatalf("expected no entries error, got %v", err)
	}
}

func TestGetFirstEntry(t *testing.T) {
	path := filepath.Join(t.TempDir(), "fs-1.img")
	writeTestImage(t, path, &CriuImage{
		Magic: "FS",
		Entries: []*CriuEntry{{Message: &fs.FsEntry{
			CwdId:  uint32Pointer(3),
			RootId: uint32Pointer(4),
		}}},
	})

	entry, err := getFirstEntry(path, &fs.FsEntry{})
	if err != nil {
		t.Fatal(err)
	}
	if entry.GetCwdId() != 3 || entry.GetRootId() != 4 {
		t.Fatalf("unexpected entry: %v", entry)
	}
}

// Every explore command must return an error, not panic,
// when a task's core image has no entries.
func TestExploreEmptyImage(t *testing.T) {
	dir := t.TempDir()
	writeTestImage(t, filepath.Join(dir, "pstree.img"), &CriuImage{
		Magic: "PSTREE",
		Entries: []*CriuEntry{{Message: &pstree.PstreeEntry{
			Pid:  uint32Pointer(1),
			Ppid: uint32Pointer(0),
			Pgid: uint32Pointer(1),
			Sid:  uint32Pointer(1),
		}}},
	})
	writeTestImage(t, filepath.Join(dir, "core-1.img"), &CriuImage{Magic: "CORE"})

	c := &crit{inputDirPath: dir}
	explorers := map[string]func() error{
		"ps":   func() error { _, err := c.ExplorePs(); return err },
		"fds":  func() error { _, err := c.ExploreFds(); return err },
		"mems": func() error { _, err := c.ExploreMems(); return err },
		"rss":  func() error { _, err := c.ExploreRss(); return err },
		"sk":   func() error { _, err := c.ExploreSk(); return err },
	}
	for name, explore := range explorers {
		t.Run(name, func(t *testing.T) {
			err := explore()
			if err == nil || !strings.Contains(err.Error(), "has no entries") {
				t.Fatalf("expected no entries error, got %v", err)
			}
		})
	}
}

func TestMemoryReaderEmptyMmImage(t *testing.T) {
	dir := t.TempDir()
	writeTestImage(t, filepath.Join(dir, "mm-1.img"), &CriuImage{Magic: "MM"})

	mr := &MemoryReader{checkpointDir: dir, pid: 1}
	if _, err := mr.GetPsArgs(); err == nil {
		t.Error("GetPsArgs: expected error for empty mm image")
	}
	if _, err := mr.GetPsEnvVars(); err == nil {
		t.Error("GetPsEnvVars: expected error for empty mm image")
	}
	if _, err := mr.GetShmemSize(); err == nil {
		t.Error("GetShmemSize: expected error for empty mm image")
	}
}

func TestGetStatsEmptyImage(t *testing.T) {
	path := filepath.Join(t.TempDir(), StatsDump)
	writeTestImage(t, path, &CriuImage{Magic: "STATS"})

	if _, err := getStats(path); err == nil {
		t.Fatal("expected error for empty stats image")
	}
}
