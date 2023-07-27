GO ?= go
CRIU ?= criu
CRIT_SRC := $(shell find . -type f -name '*.go')
CRIT_CLI := cmd/main.go

bin/crit: $(CRIT_SRC)
	$(GO) build ${GOFLAGS} -o $@ $(CRIT_CLI)

../test/loop/loop:
	$(MAKE) -C ../test/loop

test-imgs: ../test/loop/loop
	$(eval PID := $(shell ../test/loop/loop))
	mkdir -p $@
	$(CRIU) dump -v4 -o dump.log -D $@ -t $(PID)
	$(CRIU) restore -v4 -o restore.log -D $@ -d
	pkill -9 loop

unit-test: test-imgs
	$(eval GOFLAGS ?= -cover)
	$(GO) test ${GOFLAGS} -v ./...

clean:
	@rm -f bin/crit
	@rm -rf test-imgs

.PHONY: unit-test clean
