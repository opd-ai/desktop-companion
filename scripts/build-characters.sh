#!/bin/bash

# Character-specific build automation script
# Implements Phase 1, Task 3: Create build automation scripts
# Enhanced with Phase 2, Task 3: Artifact management and retention

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
BUILD_DIR="$PROJECT_ROOT/build"
ARTIFACTS_DIR="$PROJECT_ROOT/build/artifacts"
CHARACTERS_DIR="$PROJECT_ROOT/assets/characters"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
MAX_PARALLEL=${MAX_PARALLEL:-4}
PLATFORMS=${PLATFORMS:-"$(go env GOOS)/$(go env GOARCH)"}
LDFLAGS=${LDFLAGS:-"-s -w"}
ENABLE_ARTIFACT_MGMT=${ENABLE_ARTIFACT_MGMT:-"true"}

# Print colored output
log() {
    echo -e "${BLUE}[$(date +'%H:%M:%S')]${NC} $1"
}

success() {
    echo -e "${GREEN}✓${NC} $1"
}

warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

error() {
    echo -e "${RED}✗${NC} $1" >&2
}

# Show usage information
show_usage() {
    cat << EOF
Usage: $0 [OPTIONS] [COMMAND]

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
    MAX_PARALLEL       Maximum parallel builds
    PLATFORMS          Target platforms for builds
    LDFLAGS           Linker flags for optimization
    ENABLE_ARTIFACT_MGMT   Enable automatic artifact management (true/false)
EOF
}

# Show platform matrix configuration information
show_platform_info() {
    echo
    log "Platform Matrix Configuration:"
    echo "  Current OS: $(go env GOOS)/$(go env GOARCH)"
    echo
    echo "  Supported Platforms:"
    echo "    • linux/amd64   - Linux 64-bit"
    echo "    • windows/amd64 - Windows 64-bit"  
    echo "    • darwin/amd64  - macOS 64-bit"
    echo "    • android/arm64 - Android 64-bit (requires fyne CLI tool)"
    echo "    • android/arm   - Android 32-bit (requires fyne CLI tool)"
    echo
    echo "  Cross-Compilation Limitations:"
    echo "    Due to Fyne GUI framework CGO requirements, cross-compilation"
    echo "    between different operating systems may fail. For production"
    echo "    builds, use GitHub Actions matrix builds which run on native"
    echo "    environments for each target platform."
    echo
    echo "  Android Build Requirements:"
    echo "    • fyne CLI tool: go install fyne.io/tools/cmd/fyne@latest"
    echo "    • Java 8+ installed for APK packaging"
    echo "    • Android SDK (optional for basic builds)"
    echo
    echo "  Recommended Approach:"
    echo "    • Local development: Build for current platform only"
    echo "    • Production releases: Use GitHub Actions workflow"
    echo "    • Testing: Use './scripts/build-characters.sh build CHAR --platforms \$(go env GOOS)/\$(go env GOARCH)'"
    echo
}

# List all available characters (excluding examples and templates)
list_characters() {
    log "Available characters:"
    find "$CHARACTERS_DIR" -maxdepth 1 -type d \
        -not -path "$CHARACTERS_DIR" \
        -not -path "*/examples" \
        -not -path "*/templates" \
        -exec basename {} \; | \
        sort | \
        while read -r char; do
            echo "  • $char"
        done
}

# Get list of buildable characters
get_characters() {
    find "$CHARACTERS_DIR" -maxdepth 1 -type d \
        -not -path "$CHARACTERS_DIR" \
        -not -path "*/examples" \
        -not -path "*/templates" \
        -exec basename {} \; | \
        sort
}

# Generate embedded character code
generate_character() {
    local char="$1"
    local output_dir="$PROJECT_ROOT/cmd/$char-embedded"
    
    log "Generating embedded assets for character: $char"
    
    if ! go run "$PROJECT_ROOT/scripts/embed-character.go" -character "$char" -output "$output_dir"; then
        error "Failed to generate embedded assets for $char"
        return 1
    fi
    
    success "Generated embedded assets for $char"
    return 0
}

