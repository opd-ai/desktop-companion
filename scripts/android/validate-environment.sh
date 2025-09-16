#!/bin/bash

# scripts/android/validate-environment.sh
# Android Build Environment Validation Script
#
# Validates compatibility between Go, Fyne, and Android NDK versions
# and ensures proper Android development environment setup.
#
# Usage: ./scripts/android/validate-environment.sh [OPTIONS]
#
# Dependencies:
# - Go 1.21+
# - Fyne CLI tool
# - Android SDK/NDK (optional)
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
# ANDROID ENVIRONMENT CONFIGURATION
# ============================================================================

# Android environment settings
ANDROID_HOME="${DDS_ANDROID_HOME}"
ANDROID_NDK_ROOT="${DDS_ANDROID_NDK_ROOT}"
ANDROID_MIN_SDK="${DDS_ANDROID_MIN_SDK}"
ANDROID_TARGET_SDK="${DDS_ANDROID_TARGET_SDK}"

# Validation settings
STRICT_MODE=false
INSTALL_MISSING=false

# ============================================================================
# HELP AND USAGE
# ============================================================================

show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Validate Android build environment for Desktop Companion.

OPTIONS:
    --strict              Enable strict validation (require all components)
    --install-missing     Attempt to install missing components
    --show-compatibility  Show known compatibility matrix
    --help               Show this help message

VALIDATION CHECKS:
    ✓ Go version compatibility
    ✓ Fyne CLI tool availability and version
    ✓ Android SDK/NDK presence and versions
    ✓ Known compatibility combinations
    ✓ Environment variable configuration

ENVIRONMENT VARIABLES:
    ANDROID_HOME         Android SDK installation path
    ANDROID_NDK_ROOT     Android NDK installation path
    DDS_ANDROID_MIN_SDK  Minimum Android SDK version (default: $ANDROID_MIN_SDK)
    DDS_ANDROID_TARGET_SDK Target Android SDK version (default: $ANDROID_TARGET_SDK)

EXAMPLES:
    $0                           # Basic environment validation
    $0 --strict                  # Strict validation (fail if anything missing)
    $0 --install-missing         # Try to install missing components
    $0 --show-compatibility      # Show compatibility matrix

See: docs/ANDROID_BUILD_GUIDE.md for detailed setup instructions.
EOF
}

# ============================================================================
# COMPATIBILITY MATRIX
# ============================================================================

# Show known compatibility combinations
show_compatibility_matrix() {
    cat << EOF
Android Build Environment Compatibility Matrix
=============================================

KNOWN COMPATIBLE COMBINATIONS:

Fyne v2.5.x + NDK 25.x-27.x:
  ✅ Recommended for new projects
  ✅ Support for latest Android features
  ✅ Go 1.21+ required

Fyne v2.4.x + NDK 25.x:
  ✅ Stable combination
  ✅ Good for production builds
  ✅ Go 1.19+ required

Fyne v2.3.x + NDK 23.x-25.x:
  ⚠️  Legacy support
  ⚠️  Limited Android 14+ features
  ✅ Go 1.18+ required

ANDROID SDK REQUIREMENTS:

Minimum SDK: API Level $ANDROID_MIN_SDK (Android 5.0)
Target SDK:  API Level $ANDROID_TARGET_SDK (Android 14)
Build Tools: 34.0.0 or higher
Platform Tools: Latest version

NDK REQUIREMENTS:

Recommended: NDK 26.x or 27.x
Minimum:     NDK 23.x
Architecture: arm64-v8a, armeabi-v7a

GO VERSION COMPATIBILITY:

Fyne v2.5.x: Go 1.21+
Fyne v2.4.x: Go 1.19+
CGO Required: Yes (for native Android builds)

COMMON ISSUES:

❌ Cross-compilation not supported (CGO requirement)
❌ NDK path with spaces causes build failures
❌ Outdated Java version conflicts
❌ Missing platform-tools in PATH

EOF
}

# ============================================================================
# VALIDATION FUNCTIONS
# ============================================================================

