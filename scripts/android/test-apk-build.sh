#!/bin/bash

# scripts/android/test-apk-build.sh
# Android APK build testing script
#
# Tests the Android APK build process for different characters
# and validates the generated APK files.
#
# Usage: ./scripts/android/test-apk-build.sh [OPTIONS] [CHARACTER]
#
# Dependencies:
# - Go 1.21+
# - Fyne CLI tool
# - scripts/lib/common.sh
# - scripts/lib/config.sh
# - scripts/build/build-characters.sh

# Load shared libraries
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LIB_DIR="$(dirname "$SCRIPT_DIR")/lib"

# shellcheck source=../lib/common.sh
source "$LIB_DIR/common.sh"
# shellcheck source=../lib/config.sh
source "$LIB_DIR/config.sh"

# ============================================================================
# TEST CONFIGURATION
# ============================================================================

# Test settings
TEST_CHARACTER="${1:-default}"
TEST_ARCHITECTURES=("arm64" "arm")
DRY_RUN="${DDS_DRY_RUN}"
CLEAN_AFTER_TEST=true

# APK validation settings
VALIDATE_APK_INTEGRITY=true
VALIDATE_APK_CONTENTS=true

# ============================================================================
# HELP AND USAGE
# ============================================================================

show_usage() {
    cat << EOF
Usage: $0 [OPTIONS] [CHARACTER]

Test Android APK build process for Desktop Companion characters.

ARGUMENTS:
    CHARACTER            Character to build (default: default)

OPTIONS:
    --dry-run           Simulate build without creating actual APK
    --no-clean          Don't clean test artifacts after completion
    --no-validation     Skip APK validation steps
    --arch ARCH         Test specific architecture only (arm64, arm)
    --help             Show this help message

TEST PROCESS:
    1. Validate Android environment
    2. Create test build directory
    3. Generate embedded character code
    4. Build APK for specified architectures
    5. Validate APK integrity and contents
    6. Clean up test artifacts (optional)

EXAMPLES:
    $0                          # Test default character
    $0 romance                  # Test romance character
    $0 --arch arm64 default     # Test only arm64 architecture
    $0 --dry-run --no-clean     # Simulate build, keep artifacts

See: docs/ANDROID_BUILD_GUIDE.md for detailed build information.
EOF
}

# ============================================================================
# ENVIRONMENT VALIDATION
# ============================================================================

# Validate Android build environment
validate_test_environment() {
    log "Validating Android build environment..."
    
    # Check if Android environment validation script exists
    local android_env_script="$SCRIPT_DIR/validate-environment.sh"
    if [[ -x "$android_env_script" ]]; then
        if ! "$android_env_script"; then
            error "Android environment validation failed"
            return 1
        fi
    else
        warning "Android environment validation script not found"
        
        # Basic checks
        require_commands go fyne
        
        if ! command_exists fyne; then
            error "Fyne CLI tool not found. Install with: go install fyne.io/tools/cmd/fyne@latest"
            return 1
        fi
    fi
    
    success "Android build environment is ready"
    return 0
}

# ============================================================================
# APK BUILD TESTING
# ============================================================================

# Create test environment
setup_test_environment() {
    local character="$1"
    
    log "Setting up test environment for character: $character"
    
    # Create test directory
    local test_dir="$BUILD_DIR/android-test-$(date +%Y%m%d-%H%M%S)"
    mkdir -p "$test_dir"
    
    # Export test directory for cleanup
    export TEST_DIR="$test_dir"
    
    log "Test directory: $test_dir"
    
    # Verify character exists
    local character_file="$CHARACTERS_DIR/$character/character.json"
    if [[ ! -f "$character_file" ]]; then
        error "Character not found: $character"
        error "Character file expected: $character_file"
        return 1
    fi
    
    success "Test environment setup complete"
    return 0
}

# Test APK build for specific architecture
test_apk_build_architecture() {
    local character="$1"
    local architecture="$2"
    
    log "Testing APK build: $character for $architecture"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log "DRY RUN: Would build APK for $character ($architecture)"
        return 0
    fi
    
    # Use the character build script for APK generation
    local build_script="$SCRIPT_DIR/../build/build-characters.sh"
    
    if [[ ! -x "$build_script" ]]; then
        error "Character build script not found: $build_script"
        return 1
    fi
    
    # Set Android-specific platform and build
    log "Building APK using character build system..."
    
    if PLATFORMS="android/$architecture" "$build_script" build "$character"; then
        success "APK build succeeded for $character ($architecture)"
        
        # Find generated APK
        local apk_file="$BUILD_DIR/${character}_android_${architecture}.apk"
        if [[ -f "$apk_file" ]]; then
            export GENERATED_APK="$apk_file"
            log "Generated APK: $(basename "$apk_file")"
            return 0
        else
            error "APK file not found: $apk_file"
            return 1
        fi
    else
        error "APK build failed for $character ($architecture)"
        return 1
    fi
}

# ============================================================================
# APK VALIDATION
# ============================================================================

# Validate APK file integrity
validate_apk_file() {
    local apk_file="$1"
    
    log "Validating APK integrity: $(basename "$apk_file")"
    
    if [[ ! -f "$apk_file" ]]; then
        error "APK file not found: $apk_file"
        return 1
    fi
    
    # Check if file is a valid ZIP (APK is a ZIP file)
    if file "$apk_file" | grep -q "Zip archive"; then
        success "APK file is valid ZIP archive"
    else
        error "APK file is not a valid ZIP archive"
        return 1
    fi
    
    # Check APK size (should be reasonable)
    local size_mb
    size_mb=$(du -m "$apk_file" | cut -f1)
    log "APK size: ${size_mb}MB"
    
    if [[ $size_mb -gt 100 ]]; then
        warning "APK size is very large: ${size_mb}MB"
        warning "Consider optimizing assets or build settings"
    fi
    
    return 0
}

