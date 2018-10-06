all: build test phaul phaul-test

lint:
	@golint . test phaul
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

phaul:
	@cd phaul; go build -v

test/phaul: test/phaul-main.go
	@go build -v -o test/phaul test/phaul-main.go

phaul-test: test/phaul test/piggie
	rm -rf image
	test/piggie
	test/phaul `pidof piggie`
	killall -9 piggie || :

clean:
	@rm -f test/test test/piggie test/phaul
	@rm -rf image

.PHONY: build test clean lint phaul
