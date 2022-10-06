SHELL = /bin/bash
GO ?= go
CC ?= gcc

all: build

lint:
	golangci-lint run ./...

build: rpc/rpc.pb.go
	$(GO) build -v ./...
	# Build crit binary
	$(MAKE) -C crit bin/crit

test: build
	$(MAKE) -C test

coverage:
	$(MAKE) -C test coverage

codecov:
	$(MAKE) -C test codecov

rpc/rpc.proto:
	curl -sSL https://raw.githubusercontent.com/checkpoint-restore/criu/master/images/rpc.proto -o $@

rpc/rpc.pb.go: rpc/rpc.proto
	protoc --go_out=. --go_opt=M$^=rpc/ $^

vendor:
	GO111MODULE=on $(GO) mod tidy
	GO111MODULE=on $(GO) mod vendor
	GO111MODULE=on $(GO) mod verify

.PHONY: build test lint vendor coverage codecov