# Validate platform compatibility and provide warnings
# Due to Fyne CGO requirements, cross-compilation has limitations
# Android builds use fyne CLI tool instead of standard go build
validate_platform() {
    local platform="$1"
    local goos="${platform%/*}"
    local goarch="${platform#*/}"
    local current_os
    current_os=$(go env GOOS)
    
    # Special handling for Android builds
    if [[ "$goos" == "android" ]]; then
        # Check if fyne CLI is available
        if ! command -v fyne >/dev/null 2>&1; then
            error "Android builds require fyne CLI tool. Install with: go install fyne.io/tools/cmd/fyne@latest"
            return 1
        fi
        log "Android platform detected - will use fyne CLI for APK generation"
        return 0
    fi
    
    # Check if we're attempting cross-compilation
    if [[ "$goos" != "$current_os" ]]; then
        case "$current_os" in
            "linux")
                case "$goos" in
                    "windows"|"darwin")
                        warning "Cross-compiling from Linux to $goos may fail due to CGO/Fyne requirements"
                        warning "For production builds, use native $goos environment or GitHub Actions matrix"
                        return 1
                        ;;
                esac
                ;;
            "darwin")
                case "$goos" in
                    "linux"|"windows")
                        warning "Cross-compiling from macOS to $goos may fail due to CGO/Fyne requirements"
                        warning "For production builds, use native $goos environment or GitHub Actions matrix"
                        return 1
                        ;;
                esac
                ;;
            "windows")
                case "$goos" in
                    "linux"|"darwin")
                        warning "Cross-compiling from Windows to $goos may fail due to CGO/Fyne requirements"
                        warning "For production builds, use native $goos environment or GitHub Actions matrix"
                        return 1
                        ;;
                esac
                ;;
        esac
    fi
    
    return 0
}

