package crit

import (
	"fmt"
	"strconv"

	"github.com/checkpoint-restore/go-criu/v6/crit/images"
)

// PsTree represents the process tree
type PsTree struct {
	PId      uint32              `json:"pId"`
	PgId     uint32              `json:"pgId"`
	SId      uint32              `json:"sId"`
	Comm     string              `json:"comm"`
	Process  *images.PstreeEntry `json:"-"`
	Core     *images.CoreEntry   `json:"-"`
	Children []*PsTree           `json:"children,omitempty"`
}

// ExplorePs constructs the process tree and returns the root process
func (c *crit) ExplorePs() (*PsTree, error) {
	psTreeImg, err := getImg(fmt.Sprintf("%s/pstree.img", c.inputDirPath))
	if err != nil {
		return nil, err
	}

	processes := make(map[uint32]*PsTree)
	var psTreeRoot *PsTree
	for _, entry := range psTreeImg.Entries {
		process := entry.Message.(*images.PstreeEntry)
		pId := process.GetPid()

		coreImg, err := getImg(fmt.Sprintf("%s/core-%d.img", c.inputDirPath, pId))
		if err != nil {
			return nil, err
		}
		coreData := coreImg.Entries[0].Message.(*images.CoreEntry)

		ps := &PsTree{
			PId:     pId,
			PgId:    process.GetPgid(),
			SId:     process.GetSid(),
			Comm:    coreData.Tc.GetComm(),
			Process: process,
			Core:    coreData,
		}
		// If there is no parent process, then it is the root
		if process.GetPpid() == 0 {
			psTreeRoot = ps
		}
		processes[pId] = ps
	}

	for _, ps := range processes {
		parent := ps.Process.GetPpid()
		if parent != 0 {
			processes[parent].Children = append(processes[parent].Children, ps)
		}
	}

	return psTreeRoot, nil
}

// Fd represents the file descriptors opened in a single process
type Fd struct {
	PId   uint32  `json:"pId"`
	Files []*File `json:"files,omitempty"`
}

// File represents a single opened file
type File struct {
	Fd   string `json:"fd"`
	Path string `json:"path"`
}

// ExploreFds searches the process tree for open files
// and returns a list of PIDs with the corresponding files
func (c *crit) ExploreFds() ([]*Fd, error) {
	psTreeImg, err := getImg(fmt.Sprintf("%s/pstree.img", c.inputDirPath))
	if err != nil {
		return nil, err
	}

	fds := make([]*Fd, 0)
	for _, entry := range psTreeImg.Entries {
		process := entry.Message.(*images.PstreeEntry)
		pId := process.GetPid()
		// Get file with object IDs
		idsImg, err := getImg(fmt.Sprintf("%s/ids-%d.img", c.inputDirPath, pId))
		if err != nil {
			return nil, err
		}
		filesId := idsImg.Entries[0].Message.(*images.TaskKobjIdsEntry).GetFilesId()
		// Get open file descriptors
		fdInfoImg, err := getImg(fmt.Sprintf("%s/fdinfo-%d.img", c.inputDirPath, filesId))
		if err != nil {
			return nil, err
		}

		fdEntry := Fd{PId: pId}
		for _, fdInfoEntry := range fdInfoImg.Entries {
			fdInfo := fdInfoEntry.Message.(*images.FdinfoEntry)
			filePath, err := getFilePath(c.inputDirPath,
				fdInfo.GetId(), fdInfo.GetType())
			if err != nil {
				return nil, err
			}
			file := File{
				Fd:   strconv.FormatUint(uint64(fdInfo.GetFd()), 10),
				Path: filePath,
			}
			fdEntry.Files = append(fdEntry.Files, &file)
		}
		// Get chroot and chdir info
		fsImg, err := getImg(fmt.Sprintf("%s/fs-%d.img", c.inputDirPath, pId))
		if err != nil {
			return nil, err
		}
		fs := fsImg.Entries[0].Message.(*images.FsEntry)
		filePath, err := getFilePath(c.inputDirPath,
			fs.GetCwdId(), images.FdTypes_REG)
		if err != nil {
			return nil, err
		}
		fdEntry.Files = append(fdEntry.Files, &File{
			Fd:   "cwd",
			Path: filePath,
		})
		filePath, err = getFilePath(c.inputDirPath,
			fs.GetRootId(), images.FdTypes_REG)
		if err != nil {
			return nil, err
		}
		fdEntry.Files = append(fdEntry.Files, &File{
			Fd:   "root",
			Path: filePath,
		})

		fds = append(fds, &fdEntry)
	}

	return fds, nil
}

