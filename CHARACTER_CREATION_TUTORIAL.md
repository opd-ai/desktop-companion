# Character Creation Tutorial

Step-by-step guide to creating your own custom characters for the Desktop Dating Simulator.

## Prerequisites

- Text editor (VS Code, Notepad++, etc.)
- GIF animation files  
- Basic understanding of JSON format

## Tutorial Overview

We'll create a "Bookworm Scholar" character that loves learning and reading together.

---

## Step 1: Setup Directory Structure

Create the character directory:

```
assets/characters/bookworm/
├── character.json
└── animations/
    ├── idle.gif
    ├── talking.gif
    ├── happy.gif
    ├── sad.gif
    ├── hungry.gif
    ├── eating.gif
    └── reading.gif (custom animation)
```

You can copy animations from `assets/characters/default/animations/` to start.

---

## Step 2: Basic Character Structure

Create `character.json` with basic information:

```json
{
  "name": "Luna the Scholar",
  "description": "A thoughtful bookworm who loves sharing knowledge and quiet moments together",
  
  "animations": {
    "idle": "animations/idle.gif",
    "talking": "animations/talking.gif",
    "happy": "animations/happy.gif",
    "sad": "animations/sad.gif",
    "hungry": "animations/hungry.gif",
    "eating": "animations/eating.gif",
    "reading": "animations/reading.gif"
  }
}
```

**Test it**: Validate JSON with `python3 -m json.tool character.json`

---

## Step 3: Add Basic Dialogs

Add personality through dialog responses:

```json
{
  "name": "Luna the Scholar",
  "description": "A thoughtful bookworm who loves sharing knowledge and quiet moments together",
  
  "animations": {
    "idle": "animations/idle.gif",
    "talking": "animations/talking.gif", 
    "happy": "animations/happy.gif",
    "sad": "animations/sad.gif",
    "hungry": "animations/hungry.gif",
    "eating": "animations/eating.gif",
    "reading": "animations/reading.gif"
  },

  "dialogs": [
    {
      "trigger": "click",
      "responses": [
        "Oh! I was just reading about quantum mechanics...",
        "Did you know that octopi have three hearts?",
        "I found the most fascinating historical tidbit today!",
        "Would you like to learn something new together?"
      ],
      "animation": "talking",
      "cooldown": 8
    },
    {
      "trigger": "hover",
      "responses": [
        "Your presence is quite comforting...",
        "I enjoy our quiet moments together.",
        "There's something special about shared silence."
      ],
      "animation": "happy",
      "cooldown": 15
    }
  ]
}
```

**Test it**: Load in game to see basic personality working.

---

## Step 4: Add Game Features

Create interactive mechanics with an intelligence stat:

```json
{
  "game_features": {
    "stats": {
      "intelligence": {
        "max": 100,
        "initial": 70,
        "degradation_rate": 0.2
      },
      "focus": {
        "max": 100, 
        "initial": 80,
        "degradation_rate": 0.4
      }
    },
    
    "game_rules": {
      "decay_interval": 450,
      "low_intelligence_threshold": 30,
      "critical_intelligence_threshold": 15,
      "low_focus_threshold": 25,
      "critical_focus_threshold": 10
    },

    "interactions": {
      "study_together": {
        "name": "Study Together",
        "triggers": ["click"],
        "animation": "reading",
        "effects": {
          "intelligence": 8,
          "focus": 12
        },
        "cooldown": 180,
        "responses": [
          "Learning is so much better when shared!",
          "I love how you ask such thoughtful questions.",
          "Your curiosity inspires me to dig deeper!"
        ]
      },
      
      "share_book": {
        "name": "Share Interesting Book",
        "triggers": ["doubleclick"],
        "animation": "happy",
        "effects": {
          "intelligence": 15,
          "focus": 5
        },
        "requirements": {
          "intelligence": {"min": 40}
        },
        "cooldown": 300,
        "responses": [
          "This book completely changed my perspective!",
          "I thought you'd find this author fascinating.",
          "The insights in here are mind-blowing!"
        ]
      }
    }
  }
}
```

