package crit

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	bpfmap_data "github.com/checkpoint-restore/go-criu/v8/crit/images/bpfmap-data"
	ipc_desc "github.com/checkpoint-restore/go-criu/v8/crit/images/ipc-desc"
	ipc_msg "github.com/checkpoint-restore/go-criu/v8/crit/images/ipc-msg"
	ipc_sem "github.com/checkpoint-restore/go-criu/v8/crit/images/ipc-sem"
	ipc_shm "github.com/checkpoint-restore/go-criu/v8/crit/images/ipc-shm"
	pipe_data "github.com/checkpoint-restore/go-criu/v8/crit/images/pipe-data"
	sk_packet "github.com/checkpoint-restore/go-criu/v8/crit/images/sk-packet"
	tcp_stream "github.com/checkpoint-restore/go-criu/v8/crit/images/tcp-stream"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestExtraPayloadImageTypesRoundTripAndInfo(t *testing.T) {
	desc := &ipc_desc.IpcDescEntry{
		Key: proto.Uint32(1), Uid: proto.Uint32(2), Gid: proto.Uint32(3),
		Cuid: proto.Uint32(4), Cgid: proto.Uint32(5), Mode: proto.Uint32(0o600),
		Id: proto.Uint32(6),
	}
	message := &ipc_msg.IpcMsg{Mtype: proto.Uint64(1), Msize: proto.Uint32(3)}
	messageJSON, err := protojson.Marshal(message)
	if err != nil {
		t.Fatal(err)
	}
	messageExtra, err := json.Marshal([]string{
		string(messageJSON), base64.StdEncoding.EncodeToString([]byte("msg")),
	})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name  string
		magic string
		entry *CriuEntry
	}{
		{
			name: "pipe", magic: "PIPES_DATA",
			entry: &CriuEntry{
				Message: &pipe_data.PipeDataEntry{PipeId: proto.Uint32(1), Bytes: proto.Uint32(3)},
				Extra:   base64.StdEncoding.EncodeToString([]byte("abc")),
			},
		},
		{
			name: "FIFO", magic: "FIFO_DATA",
			entry: &CriuEntry{
				Message: &pipe_data.PipeDataEntry{PipeId: proto.Uint32(1), Bytes: proto.Uint32(3)},
				Extra:   base64.StdEncoding.EncodeToString([]byte("abc")),
			},
		},
		{
			name: "socket queue", magic: "SK_QUEUES",
			entry: &CriuEntry{
				Message: &sk_packet.SkPacketEntry{IdFor: proto.Uint32(1), Length: proto.Uint32(3)},
				Extra:   base64.StdEncoding.EncodeToString([]byte("abc")),
			},
		},
		{
			name: "TCP stream", magic: "TCP_STREAM",
			entry: &CriuEntry{
				Message: &tcp_stream.TcpStreamEntry{
					InqLen: proto.Uint32(2), InqSeq: proto.Uint32(0),
					OutqLen: proto.Uint32(1), OutqSeq: proto.Uint32(0),
					OptMask: proto.Uint32(0), SndWscale: proto.Uint32(0),
					MssClamp: proto.Uint32(0),
				},
				Extra: `{"in_q":"YWI=","out_q":"Yw=="}`,
			},
		},
		{
			name: "BPF map", magic: "BPFMAP_DATA",
			entry: &CriuEntry{
				Message: &bpfmap_data.BpfmapDataEntry{
					MapId: proto.Uint32(1), KeysBytes: proto.Uint32(1),
					ValuesBytes: proto.Uint32(2), Count: proto.Uint32(1),
				},
				Extra: base64.StdEncoding.EncodeToString([]byte("abc")),
			},
		},
		{
			name: "IPC semaphore unaligned", magic: "IPCNS_SEM",
			entry: &CriuEntry{
				Message: &ipc_sem.IpcSemEntry{Desc: desc, Nsems: proto.Uint32(3)},
				Extra:   `[1,2,3]`,
			},
		},
		{
			name: "IPC semaphore aligned", magic: "IPCNS_SEM",
			entry: &CriuEntry{
				Message: &ipc_sem.IpcSemEntry{Desc: desc, Nsems: proto.Uint32(4)},
				Extra:   `[1,2,3,4]`,
			},
		},
		{
			name: "IPC shared memory unaligned", magic: "IPCNS_SHM",
			entry: &CriuEntry{
				Message: &ipc_shm.IpcShmEntry{Desc: desc, Size: proto.Uint64(3)},
				Extra:   base64.StdEncoding.EncodeToString([]byte("abc")),
			},
		},
		{
			name: "IPC shared memory aligned", magic: "IPCNS_SHM",
			entry: &CriuEntry{
				Message: &ipc_shm.IpcShmEntry{Desc: desc, Size: proto.Uint64(4)},
				Extra:   base64.StdEncoding.EncodeToString([]byte("abcd")),
			},
		},
		{
			name: "IPC message", magic: "IPCNS_MSG",
			entry: &CriuEntry{
				Message: &ipc_msg.IpcMsgEntry{
					Desc: desc, Qbytes: proto.Uint32(3), Qnum: proto.Uint32(1),
				},
				Extra: string(messageExtra),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "image.img")
			file, err := os.Create(path)
			if err != nil {
				t.Fatal(err)
			}
			image := &CriuImage{
				Magic: test.magic,
				Entries: []*CriuEntry{
					test.entry,
					{Message: proto.Clone(test.entry.Message), Extra: test.entry.Extra},
				},
			}
			if err := New(nil, file, "", false, false).Encode(image); err != nil {
				_ = file.Close()
				t.Fatal(err)
			}
			if err := file.Close(); err != nil {
				t.Fatal(err)
			}

			file, err = os.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			counted, countErr := New(file, nil, "", false, false).Info()
			closeErr := file.Close()
			if countErr != nil {
				t.Fatal(countErr)
			}
			if closeErr != nil {
				t.Fatal(closeErr)
			}
			if got := counted.Entries[0].Extra; got != "2" {
				t.Fatalf("Info() count = %s, want 2", got)
			}

			file, err = os.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			decoded, decodeErr := New(file, nil, "", false, false).Decode(proto.Clone(test.entry.Message))
			closeErr = file.Close()
			if decodeErr != nil {
				t.Fatal(decodeErr)
			}
			if closeErr != nil {
				t.Fatal(closeErr)
			}
			if len(decoded.Entries) != len(image.Entries) {
				t.Fatalf("Decode() entries = %d, want %d", len(decoded.Entries), len(image.Entries))
			}
			for index, got := range decoded.Entries {
				want := image.Entries[index]
				if !proto.Equal(got.Message, want.Message) {
					t.Errorf("Decode() entry %d = %+v, want %+v", index, got.Message, want.Message)
				}
				if !equivalentExtraPayload(got.Extra, want.Extra) {
					t.Errorf("Decode() extra %d = %q, want %q", index, got.Extra, want.Extra)
				}
			}
		})
	}
}

