#!/bin/bash

# scripts/character-management/fix-validation-issues.sh
# Character validation issue fixing script
#
# Automatically fixes common validation issues in character JSON files
# including schema compliance, missing animations, and invalid categories.
#
# Usage: ./scripts/character-management/fix-validation-issues.sh [OPTIONS]
#
# Dependencies:
# - Python 3.x (for JSON manipulation)
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
# FIX CONFIGURATION
# ============================================================================

# Fix settings
CREATE_BACKUPS=true
VERBOSE="${DDS_VERBOSE}"
DRY_RUN="${DDS_DRY_RUN}"
FIX_ALL_ISSUES=true

# Valid values from config
VALID_CATEGORIES=("${DDS_VALID_CATEGORIES[@]}")
VALID_TRIGGERS=("${DDS_VALID_TRIGGERS[@]}")

# ============================================================================
# HELP AND USAGE
# ============================================================================

show_usage() {
    cat << EOF
Usage: $0 [OPTIONS] [CHARACTER_NAME]

Automatically fix common validation issues in character JSON files.

ARGUMENTS:
    CHARACTER_NAME      Fix specific character only (default: fix all)

OPTIONS:
    --dry-run          Show what would be fixed without making changes
    --no-backup        Don't create backup files
    --verbose          Show detailed fix information
    --specific-fix FIX Apply only specific fix type
    --help            Show this help message

FIX TYPES:
    categories         Fix invalid event categories
    triggers          Fix invalid event triggers
    animations        Add missing animation references
    descriptions      Fix description length issues
    battle-stats      Fix battle system statistics
    romance-dialogs   Fix romance dialog structure

COMMON FIXES:
    ✓ Replace invalid categories with 'conversation'
    ✓ Fix invalid trigger combinations
    ✓ Add missing animation files to character definitions
    ✓ Truncate overly long descriptions
    ✓ Add missing max values for battle stats
    ✓ Fix romance dialog trigger formats

EXAMPLES:
    $0                           # Fix all characters
    $0 --dry-run                 # Show what would be fixed
    $0 romance                   # Fix only romance character
    $0 --specific-fix categories # Fix only category issues

BACKUP:
    Backup files are created with .backup extension unless --no-backup is used.

EOF
}

# ============================================================================
# PYTHON HELPERS
# ============================================================================

