# Dialog Backend Configuration Guide

## Overview

The advanced dialog system allows your desktop companion to generate dynamic, context-aware responses using various AI backends. Instead of fixed response lists, your character can create unique dialogs based on personality, mood, and conversation history.

## Quick Start

### Step 1: Enable Advanced Dialogs

Add a `dialogBackend` section to your character card:

```json
{
  "name": "My Character",
  "description": "An enhanced companion with dynamic dialogs",
  
  // ... your existing animations, dialogs, etc ...
  
  "dialogBackend": {
    "enabled": true,
    "defaultBackend": "markov_chain",
    "confidenceThreshold": 0.6
  }
}
```

### Step 2: Choose a Backend Type

#### Simple Random (Recommended for Beginners)
Uses your existing dialog responses with smarter selection:

```json
"dialogBackend": {
  "enabled": true,
  "defaultBackend": "simple_random",
  "backends": {
    "simple_random": {
      "type": "basic"
    }
  }
}
```

#### Markov Chain (Recommended for Dynamic Dialog)
Generates new responses based on training examples:

```json
"dialogBackend": {
  "enabled": true,
  "defaultBackend": "markov_chain",
  "backends": {
    "markov_chain": {
      "chainOrder": 2,
      "minWords": 3,
      "maxWords": 12,
      "temperatureMin": 0.4,
      "temperatureMax": 0.7,
      "trainingData": [
        "Hello! I'm so happy to see you!",
        "Thank you for spending time with me!",
        "Your presence always brightens my day!"
      ]
    }
  }
}
```

## Backend Types

### 1. Simple Random Backend

**Best for**: Beginners, characters with fixed personalities, testing

**Configuration**:
```json
"simple_random": {
  "type": "basic",
  "personalityInfluence": 0.3
}
```

**Pros**:
- ✅ Easy to configure
- ✅ Predictable responses
- ✅ No additional memory usage
- ✅ Works with existing dialog lists

**Cons**:
- ❌ Limited variety
- ❌ No learning capability
- ❌ Responses can feel repetitive

### 2. Markov Chain Backend

**Best for**: Dynamic personalities, characters that should "learn" from conversations, variety

**Configuration**:
```json
"markov_chain": {
  "chainOrder": 2,
  "minWords": 4,
  "maxWords": 15,
  "temperatureMin": 0.3,
  "temperatureMax": 0.8,
  "trainingData": [
    "Your training phrases go here",
    "Each phrase should match your character's voice",
    "Include emotional expressions and personality"
  ],
  "usePersonality": true,
  "triggerSpecific": true
}
```

**Pros**:
- ✅ Generates unique responses
- ✅ Learns from examples
- ✅ Personality-aware
- ✅ Context-sensitive

**Cons**:
- ❌ Requires good training data
- ❌ May generate unexpected responses
- ❌ Uses more memory
- ❌ More complex to configure

## Markov Chain Configuration

### Essential Parameters

#### Chain Order (1-5)
Controls how sophisticated the text generation is:

```json
"chainOrder": 1  // Simple, random-ish responses
"chainOrder": 2  // Good balance (RECOMMENDED)
"chainOrder": 3  // More coherent, needs more training data
```

**Recommendation**: Start with `2`, increase to `3` only if you have 50+ training phrases.

#### Response Length
```json
"minWords": 3,     // Shortest possible response
"maxWords": 15     // Longest possible response
```

**Guidelines**:
- **Shy characters**: 3-8 words
- **Normal characters**: 4-12 words  
- **Talkative characters**: 6-20 words

#### Temperature (Randomness)
```json
"temperatureMin": 0.3,  // Most predictable responses
"temperatureMax": 0.8   // Most random responses
```

**Character Types**:
- **Predictable/Shy**: 0.2-0.5
- **Balanced**: 0.4-0.7 (RECOMMENDED)
- **Spontaneous**: 0.6-1.0
- **Chaotic**: 0.8-1.2

### Training Data

#### Quality over Quantity
Better to have 20 excellent examples than 100 mediocre ones:

```json
"trainingData": [
  "Hello there! I've been waiting for you to visit me today.",
  "Your presence always makes me feel so much happier and loved.",
  "Thank you for taking the time to spend these moments with me.",
  "I hope you're having a wonderful day, you truly deserve happiness."
]
```

