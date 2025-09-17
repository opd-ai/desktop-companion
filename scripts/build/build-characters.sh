#!/bin/bash

# scripts/build/build-characters.sh
# Character-specific build automation script
# 
# This script builds character-specific binaries for the Desktop Companion.
# It supports multiple platforms, Android APK generation, and artifact management.
#
# Usage: ./scripts/build/build-characters.sh [OPTIONS] [COMMAND]
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
# SCRIPT CONFIGURATION
# ============================================================================

# Local configuration (can be overridden by environment)
MAX_PARALLEL="${DDS_MAX_PARALLEL}"
PLATFORMS="${DDS_PLATFORMS}"
LDFLAGS="${DDS_LDFLAGS}"
ENABLE_ARTIFACT_MGMT="${DDS_ENABLE_ARTIFACT_MGMT}"
SPECIFIC_CHARACTER=""
COMMAND=""

# ============================================================================
# HELP AND USAGE
# ============================================================================

show_usage() {
    cat << EOF
Usage: $0 [OPTIONS] [COMMAND]

Build character-specific binaries for Desktop Companion.

COMMANDS:
    list                List all available characters
    build [CHARACTER]   Build specific character (or all if none specified)
    clean              Clean build artifacts
    manage             Manage stored artifacts (list, cleanup, compress)
    platforms          Show platform matrix configuration and limitations
    help               Show this help message

OPTIONS:
    -p, --parallel N    Set maximum parallel builds (default: $MAX_PARALLEL)
    -o, --output DIR    Set output directory (default: $BUILD_DIR)
    --platforms LIST    Comma-separated list of platforms (default: $PLATFORMS)
    --ldflags FLAGS     Linker flags for builds (default: "$LDFLAGS")
    --no-artifact-mgmt  Disable automatic artifact management

EXAMPLES:
    $0 list                           # List available characters
    $0 build                          # Build all characters
    $0 build default                  # Build only default character
    $0 build --parallel 8             # Build with 8 parallel workers
    $0 manage                         # Access artifact management tools
    $0 clean                          # Clean build directory

ENVIRONMENT VARIABLES:
    DDS_MAX_PARALLEL       Maximum parallel builds
    DDS_PLATFORMS          Target platforms for builds
    DDS_LDFLAGS           Linker flags for optimization
    DDS_ENABLE_ARTIFACT_MGMT   Enable automatic artifact management (true/false)

See: docs/CHARACTER_BINARY_VALIDATION_GUIDE.md for more information.
EOF
}

show_platform_info() {
    cat << EOF

Platform Matrix Configuration:
===============================

Current OS: $(go env GOOS)/$(go env GOARCH)

Supported Platforms:
  • linux/amd64   - Linux 64-bit
  • windows/amd64 - Windows 64-bit  
  • darwin/amd64  - macOS 64-bit
  • android/arm64 - Android 64-bit (requires fyne CLI tool)
  • android/arm   - Android 32-bit (requires fyne CLI tool)

Cross-Compilation Limitations:
  Due to Fyne GUI framework CGO requirements, cross-compilation
  between different operating systems may fail. For production
  builds, use GitHub Actions matrix builds which run on native
  environments for each target platform.

Android Build Requirements:
  • Fyne CLI tool: go install fyne.io/tools/cmd/fyne@latest
  • Android NDK (optional for basic builds)
  • ANDROID_HOME environment variable (recommended)

For more details, see: docs/ANDROID_BUILD_GUIDE.md

EOF
}

# ============================================================================
# CHARACTER MANAGEMENT
# ============================================================================