# Check Go version and compatibility
validate_go_version() {
    log "Validating Go environment..."
    
    # Get Go version
    local go_version
    if ! go_version=$(go version 2>/dev/null); then
        error "Go is not installed or not in PATH"
        return 1
    fi
    
    success "Go version: $go_version"
    
    # Extract version number
    local version_number
    version_number=$(echo "$go_version" | grep -o 'go[0-9]\+\.[0-9]\+' | sed 's/go//')
    
    # Check minimum version (1.18 for basic Fyne support)
    local major minor
    IFS='.' read -r major minor <<< "$version_number"
    
    if [[ $major -lt 1 ]] || [[ $major -eq 1 && $minor -lt 18 ]]; then
        error "Go version too old. Minimum required: 1.18, found: $version_number"
        return 1
    fi
    
    if [[ $major -eq 1 && $minor -lt 21 ]]; then
        warning "Go version $version_number may have limited Fyne compatibility"
        warning "Recommended: Go 1.21+ for best Android support"
    fi
    
    # Check CGO support
    if [[ "$(go env CGO_ENABLED)" != "1" ]]; then
        error "CGO is disabled. Android builds require CGO_ENABLED=1"
        return 1
    fi
    
    success "Go environment is compatible"
    return 0
}

# Check Fyne CLI tool
validate_fyne_cli() {
    log "Validating Fyne CLI tool..."
    
    if ! command_exists fyne; then
        if [[ "$INSTALL_MISSING" == "true" ]]; then
            log "Installing Fyne CLI tool..."
            go install fyne.io/tools/cmd/fyne@latest
            
            if ! command_exists fyne; then
                error "Failed to install Fyne CLI tool"
                return 1
            fi
        else
            error "Fyne CLI tool not found"
            error "Install with: go install fyne.io/tools/cmd/fyne@latest"
            return 1
        fi
    fi
    
    # Get Fyne version
    local fyne_version
    if fyne_version=$(fyne version 2>/dev/null); then
        success "Fyne CLI version: $fyne_version"
    else
        warning "Could not determine Fyne CLI version"
    fi
    
    return 0
}

# Check Android SDK
validate_android_sdk() {
    log "Validating Android SDK..."
    
    if [[ -z "$ANDROID_HOME" ]]; then
        if [[ "$STRICT_MODE" == "true" ]]; then
            error "ANDROID_HOME environment variable not set"
            return 1
        else
            warning "ANDROID_HOME not set. Some Android features may not work."
            warning "Set ANDROID_HOME to your Android SDK installation path"
            return 0
        fi
    fi
    
    if [[ ! -d "$ANDROID_HOME" ]]; then
        error "ANDROID_HOME directory does not exist: $ANDROID_HOME"
        return 1
    fi
    
    success "Android SDK found: $ANDROID_HOME"
    
    # Check for required SDK components
    local sdk_components=(
        "platform-tools"
        "build-tools"
        "platforms"
    )
    
    for component in "${sdk_components[@]}"; do
        if [[ -d "$ANDROID_HOME/$component" ]]; then
            debug "Found SDK component: $component"
        else
            warning "Missing SDK component: $component"
        fi
    done
    
    # Check for adb in PATH
    if command_exists adb; then
        debug "ADB is available in PATH"
    else
        warning "ADB not found in PATH. Add \$ANDROID_HOME/platform-tools to PATH"
    fi
    
    return 0
}

# Check Android NDK
validate_android_ndk() {
    log "Validating Android NDK..."
    
    if [[ -z "$ANDROID_NDK_ROOT" ]]; then
        # Try to find NDK in ANDROID_HOME
        if [[ -n "$ANDROID_HOME" && -d "$ANDROID_HOME/ndk" ]]; then
            local ndk_versions
            readarray -t ndk_versions < <(find "$ANDROID_HOME/ndk" -maxdepth 1 -type d -name "*.*.*" | sort -V)
            
            if [[ ${#ndk_versions[@]} -gt 0 ]]; then
                ANDROID_NDK_ROOT="${ndk_versions[-1]}" # Use latest version
                warning "ANDROID_NDK_ROOT not set, using: $ANDROID_NDK_ROOT"
            fi
        fi
    fi
    
    if [[ -z "$ANDROID_NDK_ROOT" ]]; then
        if [[ "$STRICT_MODE" == "true" ]]; then
            error "Android NDK not found"
            error "Set ANDROID_NDK_ROOT or install NDK in \$ANDROID_HOME/ndk/"
            return 1
        else
            warning "Android NDK not found. Basic APK builds may still work."
            return 0
        fi
    fi
    
    if [[ ! -d "$ANDROID_NDK_ROOT" ]]; then
        error "ANDROID_NDK_ROOT directory does not exist: $ANDROID_NDK_ROOT"
        return 1
    fi
    
    # Get NDK version
    local ndk_version
    ndk_version=$(basename "$ANDROID_NDK_ROOT")
    success "Android NDK found: $ndk_version"
    
    # Check NDK compatibility
    case "$ndk_version" in
        27.*|26.*|25.*)
            success "NDK version is compatible"
            ;;
        24.*|23.*)
            warning "NDK version $ndk_version may have limited compatibility"
            warning "Consider upgrading to NDK 25.x or newer"
            ;;
        *)
            warning "Unknown NDK version: $ndk_version"
            warning "Compatibility not verified"
            ;;
    esac
    
    return 0
}

