# Romance Character Animation Setup

This directory contains romance-specific animations for the dating simulator companion.

## Required Animation Files

### Basic Animations (inherited from default character)
- `idle.gif` - Default idle animation
- `talking.gif` - Speaking animation
- `happy.gif` - Happy/excited animation
- `sad.gif` - Sad/disappointed animation
- `hungry.gif` - Hungry animation
- `eating.gif` - Eating animation

### Romance-Specific Animations
- `blushing.gif` - Character blushing from compliments
- `heart_eyes.gif` - Character with heart eyes, very happy
- `shy.gif` - Shy/embarrassed animation
- `flirty.gif` - Flirtatious animation
- `romantic_idle.gif` - Special idle animation for romantic relationship level
- `jealous.gif` - Jealous/upset animation
- `excited_romance.gif` - Excited about romantic interaction

## Animation Requirements

- **Format**: Animated GIF with transparency
- **Size**: 64x64 to 256x256 pixels recommended
- **File size**: <1MB each for best performance
- **Frames**: 2-10 frames per animation for smooth playback
- **Background**: Transparent background recommended

## Setup Instructions

1. Copy basic animations from `../default/animations/` as a starting point
2. Create or commission romance-specific animations
3. Ensure all animation files match the names in `character.json`
4. Test with: `go run cmd/companion/main.go -game -stats -character assets/characters/romance/character.json`

## Romance Features

This character includes:
- **Romance stats**: affection, trust, intimacy, jealousy
- **Personality traits**: shyness, romanticism, flirtiness, etc.
- **Romance interactions**: compliment, give_gift, deep_conversation
- **Relationship progression**: Stranger → Friend → Close Friend → Romantic Interest
- **Romance-specific dialogs**: Context-aware responses based on relationship level
- **Romance events**: Random romantic thoughts and memories

## Personality Configuration

The character is configured as:
- Moderately shy (0.6) - responds well to patient interaction
- Highly romantic (0.8) - appreciates romantic gestures
- Low jealousy tendency (0.3) - generally trusting
- Moderate trust difficulty (0.4) - builds trust at reasonable pace
- High affection responsiveness (0.9) - responds strongly to kind treatment
- High flirtiness (0.7) - naturally flirtatious personality

## Interaction Guide

- **Shift+Click**: Compliment (requires trust ≥10)
- **Ctrl+Click**: Give gift (requires affection ≥15)
- **Alt+Click**: Deep conversation (requires trust ≥20)
- **Regular Click**: Pet/talk
- **Right-click**: Feed
- **Double-click**: Play

Watch the stats overlay to see how romance stats develop over time!
