PY ?= python3
PROTO_PATH ?= ../../crit/images
PB_PATH ?= ../../crit/images
GIT_BRANCH ?= master

pb-gen: clean-pb
	@$(PY) protogen.py $(PROTO_PATH) $(PB_PATH)

proto-update: clean-proto
	git clone --depth 1 --branch $(GIT_BRANCH) https://github.com/checkpoint-restore/criu criu-temp
	cp criu-temp/images/*.proto $(PROTO_PATH)/
	# rpc.proto is not an image and it is used only to communicate criu-service and swrk.
	rm -rf criu-temp $(PROTO_PATH)/rpc.proto
	# To prevent namespace conflict with proto files
	# in github.com/letsencrypt/boulder, we prepend
	# a prefix to the filenames.
	mv $(PROTO_PATH)/sa.proto $(PROTO_PATH)/criu-sa.proto
	sed -i 's/sa\.proto/criu-sa\.proto/g' $(PROTO_PATH)/*.proto
	mv $(PROTO_PATH)/core.proto $(PROTO_PATH)/criu-core.proto
	sed -i 's/core\.proto/criu-core\.proto/g' $(PROTO_PATH)/*.proto

clean-proto:
	@echo "Removing existing .proto files..."
	rm $(PROTO_PATH)/*.proto || true

clean-pb:
	@echo "Removing existing .pb.go files..."
	find $(PB_PATH) -type f -name '*.pb.go' -delete

.PHONY: pb-gen proto-update clean-proto clean-pb
