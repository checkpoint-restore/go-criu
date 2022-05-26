package crit

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/checkpoint-restore/go-criu/v5/crit/images"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

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

func loadPipesData(
	f *os.File,
	payload proto.Message,
	noPayload bool,
) (string, error) {
	p, ok := payload.(*images.PipeDataEntry)
	if !ok {
		return "", errors.New("Unable to assert payload type")
	}
	extraSize := p.GetBytes()

	if noPayload {
		f.Seek(int64(extraSize), 1)
		return countBytes(int64(extraSize)), nil
	}
	extraBuf := make([]byte, extraSize)
	if _, err := f.Read(extraBuf); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(extraBuf), nil
}

func loadSkQueues(
	f *os.File,
	payload proto.Message,
	noPayload bool,
) (string, error) {
	p, ok := payload.(*images.SkPacketEntry)
	if !ok {
		return "", errors.New("Unable to assert payload type")
	}
	extraSize := p.GetLength()

	if noPayload {
		f.Seek(int64(extraSize), 1)
		return countBytes(int64(extraSize)), nil
	}
	extraBuf := make([]byte, extraSize)
	if _, err := f.Read(extraBuf); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(extraBuf), nil
}

func loadTcpStream(
	f *os.File,
	payload proto.Message,
	noPayload bool,
) (string, error) {
	p, ok := payload.(*images.TcpStreamEntry)
	if !ok {
		return "", errors.New("Unable to assert payload type")
	}
	inqLen := p.GetInqLen()
	outqLen := p.GetOutqLen()

	if noPayload {
		f.Seek(0, 2)
		return countBytes(int64(inqLen + outqLen)), nil
	}
	extra := struct {
		InQ  string `json:"inQ"`
		OutQ string `json:"outQ"`
	}{}
	extraBuf := make([]byte, inqLen)
	if _, err := f.Read(extraBuf); err != nil {
		return "", err
	}
	extra.InQ = base64.StdEncoding.EncodeToString(extraBuf)
	extraBuf = make([]byte, outqLen)
	if _, err := f.Read(extraBuf); err != nil {
		return "", err
	}
	extra.OutQ = base64.StdEncoding.EncodeToString(extraBuf)
	extraJson, err := json.Marshal(extra)
	return string(extraJson), err
}

func loadBpfmapData(
	f *os.File,
	payload proto.Message,
	noPayload bool,
) (string, error) {
	p, ok := payload.(*images.BpfmapDataEntry)
	if !ok {
		return "", errors.New("Unable to assert payload type")
	}
	extraSize := p.GetKeysBytes() + p.GetValuesBytes()

	if noPayload {
		f.Seek(int64(extraSize), 1)
		return countBytes(int64(extraSize)), nil
	}
	extraBuf := make([]byte, extraSize)
	if _, err := f.Read(extraBuf); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(extraBuf), nil
}

func loadIpcSem(
	f *os.File,
	payload proto.Message,
	noPayload bool,
) (string, error) {
	p, ok := payload.(*images.IpcSemEntry)
	if !ok {
		return "", errors.New("Unable to assert payload type")
	}
	// Each semaphore is 16-bit
	extraSize := int64(p.GetNsems()) * 2
	// Round off to nearest 64-bit multiple
	roundedSize := (extraSize/8 + 1) * 8

	if noPayload {
		f.Seek(roundedSize, 1)
		return countBytes(extraSize), nil
	}
	extraPayload := []uint16{}
	for i := 0; i < int(extraSize/2); i++ {
		// Create 16-bit buffer
		extraBuf := make([]byte, 2)
		if _, err := f.Read(extraBuf); err != nil {
			return "", err
		}
		extraPayload = append(extraPayload, binary.LittleEndian.Uint16(extraBuf))
	}
	f.Seek(roundedSize-extraSize, 1)
	extraJson, err := json.Marshal(extraPayload)
	return string(extraJson), err
}

func loadIpcShm(
	f *os.File,
	payload proto.Message,
	noPayload bool,
) (string, error) {
	p, ok := payload.(*images.IpcShmEntry)
	if !ok {
		return "", errors.New("Unable to assert payload type")
	}
	extraSize := int64(p.GetSize())
	// Round off to nearest 32-bit multiple
	roundedSize := (extraSize/4 + 1) * 4

	if noPayload {
		f.Seek(roundedSize, 1)
		return countBytes(extraSize), nil
	}
	extraBuf := make([]byte, extraSize)
	if _, err := f.Read(extraBuf); err != nil {
		return "", err
	}
	f.Seek(roundedSize-extraSize, 1)
	return base64.StdEncoding.EncodeToString(extraBuf), nil
}

func loadIpcMsg(
	f *os.File,
	payload proto.Message,
	noPayload bool,
) (string, error) {
	p, ok := payload.(*images.IpcMsgEntry)
	if !ok {
		return "", errors.New("Unable to assert payload type")
	}
	msgQNum := int64(p.GetQnum())
	extraBuf := make([]byte, 4)
	// Store payload size if noPayload is true
	var totalSize int64 = 0
	// Store messages as string slice
	extraPayload := []string{}

	for i := 0; i < int(msgQNum); i++ {
		n, err := f.Read(extraBuf)
		if n == 0 {
			if err == io.EOF {
				break
			}
			return "", err
		}
		extraSize := uint64(binary.LittleEndian.Uint32(extraBuf))
		msgBuf := make([]byte, extraSize)
		if _, err = f.Read(msgBuf); err != nil {
			return "", err
		}
		msg := &images.IpcMsg{}
		if err = proto.Unmarshal(msgBuf, msg); err != nil {
			return "", err
		}
		msgSize := int64(msg.GetMsize())
		// Round off to nearest 64-bit multiple
		roundedMsgSize := (msgSize/8 + 1) * 8

		if noPayload {
			f.Seek(roundedMsgSize, 1)
			totalSize += int64(extraSize) + msgSize
		} else {
			jsonMsg, err := protojson.Marshal(msg)
			if err != nil {
				return "", err
			}
			extraPayload = append(extraPayload, string(jsonMsg))

			msgDataBuf := make([]byte, msgSize)
			if _, err = f.Read(msgDataBuf); err != nil {
				return "", err
			}
			f.Seek(roundedMsgSize-msgSize, 1)
			msgData := base64.StdEncoding.EncodeToString(msgDataBuf)
			extraPayload = append(extraPayload, msgData)
		}
	}

	if noPayload {
		return countBytes(totalSize), nil
	}
	extraJson, err := json.Marshal(extraPayload)
	if err != nil {
		return "", err
	}
	return string(extraJson), nil
}
