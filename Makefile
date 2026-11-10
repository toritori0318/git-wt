.PHONY: build install clean test help

# Binary name
BINARY_NAME=wt

# Build directory
BUILD_DIR=.

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Stamped into `wt version` / `wt --version`. Override with make VERSION=v1.2.3.
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse --short=7 HEAD 2>/dev/null || echo unknown)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

# Build the project
build:
	$(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/wt

# Install to $GOPATH/bin
install:
	$(GOCMD) install -ldflags "$(LDFLAGS)" ./cmd/wt

# Clean build files
clean:
	$(GOCLEAN)
	rm -f $(BUILD_DIR)/$(BINARY_NAME)

# Run tests
test:
	$(GOTEST) -v ./...

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Build for multiple platforms
build-all: build-linux build-darwin build-windows

build-linux:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/wt

build-darwin:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/wt
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/wt

build-windows:
	GOOS=windows GOARCH=amd64 $(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/wt

# Show help
help:
	@echo "Available targets:"
	@echo "  build       - Build the binary"
	@echo "  install     - Install to \$$GOPATH/bin"
	@echo "  clean       - Remove build files"
	@echo "  test        - Run tests"
	@echo "  deps        - Download and tidy dependencies"
	@echo "  build-all   - Build for all platforms"
	@echo "  help        - Show this help message"
	@echo ""
	@echo "Version stamped into the binary: $(VERSION)"
