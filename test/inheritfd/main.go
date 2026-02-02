package main

import (
	"fmt"
	"log"
	"os"
	"syscall"
	"time"

	"github.com/checkpoint-restore/go-criu/v8"
	"github.com/checkpoint-restore/go-criu/v8/rpc"
	"google.golang.org/protobuf/proto"
)

func getSwrkPid() int {
	myPid := os.Getpid()
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/task/%d/children", myPid, myPid))
	if err != nil {
		return 0
	}
	var swrkPid int
	if _, err := fmt.Sscanf(string(data), "%d", &swrkPid); err != nil {
		return 0
	}
	return swrkPid
}

func waitForSwrkPid() int {
	for i := 0; i < 50; i++ {
		swrkPid := getSwrkPid()
		if swrkPid != 0 {
			return swrkPid
		}
		time.Sleep(10 * time.Millisecond)
	}
	return 0
}

func swrkHasInode(swrkPid int, ino uint64) (bool, error) {
	entries, err := os.ReadDir(fmt.Sprintf("/proc/%d/fd", swrkPid))
	if err != nil {
		return false, err
	}
	for _, entry := range entries {
		var stat syscall.Stat_t
		if err := syscall.Stat(fmt.Sprintf("/proc/%d/fd/%s", swrkPid, entry.Name()), &stat); err != nil {
			continue
		}
		if stat.Ino == ino {
			return true, nil
		}
	}
	return false, nil
}

func testInheritFd(netnsIno uint64, netnsFile *os.File) {
	c := criu.MakeCriu()
	c.AddInheritFd("net", netnsFile)
	if err := c.Prepare(); err != nil {
		log.Fatalln(err)
	}

	swrkPid := waitForSwrkPid()
	if swrkPid == 0 {
		log.Fatalln("no swrk pid found")
	}
	has, err := swrkHasInode(swrkPid, netnsIno)
	if err != nil {
		log.Fatalln(err)
	}
	if !has {
		log.Fatalln("fd not inherited with AddInheritFd")
	}

	// Send a dummy RPC so swrk exits cleanly
	if _, err := c.GetCriuVersion(); err != nil {
		log.Fatalln(err)
	}

	if err := c.Cleanup(); err != nil {
		log.Fatalln(err)
	}
}

func testNoInheritFd(netnsIno uint64) {
	c := criu.MakeCriu()
	if err := c.Prepare(); err != nil {
		log.Fatalln(err)
	}

	swrkPid := waitForSwrkPid()
	if swrkPid == 0 {
		log.Fatalln("no swrk pid found")
	}
	has, err := swrkHasInode(swrkPid, netnsIno)
	if err != nil {
		log.Fatalln(err)
	}
	if has {
		log.Fatalln("fd incorrectly inherited without AddInheritFd")
	}

	// Send a dummy RPC so swrk exits cleanly
	if _, err := c.GetCriuVersion(); err != nil {
		log.Fatalln(err)
	}

	if err := c.Cleanup(); err != nil {
		log.Fatalln(err)
	}
}

func testAutoPopulateInheritFd(netnsFile *os.File) {
	c := criu.MakeCriu()
	c.AddInheritFd("testKey", netnsFile)

	opts := &rpc.CriuOpts{
		ImagesDir: proto.String("/nonexistent"),
	}

	// Call will fail (no images), but ensureInheritFd() runs first
	_ = c.PreDump(opts, nil)

	// Verify opts.InheritFd was auto-populated
	inheritFds := opts.GetInheritFd()
	if len(inheritFds) != 1 {
		log.Fatalf("opts.InheritFd not auto-populated: got %d, want 1", len(inheritFds))
	}
	if inheritFds[0].GetKey() != "testKey" || inheritFds[0].GetFd() != 3 {
		log.Fatalf("opts.InheritFd wrong: key=%s fd=%d", inheritFds[0].GetKey(), inheritFds[0].GetFd())
	}
}

// Usage: test-inheritfd
func main() {
	netnsFile, err := os.Open("/proc/self/ns/net")
	if err != nil {
		log.Fatalln(err)
	}
	defer netnsFile.Close()

	var stat syscall.Stat_t
	if err := syscall.Fstat(int(netnsFile.Fd()), &stat); err != nil {
		log.Fatalln(err)
	}
	netnsIno := stat.Ino

	testNoInheritFd(netnsIno)
	testInheritFd(netnsIno, netnsFile)
	testAutoPopulateInheritFd(netnsFile)
	log.Println("PASS")
}