#### Match Your Character's Voice

**Shy Character**:
```json
"trainingData": [
  "Oh... hi there... I'm happy you came to see me...",
  "Um... thank you for visiting... it means a lot to me...",
  "I hope... I hope you don't mind me being a bit quiet today..."
]
```

**Flirty Character**:
```json
"trainingData": [
  "Well hello there, gorgeous! I was hoping you'd come see me!",
  "You always know how to make my heart race with excitement!",
  "Come closer, darling, I want to tell you something special!"
]
```

**Romantic Character**:
```json
"trainingData": [
  "My dearest love, every moment with you feels like a precious gift.",
  "Your gentle touch and caring words fill my heart with pure joy.",
  "I treasure these intimate moments we share together, my beloved."
]
```

### Advanced Features

#### Personality Integration
```json
"usePersonality": true,
"personalityBoost": 0.5
```

When enabled, responses adapt based on your character's personality traits:
- `shyness`: Affects response length and boldness
- `romanticism`: Influences emotional expression
- `flirtiness`: Controls playful language
- `creativity`: Affects response variety

#### Mood Influence
```json
"moodInfluence": 0.3
```

Responses change based on character's current mood:
- **Happy mood**: More energetic, longer responses
- **Sad mood**: Shorter, more subdued responses
- **Excited mood**: Higher variety and creativity

#### Trigger-Specific Chains
```json
"triggerSpecific": true
```

Creates separate response patterns for different interactions:
- `click`: General conversation
- `hover`: Brief acknowledgments  
- `compliment`: Grateful responses
- `give_gift`: Excited reactions

### Quality Control

#### Forbidden Words
```json
"forbiddenWords": ["hate", "stupid", "ugly", "boring"]
```

Prevents the system from generating responses containing these words.

#### Required Words (Optional)
```json
"requiredWords": ["love", "happiness", "joy", "grateful"]
```

Encourages responses containing these character-appropriate terms.

#### Fallback Phrases
```json
"fallbackPhrases": [
  "I'm so happy you're here with me!",
  "Thank you for being such a wonderful friend!",
  "You always know how to make me smile!"
]
```

High-quality responses used when generation fails.

## Complete Examples

### Shy Romance Character
```json
{
  "name": "Shy Sweetheart",
  "dialogBackend": {
    "enabled": true,
    "defaultBackend": "markov_chain",
    "confidenceThreshold": 0.7,
    "backends": {
      "markov_chain": {
        "chainOrder": 2,
        "minWords": 3,
        "maxWords": 10,
        "temperatureMin": 0.2,
        "temperatureMax": 0.5,
        "usePersonality": true,
        "personalityBoost": 0.6,
        "triggerSpecific": true,
        "trainingData": [
          "Oh... hi there... I'm so glad you came to see me...",
          "Um... thank you for being so patient with me...",
          "I feel so safe and comfortable when you're here with me...",
          "Your gentle presence always makes my heart feel warm...",
          "I hope... I hope I'm not too shy for you...",
          "Every moment with you feels like a precious gift to treasure..."
        ],
        "forbiddenWords": ["loud", "bold", "aggressive"],
        "fallbackPhrases": [
          "I'm happy you're here... 😊",
          "Thank you for being so kind...",
          "You make me feel safe..."
        ]
      }
    }
  }
}
```

### Energetic Companion
```json
{
  "name": "Energetic Buddy",
  "dialogBackend": {
    "enabled": true,
    "defaultBackend": "markov_chain",
    "backends": {
      "markov_chain": {
        "chainOrder": 2,
        "minWords": 6,
        "maxWords": 18,
        "temperatureMin": 0.6,
        "temperatureMax": 1.0,
        "usePersonality": true,
        "moodInfluence": 0.4,
        "triggerSpecific": true,
        "trainingData": [
          "Hey there, superstar! I've been waiting all day for you to visit!",
          "Wow, you look absolutely amazing today! Tell me about your adventures!",
          "I'm so excited to spend time with you! What fun things should we do?",
          "Your energy is absolutely contagious! You always brighten my entire day!",
          "Come on, let's chat about everything! I want to hear all your stories!",
          "Thank you for bringing so much joy and excitement into my life!"
        ],
        "requiredWords": ["excited", "amazing", "wonderful", "fantastic"],
        "fallbackPhrases": [
          "You're absolutely wonderful!",
          "I'm so excited to see you!",
          "Let's have an amazing time together!"
        ]
      }
    }
  }
}
```

