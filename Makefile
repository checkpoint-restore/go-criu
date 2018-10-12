GO ?= go
CC ?= gcc
ifeq ($(GOPATH),)
export GOPATH := $(shell $(GO) env GOPATH)
endif
FIRST_GOPATH := $(firstword $(subst :, ,$(GOPATH)))
GOBIN := $(shell $(GO) env GOBIN)
ifeq ($(GOBIN),)
	GOBIN := $(FIRST_GOPATH)/bin
endif

all: build test phaul phaul-test

lint:
	@golint . test phaul
build:
	@$(GO) build -v

test/piggie: test/piggie.c
	@$(CC) $^ -o $@

test/test: test/main.go
	@$(GO) build -v -o test/test test/main.go

test: test/test test/piggie
	mkdir -p image
	test/piggie
	test/test dump `pidof piggie` image
	test/test restore image
	killall -9 piggie || :

phaul:
	@cd phaul; go build -v

test/phaul: test/phaul-main.go
	@$(GO) build -v -o test/phaul test/phaul-main.go

phaul-test: test/phaul test/piggie
	rm -rf image
	test/piggie
	test/phaul `pidof piggie`
	killall -9 piggie || :

clean:
	@rm -f test/test test/piggie test/phaul
	@rm -rf image

install.tools:
	if [ ! -x "$(GOBIN)/golint" ]; then \
		$(GO) get -u golang.org/x/lint/golint; \
	fi

.PHONY: build test clean lint phaul