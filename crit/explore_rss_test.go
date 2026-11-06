package crit

import (
	"fmt"
	"path/filepath"
	"testing"

	criu_core "github.com/checkpoint-restore/go-criu/v8/crit/images/criu-core"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/mm"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/pstree"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/vma"
)

// A pagemap entry that covers the last VMA must not index past the VMA list.
func TestExploreRssPagemapCoversLastVma(t *testing.T) {
	dir := t.TempDir()
	const pid uint32 = 1

	writeTestImage(t, filepath.Join(dir, "pstree.img"), &CriuImage{
		Magic: "PSTREE",
		Entries: []*CriuEntry{{Message: &pstree.PstreeEntry{
			Pid:  uint32Pointer(pid),
			Ppid: uint32Pointer(0),
			Pgid: uint32Pointer(pid),
			Sid:  uint32Pointer(pid),
		}}},
	})
	comm := "test"
	writeTestImage(t, filepath.Join(dir, fmt.Sprintf("core-%d.img", pid)), &CriuImage{
		Magic: "CORE",
		Entries: []*CriuEntry{{Message: &criu_core.CoreEntry{
			Mtype: criu_core.CoreEntry_X86_64.Enum(),
			Tc: &criu_core.TaskCoreEntry{
				TaskState:   uint32Pointer(uint32(TaskAlive)),
				ExitCode:    uint32Pointer(0),
				Personality: uint32Pointer(0),
				Flags:       uint32Pointer(0),
				BlkSigset:   uint64Pointer(0),
				Comm:        &comm,
			},
		}}},
	})
	fd := int64(-1)
	vmaEntry := func(start, end uint64) *vma.VmaEntry {
		return &vma.VmaEntry{
			Start:  uint64Pointer(start),
			End:    uint64Pointer(end),
			Pgoff:  uint64Pointer(0),
			Shmid:  uint64Pointer(0),
			Prot:   uint32Pointer(0),
			Flags:  uint32Pointer(0),
			Status: uint32Pointer(0),
			Fd:     &fd,
		}
	}
	writeTestImage(t, filepath.Join(dir, fmt.Sprintf("mm-%d.img", pid)), &CriuImage{
		Magic: "MM",
		Entries: []*CriuEntry{{Message: &mm.MmEntry{
			MmStartCode:  uint64Pointer(0),
			MmEndCode:    uint64Pointer(0),
			MmStartData:  uint64Pointer(0),
			MmEndData:    uint64Pointer(0),
			MmStartStack: uint64Pointer(0),
			MmStartBrk:   uint64Pointer(0),
			MmBrk:        uint64Pointer(0),
			MmArgStart:   uint64Pointer(0),
			MmArgEnd:     uint64Pointer(0),
			MmEnvStart:   uint64Pointer(0),
			MmEnvEnd:     uint64Pointer(0),
			ExeFileId:    uint32Pointer(0),
			Vmas: []*vma.VmaEntry{
				vmaEntry(0x1000, 0x3000),
				// Last VMA, fully covered by the last pagemap entry below
				vmaEntry(0x7000, 0x9000),
			},
		}}},
	})
	writeTestPagemap(t, dir, pid, 1,
		presentTestEntry(0x1000, 1),
		presentTestEntry(0x7000, 2),
	)

	c := &crit{inputDirPath: dir}
	var rssMaps []*RssMap
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("ExploreRss panicked: %v", r)
			}
		}()
		rssMaps, err = c.ExploreRss()
	}()
	if err != nil {
		t.Fatal(err)
	}
	if len(rssMaps) != 1 || len(rssMaps[0].Rsses) != 2 {
		t.Fatalf("unexpected result: %+v", rssMaps)
	}
	last := rssMaps[0].Rsses[1]
	if len(last.Vmas) != 1 || last.Vmas[0].Addr != "7000" || last.Vmas[0].Pages != 2 {
		t.Fatalf("expected last entry to map VMA 7000 with 2 pages, got %+v", last.Vmas)
	}
}
