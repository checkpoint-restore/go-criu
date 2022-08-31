package crit

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"os"

	"github.com/checkpoint-restore/go-criu/v6/crit/images"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Extra data handler for pipe and FIFO data
func decodePipesData(
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
		_, err := f.Seek(int64(extraSize), 1)
		if err != nil {
			return "", err
		}
		return countBytes(int64(extraSize)), nil
	}
	extraBuf := make([]byte, extraSize)
	if _, err := f.Read(extraBuf); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(extraBuf), nil
}

// Extra data handler for socket queues
func decodeSkQueues(
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
		_, err := f.Seek(int64(extraSize), 1)
		if err != nil {
			return "", err
		}
		return countBytes(int64(extraSize)), nil
	}
	extraBuf := make([]byte, extraSize)
	if _, err := f.Read(extraBuf); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(extraBuf), nil
}

type tcpStreamExtra struct {
	InQ  string `json:"inQ"`
	OutQ string `json:"outQ"`
}

// Extra data handler for TCP streams
func decodeTcpStream(
	f *os.File,
	payload proto.Message,
	noPayload bool,
) (string, error) {
	p, ok := payload.(*images.TcpStreamEntry)
	if !ok {
		return "", errors.New("Unable to assert payload type")
	}
	inQLen := p.GetInqLen()
	outQLen := p.GetOutqLen()

	if noPayload {
		_, err := f.Seek(0, 2)
		if err != nil {
			return "", err
		}
		return countBytes(int64(inQLen + outQLen)), nil
	}

	extra := tcpStreamExtra{}
	extraBuf := make([]byte, inQLen)
	if _, err := f.Read(extraBuf); err != nil {
		return "", err
	}
	extra.InQ = base64.StdEncoding.EncodeToString(extraBuf)
	extraBuf = make([]byte, outQLen)
	if _, err := f.Read(extraBuf); err != nil {
		return "", err
	}
	extra.OutQ = base64.StdEncoding.EncodeToString(extraBuf)

	extraJson, err := json.Marshal(extra)
	return string(extraJson), err
}

// Extra data handler for BPF map data
func decodeBpfmapData(
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
		_, err := f.Seek(int64(extraSize), 1)
		if err != nil {
			return "", err
		}
		return countBytes(int64(extraSize)), nil
	}
	extraBuf := make([]byte, extraSize)
	if _, err := f.Read(extraBuf); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(extraBuf), nil
}

// Extra data handler for IPC semaphores
func decodeIpcSem(
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
		_, err := f.Seek(roundedSize, 1)
		if err != nil {
			return "", err
		}
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
	_, err := f.Seek(roundedSize-extraSize, 1)
	if err != nil {
		return "", err
	}
	extraJson, err := json.Marshal(extraPayload)
	return string(extraJson), err
}

// Extra data handler for IPC shared memory
func decodeIpcShm(
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
		_, err := f.Seek(roundedSize, 1)
		if err != nil {
			return "", err
		}
		return countBytes(extraSize), nil
	}
	extraBuf := make([]byte, extraSize)
	if _, err := f.Read(extraBuf); err != nil {
		return "", err
	}
	_, err := f.Seek(roundedSize-extraSize, 1)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(extraBuf), nil
}

// Extra data handler for IPC messages
func decodeIpcMsg(
	f *os.File,
	payload proto.Message,
	noPayload bool,
) (string, error) {
	p, ok := payload.(*images.IpcMsgEntry)
	if !ok {
		return "", errors.New("Unable to assert payload type")
	}
	msgQNum := int64(p.GetQnum())
	sizeBuf := make([]byte, 4)
	// Store payload size if noPayload is true
	var totalSize int64 = 0
	// Store messages as string slice
	extraPayload := []string{}

	for i := 0; i < int(msgQNum); i++ {
		n, err := f.Read(sizeBuf)
		if n == 0 {
			if errors.Is(err, io.EOF) {
				break
			}
			return "", err
		}
		extraSize := uint64(binary.LittleEndian.Uint32(sizeBuf))
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
			_, err = f.Seek(roundedMsgSize, 1)
			if err != nil {
				return "", err
			}
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
			msgData := base64.StdEncoding.EncodeToString(msgDataBuf)
			extraPayload = append(extraPayload, msgData)
			_, err = f.Seek(roundedMsgSize-msgSize, 1)
			if err != nil {
				return "", err
			}
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
