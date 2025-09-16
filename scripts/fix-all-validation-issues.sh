#!/bin/bash

# fix-all-validation-issues.sh
# Comprehensive script to fix all remaining validation issues in character JSON files

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
CHARACTERS_DIR="$PROJECT_ROOT/assets/characters"

log() {
    echo "[$(date '+%H:%M:%S')] $*"
}

# Function to fix battle system stats - add missing max values
fix_battle_stats() {
    local file="$1"
    local character_name="$(basename "$(dirname "$file")")"
    
    log "Fixing battle stats in $character_name"
    
    python3 -c "
import json
import sys

file_path = '$file'
try:
    with open(file_path, 'r') as f:
        data = json.load(f)
    
    fixed = False
    
    # Fix battle system stats
    if 'battleSystem' in data and 'battleStats' in data['battleSystem']:
        battle_stats = data['battleSystem']['battleStats']
        
        # Standard max values for different stats
        stat_max_values = {
            'hp': 300,
            'attack': 100, 
            'defense': 100,
            'speed': 100,
            'magic': 150,
            'mana': 200,
            'stamina': 150
        }
        
        for stat_name, stat_config in battle_stats.items():
            if isinstance(stat_config, dict):
                # Add max value if missing or 0
                if 'max' not in stat_config or stat_config.get('max', 0) == 0:
                    base_value = stat_config.get('base', 50)
                    default_max = stat_max_values.get(stat_name, 100)
                    
                    # Ensure max is at least 2x the base value, but use reasonable defaults
                    calculated_max = max(base_value * 2, default_max)
                    stat_config['max'] = calculated_max
                    
                    print(f'    Added max value {calculated_max} for {stat_name} (base: {base_value})')
                    fixed = True
                
                # Ensure max >= base
                elif stat_config.get('max', 0) < stat_config.get('base', 0):
                    base_value = stat_config.get('base', 50)
                    stat_config['max'] = max(base_value * 2, stat_max_values.get(stat_name, 100))
                    print(f'    Fixed max value for {stat_name} to be >= base value')
                    fixed = True
    
    if fixed:
        with open(file_path, 'w') as f:
            json.dump(data, f, indent=2)
        print(f'    ✅ Fixed battle stats in {file_path}')
    else:
        print(f'    ✅ Battle stats already valid or not present')
        
except Exception as e:
    print(f'    ❌ Error processing {file_path}: {e}')
    sys.exit(1)
"
}

# Function to fix invalid triggers
fix_invalid_triggers() {
    local file="$1"
    local character_name="$(basename "$(dirname "$file")")"
    
    log "Fixing invalid triggers in $character_name"
    
    # Fix common invalid trigger patterns
    sed -i 's/"ctrl+click"/"ctrl+shift+click"/g' "$file"
    sed -i 's/"alt+click"/"alt+shift+click"/g' "$file"
    sed -i 's/"meta+click"/"ctrl+shift+click"/g' "$file"
    sed -i 's/"cmd+click"/"ctrl+shift+click"/g' "$file"
    
    echo "    ✅ Fixed invalid triggers"
}

# Function to fix missing animation references
fix_missing_animations() {
    local file="$1"
    local character_name="$(basename "$(dirname "$file")")"
    
    log "Fixing missing animation references in $character_name"
    
    python3 -c "
import json
import sys

file_path = '$file'
try:
    with open(file_path, 'r') as f:
        data = json.load(f)
    
    fixed = False
    
    # Get available animations
    available_animations = set()
    if 'animations' in data and isinstance(data['animations'], dict):
        available_animations = set(data['animations'].keys())
    
    # Add basic animations if missing
    basic_animations = ['idle', 'talking', 'happy', 'attack', 'defend', 'heal']
    if 'animations' not in data:
        data['animations'] = {}
        
    for anim in basic_animations:
        if anim not in data['animations']:
            data['animations'][anim] = f'animations/{anim}.gif'
            available_animations.add(anim)
            fixed = True
            print(f'    Added missing animation: {anim}')
    
    # Fix battle system animations if battleSystem exists
    if 'battleSystem' in data and data['battleSystem'].get('enabled', False):
        battle_animations = ['attack', 'defend', 'heal']
        for anim in battle_animations:
            if anim not in data['animations']:
                data['animations'][anim] = f'animations/{anim}.gif'
                available_animations.add(anim)
                fixed = True
                print(f'    Added missing battle animation: {anim}')
    
    if fixed:
        with open(file_path, 'w') as f:
            json.dump(data, f, indent=2)
        print(f'    ✅ Fixed animations in {file_path}')
    else:
        print(f'    ✅ Animations already complete')
        
except Exception as e:
    print(f'    ❌ Error processing {file_path}: {e}')
    sys.exit(1)
"
}