# Build character for Android platform using fyne CLI tool
# Android builds create APK files instead of native executables
build_character_android() {
    local char="$1"
    local goarch="$2"
    local source_dir="$PROJECT_ROOT/cmd/$char-embedded"
    local app_id="ai.opd.${char}"
    local app_name="${char^} Companion"  # Capitalize first letter
    local output_file="$BUILD_DIR/${char}_android_${goarch}.apk"
    
    if [[ ! -d "$source_dir" ]]; then
        error "Source directory not found: $source_dir"
        return 1
    fi
    
    log "Building Android APK for $char (${goarch})"
    
    # Create temporary directory for fyne build
    local temp_dir
    temp_dir=$(mktemp -d)
    trap "rm -rf '$temp_dir'" EXIT
    
    # Copy source to temp directory
    cp -r "$source_dir"/* "$temp_dir/"
    
    # Copy go.mod and go.sum for fyne CLI
    cp "$PROJECT_ROOT/go.mod" "$temp_dir/"
    cp "$PROJECT_ROOT/go.sum" "$temp_dir/"
    
    # Generate Android-specific app metadata (with icon)
    cat > "$temp_dir/FyneApp.toml" << EOF
[Details]
Icon = "Icon.png"
Name = "$app_name"
ID = "$app_id"
Version = "1.0.0"

[Development]
AutoInject = true
EOF

    # Create a simple icon for the APK
    local char_icon="$PROJECT_ROOT/assets/characters/$char/icon.png"
    local default_icon="$PROJECT_ROOT/assets/app/icon.png"
    
    if [[ -f "$char_icon" ]]; then
        cp "$char_icon" "$temp_dir/Icon.png"
        log "Using character-specific icon: $char_icon"
    elif [[ -f "$default_icon" ]]; then
        cp "$default_icon" "$temp_dir/Icon.png"
        log "Using default icon: $default_icon"
    else
        error "No icon found at $char_icon or $default_icon"
        return 1
    fi

    # Build APK using fyne CLI with NDK support
    cd "$temp_dir"
    
    # Check for Android NDK and set up environment if available
    if [[ -n "$ANDROID_NDK_ROOT" && -d "$ANDROID_NDK_ROOT" ]]; then
        log "Android NDK found at $ANDROID_NDK_ROOT - building optimized APK"
        
        # Set up NDK compiler for the target architecture
        if [[ "$goarch" == "arm64" ]]; then
            export CC="$ANDROID_NDK_ROOT/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android33-clang"
            export CXX="$ANDROID_NDK_ROOT/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android33-clang++"
        elif [[ "$goarch" == "arm" ]]; then
            export CC="$ANDROID_NDK_ROOT/toolchains/llvm/prebuilt/linux-x86_64/bin/armv7a-linux-androideabi33-clang"
            export CXX="$ANDROID_NDK_ROOT/toolchains/llvm/prebuilt/linux-x86_64/bin/armv7a-linux-androideabi33-clang++"
        fi
        
        # Build optimized release APK
        if fyne package \
            --target "android/$goarch" \
            --name "$app_name" \
            --app-id "$app_id" \
            --app-version "1.0.0" \
            --release; then
            log "Successfully built optimized release APK with NDK"
        else
            log "Release build failed, falling back to debug build"
            fyne package \
                --target "android/$goarch" \
                --name "$app_name" \
                --app-id "$app_id" \
                --app-version "1.0.0"
        fi
    else
        log "Android NDK not available - building basic APK without native optimizations"
        
        # Build the APK with basic configuration
        if ! fyne package \
            --target "android/$goarch" \
            --name "$app_name" \
            --app-id "$app_id" \
            --app-version "1.0.0" \
            --release 2>/dev/null; then
            # Try without --release flag if it fails
            log "Retrying Android build without release flag..."
            if ! fyne package \
                --target "android/$goarch" \
                --name "$app_name" \
                --app-id "$app_id" \
                --app-version "1.0.0"; then
                error "Failed to build Android APK for $char"
                cd - >/dev/null
                return 1
            fi
        fi
    fi
    
    # Find the generated APK file (fyne names it automatically)
    local generated_apk
    generated_apk=$(find . -name "*.apk" -type f | head -1)
    if [[ -z "$generated_apk" ]]; then
        error "No APK file generated for $char"
        cd - >/dev/null
        return 1
    fi
    
    # Validate APK before moving
    log "Validating generated APK..."
    local apk_size
    apk_size=$(stat -f%z "$generated_apk" 2>/dev/null || stat -c%s "$generated_apk")
    local min_size=$((512 * 1024))  # 512KB minimum
    local max_size=$((100 * 1024 * 1024))  # 100MB maximum
    
    if [[ $apk_size -lt $min_size ]]; then
        error "Generated APK is too small ($apk_size bytes) - likely corrupted"
        cd - >/dev/null
        return 1
    elif [[ $apk_size -gt $max_size ]]; then
        warning "Generated APK is quite large ($apk_size bytes)"
    else
        log "APK size validation passed: $apk_size bytes"
    fi
    
    # Validate APK structure
    if command -v unzip >/dev/null 2>&1; then
        local required_files=("AndroidManifest.xml" "classes.dex")
        for required in "${required_files[@]}"; do
            if ! unzip -l "$generated_apk" | grep -q "$required"; then
                error "APK missing required component: $required"
                cd - >/dev/null
                return 1
            fi
        done
        log "APK structure validation passed"
    fi

    # Move APK to desired output location
    if ! mv "$generated_apk" "$output_file"; then
        error "Failed to move APK to output location: $output_file"
        cd - >/dev/null
        return 1
    fi
    
    cd - >/dev/null
    
    success "Built Android APK for $char → $output_file"
    return 0
}

# Build character binary for a specific platform
# Android builds use fyne CLI tool for APK generation
build_character_platform() {
    local char="$1"
    local platform="$2"
    local goos="${platform%/*}"
    local goarch="${platform#*/}"
    local ext=""
    
    # Validate platform compatibility
    if ! validate_platform "$platform"; then
        warning "Skipping build for $char to $platform (validation failed)"
        return 0  # Don't fail the entire build, just skip this platform
    fi
    
    # Special handling for Android builds
    if [[ "$goos" == "android" ]]; then
        build_character_android "$char" "$goarch"
        return $?
    fi
    
    if [[ "$goos" == "windows" ]]; then
        ext=".exe"
    fi
    
    local source_dir="$PROJECT_ROOT/cmd/$char-embedded"
    local output_file="$BUILD_DIR/${char}_${goos}_${goarch}${ext}"
    
    if [[ ! -d "$source_dir" ]]; then
        error "Source directory not found: $source_dir"
        return 1
    fi
    
    log "Building $char for $platform"
    
    if ! CGO_ENABLED=1 GOOS="$goos" GOARCH="$goarch" \
        go build -ldflags="$LDFLAGS" -o "$output_file" "$source_dir/main.go"; then
        error "Failed to build $char for $platform"
        return 1
    fi
    
    success "Built $char for $platform → $output_file"
    return 0
}