# Validate overall compatibility
validate_compatibility() {
    log "Validating environment compatibility..."
    
    # Get versions
    local go_version fyne_version ndk_version
    go_version=$(go version 2>/dev/null | grep -o 'go[0-9]\+\.[0-9]\+' | sed 's/go//' || echo "unknown")
    fyne_version=$(fyne version 2>/dev/null | grep -o 'v[0-9]\+\.[0-9]\+' || echo "unknown")
    ndk_version=$(basename "${ANDROID_NDK_ROOT:-unknown}")
    
    log "Environment summary:"
    log "  Go: $go_version"
    log "  Fyne: $fyne_version"
    log "  NDK: $ndk_version"
    
    # Check known compatible combinations
    local compatible=false
    
    if [[ "$fyne_version" =~ ^v2\.5\. ]] && [[ "$ndk_version" =~ ^(25|26|27)\. ]]; then
        success "✅ Fyne v2.5.x is compatible with NDK 25.x-27.x"
        compatible=true
    elif [[ "$fyne_version" =~ ^v2\.4\. ]] && [[ "$ndk_version" =~ ^25\. ]]; then
        success "✅ Fyne v2.4.x is compatible with NDK 25.x"
        compatible=true
    elif [[ "$fyne_version" == "unknown" || "$ndk_version" == "unknown" ]]; then
        warning "⚠️ Cannot verify compatibility (missing version information)"
    else
        warning "⚠️ Untested compatibility: Fyne $fyne_version with NDK $ndk_version"
        warning "This combination may work but has not been verified"
    fi
    
    if [[ "$compatible" == "true" ]]; then
        success "Environment compatibility verified"
        return 0
    elif [[ "$STRICT_MODE" == "true" ]]; then
        error "Environment compatibility check failed"
        return 1
    else
        warning "Environment compatibility could not be verified"
        return 0
    fi
}

# ============================================================================
# MAIN VALIDATION FUNCTION
# ============================================================================

# Run complete environment validation
validate_environment() {
    log "Starting Android environment validation..."
    
    local validation_errors=0
    
    # Run individual validations
    validate_go_version || ((validation_errors++))
    validate_fyne_cli || ((validation_errors++))
    validate_android_sdk || ((validation_errors++))
    validate_android_ndk || ((validation_errors++))
    validate_compatibility || ((validation_errors++))
    
    echo
    if [[ $validation_errors -eq 0 ]]; then
        success "✅ Android environment validation passed!"
        log "Environment is ready for Android APK builds."
    else
        if [[ "$STRICT_MODE" == "true" ]]; then
            error "❌ Android environment validation failed ($validation_errors errors)"
            log "Fix the errors above before attempting Android builds."
            return 1
        else
            warning "⚠️ Android environment has $validation_errors warnings"
            log "Basic builds may still work, but some features may be limited."
        fi
    fi
    
    return 0
}

# ============================================================================
# ARGUMENT PARSING AND MAIN
# ============================================================================

# Parse command line arguments
parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_usage
                exit 0
                ;;
            --strict)
                STRICT_MODE=true
                shift
                ;;
            --install-missing)
                INSTALL_MISSING=true
                shift
                ;;
            --show-compatibility)
                show_compatibility_matrix
                exit 0
                ;;
            *)
                error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done
}

# Main entry point
main() {
    # Set up error handling
    setup_error_handling
    init_common
    
    # Parse arguments
    parse_arguments "$@"
    
    # Run validation
    validate_environment
}

# Run main function if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
