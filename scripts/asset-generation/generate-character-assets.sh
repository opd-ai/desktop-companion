#!/bin/bash

# scripts/asset-generation/generate-character-assets.sh
# Character asset generation pipeline script
#
# Generates GIF assets for all character JSON files using the gif-generator tool.
# This script replaces placeholder assets with properly generated ones for each
# character archetype.
#
# Usage: ./scripts/asset-generation/generate-character-assets.sh [OPTIONS] [COMMAND]
#
# Dependencies:
# - Go 1.21+
# - gif-generator tool
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
# ASSET GENERATION CONFIGURATION
# ============================================================================

# Generator configuration
GIF_GENERATOR_CMD="$PROJECT_ROOT/cmd/gif-generator"
GIF_GENERATOR_BINARY="$BUILD_DIR/gif-generator"

# Generation settings from shared config or defaults
PARALLEL_JOBS="${DDS_MAX_PARALLEL:-2}"
BACKUP_ASSETS="${DDS_BACKUP_ASSETS:-true}"
VALIDATE_ASSETS="${DDS_VALIDATE_ASSETS:-true}"
FORCE_REBUILD="${DDS_FORCE_REBUILD:-false}"

# Styling options
DEFAULT_STYLE="${DDS_DEFAULT_STYLE:-anime}"
DEFAULT_MODEL="${DDS_DEFAULT_MODEL:-sd15}"

# Asset generation modes
COMPREHENSIVE_MODE=true
QUICK_MODE=false

# ============================================================================
# HELP AND USAGE
# ============================================================================

show_usage() {
    cat << EOF
Usage: $0 [OPTIONS] [COMMAND]

Generate GIF assets for all character JSON files in assets/characters/.

COMMANDS:
    generate [CHARACTER]   Generate assets for specific character or all characters (default)
    simple                 Use simplified generation (faster, basic assets)
    validate              Validate existing assets without regenerating
    rebuild               Force rebuild all assets (ignore existing)
    help                  Show this help message

OPTIONS:
    -j, --parallel N      Number of parallel jobs (default: $PARALLEL_JOBS)
    -s, --style STYLE     Art style for generation (default: $DEFAULT_STYLE)
    -m, --model MODEL     Model to use for generation (default: $DEFAULT_MODEL)
    --no-backup          Don't backup existing assets
    --no-validate        Skip asset validation after generation
    --force              Force rebuild existing assets
    --quick              Enable quick mode (faster generation)
    -v, --verbose        Enable verbose output
    --dry-run           Show what would be generated without creating files

EXAMPLES:
    $0                   # Generate assets for all characters
    $0 generate default  # Generate assets for default character only
    $0 simple            # Use simplified generation for all characters
    $0 --style realistic # Generate with realistic art style
    $0 --parallel 4      # Use 4 parallel generation processes
    $0 --force rebuild   # Force rebuild all assets

ASSET TYPES:
    - idle.gif           # Character idle animation
    - talking.gif        # Character talking animation
    - (additional character-specific animations based on archetype)

OUTPUT:
    Generated assets: assets/characters/[CHARACTER_NAME]/
    Generation log: $TEST_OUTPUT_DIR/asset-generation-*.log
    Backup location: $BUILD_DIR/asset-backups/ (if enabled)

EOF
}

# ============================================================================
# ASSET GENERATION FUNCTIONS
# ============================================================================

# Build gif-generator if needed
build_gif_generator() {
    if [[ -f "$GIF_GENERATOR_BINARY" && "$FORCE_REBUILD" != "true" ]]; then
        log "gif-generator already exists: $GIF_GENERATOR_BINARY"
        return 0
    fi
    
    log "Building gif-generator tool..."
    
    if [[ ! -d "$GIF_GENERATOR_CMD" ]]; then
        error "gif-generator source not found: $GIF_GENERATOR_CMD"
        return 1
    fi
    
    if (cd "$GIF_GENERATOR_CMD" && go build -o "$GIF_GENERATOR_BINARY" .); then
        success "gif-generator built successfully"
        return 0
    else
        error "Failed to build gif-generator"
        return 1
    fi
}