# Build a single character for all platforms
build_character() {
    local char="$1"
    
    # Generate embedded assets
    if ! generate_character "$char"; then
        return 1
    fi
    
    # Build for each platform
    local platforms_array
    IFS=',' read -ra platforms_array <<< "$PLATFORMS"
    
    for platform in "${platforms_array[@]}"; do
        if ! build_character_platform "$char" "$platform"; then
            warning "Failed to build $char for $platform, continuing..."
        fi
    done
    
    # Cleanup generated source
    rm -rf "$PROJECT_ROOT/cmd/$char-embedded"
    
    return 0
}

# Build all characters
build_all_characters() {
    local characters
    mapfile -t characters < <(get_characters)
    
    log "Building ${#characters[@]} characters for platforms: $PLATFORMS"
    
    mkdir -p "$BUILD_DIR"
    
    # Build characters in parallel
    local pids=()
    local active_jobs=0
    
    for char in "${characters[@]}"; do
        # Wait if we've reached max parallel jobs
        while [[ $active_jobs -ge $MAX_PARALLEL ]]; do
            for i in "${!pids[@]}"; do
                if ! kill -0 "${pids[$i]}" 2>/dev/null; then
                    wait "${pids[$i]}"
                    unset "pids[$i]"
                    ((active_jobs--))
                fi
            done
            pids=("${pids[@]}")  # Reindex array
            sleep 0.1
        done
        
        # Start new build job
        build_character "$char" &
        pids+=($!)
        ((active_jobs++))
    done
    
    # Wait for remaining jobs
    for pid in "${pids[@]}"; do
        wait "$pid"
    done
    
    log "Build complete! Binaries available in $BUILD_DIR/"
    ls -la "$BUILD_DIR/"
}