// MemMap represents the memory mapping of a single process
type MemMap struct {
	PId  uint32 `json:"pId"`
	Exe  string `json:"exe"`
	Mems []*Mem `json:"mems,omitempty"`
}

// Mem represents the memory mapping of a single file
type Mem struct {
	Start      string `json:"start"`
	End        string `json:"end"`
	Protection string `json:"protection"`
	Resource   string `json:"resource,omitempty"`
}

// ExploreMems traverses the process tree and returns a
// list of processes with the corresponding memory mapping
func (c *crit) ExploreMems() ([]*MemMap, error) {
	psTreeImg, err := getImg(fmt.Sprintf("%s/pstree.img", c.inputDirPath))
	if err != nil {
		return nil, err
	}

	vmaIdMap, vmaId := make(map[uint64]int), 0
	// Use a closure to handle the ID counter
	getVmaId := func(shmId uint64) int {
		if _, ok := vmaIdMap[shmId]; !ok {
			vmaIdMap[shmId] = vmaId
			vmaId++
		}
		return vmaIdMap[shmId]
	}

	memMaps := make([]*MemMap, 0)
	for _, entry := range psTreeImg.Entries {
		process := entry.Message.(*images.PstreeEntry)
		pId := process.GetPid()
		// Get memory mappings
		mmImg, err := getImg(fmt.Sprintf("%s/mm-%d.img", c.inputDirPath, pId))
		if err != nil {
			return nil, err
		}
		mmInfo := mmImg.Entries[0].Message.(*images.MmEntry)
		exePath, err := getFilePath(c.inputDirPath,
			mmInfo.GetExeFileId(), images.FdTypes_REG)
		if err != nil {
			return nil, err
		}

		memMap := MemMap{
			PId: pId,
			Exe: exePath,
		}
		for _, vma := range mmInfo.GetVmas() {
			mem := Mem{
				Start: strconv.FormatUint(vma.GetStart(), 16),
				End:   strconv.FormatUint(vma.GetEnd(), 16),
			}

			switch status := vma.GetStatus(); {
			// Pages used by a file
			case status&((1<<7)|(1<<6)) != 0:
				file, err := getFilePath(c.inputDirPath,
					uint32(vma.GetShmid()), images.FdTypes_REG)
				if err != nil {
					return nil, err
				}
				if vma.GetPgoff() != 0 {
					mem.Resource = fmt.Sprintf("%s + 0x%x", file, vma.GetPgoff())
				}
				if status&(1<<7) != 0 {
					mem.Resource = fmt.Sprint(mem.Resource, " (s)")
				}
			case status&(1<<3) != 0:
				mem.Resource = "[vdso]"
			case status&(1<<2) != 0:
				mem.Resource = "[vsyscall]"
			case status&(1<<1) != 0:
				mem.Resource = "[stack]"
			// 0x0100 (256) indicates that the page grows downwards
			case vma.GetFlags()&0x0100 != 0:
				mem.Resource = "[stack?]"
			case status&(1<<11) != 0:
				mem.Resource = fmt.Sprintf("packet[%d]", getVmaId(vma.GetShmid()))
			case status&(1<<10) != 0:
				mem.Resource = fmt.Sprintf("ips[%d]", getVmaId(vma.GetShmid()))
			case status&(1<<8) != 0:
				mem.Resource = fmt.Sprintf("shmem[%d]", getVmaId(vma.GetShmid()))
			}
			if vma.GetStatus()&1 == 0 {
				mem.Resource = fmt.Sprint(mem.Resource, " *")
			}

			// Check page protection
			r, w, x := "-", "-", "-"
			prot := vma.GetProt()
			if prot&1 != 0 {
				r = "r"
			}
			if prot&2 != 0 {
				w = "w"
			}
			if prot&4 != 0 {
				x = "x"
			}
			mem.Protection = fmt.Sprint(r, w, x)

			memMap.Mems = append(memMap.Mems, &mem)
		}

		memMaps = append(memMaps, &memMap)
	}

	return memMaps, nil
}

