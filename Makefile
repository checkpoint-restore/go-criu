all: build

lint:
	@golint . test
build:
	@go build -v

test/piggie: test/piggie.c
	@gcc $^ -o $@

test/test: test/main.go
	@go build -v -o test/test test/main.go

test: test/test test/piggie
	mkdir -p image
	test/piggie
	test/test dump `pidof piggie` image
	test/test restore image
	killall -9 piggie || :

clean:
	@rm -f test/test test/piggie
	@rm -rf image

.PHONY: build test clean lint