func equivalentExtraPayload(got, want string) bool {
	var gotJSON any
	var wantJSON any
	if err := json.Unmarshal([]byte(want), &wantJSON); err != nil {
		return got == want
	}
	if err := json.Unmarshal([]byte(got), &gotJSON); err != nil {
		return false
	}
	return reflect.DeepEqual(gotJSON, wantJSON)
}

func TestEncodeIPCAlignment(t *testing.T) {
	for _, test := range []struct {
		name   string
		encode func(proto.Message, string) ([]byte, error)
		entry  proto.Message
		extra  string
		want   int
	}{
		{name: "semaphore aligned", encode: encodeIpcSem, entry: &ipc_sem.IpcSemEntry{Nsems: proto.Uint32(4)}, extra: `[1,2,3,4]`, want: 8},
		{name: "semaphore unaligned", encode: encodeIpcSem, entry: &ipc_sem.IpcSemEntry{Nsems: proto.Uint32(3)}, extra: `[1,2,3]`, want: 8},
		{name: "shared memory aligned", encode: encodeIpcShm, entry: &ipc_shm.IpcShmEntry{Size: proto.Uint64(4)}, extra: base64.StdEncoding.EncodeToString([]byte{1, 2, 3, 4}), want: 4},
		{name: "shared memory unaligned", encode: encodeIpcShm, entry: &ipc_shm.IpcShmEntry{Size: proto.Uint64(3)}, extra: base64.StdEncoding.EncodeToString([]byte{1, 2, 3}), want: 4},
	} {
		t.Run(test.name, func(t *testing.T) {
			payload, err := test.encode(test.entry, test.extra)
			if err != nil {
				t.Fatal(err)
			}
			if len(payload) != test.want {
				t.Fatalf("encoded payload length = %d, want %d", len(payload), test.want)
			}
		})
	}
}

