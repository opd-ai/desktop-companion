# Aria Luna Character Assets

This directory contains the assets for Aria Luna, a mystical anime character designed to demonstrate the new assetGeneration features of the Desktop Companion system.

## Character Description

Aria Luna is a celestial-themed anime character with silver hair and purple eyes. She embodies mystical wisdom and gentle companionship, making her perfect for desktop interaction.

## Asset Generation Features Demonstrated

This character showcases:

### 1. Comprehensive Base Prompt
- Detailed character description optimized for SDXL/Flux.1d
- Consistent visual style across all animations
- Mystical/celestial theme integration

### 2. Advanced Animation Mappings
- **idle**: Serene default state with flowing hair and twinkling stars
- **talking**: Interactive speaking pose with magical gestures
- **happy**: Joyful expression with magical sparkles
- **sad**: Melancholic but beautiful downcast pose
- **blushing**: Romantic shy expression with warm lighting
- **heart_eyes**: Love-struck state with heart effects
- **magical**: Spell-casting with dramatic effects (custom high-quality settings)
- **sleeping**: Peaceful resting pose with dream particles

### 3. Model and Quality Settings
- Uses Flux.1d model for high-quality generation
- Anime art style with optimized parameters
- 128x128 resolution for desktop compatibility
- Custom quality settings for magical animations

### 4. ComfyUI Integration
- Custom workflow template specification
- Advanced ControlNet configuration for pose consistency
- Batch processing settings
- Custom nodes for enhanced animation generation

### 5. Validation and Backup
- Comprehensive quality validation
- Automatic asset backup before regeneration
- Generation history tracking
- File hash verification

## Usage Example

To regenerate assets for this character:

```bash
# Generate all animations using the assetGeneration configuration
gif-generator --file assets/characters/aria_luna/character.json --model flux1d

# Generate specific animation states only
gif-generator --file assets/characters/aria_luna/character.json --states idle,talking,magical

# Generate with validation and backup
gif-generator --file assets/characters/aria_luna/character.json --validate --backup

# Dry run to see what would be generated
gif-generator --file assets/characters/aria_luna/character.json --dry-run
```

## Asset Generation Metadata

The character.json includes comprehensive metadata tracking:
- Generation timestamps and settings
- File hashes for change detection
- Quality validation scores
- Processing duration history

## Backward Compatibility

This character maintains full backward compatibility:
- All existing animations work without assetGeneration
- Manual asset workflows remain supported
- Standard character.json features function normally

The assetGeneration section is purely additive and optional.
