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

# Android APK build (requires Android SDK and NDK)
android-setup:
	@echo "Setting up Android build environment..."
	@echo "Please ensure you have:"
	@echo "  - Android SDK installed and ANDROID_HOME set"
	@echo "  - Android NDK installed"
	@echo "  - Java 8+ installed"
	@echo "  - fyne CLI tool installed (already done)"

# Build Android APK
android-apk: $(BUILD_DIR)
	@echo "Building Android APK..."
	cd $(BUILD_DIR) && fyne package --target android --app-id ai.opd.dds --name "Desktop Companion" \
		--app-version "1.0.0" --app-build 1 --icon ../assets/characters/default/animations/idle.gif \
		--src ../$(CMD_DIR) --release

# Build Android APK (debug version)
android-debug: $(BUILD_DIR)
	@echo "Building Android APK (debug)..."
	cd $(BUILD_DIR) && fyne package --target android --app-id ai.opd.dds.debug --name "DDS Debug" \
		--app-version "1.0.0-debug" --app-build 1 --icon ../assets/characters/default/animations/idle.gif \
		--src ../$(CMD_DIR)

# Install APK to connected Android device
android-install: android-apk
	@echo "Installing APK to Android device..."
	adb install -r $(BUILD_DIR)/companion.apk

# Install debug APK to connected Android device
android-install-debug: android-debug
	@echo "Installing debug APK to Android device..."
	adb install -r $(BUILD_DIR)/companion-debug.apk

# Cross-platform CI/CD preparation
ci-prepare:
	@echo "Preparing for CI/CD..."
	go mod download
	go install fyne.io/tools/cmd/fyne@latest

# Cross-platform release builds (when running on appropriate platforms)
release-windows: $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(CMD_DIR)/main.go

release-macos: $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-macos-amd64 $(CMD_DIR)/main.go

release-linux: $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_DIR)/main.go

# Create all release packages (must run on respective platforms for Fyne compatibility)
package-all: package-native
	@echo "Packaging complete. Note: Android APK must be built separately using 'make android-apk'"

# Help target
help:
	@echo "Available targets:"
	@echo "  build              - Build for current platform"
	@echo "  run                - Run application locally"
	@echo "  run-debug          - Run with debug output"
	@echo "  test               - Run unit tests"
	@echo "  coverage           - Generate test coverage report"
	@echo "  clean              - Remove build artifacts"
	@echo "  deps               - Install/update dependencies"
	@echo "  package-native     - Create release package for current platform"
	@echo ""
	@echo "Android builds:"
	@echo "  android-setup      - Show Android build requirements"
	@echo "  android-apk        - Build Android APK (release)"
	@echo "  android-debug      - Build Android APK (debug)"
	@echo "  android-install    - Install APK to connected device"
	@echo "  android-install-debug - Install debug APK to device"
	@echo ""
	@echo "Cross-platform:"
	@echo "  ci-prepare         - Prepare environment for CI/CD"
	@echo "  release-windows    - Build Windows binary (requires Windows or cross-compilation setup)"
	@echo "  release-macos      - Build macOS binary (requires macOS or cross-compilation setup)"
	@echo "  release-linux      - Build Linux binary"
	@echo ""
	@echo "Note: Fyne mobile builds require platform-specific setup and Android SDK"
