.PHONY: build install clean test help

# Binary name
BINARY_NAME=gwt

# Build directory
BUILD_DIR=.

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build the project
build:
	$(GOBUILD) -buildvcs=false -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/gwt

# Install to $GOPATH/bin
install:
	$(GOCMD) install ./cmd/gwt

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
	GOOS=linux GOARCH=amd64 $(GOBUILD) -buildvcs=false -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/gwt

build-darwin:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -buildvcs=false -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/gwt
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -buildvcs=false -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/gwt

build-windows:
	GOOS=windows GOARCH=amd64 $(GOBUILD) -buildvcs=false -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/gwt

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
