//go:build unix

package cli

import (
	"context"
	"errors"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestTerminationSignalContextCancels(t *testing.T) {
	ctx, stop := terminationSignalContext(context.Background())
	defer stop()
	if err := syscall.Kill(os.Getpid(), syscall.SIGHUP); err != nil {
		t.Fatal(err)
	}
	select {
	case <-ctx.Done():
		if !errors.Is(ctx.Err(), context.Canceled) {
			t.Fatalf("signal context error = %v", ctx.Err())
		}
	case <-time.After(5 * time.Second):
		t.Fatal("termination signal did not cancel command context")
	}
}
