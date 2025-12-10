package crit

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"

	ipc_msg "github.com/checkpoint-restore/go-criu/v8/crit/images/ipc-msg"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Extra payload handler for pipe and FIFO data
func encodePipesData(extra string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(extra)
}

// Extra payload handler for socket queues
func encodeSkQueues(extra string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(extra)
}

// Extra payload handler for TCP streams
func encodeTCPStream(extra string) ([]byte, error) {
	extraPayload := tcpStreamExtra{}
	if err := json.Unmarshal([]byte(extra), &extraPayload); err != nil {
		return nil, err
	}

	inqBytes, err := base64.StdEncoding.DecodeString(extraPayload.InQ)
	if err != nil {
		return nil, err
	}
	outQBytes, err := base64.StdEncoding.DecodeString(extraPayload.OutQ)
	if err != nil {
		return nil, err
	}

	return append(inqBytes, outQBytes...), nil
}

// Extra payload handler for BPF map data
func encodeBpfmapData(extra string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(extra)
}

// Extra payload handler for IPC semaphores
func encodeIpcSem(extra string) ([]byte, error) {
	extraEntries := []uint16{}
	if err := json.Unmarshal([]byte(extra), &extraEntries); err != nil {
		return nil, err
	}
	extraPayload := []byte{}
	extraBuf := make([]byte, 2)

	for _, entry := range extraEntries {
		binary.LittleEndian.PutUint16(extraBuf, entry)
		extraPayload = append(extraPayload, extraBuf...)
	}
	// Each semaphore is 16-bit
	extraSize := len(extraEntries) * 2
	// Round off to nearest 64-bit multiple
	roundedSize := (extraSize/8 + 1) * 8
	// Append zeroes for the remaining bytes
	extraPayload = append(extraPayload, make([]byte, roundedSize-extraSize)...)

	return extraPayload, nil
}

// Extra payload handler for IPC shared memory
func encodeIpcShm(extra string) ([]byte, error) {
	extraPayload, err := base64.StdEncoding.DecodeString(extra)
	if err != nil {
		return nil, err
	}
	// Round off to nearest 32-bit multiple
	roundedSize := len(extraPayload)
	// Append zeroes for remaining bytes
	extraPayload = append(extraPayload, make([]byte, roundedSize-len(extraPayload))...)

	return extraPayload, nil
}

// Extra payload handler for IPC messages
func encodeIpcMsg(extra string) ([]byte, error) {
	extraEntries := []string{}
	if err := json.Unmarshal([]byte(extra), &extraEntries); err != nil {
		return nil, err
	}
	extraPayload := []byte{}
	sizeBuf := make([]byte, 4)

	for i := 0; i < len(extraEntries)/2; i++ {
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
		// Append size of message, followed by the message
		binary.LittleEndian.PutUint32(sizeBuf, uint32(len(msgPayload)))
		extraPayload = append(extraPayload, sizeBuf...)
		extraPayload = append(extraPayload, msgPayload...)
		// Append message data
		msgData, err := base64.StdEncoding.DecodeString(extraEntries[i+1])
		if err != nil {
			return nil, err
		}
		extraPayload = append(extraPayload, msgData...)

		msgSize := int64(msg.GetMsize())
		// Round off to nearest 64-bit multiple
		roundedMsgSize := (msgSize/8 + 1) * 8
		// Append zeroes for remaining bytes
		extraPayload = append(extraPayload, make([]byte, roundedMsgSize-msgSize)...)
	}

	return extraPayload, nil
}
