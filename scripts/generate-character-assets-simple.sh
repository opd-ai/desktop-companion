#!/bin/bash

# DEPRECATED: Legacy wrapper for generate-character-assets-simple.sh
# This script is maintained for backward compatibility.
# New usage: ./scripts/dds-scripts.sh asset-generation simple
# Direct usage: ./scripts/asset-generation/generate-character-assets.sh simple

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Forward all arguments to the new refactored script
exec "$SCRIPT_DIR/asset-generation/generate-character-assets.sh" simple "$@"

# Default settings
STYLE="anime"
MODEL="sd15"
DRY_RUN=false
VERBOSE=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --verbose)
            VERBOSE=true
            shift
            ;;
        --style)
            STYLE="$2"
            shift 2
            ;;
        --model)
            MODEL="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 [--dry-run] [--verbose] [--style STYLE] [--model MODEL]"
            echo "Generate GIF assets for all character JSON files"
            echo ""
            echo "Options:"
            echo "  --dry-run     Show what would be done without executing"
            echo "  --verbose     Show detailed output"
            echo "  --style       Character style (default: anime)"
            echo "  --model       AI model (default: sd15)"
            echo "  --help        Show this help"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Logging functions
log() {
    echo "[$(date '+%H:%M:%S')] $*"
}

# Build gif-generator if needed
if [[ ! -f "$GIF_GENERATOR_BINARY" ]]; then
    log "Building gif-generator..."
    cd "$PROJECT_ROOT"
    mkdir -p "$BUILD_DIR"
    if ! go build -ldflags="-s -w" -o "$GIF_GENERATOR_BINARY" cmd/gif-generator/main.go; then
        echo "ERROR: Failed to build gif-generator"
        exit 1
    fi
    log "gif-generator built successfully"
fi

# Find all character files
character_files=()
while IFS= read -r file; do
    character_files+=("$file")
done < <(find "$CHARACTERS_DIR" -mindepth 2 -maxdepth 2 -name "character.json" -type f | sort)

log "Found ${#character_files[@]} character files to process"

if [[ "$VERBOSE" == "true" ]]; then
    for file in "${character_files[@]}"; do
        char_name="$(basename "$(dirname "$file")")"
        log "  - $char_name ($file)"
    done
fi

# Process each character
success_count=0
failure_count=0

for character_file in "${character_files[@]}"; do
    character_dir="$(dirname "$character_file")"
    character_name="$(basename "$character_dir")"
    
    log "Processing: $character_name"
    
    # Build command
    cmd_args=(
        "character"
        "--file" "$character_file"
        "--style" "$STYLE"
        "--model" "$MODEL"
        "--output" "$character_dir"
        "--backup"
        "--validate"
    )
    
    if [[ "$VERBOSE" == "true" ]]; then
        cmd_args+=("--verbose")
    fi
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log "  DRY RUN: $GIF_GENERATOR_BINARY ${cmd_args[*]}"
        ((success_count++))
    else
        if [[ "$VERBOSE" == "true" ]]; then
            log "  Command: $GIF_GENERATOR_BINARY ${cmd_args[*]}"
        fi
        
        if "$GIF_GENERATOR_BINARY" "${cmd_args[@]}"; then
            log "  ✓ SUCCESS: $character_name"
            ((success_count++))
        else
            log "  ✗ FAILED: $character_name"
            ((failure_count++))
        fi
    fi
done

# Summary
log "Processing complete: $success_count successful, $failure_count failed"

if [[ $failure_count -gt 0 ]]; then
    exit 1
fi
