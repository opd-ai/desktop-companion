# Build Makefile for Desktop Companion
# Note: Due to Fyne GUI framework limitations, only native builds are supported

.PHONY: all build clean test deps

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

# Native build for current platform
build: $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)/main.go

# Note: Cross-platform builds not supported due to Fyne GUI framework limitations
# Fyne requires platform-specific CGO libraries for OpenGL/graphics drivers
# Build on the target platform for proper binary distribution

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

# Package releases (with assets) - requires manual build on target platforms
package-native:
	cd $(BUILD_DIR) && \
	cp -r ../assets . && \
	tar -czf $(BINARY_NAME)-$(shell go env GOOS)-$(shell go env GOARCH).tar.gz $(BINARY_NAME) assets/

# Help target
help:
	@echo "Available targets:"
	@echo "  build           - Build for current platform"
	@echo "  run             - Run application locally"
	@echo "  run-debug       - Run with debug output"
	@echo "  test            - Run unit tests"
	@echo "  coverage        - Generate test coverage report"
	@echo "  clean           - Remove build artifacts"
	@echo "  deps            - Install/update dependencies"
	@echo "  package-native  - Create release package for current platform"
	@echo ""
	@echo "Note: Cross-platform builds require building on target platform"
