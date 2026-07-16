package criu

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"testing"

	proto "github.com/checkpoint-restore/go-criu/v8/internal/proto"
	"github.com/checkpoint-restore/go-criu/v8/rpc"
)

func TestFeatureCheckReturnsResponseFeatures(t *testing.T) {
	responseFeatures := &rpc.CriuFeatures{
		MemTrack:       proto.Ptr(false),
		MemCompression: proto.Ptr(true),
	}
	c, serverErr := newFeatureCheckTestClient(t, responseFeatures)
	requestedFeatures := &rpc.CriuFeatures{
		MemTrack:       proto.Ptr(true),
		MemCompression: proto.Ptr(true),
	}
	got, featureErr := c.FeatureCheck(requestedFeatures)
	if err := <-serverErr; err != nil {
		t.Fatal(err)
	}
	if featureErr != nil {
		t.Fatal(featureErr)
	}
	if !got.EqualVT(responseFeatures) {
		t.Fatalf("FeatureCheck() = %v, want response features %v", got, responseFeatures)
	}
}

func TestFeatureCheckRejectsMissingResponseFeatures(t *testing.T) {
	c, serverErr := newFeatureCheckTestClient(t, nil)
	_, featureErr := c.FeatureCheck(&rpc.CriuFeatures{MemCompression: proto.Ptr(true)})
	if err := <-serverErr; err != nil {
		t.Fatal(err)
	}
	if featureErr == nil || featureErr.Error() != "CRIU RPC response does not contain features" {
		t.Fatalf("FeatureCheck() error = %v, want missing-features error", featureErr)
	}
}

func newFeatureCheckTestClient(t *testing.T, responseFeatures *rpc.CriuFeatures) (*Criu, <-chan error) {
	t.Helper()
	fds, err := syscall.Socketpair(syscall.AF_LOCAL, syscall.SOCK_SEQPACKET, 0)
	if err != nil {
		t.Fatal(err)
	}
	client := os.NewFile(uintptr(fds[0]), "feature-check-client")
	server := os.NewFile(uintptr(fds[1]), "feature-check-server")
	t.Cleanup(func() { _ = client.Close() })

	serverErr := make(chan error, 1)
	go func() {
		defer func() { _ = server.Close() }()
		serverErr <- serveFeatureCheck(server, responseFeatures)
	}()

	c := MakeCriu()
	c.swrkCmd = &exec.Cmd{}
	c.swrkSk = client
	return c, serverErr
}

func serveFeatureCheck(server *os.File, responseFeatures *rpc.CriuFeatures) error {
	requestBytes := make([]byte, 8192)
	n, err := server.Read(requestBytes)
	if err != nil {
		return fmt.Errorf("read feature-check request: %w", err)
	}
	request := &rpc.CriuReq{}
	if err := request.UnmarshalVT(requestBytes[:n]); err != nil {
		return fmt.Errorf("unmarshal feature-check request: %w", err)
	}
	if request.GetType() != rpc.CriuReqType_FEATURE_CHECK {
		return fmt.Errorf("request type = %s, want FEATURE_CHECK", request.GetType())
	}
	if request.GetFeatures() == nil {
		return fmt.Errorf("feature-check request does not contain features")
	}

	response := &rpc.CriuResp{
		Type:     proto.Ptr(rpc.CriuReqType_FEATURE_CHECK),
		Success:  proto.Ptr(true),
		Features: responseFeatures,
	}
	responseBytes, err := response.MarshalVT()
	if err != nil {
		return fmt.Errorf("marshal feature-check response: %w", err)
	}
	if _, err := server.Write(responseBytes); err != nil {
		return fmt.Errorf("write feature-check response: %w", err)
	}
	return nil
}