func TestIPCMessageExtraRoundTrip(t *testing.T) {
	messages := []*ipc_msg.IpcMsg{
		{Mtype: proto.Uint64(1), Msize: proto.Uint32(8)},
		{Mtype: proto.Uint64(2), Msize: proto.Uint32(3)},
	}
	data := [][]byte{[]byte("12345678"), []byte("xyz")}
	extraEntries := make([]string, 0, 2*len(messages))
	for index, message := range messages {
		metadata, err := protojson.Marshal(message)
		if err != nil {
			t.Fatal(err)
		}
		extraEntries = append(
			extraEntries,
			string(metadata),
			base64.StdEncoding.EncodeToString(data[index]),
		)
	}
	extra, err := json.Marshal(extraEntries)
	if err != nil {
		t.Fatal(err)
	}
	outer := &ipc_msg.IpcMsgEntry{Qnum: proto.Uint32(uint32(len(messages)))}
	payload, err := encodeIpcMsg(outer, string(extra))
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "ipc-msg-extra")
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = file.Close() }()
	decoded, err := decodeIpcMsg(file, outer, false)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	if err := json.Unmarshal([]byte(decoded), &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != len(extraEntries) {
		t.Fatalf("decoded entries = %d, want %d", len(got), len(extraEntries))
	}
	for index, want := range data {
		decodedData, err := base64.StdEncoding.DecodeString(got[2*index+1])
		if err != nil {
			t.Fatal(err)
		}
		if string(decodedData) != string(want) {
			t.Fatalf("decoded message %d data = %q, want %q", index, decodedData, want)
		}
	}
}

func TestEncodeIPCMessageRejectsMalformedPairs(t *testing.T) {
	message := &ipc_msg.IpcMsg{Mtype: proto.Uint64(1), Msize: proto.Uint32(2)}
	metadata, err := protojson.Marshal(message)
	if err != nil {
		t.Fatal(err)
	}
	validPair := []string{string(metadata), base64.StdEncoding.EncodeToString([]byte{1, 2})}
	for _, test := range []struct {
		name    string
		qnum    uint32
		entries []string
	}{
		{name: "missing data", qnum: 1, entries: []string{string(metadata)}},
		{name: "wrong data length", qnum: 1, entries: []string{string(metadata), base64.StdEncoding.EncodeToString([]byte{1})}},
		{name: "wrong message count", qnum: 2, entries: validPair},
	} {
		t.Run(test.name, func(t *testing.T) {
			extra, err := json.Marshal(test.entries)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := encodeIpcMsg(&ipc_msg.IpcMsgEntry{Qnum: proto.Uint32(test.qnum)}, string(extra)); err == nil {
				t.Fatal("encodeIpcMsg() accepted malformed message/data pairs")
			}
		})
	}
}

func TestEncodeExtraRejectsDeclaredLengthMismatch(t *testing.T) {
	for _, test := range []struct {
		name   string
		encode func(proto.Message, string) ([]byte, error)
		entry  proto.Message
		extra  string
	}{
		{
			name: "pipe", encode: encodePipesData,
			entry: &pipe_data.PipeDataEntry{Bytes: proto.Uint32(2)},
			extra: base64.StdEncoding.EncodeToString([]byte{1}),
		},
		{
			name: "socket queue", encode: encodeSkQueues,
			entry: &sk_packet.SkPacketEntry{Length: proto.Uint32(2)},
			extra: base64.StdEncoding.EncodeToString([]byte{1}),
		},
		{
			name: "TCP stream", encode: encodeTCPStream,
			entry: &tcp_stream.TcpStreamEntry{InqLen: proto.Uint32(2), OutqLen: proto.Uint32(0)},
			extra: `{"in_q":"AQ==","out_q":""}`,
		},
		{
			name: "BPF map", encode: encodeBpfmapData,
			entry: &bpfmap_data.BpfmapDataEntry{KeysBytes: proto.Uint32(1), ValuesBytes: proto.Uint32(1)},
			extra: base64.StdEncoding.EncodeToString([]byte{1}),
		},
		{
			name: "IPC semaphore", encode: encodeIpcSem,
			entry: &ipc_sem.IpcSemEntry{Nsems: proto.Uint32(2)},
			extra: `[1]`,
		},
		{
			name: "IPC shared memory", encode: encodeIpcShm,
			entry: &ipc_shm.IpcShmEntry{Size: proto.Uint64(2)},
			extra: base64.StdEncoding.EncodeToString([]byte{1}),
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := test.encode(test.entry, test.extra); err == nil {
				t.Fatal("extra payload encoder accepted a declared-length mismatch")
			}
		})
	}
}

func TestStructuredExtraJSONRoundTrip(t *testing.T) {
	want := `{"in_q":"AQI=","out_q":"Aw=="}`
	entry := &CriuEntry{
		Message: &pipe_data.PipeDataEntry{PipeId: proto.Uint32(1), Bytes: proto.Uint32(3)},
		Extra:   want,
	}
	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(data) {
		t.Fatalf("MarshalJSON() produced invalid JSON: %s", data)
	}
	payload, got, err := splitJSONData(data)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("extra payload = %q, want %q", got, want)
	}
	decoded := &pipe_data.PipeDataEntry{}
	if err := protojson.Unmarshal(payload, decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.GetPipeId() != 1 || decoded.GetBytes() != 3 {
		t.Fatalf("decoded protobuf = %+v", decoded)
	}
}