// RssMap represents the resident set size mapping of a single process
type RssMap struct {
	PId uint32 `json:"pId"`
	/*
		walrus -> walruses
		radius -> radii
		If you code without breaks,
		rss -> rsi :P
	*/
	Rsses []*Rss `json:"rss,omitempty"`
}

// Rss represents a single resident set size mapping
type Rss struct {
	PhyAddr  string `json:"phyAddr,omitempty"`
	PhyPages int64  `json:"phyPages,omitempty"`
	Vmas     []*Vma `json:"vmas,omitempty"`
	Resource string `json:"resource,omitempty"`
}

// Vma represents a single virtual memory area
type Vma struct {
	Addr  string `json:"addr,omitempty"`
	Pages int64  `json:"pages,omitempty"`
}

// ExploreRss traverses the process tree and returns
// a list of processes with their RSS mappings
func (c *crit) ExploreRss() ([]*RssMap, error) {
	psTreeImg, err := getImg(fmt.Sprintf("%s/pstree.img", c.inputDirPath))
	if err != nil {
		return nil, err
	}

	rssMaps := make([]*RssMap, 0)
	for _, entry := range psTreeImg.Entries {
		process := entry.Message.(*images.PstreeEntry)
		pId := process.GetPid()
		// Get virtual memory addresses
		mmImg, err := getImg(fmt.Sprintf("%s/mm-%d.img", c.inputDirPath, pId))
		if err != nil {
			return nil, err
		}
		vmas := mmImg.Entries[0].Message.(*images.MmEntry).GetVmas()
		// Get physical memory addresses
		pagemapImg, err := getImg(fmt.Sprintf("%s/pagemap-%d.img", c.inputDirPath, pId))
		if err != nil {
			return nil, err
		}

		vmaIndex, vmaIndexPrev := 0, -1
		rssMap := RssMap{PId: pId}
		// Skip pagemap head entry
		for _, pagemapEntry := range pagemapImg.Entries[1:] {
			pagemapData := pagemapEntry.Message.(*images.PagemapEntry)
			rss := Rss{
				PhyAddr:  fmt.Sprintf("%x", pagemapData.GetVaddr()),
				PhyPages: int64(pagemapData.GetNrPages()),
			}

			for vmas[vmaIndex].GetEnd() <= pagemapData.GetVaddr() {
				vmaIndex++
			}
			// Compute last virtual address
			pagemapEnd := pagemapData.GetVaddr() + (uint64(pagemapData.GetNrPages()) << 12)

			for vmas[vmaIndex].GetStart() < pagemapEnd {
				if vmaIndex == vmaIndexPrev {
					// Use tilde to indicate that VMA is of the previous pagemap entry
					rss.Vmas = append(rss.Vmas, &Vma{Addr: "~"})
					vmaIndex++
					continue
				}

				rss.Vmas = append(rss.Vmas, &Vma{
					Addr:  fmt.Sprintf("%x", vmas[vmaIndex].GetStart()),
					Pages: int64(vmas[vmaIndex].GetEnd()-vmas[vmaIndex].GetStart()) >> 12,
				})
				// Pages used by a file
				if vmas[vmaIndex].GetStatus()&((1<<6)|(1<<7)) != 0 {
					file, err := getFilePath(c.inputDirPath,
						uint32(vmas[vmaIndex].GetShmid()), images.FdTypes_REG)
					if err != nil {
						return nil, err
					}
					rss.Resource = file
				}
				// Set reference to current index before increment
				vmaIndexPrev = vmaIndex
				vmaIndex++
			}

			vmaIndex--
			rssMap.Rsses = append(rssMap.Rsses, &rss)
		}

		rssMaps = append(rssMaps, &rssMap)
	}

	return rssMaps, nil
}