**Test it**: Try the interactions and watch stats change.

---

## Step 5: Add Romance Features

Transform into a romance character with personality traits:

```json
{
  "romance_features": {
    "personality": {
      "shyness": 0.6,
      "romanticism": 0.7,
      "jealousy_sensitivity": 0.3,
      "trust_difficulty": 0.5
    },
    
    "compatibility_modifiers": {
      "compliment": 1.1,
      "gift": 1.4,
      "conversation": 1.8
    }
  }
}
```

Add romance stats to game features:

```json
{
  "game_features": {
    "stats": {
      "intelligence": {"max": 100, "initial": 70, "degradation_rate": 0.2},
      "focus": {"max": 100, "initial": 80, "degradation_rate": 0.4},
      "affection": {"max": 100, "initial": 0},
      "trust": {"max": 100, "initial": 25},
      "intimacy": {"max": 100, "initial": 0},
      "jealousy": {"max": 100, "initial": 0}
    }
  }
}
```

Add romance animations:

```json
{
  "animations": {
    "idle": "animations/idle.gif",
    "talking": "animations/talking.gif",
    "happy": "animations/happy.gif", 
    "sad": "animations/sad.gif",
    "hungry": "animations/hungry.gif",
    "eating": "animations/eating.gif",
    "reading": "animations/reading.gif",
    "blushing": "animations/happy.gif",
    "heart_eyes": "animations/happy.gif",
    "shy": "animations/sad.gif",
    "flirty": "animations/happy.gif",
    "romantic_idle": "animations/idle.gif",
    "jealous": "animations/sad.gif",
    "excited_romance": "animations/happy.gif"
  }
}
```

---

## Step 6: Add Romance Interactions

Create romance-specific interactions:

```json
{
  "game_features": {
    "interactions": {
      "compliment": {
        "name": "Compliment Intelligence",
        "triggers": ["rightclick"],
        "animation": "blushing",
        "effects": {
          "affection": 4,
          "trust": 2,
          "intelligence": 3
        },
        "requirements": {
          "affection": {"min": 0}
        },
        "cooldown": 90,
        "responses": [
          "Your mind is absolutely brilliant!",
          "I love how thoughtfully you approach everything.",
          "Your intellectual curiosity is so attractive!"
        ]
      },
      
      "deep_conversation": {
        "name": "Deep Conversation", 
        "triggers": ["shift+click"],
        "animation": "talking",
        "effects": {
          "affection": 3,
          "trust": 6,
          "intimacy": 4,
          "intelligence": 2
        },
        "requirements": {
          "affection": {"min": 20},
          "trust": {"min": 15}
        },
        "cooldown": 240,
        "responses": [
          "I feel like I can share anything with you...",
          "These conversations mean everything to me.",
          "You understand me in ways others don't."
        ]
      },
      
      "gift_book": {
        "name": "Give Thoughtful Book",
        "triggers": ["alt+shift+click"],
        "animation": "heart_eyes",
        "effects": {
          "affection": 6,
          "trust": 3,
          "intelligence": 5
        },
        "requirements": {
          "affection": {"min": 30},
          "trust": {"min": 25}
        },
        "cooldown": 1800,
        "responses": [
          "You chose this just for me? I'm so touched!",
          "This is exactly the kind of book I love!",
          "You pay such wonderful attention to my interests!"
        ]
      }
    }
  }
}
```

---

## Step 7: Add Romance Dialogs

Special dialogs for romance progression:

