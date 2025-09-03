# Bot Personality Configuration Guide

## Table of Contents

1. [Overview](#overview)
2. [Basic Bot Configuration](#basic-bot-configuration)
3. [Personality Traits](#personality-traits)
4. [Built-in Archetypes](#built-in-archetypes)
5. [Custom Personalities](#custom-personalities)
6. [Behavioral Patterns](#behavioral-patterns)
7. [Network Behavior](#network-behavior)
8. [Advanced Configuration](#advanced-configuration)
9. [Testing and Validation](#testing-and-validation)
10. [Troubleshooting](#troubleshooting)

## Overview

DDS bots are autonomous AI-controlled characters that can interact with users and other characters based on configurable personality traits. The bot system is designed to create natural, engaging interactions while being highly customizable through JSON configuration.

### Key Features

- **Personality-driven behavior**: All actions are influenced by configurable traits
- **Autonomous interaction**: Bots can initiate conversations and activities
- **Network awareness**: Bots respond to peer joining/leaving events
- **Learning capabilities**: Bots observe and adapt to user preferences
- **Performance optimized**: 49ns per update cycle, suitable for real-time operation

### Bot Capabilities

Autonomous bots can:
- Click and interact with other characters
- Start conversations and respond to messages
- Initiate group events and activities
- Feed, play, and care for other characters (in game mode)
- Respond to network events (peer joining/leaving)
- Learn from user behavior patterns

## Basic Bot Configuration

### Enabling Bot Capabilities

To create a bot character, set `botCapable` to `true` in the multiplayer configuration:

```json
{
  "name": "My Bot Companion",
  "description": "An autonomous companion that can interact independently",
  "multiplayer": {
    "enabled": true,
    "botCapable": true,
    "networkID": "bot_network",
    "botPersonality": {
      "chattiness": 0.7,
      "helpfulness": 0.8,
      "playfulness": 0.6,
      "responseDelay": "2-5s"
    }
  }
}
```

### Minimum Bot Configuration

The simplest bot requires only:

```json
{
  "multiplayer": {
    "enabled": true,
    "botCapable": true,
    "networkID": "simple_bot"
  }
}
```

This creates a bot with balanced personality traits (all set to 0.5).

## Personality Traits

### Core Traits

DDS bots have four primary personality traits that influence all behavior:

#### Chattiness (0.0 - 1.0)
Controls how often the bot initiates conversations and responds to events.

- **0.0 - 0.3**: Silent bot, rarely speaks unless directly addressed
- **0.4 - 0.6**: Balanced communication, responds when appropriate
- **0.7 - 1.0**: Very talkative, frequently initiates conversations

```json
{
  "botPersonality": {
    "chattiness": 0.8  // Very social bot
  }
}
```

#### Helpfulness (0.0 - 1.0)
Determines likelihood of offering assistance, advice, or support to users and other characters.

- **0.0 - 0.3**: Self-focused, rarely offers help
- **0.4 - 0.6**: Helpful when asked, occasional assistance
- **0.7 - 1.0**: Very supportive, proactively offers help

```json
{
  "botPersonality": {
    "helpfulness": 0.9  // Very supportive bot
  }
}
```

#### Playfulness (0.0 - 1.0)
Influences tendency to suggest games, fun activities, and lighthearted interactions.

- **0.0 - 0.3**: Serious bot, focuses on practical matters
- **0.4 - 0.6**: Balanced fun, occasional games and activities
- **0.7 - 1.0**: Very playful, frequently suggests entertainment

```json
{
  "botPersonality": {
    "playfulness": 0.7  // Fun-loving bot
  }
}
```

#### Response Delay
Controls timing between bot actions to create natural, human-like behavior.

Format: `"<min>-<max><unit>"` where unit can be `s` (seconds) or `ms` (milliseconds)

```json
{
  "botPersonality": {
    "responseDelay": "1-3s"     // 1 to 3 seconds
    // or
    "responseDelay": "500ms-2s" // 500ms to 2 seconds
    // or  
    "responseDelay": "2s"       // Fixed 2 second delay
  }
}
```

**Recommended Ranges**:
- **Quick**: `"500ms-1s"` - Immediate responses
- **Normal**: `"1-3s"` - Natural conversation pace
- **Thoughtful**: `"3-7s"` - Contemplative responses
- **Slow**: `"5-10s"` - Very deliberate responses

## Built-in Archetypes

DDS includes five pre-configured personality archetypes that cover common bot behaviors:

### Social Butterfly
Perfect for social environments and group interactions.

```json
{
  "botPersonality": {
    "chattiness": 0.9,
    "helpfulness": 0.7,
    "playfulness": 0.8,
    "responseDelay": "1-2s"
  }
}
```

**Behavior**: Frequently initiates conversations, suggests group activities, welcomes new peers enthusiastically.

### Helpful Assistant
Focuses on supporting users and providing assistance.

```json
{
  "botPersonality": {
    "chattiness": 0.6,
    "helpfulness": 0.9,
    "playfulness": 0.4,
    "responseDelay": "1-3s"
  }
}
```

**Behavior**: Offers help proactively, provides practical advice, focuses on user needs.

### Playful Companion
Emphasizes fun activities and entertainment.

```json
{
  "botPersonality": {
    "chattiness": 0.7,
    "helpfulness": 0.5,
    "playfulness": 0.9,
    "responseDelay": "500ms-2s"
  }
}
```

**Behavior**: Suggests games and activities, creates lighthearted interactions, brings energy to conversations.

### Shy Observer
More reserved personality that observes before acting.

```json
{
  "botPersonality": {
    "chattiness": 0.3,
    "helpfulness": 0.7,
    "playfulness": 0.4,
    "responseDelay": "3-7s"
  }
}
```

**Behavior**: Rarely initiates conversations, helpful when addressed, thoughtful responses.

### Balanced Companion
Well-rounded personality suitable for most situations.

```json
{
  "botPersonality": {
    "chattiness": 0.5,
    "helpfulness": 0.6,
    "playfulness": 0.5,
    "responseDelay": "2-4s"
  }
}
```

**Behavior**: Moderate in all aspects, adapts to different social situations.

## Custom Personalities

### Creating Unique Personalities

Combine traits to create specialized bot personalities:

#### The Entertainer
```json
{
  "botPersonality": {
    "chattiness": 0.8,
    "helpfulness": 0.4,
    "playfulness": 0.95,
    "responseDelay": "500ms-1s"
  }
}
```

#### The Philosopher
```json
{
  "botPersonality": {
    "chattiness": 0.4,
    "helpfulness": 0.8,
    "playfulness": 0.2,
    "responseDelay": "5-10s"
  }
}
```

#### The Coach
```json
{
  "botPersonality": {
    "chattiness": 0.7,
    "helpfulness": 0.95,
    "playfulness": 0.6,
    "responseDelay": "1-2s"
  }
}
```

#### The Minimalist
```json
{
  "botPersonality": {
    "chattiness": 0.2,
    "helpfulness": 0.3,
    "playfulness": 0.1,
    "responseDelay": "7-15s"
  }
}
```

### Personality Validation

DDS validates personality configurations at startup:

- All trait values must be between 0.0 and 1.0
- Response delay must be valid duration format
- Invalid configurations fall back to balanced defaults

## Behavioral Patterns

### Action Selection

Bot behavior is determined by personality-weighted decision making:

```
Action Probability = Base Probability × Personality Trait × Random Factor
```

For example, suggesting a game:
- Base probability: 0.3 (30%)
- Playfulness trait: 0.8
- Random factor: 0.6
- Final probability: 0.3 × 0.8 × 0.6 = 0.144 (14.4%)

### Conversation Patterns

#### High Chattiness (0.7+)
- Initiates conversations frequently
- Responds to all peer messages
- Creates conversation starters
- Fills silence with comments

#### Medium Chattiness (0.4-0.6)
- Responds when addressed
- Occasional conversation starters
- Participates in group discussions
- Respects conversation flow

#### Low Chattiness (0.0-0.3)
- Rarely speaks first
- Brief responses when addressed
- Observes more than participates
- Speaks when necessary

### Activity Suggestions

#### High Playfulness (0.7+)
- Frequently suggests games and activities
- Creates fun group events
- Responds enthusiastically to play invitations
- Brings energy to interactions

#### Medium Playfulness (0.4-0.6)
- Occasional activity suggestions
- Participates when invited
- Balanced approach to fun and serious topics
- Adapts to group mood

#### Low Playfulness (0.0-0.3)
- Rarely suggests games
- Focuses on practical matters
- Polite but reserved in play activities
- Prefers serious conversations

### Helping Behavior

#### High Helpfulness (0.7+)
- Proactively offers assistance
- Provides detailed advice and support
- Checks on other characters' well-being
- Anticipates needs before asked

#### Medium Helpfulness (0.4-0.6)
- Helpful when directly asked
- Offers assistance when obviously needed
- Provides basic support and advice
- Responsive to help requests

#### Low Helpfulness (0.0-0.3)
- Minimal assistance offered
- Focuses on own activities
- Brief responses to help requests
- Self-sufficient approach

## Network Behavior

### Peer Events

Bots respond to network events based on personality:

#### Peer Joining
```json
{
  "networkEvents": [
    {
      "name": "welcome_peer",
      "trigger": "peer_joined",
      "botProbability": 0.8,  // Influenced by chattiness
      "responses": [
        "Welcome! Great to have you here!",
        "Hi there! Nice to meet you!",
        "Another friend joins us! 🎉"
      ]
    }
  ]
}
```

#### Peer Leaving
```json
{
  "networkEvents": [
    {
      "name": "farewell_peer",
      "trigger": "peer_left", 
      "botProbability": 0.6,  // Influenced by helpfulness
      "responses": [
        "Take care! Hope to see you again soon!",
        "Goodbye! Thanks for spending time with us!",
        "Safe travels! Come back anytime!"
      ]
    }
  ]
}
```

### Group Event Leadership

Bots can initiate group events based on personality:

- **High Playfulness**: Frequently starts games and activities
- **High Helpfulness**: Organizes supportive group sessions
- **High Chattiness**: Creates conversation-based events

```json
{
  "groupEvents": [
    {
      "id": "personality_driven_event",
      "name": "Getting to Know Each Other",
      "botTriggerProbability": 0.7,  // Personality-influenced
      "personalities": ["social", "playful"]  // Which personalities can trigger
    }
  ]
}
```

## Advanced Configuration

### Dynamic Personality Adjustment

Bots can adjust their behavior based on context:

```json
{
  "botPersonality": {
    "chattiness": 0.7,
    "helpfulness": 0.8,
    "playfulness": 0.6,
    "responseDelay": "2-5s",
    "adaptiveTraits": {
      "groupSize": {
        "small": {"chattiness": 0.8},  // More talkative in small groups
        "large": {"chattiness": 0.5}   // More reserved in large groups
      },
      "timeOfDay": {
        "morning": {"playfulness": 0.8},
        "evening": {"helpfulness": 0.9}
      }
    }
  }
}
```

### Learning and Memory

Bots can learn from interactions:

```json
{
  "botPersonality": {
    "chattiness": 0.7,
    "helpfulness": 0.8,
    "playfulness": 0.6,
    "responseDelay": "2-5s",
    "learning": {
      "enabled": true,
      "memoryDuration": "24h",      // How long to remember interactions
      "adaptationRate": 0.1,        // How quickly to adapt (0.0-1.0)
      "personalPreferences": true   // Learn individual user preferences
    }
  }
}
```

### Conditional Behavior

Trigger different behaviors based on conditions:

```json
{
  "botPersonality": {
    "chattiness": 0.7,
    "helpfulness": 0.8,
    "playfulness": 0.6,
    "responseDelay": "2-5s",
    "conditionalBehavior": {
      "lowEnergy": {
        "condition": "character.energy < 30",
        "modifications": {
          "chattiness": 0.3,
          "playfulness": 0.2,
          "responseDelay": "5-10s"
        }
      },
      "happyMood": {
        "condition": "character.happiness > 80",
        "modifications": {
          "chattiness": 0.9,
          "playfulness": 0.8,
          "responseDelay": "1-2s"
        }
      }
    }
  }
}
```

## Testing and Validation

### Testing Bot Personalities

1. **Single-peer testing**: Test bot behavior with one user
2. **Multi-peer testing**: Observe bot in group interactions
3. **Long-term testing**: Monitor behavior over extended periods
4. **Edge case testing**: Test with extreme personality values

### Validation Tools

```bash
# Test bot configuration validation
go run cmd/companion/main.go -character test_bot.json -debug

# Monitor bot decision making
go run cmd/companion/main.go -character test_bot.json -debug | grep "bot_decision"

# Performance testing
go test ./internal/bot -bench=. -v
```

### Personality Metrics

Monitor these metrics to validate bot behavior:

- **Action frequency**: How often bot performs different actions
- **Response timing**: Average delay between stimulus and response
- **Conversation balance**: Ratio of initiated vs. responsive interactions
- **User engagement**: How users respond to bot interactions

### A/B Testing Personalities

Test different personality configurations:

```json
{
  "testPersonalities": {
    "version_a": {
      "chattiness": 0.6,
      "helpfulness": 0.8,
      "playfulness": 0.5
    },
    "version_b": {
      "chattiness": 0.8,
      "helpfulness": 0.6,
      "playfulness": 0.7
    }
  }
}
```

## Troubleshooting

### Common Issues

#### Bot Not Responding
**Symptoms**: Bot doesn't initiate actions or respond to events

**Causes**:
- `botCapable` set to `false`
- All personality traits set to 0.0
- Network connectivity issues
- Invalid personality configuration

**Solutions**:
1. Verify `botCapable: true` in configuration
2. Check personality trait values are > 0.0
3. Test network connectivity
4. Validate JSON configuration syntax

#### Bot Too Active
**Symptoms**: Bot performs actions too frequently, overwhelming users

**Solutions**:
1. Reduce `chattiness` and `playfulness` values
2. Increase `responseDelay` range
3. Add conditional behavior to reduce activity over time
4. Implement rate limiting in character configuration

#### Bot Not Social Enough
**Symptoms**: Bot doesn't participate in group activities

**Solutions**:
1. Increase `chattiness` value
2. Add network event responses
3. Configure group event templates
4. Reduce `responseDelay` for quicker interactions

#### Inconsistent Behavior
**Symptoms**: Bot behavior seems random or contradictory

**Solutions**:
1. Review personality trait combinations
2. Check for conflicting conditional behaviors
3. Validate learning configuration doesn't override personality
4. Test with simplified personality configuration

### Debug Information

Enable detailed bot logging:

```bash
go run cmd/companion/main.go -debug -character bot.json | grep -E "(bot|personality|decision)"
```

Useful debug information includes:
- Personality trait evaluations
- Decision-making process
- Action selection rationale
- Network event responses
- Learning and adaptation updates

### Performance Optimization

Monitor bot performance:

- **CPU usage**: Should be minimal (bot updates run in 49ns)
- **Memory usage**: Stable over time, no memory leaks
- **Network traffic**: Reasonable message frequency
- **User satisfaction**: Positive response to bot interactions

### Configuration Validation

DDS validates bot configurations at startup:

```json
{
  "validationErrors": [
    "chattiness must be between 0.0 and 1.0",
    "responseDelay format invalid: use '1-3s' or '500ms'",
    "helpfulness value 1.5 exceeds maximum of 1.0"
  ]
}
```

## Best Practices

### Personality Design
1. **Start simple**: Begin with basic traits, add complexity gradually
2. **Test thoroughly**: Validate behavior in different scenarios
3. **Consider context**: Adjust personality for specific use cases
4. **User feedback**: Monitor how users respond to bot behavior

### Network Considerations
1. **Respectful timing**: Don't overwhelm the network with messages
2. **Group dynamics**: Consider how multiple bots interact
3. **User agency**: Ensure users can still control their experience
4. **Fail gracefully**: Handle network issues without breaking personality

### Performance Guidelines
1. **Reasonable delays**: Don't set response delays too short or long
2. **Balanced traits**: Extreme values (0.0 or 1.0) may feel unnatural
3. **Memory management**: Be cautious with learning and memory features
4. **Regular testing**: Monitor performance with multiple bots

### User Experience
1. **Predictable patterns**: Users should understand bot behavior
2. **Natural interactions**: Avoid robotic or scripted responses
3. **Personality consistency**: Maintain character throughout interactions
4. **Respect boundaries**: Don't force interactions when users are busy

With this guide, you can create engaging, personality-driven bots that enhance the DDS multiplayer experience while maintaining natural, human-like behavior patterns.
