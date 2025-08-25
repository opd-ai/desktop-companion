# Romance Character Setup Guide

This character requires both basic and romance-specific animation files.

## Quick Setup for Testing

To quickly test the romance character with existing animations:

```bash
# Copy basic animations from default character
cp /workspaces/DDS/assets/characters/default/animations/*.gif /workspaces/DDS/assets/characters/romance/animations/

# Create placeholder romance animations (copy from basic ones for now)
cd /workspaces/DDS/assets/characters/romance/animations/
cp idle.gif blushing.gif
cp happy.gif heart_eyes.gif
cp sad.gif shy.gif
cp happy.gif flirty.gif
cp idle.gif romantic_idle.gif
cp sad.gif jealous.gif
cp happy.gif excited_romance.gif
```

## Testing the Romance Character

```bash
# Run with romance character
go run cmd/companion/main.go -game -stats -character assets/characters/romance/character.json

# Test romance interactions:
# Shift+Click: Compliment (builds affection, trust)
# Ctrl+Click: Give gift (builds affection, happiness, trust) 
# Alt+Click: Deep conversation (builds trust, affection, intimacy)
```

## Romance Features to Test

1. **Stats System**: Watch affection, trust, intimacy, jealousy in stats overlay
2. **Personality Effects**: Character responds differently based on personality traits
3. **Relationship Progression**: Character evolves from Stranger → Friend → Close Friend → Romantic Interest
4. **Romance Dialogs**: Special dialogs unlock as relationship develops
5. **Romance Events**: Random romantic thoughts and memories occur
6. **Requirements System**: Some interactions require minimum stat levels

## Expected Behavior

- **Early relationship**: Basic interactions, character is more reserved
- **As affection grows**: More romantic responses, access to gift-giving
- **High trust**: Unlocks deep conversations, more intimate interactions
- **Romantic Interest level**: Special romantic idle animation, loving dialogs

## Development Notes

This character demonstrates the JSON-based romance system from the dating simulator plan:
- All romance behavior configured via JSON (no code changes needed)
- Personality traits influence interaction effectiveness
- Progressive unlocking based on relationship stats
- Backward compatible with existing Tamagotchi features