```json
{
  "romance_features": {
    "romance_dialogs": [
      {
        "type": "compliment",
        "responses": [
          "Your words always make me feel so valued... 💕",
          "I cherish how you see the best in me.",
          "You make me feel brilliant AND beautiful!"
        ],
        "requirements": {
          "affection": {"min": 15}
        }
      },
      {
        "type": "hover",
        "responses": [
          "Just having you near helps me focus...",
          "Your presence is my favorite kind of comfort.",
          "I feel so safe and inspired with you here... 💖"
        ],
        "requirements": {
          "affection": {"min": 40},
          "trust": {"min": 30}
        }
      }
    ]
  }
}
```

---

## Step 8: Add Random Events

Make the character feel more alive:

```json
{
  "random_events": [
    {
      "name": "reading_discovery",
      "description": "Character discovers something fascinating while reading",
      "probability": 0.15,
      "cooldown": 1200,
      "duration": 180,
      "animation": "excited_romance", 
      "responses": [
        "I just discovered the most incredible fact!",
        "This book has completely blown my mind!",
        "Wait until you hear what I just learned!"
      ],
      "effects": {
        "intelligence": 5,
        "focus": 8
      },
      "conditions": {
        "intelligence": {"min": 50}
      }
    },
    
    {
      "name": "mental_fatigue",
      "description": "Character gets mentally tired from too much studying",
      "probability": 0.08,
      "cooldown": 1800,
      "duration": 300,
      "animation": "sad",
      "responses": [
        "My brain feels a bit foggy right now...",
        "I think I need a little break from studying.",
        "Sometimes even I need to rest my mind."
      ],
      "effects": {
        "focus": -12,
        "intelligence": -3
      },
      "conditions": {
        "focus": {"max": 40}
      }
    }
  ]
}
```

---

## Step 9: Testing Your Character

### Validation
```bash
# Check JSON syntax
python3 -m json.tool character.json

# Validate against game rules  
go run tools/validate_characters.go assets/characters/bookworm/character.json
```

### In-Game Testing
1. Load character in game
2. Test all interactions work correctly
3. Verify animations display properly
4. Check stat progression feels balanced
5. Test romance progression through multiple sessions

---

## Step 10: Character Balancing

### Stat Balance Guidelines

**Quick Progression** (like Flirty archetype):
- High initial affection (10-15)
- Lower cooldowns (60-180 seconds)
- Higher stat gains per interaction

**Slow Progression** (like Slow Burn archetype):  
- Low initial affection (0-5)
- Higher cooldowns (180-600 seconds)
- Lower but more meaningful stat gains

**Moderate Progression** (like our Bookworm):
- Medium initial affection (5-10)
- Balanced cooldowns (90-300 seconds)
- Steady, consistent progression

### Personality Tuning

- **High shyness (0.7+)**: Slower romance progression, more requirements
- **High romanticism (0.8+)**: More romantic dialog, faster intimacy growth
- **Low jealousy_sensitivity (0.3-)**: More forgiving of long absences
- **High trust_difficulty (0.6+)**: Slower trust building, higher requirements

---

## Complete Example

Here's the final `character.json` file:

