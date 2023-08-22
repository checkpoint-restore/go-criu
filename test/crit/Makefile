GO ?= go
CRIU ?= criu

all: integration-test e2e-test clean

integration-test: test-imgs crit-test
	@echo "Running integration test"
	@./crit-test

e2e-test: ../../crit/bin/crit test-imgs crit-test.sh
	@echo "Running E2E test"
	@./crit-test.sh

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

test-imgs: ../loop/loop ../mmapper/mmapper
	$(eval LOOP_PID := $(shell ../loop/loop))
	$(eval LOOP_DIR := $@/loop)
	mkdir -p ${LOOP_DIR}
	cp /proc/${LOOP_PID}/environ ${LOOP_DIR}
	cp /proc/${LOOP_PID}/cmdline ${LOOP_DIR}
	$(CRIU) dump -v4 -o dump.log -D ${LOOP_DIR} -t $(LOOP_PID)

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

crit-test: main.go
	$(GO) build -v -o $@ $^

clean:
	@rm -rf test-imgs crit-test /tmp/mmapper-*.log

.PHONY: all test integration-test e2e-test clean
