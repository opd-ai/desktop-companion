#!/bin/bash
# Cross-Platform CI/CD Build Script for DDS
# This script handles building for multiple platforms including Android APK

set -e

# Configuration
VERSION="1.0.0"
BUILD_NUMBER="${GITHUB_RUN_NUMBER:-1}"
APP_ID="ai.opd.dds"
BUILD_DIR="build"
ARTIFACTS_DIR="artifacts"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if fyne tool is available
check_fyne() {
    if ! command -v fyne &> /dev/null; then
        log_warn "Fyne CLI tool not found. Installing..."
        go install fyne.io/tools/cmd/fyne@latest
        if ! command -v fyne &> /dev/null; then
            log_error "Failed to install fyne CLI tool"
            exit 1
        fi
    fi
    log_info "Fyne CLI version: $(fyne version)"
}

# Prepare build environment
prepare_env() {
    log_info "Preparing build environment..."
    
    # Create directories
    mkdir -p "$BUILD_DIR" "$ARTIFACTS_DIR"
    
    # Install dependencies
    go mod download
    go mod tidy
    
    # Install/update fyne tool
    check_fyne
}

# Run tests
run_tests() {
    log_info "Running tests..."
    go test ./... -v -cover
    
    # Generate coverage report
    go test ./... -coverprofile=coverage.out
    go tool cover -html=coverage.out -o coverage.html
    mv coverage.html "$ARTIFACTS_DIR/"
    
    log_info "Tests completed successfully"
}

# Build for desktop platforms
build_desktop() {
    local platform=$1
    local arch=$2
    local ext=$3
    
    log_info "Building for $platform/$arch..."
    
    local output="$BUILD_DIR/companion-$platform-$arch$ext"
    GOOS=$platform GOARCH=$arch go build -ldflags="-s -w" -o "$output" cmd/companion/main.go
    
    if [ -f "$output" ]; then
        log_info "Successfully built: $output"
        cp "$output" "$ARTIFACTS_DIR/"
    else
        log_error "Build failed for $platform/$arch"
        return 1
    fi
}

# Build Android APK
build_android() {
    log_info "Building Android APK..."
    
    # Check for Android SDK (optional - fyne can work without it in some cases)
    if [ -z "$ANDROID_HOME" ]; then
        log_warn "ANDROID_HOME not set. Android build may not work without Android SDK."
    fi
    
    # Create temporary directory for Android build
    local android_build_dir="$BUILD_DIR/android"
    mkdir -p "$android_build_dir"
    
    cd "$android_build_dir"
    
    # Build debug APK (doesn't require signing)
    log_info "Building debug APK..."
    if fyne package --target android --app-id "$APP_ID.debug" --name "DDS Debug" \
        --app-version "$VERSION-debug" --app-build "$BUILD_NUMBER" \
        --icon "../../assets/characters/default/animations/idle.gif" \
        --src "../../cmd/companion"; then
        
        # Find generated APK
        local apk_file=$(find . -name "*.apk" | head -1)
        if [ -n "$apk_file" ]; then
            log_info "Android APK built successfully: $apk_file"
            cp "$apk_file" "../../$ARTIFACTS_DIR/companion-debug.apk"
        else
            log_warn "APK file not found after build"
        fi
    else
        log_error "Android APK build failed"
        cd ../..
        return 1
    fi
    
    cd ../..
}

# Package with assets
package_with_assets() {
    log_info "Creating packages with assets..."
    
    cd "$BUILD_DIR"
    
    # Copy assets
    cp -r ../assets .
    
    # Create archives for each platform
    for binary in companion-*; do
        if [ -f "$binary" ]; then
            local platform_arch=$(echo "$binary" | sed 's/companion-//' | sed 's/\.exe$//')
            local archive_name="companion-$platform_arch.tar.gz"
            
            log_info "Creating package: $archive_name"
            tar -czf "$archive_name" "$binary" assets/
            mv "$archive_name" "../$ARTIFACTS_DIR/"
        fi
    done
    
    cd ..
}

# Main build function
main() {
    log_info "Starting DDS Cross-Platform Build"
    log_info "Version: $VERSION, Build: $BUILD_NUMBER"
    
    # Prepare environment
    prepare_env
    
    # Run tests
    run_tests
    
    # Build for current platform (always works)
    local current_os=$(go env GOOS)
    local current_arch=$(go env GOARCH)
    local current_ext=""
    
    if [ "$current_os" = "windows" ]; then
        current_ext=".exe"
    fi
    
    build_desktop "$current_os" "$current_arch" "$current_ext"
    
    # Try to build for other platforms (may fail if cross-compilation not supported)
    log_info "Attempting cross-platform builds..."
    
    # Linux builds
    if [ "$current_os" != "linux" ]; then
        build_desktop "linux" "amd64" "" || log_warn "Linux build failed (cross-compilation may not be supported)"
    fi
    
    # Windows builds
    if [ "$current_os" != "windows" ]; then
        build_desktop "windows" "amd64" ".exe" || log_warn "Windows build failed (cross-compilation may not be supported)"
    fi
    
    # macOS builds (require macOS host for Fyne)
    if [ "$current_os" = "darwin" ]; then
        build_desktop "darwin" "amd64" "" || log_warn "macOS build failed"
        build_desktop "darwin" "arm64" "" || log_warn "macOS ARM64 build failed"
    else
        log_warn "Skipping macOS builds (require macOS host for Fyne)"
    fi
    
    # Android build
    build_android || log_warn "Android build failed (requires Android SDK setup)"
    
    # Package with assets
    package_with_assets
    
    # Summary
    log_info "Build complete. Artifacts:"
    ls -la "$ARTIFACTS_DIR/"
    
    log_info "Cross-platform build finished successfully!"
}

# Handle script arguments
case "${1:-build}" in
    "prepare")
        prepare_env
        ;;
    "test")
        run_tests
        ;;
    "android")
        prepare_env
        build_android
        ;;
    "desktop")
        prepare_env
        build_desktop $(go env GOOS) $(go env GOARCH) ""
        ;;
    "package")
        package_with_assets
        ;;
    "build"|*)
        main
        ;;
esac
