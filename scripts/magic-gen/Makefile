PY ?= python3

MAGIC_DEST ?= ../../magic/magic.go

magic-gen: clean magic.h
	@$(PY) magicgen.py magic.h $(MAGIC_DEST)

magic.h:
	curl -s https://raw.githubusercontent.com/checkpoint-restore/criu/criu-dev/criu/include/magic.h -o magic.h

test:
	@./magicgen-test.sh
	@rm -f input.h output.go expected.go

clean:
	rm -f $(MAGIC_DEST) magic.h

.PHONY: magic-gen test clean
