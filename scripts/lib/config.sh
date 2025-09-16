#!/bin/bash

# scripts/lib/config.sh
# Configuration settings for Desktop Companion scripts
#
# This file contains default configuration values that can be overridden
# by environment variables or command-line arguments.
#
# Usage: source scripts/lib/config.sh

# Prevent multiple inclusion
[[ -n "${DDS_CONFIG_LOADED:-}" ]] && return 0
DDS_CONFIG_LOADED=1

# ============================================================================
# BUILD CONFIGURATION
# ============================================================================

# Default build settings
export DDS_MAX_PARALLEL="${MAX_PARALLEL:-4}"
export DDS_LDFLAGS="${LDFLAGS:--s -w}"
export DDS_PLATFORMS="${PLATFORMS:-$(go env GOOS)/$(go env GOARCH)}"

# Android build settings
export DDS_APP_ID="${APP_ID:-ai.opd.dds}"
export DDS_APP_VERSION="${VERSION:-1.0.0}"
export DDS_BUILD_NUMBER="${BUILD_NUMBER:-${GITHUB_RUN_NUMBER:-1}}"

# Build output settings
export DDS_ENABLE_ARTIFACT_MGMT="${ENABLE_ARTIFACT_MGMT:-true}"
export DDS_COMPRESSION_LEVEL="${COMPRESSION_LEVEL:-6}"

# ============================================================================
# VALIDATION CONFIGURATION
# ============================================================================

# Validation timeouts and limits
export DDS_VALIDATION_TIMEOUT="${VALIDATION_TIMEOUT:-30}"
export DDS_MEMORY_LIMIT_MB="${MEMORY_LIMIT_MB:-100}"
export DDS_STARTUP_TIME_LIMIT_SEC="${STARTUP_TIME_LIMIT_SEC:-5}"

# Performance targets
export DDS_TARGET_MEMORY_MB="${TARGET_MEMORY_MB:-50}"
export DDS_TARGET_FPS="${TARGET_FPS:-30}"
export DDS_TEST_TIMEOUT="${TEST_TIMEOUT:-60s}"

# ============================================================================
# CHARACTER ASSET GENERATION
# ============================================================================

# Asset generation settings
export DDS_BACKUP_ASSETS="${BACKUP_ASSETS:-true}"
export DDS_VALIDATE_ASSETS="${VALIDATE_ASSETS:-true}"
export DDS_FORCE_REBUILD="${FORCE_REBUILD:-false}"

# Default styling options
export DDS_DEFAULT_STYLE="${DEFAULT_STYLE:-anime}"
export DDS_DEFAULT_MODEL="${DEFAULT_MODEL:-sd15}"

# ============================================================================
# RELEASE VALIDATION CONFIGURATION
# ============================================================================

# Release validation settings  
export DDS_QUICK_MODE="${QUICK_MODE:-false}"
export DDS_ANDROID_TESTS="${ANDROID_TESTS:-true}"
export DDS_PARALLEL_TESTS="${PARALLEL_TESTS:-4}"

# ============================================================================
# Asset generation defaults
export DDS_DEFAULT_STYLE="${STYLE:-anime}"
export DDS_DEFAULT_MODEL="${MODEL:-sd15}"
export DDS_PARALLEL_JOBS="${PARALLEL_JOBS:-2}"

# Asset validation settings
export DDS_BACKUP_ASSETS="${BACKUP_ASSETS:-true}"
export DDS_VALIDATE_ASSETS="${VALIDATE_ASSETS:-true}"
export DDS_FORCE_REBUILD="${FORCE_REBUILD:-false}"

# ============================================================================
# ANDROID SPECIFIC CONFIGURATION
# ============================================================================

# Android SDK and NDK settings
export DDS_ANDROID_HOME="${ANDROID_HOME:-}"
export DDS_ANDROID_NDK_ROOT="${ANDROID_NDK_ROOT:-}"

