# xBasic Makefile

BINARY_NAME=xbasic
VERSION=0.1.0
BUILD_DIR=build

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags "-s -w"

.PHONY: all build build-linux build-macos build-windows test clean deps run

all: deps build

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Build for current platform
build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) ./cmd/xbasic

# Build for Linux (amd64)
build-linux:
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/xbasic

# Build for Linux (arm64)
build-linux-arm64:
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/xbasic

# Build for macOS (amd64)
build-macos:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-macos-amd64 ./cmd/xbasic

# Build for macOS (arm64 - Apple Silicon)
build-macos-arm64:
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-macos-arm64 ./cmd/xbasic

# Build all platforms
build-all: build-linux build-linux-arm64 build-macos build-macos-arm64

# Run tests
test:
	$(GOTEST) -v ./...

# Run the application
run: build
	./$(BINARY_NAME)

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -rf $(BUILD_DIR)

# Install locally
install: build
	cp $(BINARY_NAME) /usr/local/bin/

# Create build directory
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

build-linux build-linux-arm64 build-macos build-macos-arm64: $(BUILD_DIR)

# Format code
fmt:
	$(GOCMD) fmt ./...

# Lint code
lint:
	golangci-lint run

# Generate version info
version:
	@echo $(VERSION)
