package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/checkpoint-restore/go-criu/v4"
	"github.com/checkpoint-restore/go-criu/v4/rpc"
	"google.golang.org/protobuf/proto"
)

// TestNfy struct
type TestNfy struct {
	criu.NoNotify
}

// PreDump test function
func (c TestNfy) PreDump() error {
	log.Println("TEST PRE DUMP")
	return nil
}

func doDump(c *criu.Criu, pidS string, imgDir string, pre bool, prevImg string) error {
	log.Println("Dumping")
	pid, _ := strconv.Atoi(pidS)
	img, err := os.Open(imgDir)
	if err != nil {
		return fmt.Errorf("can't open image dir: %v", err)
	}
	defer img.Close()

	opts := rpc.CriuOpts{
		Pid:         proto.Int32(int32(pid)),
		ImagesDirFd: proto.Int32(int32(img.Fd())),
		LogLevel:    proto.Int32(4),
		LogFile:     proto.String("dump.log"),
	}

	if prevImg != "" {
		opts.ParentImg = proto.String(prevImg)
		opts.TrackMem = proto.Bool(true)
	}

	if pre {
		err = c.PreDump(opts, TestNfy{})
	} else {
		err = c.Dump(opts, TestNfy{})
	}
	if err != nil {
		return fmt.Errorf("dump fail: %v", err)
	}

	return nil
}

// Usage: test $act $pid $images_dir
func main() {
	c := criu.MakeCriu()
	// Read out CRIU version
	version, err := c.GetCriuVersion()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	log.Println("CRIU version", version)
	// Check if version at least 3.2
	result, err := c.IsCriuAtLeast(30200)
	if err != nil {
		log.Fatalln(err)
	}
	if !result {
		log.Fatalln("CRIU version to old")
	}
	act := os.Args[1]
	switch act {
	case "dump":
		err := doDump(c, os.Args[2], os.Args[3], false, "")
		if err != nil {
			log.Fatalln("dump failed:", err)
		}
	case "dump2":
		err := c.Prepare()
		if err != nil {
			fmt.Print(err)
			os.Exit(1)
		}

		err = doDump(c, os.Args[2], os.Args[3]+"/pre", true, "")
		if err != nil {
			log.Fatalln("pre-dump failed:", err)
		}
		err = doDump(c, os.Args[2], os.Args[3], false, "./pre")
		if err != nil {
			log.Fatalln("dump failed: ", err)
		}

		c.Cleanup()
	case "restore":
		log.Println("Restoring")
		img, err := os.Open(os.Args[2])
		if err != nil {
			log.Fatalln("can't open image dir:", err)
		}
		defer img.Close()

		opts := rpc.CriuOpts{
			ImagesDirFd: proto.Int32(int32(img.Fd())),
			LogLevel:    proto.Int32(4),
			LogFile:     proto.String("restore.log"),
		}

		err = c.Restore(opts, nil)
		if err != nil {
			log.Fatalln("restore failed:", err)
		}
	default:
		log.Fatalln("unknown action")
	}

	log.Println("Success")
}
