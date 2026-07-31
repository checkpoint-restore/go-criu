package crit

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"io"
	"os"

	ghost_file "github.com/checkpoint-restore/go-criu/v8/crit/images/ghost-file"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/pagemap"
	"google.golang.org/protobuf/proto"
)

type decodeExtraFunc func(*os.File, proto.Message, bool) (string, error)

type entryVisitor func(*CriuEntry)

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
	default:
		var decodeExtra decodeExtraFunc
		if _, decoder, ok := extraPayloadDecoderForMagic(img.Magic); ok {
			decodeExtra = decoder
		}
		err = img.decodeDefault(f, decodeExtra, noPayload)
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
	decodeExtra decodeExtraFunc,
	noPayload bool,
) error {
	return walkDefaultEntries(f, img.EntryType, decodeExtra, noPayload, func(entry *CriuEntry) {
		img.Entries = append(img.Entries, entry)
	})
}

func walkDefaultEntries(
	f *os.File,
	entryType proto.Message,
	decodeExtra decodeExtraFunc,
	noPayload bool,
	visit entryVisitor,
) error {
	sizeBuf := make([]byte, 4)
	// Read payload size and payload until EOF
	for {
		if _, err := io.ReadFull(f, sizeBuf); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		// Create proto struct to hold payload
		payload := proto.Clone(entryType)
		payloadSize := uint64(binary.LittleEndian.Uint32(sizeBuf))
		payloadBuf, err := readExactBytes(f, payloadSize)
		if err != nil {
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
		visit(&entry)
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
		if _, err := io.ReadFull(f, sizeBuf); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		payloadSize := uint64(binary.LittleEndian.Uint32(sizeBuf))
		payloadBuf, err := readExactBytes(f, payloadSize)
		if err != nil {
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
	return walkGhostFileEntries(f, noPayload, func(entry *CriuEntry) {
		img.Entries = append(img.Entries, entry)
	})
}

func walkGhostFileEntries(f *os.File, noPayload bool, visit entryVisitor) error {
	sizeBuf := make([]byte, 4)
	if _, err := io.ReadFull(f, sizeBuf); err != nil {
		return err
	}
	// Create proto struct for primary entry
	payload := &ghost_file.GhostFileEntry{}
	payloadSize := uint64(binary.LittleEndian.Uint32(sizeBuf))
	payloadBuf, err := readExactBytes(f, payloadSize)
	if err != nil {
		return err
	}
	if err := proto.Unmarshal(payloadBuf, payload); err != nil {
		return err
	}
	entry := &CriuEntry{Message: payload}

	if payload.GetChunks() {
		visit(entry)
		for {
			_, err := io.ReadFull(f, sizeBuf)
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return err
			}
			// Create proto struct for chunk
			payload := &ghost_file.GhostChunkEntry{}
			payloadSize := uint64(binary.LittleEndian.Uint32(sizeBuf))
			payloadBuf, err := readExactBytes(f, payloadSize)
			if err != nil {
				return err
			}
			if err := proto.Unmarshal(payloadBuf, payload); err != nil {
				return err
			}
			chunkEntry := &CriuEntry{Message: payload}
			if noPayload {
				if err = skipExactBytes(f, payload.GetLen()); err != nil {
					return err
				}
			} else {
				extraBuf, err := readExactBytes(f, payload.GetLen())
				if err != nil {
					return err
				}
				chunkEntry.Extra = base64.StdEncoding.EncodeToString(extraBuf)
			}
			visit(chunkEntry)
		}
	} else {
		if noPayload {
			// Seek to the end of the file
			if _, err := f.Seek(0, io.SeekEnd); err != nil {
				return err
			}
		} else {
			extraBuf, err := io.ReadAll(f)
			if err != nil {
				return err
			}
			entry.Extra = base64.StdEncoding.EncodeToString(extraBuf)
		}
		visit(entry)
	}
	return nil
}
