CC ?= gcc

all: test clean

test: gen-imgs
	@echo "Running E2E test"
	@./crit-test.sh

gen-imgs: loop
	$(eval PID := $(shell ./loop))
	criu dump -v4 -o dump.log -D ./ -t $(PID)

loop: loop.c
	$(CC) $^ -o $@

clean:
	@rm -f *.img *.log *.json stats-*

.PHONY: all test gen-imgs clean
