#!/bin/bash

# fix-remaining-validation-issues.sh
# Additional fixes for remaining validation issues

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
CHARACTERS_DIR="$PROJECT_ROOT/assets/characters"

log() {
    echo "[$(date '+%H:%M:%S')] $*"
}

# Function to fix romance dialog triggers
fix_romance_triggers() {
    local file="$1"
    local character_name="$(basename "$(dirname "$file")")"
    
    log "Fixing romance triggers in $character_name"
    
    python3 -c "
import json
import sys

file_path = '$file'
try:
    with open(file_path, 'r') as f:
        data = json.load(f)
    
    fixed = False
    valid_triggers = ['click', 'rightclick', 'hover']
    
    # Fix romance features dialogs
    if 'romanceFeatures' in data and 'dialogs' in data['romanceFeatures']:
        for dialog in data['romanceFeatures']['dialogs']:
            if isinstance(dialog, dict) and 'trigger' in dialog:
                if dialog['trigger'] not in valid_triggers:
                    old_trigger = dialog['trigger']
                    dialog['trigger'] = 'click'  # Default to click
                    print(f'    Fixed romance dialog trigger: {old_trigger} -> click')
                    fixed = True
    
    # Fix romanceDialogs (alternative structure)
    if 'romanceDialogs' in data:
        for dialog in data['romanceDialogs']:
            if isinstance(dialog, dict) and 'trigger' in dialog:
                if dialog['trigger'] not in valid_triggers:
                    old_trigger = dialog['trigger']
                    dialog['trigger'] = 'click'
                    print(f'    Fixed romance dialog trigger: {old_trigger} -> click')
                    fixed = True
    
    # Fix romance features events
    if 'romanceFeatures' in data and 'events' in data['romanceFeatures']:
        for event in data['romanceFeatures']['events']:
            if isinstance(event, dict) and 'trigger' in event:
                if event['trigger'] not in valid_triggers:
                    old_trigger = event['trigger']
                    event['trigger'] = 'click'
                    print(f'    Fixed romance event trigger: {old_trigger} -> click')
                    fixed = True
    
    if fixed:
        with open(file_path, 'w') as f:
            json.dump(data, f, indent=2)
        print(f'    ✅ Fixed romance triggers in {file_path}')
    else:
        print(f'    ✅ Romance triggers already valid')
        
except Exception as e:
    print(f'    ❌ Error processing {file_path}: {e}')
    sys.exit(1)
"
}

# Function to fix description length
fix_description_length() {
    local file="$1"
    local character_name="$(basename "$(dirname "$file")")"
    
    log "Fixing description length in $character_name"
    
    python3 -c "
import json
import sys

file_path = '$file'
try:
    with open(file_path, 'r') as f:
        data = json.load(f)
    
    fixed = False
    
    # Fix main description
    if 'description' in data:
        desc = data['description']
        if len(desc) > 200:
            # Truncate to 197 chars and add ellipsis
            data['description'] = desc[:197] + '...'
            print(f'    Truncated description from {len(desc)} to 200 characters')
            fixed = True
    
    if fixed:
        with open(file_path, 'w') as f:
            json.dump(data, f, indent=2)
        print(f'    ✅ Fixed description in {file_path}')
    else:
        print(f'    ✅ Description length already valid')
        
except Exception as e:
    print(f'    ❌ Error processing {file_path}: {e}')
    sys.exit(1)
"
}

