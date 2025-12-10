package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"strconv"

	"github.com/checkpoint-restore/go-criu/v8"
	"github.com/checkpoint-restore/go-criu/v8/rpc"
	"github.com/checkpoint-restore/go-criu/v8/utils"
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
	pid, err := strconv.ParseInt(pidS, 10, 32)
	if err != nil {
		return fmt.Errorf("can't parse pid: %w", err)
	}
	img, err := os.Open(imgDir)
	if err != nil {
		return fmt.Errorf("can't open image dir: %w", err)
	}
	defer img.Close()

	opts := &rpc.CriuOpts{
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
		return fmt.Errorf("dump fail: %w", err)
	}

	return nil
}

func featureCheck(c *criu.Criu) error {
	features := &rpc.CriuFeatures{
		MemTrack:   proto.Bool(false),
		LazyPages:  proto.Bool(false),
		PidfdStore: proto.Bool(false),
	}
	featuresToCompare := &rpc.CriuFeatures{
		MemTrack:   proto.Bool(false),
		LazyPages:  proto.Bool(false),
		PidfdStore: proto.Bool(false),
	}
	env := os.Getenv("CRIU_FEATURE_MEM_TRACK")
	if env != "" {
		val, err := strconv.Atoi(env)
		if err != nil {
			return err
		}
		features.MemTrack = proto.Bool(val != 0)
		featuresToCompare.MemTrack = proto.Bool(val != 0)
	}
	env = os.Getenv("CRIU_FEATURE_LAZY_PAGES")
	if env != "" {
		val, err := strconv.Atoi(env)
		if err != nil {
			return err
		}
		features.LazyPages = proto.Bool(val != 0)
		featuresToCompare.LazyPages = proto.Bool(val != 0)
	}
	env = os.Getenv("CRIU_FEATURE_PIDFD_STORE")
	if env != "" {
		val, err := strconv.Atoi(env)
		if err != nil {
			return err
		}
		features.PidfdStore = proto.Bool(val != 0)
		featuresToCompare.PidfdStore = proto.Bool(val != 0)
	}

	features, err := c.FeatureCheck(features)
	if err != nil {
		return err
	}

	if features.GetMemTrack() != featuresToCompare.GetMemTrack() {
		return fmt.Errorf(
			"unexpected MemTrack FeatureCheck result %v:%v",
			features.GetMemTrack(),
			featuresToCompare.GetMemTrack(),
		)
	}

	if features.GetLazyPages() != featuresToCompare.GetLazyPages() {
		return fmt.Errorf(
			"unexpected LazyPages FeatureCheck result %v:%v",
			features.GetLazyPages(),
			featuresToCompare.GetLazyPages(),
		)
	}

	if features.GetPidfdStore() != featuresToCompare.GetPidfdStore() {
		return fmt.Errorf(
			"unexpected PidfdStore FeatureCheck result %v:%v",
			features.GetPidfdStore(),
			featuresToCompare.GetPidfdStore(),
		)
	}

	isMemTrack := utils.IsMemTrack()
	if isMemTrack != featuresToCompare.GetMemTrack() {
		return fmt.Errorf(
			"unexpected MemTrack FeatureCheck result %v:%v",
			isMemTrack,
			featuresToCompare.GetMemTrack(),
		)
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
	// Compare if version from convenience function matches
	version2, err := utils.GetCriuVersion()
	if err != nil {
		log.Fatalln(err)
	}
	if version != version2 {
		log.Fatalf("Detected versions do not match (%d != %d)", version, version2)
	}
	// Check if version at least 3.2
	result, err := c.IsCriuAtLeast(30200)
	if err != nil {
		log.Fatalln(err)
	}
	if !result {
		log.Fatalln("CRIU version to old")
	}

	if err := utils.CheckForCriu(30200); err != nil {
		log.Fatalln(err)
	}

	if err := utils.CheckForCriu(math.MaxInt); err == nil {
		log.Fatalf("Checking for CRIU version %d should have failed.", math.MaxInt)
	}

	if err = featureCheck(c); err != nil {
		log.Fatalln(err)
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

		err = c.Cleanup()
		if err != nil {
			log.Fatalln("cleanup failed: ", err)
		}
	case "restore":
		log.Println("Restoring")
		img, err := os.Open(os.Args[2])
		if err != nil {
			log.Fatalln("can't open image dir:", err)
		}
		defer img.Close()

		opts := &rpc.CriuOpts{
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
