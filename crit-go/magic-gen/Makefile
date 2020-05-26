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

all: build test magicgen

lint: 
	@golint magic-gen.go

build:
	 @$(GO) build -v

test/magicgen: build
	echo "#define TEST_MAGIC 0x12345678" > test_input
	@rm -f test_output 
	./magic-gen test_input test_output
	cmp samplemagic test_output
	if [ $$? -eq 0 ]; then \
		echo PASS; \
		exit 0; \
	else \
		echo FAIL; \
		exit 1; \
	fi
	@rm -f test_input
	@rm -f test_output

test: test/magicgen

magic.h:
	curl -s https://raw.githubusercontent.com/checkpoint-restore/criu/criu-dev/criu/include/magic.h -o magic.h

magicgen: build magic.h
	./magic-gen magic.h magic.json
	@rm -f magic.h

install.tools:
	if [ ! -x "$(GOBIN)/golint" ]; then \
		$(GO) get -u golang.org/x/lint/golint; \
	fi

clean:
	@rm -f testmagic.json
	@rm -f test_output

.PHONY: build test clean lint magicgen
