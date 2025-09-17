#!/bin/bash

# scripts/lib/common.sh
# Shared utility functions and constants for Desktop Companion scripts
# 
# This library provides:
# - Common logging functions with consistent formatting
# - Error handling utilities
# - Path management functions
# - Color constants for terminal output
#
# Usage: source scripts/lib/common.sh

# Prevent multiple inclusion
[[ -n "${DDS_COMMON_LIB_LOADED:-}" ]] && return 0
DDS_COMMON_LIB_LOADED=1

# ============================================================================
# CONSTANTS AND CONFIGURATION
# ============================================================================

# Colors for terminal output
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m' # No Color

# Script execution settings
set -euo pipefail

# ============================================================================
# PATH MANAGEMENT
# ============================================================================

# Get the project root directory from any script location
# Returns: Absolute path to project root
get_project_root() {
    local script_path="${BASH_SOURCE[1]}"
    local script_dir="$(cd "$(dirname "$script_path")" && pwd)"
    
    # Walk up the directory tree to find go.mod
    local current_dir="$script_dir"
    while [[ "$current_dir" != "/" ]]; do
        if [[ -f "$current_dir/go.mod" ]]; then
            echo "$current_dir"
            return 0
        fi
        current_dir="$(dirname "$current_dir")"
    done
    
    # Fallback: assume scripts are in PROJECT_ROOT/scripts
    echo "$(cd "$script_dir/.." && pwd)"
}

# Initialize common paths
readonly PROJECT_ROOT="$(get_project_root)"
readonly SCRIPT_DIR="$PROJECT_ROOT/scripts"
readonly BUILD_DIR="$PROJECT_ROOT/build"
readonly ARTIFACTS_DIR="$BUILD_DIR/artifacts"
readonly CHARACTERS_DIR="$PROJECT_ROOT/assets/characters"
readonly TEST_OUTPUT_DIR="$PROJECT_ROOT/test_output"

# Ensure required directories exist
ensure_directories() {
    mkdir -p "$BUILD_DIR" "$ARTIFACTS_DIR" "$TEST_OUTPUT_DIR"
}

# ============================================================================
# LOGGING FUNCTIONS
# ============================================================================

# Log an informational message with timestamp
# Args: message
log() {
    echo -e "${BLUE}[$(date +'%H:%M:%S')]${NC} $*" >&2
}

# Log an informational message (alternative name for consistency)
# Args: message
log_info() {
    echo -e "${GREEN}[INFO]${NC} $*" >&2
}

# Log a success message
# Args: message
success() {
    echo -e "${GREEN}✓${NC} $*" >&2
}

# Log a warning message
# Args: message
warning() {
    echo -e "${YELLOW}⚠${NC} $*" >&2
}

# Log a warning message (alternative name)
# Args: message
log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $*" >&2
}

# Log an error message
# Args: message
error() {
    echo -e "${RED}✗${NC} $*" >&2
}

# Log an error message (alternative name)
# Args: message
log_error() {
    echo -e "${RED}[ERROR]${NC} $*" >&2
}

# Log a debug message (only shown when DEBUG=1)
# Args: message
debug() {
    [[ "${DEBUG:-0}" == "1" ]] && echo -e "${BLUE}[DEBUG]${NC} $*" >&2
    return 0
}

# ============================================================================
# ERROR HANDLING
# ============================================================================

# Set up error handling with cleanup
# Args: cleanup_function (optional)
setup_error_handling() {
    local cleanup_function="${1:-}"
    
    # Handle script interruption
    trap 'handle_interrupt' INT TERM
    
    # Handle script errors only (not normal exits)
    if [[ -n "$cleanup_function" ]]; then
        trap "$cleanup_function" ERR
    fi
}

# Handle script interruption (Ctrl+C)
handle_interrupt() {
    log_error "Script interrupted by user"
    exit 130
}

# Check if a command exists
# Args: command_name
# Returns: 0 if command exists, 1 otherwise
command_exists() {
    command -v "$1" &>/dev/null
}

