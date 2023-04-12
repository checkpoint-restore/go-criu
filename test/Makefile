SHELL = /bin/bash
GO ?= go
CC ?= gcc
COVERAGE_PATH ?= $(shell pwd)/.coverage
CRIU_FEATURE_MEM_TRACK = $(shell if criu check --feature mem_dirty_track > /dev/null; then echo 1; else echo 0; fi)
CRIU_FEATURE_LAZY_PAGES = $(shell if criu check --feature uffd-noncoop > /dev/null; then echo 1; else echo 0; fi)
CRIU_FEATURE_PIDFD_STORE = $(shell if criu check --feature pidfd_store > /dev/null; then echo 1; else echo 0; fi)

export CRIU_FEATURE_MEM_TRACK CRIU_FEATURE_LAZY_PAGES CRIU_FEATURE_PIDFD_STORE
export GOCOVERDIR=$(COVERAGE_PATH)

TEST_PAYLOAD := piggie/piggie
TEST_BINARIES := test $(TEST_PAYLOAD) phaul/phaul
COVERAGE_BINARIES := test.coverage phaul/phaul.coverage

all: $(TEST_BINARIES) phaul-test
	mkdir -p image
	PID=$$(piggie/piggie) && { \
	test dump $$PID image && \
	test restore image; \
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

test.coverage: *.go
	$(GO) test \
		-covermode=count \
		-coverpkg=./... \
		-mod=vendor \
		-tags coverage \
		-buildmode=pie -c -o $@ $^

phaul/phaul.coverage: phaul/*.go
	$(GO) test \
		-covermode=count \
		-coverpkg=./... \
		-mod=vendor \
		-tags coverage \
		-buildmode=pie -c -o $@ $^

coverage: $(COVERAGE_BINARIES) $(TEST_PAYLOAD)
	mkdir -p $(COVERAGE_PATH)
	mkdir -p image
	PID=$$(piggie/piggie) && { \
	./test.coverage -test.v -test.coverprofile=${COVERAGE_PATH}/coverprofile.integration.$$RANDOM COVERAGE dump $$PID image || echo ">> Dump exit code $$?"; \
	./test.coverage -test.v -test.coverprofile=${COVERAGE_PATH}/coverprofile.integration.$$RANDOM COVERAGE restore image; \
	pkill -9 piggie; \
	}
	rm -rf image
	PID=$$(piggie/piggie) && { \
	phaul/phaul.coverage -test.v -test.coverprofile=${COVERAGE_PATH}/coverprofile.integration.$$RANDOM COVERAGE $$PID; \
	pkill -9 piggie; \
	}
	echo "mode: set" > ${COVERAGE_PATH}/coverage.out && cat ${COVERAGE_PATH}/coverprofile* | \
		grep -v mode: | sort -r | awk '{if($$1 != last) {print $$0;last=$$1}}' >> ${COVERAGE_PATH}/coverage.out

codecov:
	curl -Os https://uploader.codecov.io/latest/linux/codecov
	chmod +x codecov
	./codecov -f ${COVERAGE_PATH}/coverage.out

clean:
	@rm -f $(TEST_BINARIES) $(COVERAGE_BINARIES) codecov
	@rm -rf image $(COVERAGE_PATH)

.PHONY: all clean coverage codecov crit-test phaul-test