# Android build settings
export DDS_ANDROID_TESTS="${ANDROID_TESTS:-true}"
export DDS_ANDROID_MIN_SDK="${ANDROID_MIN_SDK:-21}"
export DDS_ANDROID_TARGET_SDK="${ANDROID_TARGET_SDK:-34}"

# ============================================================================
# PLATFORM MATRIX CONFIGURATION
# ============================================================================

# Supported platform combinations
readonly DDS_SUPPORTED_PLATFORMS=(
    "linux/amd64"
    "windows/amd64"
    "darwin/amd64"
    "android/arm64"
    "android/arm"
)

# Platform-specific extensions
declare -A DDS_PLATFORM_EXTENSIONS
DDS_PLATFORM_EXTENSIONS["windows/amd64"]=".exe"
DDS_PLATFORM_EXTENSIONS["linux/amd64"]=""
DDS_PLATFORM_EXTENSIONS["darwin/amd64"]=""

# ============================================================================
# ANIMATION REQUIREMENTS
# ============================================================================

# Basic animations required for all characters
readonly DDS_BASIC_ANIMATIONS=(
    "idle.gif"
    "talking.gif"
    "happy.gif"
    "sad.gif"
    "hungry.gif"
    "eating.gif"
)

# Character-specific animation sets
declare -A DDS_SPECIFIC_ANIMATIONS
DDS_SPECIFIC_ANIMATIONS[romance]="blushing.gif heart_eyes.gif shy.gif flirty.gif romantic_idle.gif"
DDS_SPECIFIC_ANIMATIONS[tsundere]="smug.gif proud.gif embarrassed.gif defensive.gif"
DDS_SPECIFIC_ANIMATIONS[flirty]="wink.gif seductive.gif playful.gif teasing.gif"

# ============================================================================
# VALIDATION RULES
# ============================================================================

# Valid categories for character events
readonly DDS_VALID_CATEGORIES=(
    "conversation"
    "roleplay"
    "game"
    "humor"
    "romance"
)

# Valid trigger types
readonly DDS_VALID_TRIGGERS=(
    "click"
    "rightclick"
    "hover"
    "ctrl+shift+click"
    "alt+shift+click"
)

# ============================================================================
# LOGGING CONFIGURATION
# ============================================================================

# Log levels (for future enhancement)
export DDS_LOG_LEVEL="${LOG_LEVEL:-INFO}"
export DDS_DEBUG="${DEBUG:-0}"
export DDS_VERBOSE="${VERBOSE:-false}"

# Log file settings
export DDS_LOG_FILE="${LOG_FILE:-}"
export DDS_LOG_ROTATION="${LOG_ROTATION:-false}"

# ============================================================================
# DEVELOPMENT SETTINGS
# ============================================================================

# Development mode settings
export DDS_DEV_MODE="${DEV_MODE:-false}"
export DDS_QUICK_MODE="${QUICK_MODE:-false}"
export DDS_DRY_RUN="${DRY_RUN:-false}"

# Testing settings
export DDS_SKIP_TESTS="${SKIP_TESTS:-false}"
export DDS_SKIP_INTEGRATION_TESTS="${SKIP_INTEGRATION_TESTS:-false}"
export DDS_COVERAGE_THRESHOLD="${COVERAGE_THRESHOLD:-80}"

# ============================================================================
# UTILITY FUNCTIONS
# ============================================================================

