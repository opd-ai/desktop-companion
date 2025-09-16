#!/bin/bash

# scripts/validation/validate-animations.sh
# Animation file validation script
#
# Ensures all required animation files exist for each character
# and validates animation file integrity.
#
# Usage: ./scripts/validation/validate-animations.sh [OPTIONS]
#
# Dependencies:
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

# Animation validation settings
CHECK_FILE_INTEGRITY=true
VERBOSE="${DDS_VERBOSE}"
DETAILED_OUTPUT=false

# Basic animations required for all characters (from config)
BASIC_ANIMATIONS=("${DDS_BASIC_ANIMATIONS[@]}")

# Character-specific animations (from config)
declare -A SPECIFIC_ANIMATIONS
for char in "${!DDS_SPECIFIC_ANIMATIONS[@]}"; do
    SPECIFIC_ANIMATIONS["$char"]="${DDS_SPECIFIC_ANIMATIONS[$char]}"
done

# ============================================================================
# HELP AND USAGE
# ============================================================================

show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Validate animation files for all characters in the project.

OPTIONS:
    -v, --verbose         Show detailed validation output
    -d, --detailed        Show detailed error messages
    --no-integrity-check  Skip GIF file integrity validation
    --list-requirements   List animation requirements for each character type
    --help               Show this help message

VALIDATION CHECKS:
    - Basic animations present for all characters
    - Character-specific animations for special types
    - Animation file integrity (GIF format validation)
    - Animation directory structure

EXAMPLES:
    $0                          # Validate all character animations
    $0 --verbose --detailed     # Verbose output with detailed errors
    $0 --list-requirements      # Show animation requirements

See: docs/CHARACTER_CREATION_TUTORIAL.md for animation requirements.
EOF
}

# ============================================================================
# ANIMATION REQUIREMENTS
# ============================================================================

# List animation requirements for character types
list_animation_requirements() {
    cat << EOF
Animation Requirements by Character Type
======================================

Basic Animations (Required for ALL characters):
$(printf "  • %s\n" "${BASIC_ANIMATIONS[@]}")

Character-Specific Animations:
EOF
    
    for char_type in "${!SPECIFIC_ANIMATIONS[@]}"; do
        echo
        echo "$char_type characters:"
        # shellcheck disable=SC2086
        printf "  • %s\n" ${SPECIFIC_ANIMATIONS[$char_type]}
    done
    
    cat << EOF

Animation File Requirements:
  • Format: GIF with transparency support
  • Size: Recommended 64x64 to 256x256 pixels
  • Duration: 1-3 seconds for idle animations
  • Frames: 10-30 frames for smooth animation
  • Location: assets/characters/CHARACTER_NAME/animations/

EOF
}

# ============================================================================
# VALIDATION FUNCTIONS
# ============================================================================

# Check if a file is a valid GIF
validate_gif_integrity() {
    local gif_file="$1"
    
    if [[ ! -f "$gif_file" ]]; then
        return 1
    fi
    
    # Check file signature (GIF87a or GIF89a)
    local header
    header=$(head -c 6 "$gif_file" 2>/dev/null)
    
    if [[ "$header" == "GIF87a" || "$header" == "GIF89a" ]]; then
        return 0
    else
        return 1
    fi
}

