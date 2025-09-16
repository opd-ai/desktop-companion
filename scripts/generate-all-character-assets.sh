#!/bin/bash

# DEPRECATED: Legacy wrapper for generate-all-character-assets.sh
# This script is maintained for backward compatibility.
# New usage: ./scripts/dds-scripts.sh asset-generation generate
# Direct usage: ./scripts/asset-generation/generate-character-assets.sh

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Forward all arguments to the new refactored script
exec "$SCRIPT_DIR/asset-generation/generate-character-assets.sh" "$@"

# Default settings
PARALLEL_JOBS=2
VERBOSE=false
DRY_RUN=false
BACKUP_ASSETS=true
VALIDATE_ASSETS=true
FORCE_REBUILD=false

# Styling options
DEFAULT_STYLE="anime"
DEFAULT_MODEL="sd15"

# Create required directories
mkdir -p "$(dirname "$LOG_FILE")"
mkdir -p "$BUILD_DIR"

# Logging functions
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*" | tee -a "$LOG_FILE"
}

log_error() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] ERROR: $*" | tee -a "$LOG_FILE" >&2
}

log_success() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] SUCCESS: $*" | tee -a "$LOG_FILE"
}

# Usage function
usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Generate GIF assets for all character JSON files in assets/characters/

OPTIONS:
    -h, --help          Show this help message
    -v, --verbose       Enable verbose output
    -n, --dry-run       Show what would be done without executing
    -f, --force         Force rebuild of gif-generator binary
    -j, --jobs N        Number of parallel jobs (default: $PARALLEL_JOBS)
    --no-backup         Don't backup existing assets
    --no-validate       Don't validate generated assets
    --style STYLE       Character style (default: $DEFAULT_STYLE)
    --model MODEL       AI model to use (default: $DEFAULT_MODEL)

EXAMPLES:
    $0                                    # Generate assets for all characters with defaults
    $0 --verbose --jobs 4                # Verbose output with 4 parallel jobs
    $0 --dry-run                         # Preview what would be generated
    $0 --style "realistic" --model "sdxl" # Use realistic style with SDXL model
    $0 --no-backup --no-validate         # Skip backup and validation steps

OUTPUT:
    Generated assets will be placed in each character's directory.
    Log file: $LOG_FILE
EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            usage
            exit 0
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -n|--dry-run)
            DRY_RUN=true
            shift
            ;;
        -f|--force)
            FORCE_REBUILD=true
            shift
            ;;
        -j|--jobs)
            PARALLEL_JOBS="$2"
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
        --style)
            DEFAULT_STYLE="$2"
            shift 2
            ;;
        --model)
            DEFAULT_MODEL="$2"
            shift 2
            ;;
        *)
            log_error "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

# Build gif-generator if needed
build_gif_generator() {
    log "Checking gif-generator binary..."
    
    if [[ "$FORCE_REBUILD" == "true" ]] || [[ ! -f "$GIF_GENERATOR_BINARY" ]]; then
        log "Building gif-generator..."
        cd "$PROJECT_ROOT"
        
        if ! go build -ldflags="-s -w" -o "$GIF_GENERATOR_BINARY" "$GIF_GENERATOR_CMD/main.go"; then
            log_error "Failed to build gif-generator"
            return 1
        fi
        
        log_success "gif-generator built successfully"
    else
        log "gif-generator binary already exists"
    fi
    
    if [[ ! -x "$GIF_GENERATOR_BINARY" ]]; then
        log_error "gif-generator binary is not executable"
        return 1
    fi
    
    return 0
}

