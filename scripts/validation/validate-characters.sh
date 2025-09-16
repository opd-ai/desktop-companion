#!/bin/bash

# scripts/validation/validate-characters.sh
# Character JSON validation script
#
# This script validates all character JSON files for syntax errors,
# schema compliance, and common configuration issues.
#
# Usage: ./scripts/validation/validate-characters.sh [OPTIONS]
#
# Dependencies:
# - Go 1.21+
# - gif-generator tool (for validation)
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
# VALIDATION CONFIGURATION
# ============================================================================

# Validation tools
GIF_GENERATOR_BINARY="$BUILD_DIR/gif-generator"

# Validation options
VERBOSE="${DDS_VERBOSE}"
DETAILED_OUTPUT=false
STOP_ON_FIRST_ERROR=false

# ============================================================================
# HELP AND USAGE
# ============================================================================

show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Validate all character JSON files for syntax and schema compliance.

OPTIONS:
    -v, --verbose        Show detailed validation output
    -d, --detailed       Show detailed error messages for failed characters
    -s, --stop-on-error  Stop validation on first error
    --build-validator    Build gif-generator tool if needed
    --help              Show this help message

EXAMPLES:
    $0                        # Quick validation of all characters
    $0 --verbose --detailed   # Verbose validation with detailed errors
    $0 --stop-on-error        # Stop on first validation failure

The validation checks for:
- JSON syntax errors
- Character schema compliance
- Animation file references
- Required field presence
- Valid category and trigger values

See: docs/CHARACTER_VALIDATION_REPORT.md for detailed validation rules.
EOF
}

# ============================================================================
# VALIDATOR TOOLS
# ============================================================================

# Build gif-generator tool if needed
build_validator() {
    if [[ ! -f "$GIF_GENERATOR_BINARY" ]]; then
        log "Building gif-generator for validation..."
        ensure_directories
        
        local source_file="$PROJECT_ROOT/cmd/gif-generator/main.go"
        if [[ ! -f "$source_file" ]]; then
            error "gif-generator source not found: $source_file"
            return 1
        fi
        
        build_if_needed "$source_file" "$GIF_GENERATOR_BINARY"
    else
        debug "gif-generator already exists: $GIF_GENERATOR_BINARY"
    fi
}

# Check if validator tool is available
check_validator() {
    if [[ ! -x "$GIF_GENERATOR_BINARY" ]]; then
        log "gif-generator not found, building..."
        build_validator
    fi
    
    if [[ ! -x "$GIF_GENERATOR_BINARY" ]]; then
        error "Failed to build or find gif-generator validation tool"
        return 1
    fi
    
    return 0
}

# ============================================================================
# VALIDATION FUNCTIONS
# ============================================================================

# Validate a single character file
validate_character_file() {
    local character_file="$1"
    local character_name
    character_name=$(get_character_name "$character_file")
    
    debug "Validating character: $character_name ($character_file)"
    
    # Run validation using gif-generator
    local validation_output
    if validation_output=$("$GIF_GENERATOR_BINARY" character --file "$character_file" 2>&1); then
        return 0
    else
        # Validation failed
        if [[ "$DETAILED_OUTPUT" == "true" ]]; then
            error "Validation failed for $character_name:"
            echo "$validation_output" | sed 's/^/    /' >&2
        fi
        return 1
    fi
}

