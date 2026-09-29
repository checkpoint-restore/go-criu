package criu

import (
	"fmt"
	"os"
	"os/exec"
	"slices"
	"syscall"
	"testing"
	"time"

	proto "github.com/checkpoint-restore/go-criu/v8/internal/proto"
	"github.com/checkpoint-restore/go-criu/v8/rpc"
)

func TestPluginOptions(t *testing.T) {
	operations := []struct {
		name string
		kind rpc.CriuReqType
		call func(*Criu, *rpc.CriuOpts, Notify) error
	}{
		{"dump", rpc.CriuReqType_DUMP, (*Criu).Dump},
		{"pre-dump", rpc.CriuReqType_PRE_DUMP, (*Criu).PreDump},
		{"restore", rpc.CriuReqType_RESTORE, (*Criu).Restore},
	}
	for _, operation := range operations {
		t.Run(operation.name, func(t *testing.T) {
			client, server := newPluginOptionsTestClient(t)
			cases := []struct {
				name    string
				options []string
				failure bool
			}{
				{"multiple", []string{"cuda_plugin.backend=cuda-checkpoint", "cuda_plugin.timeout=600"}, false},
				{"duplicates", []string{"cuda_plugin.timeout=600", "cuda_plugin.timeout=0"}, false},
				{"nil", nil, false},
				{"empty", []string{}, false},
				{"rejected", []string{"cuda_plugin.timeout=invalid"}, true},
			}
			if operation.kind == rpc.CriuReqType_RESTORE {
				cases[0].options = append(cases[0].options, "cuda_plugin.device-map=0=1,1=0")
			}
			// Reuse the same client for successive requests to catch option retention.
			for _, tc := range cases {
				ok := t.Run(tc.name, func(t *testing.T) {
					result := make(chan error, 1)
					go func() {
						err := servePluginOptions(server, operation.kind, tc.options, tc.failure)
						if err != nil {
							_ = server.Close() // Unblock the client if request handling fails.
						}
						result <- err
					}()
					err := operation.call(client, &rpc.CriuOpts{
						ImagesDirFd:   proto.Ptr[int32](-1),
						PluginOptions: slices.Clone(tc.options),
					}, nil)
					if serverErr := <-result; serverErr != nil {
						t.Fatal(serverErr)
					}
					if tc.failure {
						want := "operation failed (msg:invalid plugin option err:22)"
						if err == nil || err.Error() != want {
							t.Fatalf("operation error = %v, want %q", err, want)
						}
					} else if err != nil {
						t.Fatal(err)
					}
				})
				if !ok {
					// A failed exchange can leave the shared socket closed or
					// out of sync, so later cases would only report noise.
					break
				}
			}
		})
	}
}

func newPluginOptionsTestClient(t *testing.T) (*Criu, *os.File) {
	t.Helper()
	fds, err := syscall.Socketpair(syscall.AF_LOCAL, syscall.SOCK_SEQPACKET|syscall.SOCK_NONBLOCK, 0)
	if err != nil {
		t.Fatal(err)
	}
	client := os.NewFile(uintptr(fds[0]), "plugin-options-client")
	server := os.NewFile(uintptr(fds[1]), "plugin-options-server")
	t.Cleanup(func() {
		_ = client.Close()
		_ = server.Close()
	})
	// Bound both ends so a transport regression cannot hang the test suite.
	deadline := time.Now().Add(10 * time.Second)
	if err := client.SetDeadline(deadline); err != nil {
		t.Fatal(err)
	}
	if err := server.SetDeadline(deadline); err != nil {
		t.Fatal(err)
	}
	c := MakeCriu()
	c.swrkCmd = &exec.Cmd{}
	c.swrkSk = client
	return c, server
}

func servePluginOptions(server *os.File, kind rpc.CriuReqType, options []string, failure bool) error {
	requestBytes := make([]byte, 8192)
	n, err := server.Read(requestBytes)
	if err != nil {
		return fmt.Errorf("read plugin-options request: %w", err)
	}
	request := &rpc.CriuReq{}
	if err := request.UnmarshalVT(requestBytes[:n]); err != nil {
		return fmt.Errorf("unmarshal plugin-options request: %w", err)
	}
	if request.GetType() != kind {
		return fmt.Errorf("request type = %s, want %s", request.GetType(), kind)
	}
	if request.GetOpts() == nil {
		return fmt.Errorf("request does not contain options")
	}
	if got := request.GetOpts().GetPluginOptions(); !slices.Equal(got, options) {
		return fmt.Errorf("plugin options = %q, want %q", got, options)
	}
	response := &rpc.CriuResp{
		Type:    &kind,
		Success: proto.Ptr(!failure),
	}
	if failure {
		response.CrErrno = proto.Ptr[int32](22)
		response.CrErrmsg = proto.Ptr("invalid plugin option")
	}
	responseBytes, err := response.MarshalVT()
	if err != nil {
		return fmt.Errorf("marshal plugin-options response: %w", err)
	}
	if _, err := server.Write(responseBytes); err != nil {
		return fmt.Errorf("write plugin-options response: %w", err)
	}
	return nil
}
