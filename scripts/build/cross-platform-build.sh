#!/bin/bash

# scripts/build/cross-platform-build.sh
# Cross-Platform CI/CD Build Script for Desktop Companion
# 
# This script handles building for multiple platforms including Android APK
# in CI/CD environments. Optimized for GitHub Actions and other automation.
#
# Usage: ./scripts/build/cross-platform-build.sh [OPTIONS]
#
# Dependencies:
# - Go 1.21+
# - Fyne CLI tool (for Android builds)
# - scripts/lib/common.sh
# - scripts/lib/config.sh

# Load shared libraries
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LIB_DIR="$(dirname "$SCRIPT_DIR")/lib"

# shellcheck source=../lib/common.sh
source "$LIB_DIR/common.sh"
# shellcheck source=../lib/config.sh
source "$LIB_DIR/config.sh"

# ============================================================================
# CI/CD SPECIFIC CONFIGURATION
# ============================================================================

# CI/CD Build configuration
VERSION="${DDS_APP_VERSION}"
BUILD_NUMBER="${DDS_BUILD_NUMBER}"
APP_ID="${DDS_APP_ID}"
ARTIFACTS_DIR="$BUILD_DIR/artifacts"

# ============================================================================
# ENVIRONMENT VALIDATION
# ============================================================================

# Check if fyne tool is available for Android builds
check_fyne_tool() {
    if ! command_exists fyne; then
        log_warn "Fyne CLI tool not found. Installing..."
        go install fyne.io/tools/cmd/fyne@latest
        
        if ! command_exists fyne; then
            log_error "Failed to install fyne CLI tool"
            exit 1
        fi
    fi
    log_info "Fyne CLI version: $(fyne version)"
}

# Prepare build environment for CI/CD
prepare_build_environment() {
    log_info "Preparing build environment..."
    
    # Create directories
    ensure_directories
    mkdir -p "$ARTIFACTS_DIR"
    
    # Install dependencies
    log_info "Installing Go dependencies..."
    go mod download
    go mod tidy
    
    # Install/update fyne tool
    check_fyne_tool
    
    # Validate Go environment
    validate_go_environment
    
    success "Build environment prepared"
}

# ============================================================================
# TESTING
# ============================================================================

# Run comprehensive tests
run_tests() {
    log_info "Running test suite..."
    
    # Run unit tests with coverage
    log_info "Running unit tests..."
    if ! go test ./... -v -cover -race; then
        log_error "Unit tests failed"
        return 1
    fi
    
    # Generate coverage report
    log_info "Generating coverage report..."
    go test ./... -coverprofile="$ARTIFACTS_DIR/coverage.out"
    go tool cover -html="$ARTIFACTS_DIR/coverage.out" -o "$ARTIFACTS_DIR/coverage.html"
    
    # Run character validation tests
    log_info "Running character validation..."
    if [[ -x "$SCRIPT_DIR/../validation/validate-characters.sh" ]]; then
        "$SCRIPT_DIR/../validation/validate-characters.sh"
    fi
    
    success "Tests completed successfully"
}

# ============================================================================
# PLATFORM BUILDS
# ============================================================================

# Build for desktop platforms
build_desktop_platform() {
    local platform="$1"
    local arch="$2"
    local ext="$3"
    
    log_info "Building for $platform/$arch..."
    
    local output="$ARTIFACTS_DIR/companion-$platform-$arch$ext"
    
    if GOOS="$platform" GOARCH="$arch" go build -ldflags="$DDS_LDFLAGS" -o "$output" cmd/companion/main.go; then
        success "Built desktop binary: $(basename "$output")"
        
        # Verify binary
        if [[ -f "$output" ]]; then
            local size_mb
            size_mb=$(du -m "$output" | cut -f1)
            log_info "Binary size: ${size_mb}MB"
            return 0
        fi
    fi
    
    log_error "Failed to build for $platform/$arch"
    return 1
}