# Find all character JSON files
find_character_files() {
    log "Scanning for character JSON files..."
    
    local character_files=()
    local seen_files=()
    
    # Find all character.json files exactly 2 levels deep in assets/characters/
    while IFS= read -r file; do
        # Skip if not a regular file
        [[ -f "$file" ]] || continue
        
        # Skip template and example directories
        if [[ "$file" =~ /templates/ ]] || [[ "$file" =~ /examples/ ]]; then
            [[ "$VERBOSE" == "true" ]] && log "Skipping template/example: $file"
            continue
        fi
        
        # Get the character directory name
        local character_dir
        character_dir="$(dirname "$file")"
        local character_name
        character_name="$(basename "$character_dir")"
        
        # Skip empty or invalid names
        if [[ -z "$character_name" ]] || [[ "$character_name" == "." ]] || [[ "$character_name" == ".." ]]; then
            [[ "$VERBOSE" == "true" ]] && log "Skipping invalid character name: $character_name"
            continue
        fi
        
        # Check for duplicates (shouldn't happen but let's be safe)
        if [[ " ${seen_files[*]} " =~ " ${character_name} " ]]; then
            [[ "$VERBOSE" == "true" ]] && log "Skipping duplicate: $file"
            continue
        fi
        
        character_files+=("$file")
        seen_files+=("$character_name")
    done < <(find "$CHARACTERS_DIR" -mindepth 2 -maxdepth 2 -name "character.json" -type f | sort)
    
    log "Found ${#character_files[@]} character files to process"
    
    if [[ "$VERBOSE" == "true" ]]; then
        for file in "${character_files[@]}"; do
            log "  - $file"
        done
    fi
    
    printf '%s\n' "${character_files[@]}"
}

# Generate assets for a single character
generate_character_assets() {
    local character_file="$1"
    local character_dir
    character_dir="$(dirname "$character_file")"
    local character_name
    character_name="$(basename "$character_dir")"
    
    # Use local log prefix for this character to avoid mixing output
    local local_log_file="${LOG_FILE}.${character_name}"
    
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] Processing character: $character_name" >> "$local_log_file"
    
    # Prepare gif-generator command
    local cmd_args=(
        "character"
        "--file" "$character_file"
        "--style" "$DEFAULT_STYLE"
        "--model" "$DEFAULT_MODEL"
        "--output" "$character_dir"
    )
    
    # Add optional flags
    if [[ "$BACKUP_ASSETS" == "true" ]]; then
        cmd_args+=("--backup")
    fi
    
    if [[ "$VALIDATE_ASSETS" == "true" ]]; then
        cmd_args+=("--validate")
    fi
    
    if [[ "$VERBOSE" == "true" ]]; then
        cmd_args+=("--verbose")
    fi
    
    if [[ "$DRY_RUN" == "true" ]]; then
        cmd_args+=("--dry-run")
    fi
    
    # Execute the command
    local full_cmd="$GIF_GENERATOR_BINARY ${cmd_args[*]}"
    
    if [[ "$VERBOSE" == "true" ]] || [[ "$DRY_RUN" == "true" ]]; then
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] Command: $full_cmd" >> "$local_log_file"
    fi
    
    if [[ "$DRY_RUN" == "false" ]]; then
        if "$GIF_GENERATOR_BINARY" "${cmd_args[@]}" >> "$local_log_file" 2>&1; then
            echo "[$(date '+%Y-%m-%d %H:%M:%S')] SUCCESS: Generated assets for $character_name" >> "$LOG_FILE"
            # Also append local log to main log
            cat "$local_log_file" >> "$LOG_FILE"
            rm -f "$local_log_file"
            return 0
        else
            echo "[$(date '+%Y-%m-%d %H:%M:%S')] ERROR: Failed to generate assets for $character_name" >> "$LOG_FILE"
            # Also append local log to main log for debugging
            cat "$local_log_file" >> "$LOG_FILE"
            rm -f "$local_log_file"
            return 1
        fi
    else
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] DRY RUN: Would execute: $full_cmd" >> "$local_log_file"
        # Append to main log immediately for dry run
        cat "$local_log_file" >> "$LOG_FILE"
        rm -f "$local_log_file"
        return 0
    fi
}

