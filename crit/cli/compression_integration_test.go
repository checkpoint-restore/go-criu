//go:build linux

package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/checkpoint-restore/go-criu/v8/crit"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/inventory"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/mm"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/pagemap"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
)

func cliTestRunCommand(t *testing.T, command *cobra.Command, directory string) (string, error) {
	t.Helper()
	var output bytes.Buffer
	command.SetOut(&output)
	command.SetContext(context.Background())
	t.Cleanup(func() { command.SetOut(nil) })
	err := command.RunE(command, []string{directory})
	return output.String(), err
}

func cliTestWriteCheckpoint(t *testing.T, directory string) {
	t.Helper()
	cliTestWriteInventory(t, directory, crit.CompressionOff)
	cliTestWriteImage(t, filepath.Join(directory, "pagemap-1.img"), "PAGEMAP",
		&pagemap.PagemapHead{PagesId: proto.Uint32(1)},
		&pagemap.PagemapEntry{
			Vaddr:         proto.Uint64(uint64(os.Getpagesize())),
			CompatNrPages: proto.Uint32(1),
			Flags:         proto.Uint32(1 << 2),
		})
	if err := os.WriteFile(filepath.Join(directory, "pages-1.img"),
		bytes.Repeat([]byte{'A'}, os.Getpagesize()), 0o600); err != nil {
		t.Fatal(err)
	}
	zero64 := func() *uint64 { return proto.Uint64(0) }
	cliTestWriteImage(t, filepath.Join(directory, "mm-1.img"), "MM", &mm.MmEntry{
		MmStartCode: zero64(), MmEndCode: zero64(), MmStartData: zero64(), MmEndData: zero64(),
		MmStartStack: zero64(), MmStartBrk: zero64(), MmBrk: zero64(), MmArgStart: zero64(),
		MmArgEnd: zero64(), MmEnvStart: zero64(), MmEnvEnd: zero64(), ExeFileId: proto.Uint32(0),
	})
}

func cliTestWriteInventory(t *testing.T, directory string, mode crit.CompressionMode) {
	t.Helper()
	version := uint32(2)
	entry := &inventory.InventoryEntry{ImgVersion: proto.Uint32(version)}
	if mode.Compressed() {
		entry.ImgVersion = proto.Uint32(3)
		entry.Compress = proto.Uint32(uint32(mode))
	}
	cliTestWriteImage(t, filepath.Join(directory, "inventory.img"), "INVENTORY", entry)
}

func cliTestWriteImage(t *testing.T, path, magic string, messages ...proto.Message) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	entries := make([]*crit.CriuEntry, 0, len(messages))
	for _, message := range messages {
		entries = append(entries, &crit.CriuEntry{Message: message})
	}
	image := &crit.CriuImage{Magic: magic, Entries: entries}
	if err := crit.New(nil, file, "", false, false).Encode(image); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}
