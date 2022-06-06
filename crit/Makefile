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
	rm -R criu-temp

gen-proto:
	@protoc \
		--proto_path=$(proto_path) \
		--go_out=$(proto_path) \
		--go_opt=paths=source_relative,$(proto_opts) \
		$(proto_files)

bin/crit: cmd/cli.go
	$(GO) build -o $@ $^

.PHONY: all gen-proto update-proto