## Troubleshooting

### Common Issues

#### "Responses are too random/incoherent"
- **Solution**: Lower `temperatureMax` to 0.5-0.6
- **Or**: Increase `chainOrder` to 3 (if you have 50+ training phrases)
- **Or**: Add more high-quality training data

#### "Responses are too repetitive"
- **Solution**: Increase `temperatureMin` to 0.4-0.5
- **Or**: Add more variety to training data
- **Or**: Enable `triggerSpecific` for context variety

#### "Character doesn't sound like themselves"
- **Solution**: Review training data for consistency
- **Or**: Enable `usePersonality` and configure personality traits
- **Or**: Add character-specific words to `requiredWords`

#### "Responses are too short/long"
- **Solution**: Adjust `minWords` and `maxWords`
- **Or**: Review training data length (responses follow examples)

#### "System falls back to simple responses too often"
- **Solution**: Lower `confidenceThreshold` to 0.4-0.5
- **Or**: Add more training data
- **Or**: Check for `forbiddenWords` blocking responses

### Testing Your Configuration

1. **Start Simple**: Begin with `simple_random` backend
2. **Add Training Data**: Switch to `markov_chain` with 10-20 examples
3. **Test Interactions**: Try different triggers (click, hover, etc.)
4. **Adjust Temperature**: Fine-tune randomness for your character
5. **Add Personality**: Enable personality features once basics work
6. **Expand Training**: Add more examples for variety

## Best Practices

### 1. Character Voice Consistency
- All training phrases should sound like the same character
- Use consistent tone, vocabulary, and personality
- Include character-specific expressions and quirks

### 2. Balanced Training Data
- Include examples for different emotional states
- Mix short and long responses
- Cover various conversation topics

### 3. Gradual Complexity
- Start with basic configuration
- Add advanced features one at a time
- Test thoroughly at each step

### 4. User Experience
- Avoid controversial or offensive content
- Ensure responses feel natural and appropriate
- Test with different interaction patterns

### 5. Performance Considerations
- More training data = better quality but more memory
- Higher chain order = more coherent but needs more examples
- Trigger-specific chains = better context but more memory

## Templates Library

### Basic Friendly Character
```json
"backends": {
  "markov_chain": {
    "chainOrder": 2,
    "minWords": 4,
    "maxWords": 12,
    "temperatureMin": 0.4,
    "temperatureMax": 0.7,
    "trainingData": [
      "Hello! I'm so happy to see you today!",
      "Thank you for spending time with me!",
      "Your visits always make my day brighter!",
      "I hope you're having a wonderful time!",
      "You're such a kind and caring person!",
      "I feel grateful for our friendship!"
    ]
  }
}
```

### Romance Character Template
```json
"backends": {
  "markov_chain": {
    "chainOrder": 2,
    "minWords": 5,
    "maxWords": 16,
    "temperatureMin": 0.3,
    "temperatureMax": 0.8,
    "usePersonality": true,
    "personalityBoost": 0.5,
    "moodInfluence": 0.3,
    "triggerSpecific": true,
    "trainingData": [
      "My heart fills with joy every time I see your beautiful face.",
      "You mean everything to me, my dearest and most precious love.",
      "Thank you for bringing so much love and warmth into my life.",
      "Every moment with you feels like a dream come true, my darling.",
      "I cherish our relationship and the beautiful bond we share together.",
      "Your love gives me strength and fills my days with happiness."
    ],
    "requiredWords": ["love", "heart", "beautiful", "precious"],
    "fallbackPhrases": [
      "You mean so much to me... 💕",
      "I love being with you!",
      "You make my heart happy!"
    ]
  }
}
```

Copy and modify these templates to match your character's unique personality and voice!

## Support

If you need help configuring your character's dialog system:

1. **Check Examples**: Review the templates above for similar characters
2. **Start Simple**: Begin with basic configuration and add complexity gradually  
3. **Test Frequently**: Make small changes and test the results
4. **Read Error Messages**: Configuration errors provide helpful guidance
5. **Community**: Share your configurations and learn from others

The advanced dialog system opens up endless possibilities for creating unique, engaging characters. Start with the templates and gradually customize them to create your perfect digital companion!
