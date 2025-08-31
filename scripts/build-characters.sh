#!/bin/bash

# Character-specific build automation script
# Implements Phase 1, Task 3: Create build automation scripts

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
BUILD_DIR="$PROJECT_ROOT/build"
CHARACTERS_DIR="$PROJECT_ROOT/assets/characters"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
MAX_PARALLEL=${MAX_PARALLEL:-4}
PLATFORMS=${PLATFORMS:-"linux/amd64"}
LDFLAGS=${LDFLAGS:-"-s -w"}

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
    help               Show this help message

OPTIONS:
    -p, --parallel N    Set maximum parallel builds (default: $MAX_PARALLEL)
    -o, --output DIR    Set output directory (default: $BUILD_DIR)
    --platforms LIST    Comma-separated list of platforms (default: $PLATFORMS)
    --ldflags FLAGS     Linker flags for builds (default: "$LDFLAGS")

EXAMPLES:
    $0 list                           # List available characters
    $0 build                          # Build all characters
    $0 build default                  # Build only default character
    $0 build --parallel 8             # Build with 8 parallel workers
    $0 clean                          # Clean build directory

ENVIRONMENT VARIABLES:
    MAX_PARALLEL       Maximum parallel builds
    PLATFORMS          Target platforms for builds
    LDFLAGS           Linker flags for optimization
EOF
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

# Build character binary for a specific platform
build_character_platform() {
    local char="$1"
    local platform="$2"
    local goos="${platform%/*}"
    local goarch="${platform#*/}"
    local ext=""
    
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
    
    success "Cleaned build artifacts"
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
