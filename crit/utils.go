package crit

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/checkpoint-restore/go-criu/v5/magic"
)

// Helper to decode magic name from hex value
func readMagic(f *os.File) (string, error) {
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
		return "", errors.New(fmt.Sprintf("Unknown magic 0x%x", magicValue))
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
	if img.Magic, err = readMagic(f); err != nil {
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
