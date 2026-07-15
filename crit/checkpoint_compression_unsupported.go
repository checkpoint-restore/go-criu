//go:build !linux

package crit

import (
	"context"
	"errors"
	"fmt"
)

func checkpointCompressionUnsupported() error {
	return fmt.Errorf("checkpoint compression transforms are supported only on Linux: %w", errors.ErrUnsupported)
}

func compressCheckpoint(
	ctx context.Context,
	_ string,
	opts CompressOptions,
) (CheckpointCompressionResult, error) {
	var result CheckpointCompressionResult
	if ctx == nil {
		return result, errors.New("nil context")
	}
	acceleration := opts.Acceleration
	if acceleration == 0 {
		acceleration = 1
	}
	if acceleration != 1 {
		return result, fmt.Errorf("%w: %d (only 1 is supported)", ErrUnsupportedAcceleration, acceleration)
	}
	if err := ctx.Err(); err != nil {
		return result, err
	}
	return result, checkpointCompressionUnsupported()
}

func decompressCheckpoint(
	ctx context.Context,
	_ string,
	_ DecompressOptions,
) (CheckpointCompressionResult, error) {
	var result CheckpointCompressionResult
	if ctx == nil {
		return result, errors.New("nil context")
	}
	if err := ctx.Err(); err != nil {
		return result, err
	}
	return result, checkpointCompressionUnsupported()
}
