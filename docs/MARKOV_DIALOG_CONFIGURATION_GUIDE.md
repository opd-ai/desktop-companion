# Markov Chain Dialog Configuration Guide

## Table of Contents
- [Quick Start](#quick-start)
- [Configuration Templates](#configuration-templates)
- [Advanced Features](#advanced-features)
- [Quality Control](#quality-control)
- [Troubleshooting](#troubleshooting)
- [Best Practices](#best-practices)

## Quick Start

### Enabling Markov Dialog for Your Character

Add this to your character's JSON configuration:

```json
{
  "name": "Your Character",
  "description": "A character with AI-powered dialog",
  
  "dialogBackend": {
    "enabled": true,
    "defaultBackend": "markov_chain",
    "confidenceThreshold": 0.6,
    "backends": {
      "markov_chain": {
        "$include": "templates/markov_basic.json"
      }
    }
  }
}
```

### Using Templates

Templates provide pre-configured settings for common character types:

- `markov_basic.json` - General-purpose friendly companion
- `markov_romance.json` - Romantic companion with affectionate responses
- `markov_shy.json` - Nervous, gentle character
- `markov_tsundere.json` - Contradictory personality with hidden kindness
- `markov_intellectual.json` - Sophisticated, philosophical character

## Configuration Templates

### Basic Character
Perfect for simple desktop pets with friendly dialog:

```json
{
  "chainOrder": 2,
  "minWords": 3,
  "maxWords": 12,
  "temperatureMin": 0.4,
  "temperatureMax": 0.7,
  "useDialogHistory": true,
  "usePersonality": false,
  "triggerSpecific": false
}
```

**Best for**: General companions, helpful assistants, friendly pets

### Romance Character
Designed for dating simulator mechanics with emotional depth:

```json
{
  "chainOrder": 2,
  "minWords": 4,
  "maxWords": 18,
  "temperatureMin": 0.3,
  "temperatureMax": 0.8,
  "useDialogHistory": true,
  "usePersonality": true,
  "triggerSpecific": true,
  "personalityBoost": 0.5,
  "relationshipWeight": 0.8
}
```

**Best for**: Dating simulators, romantic companions, emotional support characters

### Shy Character
For introverted, gentle personality types:

```json
{
  "chainOrder": 2,
  "minWords": 2,
  "maxWords": 8,
  "temperatureMin": 0.6,
  "temperatureMax": 1.0,
  "moodInfluence": 0.7,
  "personalityBoost": 0.3
}
```

**Best for**: Introverted characters, nervous personalities, gentle souls

### Tsundere Character
For contradictory personalities with hidden kindness:

```json
{
  "chainOrder": 2,
  "minWords": 3,
  "maxWords": 15,
  "triggerSpecific": true,
  "personalityBoost": 0.6,
  "forbiddenWords": ["love", "adore", "precious"],
  "requiredWords": ["hmph", "idiot", "whatever"]
}
```

**Best for**: Tsundere archetypes, contradictory personalities, complex characters

### Intellectual Character
For sophisticated, philosophical characters:

```json
{
  "chainOrder": 3,
  "minWords": 5,
  "maxWords": 25,
  "temperatureMin": 0.2,
  "temperatureMax": 0.9,
  "personalityBoost": 0.7,
  "statAwareness": 0.6,
  "requiredWords": ["consciousness", "fascinating", "analyze"]
}
```

**Best for**: Academic characters, philosophers, sophisticated AI personalities

## Advanced Features

### Personality Integration

Enable personality-aware responses by setting `usePersonality: true` and adjusting:

- `personalityBoost` (0-2): How much personality traits affect word selection
- `moodInfluence` (0-2): How much character mood affects response style
- `statAwareness` (0-1): How much character stats influence generation

### Context Awareness

Make responses context-sensitive:

- `triggerSpecific: true` - Different responses for different triggers (click, hover, gift)
- `relationshipWeight` (0-2) - Adapt responses based on relationship level
- `timeOfDayWeight` (0-1) - Optional time-based response variation

### Memory and Learning

Enable adaptive behavior:

- `memoryDecay` (0-1): How quickly old training data is forgotten
- `learningRate` (0-1): How quickly to adapt to new interactions
- `adaptationSteps`: Number of interactions before adaptation

### Training Data

Provide character-specific training phrases:

```json
{
  "trainingData": [
    "Your character's typical responses here",
    "Include personality-appropriate language",
    "Mix different emotional tones and contexts"
  ]
}
```

## Quality Control

### Basic Quality Settings

- `coherenceThreshold` (0-1): Minimum coherence for accepting responses
- `similarityPenalty` (0-1): Penalty for responses too similar to recent ones
- `forbiddenWords`: Words to avoid in responses
- `requiredWords`: Words that should appear more often

### Advanced Quality Filters

```json
{
  "qualityFilters": {
    "minCoherence": 0.7,
    "maxRepetition": 0.3,
    "requireComplete": true,
    "grammarCheck": true,
    "minUniqueWords": 3,
    "maxSimilarity": 0.5
  }
}
```

- `minCoherence`: Enhanced coherence analysis
- `maxRepetition`: Maximum word repetition ratio
- `requireComplete`: Require complete sentences
- `grammarCheck`: Basic grammar validation
- `minUniqueWords`: Minimum unique words in response
- `maxSimilarity`: Maximum similarity to recent responses

### Fallback System

Provide high-quality fallback responses:

```json
{
  "fallbackPhrases": [
    "Thanks for being here with me.",
    "You always know how to make me smile.",
    "I'm grateful for your company."
  ]
}
```

## Parameter Reference

### Core Settings

| Parameter | Range | Default | Description |
|-----------|-------|---------|-------------|
| `chainOrder` | 1-5 | 2 | N-gram complexity (2=bigram, 3=trigram) |
| `minWords` | 1+ | 3 | Minimum words in response |
| `maxWords` | 1-50 | 12 | Maximum words in response |
| `temperatureMin` | 0-2 | 0.4 | Minimum randomness |
| `temperatureMax` | 0-2 | 0.7 | Maximum randomness |

### Personality Settings

| Parameter | Range | Default | Description |
|-----------|-------|---------|-------------|
| `usePersonality` | boolean | false | Enable personality integration |
| `personalityBoost` | 0-2 | 0.0 | Personality influence strength |
| `moodInfluence` | 0-2 | 0.0 | Mood affects generation |
| `statAwareness` | 0-1 | 0.0 | Character stats influence |

### Context Settings

| Parameter | Range | Default | Description |
|-----------|-------|---------|-------------|
| `triggerSpecific` | boolean | false | Separate chains per trigger |
| `relationshipWeight` | 0-2 | 0.0 | Relationship level impact |
| `timeOfDayWeight` | 0-1 | 0.0 | Time-based variation |

## Troubleshooting

### Common Issues

**Problem**: Responses are too random/incoherent
**Solution**: Lower `temperatureMax`, increase `coherenceThreshold`

**Problem**: Responses are too repetitive
**Solution**: Increase `temperatureMin`, add more training data

**Problem**: Character doesn't match personality
**Solution**: Enable `usePersonality`, adjust `personalityBoost`, customize training data

**Problem**: Responses are inappropriate
**Solution**: Add `forbiddenWords`, increase quality filters

**Problem**: Generation fails frequently
**Solution**: Lower `confidenceThreshold`, add fallback phrases

### Debug Mode

Enable debug logging in your character configuration:
```json
{
  "dialogBackend": {
    "debug": true
  }
}
```

### Validation

Use the validation tool to check your configuration:
```bash
go run tools/validate_characters.go path/to/your/character.json
```

## Best Practices

### Training Data Guidelines

1. **Variety**: Include 10-20 diverse training phrases
2. **Personality**: Match your character's personality and speech patterns
3. **Context**: Include responses for different situations
4. **Quality**: Use well-formed, character-appropriate sentences

### Performance Tips

1. **Chain Order**: Use 2 for most characters, 3 for sophisticated ones
2. **Temperature**: Start with 0.4-0.7, adjust based on desired randomness
3. **Training Size**: 15-25 phrases optimal for most characters
4. **Quality Filters**: Enable for higher-quality responses

### Character Design

1. **Consistency**: Ensure training data matches character personality
2. **Progression**: Consider how responses should evolve with relationship
3. **Fallbacks**: Always provide high-quality fallback phrases
4. **Testing**: Test with different triggers and relationship levels

### Configuration Examples

#### Simple Pet Character
```json
{
  "dialogBackend": {
    "enabled": true,
    "defaultBackend": "markov_chain",
    "backends": {
      "markov_chain": {
        "$include": "templates/markov_basic.json",
        "trainingData": [
          "Woof! I'm happy to see you!",
          "Pet me! I love attention!",
          "Let's play together!"
        ]
      }
    }
  }
}
```

#### Sophisticated AI Assistant
```json
{
  "dialogBackend": {
    "enabled": true,
    "defaultBackend": "markov_chain",
    "backends": {
      "markov_chain": {
        "$include": "templates/markov_intellectual.json",
        "trainingData": [
          "I find our conversations intellectually stimulating.",
          "Your perspective adds valuable insights to my understanding.",
          "The complexity of human interaction fascinates me greatly."
        ]
      }
    }
  }
}
```

#### Romantic Partner
```json
{
  "dialogBackend": {
    "enabled": true,
    "defaultBackend": "markov_chain",
    "confidenceThreshold": 0.7,
    "backends": {
      "markov_chain": {
        "$include": "templates/markov_romance.json",
        "trainingData": [
          "Every moment with you feels like magic, my love.",
          "Your smile brightens my entire world completely.",
          "I treasure these precious moments we share together."
        ]
      }
    }
  }
}
```

## Advanced Customization

### Custom Backend Chain

For ultimate control, override template settings:

```json
{
  "dialogBackend": {
    "backends": {
      "markov_chain": {
        "chainOrder": 3,
        "temperatureMin": 0.2,
        "temperatureMax": 0.9,
        "usePersonality": true,
        "personalityBoost": 0.8,
        "qualityFilters": {
          "minCoherence": 0.8,
          "grammarCheck": true,
          "requireComplete": true
        },
        "trainingData": [
          "Your completely custom training data here"
        ]
      }
    }
  }
}
```

### Multiple Backends

Use multiple backends with fallback:

```json
{
  "dialogBackend": {
    "defaultBackend": "markov_chain",
    "fallbackChain": ["simple_random"],
    "backends": {
      "markov_chain": { /* markov config */ },
      "simple_random": { /* simple config */ }
    }
  }
}
```

This guide should help you create engaging, personality-rich characters with sophisticated AI-powered dialog generation!
