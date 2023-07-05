package crit

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/checkpoint-restore/go-criu/v6/crit/images/fdinfo"
	"github.com/checkpoint-restore/go-criu/v6/crit/images/pipe"
	"github.com/checkpoint-restore/go-criu/v6/crit/images/regfile"
	sk_unix "github.com/checkpoint-restore/go-criu/v6/crit/images/sk-unix"
	"github.com/checkpoint-restore/go-criu/v6/magic"
	"google.golang.org/protobuf/proto"
)

// Helper to decode magic name from hex value
func ReadMagic(f *os.File) (string, error) {
	magicMap := magic.LoadMagic()
	// Read magic
	magicBuf := make([]byte, 4)
	if _, err := f.Read(magicBuf); err != nil {
		return "", err
	}
	magicValue := uint64(binary.LittleEndian.Uint32(magicBuf))
	if magicValue == magicMap.ByName["IMG_COMMON"] ||
		magicValue == magicMap.ByName["IMG_SERVICE"] {
		if _, err := f.Read(magicBuf); err != nil {
			return "", err
		}
		magicValue = uint64(binary.LittleEndian.Uint32(magicBuf))
	}

	// Identify magic
	magicName, ok := magicMap.ByValue[magicValue]
	if !ok {
		return "", fmt.Errorf("unknown magic 0x%x", magicValue)
	}

	return magicName, nil
}

// Helper to convert bytes into a more readable unit
func countBytes(n int64) string {
	const UNIT int64 = 1024
	if n < UNIT {
		return fmt.Sprint(n, " B")
	}
	div, exp := UNIT, 0
	for i := n / UNIT; i >= UNIT; i /= UNIT {
		div *= UNIT
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGTPE"[exp])
}

// Function to count number of top-level entries
func countImg(f *os.File) (*CriuImage, error) {
	img := CriuImage{}
	var err error

	// Identify magic
	if img.Magic, err = ReadMagic(f); err != nil {
		return nil, err
	}

	count := 0
	sizeBuf := make([]byte, 4)
	// Read payload size and increment counter until EOF
	for {
		if n, err := f.Read(sizeBuf); err != nil {
			if n == 0 && err == io.EOF {
				break
			}
			return nil, err
		}

		payloadSize := int64(binary.LittleEndian.Uint32(sizeBuf))
		if _, err = f.Seek(payloadSize, 1); err != nil {
			return nil, err
		}
		count++
	}
	// Decrement counter by 1 for pagemap file,
	// as pagemap head is not a top-level entry
	if img.Magic == "PAGEMAP" {
		count--
	}

	entry := CriuEntry{Extra: strconv.Itoa(count)}
	img.Entries = append(img.Entries, &entry)
	return &img, nil
}

// Helper to decode image when file path is given
func getImg(path string, entryType proto.Message) (*CriuImage, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening binary file: %w", err)
	}
	defer file.Close()

	return decodeImg(file, entryType, false)
}

// Global variables to cache loaded images
var (
	filesImg, regImg, pipeImg, unixSkImg *CriuImage
)

// Helper to get file path for exploring file descriptors
func getFilePath(dir string, fID uint32, fType fdinfo.FdTypes) (string, error) {
	var filePath string
	var err error
	// Get open files
	if filesImg == nil {
		filesImg, err = getImg(filepath.Join(dir, "files.img"), &fdinfo.FileEntry{})
		if err != nil {
			return "", err
		}
	}

	// Check if file entry is present
	var file *fdinfo.FileEntry
	for _, entry := range filesImg.Entries {
		file = entry.Message.(*fdinfo.FileEntry)
		if file.GetId() == fID {
			break
		}
	}

	switch fType {
	case fdinfo.FdTypes_REG:
		filePath, err = getRegFilePath(dir, file, fID)
	case fdinfo.FdTypes_PIPE:
		filePath, err = getPipeFilePath(dir, file, fID)
	case fdinfo.FdTypes_UNIXSK:
		filePath, err = getUnixSkFilePath(dir, file, fID)
	default:
		filePath = fmt.Sprintf("%s.%d", fType.String(), fID)
	}

	return filePath, err
}

// Helper to get file path of regular files
func getRegFilePath(dir string, file *fdinfo.FileEntry, fID uint32) (string, error) {
	var err error
	if file != nil {
		if file.GetReg() != nil {
			return file.GetReg().GetName(), nil
		}
		return "Unknown path", nil
	}

	if regImg == nil {
		regImg, err = getImg(filepath.Join(dir, "reg-files.img"), &regfile.RegFileEntry{})
		if err != nil {
			return "", err
		}
	}
	for _, entry := range regImg.Entries {
		regFile := entry.Message.(*regfile.RegFileEntry)
		if regFile.GetId() == fID {
			return regFile.GetName(), nil
		}
	}

	return "Unknown path", nil
}

// Helper to get file path of pipe files
func getPipeFilePath(dir string, file *fdinfo.FileEntry, fID uint32) (string, error) {
	var err error
	if file != nil {
		if file.GetPipe() != nil {
			return fmt.Sprintf("pipe[%d]", file.GetPipe().GetPipeId()), nil
		}
		return "pipe[?]", nil
	}

	if pipeImg == nil {
		pipeImg, err = getImg(filepath.Join(dir, "pipes.img"), &pipe.PipeEntry{})
		if err != nil {
			return "", err
		}
	}
	for _, entry := range pipeImg.Entries {
		pipeFile := entry.Message.(*pipe.PipeEntry)
		if pipeFile.GetId() == fID {
			return fmt.Sprintf("pipe[%d]", pipeFile.GetPipeId()), nil
		}
	}

	return "pipe[?]", nil
}

// Helper to get file path of UNIX socket files
func getUnixSkFilePath(dir string, file *fdinfo.FileEntry, fID uint32) (string, error) {
	var err error
	if file != nil {
		if file.GetUsk() != nil {
			return fmt.Sprintf(
				"unix[%d (%d) %s]",
				file.GetUsk().GetIno(),
				file.GetUsk().GetPeer(),
				file.GetUsk().GetName(),
			), nil
		}
		return "unix[?]", nil
	}

	if unixSkImg == nil {
		unixSkImg, err = getImg(filepath.Join(dir, "unixsk.img"), &sk_unix.UnixSkEntry{})
		if err != nil {
			return "", err
		}
	}
	for _, entry := range unixSkImg.Entries {
		unixSkFile := entry.Message.(*sk_unix.UnixSkEntry)
		if unixSkFile.GetId() == fID {
			return fmt.Sprintf(
				"unix[%d (%d) %s]",
				unixSkFile.GetIno(),
				unixSkFile.GetPeer(),
				unixSkFile.GetName(),
			), nil
		}
	}

	return "unix[?]", nil
}