# Build Android APK using character build system
build_android_apk() {
    local character="${1:-default}"
    
    log_info "Building Android APK for character: $character"
    
    # Check for Android SDK (optional - fyne can work without it in some cases)
    if [[ -z "${ANDROID_HOME:-}" ]]; then
        log_warn "ANDROID_HOME not set. Android build may not work without Android SDK."
    fi
    
    # Use the character build script for Android builds
    local character_build_script="$SCRIPT_DIR/build-characters.sh"
    
    if [[ -x "$character_build_script" ]]; then
        # Build for both Android architectures
        for arch in arm64 arm; do
            log_info "Building Android APK for $arch..."
            
            # Set platform for Android build
            PLATFORMS="android/$arch" "$character_build_script" build "$character"
            
            # Check if APK was created
            local apk_file="$BUILD_DIR/${character}_android_${arch}.apk"
            if [[ -f "$apk_file" ]]; then
                cp "$apk_file" "$ARTIFACTS_DIR/"
                success "Android APK created: $(basename "$apk_file")"
            else
                log_error "Android APK not found: $apk_file"
                return 1
            fi
        done
    else
        log_error "Character build script not found: $character_build_script"
        return 1
    fi
    
    return 0
}

# ============================================================================
# PACKAGING
# ============================================================================

# Package builds with assets for release
package_with_assets() {
    log_info "Creating release packages with assets..."
    
    cd "$ARTIFACTS_DIR" || exit 1
    
    # Copy assets for packaging
    cp -r "$PROJECT_ROOT/assets" .
    
    # Create archives for each platform binary
    for binary in companion-*; do
        if [[ -f "$binary" && ! "$binary" =~ \.apk$ ]]; then
            local platform_name
            platform_name=$(echo "$binary" | sed 's/companion-//' | sed 's/\.[^.]*$//')
            
            local archive_name="desktop-companion-${platform_name}-v${VERSION}.tar.gz"
            
            log_info "Creating package: $archive_name"
            tar -czf "$archive_name" "$binary" assets/
            
            success "Created package: $archive_name"
        fi
    done
    
    # Package Android APKs separately
    for apk in *.apk; do
        if [[ -f "$apk" ]]; then
            local apk_name
            apk_name=$(basename "$apk" .apk)
            local package_name="${apk_name}-v${VERSION}.zip"
            
            log_info "Creating Android package: $package_name"
            zip -q "$package_name" "$apk" -r assets/
            
            success "Created Android package: $package_name"
        fi
    done
    
    cd - > /dev/null
    success "Release packaging complete"
}

# ============================================================================
# MAIN BUILD WORKFLOW
# ============================================================================

# Execute the main CI/CD build workflow
run_build_workflow() {
    local build_type="${1:-all}"
    
    log_info "Starting Desktop Companion CI/CD Build"
    log_info "Version: $VERSION, Build: $BUILD_NUMBER"
    log_info "Build type: $build_type"
    
    # Prepare environment
    prepare_build_environment
    
    # Run tests first
    if [[ "${SKIP_TESTS:-false}" != "true" ]]; then
        run_tests
    else
        log_warn "Skipping tests (SKIP_TESTS=true)"
    fi
    
    # Build for current platform (always works)
    local current_os current_arch current_ext
    current_os=$(go env GOOS)
    current_arch=$(go env GOARCH)
    current_ext=""
    
    if [[ "$current_os" == "windows" ]]; then
        current_ext=".exe"
    fi
    
    log_info "Building for current platform: $current_os/$current_arch"
    build_desktop_platform "$current_os" "$current_arch" "$current_ext"
    
    # Platform-specific builds
    case "$build_type" in
        all|desktop)
            # Try cross-compilation for other desktop platforms (may fail due to CGO)
            case "$current_os" in
                linux)
                    log_info "Attempting cross-compilation on Linux..."
                    build_desktop_platform windows amd64 .exe || log_warn "Windows cross-compilation failed"
                    build_desktop_platform darwin amd64 "" || log_warn "macOS cross-compilation failed"
                    ;;
                darwin)
                    log_info "Attempting cross-compilation on macOS..."
                    build_desktop_platform linux amd64 "" || log_warn "Linux cross-compilation failed"
                    ;;
                windows)
                    log_info "Attempting cross-compilation on Windows..."
                    build_desktop_platform linux amd64 "" || log_warn "Linux cross-compilation failed"
                    ;;
            esac
            ;;
    esac
    
    # Android builds (if requested and possible)
    if [[ "$build_type" == "all" || "$build_type" == "android" ]]; then
        if command_exists fyne; then
            build_android_apk "${CHARACTER:-default}"
        else
            log_warn "Skipping Android builds (fyne CLI not available)"
        fi
    fi
    
    # Package everything
    if [[ "${SKIP_PACKAGING:-false}" != "true" ]]; then
        package_with_assets
    else
        log_warn "Skipping packaging (SKIP_PACKAGING=true)"
    fi
    
    # Generate build report
    generate_build_report
    
    success "CI/CD build completed successfully"
}