# Validate APK contents
validate_apk_contents() {
    local apk_file="$1"
    local character="$2"
    
    log "Validating APK contents..."
    
    # Create temporary directory for APK extraction
    local temp_dir
    temp_dir=$(mktemp -d)
    
    # Extract APK contents
    if ! unzip -q "$apk_file" -d "$temp_dir"; then
        error "Failed to extract APK contents"
        rm -rf "$temp_dir"
        return 1
    fi
    
    # Check required files
    local required_files=(
        "AndroidManifest.xml"
        "classes.dex"
        "resources.arsc"
    )
    
    local missing_files=()
    for file in "${required_files[@]}"; do
        if [[ ! -f "$temp_dir/$file" ]]; then
            missing_files+=("$file")
        fi
    done
    
    if [[ ${#missing_files[@]} -gt 0 ]]; then
        error "Missing required APK files: ${missing_files[*]}"
        rm -rf "$temp_dir"
        return 1
    fi
    
    # Check for native libraries
    if [[ -d "$temp_dir/lib" ]]; then
        log "Native libraries found:"
        find "$temp_dir/lib" -name "*.so" | while read -r lib; do
            debug "  $(basename "$lib")"
        done
    else
        warning "No native libraries found in APK"
    fi
    
    # Check for assets
    if [[ -d "$temp_dir/assets" ]]; then
        log "Assets found in APK"
    else
        warning "No assets directory found in APK"
    fi
    
    # Cleanup
    rm -rf "$temp_dir"
    
    success "APK contents validation passed"
    return 0
}

# ============================================================================
# TEST EXECUTION
# ============================================================================

# Run complete APK build test
run_apk_build_test() {
    local character="$1"
    
    log "Starting APK build test for character: $character"
    
    # Setup test environment
    setup_test_environment "$character"
    
    local test_failures=0
    local test_successes=0
    
    # Test each architecture
    for arch in "${TEST_ARCHITECTURES[@]}"; do
        log "Testing architecture: $arch"
        
        if test_apk_build_architecture "$character" "$arch"; then
            ((test_successes++))
            
            # Validate generated APK if validation is enabled
            if [[ "$VALIDATE_APK_INTEGRITY" == "true" && -n "${GENERATED_APK:-}" ]]; then
                if validate_apk_file "$GENERATED_APK"; then
                    success "APK integrity validation passed for $arch"
                else
                    warning "APK integrity validation failed for $arch"
                    ((test_failures++))
                fi
            fi
            
            if [[ "$VALIDATE_APK_CONTENTS" == "true" && -n "${GENERATED_APK:-}" ]]; then
                if validate_apk_contents "$GENERATED_APK" "$character"; then
                    success "APK contents validation passed for $arch"
                else
                    warning "APK contents validation failed for $arch"
                    ((test_failures++))
                fi
            fi
        else
            ((test_failures++))
        fi
        
        echo
    done
    
    # Report results
    log "APK build test completed:"
    log "  Architectures tested: ${#TEST_ARCHITECTURES[@]}"
    log "  Successful builds: $test_successes"
    log "  Failed builds: $test_failures"
    
    if [[ $test_failures -eq 0 ]]; then
        success "✅ All APK build tests passed!"
        return 0
    else
        error "❌ APK build test had $test_failures failures"
        return 1
    fi
}

# ============================================================================
# CLEANUP
# ============================================================================

# Clean up test artifacts
cleanup_test_environment() {
    if [[ "$CLEAN_AFTER_TEST" == "true" && -n "${TEST_DIR:-}" && -d "$TEST_DIR" ]]; then
        log "Cleaning up test environment..."
        rm -rf "$TEST_DIR"
        debug "Removed test directory: $TEST_DIR"
    fi
    
    # Clean up any temporary embedded character directories
    local embedded_dirs
    readarray -t embedded_dirs < <(find "$PROJECT_ROOT/cmd" -name "*-embedded" -type d 2>/dev/null || true)
    
    for dir in "${embedded_dirs[@]}"; do
        if [[ -n "$dir" && -d "$dir" ]]; then
            rm -rf "$dir"
            debug "Removed embedded directory: $(basename "$dir")"
        fi
    done
}

# ============================================================================
# ARGUMENT PARSING AND MAIN
# ============================================================================

# Parse command line arguments
parse_arguments() {
    local character="default"
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_usage
                exit 0
                ;;
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            --no-clean)
                CLEAN_AFTER_TEST=false
                shift
                ;;
            --no-validation)
                VALIDATE_APK_INTEGRITY=false
                VALIDATE_APK_CONTENTS=false
                shift
                ;;
            --arch)
                TEST_ARCHITECTURES=("$2")
                shift 2
                ;;
            -*)
                error "Unknown option: $1"
                show_usage
                exit 1
                ;;
            *)
                character="$1"
                shift
                ;;
        esac
    done
    
    echo "$character"
}

# Main entry point
main() {
    # Set up error handling with cleanup
    setup_error_handling cleanup_test_environment
    init_common
    
    # Parse arguments
    local character
    character=$(parse_arguments "$@")
    
    # Validate environment
    validate_test_environment
    
    # Run APK build test
    run_apk_build_test "$character"
    
    # Cleanup
    cleanup_test_environment
}

# Run main function if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
