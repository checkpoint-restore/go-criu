package cli

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func terminationSignalContext(parent context.Context) (context.Context, context.CancelFunc) {
	return signal.NotifyContext(parent, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM)
}
