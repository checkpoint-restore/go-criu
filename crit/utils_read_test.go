package crit

import (
	"bytes"
	"errors"
	"io"
	"math"
	"os"
	"testing"
)

func TestReadExactBytesFromPipe(t *testing.T) {
	for _, test := range []struct {
		name    string
		data    []byte
		size    uint64
		wantErr error
	}{
		{name: "complete payload", data: bytes.Repeat([]byte{'x'}, int(maxImmediateImageAllocation)+1), size: maxImmediateImageAllocation + 1},
		// A corrupted size must fail on the short read instead of allocating it.
		{name: "oversized length", data: []byte("short"), size: uint64(math.MaxInt), wantErr: io.ErrUnexpectedEOF},
	} {
		t.Run(test.name, func(t *testing.T) {
			reader, writer, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = reader.Close() }()
			go func() {
				_, _ = writer.Write(test.data)
				_ = writer.Close()
			}()

			got, err := readExactBytes(reader, test.size)
			if test.wantErr != nil {
				if !errors.Is(err, test.wantErr) {
					t.Fatalf("readExactBytes() error = %v, want %v", err, test.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, test.data) {
				t.Fatal("readExactBytes() returned different data")
			}
		})
	}
}
