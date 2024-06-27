PY ?= python3

MAGIC_DEST ?= ../../magic/magic.go

magic-gen: clean magic.h magicgen.py
	@$(PY) magicgen.py magic.h $(MAGIC_DEST)

magic.h:
	curl -s https://raw.githubusercontent.com/checkpoint-restore/criu/criu-dev/criu/include/magic.h -o magic.h

test: test_magicgen.py magicgen-test.sh
	@echo "Running unit tests..."
	@$(PY) test_magicgen.py
	@echo "Running E2E tests..."
	@./magicgen-test.sh
	@rm -f input.h output.go expected.go

clean:
	rm -f $(MAGIC_DEST) magic.h

.PHONY: magic-gen test clean
