GO ?= go
CC ?= gcc
CRIU ?= criu
BATS ?= bats

all: integration-test e2e-test clean

integration-test: test-imgs crit-test
	@echo "Running integration test"
	@./crit-test

e2e-test: ../../crit/bin/crit test-imgs crit-test.sh bats-test
	@echo "Running E2E test"
	@./crit-test.sh

bats-test: ../../crit/bin/crit test-imgs
	@echo "Running bats tests"
	@$(BATS) humanize.bats

CRIU_TREE ?= ../../criu
CRIU_BIN ?= $(CRIU_TREE)/criu/criu

compression-test: ../../crit/bin/crit
	$(abspath $(CRIU_BIN)) check --no-default-config --feature compress
	CRIT_BIN=$(abspath ../../crit/bin/crit) CRIT_COMPRESSION_ACTION=decompress \
	python3 $(abspath $(CRIU_TREE)/test/zdtm.py) run \
		-f h \
		--ignore-taint \
		--keep-img failed \
		--criu-bin $(abspath $(CRIU_BIN)) \
		--script $(abspath compression-action.sh) \
		-t zdtm/static/compress_pages00 \
		-t zdtm/static/compress_pages_region00
	# Run directly because zdtm.py checks dump statistics after the action script.
	CRIU_TREE=$(abspath $(CRIU_TREE)) \
		CRIU_BIN=$(abspath $(CRIU_BIN)) \
		CRIT_BIN=$(abspath ../../crit/bin/crit) \
		sh $(abspath compression-from-raw-test.sh)
	CRIT_BIN=$(abspath ../../crit/bin/crit) CRIT_COMPRESSION_ACTION=roundtrip \
	python3 $(abspath $(CRIU_TREE)/test/zdtm.py) run \
		-f h \
		--ignore-taint \
		--keep-img failed \
		--criu-bin $(abspath $(CRIU_BIN)) \
		--script $(abspath compression-action.sh) \
		--pre 2 \
		-t zdtm/static/compress_pages_region03

# gen_mmapper_imgs generates CRIU images for mmapper process.
# Parameters:
#   $(1) - Memory size for memory mapping.
#   $(2) - Mmap flags: p - private, s - shared, a - anonymous.
#   $(3) - Directory where the generated images and information will be saved.
define gen_mmapper_imgs
	$(eval MM_PID := $(shell ../mmapper/mmapper $(1) $(2) "/tmp/mmapper-${1}-${2}.log"))
	$(eval MM_DIR := $(3))
	mkdir -p $(MM_DIR)
	cp /proc/$(MM_PID)/environ $(MM_DIR)
	cp /proc/$(MM_PID)/cmdline $(MM_DIR)
	$(CRIU) dump -v4 -o dump.log -D $(MM_DIR) -t $(MM_PID)
endef

test-imgs: ../loop/loop ../mmapper/mmapper inetsk_sleeper/inetsk_sleeper
	$(eval LOOP_PID := $(shell ../loop/loop))
	$(eval LOOP_DIR := $@/loop)
	mkdir -p ${LOOP_DIR}
	cp /proc/${LOOP_PID}/environ ${LOOP_DIR}
	cp /proc/${LOOP_PID}/cmdline ${LOOP_DIR}
	$(CRIU) dump -v4 -o dump.log -D ${LOOP_DIR} -t $(LOOP_PID)

	$(eval INETSK_PID := $(shell ./inetsk_sleeper/inetsk_sleeper))
	$(eval INETSK_DIR := $@/inetsk)
	mkdir -p ${INETSK_DIR}
	$(CRIU) dump -v4 -o dump.log -D ${INETSK_DIR} -t $(INETSK_PID)
	kill -9 $(INETSK_PID) 2>/dev/null || true

	# Generate mmapper CRIU images for different memory mapping scenarios.
	# gen_mmapper_imgs <memory size> <mmap flags> <images dir>
	$(call gen_mmapper_imgs, 257, p, $@/mm_p_257b)
	$(call gen_mmapper_imgs, 257, pa, $@/mm_pa_257b)
	$(call gen_mmapper_imgs, 257, s, $@/mm_s_257b)
	$(call gen_mmapper_imgs, 257, sa, $@/mm_sa_257b)
	$(call gen_mmapper_imgs, 4096, p, $@/mm_p_4kb)
	$(call gen_mmapper_imgs, 4096, pa, $@/mm_pa_4kb)
	$(call gen_mmapper_imgs, 4096, s, $@/mm_s_4kb)
	$(call gen_mmapper_imgs, 4096, sa, $@/mm_sa_4kb)
	$(call gen_mmapper_imgs, 32768, p, $@/mm_p_32kb)
	$(call gen_mmapper_imgs, 32768, pa, $@/mm_pa_32kb)
	$(call gen_mmapper_imgs, 32768, s, $@/mm_s_32kb)
	$(call gen_mmapper_imgs, 32768, sa, $@/mm_sa_32kb)

../../crit/bin/crit:
	$(MAKE) -C ../../crit bin/crit GOFLAGS="${GOFLAGS}"

../loop/loop:
	$(MAKE) -C ../loop

../mmapper/mmapper:
	$(MAKE) -C ../mmapper

inetsk_sleeper/inetsk_sleeper:
	$(MAKE) -C inetsk_sleeper

crit-test: main.go
	$(GO) build -v -o $@ $^

clean:
	@rm -rf test-imgs crit-test /tmp/mmapper-*.log
	$(MAKE) -C inetsk_sleeper clean

.PHONY: all test integration-test e2e-test bats-test compression-test clean
