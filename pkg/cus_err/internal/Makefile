# Makefile for building proto files

# Variables
PROTO_DIR = protos
GEN_DIR = gen
PROTO_FILES = $(wildcard $(PROTO_DIR)/*.proto)
GO_OUT_DIR = $(GEN_DIR)

# Ensure the output directory exists
$(shell mkdir -p $(GO_OUT_DIR))

.PHONY:  proto_build proto_clean

# Build proto files
proto_build: $(PROTO_FILES)
	protoc --proto_path=$(PROTO_DIR) \
		--go_out=$(GO_OUT_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(GO_OUT_DIR) --go-grpc_opt=paths=source_relative \
		$(PROTO_FILES)

# Clean generated files
proto_clean:
	rm -rf $(GEN_DIR)/*



# Default target
.DEFAULT_GOAL := proto_build