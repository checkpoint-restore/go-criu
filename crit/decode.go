package crit

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/checkpoint-restore/go-criu/v5/crit/images"
	"github.com/checkpoint-restore/go-criu/v5/magic"
	"google.golang.org/protobuf/proto"
)

func decodeImg(f *os.File, noPayload bool) (*CriuImage, error) {
	magicMap := magic.LoadMagic()
	img := CriuImage{}

	// Read magic
	buf := make([]byte, 4)
	if _, err := f.Read(buf); err != nil {
		return nil, err
	}
	magic := uint64(binary.LittleEndian.Uint32(buf))
	if magic == magicMap.ByName["IMG_COMMON"] ||
		magic == magicMap.ByName["IMG_SERVICE"] {
		if _, err := f.Read(buf); err != nil {
			return nil, err
		}
		magic = uint64(binary.LittleEndian.Uint32(buf))
	}

	// Identify magic
	img.Magic = magicMap.ByValue[magic]
	if img.Magic == "" {
		return nil, errors.New(fmt.Sprintf("Unknown magic 0x%x", magic))
	}

	// Call handler for entries
	var err error
	switch img.Magic {
	// Special handlers
	case "PAGEMAP":
		err = img.decodePagemap(f)
	case "GHOST_FILE":
		err = img.decodeGhostFile(f, noPayload)
	// Default handler with func for extra data
	case "PIPES_DATA":
		err = img.decodeDefault(f, decodePipesData, noPayload)
	case "FIFO_DATA":
		err = img.decodeDefault(f, decodePipesData, noPayload)
	case "SK_QUEUES":
		err = img.decodeDefault(f, decodeSkQueues, noPayload)
	case "TCP_STREAM":
		err = img.decodeDefault(f, decodeTcpStream, noPayload)
	case "BPFMAP_DATA":
		err = img.decodeDefault(f, decodeBpfmapData, noPayload)
	case "IPCNS_SEM":
		err = img.decodeDefault(f, decodeIpcSem, noPayload)
	case "IPCNS_SHM":
		err = img.decodeDefault(f, decodeIpcShm, noPayload)
	case "IPCNS_MSG":
		err = img.decodeDefault(f, decodeIpcMsg, noPayload)
	default:
		err = img.decodeDefault(f, nil, noPayload)
	}
	if err != nil {
		return nil, err
	}

	return &img, nil
}

func (img *CriuImage) decodeDefault(
	f *os.File,
	decodeExtra func(*os.File, proto.Message, bool) (string, error),
	noPayload bool,
) error {
	sizeBuf := make([]byte, 4)
	// Read payload size and payload until EOF
	for {
		n, err := f.Read(sizeBuf)
		if n == 0 {
			if err == io.EOF {
				break
			}
			return err
		}
		// Create proto struct to hold payload
		payload, err := images.ProtoHandler(img.Magic)
		if err != nil {
			return err
		}
		payloadSize := uint64(binary.LittleEndian.Uint32(sizeBuf))
		payloadBuf := make([]byte, payloadSize)
		if _, err := f.Read(payloadBuf); err != nil {
			return err
		}
		if err := proto.Unmarshal(payloadBuf, payload); err != nil {
			return err
		}
		entry := CriuEntry{Message: payload}
		if decodeExtra != nil {
			extraPayload, err := decodeExtra(f, payload, noPayload)
			if err != nil {
				return err
			}
			entry.Extra = extraPayload
		}
		img.Entries = append(img.Entries, &entry)
	}
	return nil
}