# Show current configuration
show_config() {
    cat << EOF
Desktop Companion Scripts Configuration
======================================

Build Settings:
  Max Parallel Jobs: $DDS_MAX_PARALLEL
  Platforms: $DDS_PLATFORMS
  LDFLAGS: $DDS_LDFLAGS
  Artifact Management: $DDS_ENABLE_ARTIFACT_MGMT

Android Settings:
  App ID: $DDS_APP_ID
  Version: $DDS_APP_VERSION
  Build Number: $DDS_BUILD_NUMBER
  Android Home: ${DDS_ANDROID_HOME:-"Not set"}

Asset Generation:
  Default Style: $DDS_DEFAULT_STYLE
  Default Model: $DDS_DEFAULT_MODEL
  Parallel Jobs: $DDS_PARALLEL_JOBS
  Backup Assets: $DDS_BACKUP_ASSETS

Validation Settings:
  Timeout: ${DDS_VALIDATION_TIMEOUT}s
  Memory Limit: ${DDS_MEMORY_LIMIT_MB}MB
  Target FPS: $DDS_TARGET_FPS

Development:
  Debug Mode: $DDS_DEBUG
  Verbose: $DDS_VERBOSE
  Dry Run: $DDS_DRY_RUN

EOF
}

# Validate configuration
validate_config() {
    local errors=()
    
    # Check required numeric values
    if ! [[ "$DDS_MAX_PARALLEL" =~ ^[0-9]+$ ]] || [[ "$DDS_MAX_PARALLEL" -lt 1 ]]; then
        errors+=("DDS_MAX_PARALLEL must be a positive integer")
    fi
    
    if ! [[ "$DDS_VALIDATION_TIMEOUT" =~ ^[0-9]+$ ]] || [[ "$DDS_VALIDATION_TIMEOUT" -lt 1 ]]; then
        errors+=("DDS_VALIDATION_TIMEOUT must be a positive integer")
    fi
    
    # Check boolean values
    for var in DDS_ENABLE_ARTIFACT_MGMT DDS_BACKUP_ASSETS DDS_VALIDATE_ASSETS DDS_ANDROID_TESTS; do
        local value="${!var}"
        if [[ "$value" != "true" && "$value" != "false" ]]; then
            errors+=("$var must be 'true' or 'false'")
        fi
    done
    
    # Report errors
    if [[ ${#errors[@]} -gt 0 ]]; then
        echo "Configuration validation errors:" >&2
        printf "  - %s\n" "${errors[@]}" >&2
        return 1
    fi
    
    return 0
}

# Load configuration from file if it exists
load_config_file() {
    local config_file="${1:-$HOME/.dds-config}"
    
    if [[ -f "$config_file" ]]; then
        echo "Loading configuration from $config_file" >&2
        # shellcheck source=/dev/null
        source "$config_file"
    fi
}

# Save current configuration to file
save_config_file() {
    local config_file="${1:-$HOME/.dds-config}"
    
    cat > "$config_file" << EOF
# Desktop Companion Scripts Configuration
# Generated on $(date)

# Build settings
export DDS_MAX_PARALLEL="$DDS_MAX_PARALLEL"
export DDS_LDFLAGS="$DDS_LDFLAGS"
export DDS_PLATFORMS="$DDS_PLATFORMS"
export DDS_ENABLE_ARTIFACT_MGMT="$DDS_ENABLE_ARTIFACT_MGMT"

# Android settings
export DDS_APP_ID="$DDS_APP_ID"
export DDS_APP_VERSION="$DDS_APP_VERSION"
export DDS_ANDROID_HOME="$DDS_ANDROID_HOME"

# Asset generation
export DDS_DEFAULT_STYLE="$DDS_DEFAULT_STYLE"
export DDS_DEFAULT_MODEL="$DDS_DEFAULT_MODEL"
export DDS_PARALLEL_JOBS="$DDS_PARALLEL_JOBS"

# Development
export DDS_DEBUG="$DDS_DEBUG"
export DDS_VERBOSE="$DDS_VERBOSE"
EOF
    
    echo "Configuration saved to $config_file" >&2
}

# ============================================================================
# INITIALIZATION
# ============================================================================

# Initialize configuration
init_config() {
    # Load user configuration if available
    load_config_file
    
    # Validate configuration
    if ! validate_config; then
        echo "Using default configuration due to validation errors" >&2
    fi
}

# Auto-initialize when sourced
init_config

# Show configuration if run directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    show_config
fi
