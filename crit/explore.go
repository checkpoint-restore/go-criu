package crit

import (
	"fmt"
	"path/filepath"
	"strconv"

	criu_core "github.com/checkpoint-restore/go-criu/v7/crit/images/criu-core"
	"github.com/checkpoint-restore/go-criu/v7/crit/images/fdinfo"
	"github.com/checkpoint-restore/go-criu/v7/crit/images/fs"
	"github.com/checkpoint-restore/go-criu/v7/crit/images/mm"
	"github.com/checkpoint-restore/go-criu/v7/crit/images/pagemap"
	"github.com/checkpoint-restore/go-criu/v7/crit/images/pstree"
)

// PsTree represents the process tree
type PsTree struct {
	PID      uint32               `json:"pId"`
	PgID     uint32               `json:"pgId"`
	SID      uint32               `json:"sId"`
	Comm     string               `json:"comm"`
	Process  *pstree.PstreeEntry  `json:"-"`
	Core     *criu_core.CoreEntry `json:"-"`
	Children []*PsTree            `json:"children,omitempty"`
}

// ExplorePs constructs the process tree and returns the root process
func (c *crit) ExplorePs() (*PsTree, error) {
	psTreeImg, err := getImg(filepath.Join(c.inputDirPath, "pstree.img"), &pstree.PstreeEntry{})
	if err != nil {
		return nil, err
	}

	processes := make(map[uint32]*PsTree)
	var psTreeRoot *PsTree
	for _, entry := range psTreeImg.Entries {
		process := entry.Message.(*pstree.PstreeEntry)
		pID := process.GetPid()

		coreImg, err := getImg(filepath.Join(c.inputDirPath, fmt.Sprintf("core-%d.img", pID)), &criu_core.CoreEntry{})
		if err != nil {
			return nil, err
		}
		coreData := coreImg.Entries[0].Message.(*criu_core.CoreEntry)

		ps := &PsTree{
			PID:     pID,
			PgID:    process.GetPgid(),
			SID:     process.GetSid(),
			Comm:    coreData.GetTc().GetComm(),
			Process: process,
			Core:    coreData,
		}
		// If there is no parent process, then it is the root
		if process.GetPpid() == 0 {
			psTreeRoot = ps
		}
		processes[pID] = ps
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
	Files []*File `json:"files"`
}

// File represents a single opened file
type File struct {
	Fd   string `json:"fd"`
	Type string `json:"type,omitempty"`
	Path string `json:"path"`
}

// ExploreFds searches the process tree for open files
// and returns a list of PIDs with the corresponding files
func (c *crit) ExploreFds() ([]*Fd, error) {
	psTreeImg, err := getImg(filepath.Join(c.inputDirPath, "pstree.img"), &pstree.PstreeEntry{})
	if err != nil {
		return nil, err
	}

	fds := make([]*Fd, 0)
	for _, entry := range psTreeImg.Entries {
		process := entry.Message.(*pstree.PstreeEntry)
		pID := process.GetPid()
		// Get file with object IDs
		idsImg, err := getImg(filepath.Join(c.inputDirPath, fmt.Sprintf("ids-%d.img", pID)), &criu_core.TaskKobjIdsEntry{})
		if err != nil {
			return nil, err
		}
		filesID := idsImg.Entries[0].Message.(*criu_core.TaskKobjIdsEntry).GetFilesId()
		// Get open file descriptors
		fdInfoImg, err := getImg(filepath.Join(c.inputDirPath, fmt.Sprintf("fdinfo-%d.img", filesID)), &fdinfo.FdinfoEntry{})
		if err != nil {
			return nil, err
		}

		fdEntry := Fd{PId: pID}
		for _, fdInfoEntry := range fdInfoImg.Entries {
			fdInfo := fdInfoEntry.Message.(*fdinfo.FdinfoEntry)
			filePath, err := getFilePath(c.inputDirPath,
				fdInfo.GetId(), fdInfo.GetType())
			if err != nil {
				return nil, err
			}
			file := File{
				Fd:   strconv.FormatUint(uint64(fdInfo.GetFd()), 10),
				Type: fdInfo.GetType().String(),
				Path: filePath,
			}
			fdEntry.Files = append(fdEntry.Files, &file)
		}
		// Get chroot and chdir info
		fsImg, err := getImg(filepath.Join(c.inputDirPath, fmt.Sprintf("fs-%d.img", pID)), &fs.FsEntry{})
		if err != nil {
			return nil, err
		}
		fs := fsImg.Entries[0].Message.(*fs.FsEntry)
		filePath, err := getFilePath(c.inputDirPath,
			fs.GetCwdId(), fdinfo.FdTypes_REG)
		if err != nil {
			return nil, err
		}
		fdEntry.Files = append(fdEntry.Files, &File{
			Fd:   "cwd",
			Path: filePath,
		})
		filePath, err = getFilePath(c.inputDirPath,
			fs.GetRootId(), fdinfo.FdTypes_REG)
		if err != nil {
			return nil, err
		}
		fdEntry.Files = append(fdEntry.Files, &File{
			Fd:   "root",
			Path: filePath,
		})

		// Omit if the process has no file descriptors
		if len(fdEntry.Files) == 0 {
			continue
		}

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
	psTreeImg, err := getImg(filepath.Join(c.inputDirPath, "pstree.img"), &pstree.PstreeEntry{})
	if err != nil {
		return nil, err
	}

	vmaIDMap, vmaID := make(map[uint64]int), 0
	// Use a closure to handle the ID counter
	getVmaID := func(shmId uint64) int {
		if _, ok := vmaIDMap[shmId]; !ok {
			vmaIDMap[shmId] = vmaID
			vmaID++
		}
		return vmaIDMap[shmId]
	}

	memMaps := make([]*MemMap, 0)
	for _, entry := range psTreeImg.Entries {
		process := entry.Message.(*pstree.PstreeEntry)
		pID := process.GetPid()
		// Get memory mappings
		mmImg, err := getImg(filepath.Join(c.inputDirPath, fmt.Sprintf("mm-%d.img", pID)), &mm.MmEntry{})
		if err != nil {
			return nil, err
		}
		mmInfo := mmImg.Entries[0].Message.(*mm.MmEntry)
		exePath, err := getFilePath(c.inputDirPath,
			mmInfo.GetExeFileId(), fdinfo.FdTypes_REG)
		if err != nil {
			return nil, err
		}

		memMap := MemMap{
			PId: pID,
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
					uint32(vma.GetShmid()), fdinfo.FdTypes_REG)
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
				mem.Resource = fmt.Sprintf("packet[%d]", getVmaID(vma.GetShmid()))
			case status&(1<<10) != 0:
				mem.Resource = fmt.Sprintf("ips[%d]", getVmaID(vma.GetShmid()))
			case status&(1<<8) != 0:
				mem.Resource = fmt.Sprintf("shmem[%d]", getVmaID(vma.GetShmid()))
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
	psTreeImg, err := getImg(filepath.Join(c.inputDirPath, "pstree.img"), &pstree.PstreeEntry{})
	if err != nil {
		return nil, err
	}

	rssMaps := make([]*RssMap, 0)
	for _, entry := range psTreeImg.Entries {
		process := entry.Message.(*pstree.PstreeEntry)
		pID := process.GetPid()
		// Get virtual memory addresses
		mmImg, err := getImg(filepath.Join(c.inputDirPath, fmt.Sprintf("mm-%d.img", pID)), &mm.MmEntry{})
		if err != nil {
			return nil, err
		}
		vmas := mmImg.Entries[0].Message.(*mm.MmEntry).GetVmas()
		// Get physical memory addresses
		pagemapImg, err := getImg(filepath.Join(c.inputDirPath, fmt.Sprintf("pagemap-%d.img", pID)), &pagemap.PagemapEntry{})
		if err != nil {
			return nil, err
		}

		vmaIndex, vmaIndexPrev := 0, -1
		rssMap := RssMap{PId: pID}
		// Skip pagemap head entry
		for _, pagemapEntry := range pagemapImg.Entries[1:] {
			pagemapData := pagemapEntry.Message.(*pagemap.PagemapEntry)
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
						uint32(vmas[vmaIndex].GetShmid()), fdinfo.FdTypes_REG)
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

// Sk represents the sockets associated with a single process
type Sk struct {
	PId     uint32    `json:"pId"`
	Sockets []*Socket `json:"sockets"`
}

// Socket represents a single socket
type Socket struct {
	Fd       uint32 `json:"fd"`
	FdType   string `json:"fdType"`
	Family   string `json:"family,omitempty"`
	Protocol string `json:"protocol,omitempty"`
	Type     string `json:"type,omitempty"`
	State    string `json:"state,omitempty"`
	SrcAddr  string `json:"srcAddr,omitempty"`
	SrcPort  uint32 `json:"srcPort,omitempty"`
	DestAddr string `json:"destAddr,omitempty"`
	DestPort uint32 `json:"destPort,omitempty"`
	SendBuf  string `json:"sendBuf,omitempty"`
	RecvBuf  string `json:"recvBuf,omitempty"`
}

// ExploreSk searches the process tree for sockets
// and returns a list of PIDs with the associated sockets
func (c *crit) ExploreSk() ([]*Sk, error) {
	psTreeImg, err := getImg(filepath.Join(c.inputDirPath, "pstree.img"), &pstree.PstreeEntry{})
	if err != nil {
		return nil, err
	}

	sks := make([]*Sk, 0)
	for _, entry := range psTreeImg.Entries {
		process := entry.Message.(*pstree.PstreeEntry)
		pID := process.GetPid()
		// Get file with object IDs
		idsImg, err := getImg(filepath.Join(c.inputDirPath, fmt.Sprintf("ids-%d.img", pID)), &criu_core.TaskKobjIdsEntry{})
		if err != nil {
			return nil, err
		}
		filesID := idsImg.Entries[0].Message.(*criu_core.TaskKobjIdsEntry).GetFilesId()
		// Get open file descriptors
		fdInfoImg, err := getImg(filepath.Join(c.inputDirPath, fmt.Sprintf("fdinfo-%d.img", filesID)), &fdinfo.FdinfoEntry{})
		if err != nil {
			return nil, err
		}
		skEntry := Sk{PId: pID}
		for _, fdInfoEntry := range fdInfoImg.Entries {
			fdInfo := fdInfoEntry.Message.(*fdinfo.FdinfoEntry)
			file, err := getFile(c.inputDirPath, fdInfo.GetId())
			if err != nil {
				return nil, err
			}
			socket := Socket{
				Fd:     fdInfo.GetFd(),
				FdType: fdInfo.GetType().String(),
			}
			switch fdInfo.GetType() {
			case fdinfo.FdTypes_INETSK:
				if isk := file.GetIsk(); isk != nil {
					socket.State = getSkState(tcpState(isk.GetState()))
					socket.Family = getAddressFamily(isk.GetFamily())
					socket.Protocol = getSkProtocol(isk.GetProto())
					socket.Type = getSkType(isk.GetType())
					socket.SrcAddr = processIP(isk.GetSrcAddr())
					socket.SrcPort = isk.GetSrcPort()
					socket.DestAddr = processIP(isk.GetDstAddr())
					socket.DestPort = isk.GetDstPort()
					socket.SendBuf = countBytes(int64(isk.GetOpts().GetSoSndbuf()))
					socket.RecvBuf = countBytes(int64(isk.GetOpts().GetSoRcvbuf()))
				}
			case fdinfo.FdTypes_UNIXSK:
				if usk := file.GetUsk(); usk != nil {
					socket.State = getSkState(tcpState(usk.GetState()))
					socket.Type = getSkType(usk.GetType())
					socket.SrcAddr = string(usk.GetName())
					socket.SendBuf = countBytes(int64(usk.GetOpts().GetSoSndbuf()))
					socket.RecvBuf = countBytes(int64(usk.GetOpts().GetSoRcvbuf()))
				}
			case fdinfo.FdTypes_PACKETSK:
				if psk := file.GetPsk(); psk != nil {
					socket.Type = getSkType(psk.GetProtocol())
					socket.Protocol = getSkProtocol(psk.GetProtocol())
					socket.SendBuf = countBytes(int64(psk.GetOpts().GetSoSndbuf()))
					socket.RecvBuf = countBytes(int64(psk.GetOpts().GetSoRcvbuf()))
				}
			case fdinfo.FdTypes_NETLINKSK:
				if nlsk := file.GetNlsk(); nlsk != nil {
					socket.State = getSkState(tcpState(nlsk.GetState()))
					socket.Protocol = getSkProtocol(nlsk.GetProtocol())
					socket.Type = getSkType(nlsk.GetProtocol())
					socket.SendBuf = countBytes(int64(nlsk.GetOpts().GetSoSndbuf()))
					socket.RecvBuf = countBytes(int64(nlsk.GetOpts().GetSoRcvbuf()))
				}
			default:
				continue
			}

			skEntry.Sockets = append(skEntry.Sockets, &socket)
		}

		// Omit if the process has no associated sockets
		if len(skEntry.Sockets) == 0 {
			continue
		}

		sks = append(sks, &skEntry)
	}

	return sks, nil
}
