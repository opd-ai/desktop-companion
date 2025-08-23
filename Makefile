# Cross-Platform Build Makefile for Desktop Companion
# Uses Go's built-in cross-compilation - no external tools needed

.PHONY: all build-windows build-macos build-linux build-all clean test deps

# Build configuration
BINARY_NAME=companion
BUILD_DIR=build
CMD_DIR=cmd/companion
LDFLAGS=-ldflags="-s -w"

# Default target
all: build-all

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run tests
test:
	go test ./... -v -cover

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	go clean

# Create build directory
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

# Windows build (64-bit)
build-windows: $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows.exe $(CMD_DIR)/main.go

# macOS build (64-bit Intel)
build-macos: $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-macos $(CMD_DIR)/main.go

# macOS build (ARM64 Apple Silicon)
build-macos-arm: $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-macos-arm64 $(CMD_DIR)/main.go

# Linux build (64-bit)
build-linux: $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux $(CMD_DIR)/main.go

# Build for all platforms
build-all: build-windows build-macos build-macos-arm build-linux

# Development build (current platform)
build:
	go build $(LDFLAGS) -o $(BINARY_NAME) $(CMD_DIR)/main.go

# Run locally
run:
	go run $(CMD_DIR)/main.go

# Run with debug output
run-debug:
	go run $(CMD_DIR)/main.go -debug

# Run with custom character
run-custom:
	go run $(CMD_DIR)/main.go -character assets/characters/default/character.json -debug

# Performance profiling builds
build-profile: $(BUILD_DIR)
	go build -ldflags="-s -w" -tags profile -o $(BUILD_DIR)/$(BINARY_NAME)-profile $(CMD_DIR)/main.go

# Static analysis
lint:
	go vet ./...
	go fmt ./...

# Security check
security:
	go list -json -m all | nancy sleuth

# Generate coverage report
coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# Benchmark tests
bench:
	go test ./... -bench=. -benchmem

# Install development tools
install-tools:
	go install golang.org/x/tools/cmd/goimports@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest

# Package releases (with assets)
package-all: build-all
	cd $(BUILD_DIR) && \
	cp -r ../assets . && \
	tar -czf $(BINARY_NAME)-windows-amd64.tar.gz $(BINARY_NAME)-windows.exe assets/ && \
	tar -czf $(BINARY_NAME)-macos-amd64.tar.gz $(BINARY_NAME)-macos assets/ && \
	tar -czf $(BINARY_NAME)-macos-arm64.tar.gz $(BINARY_NAME)-macos-arm64 assets/ && \
	tar -czf $(BINARY_NAME)-linux-amd64.tar.gz $(BINARY_NAME)-linux assets/

# Help target
help:
	@echo "Available targets:"
	@echo "  build-all       - Build for all platforms"
	@echo "  build-windows   - Build for Windows"
	@echo "  build-macos     - Build for macOS (Intel)"
	@echo "  build-macos-arm - Build for macOS (Apple Silicon)"
	@echo "  build-linux     - Build for Linux"
	@echo "  build           - Build for current platform"
	@echo "  run             - Run application locally"
	@echo "  run-debug       - Run with debug output"
	@echo "  test            - Run unit tests"
	@echo "  coverage        - Generate test coverage report"
	@echo "  clean           - Remove build artifacts"
	@echo "  deps            - Install/update dependencies"
	@echo "  package-all     - Create release packages"