# Validate animations for a single character
validate_character_animations() {
    local char_dir="$1"
    local char_name
    char_name=$(basename "$char_dir")
    
    debug "Validating animations for character: $char_name"
    
    local animations_dir="$char_dir/animations"
    local validation_errors=()
    local validation_warnings=()
    
    # Check if animations directory exists
    if [[ ! -d "$animations_dir" ]]; then
        validation_errors+=("Missing animations directory")
        return 1
    fi
    
    # Check basic animations
    local missing_basic=()
    for animation in "${BASIC_ANIMATIONS[@]}"; do
        local animation_file="$animations_dir/$animation"
        
        if [[ ! -f "$animation_file" ]]; then
            missing_basic+=("$animation")
        elif [[ "$CHECK_FILE_INTEGRITY" == "true" ]]; then
            if ! validate_gif_integrity "$animation_file"; then
                validation_errors+=("Invalid GIF format: $animation")
            fi
        fi
    done
    
    if [[ ${#missing_basic[@]} -gt 0 ]]; then
        validation_errors+=("Missing basic animations: ${missing_basic[*]}")
    fi
    
    # Check character-specific animations
    if [[ -n "${SPECIFIC_ANIMATIONS[$char_name]:-}" ]]; then
        local missing_specific=()
        # shellcheck disable=SC2086
        for animation in ${SPECIFIC_ANIMATIONS[$char_name]}; do
            local animation_file="$animations_dir/$animation"
            
            if [[ ! -f "$animation_file" ]]; then
                missing_specific+=("$animation")
            elif [[ "$CHECK_FILE_INTEGRITY" == "true" ]]; then
                if ! validate_gif_integrity "$animation_file"; then
                    validation_warnings+=("Invalid GIF format: $animation")
                fi
            fi
        done
        
        if [[ ${#missing_specific[@]} -gt 0 ]]; then
            validation_warnings+=("Missing character-specific animations: ${missing_specific[*]}")
        fi
    fi
    
    # Report results
    if [[ ${#validation_errors[@]} -gt 0 ]]; then
        if [[ "$DETAILED_OUTPUT" == "true" ]]; then
            error "Animation validation failed for $char_name:"
            for err in "${validation_errors[@]}"; do
                echo "    ✗ $err" >&2
            done
        fi
        return 1
    fi
    
    # Show warnings
    if [[ ${#validation_warnings[@]} -gt 0 && "$DETAILED_OUTPUT" == "true" ]]; then
        warning "Animation warnings for $char_name:"
        for warn in "${validation_warnings[@]}"; do
            echo "    ⚠ $warn" >&2
        done
    fi
    
    return 0
}

# Validate all character animations
validate_all_animations() {
    log "Starting animation validation..."
    
    # Find all character directories
    local character_dirs
    readarray -t character_dirs < <(find "$CHARACTERS_DIR" -maxdepth 1 -type d -not -path "$CHARACTERS_DIR")
    
    if [[ ${#character_dirs[@]} -eq 0 ]]; then
        warning "No character directories found in $CHARACTERS_DIR"
        return 1
    fi
    
    log "Found ${#character_dirs[@]} character directories to validate"
    
    # Validation counters
    local success_count=0
    local failure_count=0
    local warning_count=0
    local failed_characters=()
    
    for char_dir in "${character_dirs[@]}"; do
        local char_name
        char_name=$(basename "$char_dir")
        
        # Skip non-character directories
        if [[ ! -f "$char_dir/character.json" ]]; then
            debug "Skipping $char_name (no character.json)"
            continue
        fi
        
        if [[ "$VERBOSE" == "true" ]]; then
            log "Validating animations: $char_name"
        fi
        
        if validate_character_animations "$char_dir"; then
            if [[ "$VERBOSE" == "true" ]]; then
                success "$char_name animations valid"
            else
                echo "✓ $char_name"
            fi
            ((success_count++))
        else
            if [[ "$VERBOSE" == "true" ]]; then
                error "$char_name animations failed validation"
            else
                echo "✗ $char_name"
            fi
            
            failed_characters+=("$char_name")
            ((failure_count++))
        fi
        
        # Show progress for large numbers of characters
        if [[ ${#character_dirs[@]} -gt 10 && "$VERBOSE" != "true" ]]; then
            show_progress $((success_count + failure_count)) ${#character_dirs[@]} "Validating animations"
        fi
    done
    
    # Report final results
    echo
    log "Animation validation complete: $success_count passed, $failure_count failed"
    
    if [[ $failure_count -gt 0 ]]; then
        echo
        warning "Characters with animation issues:"
        for char in "${failed_characters[@]}"; do
            echo "  - $char"
        done
        
        if [[ "$DETAILED_OUTPUT" != "true" ]]; then
            echo
            log "For detailed error messages, run with --detailed option"
        fi
        
        return 1
    fi
    
    success "All character animations validated successfully! ✨"
    return 0
}

# ============================================================================
# REPORTING FUNCTIONS
# ============================================================================

# Generate animation validation report
generate_animation_report() {
    local report_file="$TEST_OUTPUT_DIR/animation-validation-report-$(date +%Y%m%d-%H%M%S).md"
    
    log "Generating animation validation report: $(basename "$report_file")"
    
    cat > "$report_file" << EOF
# Animation Validation Report

**Generated:** $(date -u '+%Y-%m-%d %H:%M:%S UTC')
**Script:** validate-animations.sh

## Requirements

### Basic Animations (All Characters)
$(printf "- %s\n" "${BASIC_ANIMATIONS[@]}")

### Character-Specific Animations
$(for char_type in "${!SPECIFIC_ANIMATIONS[@]}"; do
    echo "#### $char_type"
    # shellcheck disable=SC2086
    printf "- %s\n" ${SPECIFIC_ANIMATIONS[$char_type]}
    echo
done)

## Validation Results

EOF
    
    # Add validation results for each character
    local character_dirs
    readarray -t character_dirs < <(find "$CHARACTERS_DIR" -maxdepth 1 -type d -not -path "$CHARACTERS_DIR")
    
    for char_dir in "${character_dirs[@]}"; do
        local char_name
        char_name=$(basename "$char_dir")
        
        # Skip non-character directories
        [[ ! -f "$char_dir/character.json" ]] && continue
        
        echo "### $char_name" >> "$report_file"
        echo "" >> "$report_file"
        
        if validate_character_animations "$char_dir"; then
            echo "**Status:** ✅ PASSED" >> "$report_file"
        else
            echo "**Status:** ❌ FAILED" >> "$report_file"
        fi
        
        # List available animations
        local animations_dir="$char_dir/animations"
        if [[ -d "$animations_dir" ]]; then
            echo "" >> "$report_file"
            echo "**Available Animations:**" >> "$report_file"
            find "$animations_dir" -name "*.gif" -exec basename {} \; | sort | while read -r anim; do
                echo "- $anim" >> "$report_file"
            done
        fi
        
        echo "" >> "$report_file"
    done
    
    success "Animation validation report generated: $report_file"
}

# ============================================================================
# ARGUMENT PARSING AND MAIN
# ============================================================================

# Parse command line arguments
parse_arguments() {
    local action="validate"
    
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
            --no-integrity-check)
                CHECK_FILE_INTEGRITY=false
                shift
                ;;
            --list-requirements)
                action="list-requirements"
                shift
                ;;
            --report)
                action="report"
                shift
                ;;
            *)
                error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done
    
    echo "$action"
}

# Main entry point
main() {
    # Set up error handling
    setup_error_handling
    init_common
    
    # Parse arguments
    local action
    action=$(parse_arguments "$@")
    
    # Execute requested action
    case "$action" in
        validate)
            validate_all_animations
            ;;
        list-requirements)
            list_animation_requirements
            ;;
        report)
            generate_animation_report
            ;;
        *)
            error "Unknown action: $action"
            exit 1
            ;;
    esac
}

# Run main function if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