# Function to fix general validation issues
fix_general_issues() {
    local file="$1"
    local character_name="$(basename "$(dirname "$file")")"
    
    log "Fixing general validation issues in $character_name"
    
    python3 -c "
import json
import sys

file_path = '$file'
try:
    with open(file_path, 'r') as f:
        data = json.load(f)
    
    fixed = False
    
    # Ensure required fields exist
    if 'categories' not in data:
        data['categories'] = ['conversation']
        fixed = True
        print('    Added default categories')
    
    if 'triggers' not in data:
        data['triggers'] = ['click']
        fixed = True
        print('    Added default triggers')
    
    # Fix behavior settings if missing
    if 'behavior' not in data:
        data['behavior'] = {
            'idleTimeout': 30,
            'defaultSize': 128
        }
        fixed = True
        print('    Added default behavior settings')
    else:
        if 'idleTimeout' not in data['behavior'] or data['behavior'].get('idleTimeout', 0) < 10:
            data['behavior']['idleTimeout'] = 30
            fixed = True
            print('    Fixed idleTimeout')
        
        if 'defaultSize' not in data['behavior'] or data['behavior'].get('defaultSize', 0) < 64:
            data['behavior']['defaultSize'] = 128
            fixed = True
            print('    Fixed defaultSize')
    
    # Ensure dialogs have proper structure
    if 'dialogs' in data:
        for i, dialog in enumerate(data['dialogs']):
            if isinstance(dialog, dict):
                if 'animation' not in dialog:
                    dialog['animation'] = 'talking'
                    fixed = True
                    print(f'    Added missing animation to dialog {i}')
    
    # Fix general events
    if 'generalEvents' in data:
        for i, event in enumerate(data['generalEvents']):
            if isinstance(event, dict):
                if 'description' not in event:
                    event['description'] = f\"Event: {event.get('name', f'event_{i}')}\"
                    fixed = True
                    print(f'    Added description to general event {i}')
                
                if 'category' not in event:
                    event['category'] = 'conversation'
                    fixed = True
                    print(f'    Added category to general event {i}')
                
                if 'trigger' not in event:
                    event['trigger'] = 'click'
                    fixed = True
                    print(f'    Added trigger to general event {i}')
    
    if fixed:
        with open(file_path, 'w') as f:
            json.dump(data, f, indent=2)
        print(f'    ✅ Fixed general issues in {file_path}')
    else:
        print(f'    ✅ General validation already complete')
        
except Exception as e:
    print(f'    ❌ Error processing {file_path}: {e}')
    sys.exit(1)
"
}

# Main processing loop
log "Starting comprehensive character validation fixes..."

# Find all character files
character_files=()
while IFS= read -r file; do
    character_files+=("$file")
done < <(find "$CHARACTERS_DIR" -mindepth 2 -maxdepth 2 -name "character.json" -type f | sort)

log "Found ${#character_files[@]} character files to fix"

for character_file in "${character_files[@]}"; do
    character_name="$(basename "$(dirname "$character_file")")"
    echo
    log "Processing character: $character_name"
    
    # Create backup
    cp "$character_file" "${character_file}.backup"
    
    # Apply all fixes
    fix_general_issues "$character_file"
    fix_invalid_triggers "$character_file"
    fix_missing_animations "$character_file"
    fix_battle_stats "$character_file"
    
    echo "  ✅ $character_name processing complete"
done

echo
log "All character fixes applied!"
log "Backups created with .backup extension"
log "Running validation check..."

# Run validation to see results
"$PROJECT_ROOT/scripts/validate-characters.sh"