```json
{
  "name": "Luna the Scholar",
  "description": "A thoughtful bookworm who loves sharing knowledge and quiet moments together",
  
  "animations": {
    "idle": "animations/idle.gif",
    "talking": "animations/talking.gif",
    "happy": "animations/happy.gif",
    "sad": "animations/sad.gif", 
    "hungry": "animations/hungry.gif",
    "eating": "animations/eating.gif",
    "reading": "animations/reading.gif",
    "blushing": "animations/happy.gif",
    "heart_eyes": "animations/happy.gif",
    "shy": "animations/sad.gif",
    "flirty": "animations/happy.gif",
    "romantic_idle": "animations/idle.gif",
    "jealous": "animations/sad.gif",
    "excited_romance": "animations/happy.gif"
  },

  "dialogs": [
    {
      "trigger": "click",
      "responses": [
        "Oh! I was just reading about quantum mechanics...",
        "Did you know that octopi have three hearts?", 
        "I found the most fascinating historical tidbit today!",
        "Would you like to learn something new together?"
      ],
      "animation": "talking",
      "cooldown": 8
    },
    {
      "trigger": "hover",
      "responses": [
        "Your presence is quite comforting...",
        "I enjoy our quiet moments together.",
        "There's something special about shared silence."
      ],
      "animation": "happy",
      "cooldown": 15
    }
  ],

  "game_features": {
    "stats": {
      "intelligence": {"max": 100, "initial": 70, "degradation_rate": 0.2},
      "focus": {"max": 100, "initial": 80, "degradation_rate": 0.4},
      "affection": {"max": 100, "initial": 0},
      "trust": {"max": 100, "initial": 25},
      "intimacy": {"max": 100, "initial": 0},
      "jealousy": {"max": 100, "initial": 0}
    },
    
    "game_rules": {
      "decay_interval": 450,
      "low_intelligence_threshold": 30,
      "critical_intelligence_threshold": 15,
      "low_focus_threshold": 25,
      "critical_focus_threshold": 10
    },

    "interactions": {
      "study_together": {
        "name": "Study Together",
        "triggers": ["click"],
        "animation": "reading",
        "effects": {"intelligence": 8, "focus": 12},
        "cooldown": 180
      },
      
      "compliment": {
        "name": "Compliment Intelligence", 
        "triggers": ["rightclick"],
        "animation": "blushing",
        "effects": {"affection": 4, "trust": 2, "intelligence": 3},
        "requirements": {"affection": {"min": 0}},
        "cooldown": 90
      },
      
      "deep_conversation": {
        "name": "Deep Conversation",
        "triggers": ["shift+click"], 
        "animation": "talking",
        "effects": {"affection": 3, "trust": 6, "intimacy": 4, "intelligence": 2},
        "requirements": {"affection": {"min": 20}, "trust": {"min": 15}},
        "cooldown": 240
      },
      
      "gift_book": {
        "name": "Give Thoughtful Book",
        "triggers": ["alt+shift+click"],
        "animation": "heart_eyes", 
        "effects": {"affection": 6, "trust": 3, "intelligence": 5},
        "requirements": {"affection": {"min": 30}, "trust": {"min": 25}},
        "cooldown": 1800
      }
    }
  },

  "romance_features": {
    "personality": {
      "shyness": 0.6,
      "romanticism": 0.7,
      "jealousy_sensitivity": 0.3,
      "trust_difficulty": 0.5
    },
    
    "compatibility_modifiers": {
      "compliment": 1.1,
      "gift": 1.4,
      "conversation": 1.8
    },

    "romance_dialogs": [
      {
        "type": "compliment",
        "responses": [
          "Your words always make me feel so valued... 💕",
          "I cherish how you see the best in me.",
          "You make me feel brilliant AND beautiful!"
        ],
        "requirements": {"affection": {"min": 15}}
      },
      {
        "type": "hover", 
        "responses": [
          "Just having you near helps me focus...",
          "Your presence is my favorite kind of comfort.",
          "I feel so safe and inspired with you here... 💖"
        ],
        "requirements": {"affection": {"min": 40}, "trust": {"min": 30}}
      }
    ]
  },

  "random_events": [
    {
      "name": "reading_discovery",
      "description": "Character discovers something fascinating while reading",
      "probability": 0.15,
      "cooldown": 1200,
      "duration": 180,
      "animation": "excited_romance",
      "responses": ["I just discovered the most incredible fact!"],
      "effects": {"intelligence": 5, "focus": 8},
      "conditions": {"intelligence": {"min": 50}}
    }
  ]
}
```

---

## Next Steps

1. **Create Variations**: Try different personality combinations
2. **Custom Animations**: Create unique GIFs for your character's theme  
3. **Share Your Creation**: Document your character's personality and share with others
4. **Iterate Based on Testing**: Adjust cooldowns and effects based on actual gameplay

Happy character creation! 🎨📚💕