# List all available characters
list_characters() {
    log "Available characters:"
    
    local character_files
    readarray -t character_files < <(find_character_files)
    
    if [[ ${#character_files[@]} -eq 0 ]]; then
        warning "No character files found in $CHARACTERS_DIR"
        return 1
    fi
    
    for character_file in "${character_files[@]}"; do
        local char_name
        char_name=$(get_character_name "$character_file")
        
        # Check if character has required files
        local char_dir
        char_dir=$(dirname "$character_file")
        if [[ -f "$char_dir/character.json" ]]; then
            echo "  ✓ $char_name"
        else
            echo "  ✗ $char_name (missing character.json)"
        fi
    done
}

# Get list of buildable characters (excluding examples and templates)
get_buildable_characters() {
    local character_files
    readarray -t character_files < <(find_character_files)
    
    for character_file in "${character_files[@]}"; do
        local char_name
        char_name=$(get_character_name "$character_file")
        
        # Skip example and template characters
        case "$char_name" in
            *example*|*template*|*test*)
                debug "Skipping $char_name (example/template/test)"
                ;;
            *)
                echo "$char_name"
                ;;
        esac
    done
}

# ============================================================================
# CHARACTER EMBEDDING
# ============================================================================

# Generate embedded character code using embed-character.go
generate_embedded_character() {
    local char="$1"
    
    log "Generating embedded character: $char"
    
    local output_dir="$PROJECT_ROOT/cmd/$char-embedded"
    
    # Clean previous embedded version
    [[ -d "$output_dir" ]] && rm -rf "$output_dir"
    
    # Change to project root so embed-character.go can find assets/characters/
    local original_dir="$PWD"
    cd "$PROJECT_ROOT" || {
        error "Failed to change to project root: $PROJECT_ROOT"
        return 1
    }
    
    # Use absolute path to embed-character.go script
    local embed_script="$PROJECT_ROOT/scripts/embed-character.go"
    
    if ! go run "$embed_script" -character "$char" -output "$output_dir"; then
        cd "$original_dir"
        error "Failed to generate embedded character: $char"
        return 1
    fi
    
    cd "$original_dir"
    
    success "Generated embedded character: $char"
    return 0
}

# ============================================================================
# PLATFORM-SPECIFIC BUILDS
# ============================================================================

# Validate platform compatibility
validate_platform() {
    local platform="$1"
    
    case "$platform" in
        linux/amd64|windows/amd64|darwin/amd64)
            return 0
            ;;
        android/arm64|android/arm)
            if ! command_exists fyne; then
                warning "Android builds require fyne CLI tool"
                warning "Install with: go install fyne.io/tools/cmd/fyne@latest"
                return 1
            fi
            return 0
            ;;
        *)
            warning "Unsupported platform: $platform"
            return 1
            ;;
    esac
}