# Check if required commands exist
# Args: command1 [command2 ...]
# Exits: 1 if any command is missing
require_commands() {
    local missing_commands=()
    
    for cmd in "$@"; do
        if ! command_exists "$cmd"; then
            missing_commands+=("$cmd")
        fi
    done
    
    if [[ ${#missing_commands[@]} -gt 0 ]]; then
        log_error "Missing required commands: ${missing_commands[*]}"
        log_error "Please install the missing commands and try again"
        exit 1
    fi
}

# ============================================================================
# FILE OPERATIONS
# ============================================================================

# Find all character JSON files
# Returns: Array of character file paths via stdout
find_character_files() {
    find "$CHARACTERS_DIR" -mindepth 2 -maxdepth 2 -name "character.json" -type f | sort
}

# Get character name from character file path
# Args: character_file_path
# Returns: Character name
get_character_name() {
    local character_file="$1"
    basename "$(dirname "$character_file")"
}

# Check if file is older than specified time
# Args: file_path time_in_seconds
# Returns: 0 if file is older, 1 otherwise
file_older_than() {
    local file_path="$1"
    local time_seconds="$2"
    
    [[ ! -f "$file_path" ]] && return 0
    
    local file_time=$(stat -c %Y "$file_path" 2>/dev/null || stat -f %m "$file_path" 2>/dev/null)
    local current_time=$(date +%s)
    local age=$((current_time - file_time))
    
    [[ $age -gt $time_seconds ]]
}

# ============================================================================
# BUILD UTILITIES
# ============================================================================

# Build binary if needed
# Args: source_path output_path [ldflags]
build_if_needed() {
    local source_path="$1"
    local output_path="$2"
    local ldflags="${3:--s -w}"
    
    if [[ ! -f "$output_path" ]] || [[ "$source_path" -nt "$output_path" ]]; then
        log "Building $(basename "$output_path")..."
        go build -ldflags="$ldflags" -o "$output_path" "$source_path"
        success "Built $(basename "$output_path")"
    else
        debug "$(basename "$output_path") is up to date"
    fi
}

# ============================================================================
# VALIDATION UTILITIES
# ============================================================================

# Validate Go environment
validate_go_environment() {
    require_commands go
    
    local go_version
    go_version=$(go version 2>/dev/null || echo "")
    
    if [[ -z "$go_version" ]]; then
        log_error "Go is not properly installed or not in PATH"
        exit 1
    fi
    
    log "Go environment: $go_version"
    
    # Check if we're in a Go module
    if [[ ! -f "$PROJECT_ROOT/go.mod" ]]; then
        log_error "Not in a Go module directory (go.mod not found)"
        exit 1
    fi
}

# ============================================================================
# PROGRESS TRACKING
# ============================================================================

# Show progress bar for operations
# Args: current total message
show_progress() {
    local current="$1"
    local total="$2"
    local message="${3:-Processing}"
    
    local percent=$((current * 100 / total))
    local completed=$((current * 20 / total))
    
    printf "\r%s [" "$message"
    printf "%*s" $completed "" | tr ' ' '='
    printf "%*s" $((20 - completed)) "" | tr ' ' '-'
    printf "] %d%% (%d/%d)" $percent $current $total
    
    if [[ $current -eq $total ]]; then
        echo
    fi
}

# ============================================================================
# COMPATIBILITY FUNCTIONS
# ============================================================================

# Provide consistent behavior across different systems
realpath_portable() {
    local path="$1"
    
    # Use realpath if available, otherwise use Python
    if command_exists realpath; then
        realpath "$path"
    elif command_exists python3; then
        python3 -c "import os; print(os.path.realpath('$path'))"
    else
        # Fallback for systems without realpath or python
        cd "$(dirname "$path")" && pwd -P
    fi
}

# Get file modification time in seconds since epoch
get_file_mtime() {
    local file_path="$1"
    
    if [[ "$(uname)" == "Darwin" ]]; then
        # macOS
        stat -f %m "$file_path" 2>/dev/null
    else
        # Linux
        stat -c %Y "$file_path" 2>/dev/null
    fi
}

# ============================================================================
# INITIALIZATION
# ============================================================================

# Initialize common directories and environment
init_common() {
    ensure_directories
    validate_go_environment
}

# Show library information (for debugging)
show_common_info() {
    cat << EOF
Desktop Companion Scripts Common Library
======================================

Project Root: $PROJECT_ROOT
Script Dir:   $SCRIPT_DIR  
Build Dir:    $BUILD_DIR
Characters:   $CHARACTERS_DIR

Available Functions:
- Logging: log, success, warning, error, debug
- Paths: get_project_root, ensure_directories
- Files: find_character_files, get_character_name
- Build: build_if_needed
- Validation: validate_go_environment, require_commands
- Progress: show_progress

Usage: source scripts/lib/common.sh

EOF
}

# Export functions that should be available to sourcing scripts
export -f log log_info success warning log_warn error log_error debug
export -f get_project_root ensure_directories
export -f find_character_files get_character_name file_older_than
export -f build_if_needed validate_go_environment require_commands
export -f show_progress command_exists handle_interrupt
export -f realpath_portable get_file_mtime

# Initialize if not sourcing
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    show_common_info
fi
