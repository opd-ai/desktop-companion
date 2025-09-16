#!/bin/bash

# validate-characters.sh
# Quick validation script to check all character JSON files for common issues

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
CHARACTERS_DIR="$PROJECT_ROOT/assets/characters"
BUILD_DIR="$PROJECT_ROOT/build"
GIF_GENERATOR_BINARY="$BUILD_DIR/gif-generator"

log() {
    echo "[$(date '+%H:%M:%S')] $*"
}

# Build gif-generator if needed
if [[ ! -f "$GIF_GENERATOR_BINARY" ]]; then
    log "Building gif-generator for validation..."
    cd "$PROJECT_ROOT"
    mkdir -p "$BUILD_DIR"
    if ! go build -ldflags="-s -w" -o "$GIF_GENERATOR_BINARY" cmd/gif-generator/main.go; then
        echo "ERROR: Failed to build gif-generator"
        exit 1
    fi
fi

# Find all character files
character_files=()
while IFS= read -r file; do
    character_files+=("$file")
done < <(find "$CHARACTERS_DIR" -mindepth 2 -maxdepth 2 -name "character.json" -type f | sort)

log "Validating ${#character_files[@]} character files..."

success_count=0
failure_count=0
failed_characters=()

for character_file in "${character_files[@]}"; do
    character_name="$(basename "$(dirname "$character_file")")"
    character_dir="$(dirname "$character_file")"
    
    if "$GIF_GENERATOR_BINARY" validate --path "$character_dir" >/dev/null 2>&1; then
        echo "✓ $character_name"
        ((success_count++))
    else
        echo "✗ $character_name"
        failed_characters+=("$character_name")
        ((failure_count++))
    fi
done

echo
log "Validation complete: $success_count passed, $failure_count failed"

if [[ $failure_count -gt 0 ]]; then
    echo
    echo "Failed characters:"
    for char in "${failed_characters[@]}"; do
        echo "  - $char"
    done
    echo
    echo "To see detailed errors for a specific character:"
    echo "  ./build/gif-generator validate --path assets/characters/CHAR_NAME"
    exit 1
fi

echo
log "All characters passed validation! ✨"
