# Slow Burn Romance Character Archetype

## Overview
The Slow Burn Romance character represents a thoughtful, reserved personality that values deep emotional connection over quick romantic progression. This character is perfect for players who enjoy meaningful relationship development and realistic pacing.

## Personality Profile

### Core Traits
- **Shyness**: 0.7 (Reserved and thoughtful)
- **Romanticism**: 0.6 (Moderately romantic but sincere)
- **Jealousy Prone**: 0.1 (Very low jealousy, secure personality)
- **Trust Difficulty**: 0.9 (Extremely difficult to gain deep trust)
- **Affection Responsiveness**: 0.4 (Slow to respond, but meaningful)
- **Flirtiness**: 0.1 (Not flirty, prefers sincerity)

### Gameplay Characteristics

**Starting Stats:**
- Very low starting affection (0) and moderate trust (10)
- Slower stat degradation across the board
- Longer decay intervals (90 seconds vs 60)

**Interaction Style:**
- Long cooldowns emphasize quality over quantity
- Thoughtful, measured responses
- Small but meaningful stat gains

**Progression:**
- Very slow relationship progression (5 levels over ~16 days)
- High requirements emphasizing trust and time
- Dramatic payoffs for patience and consistency

## Usage Examples

```bash
# Run the slow burn romance character
go run cmd/companion/main.go -game -stats -character assets/characters/slow_burn/character.json
```

## Character Development Arc

1. **Polite Acquaintance** (0-4 days): Formal, friendly but distant
2. **Trusted Friend** (4-8 days): Growing comfort and openness
3. **Close Confidant** (8-12 days): Deep conversations and trust
4. **Emotional Partner** (12-16 days): Romantic feelings emerge
5. **Life Partner** (16+ days): Complete emotional intimacy

## Unique Features

### Special Interactions
- **Intimate Moment** (Alt+Click): Requires high trust, major intimacy gains
- **Commitment Talk** (Ctrl+Shift+Click): Serious relationship discussion
- **Deep Conversation**: Enhanced importance with 2.2x compatibility bonus

### Romance Events
- **Trust Building Moment**: Reflects on growing security
- **Vulnerability Moment**: Rare emotional openness
- **Commitment Consideration**: Thinking about the future
- **Appreciation for Patience**: Values partner's understanding

### Dialogue Evolution
Early: "Hello. It's nice to see you again."
Mid: "I find myself looking forward to seeing you..."
Late: "I want to spend my life with you."

## Strategy Tips

1. **Patience is Key**: This character rewards long-term commitment
2. **Quality over Quantity**: Fewer, more meaningful interactions
3. **Deep Conversations**: Primary relationship building tool (2.2x bonus)
4. **Consistency Matters**: Regular interaction more important than intensity
5. **Trust First**: Focus on trust building before romantic advancement

## Interaction Requirements

Many interactions have trust prerequisites:
- **Pet**: Requires trust ≥ 5
- **Play**: Requires trust ≥ 10  
- **Compliment**: Requires trust ≥ 15
- **Gift**: Requires trust ≥ 25 AND affection ≥ 15
- **Intimate Moment**: Requires trust ≥ 50, affection ≥ 40, intimacy ≥ 25

## Animation Mapping
- Subtle animation usage emphasizing "talking" and "happy"
- "romantic_idle" reserved for advanced relationship stages
- Minimal dramatic animations, focusing on sincerity

## Compatibility Bonuses

- **Consistent Interaction**: 2.0x multiplier (highest value)
- **Conversation Lover**: 2.2x multiplier (extremely important)
- **Gift Appreciation**: 1.3x multiplier (moderate)
- **Variety Preference**: 0.6x multiplier (prefers consistency)

## Ideal For Players Who Want

- **Realistic Relationship Pacing**: Mirrors real-world relationship development
- **Meaningful Progression**: Every advancement feels earned
- **Deep Emotional Content**: Focus on trust and emotional intimacy
- **Long-term Commitment**: Designed for extended play sessions
- **Character Development**: Gradual personality evolution

## Performance Notes

- Slower stat degradation reduces maintenance requirements
- Long cooldowns prevent rapid stat building
- Events are rare but highly meaningful
- Character progression requires real-time investment

## Achievement Philosophy

Achievements focus on patience and consistency:
- **First Trust**: Building initial foundation
- **Deep Connection**: Balanced trust and affection
- **Soulmate Bond**: Maximum relationship depth requiring significant time investment

This character archetype is perfect for players who enjoy:
- Realistic relationship development
- Long-term character investment
- Deep, meaningful dialogue
- Patient, thoughtful gameplay
- Emotional storytelling and character growth
