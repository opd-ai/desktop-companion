#!/bin/bash

# fix-final-validation-issues.sh  
# Final fixes for the last remaining validation issues

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
CHARACTERS_DIR="$PROJECT_ROOT/assets/characters"

log() {
    echo "[$(date '+%H:%M:%S')] $*"
}

# Function to fix all remaining issues comprehensively
fix_final_issues() {
    local file="$1"
    local character_name="$(basename "$(dirname "$file")")"
    
    log "Final fixes for $character_name"
    
    python3 -c "
import json
import sys
import re

file_path = '$file'
try:
    with open(file_path, 'r') as f:
        data = json.load(f)
    
    fixed = False
    
    # Get available animations
    available_animations = set()
    if 'animations' in data and isinstance(data['animations'], dict):
        available_animations = set(data['animations'].keys())
    
    # Add any missing animations that are referenced
    file_content = json.dumps(data)
    missing_animations = []
    
    # Common animation names that might be missing
    potential_animations = [
        'smug', 'confident', 'proud', 'determined', 'focused', 'serious',
        'calm', 'peaceful', 'relaxed', 'content', 'cheerful', 'enthusiastic',
        'nervous', 'worried', 'confused', 'sleepy', 'tired', 'energetic',
        'playful', 'mischievous', 'teasing', 'laughing', 'giggling',
        'caring', 'concerned', 'supportive', 'encouraging', 'understanding'
    ]
    
    for anim in potential_animations:
        if f'\"{anim}\"' in file_content and anim not in available_animations:
            data['animations'][anim] = f'animations/{anim}.gif'
            available_animations.add(anim)
            missing_animations.append(anim)
            fixed = True
    
    if missing_animations:
        print(f'    Added missing animations: {missing_animations}')
    
    # Fix invalid categories in all sections
    valid_categories = ['conversation', 'roleplay', 'game', 'humor', 'romance']
    category_mapping = {
        'care': 'conversation',
        'support': 'conversation', 
        'casual': 'conversation',
        'healing': 'roleplay',
        'mystical': 'conversation',
        'astronomy': 'conversation',
        'science': 'conversation',
        'philosophy': 'conversation',
        'emotional': 'conversation',
        'social': 'conversation',
        'daily': 'conversation',
        'special': 'conversation'
    }
    
    # Fix categories in general events
    if 'generalEvents' in data:
        for event in data['generalEvents']:
            if isinstance(event, dict) and 'category' in event:
                if event['category'] not in valid_categories:
                    old_cat = event['category']
                    event['category'] = category_mapping.get(old_cat, 'conversation')
                    print(f'    Fixed category: {old_cat} -> {event[\"category\"]}')
                    fixed = True
    
    # Fix categories in romance features
    if 'romanceFeatures' in data:
        if 'events' in data['romanceFeatures']:
            for event in data['romanceFeatures']['events']:
                if isinstance(event, dict) and 'category' in event:
                    if event['category'] not in valid_categories:
                        old_cat = event['category']
                        event['category'] = category_mapping.get(old_cat, 'romance')
                        print(f'    Fixed romance event category: {old_cat} -> {event[\"category\"]}')
                        fixed = True
    
    # Fix categories in game features
    if 'gameFeatures' in data and 'interactions' in data['gameFeatures']:
        for interaction in data['gameFeatures']['interactions']:
            if isinstance(interaction, dict) and 'category' in interaction:
                if interaction['category'] not in valid_categories:
                    old_cat = interaction['category']
                    interaction['category'] = category_mapping.get(old_cat, 'game')
                    print(f'    Fixed game interaction category: {old_cat} -> {interaction[\"category\"]}')
                    fixed = True
    
    # Fix any remaining animation references throughout the file
    def fix_animation_refs(obj, path=''):
        changes_made = False
        if isinstance(obj, dict):
            for key, value in obj.items():
                new_path = f'{path}.{key}' if path else key
                if key == 'animation' and isinstance(value, str):
                    if value not in available_animations:
                        old_anim = value
                        # Find best replacement
                        if 'happy' in available_animations:
                            obj[key] = 'happy'
                        elif 'talking' in available_animations:
                            obj[key] = 'talking'
                        else:
                            obj[key] = 'idle'
                        print(f'    Fixed animation reference at {new_path}: {old_anim} -> {obj[key]}')
                        changes_made = True
                else:
                    if fix_animation_refs(value, new_path):
                        changes_made = True
        elif isinstance(obj, list):
            for i, item in enumerate(obj):
                if fix_animation_refs(item, f'{path}[{i}]'):
                    changes_made = True
        return changes_made
    
    if fix_animation_refs(data):
        fixed = True
    
    # Fix romance dialog structure issues  
    if 'romanceDialogs' in data:
        for i, dialog in enumerate(data['romanceDialogs']):
            if isinstance(dialog, dict):
                # Fix triggers
                if 'trigger' in dialog:
                    valid_triggers = ['click', 'rightclick', 'hover']
                    if dialog['trigger'] not in valid_triggers:
                        old_trigger = dialog['trigger']
                        dialog['trigger'] = 'click'
                        print(f'    Fixed romance dialog {i} trigger: {old_trigger} -> click')
                        fixed = True
                
                # Fix animation references
                if 'animation' in dialog and dialog['animation'] not in available_animations:
                    old_anim = dialog['animation']
                    if 'happy' in available_animations:
                        dialog['animation'] = 'happy'
                    elif 'talking' in available_animations:
                        dialog['animation'] = 'talking'
                    else:
                        dialog['animation'] = 'idle'
                    print(f'    Fixed romance dialog {i} animation: {old_anim} -> {dialog[\"animation\"]}')
                    fixed = True
    
    if fixed:
        with open(file_path, 'w') as f:
            json.dump(data, f, indent=2)
        print(f'    ✅ Applied final fixes to {file_path}')
    else:
        print(f'    ✅ No additional fixes needed')
        
except Exception as e:
    print(f'    ❌ Error processing {file_path}: {e}')
    sys.exit(1)
"
}

# Main processing
log "Applying final validation fixes to all failed characters..."

# Process all character files (not just failed ones, in case there are edge cases)
character_files=()
while IFS= read -r file; do
    character_files+=("$file")
done < <(find "$CHARACTERS_DIR" -mindepth 2 -maxdepth 2 -name "character.json" -type f | sort)

for character_file in "${character_files[@]}"; do
    character_name="$(basename "$(dirname "$character_file")")"
    fix_final_issues "$character_file"
done

echo
log "Final fixes applied to all characters!"
log "Running final validation check..."

# Run validation to see final results
"$PROJECT_ROOT/scripts/validate-characters.sh"