# Generate build report
generate_build_report() {
    local report_file="$ARTIFACTS_DIR/build-report.md"
    
    cat > "$report_file" << EOF
# Desktop Companion Build Report

**Build Date:** $(date -u '+%Y-%m-%d %H:%M:%S UTC')
**Version:** $VERSION
**Build Number:** $BUILD_NUMBER
**Platform:** $(go env GOOS)/$(go env GOARCH)

## Artifacts Generated

$(find "$ARTIFACTS_DIR" -type f -name "*.tar.gz" -o -name "*.zip" -o -name "*.apk" | wc -l) release artifacts

### Desktop Binaries
$(find "$ARTIFACTS_DIR" -name "companion-*" ! -name "*.apk" | sed 's|.*/||' | while read -r f; do echo "- $f"; done)

### Android APKs
$(find "$ARTIFACTS_DIR" -name "*.apk" | sed 's|.*/||' | while read -r f; do echo "- $f"; done)

### Release Packages
$(find "$ARTIFACTS_DIR" -name "*.tar.gz" -o -name "*.zip" | sed 's|.*/||' | while read -r f; do echo "- $f"; done)

## Build Environment

- Go Version: $(go version)
- Fyne Version: $(fyne version 2>/dev/null || echo "Not available")
- Platform: $(uname -a)

## Test Results

$(if [[ -f "$ARTIFACTS_DIR/coverage.out" ]]; then
    echo "- Coverage report: coverage.html"
    echo "- Coverage data: coverage.out"
else
    echo "- Tests skipped or failed"
fi)

EOF
    
    log_info "Build report generated: $(basename "$report_file")"
}

# ============================================================================
# ARGUMENT PARSING AND MAIN
# ============================================================================

# Show usage information
show_usage() {
    cat << EOF
Usage: $0 [OPTIONS] [BUILD_TYPE]

Cross-platform CI/CD build script for Desktop Companion.

BUILD_TYPE:
    all       Build for all platforms (default)
    desktop   Build desktop platforms only
    android   Build Android APKs only
    current   Build for current platform only

OPTIONS:
    --skip-tests      Skip test execution
    --skip-packaging  Skip release packaging
    --character NAME  Character for Android builds (default: default)
    --help           Show this help message

ENVIRONMENT VARIABLES:
    VERSION          Application version
    BUILD_NUMBER     CI/CD build number
    ANDROID_HOME     Android SDK path
    SKIP_TESTS       Skip tests (true/false)
    SKIP_PACKAGING   Skip packaging (true/false)

EXAMPLES:
    $0                          # Build everything
    $0 desktop                  # Desktop platforms only
    $0 android --character romance  # Android APK for romance character
    $0 --skip-tests current     # Current platform, no tests

EOF
}

# Parse command line arguments
parse_ci_arguments() {
    local build_type="all"
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --help|-h)
                show_usage
                exit 0
                ;;
            --skip-tests)
                export SKIP_TESTS=true
                shift
                ;;
            --skip-packaging)
                export SKIP_PACKAGING=true
                shift
                ;;
            --character)
                export CHARACTER="$2"
                shift 2
                ;;
            all|desktop|android|current)
                build_type="$1"
                shift
                ;;
            *)
                log_error "Unknown argument: $1"
                show_usage
                exit 1
                ;;
        esac
    done
    
    echo "$build_type"
}

# Main entry point
main() {
    # Set up error handling
    setup_error_handling
    
    # Parse arguments
    local build_type
    build_type=$(parse_ci_arguments "$@")
    
    # Run the build workflow
    run_build_workflow "$build_type"
}

# Run main function if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
