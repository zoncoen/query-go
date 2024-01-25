export

.DEFAULT_GOAL := generate

UNAME_OS := $(shell uname -s)
UNAME_ARCH := $(shell uname -m)

BIN_DIR := $(CURDIR)/bin
PROTO_DIR := $(CURDIR)/testdata/proto
GEN_PB_DIR := $(CURDIR)/testdata/gen
GOBIN := $(BIN_DIR)
PATH := $(GOBIN):$(PATH)

$(BIN_DIR):
	@mkdir -p $(BIN_DIR)

PROTOC := $(BIN_DIR)/protoc
PROTOC_VERSION := 25.1
PROTOC_OS := $(UNAME_OS)
ifeq "$(UNAME_OS)" "Darwin"
	PROTOC_OS = osx
endif
PROTOC_ARCH := $(UNAME_ARCH)
ifeq "$(UNAME_ARCH)" "arm64"
	PROTOC_ARCH = aarch_64
endif
PROTOC_ZIP := protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH).zip
$(PROTOC): | $(BIN_DIR)
	@curl -sSOL \
		"https://github.com/protocolbuffers/protobuf/releases/download/v$(PROTOC_VERSION)/$(PROTOC_ZIP)"
	@unzip -j -o $(PROTOC_ZIP) -d $(BIN_DIR) bin/protoc
	@unzip -o $(PROTOC_ZIP) -d $(BIN_DIR) "include/*"
	@rm -f $(PROTOC_ZIP)

PROTOC_GEN_GO := $(BIN_DIR)/protoc-gen-go
$(PROTOC_GEN_GO): | $(BIN_DIR)
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.32.0

PROTOC_GEN_GO_GRPC := $(BIN_DIR)/protoc-gen-go-grpc
$(PROTOC_GEN_GO_GRPC): | $(BIN_DIR)
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0

PROTOC_OPTION := -I$(PROTO_DIR)
PROTOC_GO_OPTION := --plugin=${PROTOC_GEN_GO} --go_out=$(GEN_PB_DIR) --go_opt=paths=source_relative
PROTOC_GO_GRPC_OPTION := --go-grpc_out=require_unimplemented_servers=false:$(GEN_PB_DIR) --go-grpc_opt=paths=source_relative
.PHONY: generate
generate: $(PROTOC) $(PROTOC_GEN_GO) $(PROTOC_GEN_GO_GRPC)
	@mkdir -p $(GEN_PB_DIR)
	@find $(PROTO_DIR) -name '*.proto' | xargs -P8 protoc $(PROTOC_OPTION) $(PROTOC_GO_OPTION) $(PROTOC_GO_GRPC_OPTION)
