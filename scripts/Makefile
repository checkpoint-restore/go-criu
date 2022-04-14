GO ?= go

magic-gen: clean magic.h
	$(GO) run magicgen.go magic.h ../magic/magic.go

magic.h:
	curl -s https://raw.githubusercontent.com/checkpoint-restore/criu/criu-dev/criu/include/magic.h -o magic.h

test:
	@./magicgen-test.sh
	@rm -f input.h output.go expected.go

clean:
	rm -f ../magic/magic.go

.PHONY: magic-gen test clean