# Validate all character files
validate_all_characters() {
    log "Starting character validation..."
    
    # Find all character files
    local character_files
    readarray -t character_files < <(find_character_files)
    
    if [[ ${#character_files[@]} -eq 0 ]]; then
        warning "No character files found in $CHARACTERS_DIR"
        return 1
    fi
    
    log "Found ${#character_files[@]} character files to validate"
    
    # Validation counters
    local success_count=0
    local failure_count=0
    local failed_characters=()
    
    # Temporarily disable error exit for validation loop
    set +e
    
    for character_file in "${character_files[@]}"; do
        local character_name
        character_name=$(get_character_name "$character_file")
        
        if [[ "$VERBOSE" == "true" ]]; then
            log "Validating: $character_name"
        fi
        
        if validate_character_file "$character_file"; then
            if [[ "$VERBOSE" == "true" ]]; then
                success "$character_name"
            else
                echo "✓ $character_name"
            fi
            ((success_count++))
        else
            if [[ "$VERBOSE" == "true" ]]; then
                error "$character_name"
            else
                echo "✗ $character_name"
            fi
            
            failed_characters+=("$character_name")
            ((failure_count++))
            
            # Stop on first error if requested
            if [[ "$STOP_ON_FIRST_ERROR" == "true" ]]; then
                break
            fi
        fi
        
        # Show progress for large numbers of characters
        if [[ ${#character_files[@]} -gt 10 && "$VERBOSE" != "true" ]]; then
            show_progress $((success_count + failure_count)) ${#character_files[@]} "Validating"
        fi
    done
    
    # Re-enable error exit
    set -e
    
    # Report final results
    echo
    log "Validation complete: $success_count passed, $failure_count failed"
    
    if [[ $failure_count -gt 0 ]]; then
        echo
        warning "Failed characters:"
        for char in "${failed_characters[@]}"; do
            echo "  - $char"
        done
        
        if [[ "$DETAILED_OUTPUT" != "true" ]]; then
            echo
            log "For detailed error messages, run with --detailed option"
            log "To see specific errors for a character:"
            log "  $GIF_GENERATOR_BINARY character --file assets/characters/CHARACTER_NAME/character.json"
        fi
        
        return 1
    fi
    
    success "All characters passed validation! ✨"
    return 0
}

# ============================================================================
# QUICK VALIDATION FUNCTIONS
# ============================================================================

# Quick syntax check using basic JSON parsing
quick_json_check() {
    local character_file="$1"
    local character_name
    character_name=$(get_character_name "$character_file")
    
    # Check if file exists and is readable
    if [[ ! -f "$character_file" ]]; then
        error "$character_name: File not found"
        return 1
    fi
    
    if [[ ! -r "$character_file" ]]; then
        error "$character_name: File not readable"
        return 1
    fi
    
    # Basic JSON syntax check
    if ! jq empty "$character_file" >/dev/null 2>&1; then
        # Try with Python if jq is not available
        if command_exists python3; then
            if ! python3 -c "import json; json.load(open('$character_file'))" >/dev/null 2>&1; then
                error "$character_name: Invalid JSON syntax"
                return 1
            fi
        else
            warning "$character_name: Could not validate JSON syntax (no jq or python3)"
        fi
    fi
    
    return 0
}

# Run quick validation for all characters
quick_validate_all() {
    log "Running quick JSON syntax validation..."
    
    local character_files
    readarray -t character_files < <(find_character_files)
    
    local success_count=0
    local failure_count=0
    
    for character_file in "${character_files[@]}"; do
        if quick_json_check "$character_file"; then
            ((success_count++))
        else
            ((failure_count++))
        fi
    done
    
    log "Quick validation: $success_count passed, $failure_count failed"
    return $failure_count
}

# ============================================================================
# VALIDATION REPORTS
# ============================================================================

# Generate detailed validation report
generate_validation_report() {
    local report_file="$TEST_OUTPUT_DIR/character-validation-report-$(date +%Y%m%d-%H%M%S).md"
    
    log "Generating validation report: $(basename "$report_file")"
    
    cat > "$report_file" << EOF
# Character Validation Report

**Generated:** $(date -u '+%Y-%m-%d %H:%M:%S UTC')
**Script:** validate-characters.sh
**Validator:** $(basename "$GIF_GENERATOR_BINARY")

## Summary

Total characters found: $(find_character_files | wc -l)

## Individual Results

EOF
    
    # Add validation results for each character
    local character_files
    readarray -t character_files < <(find_character_files)
    
    for character_file in "${character_files[@]}"; do
        local character_name
        character_name=$(get_character_name "$character_file")
        
        echo "### $character_name" >> "$report_file"
        echo "" >> "$report_file"
        echo "**File:** \`$character_file\`" >> "$report_file"
        echo "" >> "$report_file"
        
        if validate_character_file "$character_file"; then
            echo "**Status:** ✅ PASSED" >> "$report_file"
        else
            echo "**Status:** ❌ FAILED" >> "$report_file"
            echo "" >> "$report_file"
            echo "**Errors:**" >> "$report_file"
            echo "\`\`\`" >> "$report_file"
            "$GIF_GENERATOR_BINARY" character --file "$character_file" 2>&1 || true >> "$report_file"
            echo "\`\`\`" >> "$report_file"
        fi
        
        echo "" >> "$report_file"
    done
    
    success "Validation report generated: $report_file"
}

# ============================================================================
# ARGUMENT PARSING AND MAIN
# ============================================================================

# Parse command line arguments
parse_arguments() {
    local run_validation=true
    local generate_report=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_usage
                exit 0
                ;;
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -d|--detailed)
                DETAILED_OUTPUT=true
                shift
                ;;
            -s|--stop-on-error)
                STOP_ON_FIRST_ERROR=true
                shift
                ;;
            --build-validator)
                build_validator
                exit $?
                ;;
            --quick)
                quick_validate_all
                exit $?
                ;;
            --report)
                generate_report=true
                shift
                ;;
            *)
                error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done
    
    echo "$run_validation $generate_report"
}

# Main entry point
main() {
    # Set up error handling
    setup_error_handling
    init_common
    
    # Parse arguments
    local args
    read -r run_validation generate_report <<< "$(parse_arguments "$@")"
    
    # Ensure we have the validation tool
    check_validator
    
    # Run validation
    local exit_code=0
    if [[ "$run_validation" == "true" ]]; then
        if ! validate_all_characters; then
            exit_code=1
        fi
    fi
    
    # Generate report if requested
    if [[ "$generate_report" == "true" ]]; then
        generate_validation_report
    fi
    
    exit $exit_code
}

# Run main function if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
