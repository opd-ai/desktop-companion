#!/bin/bash

# Animation validation script for character directories
# Ensures all required animation files exist for each character

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
CHARACTERS_DIR="$PROJECT_ROOT/assets/characters"

# Basic animations required for all characters
BASIC_ANIMATIONS=(
    "idle.gif"
    "talking.gif" 
    "happy.gif"
    "sad.gif"
    "hungry.gif"
    "eating.gif"
)

# Character-specific animations (optional)
declare -A SPECIFIC_ANIMATIONS
SPECIFIC_ANIMATIONS[romance]="blushing.gif heart_eyes.gif shy.gif flirty.gif romantic_idle.gif"

echo "Validating character animations..."

exit_code=0

# Check each character directory
for char_dir in "$CHARACTERS_DIR"/*/; do
    if [ ! -d "$char_dir" ]; then
        continue
    fi
    
    char_name=$(basename "$char_dir")
    animations_dir="$char_dir/animations"
    
    echo "Checking character: $char_name"
    
    if [ ! -d "$animations_dir" ]; then
        echo "  ❌ Missing animations directory"
        exit_code=1
        continue
    fi
    
    # Check basic animations
    missing_basic=0
    for animation in "${BASIC_ANIMATIONS[@]}"; do
        if [ ! -f "$animations_dir/$animation" ]; then
            echo "  ❌ Missing basic animation: $animation"
            missing_basic=1
            exit_code=1
        fi
    done
    
    if [ $missing_basic -eq 0 ]; then
        echo "  ✅ All basic animations present"
    fi
    
    # Check character-specific animations
    if [ -n "${SPECIFIC_ANIMATIONS[$char_name]}" ]; then
        missing_specific=0
        for animation in ${SPECIFIC_ANIMATIONS[$char_name]}; do
            if [ ! -f "$animations_dir/$animation" ]; then
                echo "  ⚠️  Missing specific animation: $animation"
                missing_specific=1
            fi
        done
        
        if [ $missing_specific -eq 0 ]; then
            echo "  ✅ All specific animations present"
        fi
    fi
done

if [ $exit_code -eq 0 ]; then
    echo "✅ Animation validation passed!"
else
    echo "❌ Animation validation failed - some animations are missing"
fi

exit $exit_code
