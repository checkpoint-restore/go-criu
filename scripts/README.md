This directory provides scripts used by go-criu. Each script is contained in its own directory, along with the respective Makefile and tests.

## `magic-gen`
CRIU uses a 32-bit integer value to determine the type of image file. These values are called *magics*, and are defined [here](https://github.com/checkpoint-restore/criu/tree/master/criu/include/magic.h). `magicgen.go` processes this header file and generates `magic.go` with a map of magic names and values. If the destination path is not provided, the script uses `../magic/magic.go` by default.

`Usage: magicgen.go /path/to/magic.h /path/to/magic.go`

Makefile targets provided:

- `make` or `make magic-gen`: Generate `../magic/magic.go`
- `make test`: Run unit test and E2E test for `magicgen.go`
- `make clean`: Remove `../magic/magic.go` and `magic.h`

## `proto-gen`
CRIT uses protobuf bindings to serialise and deserialise CRIU image files. The definitions for these bindings are available in [criu/images](https://github.com/checkpoint-restore/criu/tree/master/images). `protogen.py` fetches these definitions and generates the `.pb.go` files used by CRIT. If the source or destination directory is not provided, the script uses `../crit/images` by default.

`Usage: protogen.py /path/to/definitions/dir /path/to/destination/dir`

Makefile targets provided:

- `make` or `make pb-gen`: Generate the `.pb.go` bindings
- `make proto-update`: Update the `.proto` files with the latest copy from the CRIU repo. The GIT_BRANCH variable can be used to specify the branch to fetch the files from.
- `make clean-proto`: Delete all `.proto` bindings
- `make clean-pb`: Delete all `.pb.go` bindings
