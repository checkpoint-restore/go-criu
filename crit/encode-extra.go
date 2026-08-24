package crit

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"

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

// Extra payload handler for pipe and FIFO data
func decodeBase64WithSize(extra string, size uint64, description string) ([]byte, error) {
	payload, err := base64.StdEncoding.DecodeString(extra)
	if err != nil {
		return nil, err
	}
	if uint64(len(payload)) != size {
		return nil, fmt.Errorf("%s length %d does not match declared size %d", description, len(payload), size)
	}
	return payload, nil
}

func encodePipesData(payload proto.Message, extra string) ([]byte, error) {
	entry, ok := payload.(*pipe_data.PipeDataEntry)
	if !ok {
		return nil, errors.New("unable to assert pipe payload type")
	}
	return decodeBase64WithSize(extra, uint64(entry.GetBytes()), "pipe payload")
}

// Extra payload handler for socket queues
func encodeSkQueues(payload proto.Message, extra string) ([]byte, error) {
	entry, ok := payload.(*sk_packet.SkPacketEntry)
	if !ok {
		return nil, errors.New("unable to assert socket queue payload type")
	}
	return decodeBase64WithSize(extra, uint64(entry.GetLength()), "socket queue payload")
}

// Extra payload handler for TCP streams
func encodeTCPStream(payload proto.Message, extra string) ([]byte, error) {
	entry, ok := payload.(*tcp_stream.TcpStreamEntry)
	if !ok {
		return nil, errors.New("unable to assert TCP stream payload type")
	}
	extraPayload := tcpStreamExtra{}
	if err := json.Unmarshal([]byte(extra), &extraPayload); err != nil {
		return nil, err
	}

	inqBytes, err := base64.StdEncoding.DecodeString(extraPayload.InQ)
	if err != nil {
		return nil, err
	}
	if uint64(len(inqBytes)) != uint64(entry.GetInqLen()) {
		return nil, errors.New("TCP input queue length does not match inq_len")
	}
	outQBytes, err := base64.StdEncoding.DecodeString(extraPayload.OutQ)
	if err != nil {
		return nil, err
	}
	if uint64(len(outQBytes)) != uint64(entry.GetOutqLen()) {
		return nil, errors.New("TCP output queue length does not match outq_len")
	}

	return append(inqBytes, outQBytes...), nil
}

// Extra payload handler for BPF map data
func encodeBpfmapData(payload proto.Message, extra string) ([]byte, error) {
	entry, ok := payload.(*bpfmap_data.BpfmapDataEntry)
	if !ok {
		return nil, errors.New("unable to assert BPF map payload type")
	}
	size := uint64(entry.GetKeysBytes()) + uint64(entry.GetValuesBytes())
	return decodeBase64WithSize(extra, size, "BPF map payload")
}

// Extra payload handler for IPC semaphores
func encodeIpcSem(payload proto.Message, extra string) ([]byte, error) {
	entry, ok := payload.(*ipc_sem.IpcSemEntry)
	if !ok {
		return nil, errors.New("unable to assert IPC semaphore payload type")
	}
	extraEntries := []uint16{}
	if err := json.Unmarshal([]byte(extra), &extraEntries); err != nil {
		return nil, err
	}
	if uint64(len(extraEntries)) != uint64(entry.GetNsems()) {
		return nil, errors.New("IPC semaphore count does not match nsems")
	}
	extraPayload := []byte{}
	extraBuf := make([]byte, 2)

	for _, entry := range extraEntries {
		binary.LittleEndian.PutUint16(extraBuf, entry)
		extraPayload = append(extraPayload, extraBuf...)
	}
	// Each semaphore is 16-bit
	extraSize := uint64(len(extraEntries)) * 2
	// Round off to nearest 64-bit multiple
	roundedSize, err := alignPayloadSize(extraSize, 8)
	if err != nil {
		return nil, err
	}
	// Append zeroes for the remaining bytes
	extraPayload = append(extraPayload, make([]byte, int(roundedSize-extraSize))...)

	return extraPayload, nil
}

// Extra payload handler for IPC shared memory
func encodeIpcShm(payload proto.Message, extra string) ([]byte, error) {
	entry, ok := payload.(*ipc_shm.IpcShmEntry)
	if !ok {
		return nil, errors.New("unable to assert IPC shared memory payload type")
	}
	extraPayload, err := decodeBase64WithSize(extra, entry.GetSize(), "IPC shared memory payload")
	if err != nil {
		return nil, err
	}
	// Round off to nearest 32-bit multiple
	extraSize := uint64(len(extraPayload))
	roundedSize, err := alignPayloadSize(extraSize, 4)
	if err != nil {
		return nil, err
	}
	// Append zeroes for remaining bytes
	extraPayload = append(extraPayload, make([]byte, int(roundedSize-extraSize))...)

	return extraPayload, nil
}

// Extra payload handler for IPC messages
func encodeIpcMsg(payload proto.Message, extra string) ([]byte, error) {
	entry, ok := payload.(*ipc_msg.IpcMsgEntry)
	if !ok {
		return nil, errors.New("unable to assert IPC message payload type")
	}
	extraEntries := []string{}
	if err := json.Unmarshal([]byte(extra), &extraEntries); err != nil {
		return nil, err
	}
	if len(extraEntries)%2 != 0 {
		return nil, errors.New("IPC message payload must contain message and data pairs")
	}
	if uint64(len(extraEntries)/2) != uint64(entry.GetQnum()) {
		return nil, errors.New("IPC message count does not match qnum")
	}
	extraPayload := []byte{}
	sizeBuf := make([]byte, 4)

	for i := 0; i < len(extraEntries); i += 2 {
		msg := &ipc_msg.IpcMsg{}
		// Unmarshal JSON into proto struct
		if err := protojson.Unmarshal([]byte(extraEntries[i]), msg); err != nil {
			return nil, err
		}
		// Marshal proto struct into binary
		msgPayload, err := proto.Marshal(msg)
		if err != nil {
			return nil, err
		}
		if uint64(len(msgPayload)) > math.MaxUint32 {
			return nil, errors.New("IPC message metadata exceeds the image size field")
		}
		// Append size of message, followed by the message
		binary.LittleEndian.PutUint32(sizeBuf, uint32(len(msgPayload)))
		extraPayload = append(extraPayload, sizeBuf...)
		extraPayload = append(extraPayload, msgPayload...)
		// Append message data
		msgData, err := base64.StdEncoding.DecodeString(extraEntries[i+1])
		if err != nil {
			return nil, err
		}
		if uint64(len(msgData)) != uint64(msg.GetMsize()) {
			return nil, errors.New("IPC message data length does not match msize")
		}
		extraPayload = append(extraPayload, msgData...)

		msgSize := uint64(msg.GetMsize())
		// Round off to nearest 64-bit multiple
		roundedMsgSize, err := alignPayloadSize(msgSize, 8)
		if err != nil {
			return nil, err
		}
		// Append zeroes for remaining bytes
		extraPayload = append(extraPayload, make([]byte, int(roundedMsgSize-msgSize))...)
	}

	return extraPayload, nil
}
