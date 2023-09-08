package crit

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"os"

	"github.com/checkpoint-restore/go-criu/v7/magic"
	"google.golang.org/protobuf/proto"
)

// encodeImg identifies the type of image file
// and calls the appropriate encode handler
func encodeImg(img *CriuImage, f *os.File) error {
	magicMap := magic.LoadMagic()
	var err error

	// Write magic
	magic, ok := magicMap.ByName[img.Magic]
	if !ok {
		return errors.New(fmt.Sprint("unknown magic ", img.Magic))
	}
	magicBuf := make([]byte, 4)
	if img.Magic != "INVENTORY" {
		if img.Magic == "STATS" || img.Magic == "IRMAP_CACHE" {
			binary.LittleEndian.PutUint32(magicBuf, uint32(magicMap.ByName["IMG_SERVICE"]))
		} else {
			binary.LittleEndian.PutUint32(magicBuf, uint32(magicMap.ByName["IMG_COMMON"]))
		}
		if _, err = f.Write(magicBuf); err != nil {
			return err
		}
	}
	binary.LittleEndian.PutUint32(magicBuf, uint32(magic))
	if _, err = f.Write(magicBuf); err != nil {
		return err
	}

	// Call handler for entries
	switch img.Magic {
	// Special handler for ghost files
	case "GHOST_FILE":
		err = img.encodeGhostFile(f)
	// Default handler with func for extra data
	case "BPFMAP_DATA":
		err = img.encodeDefault(f, encodeBpfmapData)
	case "FIFO_DATA":
		err = img.encodeDefault(f, encodePipesData)
	case "IPCNS_MSG":
		err = img.encodeDefault(f, encodeIpcMsg)
	case "IPCNS_SEM":
		err = img.encodeDefault(f, encodeIpcSem)
	case "IPCNS_SHM":
		err = img.encodeDefault(f, encodeIpcShm)
	case "PIPES_DATA":
		err = img.encodeDefault(f, encodePipesData)
	case "SK_QUEUES":
		err = img.encodeDefault(f, encodeSkQueues)
	case "TCP_STREAM":
		err = img.encodeDefault(f, encodeTCPStream)
	default:
		err = img.encodeDefault(f, nil)
	}
	if err != nil {
		return err
	}

	return nil
}

// encodeDefault is used for all image files
// that are in the standard protobuf format
func (img *CriuImage) encodeDefault(
	f *os.File,
	encodeExtra func(string) ([]byte, error),
) error {
	sizeBuf := make([]byte, 4)

	for _, entry := range img.Entries {
		payload, err := proto.Marshal(entry.Message)
		if err != nil {
			return err
		}
		// Write size of payload into buffer
		binary.LittleEndian.PutUint32(sizeBuf, uint32(len(payload)))

		if _, err = f.Write(sizeBuf); err != nil {
			return err
		}
		if _, err = f.Write(payload); err != nil {
			return err
		}

		// Write extra data
		if encodeExtra != nil {
			if entry.Extra != "" {
				extraPayload, err := encodeExtra(entry.Extra)
				if err != nil {
					return err
				}
				if _, err = f.Write(extraPayload); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// Special handler for ghost image
func (img *CriuImage) encodeGhostFile(f *os.File) error {
	sizeBuf := make([]byte, 4)
	// Write primary entry
	payload, err := proto.Marshal(img.Entries[0].Message)
	if err != nil {
		return err
	}
	// Write size of payload into buffer
	binary.LittleEndian.PutUint32(sizeBuf, uint32(len(payload)))

	if _, err = f.Write(sizeBuf); err != nil {
		return err
	}
	if _, err = f.Write(payload); err != nil {
		return err
	}

	// If there is only one entry,
	// then no chunks are present
	if len(img.Entries) == 1 {
		// Write extra data
		extraPayload, err := base64.StdEncoding.DecodeString(img.Entries[0].Extra)
		if err != nil {
			return err
		}
		if _, err = f.Write(extraPayload); err != nil {
			return err
		}

		return nil
	}

	// Write chunks
	for _, entry := range img.Entries[1:] {
		payload, err = proto.Marshal(entry.Message)
		if err != nil {
			return err
		}
		extraPayload, err := base64.StdEncoding.DecodeString(entry.Extra)
		if err != nil {
			return err
		}

		binary.LittleEndian.PutUint32(sizeBuf, uint32(len(extraPayload)))

		if _, err = f.Write(sizeBuf); err != nil {
			return err
		}
		if _, err = f.Write(payload); err != nil {
			return err
		}
		if _, err = f.Write(extraPayload); err != nil {
			return err
		}
	}

	return nil
}
