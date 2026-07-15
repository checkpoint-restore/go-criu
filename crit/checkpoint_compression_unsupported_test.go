//go:build !linux

package crit

import (
	"context"
	"errors"
	"testing"
)

func TestCheckpointCompressionUnsupported(t *testing.T) {
	for _, test := range []struct {
		name string
		run  func() error
	}{
		{
			name: "compress",
			run: func() error {
				_, err := CompressCheckpoint(context.Background(), t.TempDir(), CompressOptions{})
				return err
			},
		},
		{
			name: "decompress",
			run: func() error {
				_, err := DecompressCheckpoint(context.Background(), t.TempDir(), DecompressOptions{})
				return err
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := test.run(); !errors.Is(err, errors.ErrUnsupported) {
				t.Fatalf("error = %v, want errors.ErrUnsupported", err)
			}
		})
	}
}
