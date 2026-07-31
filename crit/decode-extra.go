package crit

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"math"
	"os"

	bpfmap_data "github.com/checkpoint-restore/go-criu/v8/crit/images/bpfmap-data"
	ipc_msg "github.com/checkpoint-restore/go-criu/v8/crit/images/ipc-msg"
	ipc_sem "github.com/checkpoint-restore/go-criu/v8/crit/images/ipc-sem"
	ipc_shm "github.com/checkpoint-restore/go-criu/v8/crit/images/ipc-shm"
	pipe_data "github.com/checkpoint-restore/go-criu/v8/crit/images/pipe-data"
	sk_packet "github.com/checkpoint-restore/go-criu/v8/crit/images/sk-packet"
	tcp_stream "github.com/checkpoint-restore/go-criu/v8/crit/images/tcp-stream"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Extra data handler for pipe and FIFO data
func decodePipesData(
	f *os.File,
	payload proto.Message,
	noPayload bool,
) (string, error) {
	p, ok := payload.(*pipe_data.PipeDataEntry)
	if !ok {
		return "", errors.New("unable to assert payload type")
	}
	extraSize := p.GetBytes()

	if noPayload {
		if err := skipExactBytes(f, uint64(extraSize)); err != nil {
			return "", err
		}
		return countBytes(int64(extraSize)), nil
	}
	extraBuf, err := readExactBytes(f, uint64(extraSize))
	if err != nil {
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
	p, ok := payload.(*sk_packet.SkPacketEntry)
	if !ok {
		return "", errors.New("unable to assert payload type")
	}
	extraSize := p.GetLength()

	if noPayload {
		if err := skipExactBytes(f, uint64(extraSize)); err != nil {
			return "", err
		}
		return countBytes(int64(extraSize)), nil
	}
	extraBuf, err := readExactBytes(f, uint64(extraSize))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(extraBuf), nil
}

type tcpStreamExtra struct {
	InQ  string `json:"in_q"`
	OutQ string `json:"out_q"`
}

func extraPayloadDecoderForMagic(
	magic string,
) (proto.Message, decodeExtraFunc, bool) {
	switch magic {
	case "PIPES_DATA", "FIFO_DATA":
		return &pipe_data.PipeDataEntry{}, decodePipesData, true
	case "SK_QUEUES":
		return &sk_packet.SkPacketEntry{}, decodeSkQueues, true
	case "TCP_STREAM":
		return &tcp_stream.TcpStreamEntry{}, decodeTCPStream, true
	case "BPFMAP_DATA":
		return &bpfmap_data.BpfmapDataEntry{}, decodeBpfmapData, true
	case "IPCNS_SEM":
		return &ipc_sem.IpcSemEntry{}, decodeIpcSem, true
	case "IPCNS_SHM":
		return &ipc_shm.IpcShmEntry{}, decodeIpcShm, true
	case "IPCNS_MSG":
		return &ipc_msg.IpcMsgEntry{}, decodeIpcMsg, true
	default:
		return nil, nil, false
	}
}

// Extra data handler for TCP streams
func decodeTCPStream(
	f *os.File,
	payload proto.Message,
	noPayload bool,
) (string, error) {
	p, ok := payload.(*tcp_stream.TcpStreamEntry)
	if !ok {
		return "", errors.New("unable to assert payload type")
	}
	inQLen := p.GetInqLen()
	outQLen := p.GetOutqLen()
	extraSize := uint64(inQLen) + uint64(outQLen)

	if noPayload {
		if err := skipExactBytes(f, extraSize); err != nil {
			return "", err
		}
		return countBytes(int64(extraSize)), nil
	}

	extra := tcpStreamExtra{}
	extraBuf, err := readExactBytes(f, uint64(inQLen))
	if err != nil {
		return "", err
	}
	extra.InQ = base64.StdEncoding.EncodeToString(extraBuf)
	extraBuf, err = readExactBytes(f, uint64(outQLen))
	if err != nil {
		return "", err
	}
	extra.OutQ = base64.StdEncoding.EncodeToString(extraBuf)

	extraJSON, err := json.Marshal(extra)
	return string(extraJSON), err
}

// Extra data handler for BPF map data
func decodeBpfmapData(
	f *os.File,
	payload proto.Message,
	noPayload bool,
) (string, error) {
	p, ok := payload.(*bpfmap_data.BpfmapDataEntry)
	if !ok {
		return "", errors.New("unable to assert payload type")
	}
	extraSize := uint64(p.GetKeysBytes()) + uint64(p.GetValuesBytes())

	if noPayload {
		if err := skipExactBytes(f, extraSize); err != nil {
			return "", err
		}
		return countBytes(int64(extraSize)), nil
	}
	extraBuf, err := readExactBytes(f, extraSize)
	if err != nil {
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
	p, ok := payload.(*ipc_sem.IpcSemEntry)
	if !ok {
		return "", errors.New("unable to assert payload type")
	}
	// Each semaphore is 16-bit
	extraSize := uint64(p.GetNsems()) * 2
	// Round off to nearest 64-bit multiple
	roundedSize, err := alignPayloadSize(extraSize, 8)
	if err != nil {
		return "", err
	}

	if noPayload {
		if err := skipExactBytes(f, roundedSize); err != nil {
			return "", err
		}
		return countBytes(int64(extraSize)), nil
	}
	extraBuf, err := readExactBytes(f, extraSize)
	if err != nil {
		return "", err
	}
	extraPayload := make([]uint16, p.GetNsems())
	for i := range extraPayload {
		offset := i * 2
		extraPayload[i] = binary.LittleEndian.Uint16(extraBuf[offset : offset+2])
	}
	if err := skipExactBytes(f, roundedSize-extraSize); err != nil {
		return "", err
	}
	extraJSON, err := json.Marshal(extraPayload)
	return string(extraJSON), err
}

// Extra data handler for IPC shared memory
func decodeIpcShm(
	f *os.File,
	payload proto.Message,
	noPayload bool,
) (string, error) {
	p, ok := payload.(*ipc_shm.IpcShmEntry)
	if !ok {
		return "", errors.New("unable to assert payload type")
	}
	extraSize := p.GetSize()
	// Round off to nearest 32-bit multiple
	roundedSize, err := alignPayloadSize(extraSize, 4)
	if err != nil {
		return "", err
	}

	if noPayload {
		if err := skipExactBytes(f, roundedSize); err != nil {
			return "", err
		}
		return countBytes(int64(extraSize)), nil
	}
	extraBuf, err := readExactBytes(f, extraSize)
	if err != nil {
		return "", err
	}
	if err := skipExactBytes(f, roundedSize-extraSize); err != nil {
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
	p, ok := payload.(*ipc_msg.IpcMsgEntry)
	if !ok {
		return "", errors.New("unable to assert payload type")
	}
	// Store payload size if noPayload is true
	var totalSize uint64
	// Store messages as string slice
	extraPayload := []string{}

	for i := uint32(0); i < p.GetQnum(); i++ {
		sizeBuf, err := readExactBytes(f, 4)
		if err != nil {
			return "", err
		}
		extraSize := uint64(binary.LittleEndian.Uint32(sizeBuf))
		msgBuf, err := readExactBytes(f, extraSize)
		if err != nil {
			return "", err
		}
		msg := &ipc_msg.IpcMsg{}
		if err = proto.Unmarshal(msgBuf, msg); err != nil {
			return "", err
		}
		msgSize := uint64(msg.GetMsize())
		// Round off to nearest 64-bit multiple
		roundedMsgSize, err := alignPayloadSize(msgSize, 8)
		if err != nil {
			return "", err
		}

		if noPayload {
			if err = skipExactBytes(f, roundedMsgSize); err != nil {
				return "", err
			}
			if totalSize > math.MaxUint64-extraSize-msgSize {
				return "", errors.New("IPC message payload size overflows")
			}
			totalSize += extraSize + msgSize
		} else {
			jsonMsg, err := protojson.Marshal(msg)
			if err != nil {
				return "", err
			}
			extraPayload = append(extraPayload, string(jsonMsg))

			msgDataBuf, readErr := readExactBytes(f, msgSize)
			if readErr != nil {
				return "", readErr
			}
			msgData := base64.StdEncoding.EncodeToString(msgDataBuf)
			extraPayload = append(extraPayload, msgData)
			if err = skipExactBytes(f, roundedMsgSize-msgSize); err != nil {
				return "", err
			}
		}
	}

	if noPayload {
		if totalSize > math.MaxInt64 {
			return "", errors.New("IPC message payload size exceeds reporting range")
		}
		return countBytes(int64(totalSize)), nil
	}
	extraJSON, err := json.Marshal(extraPayload)
	if err != nil {
		return "", err
	}
	return string(extraJSON), nil
}