# Generate Python script for JSON manipulation
create_json_fixer_script() {
    cat << 'EOF'
import json
import sys
import re
from pathlib import Path

class CharacterJSONFixer:
    def __init__(self, file_path, valid_categories, valid_triggers):
        self.file_path = file_path
        self.valid_categories = valid_categories
        self.valid_triggers = valid_triggers
        self.fixes_applied = []
        
    def load_json(self):
        """Load JSON file"""
        try:
            with open(self.file_path, 'r', encoding='utf-8') as f:
                return json.load(f)
        except Exception as e:
            print(f"Error loading JSON: {e}", file=sys.stderr)
            return None
    
    def save_json(self, data):
        """Save JSON file"""
        try:
            with open(self.file_path, 'w', encoding='utf-8') as f:
                json.dump(data, f, indent=2, ensure_ascii=False)
            return True
        except Exception as e:
            print(f"Error saving JSON: {e}", file=sys.stderr)
            return False
    
    def fix_invalid_categories(self, data):
        """Fix invalid categories in events"""
        category_mapping = {
            'mystical': 'conversation',
            'support': 'conversation',
            'astronomy': 'conversation',
            'science': 'conversation',
            'philosophy': 'conversation',
            'care': 'conversation'
        }
        
        fixed_count = 0
        
        # Fix general events
        if 'generalEvents' in data and isinstance(data['generalEvents'], list):
            for event in data['generalEvents']:
                if 'category' in event and event['category'] in category_mapping:
                    old_category = event['category']
                    event['category'] = category_mapping[old_category]
                    fixed_count += 1
                    
        # Fix romance features
        if 'romanceFeatures' in data:
            if 'events' in data['romanceFeatures'] and isinstance(data['romanceFeatures']['events'], list):
                for event in data['romanceFeatures']['events']:
                    if 'category' in event and event['category'] in category_mapping:
                        old_category = event['category']
                        event['category'] = category_mapping[old_category]
                        fixed_count += 1
        
        # Fix game features
        if 'gameFeatures' in data and 'interactions' in data['gameFeatures']:
            if isinstance(data['gameFeatures']['interactions'], list):
                for interaction in data['gameFeatures']['interactions']:
                    if 'category' in interaction and interaction['category'] in category_mapping:
                        old_category = interaction['category']
                        interaction['category'] = category_mapping[old_category]
                        fixed_count += 1
        
        if fixed_count > 0:
            self.fixes_applied.append(f"Fixed {fixed_count} invalid categories")
        
        return data
    
    def fix_invalid_triggers(self, data):
        """Fix invalid trigger combinations"""
        trigger_mapping = {
            'ctrl+click': 'ctrl+shift+click',
            'alt+click': 'alt+shift+click',
            'meta+click': 'ctrl+shift+click',
            'cmd+click': 'ctrl+shift+click'
        }
        
        fixed_count = 0
        
        def fix_triggers_in_list(events_list):
            nonlocal fixed_count
            for event in events_list:
                if 'trigger' in event and event['trigger'] in trigger_mapping:
                    old_trigger = event['trigger']
                    event['trigger'] = trigger_mapping[old_trigger]
                    fixed_count += 1
        
        # Fix general events
        if 'generalEvents' in data and isinstance(data['generalEvents'], list):
            fix_triggers_in_list(data['generalEvents'])
        
        # Fix romance features
        if 'romanceFeatures' in data:
            if 'events' in data['romanceFeatures'] and isinstance(data['romanceFeatures']['events'], list):
                fix_triggers_in_list(data['romanceFeatures']['events'])
            if 'dialogs' in data['romanceFeatures'] and isinstance(data['romanceFeatures']['dialogs'], list):
                fix_triggers_in_list(data['romanceFeatures']['dialogs'])
        
        # Fix romance dialogs (alternative structure)
        if 'romanceDialogs' in data and isinstance(data['romanceDialogs'], list):
            fix_triggers_in_list(data['romanceDialogs'])
        
        if fixed_count > 0:
            self.fixes_applied.append(f"Fixed {fixed_count} invalid triggers")
        
        return data
    
    def fix_missing_animations(self, data):
        """Add missing animation references"""
        if 'animations' not in data:
            data['animations'] = {}
        
        # Common animations that might be missing
        common_animations = {
            'smug': 'animations/smug.gif',
            'confident': 'animations/confident.gif',
            'proud': 'animations/proud.gif',
            'calm': 'animations/calm.gif',
            'serious': 'animations/serious.gif',
            'cheerful': 'animations/cheerful.gif',
            'nervous': 'animations/nervous.gif',
            'playful': 'animations/playful.gif',
            'caring': 'animations/caring.gif'
        }
        
        added_count = 0
        file_content = json.dumps(data)
        
        for anim_name, anim_path in common_animations.items():
            if f'"{anim_name}"' in file_content and anim_name not in data['animations']:
                data['animations'][anim_name] = anim_path
                added_count += 1
        
        if added_count > 0:
            self.fixes_applied.append(f"Added {added_count} missing animation references")
        
        return data
    
    def fix_description_length(self, data):
        """Fix overly long descriptions"""
        fixed_count = 0
        max_length = 500
        
        if 'description' in data and len(data['description']) > max_length:
            data['description'] = data['description'][:max_length].rsplit(' ', 1)[0] + '...'
            fixed_count += 1
        
        if fixed_count > 0:
            self.fixes_applied.append(f"Fixed {fixed_count} overly long descriptions")
        
        return data
    
    def fix_battle_stats(self, data):
        """Fix battle system stats - add missing max values"""
        fixed_count = 0
        
        if 'battleSystem' in data and 'battleStats' in data['battleSystem']:
            stats = data['battleSystem']['battleStats']
            
            # Add missing max values
            if 'health' in stats and 'maxHealth' not in stats:
                stats['maxHealth'] = stats.get('health', 100)
                fixed_count += 1
            
            if 'mana' in stats and 'maxMana' not in stats:
                stats['maxMana'] = stats.get('mana', 50)
                fixed_count += 1
            
            if 'energy' in stats and 'maxEnergy' not in stats:
                stats['maxEnergy'] = stats.get('energy', 100)
                fixed_count += 1
        
        if fixed_count > 0:
            self.fixes_applied.append(f"Fixed {fixed_count} battle stat issues")
        
        return data
    
    def fix_romance_dialogs(self, data):
        """Fix romance dialog structure issues"""
        fixed_count = 0
        
        # Fix romance features dialogs
        if 'romanceFeatures' in data and 'dialogs' in data['romanceFeatures']:
            if isinstance(data['romanceFeatures']['dialogs'], list):
                for dialog in data['romanceFeatures']['dialogs']:
                    # Ensure trigger is valid
                    if 'trigger' in dialog and dialog['trigger'] not in self.valid_triggers:
                        dialog['trigger'] = 'click'
                        fixed_count += 1
        
        # Fix standalone romance dialogs
        if 'romanceDialogs' in data and isinstance(data['romanceDialogs'], list):
            for dialog in data['romanceDialogs']:
                if 'trigger' in dialog and dialog['trigger'] not in self.valid_triggers:
                    dialog['trigger'] = 'click'
                    fixed_count += 1
        
        if fixed_count > 0:
            self.fixes_applied.append(f"Fixed {fixed_count} romance dialog issues")
        
        return data
    
    def apply_all_fixes(self, data):
        """Apply all available fixes"""
        data = self.fix_invalid_categories(data)
        data = self.fix_invalid_triggers(data)
        data = self.fix_missing_animations(data)
        data = self.fix_description_length(data)
        data = self.fix_battle_stats(data)
        data = self.fix_romance_dialogs(data)
        return data
    
    def fix_character_file(self, specific_fix=None):
        """Fix character file with specified fixes"""
        data = self.load_json()
        if data is None:
            return False
        
        if specific_fix:
            if specific_fix == 'categories':
                data = self.fix_invalid_categories(data)
            elif specific_fix == 'triggers':
                data = self.fix_invalid_triggers(data)
            elif specific_fix == 'animations':
                data = self.fix_missing_animations(data)
            elif specific_fix == 'descriptions':
                data = self.fix_description_length(data)
            elif specific_fix == 'battle-stats':
                data = self.fix_battle_stats(data)
            elif specific_fix == 'romance-dialogs':
                data = self.fix_romance_dialogs(data)
            else:
                print(f"Unknown fix type: {specific_fix}", file=sys.stderr)
                return False
        else:
            data = self.apply_all_fixes(data)
        
        if self.fixes_applied:
            return self.save_json(data)
        
        return True  # No fixes needed

def main():
    if len(sys.argv) < 4:
        print("Usage: python3 fixer.py <file_path> <valid_categories> <valid_triggers> [specific_fix]")
        sys.exit(1)
    
    file_path = sys.argv[1]
    valid_categories = sys.argv[2].split(',')
    valid_triggers = sys.argv[3].split(',')
    specific_fix = sys.argv[4] if len(sys.argv) > 4 else None
    
    fixer = CharacterJSONFixer(file_path, valid_categories, valid_triggers)
    
    if fixer.fix_character_file(specific_fix):
        if fixer.fixes_applied:
            for fix in fixer.fixes_applied:
                print(f"    ✅ {fix}")
        else:
            print("    ✓ No fixes needed")
    else:
        print(f"    ❌ Failed to fix {file_path}", file=sys.stderr)
        sys.exit(1)

if __name__ == "__main__":
    main()
EOF
}