# Clean build artifacts
clean_builds() {
    log "Cleaning build artifacts..."
    
    rm -rf "$BUILD_DIR"
    rm -rf "$PROJECT_ROOT"/cmd/*-embedded
    if [[ -d "$ARTIFACTS_DIR" ]]; then
        rm -rf "$ARTIFACTS_DIR"
    fi
    
    success "Cleaned build artifacts"
}

# Manage stored artifacts using the artifact manager tool
manage_artifacts() {
    log "Artifact Management Tools"
    
    # Check if artifact manager is built
    ARTIFACT_MANAGER="$BUILD_DIR/artifact-manager"
    if [[ ! -f "$ARTIFACT_MANAGER" ]]; then
        log "Building artifact manager..."
        mkdir -p "$BUILD_DIR"
        go build -ldflags="$LDFLAGS" -o "$ARTIFACT_MANAGER" cmd/artifact-manager/main.go
        if [[ $? -ne 0 ]]; then
            error "Failed to build artifact manager"
            return 1
        fi
        success "Built artifact manager"
    fi
    
    echo
    echo "Available artifact management commands:"
    echo "  list     - List stored artifacts"
    echo "  stats    - Show artifact statistics"
    echo "  cleanup  - Clean up expired artifacts"
    echo "  compress - Compress old artifacts"
    echo "  policies - Show retention policies"
    echo
    
    while true; do
        echo -n "Enter command (or 'quit' to exit): "
        read -r cmd
        
        case "$cmd" in
            list)
                echo -n "Character (optional): "
                read -r character
                if [[ -n "$character" ]]; then
                    "$ARTIFACT_MANAGER" -dir "$ARTIFACTS_DIR" list "$character"
                else
                    "$ARTIFACT_MANAGER" -dir "$ARTIFACTS_DIR" list
                fi
                ;;
            stats)
                "$ARTIFACT_MANAGER" -dir "$ARTIFACTS_DIR" stats
                ;;
            cleanup)
                echo "Available policies: development, production, release"
                echo -n "Enter policy name: "
                read -r policy
                if [[ -n "$policy" ]]; then
                    "$ARTIFACT_MANAGER" -dir "$ARTIFACTS_DIR" cleanup "$policy"
                fi
                ;;
            compress)
                echo "Available policies: development, production, release"
                echo -n "Enter policy name: "
                read -r policy
                if [[ -n "$policy" ]]; then
                    "$ARTIFACT_MANAGER" -dir "$ARTIFACTS_DIR" compress "$policy"
                fi
                ;;
            policies)
                "$ARTIFACT_MANAGER" -dir "$ARTIFACTS_DIR" policies
                ;;
            quit|exit|q)
                break
                ;;
            help|h)
                echo "Available commands: list, stats, cleanup, compress, policies, quit"
                ;;
            *)
                echo "Unknown command: $cmd (type 'help' for available commands)"
                ;;
        esac
        echo
    done
}

# Store artifact using artifact manager (if enabled)
store_artifact() {
    local binary_path="$1"
    local character="$2" 
    local platform="$3"
    local arch="$4"
    
    if [[ "$ENABLE_ARTIFACT_MGMT" != "true" ]]; then
        return 0
    fi
    
    # Build artifact manager if needed
    ARTIFACT_MANAGER="$BUILD_DIR/artifact-manager"
    if [[ ! -f "$ARTIFACT_MANAGER" ]]; then
        go build -ldflags="$LDFLAGS" -o "$ARTIFACT_MANAGER" cmd/artifact-manager/main.go >/dev/null 2>&1
    fi
    
    # Store the artifact
    if [[ -f "$ARTIFACT_MANAGER" && -f "$binary_path" ]]; then
        "$ARTIFACT_MANAGER" -dir "$ARTIFACTS_DIR" store "$character" "$platform" "$arch" "$binary_path" >/dev/null 2>&1
    fi
}

# Validate build environment
validate_environment() {
    # Check if we're in the right directory
    if [[ ! -f "$PROJECT_ROOT/go.mod" ]]; then
        error "Must be run from DDS project root (go.mod not found)"
        exit 1
    fi
    
    # Check if embed script exists
    if [[ ! -f "$PROJECT_ROOT/scripts/embed-character.go" ]]; then
        error "Embed script not found: scripts/embed-character.go"
        exit 1
    fi
    
    # Check if characters directory exists
    if [[ ! -d "$CHARACTERS_DIR" ]]; then
        error "Characters directory not found: $CHARACTERS_DIR"
        exit 1
    fi
    
    # Check Go installation
    if ! command -v go >/dev/null 2>&1; then
        error "Go is not installed or not in PATH"
        exit 1
    fi
    
    success "Build environment validated"
}

# Parse command line arguments
parse_args() {
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
            -h|--help|help)
                show_usage
                exit 0
                ;;
            list)
                list_characters
                exit 0
                ;;
            build)
                if [[ -n "$2" && "$2" != -* ]]; then
                    # Build specific character
                    validate_environment
                    mkdir -p "$BUILD_DIR"
                    build_character "$2"
                    exit $?
                else
                    # Build all characters
                    validate_environment
                    build_all_characters
                    exit $?
                fi
                ;;
            clean)
                clean_builds
                exit 0
                ;;
            manage)
                manage_artifacts
                exit 0
                ;;
            platforms)
                show_platform_info
                exit 0
                ;;
            *)
                error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done
    
    # Default action if no command provided
    show_usage
}

# Main entry point
main() {
    cd "$PROJECT_ROOT"
    
    if [[ $# -eq 0 ]]; then
        show_usage
        exit 0
    fi
    
    parse_args "$@"
}

# Run main function with all arguments
main "$@"
