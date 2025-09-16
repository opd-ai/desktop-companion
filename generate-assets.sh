#!/bin/bash

# Quick asset generation script - wrapper for character asset generation
# This provides a simple interface for the most common use cases

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SIMPLE_SCRIPT="$SCRIPT_DIR/scripts/generate-character-assets-simple.sh"
ADVANCED_SCRIPT="$SCRIPT_DIR/scripts/generate-all-character-assets.sh"

echo "Desktop Companion - Character Asset Generator"
echo "============================================="
echo

# Quick usage menu
if [[ $# -eq 0 ]]; then
    echo "Quick options:"
    echo "  1. Generate all assets (default settings)"
    echo "  2. Dry run (preview what will be generated)"
    echo "  3. Verbose generation"
    echo "  4. Custom style/model"
    echo "  5. Advanced options (parallel processing)"
    echo "  6. Help"
    echo
    read -p "Choose an option (1-6): " choice
    
    case $choice in
        1)
            echo "Generating assets with default settings..."
            exec "$SIMPLE_SCRIPT"
            ;;
        2)
            echo "Dry run - showing what would be generated..."
            exec "$SIMPLE_SCRIPT" --dry-run --verbose
            ;;
        3)
            echo "Verbose generation..."
            exec "$SIMPLE_SCRIPT" --verbose
            ;;
        4)
            echo "Available styles: anime, realistic, cartoon, pixel"
            echo "Available models: sd15, sdxl, dall-e"
            read -p "Enter style (default: anime): " style
            read -p "Enter model (default: sd15): " model
            style=${style:-anime}
            model=${model:-sd15}
            echo "Generating with style=$style, model=$model..."
            exec "$SIMPLE_SCRIPT" --style "$style" --model "$model" --verbose
            ;;
        5)
            echo "Advanced options with parallel processing..."
            exec "$ADVANCED_SCRIPT" --help
            ;;
        6)
            echo "Simple script help:"
            exec "$SIMPLE_SCRIPT" --help
            ;;
        *)
            echo "Invalid choice. Running with defaults..."
            exec "$SIMPLE_SCRIPT"
            ;;
    esac
else
    # Check if user wants advanced features
    if [[ "$*" =~ --jobs ]] || [[ "$*" =~ --parallel ]]; then
        # Use advanced script for parallel processing
        exec "$ADVANCED_SCRIPT" "$@"
    else
        # Use simple script for everything else
        exec "$SIMPLE_SCRIPT" "$@"
    fi
fi
    exec "$MAIN_SCRIPT" "$@"
fi
