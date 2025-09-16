# Character JSON Schema Documentation

Complete reference for creating custom characters in the Desktop Dating Simulator.

## Table of Contents

1. [Basic Structure](#basic-structure)
2. [Core Properties](#core-properties)
3. [Animation System](#animation-system)
4. [Dialog System](#dialog-system)
5. [Game Features](#game-features)
6. [Romance Features](#romance-features)
7. [Random Events](#random-events)
8. [Validation Rules](#validation-rules)
9. [Examples](#examples)

---

## Basic Structure

Every character JSON file must have this basic structure:

```json
{
  "name": "Character Name",
  "description": "Brief character description",
  "animations": { /* animation definitions */ },
  "dialogs": [ /* dialog interactions */ ],
  "behavior": { /* character behavior settings */ },
  "stats": { /* game stats configuration */ },
  "gameRules": { /* game mechanics settings */ },
  "interactions": { /* game interactions */ },
  "progression": { /* age and evolution */ },
  "randomEvents": [ /* game random events */ ],
  "personality": { /* romance personality traits */ },
  "romanceDialogs": [ /* romance-specific dialogs */ ],
  "romanceEvents": [ /* romance random events */ ],
  "dialogBackend": { /* AI dialog configuration */ },
  "generalEvents": [ /* interactive dialog events */ },
  "giftSystem": { /* gift system configuration */ },
  "multiplayer": { /* networking configuration */ },
  "battleSystem": { /* combat system settings */ },
  "newsFeatures": { /* RSS/news integration */ },
  "platformConfig": { /* platform-specific settings */ }
}
```

---

## Core Properties

### Required Fields

- **`name`** (string): Character's display name
- **`description`** (string): Brief character description  
- **`animations`** (object): Animation file mappings

### Optional Fields

- **`dialogs`** (array): Interactive dialog options
- **`behavior`** (object): Character behavior settings
- **`stats`** (object): Game stats configuration (hunger, happiness, health, energy)
- **`gameRules`** (object): Game mechanics settings (decay intervals, auto-save, etc.)
- **`interactions`** (object): Game interactions (feed, play, pet)
- **`progression`** (object): Age-based evolution configuration
- **`randomEvents`** (array): Game random events
- **`personality`** (object): Romance personality traits and preferences
- **`romanceDialogs`** (array): Romance-specific dialog interactions
- **`romanceEvents`** (array): Romance random events
- **`dialogBackend`** (object): AI-powered dialog configuration
- **`generalEvents`** (array): Interactive dialog event scenarios
- **`giftSystem`** (object): Gift system configuration
- **`multiplayer`** (object): Networking and multiplayer settings
- **`battleSystem`** (object): Combat system configuration
- **`newsFeatures`** (object): RSS/Atom news integration settings
- **`platformConfig`** (object): Platform-specific behavior overrides

---

## Animation System

Animations are GIF files that provide visual feedback for character states.

### Required Animations

Every character must have these core animations:

```json
{
  "animations": {
    "idle": "animations/idle.gif",
    "talking": "animations/talking.gif", 
    "happy": "animations/happy.gif",
    "sad": "animations/sad.gif",
    "hungry": "animations/hungry.gif",
    "eating": "animations/eating.gif"
  }
}
```

### Romance Animations

Additional animations for romance features:

```json
{
  "animations": {
    "blushing": "animations/sad.gif",
    "heart_eyes": "animations/happy.gif", 
    "shy": "animations/sad.gif",
    "flirty": "animations/happy.gif",
    "romantic_idle": "animations/idle.gif",
    "jealous": "animations/sad.gif",
    "excited_romance": "animations/happy.gif"
  }
}
```

### Path Resolution

- Paths are relative to the character JSON file location
- Use forward slashes `/` for cross-platform compatibility
- Ensure GIF files exist at specified paths

---

## Dialog System

Dialogs provide character personality through text responses to user interactions.

### Basic Dialog Structure

```json
{
  "dialogs": [
    {
      "trigger": "click",
      "responses": [
        "Hello there!",
        "How are you today?",
        "Nice to see you!"
      ],
      "animation": "talking",
      "cooldown": 5
    }
  ]
}
```

### Dialog Properties

- **`trigger`** (string): User interaction that triggers dialog
  - Valid triggers: `click`, `rightclick`, `hover`, `doubleclick`, `shift+click`, `ctrl+shift+click`, `alt+shift+click`
- **`responses`** (array): List of possible text responses (random selection)
- **`animation`** (string): Animation to play during dialog
- **`cooldown`** (integer): Seconds before dialog can trigger again (0-300)

### Romance Dialogs

Romance dialogs include requirement conditions:

```json
{
  "dialogs": [
    {
      "trigger": "hover",
      "responses": ["You make my heart flutter... 💕"],
      "animation": "blushing",
      "cooldown": 10,
      "requirements": {
        "affection": {"min": 30},
        "trust": {"min": 20}
      }
    }
  ]
}
```

---

## Game Features

Game features add interactive mechanics with stats, progression, and gameplay elements.

### Basic Game Features Structure

```json
{
  "game_features": {
    "stats": {
      "hunger": {"max": 100, "initial": 80, "degradation_rate": 0.5}
    },
    "game_rules": {
      "decay_interval": 300,
      "low_hunger_threshold": 20,
      "critical_hunger_threshold": 5
    },
    "interactions": {
      "feed": {
        "name": "Feed",
        "triggers": ["click"],
        "animation": "eating",
        "effects": {"hunger": 20},
        "cooldown": 30
      }
    }
  }
}
```

### Stats Configuration

Each stat has these properties:

- **`max`** (integer): Maximum value (1-100)
- **`initial`** (integer): Starting value (0-max)  
- **`degradation_rate`** (float): Decay per interval (0.0-5.0)

### Game Rules

Global rules affecting all stats:

- **`decay_interval`** (integer): Seconds between stat decay (60-3600)
- **`low_[stat]_threshold`** (integer): Warning level
- **`critical_[stat]_threshold`** (integer): Crisis level

### Interactions

Interactive elements with effects:

- **`name`** (string): Display name
- **`triggers`** (array): User inputs that activate interaction
- **`animation`** (string): Animation during interaction
- **`effects`** (object): Stat changes `{"stat_name": change_amount}`
- **`cooldown`** (integer): Seconds between uses (0-3600)
- **`requirements`** (object): Stat conditions to unlock

---

## Romance Features

Romance features add personality traits, relationships, and romantic interactions.

### Personality Configuration

```json
{
  "romance_features": {
    "personality": {
      "shyness": 0.7,
      "romanticism": 0.8,
      "jealousy_sensitivity": 0.6,
      "trust_difficulty": 0.4
    },
    "compatibility_modifiers": {
      "compliment": 1.2,
      "gift": 0.9,
      "conversation": 1.5
    }
  }
}
```

### Personality Traits

All values range from 0.0 to 1.0:

- **`shyness`**: How reserved the character is (higher = more shy)
- **`romanticism`**: Romantic responsiveness (higher = more romantic)  
- **`jealousy_sensitivity`**: Jealousy trigger threshold (higher = more jealous)
- **`trust_difficulty`**: How hard trust is to build (higher = slower trust)

### Compatibility Modifiers

Multipliers for interaction effectiveness (0.5 to 2.0):

- **`compliment`**: Compliment interaction bonus
- **`gift`**: Gift giving bonus
- **`conversation`**: Deep conversation bonus

### Romance Stats

Required stats for romance features:

```json
{
  "game_features": {
    "stats": {
      "affection": {"max": 100, "initial": 0},
      "trust": {"max": 100, "initial": 20}, 
      "intimacy": {"max": 100, "initial": 0},
      "jealousy": {"max": 100, "initial": 0}
    }
  }
}
```

### Romance Interactions

Special interactions with relationship requirements:

```json
{
  "game_features": {
    "interactions": {
      "compliment": {
        "name": "Compliment",
        "triggers": ["click"],
        "animation": "blushing",
        "effects": {
          "affection": 3,
          "trust": 1
        },
        "requirements": {
          "affection": {"min": 0}
        },
        "cooldown": 60
      }
    }
  }
}
```

### Romance Dialogs

Special dialogs that trigger based on relationship level:

```json
{
  "romance_features": {
    "romance_dialogs": [
      {
        "type": "compliment",
        "responses": [
          "Thank you! That means a lot! 💕",
          "You always know what to say..."
        ],
        "requirements": {
          "affection": {"min": 15}
        }
      }
    ]
  }
}
```

### Romance Events

Special events that trigger during relationship progression:

```json
{
  "romance_features": {
    "romance_events": [
      {
        "name": "First Blush",
        "description": "Character gets flustered from attention",
        "triggers": {
          "interaction_count": {"compliment": 3}
        },
        "effects": {
          "affection": 5
        },
        "animation": "blushing",
        "responses": [
          "I... I can't help but blush when you say things like that! 😊"
        ]
      }
    ]
  }
}
```

---

## Random Events

Random events add unpredictability and dynamic character behavior.

### Event Structure

```json
{
  "random_events": [
    {
      "name": "hunger_attack",
      "description": "Character gets suddenly hungry",
      "probability": 0.1,
      "cooldown": 1800,
      "duration": 300,
      "animation": "hungry",
      "responses": [
        "I'm getting really hungry...",
        "My stomach is growling!"
      ],
      "effects": {
        "hunger": -15
      },
      "conditions": {
        "hunger": {"min": 40}
      }
    }
  ]
}
```

### Event Properties

- **`name`** (string): Unique event identifier
- **`description`** (string): Event description
- **`probability`** (float): Chance per check (0.0-1.0)
- **`cooldown`** (integer): Seconds between possible triggers (0-7200)
- **`duration`** (integer): How long event lasts (0-3600)
- **`animation`** (string): Animation during event
- **`responses`** (array): Text shown during event (max 3)
- **`effects`** (object): Stat changes caused by event
- **`conditions`** (object): Stat requirements for event to trigger

---

## Validation Rules

The system enforces these validation rules:

### String Limits
- Names: 1-50 characters
- Descriptions: 1-200 characters  
- Responses: 1-100 characters each
- Max 3 responses per dialog/event

### Numeric Ranges
- Stats: max 1-100, initial 0-max, degradation 0.0-5.0
- Personality traits: 0.0-1.0
- Compatibility modifiers: 0.5-2.0
- Probabilities: 0.0-1.0
- Cooldowns: 0-7200 seconds
- Game interaction cooldowns: 0-3600 seconds

### Required References
- Animations must reference existing files
- Stat effects must reference defined stats
- Requirements must reference defined stats

---

## Examples

### Minimal Character

```json
{
  "name": "Simple Companion",
  "description": "A basic virtual companion",
  "animations": {
    "idle": "animations/idle.gif",
    "talking": "animations/talking.gif", 
    "happy": "animations/happy.gif",
    "sad": "animations/sad.gif",
    "hungry": "animations/hungry.gif",
    "eating": "animations/eating.gif"
  },
  "dialogs": [
    {
      "trigger": "click",
      "responses": ["Hello!", "Hi there!", "Nice to see you!"],
      "animation": "talking",
      "cooldown": 5
    }
  ]
}
```

### Game-Enabled Character

```json
{
  "name": "Pet Companion",
  "description": "A virtual pet with needs",
  "animations": {
    "idle": "animations/idle.gif",
    "talking": "animations/talking.gif",
    "happy": "animations/happy.gif", 
    "sad": "animations/sad.gif",
    "hungry": "animations/hungry.gif",
    "eating": "animations/eating.gif"
  },
  "dialogs": [
    {
      "trigger": "click",
      "responses": ["Woof!", "Play with me!", "*happy noises*"],
      "animation": "happy",
      "cooldown": 3
    }
  ],
  "game_features": {
    "stats": {
      "hunger": {"max": 100, "initial": 80, "degradation_rate": 0.8},
      "happiness": {"max": 100, "initial": 60, "degradation_rate": 0.3}
    },
    "game_rules": {
      "decay_interval": 180,
      "low_hunger_threshold": 25,
      "critical_hunger_threshold": 10
    },
    "interactions": {
      "feed": {
        "name": "Feed",
        "triggers": ["click"],
        "animation": "eating", 
        "effects": {"hunger": 25, "happiness": 5},
        "cooldown": 45
      },
      "play": {
        "name": "Play",
        "triggers": ["doubleclick"],
        "animation": "happy",
        "effects": {"happiness": 15},
        "requirements": {"hunger": {"min": 30}},
        "cooldown": 120
      }
    }
  }
}
```

### Romance Character

See the complete examples in:
- `assets/characters/tsundere/character.json`
- `assets/characters/flirty/character.json` 
- `assets/characters/slow_burn/character.json`

---

## Best Practices

1. **Start Simple**: Begin with basic dialogs, add complexity gradually
2. **Test Incrementally**: Validate JSON after each major addition
3. **Balance Stats**: Ensure reasonable progression rates
4. **Meaningful Interactions**: Each interaction should feel purposeful
5. **Personality Consistency**: Align all elements with character archetype
6. **Animation Reuse**: It's fine to reuse animations for similar emotions
7. **Cooldown Balance**: Avoid too frequent or too rare interactions

---

## Troubleshooting

### Common Issues

**Animation not found**: Check file paths are correct and files exist
**JSON parse error**: Use `python3 -m json.tool file.json` to validate
**Interaction not working**: Verify trigger names match valid options
**Stats not changing**: Check stat names match exactly in effects
**Romance not progressing**: Ensure requirements are reasonable and achievable

### Validation Tool

Use the character validation tool:

```bash
go run tools/validate_characters.go assets/characters/your_character/character.json
```

This will check for common issues and provide detailed error messages.
