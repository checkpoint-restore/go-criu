GO ?= go
CRIU ?= criu

bin/crit: cmd/main.go
	$(GO) build ${GOFLAGS} -o $@ $^

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