func (img *CriuImage) decodePagemap(f *os.File) error {
	var head bool = true
	sizeBuf := make([]byte, 4)
	// Read payload size and payload until EOF
	for {
		n, err := f.Read(sizeBuf)
		if n == 0 {
			if err == io.EOF {
				break
			}
			return err
		}
		// Create proto struct as pagemapHead for first entry
		// and as pagemapEntry for remaining
		var payload proto.Message
		if head {
			payload = &images.PagemapHead{}
			head = false
		} else {
			payload = &images.PagemapEntry{}
		}

		payloadSize := uint64(binary.LittleEndian.Uint32(sizeBuf))
		payloadBuf := make([]byte, payloadSize)
		if _, err := f.Read(payloadBuf); err != nil {
			return err
		}
		if err := proto.Unmarshal(payloadBuf, payload); err != nil {
			return err
		}
		entry := CriuEntry{Message: payload}
		img.Entries = append(img.Entries, &entry)
	}
	return nil
}

func (img *CriuImage) decodeGhostFile(f *os.File, noPayload bool) error {
	sizeBuf := make([]byte, 4)
	if _, err := f.Read(sizeBuf); err != nil {
		return err
	}
	// Create proto struct for primary entry
	payload := &images.GhostFileEntry{}
	payloadSize := uint64(binary.LittleEndian.Uint32(sizeBuf))
	payloadBuf := make([]byte, payloadSize)
	if _, err := f.Read(payloadBuf); err != nil {
		return err
	}
	if err := proto.Unmarshal(payloadBuf, payload); err != nil {
		return err
	}
	entry := CriuEntry{Message: payload}

	if payload.GetChunks() {
		img.Entries = append(img.Entries, &entry)
		for {
			n, err := f.Read(sizeBuf)
			if n == 0 {
				if err == io.EOF {
					break
				}
				return err
			}
			// Create proto struct for chunk
			payload := &images.GhostChunkEntry{}
			payloadSize := uint64(binary.LittleEndian.Uint32(sizeBuf))
			payloadBuf := make([]byte, payloadSize)
			if _, err := f.Read(payloadBuf); err != nil {
				return err
			}
			if err := proto.Unmarshal(payloadBuf, payload); err != nil {
				return err
			}
			entry = CriuEntry{Message: payload}
			if noPayload {
				if _, err = f.Seek(int64(payload.GetLen()), 1); err != nil {
					return err
				}
			} else {
				extraBuf := make([]byte, payload.GetLen())
				if _, err := f.Read(extraBuf); err != nil {
					return err
				}
				entry.Extra = base64.StdEncoding.EncodeToString(extraBuf)
			}
			img.Entries = append(img.Entries, &entry)
		}
	} else {
		if noPayload {
			if _, err := f.Seek(0, 2); err != nil {
				return err
			}
		} else {
			fInfo, err := f.Stat()
			if err != nil {
				return err
			}
			extraBuf := make([]byte, uint64(fInfo.Size())-4-payloadSize)
			if _, err := f.Read(extraBuf); err != nil {
				return err
			}
			entry.Extra = base64.StdEncoding.EncodeToString(extraBuf)
		}
		img.Entries = append(img.Entries, &entry)
	}
	return nil
}

// Function to count number of top-level entries
func countImg(f *os.File) (*CriuImage, error) {
	magicMap := magic.LoadMagic()
	img := CriuImage{}

	// Read magic
	buf := make([]byte, 4)
	if _, err := f.Read(buf); err != nil {
		return nil, err
	}
	magic := uint64(binary.LittleEndian.Uint32(buf))
	if magic == magicMap.ByName["IMG_COMMON"] ||
		magic == magicMap.ByName["IMG_SERVICE"] {
		if _, err := f.Read(buf); err != nil {
			return nil, err
		}
		magic = uint64(binary.LittleEndian.Uint32(buf))
	}

	// Identify magic
	img.Magic = magicMap.ByValue[magic]
	if img.Magic == "" {
		return nil, errors.New(fmt.Sprintf("Unknown magic 0x%x", magic))
	}

	count := 0
	sizeBuf := make([]byte, 4)
	// Read payload size and increment counter until EOF
	for {
		n, err := f.Read(sizeBuf)
		if n == 0 {
			if err == io.EOF {
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