# Find all character JSON files
find_character_files() {
    local character_files=()
    
    while IFS= read -r file; do
        character_files+=("$file")
    done < <(find "$CHARACTERS_DIR" -mindepth 2 -maxdepth 2 -name "character.json" -type f | sort)
    
    if [[ ${#character_files[@]} -eq 0 ]]; then
        error "No character files found in $CHARACTERS_DIR"
        return 1
    fi
    
    if [[ "${DDS_VERBOSE:-false}" == "true" ]]; then
        log "Found ${#character_files[@]} character files:"
        for file in "${character_files[@]}"; do
            local char_dir=$(dirname "$file")
            local char_name=$(basename "$char_dir")
            log "  - $char_name: $file"
        done
    else
        log "Found ${#character_files[@]} character files to process"
    fi
    
    printf '%s\n' "${character_files[@]}"
}

# Backup existing assets
backup_existing_assets() {
    if [[ "$BACKUP_ASSETS" != "true" ]]; then
        return 0
    fi
    
    local character_dir="$1"
    local char_name=$(basename "$character_dir")
    local backup_dir="$BUILD_DIR/asset-backups/$char_name-$(date +%Y%m%d-%H%M%S)"
    
    # Check if there are any existing assets to backup
    local existing_assets=()
    for ext in gif png jpg jpeg; do
        while IFS= read -r -d '' file; do
            existing_assets+=("$file")
        done < <(find "$character_dir" -maxdepth 1 -name "*.$ext" -print0 2>/dev/null)
    done
    
    if [[ ${#existing_assets[@]} -eq 0 ]]; then
        return 0
    fi
    
    log "Backing up existing assets for $char_name..."
    mkdir -p "$backup_dir"
    
    for asset in "${existing_assets[@]}"; do
        cp "$asset" "$backup_dir/" || warning "Failed to backup $(basename "$asset")"
    done
    
    log "Assets backed up to: $backup_dir"
}

# Generate assets for a single character
generate_character_assets() {
    local character_file="$1"
    local character_dir=$(dirname "$character_file")
    local char_name=$(basename "$character_dir")
    
    log "Generating assets for character: $char_name"
    
    # Backup existing assets if enabled
    backup_existing_assets "$character_dir"
    
    # Prepare generation command
    local generation_cmd=(
        "$GIF_GENERATOR_BINARY"
    )
    
    # Add global flags first (before command)
    if [[ "${DDS_VERBOSE:-false}" == "true" ]]; then
        generation_cmd+=("--verbose")
    fi
    
    if [[ "${DDS_DRY_RUN:-false}" == "true" ]]; then
        generation_cmd+=("--dry-run")
    fi
    
    # Add the command
    generation_cmd+=("character")
    
    # Add command-specific options
    generation_cmd+=(
        "--file" "$character_file"
        "--output" "$character_dir"
        "--style" "$DEFAULT_STYLE"
    )
    
    if [[ "$QUICK_MODE" == "true" ]]; then
        generation_cmd+=("--quick")
    fi
    
    # Execute generation
    local generation_log="$TEST_OUTPUT_DIR/generation-${char_name}-$(date +%H%M%S).log"
    
    if [[ "${DDS_DRY_RUN:-false}" == "true" ]]; then
        log "DRY RUN: Would execute: ${generation_cmd[*]}"
        return 0
    fi
    
    if "${generation_cmd[@]}" > "$generation_log" 2>&1; then
        success "✅ Assets generated for $char_name"
        
        # Validate generated assets if enabled
        if [[ "$VALIDATE_ASSETS" == "true" ]]; then
            validate_character_assets "$character_dir"
        fi
        
        return 0
    else
        error "❌ Asset generation failed for $char_name (see: $generation_log)"
        return 1
    fi
}

# Validate character assets
validate_character_assets() {
    local character_dir="$1"
    local char_name=$(basename "$character_dir")
    
    # Check for required basic assets
    local required_assets=("idle.gif" "talking.gif")
    local missing_assets=()
    
    for asset in "${required_assets[@]}"; do
        if [[ ! -f "$character_dir/$asset" ]]; then
            missing_assets+=("$asset")
        fi
    done
    
    if [[ ${#missing_assets[@]} -gt 0 ]]; then
        warning "Missing required assets for $char_name: ${missing_assets[*]}"
        return 1
    fi
    
    # Validate asset file sizes (should not be empty)
    for asset in "$character_dir"/*.gif; do
        [[ -f "$asset" ]] || continue
        
        local size=$(stat -f%z "$asset" 2>/dev/null || stat -c%s "$asset" 2>/dev/null || echo "0")
        if [[ $size -eq 0 ]]; then
            warning "Empty asset file: $(basename "$asset")"
        fi
    done
    
    log "✓ Asset validation completed for $char_name"
    return 0
}

# Process all characters with parallel execution
process_all_characters() {
    local character_files
    mapfile -t character_files < <(find_character_files)
    
    if [[ ${#character_files[@]} -eq 0 ]]; then
        return 1
    fi
    
    local success_count=0
    local failure_count=0
    
    # Process characters in parallel batches
    local batch_size=$PARALLEL_JOBS
    local total_chars=${#character_files[@]}
    
    for ((i=0; i<total_chars; i+=batch_size)); do
        local batch=("${character_files[@]:i:batch_size}")
        local pids=()
        
        # Start parallel generation for current batch
        for character_file in "${batch[@]}"; do
            generate_character_assets "$character_file" &
            pids+=($!)
        done
        
        # Wait for batch to complete
        for pid in "${pids[@]}"; do
            if wait "$pid"; then
                ((success_count++))
            else
                ((failure_count++))
            fi
        done
        
        log "Batch $((i/batch_size + 1)) completed: $((i + ${#batch[@]}))/$total_chars characters processed"
    done
    
    # Final summary
    log "Asset generation complete: $success_count successful, $failure_count failed"
    
    if [[ $failure_count -gt 0 ]]; then
        return 1
    fi
    
    return 0
}

# Simple generation mode (faster, basic assets only)
simple_generation_mode() {
    log "Using simplified asset generation..."
    
    QUICK_MODE=true
    DEFAULT_STYLE="simple"
    VALIDATE_ASSETS=false
    
    process_all_characters
}

# ============================================================================
# MAIN EXECUTION
# ============================================================================

# Parse command line arguments
COMMAND="generate"
TARGET_CHARACTER=""

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help|help)
            show_usage
            exit 0
            ;;
        -j|--parallel)
            PARALLEL_JOBS="$2"
            shift 2
            ;;
        -s|--style)
            DEFAULT_STYLE="$2"
            shift 2
            ;;
        -m|--model)
            DEFAULT_MODEL="$2"
            shift 2
            ;;
        --no-backup)
            BACKUP_ASSETS=false
            shift
            ;;
        --no-validate)
            VALIDATE_ASSETS=false
            shift
            ;;
        --force)
            FORCE_REBUILD=true
            shift
            ;;
        --quick)
            QUICK_MODE=true
            shift
            ;;
        -v|--verbose)
            DDS_VERBOSE=true
            shift
            ;;
        --dry-run)
            DDS_DRY_RUN=true
            shift
            ;;
        generate|simple|validate|rebuild)
            COMMAND="$1"
            shift
            ;;
        -*)
            error "Unknown option: $1"
            show_usage
            exit 1
            ;;
        *)
            if [[ "$COMMAND" == "generate" && -z "$TARGET_CHARACTER" ]]; then
                TARGET_CHARACTER="$1"
            else
                error "Unexpected argument: $1"
                show_usage
                exit 1
            fi
            shift
            ;;
    esac
done

# Create required directories
mkdir -p "$BUILD_DIR" "$TEST_OUTPUT_DIR"

# Build gif-generator tool
build_gif_generator || exit 1

# Execute command
case $COMMAND in
    generate)
        log "Starting character asset generation..."
        
        if [[ -n "$TARGET_CHARACTER" ]]; then
            character_file="$CHARACTERS_DIR/$TARGET_CHARACTER/character.json"
            if [[ -f "$character_file" ]]; then
                generate_character_assets "$character_file"
            else
                error "Character file not found: $character_file"
                exit 1
            fi
        else
            process_all_characters
        fi
        ;;
    simple)
        log "Starting simplified asset generation..."
        simple_generation_mode
        ;;
    validate)
        log "Validating existing character assets..."
        
        local validation_failures=0
        while IFS= read -r character_file; do
            local character_dir=$(dirname "$character_file")
            if ! validate_character_assets "$character_dir"; then
                ((validation_failures++))
            fi
        done < <(find_character_files)
        
        if [[ $validation_failures -gt 0 ]]; then
            error "$validation_failures characters have asset validation issues"
            exit 1
        else
            success "All character assets validated successfully"
        fi
        ;;
    rebuild)
        log "Force rebuilding all character assets..."
        FORCE_REBUILD=true
        process_all_characters
        ;;
    *)
        error "Unknown command: $COMMAND"
        show_usage
        exit 1
        ;;
esac

success "Character asset generation completed successfully!"
exit 0