# Process characters in parallel
process_characters_parallel() {
    local character_files=("$@")
    local total_count=${#character_files[@]}
    local success_count=0
    local failure_count=0
    
    log "Starting parallel processing with $PARALLEL_JOBS jobs"
    log "Processing $total_count characters..."
    
    # Use GNU parallel if available, otherwise fall back to background jobs
    if command -v parallel >/dev/null 2>&1; then
        log "Using GNU parallel for job control"
        
        export -f generate_character_assets log log_error log_success
        export GIF_GENERATOR_BINARY DEFAULT_STYLE DEFAULT_MODEL BACKUP_ASSETS VALIDATE_ASSETS VERBOSE DRY_RUN LOG_FILE
        
        printf '%s\n' "${character_files[@]}" | \
            parallel -j "$PARALLEL_JOBS" --bar generate_character_assets {}
        
        # Note: parallel doesn't easily return individual exit codes, so we check log for success/failure
        success_count=$(grep -c "SUCCESS: Generated assets for" "$LOG_FILE" || echo "0")
        failure_count=$((total_count - success_count))
        
    else
        log "Using background jobs for parallel processing"
        
        local pids=()
        local job_count=0
        
        for character_file in "${character_files[@]}"; do
            # Wait if we've reached the parallel limit
            while [[ ${#pids[@]} -ge $PARALLEL_JOBS ]]; do
                for i in "${!pids[@]}"; do
                    if ! kill -0 "${pids[$i]}" 2>/dev/null; then
                        wait "${pids[$i]}"
                        if [[ $? -eq 0 ]]; then
                            ((success_count++))
                        else
                            ((failure_count++))
                        fi
                        unset "pids[$i]"
                    fi
                done
                pids=("${pids[@]}")  # Reindex array
                sleep 0.1
            done
            
            # Start new job
            generate_character_assets "$character_file" &
            pids+=($!)
            ((job_count++))
            
            [[ "$VERBOSE" == "true" ]] && log "Started job $job_count/$total_count (PID: $!)"
        done
        
        # Wait for remaining jobs
        for pid in "${pids[@]}"; do
            if wait "$pid"; then
                ((success_count++))
            else
                ((failure_count++))
            fi
        done
    fi
    
    log "Processing complete: $success_count successful, $failure_count failed"
    return $failure_count
}

# Main execution
main() {
    log "Starting character asset generation"
    log "Project root: $PROJECT_ROOT"
    log "Characters directory: $CHARACTERS_DIR"
    log "Log file: $LOG_FILE"
    log "Configuration: style=$DEFAULT_STYLE, model=$DEFAULT_MODEL, jobs=$PARALLEL_JOBS"
    
    # Check dependencies
    cd "$PROJECT_ROOT"
    
    if [[ ! -d "$CHARACTERS_DIR" ]]; then
        log_error "Characters directory not found: $CHARACTERS_DIR"
        exit 1
    fi
    
    # Build gif-generator
    if ! build_gif_generator; then
        log_error "Failed to build gif-generator"
        exit 1
    fi
    
    # Find character files
    local character_files
    readarray -t character_files < <(find_character_files)
    
    if [[ ${#character_files[@]} -eq 0 ]]; then
        log_error "No character files found"
        exit 1
    fi
    
    # Process characters
    local start_time
    start_time="$(date +%s)"
    
    if ! process_characters_parallel "${character_files[@]}"; then
        local end_time
        end_time="$(date +%s)"
        local duration=$((end_time - start_time))
        
        log_error "Some character processing failed (duration: ${duration}s)"
        log "Check the log file for details: $LOG_FILE"
        exit 1
    fi
    
    local end_time
    end_time="$(date +%s)"
    local duration=$((end_time - start_time))
    
    log_success "All character assets generated successfully (duration: ${duration}s)"
    log "Generated assets for ${#character_files[@]} characters"
    log "Log file: $LOG_FILE"
}

# Check for required commands
check_dependencies() {
    local missing_deps=()
    
    if ! command -v go >/dev/null 2>&1; then
        missing_deps+=("go")
    fi
    
    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        log_error "Missing required dependencies: ${missing_deps[*]}"
        log "Please install the missing dependencies and try again"
        exit 1
    fi
}

# Script entry point
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    check_dependencies
    main "$@"
fi
