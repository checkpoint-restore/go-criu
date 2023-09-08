package crit

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"io"
	"os"

	ghost_file "github.com/checkpoint-restore/go-criu/v7/crit/images/ghost-file"
	"github.com/checkpoint-restore/go-criu/v7/crit/images/pagemap"
	"google.golang.org/protobuf/proto"
)

// decodeImg identifies the type of image file
// and calls the appropriate decode handler
func decodeImg(f *os.File, entryType proto.Message, noPayload bool) (*CriuImage, error) {
	img := CriuImage{EntryType: entryType}
	var err error

	// Identify magic
	if img.Magic, err = ReadMagic(f); err != nil {
		return nil, err
	}

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
		err = img.decodeDefault(f, decodeTCPStream, noPayload)
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

// decodeDefault is used for all image files
// that are in the standard protobuf format
func (img *CriuImage) decodeDefault(
	f *os.File,
	decodeExtra func(*os.File, proto.Message, bool) (string, error),
	noPayload bool,
) error {
	sizeBuf := make([]byte, 4)
	// Read payload size and payload until EOF
	for {
		if n, err := f.Read(sizeBuf); err != nil {
			if n == 0 && err == io.EOF {
				break
			}
			return err
		}
		// Create proto struct to hold payload
		payload := proto.Clone(img.EntryType)
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

// Special handler for pagemap image
func (img *CriuImage) decodePagemap(f *os.File) error {
	sizeBuf := make([]byte, 4)
	// First entry is pagemap head
	var payload proto.Message = &pagemap.PagemapHead{}
	// Read payload size and payload until EOF
	for {
		if n, err := f.Read(sizeBuf); err != nil {
			if n == 0 && err == io.EOF {
				break
			}
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
		img.Entries = append(img.Entries, &entry)
		// Create struct for next entry
		payload = &pagemap.PagemapEntry{}
	}
	return nil
}

// Special handler for ghost image
func (img *CriuImage) decodeGhostFile(f *os.File, noPayload bool) error {
	sizeBuf := make([]byte, 4)
	if _, err := f.Read(sizeBuf); err != nil {
		return err
	}
	// Create proto struct for primary entry
	payload := &ghost_file.GhostFileEntry{}
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
				if errors.Is(err, io.EOF) {
					break
				}
				return err
			}
			// Create proto struct for chunk
			payload := &ghost_file.GhostChunkEntry{}
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
			// Seek to the end of the file
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