# ============================================================================
# CHARACTER FIXING FUNCTIONS
# ============================================================================

# Fix a single character file
fix_character_file() {
    local character_file="$1"
    local specific_fix="${2:-}"
    local character_name
    character_name=$(get_character_name "$character_file")
    
    log "Fixing character: $character_name"
    
    # Create backup if requested
    if [[ "$CREATE_BACKUPS" == "true" ]]; then
        local backup_file="${character_file}.backup-$(date +%Y%m%d-%H%M%S)"
        cp "$character_file" "$backup_file"
        debug "Created backup: $(basename "$backup_file")"
    fi
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log "  DRY RUN: Would fix $(basename "$character_file")"
        return 0
    fi
    
    # Create temporary Python script
    local temp_script
    temp_script=$(mktemp /tmp/character_fixer_XXXXXX.py)
    create_json_fixer_script > "$temp_script"
    
    # Prepare arguments
    local categories_arg
    local triggers_arg
    categories_arg=$(IFS=','; echo "${VALID_CATEGORIES[*]}")
    triggers_arg=$(IFS=','; echo "${VALID_TRIGGERS[*]}")
    
    # Run the fixer
    if python3 "$temp_script" "$character_file" "$categories_arg" "$triggers_arg" "$specific_fix"; then
        success "Fixed: $character_name"
    else
        error "Failed to fix: $character_name"
        rm -f "$temp_script"
        return 1
    fi
    
    # Cleanup
    rm -f "$temp_script"
    return 0
}

