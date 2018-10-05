all: build

build:
	@go build -v

test/test: test/main.go
	@go build -v -o test/test test/main.go

test: test/test

clean:
	@rm -f test/test

.PHONY: build test clean