# Function to fix missing animation references
fix_missing_animation_references() {
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
    
    # Add commonly referenced animations if missing
    common_animations = {
        'heart_eyes': 'animations/heart_eyes.gif',
        'blushing': 'animations/blushing.gif',
        'shy': 'animations/shy.gif',
        'excited': 'animations/excited.gif',
        'thinking': 'animations/thinking.gif',
        'winking': 'animations/winking.gif',
        'surprised': 'animations/surprised.gif'
    }
    
    # Check and add missing animations referenced in the character
    file_content = json.dumps(data)
    for anim_name, anim_path in common_animations.items():
        if anim_name in file_content and anim_name not in available_animations:
            data['animations'][anim_name] = anim_path
            available_animations.add(anim_name)
            print(f'    Added missing animation: {anim_name}')
            fixed = True
    
    # Fix romance features that reference missing animations
    if 'romanceFeatures' in data:
        # Fix romance events
        if 'events' in data['romanceFeatures']:
            for event in data['romanceFeatures']['events']:
                if isinstance(event, dict):
                    # Fix animation references in choices
                    if 'choices' in event:
                        for choice in event['choices']:
                            if isinstance(choice, dict) and 'animation' in choice:
                                if choice['animation'] not in available_animations:
                                    old_anim = choice['animation']
                                    # Try to find a suitable replacement
                                    if 'happy' in available_animations:
                                        choice['animation'] = 'happy'
                                    elif 'talking' in available_animations:
                                        choice['animation'] = 'talking'
                                    else:
                                        choice['animation'] = 'idle'
                                    print(f'    Fixed animation reference: {old_anim} -> {choice[\"animation\"]}')
                                    fixed = True
                    
                    # Fix animation references in the event itself
                    if 'animation' in event and event['animation'] not in available_animations:
                        old_anim = event['animation']
                        if 'happy' in available_animations:
                            event['animation'] = 'happy'
                        elif 'talking' in available_animations:
                            event['animation'] = 'talking'
                        else:
                            event['animation'] = 'idle'
                        print(f'    Fixed event animation: {old_anim} -> {event[\"animation\"]}')
                        fixed = True
    
    # Fix general events that reference missing animations
    if 'generalEvents' in data:
        for event in data['generalEvents']:
            if isinstance(event, dict) and 'animation' in event:
                if event['animation'] not in available_animations:
                    old_anim = event['animation']
                    if 'talking' in available_animations:
                        event['animation'] = 'talking'
                    else:
                        event['animation'] = 'idle'
                    print(f'    Fixed general event animation: {old_anim} -> {event[\"animation\"]}')
                    fixed = True
    
    if fixed:
        with open(file_path, 'w') as f:
            json.dump(data, f, indent=2)
        print(f'    ✅ Fixed animation references in {file_path}')
    else:
        print(f'    ✅ Animation references already valid')
        
except Exception as e:
    print(f'    ❌ Error processing {file_path}: {e}')
    sys.exit(1)
"
}

# Function to fix game features triggers
fix_game_features_triggers() {
    local file="$1"
    local character_name="$(basename "$(dirname "$file")")"
    
    log "Fixing game features triggers in $character_name"
    
    python3 -c "
import json
import sys

file_path = '$file'
try:
    with open(file_path, 'r') as f:
        data = json.load(f)
    
    fixed = False
    valid_triggers = ['click', 'rightclick', 'doubleclick', 'shift+click', 'hover', 'ctrl+shift+click', 'alt+shift+click', 'daily_interaction_bonus']
    
    # Fix game features interactions
    if 'gameFeatures' in data and 'interactions' in data['gameFeatures']:
        for interaction in data['gameFeatures']['interactions']:
            if isinstance(interaction, dict) and 'trigger' in interaction:
                if interaction['trigger'] not in valid_triggers:
                    old_trigger = interaction['trigger']
                    # Map common invalid triggers to valid ones
                    if 'ctrl+click' in old_trigger:
                        interaction['trigger'] = 'ctrl+shift+click'
                    elif 'alt+click' in old_trigger:
                        interaction['trigger'] = 'alt+shift+click'
                    else:
                        interaction['trigger'] = 'click'
                    print(f'    Fixed game features trigger: {old_trigger} -> {interaction[\"trigger\"]}')
                    fixed = True
    
    if fixed:
        with open(file_path, 'w') as f:
            json.dump(data, f, indent=2)
        print(f'    ✅ Fixed game features triggers in {file_path}')
    else:
        print(f'    ✅ Game features triggers already valid')
        
except Exception as e:
    print(f'    ❌ Error processing {file_path}: {e}')
    sys.exit(1)
"
}

# Main processing loop
log "Starting additional validation fixes..."

# Find all character files that still fail validation
failed_characters=(
    "challenge" "easy" "flirty" "hard" "llm_example" "markov_example" 
    "news_example" "normal" "romance" "romance_flirty" "romance_slowburn" 
    "romance_supportive" "romance_tsundere" "slow_burn" "specialist"
)

for character_name in "${failed_characters[@]}"; do
    character_file="$CHARACTERS_DIR/$character_name/character.json"
    
    if [[ -f "$character_file" ]]; then
        echo
        log "Processing character: $character_name"
        
        # Apply additional fixes
        fix_description_length "$character_file"
        fix_romance_triggers "$character_file"
        fix_missing_animation_references "$character_file"
        fix_game_features_triggers "$character_file"
        
        echo "  ✅ $character_name additional fixes complete"
    fi
done

echo
log "Additional character fixes applied!"
log "Running validation check..."

# Run validation to see results
"$PROJECT_ROOT/scripts/validate-characters.sh"
