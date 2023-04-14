SHELL = /bin/bash
GO ?= go
CC ?= gcc
COVERAGE_PATH ?= $(shell pwd)/.coverage
CRIU_FEATURE_MEM_TRACK = $(shell if criu check --feature mem_dirty_track > /dev/null; then echo 1; else echo 0; fi)
CRIU_FEATURE_LAZY_PAGES = $(shell if criu check --feature uffd-noncoop > /dev/null; then echo 1; else echo 0; fi)
CRIU_FEATURE_PIDFD_STORE = $(shell if criu check --feature pidfd_store > /dev/null; then echo 1; else echo 0; fi)

export CRIU_FEATURE_MEM_TRACK CRIU_FEATURE_LAZY_PAGES CRIU_FEATURE_PIDFD_STORE

TEST_PAYLOAD := piggie/piggie
TEST_BINARIES := test $(TEST_PAYLOAD) phaul/phaul
COVERAGE_BINARIES := test.coverage phaul/phaul.coverage

GO_MAJOR_VER = $(shell $(GO) version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f1)
GO_MINOR_VER = $(shell $(GO) version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f2)
MIN_GO_MAJOR_VER = 1
MIN_GO_MINOR_VER = 20
GO_VALIDATION_ERR = Go version is not supported. Please update to at least $(MIN_GO_MAJOR_VER).$(MIN_GO_MINOR_VER)

check-go-version:
	@if [ $(GO_MAJOR_VER) -gt $(MIN_GO_MAJOR_VER) ]; then \
		exit 0 ;\
	elif [ $(GO_MAJOR_VER) -lt $(MIN_GO_MAJOR_VER) ]; then \
		echo '$(GO_VALIDATION_ERR)';\
		exit 1; \
	elif [ $(GO_MINOR_VER) -lt $(MIN_GO_MINOR_VER) ] ; then \
		echo '$(GO_VALIDATION_ERR)';\
		exit 1; \
	fi


all: $(TEST_BINARIES) phaul-test
	mkdir -p image
	PID=$$(piggie/piggie) && { \
	./test dump $$PID image && \
	./test restore image; \
	pkill -9 piggie; \
	}
	rm -rf image

piggie/piggie: piggie/piggie.c
	$(CC) $^ -o $@

test: main.go
	$(GO) build -v -o $@ $^

phaul/phaul: phaul/main.go
	$(GO) build -v -o $@ $^

phaul-test: $(TEST_BINARIES)
	rm -rf image
	PID=$$(piggie/piggie) && { \
	phaul/phaul $$PID; \
	pkill -9 piggie; \
	}

crit-test:
	$(MAKE) -C crit

test.coverage: check-go-version *.go
	$(GO) build \
		-cover \
		-o $@ main.go

phaul/phaul.coverage: check-go-version phaul/*.go
	$(GO) build \
		-cover \
		-o $@ phaul/main.go

coverage: check-go-version $(COVERAGE_BINARIES) $(TEST_PAYLOAD)
	mkdir -p $(COVERAGE_PATH)
	mkdir -p image
	PID=$$(piggie/piggie) && { \
	GOCOVERDIR=${COVERAGE_PATH} ./test.coverage dump $$PID image && \
	GOCOVERDIR=${COVERAGE_PATH} ./test.coverage restore image; \
	pkill -9 piggie; \
	}
	rm -rf image
	PID=$$(piggie/piggie) && { \
	GOCOVERDIR=${COVERAGE_PATH} phaul/phaul.coverage $$PID; \
	pkill -9 piggie; \
	}
	# Print coverage from this run
	$(GO) tool covdata percent -i=${COVERAGE_PATH}
	$(GO) tool covdata textfmt -i=${COVERAGE_PATH} -o ${COVERAGE_PATH}/coverage.out

codecov:
	curl -Os https://uploader.codecov.io/latest/linux/codecov
	chmod +x codecov
	./codecov -f "$(COVERAGE_PATH)/coverage.out"

clean:
	@rm -f $(TEST_BINARIES) $(COVERAGE_BINARIES) codecov
	@rm -rf image $(COVERAGE_PATH)

.PHONY: all clean coverage codecov crit-test phaul-test check-go-version