# Build character for Android platform using fyne CLI tool
build_character_android() {
    local char="$1"
    local goarch="$2"
    
    log "Building Android APK for $char ($goarch)"
    
    local source_dir="$PROJECT_ROOT/cmd/$char-embedded"
    local output_file="$BUILD_DIR/${char}_android_${goarch}.apk"
    local app_name="${char^} Companion"
    local app_id="${DDS_APP_ID}.${char}"
    
    # Ensure source directory exists
    if [[ ! -d "$source_dir" ]]; then
        error "Embedded character not found: $source_dir"
        return 1
    fi
    
    # Create temporary directory for Android build
    local temp_dir
    temp_dir=$(mktemp -d)
    cp -r "$source_dir"/* "$temp_dir/"
    
    cd "$temp_dir"
    
    # Set up Go module for Android build
    log "Setting up embedded character module..."
    
    # Create FyneApp.toml for Android build
    cat > FyneApp.toml << EOF
[Details]
Icon = "Icon.png"
Name = "$app_name"
ID = "$app_id"
Version = "${DDS_APP_VERSION}"
Build = "${DDS_BUILD_NUMBER}"

[Development]
AutoInject = true
EOF
    
    # Copy icon (character-specific or default)
    local char_icon="$PROJECT_ROOT/assets/characters/$char/icon.png"
    local default_icon="$PROJECT_ROOT/assets/app/icon.png"
    
    if [[ -f "$char_icon" ]]; then
        cp "$char_icon" Icon.png
    elif [[ -f "$default_icon" ]]; then
        cp "$default_icon" Icon.png
    else
        warning "No icon found for $char, creating placeholder"
        # Create minimal PNG placeholder
        echo "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg==" | base64 -d > Icon.png
    fi
    
    # Build APK
    if fyne package --target "android/$goarch" --name "$app_name" --app-id "$app_id" --app-version "${DDS_APP_VERSION}" --release; then
        # Move generated APK to build directory
        local generated_apk
        generated_apk=$(find . -name "*.apk" -type f | head -1)
        
        if [[ -n "$generated_apk" ]]; then
            mv "$generated_apk" "$output_file"
            success "Built Android APK: $(basename "$output_file")"
            
            # Store artifact if enabled
            [[ "$ENABLE_ARTIFACT_MGMT" == "true" ]] && store_artifact "$output_file"
        else
            error "APK file not found after build"
            return 1
        fi
    else
        error "Failed to build Android APK for $char"
        return 1
    fi
    
    # Cleanup
    cd - > /dev/null
    rm -rf "$temp_dir"
    
    return 0
}

# Build character binary for a specific platform
build_character_platform() {
    local char="$1"
    local platform="$2"
    
    log "Building $char for $platform"
    
    local goos goarch
    IFS='/' read -r goos goarch <<< "$platform"
    
    # Handle Android builds specially
    if [[ "$goos" == "android" ]]; then
        build_character_android "$char" "$goarch"
        return $?
    fi
    
    # Regular Go build for desktop platforms
    local source_dir="$PROJECT_ROOT/cmd/$char-embedded"
    local ext="${DDS_PLATFORM_EXTENSIONS[$platform]:-}"
    local output_file="$BUILD_DIR/${char}_${goos}_${goarch}${ext}"
    
    if [[ ! -d "$source_dir" ]]; then
        error "Embedded character not found: $source_dir"
        return 1
    fi
    
    if GOOS="$goos" GOARCH="$goarch" go build -ldflags="$LDFLAGS" -o "$output_file" "$source_dir/main.go"; then
        success "Built: $(basename "$output_file")"
        
        # Store artifact if enabled
        [[ "$ENABLE_ARTIFACT_MGMT" == "true" ]] && store_artifact "$output_file"
        return 0
    else
        error "Failed to build $char for $platform"
        return 1
    fi
}

# ============================================================================
# MAIN BUILD FUNCTIONS
# ============================================================================

# Build a single character for all platforms
build_character() {
    local char="$1"
    
    log "Building character: $char"
    
    # Generate embedded character first
    if ! generate_embedded_character "$char"; then
        return 1
    fi
    
    # Build for each platform
    local platform
    local failed_platforms=()
    
    IFS=',' read -ra PLATFORM_LIST <<< "$PLATFORMS"
    for platform in "${PLATFORM_LIST[@]}"; do
        platform=$(echo "$platform" | xargs) # trim whitespace
        
        if validate_platform "$platform"; then
            if ! build_character_platform "$char" "$platform"; then
                failed_platforms+=("$platform")
            fi
        else
            failed_platforms+=("$platform")
        fi
    done
    
    # Report results
    if [[ ${#failed_platforms[@]} -eq 0 ]]; then
        success "Successfully built $char for all platforms"
        return 0
    else
        warning "Failed to build $char for: ${failed_platforms[*]}"
        return 1
    fi
}

# Build all characters
build_all_characters() {
    log "Building all characters..."
    
    local characters
    readarray -t characters < <(get_buildable_characters)
    
    if [[ ${#characters[@]} -eq 0 ]]; then
        warning "No buildable characters found"
        return 1
    fi
    
    log "Found ${#characters[@]} buildable characters"
    
    local success_count=0
    local failed_characters=()
    
    # Build characters in parallel if requested
    if [[ "$MAX_PARALLEL" -gt 1 ]]; then
        log "Building in parallel (max $MAX_PARALLEL jobs)"
        
        # Use GNU parallel if available, otherwise fall back to background jobs
        if command_exists parallel; then
            printf "%s\n" "${characters[@]}" | parallel -j "$MAX_PARALLEL" "$(declare -f build_character); build_character {}"
        else
            # Simple background job approach
            local pids=()
            for char in "${characters[@]}"; do
                build_character "$char" &
                pids+=($!)
                
                # Limit concurrent jobs
                if [[ ${#pids[@]} -ge "$MAX_PARALLEL" ]]; then
                    wait "${pids[0]}"
                    pids=("${pids[@]:1}")
                fi
            done
            
            # Wait for remaining jobs
            for pid in "${pids[@]}"; do
                wait "$pid"
            done
        fi
    else
        # Sequential builds
        for char in "${characters[@]}"; do
            show_progress $((success_count + ${#failed_characters[@]} + 1)) ${#characters[@]} "Building characters"
            
            if build_character "$char"; then
                ((success_count++))
            else
                failed_characters+=("$char")
            fi
        done
    fi
    
    # Report final results
    echo
    log "Build complete: $success_count successful, ${#failed_characters[@]} failed"
    
    if [[ ${#failed_characters[@]} -gt 0 ]]; then
        warning "Failed characters: ${failed_characters[*]}"
        return 1
    fi
    
    return 0
}

# ============================================================================
# CLEANUP AND ARTIFACT MANAGEMENT
# ============================================================================

# Clean build artifacts
clean_builds() {
    log "Cleaning build artifacts..."
    
    if [[ -d "$BUILD_DIR" ]]; then
        rm -rf "$BUILD_DIR"/*
        success "Cleaned build directory"
    fi
    
    # Clean embedded character directories
    local embedded_dirs
    readarray -t embedded_dirs < <(find "$PROJECT_ROOT/cmd" -name "*-embedded" -type d)
    
    for dir in "${embedded_dirs[@]}"; do
        rm -rf "$dir"
        debug "Removed embedded directory: $(basename "$dir")"
    done
    
    success "Cleanup complete"
}

# Store artifact using artifact manager (if enabled)
store_artifact() {
    local file_path="$1"
    
    if [[ "$ENABLE_ARTIFACT_MGMT" != "true" ]]; then
        return 0
    fi
    
    if command_exists "$BUILD_DIR/artifact-manager"; then
        "$BUILD_DIR/artifact-manager" store "$file_path" >/dev/null 2>&1 || true
    fi
}

# Manage stored artifacts
manage_artifacts() {
    if [[ "$ENABLE_ARTIFACT_MGMT" != "true" ]]; then
        warning "Artifact management is disabled"
        return 1
    fi
    
    local artifact_manager="$BUILD_DIR/artifact-manager"
    
    if [[ ! -x "$artifact_manager" ]]; then
        log "Building artifact manager..."
        build_if_needed "$PROJECT_ROOT/cmd/artifact-manager/main.go" "$artifact_manager"
    fi
    
    log "Starting artifact management interface..."
    "$artifact_manager" "$@"
}

# ============================================================================
# ARGUMENT PARSING AND MAIN
# ============================================================================

# Parse command line arguments
parse_arguments() {
    COMMAND="build"
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            -p|--parallel)
                MAX_PARALLEL="$2"
                shift 2
                ;;
            -o|--output)
                BUILD_DIR="$2"
                shift 2
                ;;
            --platforms)
                PLATFORMS="$2"
                shift 2
                ;;
            --ldflags)
                LDFLAGS="$2"
                shift 2
                ;;
            --no-artifact-mgmt)
                ENABLE_ARTIFACT_MGMT="false"
                shift
                ;;
            list|build|clean|manage|platforms)
                COMMAND="$1"
                shift
                ;;
            *)
                # Assume it's a character name for build command
                if [[ "$COMMAND" == "build" ]]; then
                    SPECIFIC_CHARACTER="$1"
                fi
                shift
                ;;
        esac
    done
}

# Main entry point
main() {
    # Initialize environment
    setup_error_handling clean_builds
    init_common
    
    # Handle help flags before argument parsing to avoid subshell issues
    for arg in "$@"; do
        case "$arg" in
            -h|--help|help)
                show_usage
                exit 0
                ;;
        esac
    done
    
    # Parse arguments
    parse_arguments "$@"
    
    # Ensure build directory exists
    ensure_directories
    
    # Execute command
    case "$COMMAND" in
        list)
            list_characters
            ;;
        build)
            if [[ -n "${SPECIFIC_CHARACTER:-}" ]]; then
                build_character "$SPECIFIC_CHARACTER"
            else
                build_all_characters
            fi
            ;;
        clean)
            clean_builds
            ;;
        manage)
            manage_artifacts "${@:2}"
            ;;
        platforms)
            show_platform_info
            ;;
        help)
            show_usage
            ;;
        *)
            error "Unknown command: $command"
            show_usage
            exit 1
            ;;
    esac
}

# Run main function if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
