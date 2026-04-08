# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

# Binary names
COMPRESSOR_BINARY=ppma_compress
DECOMPRESSOR_BINARY=ppma_decompress

# Paths to main packages
COMPRESSOR_MAIN=./cmd/compressor
DECOMPRESSOR_MAIN=./cmd/decompressor

# Output directory
BIN_DIR=./bin

all: build

build: clean $(BIN_DIR) $(BIN_DIR)/$(COMPRESSOR_BINARY) $(BIN_DIR)/$(DECOMPRESSOR_BINARY)

$(BIN_DIR):
	mkdir -p $(BIN_DIR)

$(BIN_DIR)/$(COMPRESSOR_BINARY): $(COMPRESSOR_MAIN)
	$(GOBUILD) -o $(BIN_DIR)/$(COMPRESSOR_BINARY) $(COMPRESSOR_MAIN)

$(BIN_DIR)/$(DECOMPRESSOR_BINARY): $(DECOMPRESSOR_MAIN)
	$(GOBUILD) -o $(BIN_DIR)/$(DECOMPRESSOR_BINARY) $(DECOMPRESSOR_MAIN)

clean:
	rm -rf $(BIN_DIR)

test:
	$(GOTEST) ./...

test-e2e: build
	go run ./cmd/e2e

.PHONY: all build clean test