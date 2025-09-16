#!/bin/bash

# fix-character-validation.sh
# Fix common validation issues in character JSON files

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
CHARACTERS_DIR="$PROJECT_ROOT/assets/characters"

log() {
    echo "[$(date '+%H:%M:%S')] $*"
}

# Function to fix a single character file
fix_character_file() {
    local character_file="$1"
    local character_name="$(basename "$(dirname "$character_file")")"
    local backup_file="${character_file}.backup"
    
    log "Fixing character: $character_name"
    
    # Create backup
    cp "$character_file" "$backup_file"
    
    # Apply fixes using sed and temporary files
    local temp_file=$(mktemp)
    
    # Fix 1: Replace invalid categories with 'conversation'
    sed 's/"category": "mystical"/"category": "conversation"/g; 
         s/"category": "support"/"category": "conversation"/g; 
         s/"category": "astronomy"/"category": "conversation"/g; 
         s/"category": "science"/"category": "conversation"/g; 
         s/"category": "philosophy"/"category": "conversation"/g' \
         "$character_file" > "$temp_file"
    
    # Fix 2: Replace invalid triggers
    sed 's/"alt+click"/"alt+shift+click"/g' "$temp_file" > "${temp_file}.2"
    
    # Fix 3: Add triggers to general events that don't have them
    # This is more complex - we'll use a Python script for this
    python3 - "$temp_file.2" "$character_file" << 'EOF'
import json
import sys
import re

def fix_general_events(data):
    """Add missing triggers to general events"""
    if 'generalEvents' in data and isinstance(data['generalEvents'], list):
        for event in data['generalEvents']:
            if isinstance(event, dict) and 'trigger' not in event:
                event['trigger'] = 'automatic'
    return data

def fix_invalid_stats_in_achievements(data):
    """Fix achievements that reference non-existent stats"""
    if 'progression' in data and 'achievements' in data['progression']:
        achievements = data['progression']['achievements']
        if isinstance(achievements, list):
            for achievement in achievements:
                if isinstance(achievement, dict) and 'requirement' in achievement:
                    req = achievement['requirement']
                    # Replace 'age' requirement with 'affection'
                    if 'age' in req:
                        req['affection'] = req.pop('age')
                        # Adjust the value to be reasonable for affection (0-100)
                        if 'min' in req['affection'] and req['affection']['min'] > 100:
                            req['affection']['min'] = 80
    return data

def fix_bot_personality(data):
    """Fix botPersonality structure to use correct nested format"""
    if 'multiplayer' in data and isinstance(data['multiplayer'], dict):
        multiplayer = data['multiplayer']
        if 'botPersonality' in multiplayer and isinstance(multiplayer['botPersonality'], dict):
            bot_personality = multiplayer['botPersonality']
            
            # Check if it's using the old flat structure
            if 'interactionRate' in bot_personality and 'behavior' not in bot_personality:
                # Extract values from flat structure
                name = bot_personality.get('name', 'default_bot')
                interaction_rate = bot_personality.get('interactionRate', 0.5)
                
                # Create new nested structure
                new_bot_personality = {
                    'name': name,
                    'description': f'Bot personality for {name}',
                    'behavior': {
                        'responseDelay': '1-3s',
                        'interactionRate': interaction_rate,
                        'attention': 0.7,
                        'maxActionsPerMinute': 5,
                        'minTimeBetweenSame': 30
                    }
                }
                
                # Add traits if they exist
                traits = {}
                if 'chattiness' in bot_personality:
                    traits['chattiness'] = bot_personality['chattiness']
                if 'helpfulness' in bot_personality:
                    traits['helpfulness'] = bot_personality['helpfulness']
                if 'playfulness' in bot_personality:
                    traits['playfulness'] = bot_personality['playfulness']
                if 'socialness' in bot_personality:
                    traits['socialness'] = bot_personality['socialness']
                    
                if traits:
                    new_bot_personality['traits'] = traits
                
                # Replace the old structure
                multiplayer['botPersonality'] = new_bot_personality
            
            # Ensure name field exists
            elif 'name' not in bot_personality:
                if 'networkPersonality' in multiplayer:
                    bot_personality['name'] = multiplayer['networkPersonality']
                else:
                    bot_personality['name'] = 'default_bot'
    return data

def fix_missing_animations(data):
    """Remove references to animations that don't exist"""
    if 'animations' in data:
        available_animations = set(data['animations'].keys())
        
        # Fix romance events
        if 'romanceFeatures' in data and 'events' in data['romanceFeatures']:
            for event in data['romanceFeatures']['events']:
                if isinstance(event, dict) and 'choices' in event:
                    for choice in event['choices']:
                        if isinstance(choice, dict) and 'animation' in choice:
                            if choice['animation'] not in available_animations:
                                # Use a common fallback animation
                                if 'happy' in available_animations:
                                    choice['animation'] = 'happy'
                                elif 'talking' in available_animations:
                                    choice['animation'] = 'talking'
                                else:
                                    choice['animation'] = 'idle'
    return data

try:
    input_file = sys.argv[1]
    output_file = sys.argv[2]
    
    with open(input_file, 'r', encoding='utf-8') as f:
        data = json.load(f)
    
    # Apply fixes
    data = fix_general_events(data)
    data = fix_invalid_stats_in_achievements(data)
    data = fix_bot_personality(data)
    data = fix_missing_animations(data)
    
    with open(output_file, 'w', encoding='utf-8') as f:
        json.dump(data, f, indent=2, ensure_ascii=False)
        
except Exception as e:
    print(f"Error processing file: {e}", file=sys.stderr)
    sys.exit(1)
EOF
    
    # Clean up temp files
    rm -f "$temp_file" "${temp_file}.2"
    
    echo "  ✓ Fixed: $character_name"
}

# Find all character files and fix them
character_files=()
while IFS= read -r file; do
    character_files+=("$file")
done < <(find "$CHARACTERS_DIR" -mindepth 2 -maxdepth 2 -name "character.json" -type f | sort)

log "Found ${#character_files[@]} character files to fix"

for character_file in "${character_files[@]}"; do
    fix_character_file "$character_file"
done

log "Character fixes complete!"
log "Backups created with .backup extension"
log "Run './scripts/validate-characters.sh' to verify fixes"