# Fix all character files
fix_all_characters() {
    local specific_fix="${1:-}"
    
    log "Starting character validation fixes..."
    
    if [[ -n "$specific_fix" ]]; then
        log "Applying specific fix: $specific_fix"
    fi
    
    # Find all character files
    local character_files
    readarray -t character_files < <(find_character_files)
    
    if [[ ${#character_files[@]} -eq 0 ]]; then
        warning "No character files found in $CHARACTERS_DIR"
        return 1
    fi
    
    log "Found ${#character_files[@]} character files to fix"
    
    local success_count=0
    local failure_count=0
    local failed_characters=()
    
    for character_file in "${character_files[@]}"; do
        local character_name
        character_name=$(get_character_name "$character_file")
        
        if fix_character_file "$character_file" "$specific_fix"; then
            ((success_count++))
        else
            failed_characters+=("$character_name")
            ((failure_count++))
        fi
        
        # Show progress
        if [[ ${#character_files[@]} -gt 10 && "$VERBOSE" != "true" ]]; then
            show_progress $((success_count + failure_count)) ${#character_files[@]} "Fixing characters"
        fi
    done
    
    # Report results
    echo
    log "Character fixes complete: $success_count successful, $failure_count failed"
    
    if [[ $failure_count -gt 0 ]]; then
        warning "Failed to fix characters: ${failed_characters[*]}"
        return 1
    fi
    
    if [[ "$CREATE_BACKUPS" == "true" ]]; then
        log "Backup files created with .backup-YYYYMMDD-HHMMSS extension"
    fi
    
    success "All character fixes applied successfully! ✨"
    return 0
}

# ============================================================================
# VALIDATION VERIFICATION
# ============================================================================

# Verify fixes by running validation
verify_fixes() {
    log "Verifying fixes by running character validation..."
    
    local validation_script="$SCRIPT_DIR/../validation/validate-characters.sh"
    
    if [[ -x "$validation_script" ]]; then
        if "$validation_script"; then
            success "✅ All characters now pass validation!"
            return 0
        else
            warning "⚠️ Some characters still have validation issues"
            return 1
        fi
    else
        warning "Validation script not found, skipping verification"
        return 0
    fi
}

# ============================================================================
# ARGUMENT PARSING AND MAIN
# ============================================================================

# Parse command line arguments
parse_arguments() {
    local character_name=""
    local specific_fix=""
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_usage
                exit 0
                ;;
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            --no-backup)
                CREATE_BACKUPS=false
                shift
                ;;
            --verbose)
                VERBOSE=true
                shift
                ;;
            --specific-fix)
                specific_fix="$2"
                shift 2
                ;;
            -*)
                error "Unknown option: $1"
                show_usage
                exit 1
                ;;
            *)
                character_name="$1"
                shift
                ;;
        esac
    done
    
    echo "$character_name" "$specific_fix"
}

# Main entry point
main() {
    # Set up error handling
    setup_error_handling
    init_common
    
    # Check dependencies
    require_commands python3
    
    # Parse arguments
    local args
    read -r character_name specific_fix <<< "$(parse_arguments "$@")"
    
    # Fix characters
    if [[ -n "$character_name" ]]; then
        # Fix specific character
        local character_file="$CHARACTERS_DIR/$character_name/character.json"
        
        if [[ ! -f "$character_file" ]]; then
            error "Character not found: $character_name"
            error "Expected file: $character_file"
            exit 1
        fi
        
        fix_character_file "$character_file" "$specific_fix"
    else
        # Fix all characters
        fix_all_characters "$specific_fix"
    fi
    
    # Verify fixes if not dry run
    if [[ "$DRY_RUN" != "true" ]]; then
        echo
        verify_fixes
    fi
}

# Run main function if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
