This directory contains scripts used by go-criu.

# `magicgen`
CRIU uses a 32-bit integer value to determine the type of image file. These values are called *magics*, and are defined [here](https://github.com/checkpoint-restore/criu/tree/master/criu/include/magic.h). `magicgen.go` processes this header file and generates `../magic/magic.go` with a map of magic names and values.

`Usage: magicgen.go /path/to/magic.h /path/to/magic.go`

A set of Makefile targets are provided for convenience:
- `make` or `make magic-gen`: Generate `../magic/magic.go`
- `make test`: Run unit test and E2E test for `magicgen.go`
- `make clean`: Remove `../magic/magic.go` and `magic.h`
