GO ?= go

# The import path that protoc will use if a proto file imports another one
import_path := github.com/checkpoint-restore/go-criu/crit/images
# Path to .proto source files
proto_path := ./images
# Generate string of all .proto filenames
proto_files := $(sort $(subst $(proto_path)/,,$(wildcard $(proto_path)/*.proto)))
# Generate M flag to specify import path for all .proto files
# and replace all spaces with commas to use with go_opt flag
comma := ,
proto_opts := $(subst $() $(),$(comma),$(patsubst %,M%=$(import_path),$(proto_files)))

all: gen-proto bin/crit

update-proto:
	rm ./images/*.proto || true
	git clone --depth 1 --branch master https://github.com/checkpoint-restore/criu criu-temp
	cp criu-temp/images/*.proto ./images/
	# rpc.proto is not an image and it is used only to communicate criu-service and swrk.
	rm -rf criu-temp images/rpc.proto
	# To prevent namespace conflict with proto files
	# in github.com/letsencrypt/boulder, we prepend
	# a prefix to the filenames.
	mv ./images/sa.proto ./images/criu-sa.proto
	sed -i 's/sa\.proto/criu-sa\.proto/g' images/*.proto
	mv ./images/core.proto ./images/criu-core.proto
	sed -i 's/core\.proto/criu-core\.proto/g' images/*.proto

gen-proto:
	rm -f ./images/*.pb.go
	@protoc \
		--proto_path=$(proto_path) \
		--go_out=$(proto_path) \
		--go_opt=paths=source_relative,$(proto_opts) \
		$(proto_files)

bin/crit: cmd/cli.go
	$(GO) build -o $@ $^

.PHONY: all gen-proto update-proto
