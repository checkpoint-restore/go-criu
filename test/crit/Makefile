GO ?= go
CRIU ?= criu

all: integration-test e2e-test clean

integration-test: test-imgs crit-test
	@echo "Running integration test"
	@./crit-test

e2e-test: ../../crit/bin/crit test-imgs crit-test.sh
	@echo "Running E2E test"
	@./crit-test.sh

test-imgs: ../loop/loop
	$(eval PID := $(shell ../loop/loop))
	mkdir -p $@
	$(CRIU) dump -v4 -o dump.log -D $@ -t $(PID)
	$(CRIU) restore -v4 -o restore.log -D $@ -d
	pkill -9 loop

../../crit/bin/crit:
	$(MAKE) -C ../../crit bin/crit GOFLAGS="${GOFLAGS}"

../loop/loop:
	$(MAKE) -C ../loop

crit-test: main.go
	$(GO) build -v -o $@ $^

clean:
	@rm -rf test-imgs

.PHONY: all test integration-test e2e-test clean
