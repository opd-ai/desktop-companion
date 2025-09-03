# General Dialog Events System Guide

## Overview

The General Dialog Events system enables companions to initiate diverse interactive scenarios including conversations, roleplay adventures, mini-games, and humor sessions. This system operates alongside existing dialogs and game interactions, providing a rich, user-driven experience.

## Architecture

### Event Types

**Category Classification:**
- **Conversation**: Daily check-ins, advice sessions, deep discussions
- **Roleplay**: Fantasy adventures, sci-fi missions, detective stories  
- **Game**: Trivia questions, word games, creative challenges
- **Humor**: Joke sessions, pun competitions, silly interactions

### System Integration

The general events system uses a **parallel pipeline approach**:

1. **Existing Dialogs** → Handle basic interactions (click, hover, rightclick)
2. **Game Interactions** → Handle stat-affecting actions (feed, pet, play)
3. **General Events** → Handle complex user-initiated scenarios
4. **Advanced Dialog Backend** → Generate dynamic responses for all systems

## JSON Schema

### Basic Event Structure

```json
{
  "generalEvents": [
    {
      "name": "unique_event_identifier",
      "category": "conversation",
      "description": "Human-readable event description",
      "trigger": "keyboard_shortcut_or_auto_trigger",
      "probability": 1.0,
      "interactive": true,
      "responses": [
        "Initial event dialog to display"
      ],
      "animations": ["talking", "excited"],
      "choices": [
        {
          "text": "User choice text",
          "effects": {"stat_name": 5},
          "responses": ["Choice-specific response"],
          "nextEvent": "follow_up_event_name"
        }
      ],
      "cooldown": 3600,
      "conditions": {"stat_name": {"min": 10}}
    }
  ]
}
```

### Field Definitions

#### Core Properties
- **`name`** (string, required): Unique identifier for the event
- **`category`** (string, required): Event type - "conversation", "roleplay", "game", "humor"
- **`description`** (string, required): Human-readable description for debugging
- **`trigger`** (string, required): Trigger identifier (for keyboard shortcuts or auto-events)
- **`probability`** (number, required): 0.0-1.0 chance of triggering (usually 1.0 for user events)

#### Interactivity
- **`interactive`** (boolean): Whether event supports user choices
- **`choices`** (array): Available user choices (required if interactive=true)

#### Content
- **`responses`** (array): Initial dialog text to display
- **`animations`** (array): Animations to play when event triggers

#### Timing & Requirements
- **`cooldown`** (number): Seconds before event can trigger again
- **`conditions`** (object): Stat requirements to access the event
- **`minRelationship`** (string): Minimum relationship level required

#### Advanced Features
- **`followUpEvents`** (array): Events that can chain after this one
- **`keywords`** (array): Keywords for event discovery
- **`difficulty`** (string): "easy", "normal", "hard" for filtering

### Choice Structure

```json
{
  "text": "Display text for the choice",
  "effects": {
    "stat_name": 5,
    "another_stat": -2
  },
  "responses": [
    "Custom response for this choice"
  ],
  "nextEvent": "follow_up_event",
  "requirements": {
    "stat_name": {"min": 20}
  },
  "animation": "custom_animation",
  "disabled": false
}
```

## API Reference

### Character Methods

#### Event Management
```go
// Trigger a specific general event by name
func (c *Character) HandleGeneralEvent(eventName string) string

// Get all available events for the current character state
func (c *Character) GetAvailableGeneralEvents() []GeneralDialogEvent

// Get events filtered by category
func (c *Character) GetGeneralEventsByCategory(category string) []GeneralDialogEvent

// Check if a specific event can be triggered
func (c *Character) IsGeneralEventAvailable(eventName string) bool
```

#### Interactive Events
```go
// Submit a user choice in an active interactive event
func (c *Character) SubmitEventChoice(choiceIndex int) (string, bool)

// Get the currently active interactive event
func (c *Character) GetActiveGeneralEvent() *GeneralDialogEvent

// Cancel the active interactive event
func (c *Character) CancelActiveGeneralEvent()
```

### Event Manager Methods

```go
// Create new general event manager
func NewGeneralEventManager(events []GeneralDialogEvent, enabled bool) *GeneralEventManager

// Trigger specific event (with validation)
func (gem *GeneralEventManager) TriggerEvent(eventName string, gameState *GameState) (*GeneralDialogEvent, error)

// Handle user choice submission
func (gem *GeneralEventManager) SubmitChoice(choiceIndex int, gameState *GameState) (*EventChoice, string, error)

// Get user's choice history for learning
func (gem *GeneralEventManager) GetUserChoiceHistory(eventName string) []int
```

## Usage Examples

### Daily Conversation Event

```json
{
  "name": "daily_check_in",
  "category": "conversation",
  "description": "Daily conversation about user's day",
  "trigger": "daily_check_in",
  "probability": 1.0,
  "interactive": true,
  "responses": [
    "How has your day been going? I'd love to hear about it! 😊"
  ],
  "animations": ["talking"],
  "choices": [
    {
      "text": "It's been great!",
      "effects": {"happiness": 5, "friendship": 2},
      "responses": ["That's wonderful to hear!"],
      "nextEvent": "celebrate_good_day"
    },
    {
      "text": "Pretty challenging...",
      "effects": {"friendship": 3},
      "responses": ["Want to talk about it?"],
      "nextEvent": "supportive_conversation"
    }
  ],
  "cooldown": 86400,
  "conditions": {"friendship": {"min": 10}}
}
```

### Roleplay Adventure

```json
{
  "name": "fantasy_quest",
  "category": "roleplay",
  "description": "Epic fantasy adventure with choices",
  "trigger": "start_fantasy",
  "probability": 1.0,
  "interactive": true,
  "responses": [
    "🐉 You stand before an ancient castle. What is your class, adventurer?"
  ],
  "animations": ["excited"],
  "choices": [
    {
      "text": "I am a brave knight!",
      "effects": {"courage": 5},
      "nextEvent": "knight_path"
    },
    {
      "text": "I am a cunning rogue!",
      "effects": {"stealth": 5},
      "nextEvent": "rogue_path"
    },
    {
      "text": "I am a wise mage!",
      "effects": {"wisdom": 5},
      "nextEvent": "mage_path"
    }
  ],
  "cooldown": 10800,
  "conditions": {"level": {"min": 5}}
}
```

### Mini-Game Event

```json
{
  "name": "trivia_challenge",
  "category": "game",
  "description": "Knowledge trivia game",
  "trigger": "start_trivia",
  "probability": 1.0,
  "interactive": true,
  "responses": [
    "🧠 Ready for a trivia challenge? Pick your category!"
  ],
  "animations": ["thinking"],
  "choices": [
    {
      "text": "Science questions",
      "effects": {"knowledge": 3},
      "nextEvent": "science_trivia"
    },
    {
      "text": "History questions",
      "effects": {"knowledge": 3},
      "nextEvent": "history_trivia"
    }
  ],
  "cooldown": 3600
}
```

### Humor Session

```json
{
  "name": "joke_time",
  "category": "humor",
  "description": "Share jokes and funny content",
  "trigger": "tell_jokes",
  "probability": 1.0,
  "interactive": true,
  "responses": [
    "😄 Ready for some laughs? What kind of humor do you prefer?"
  ],
  "animations": ["happy"],
  "choices": [
    {
      "text": "Tell me a pun!",
      "effects": {"happiness": 5},
      "responses": ["Why don't scientists trust atoms? Because they make up everything!"]
    },
    {
      "text": "Something silly!",
      "effects": {"happiness": 6},
      "responses": ["What do you call a bear with no teeth? A gummy bear!"]
    }
  ],
  "cooldown": 1800
}
```

## Integration with Existing Systems

### Coexistence Strategy

1. **Dialog Priority Chain**:
   - Advanced Dialog Backend (if enabled and confident)
   - Pending General Events (if active)
   - Romance Dialogs (if romance features enabled)
   - Basic Dialogs (fallback)

2. **Stat Integration**:
   - General events can affect all existing stats
   - Choice effects use same system as game interactions
   - Requirements check against current character state

3. **Animation System**:
   - Events use existing animation framework
   - Choice-specific animations override event defaults
   - Integrates with mood-based animation selection

### Backward Compatibility

- **Existing characters continue to work unchanged**
- **New `generalEvents` field is optional**
- **No changes to existing dialog or game interaction behavior**
- **Progressive enhancement - features activate when configured**

## Command Line Usage

### New Flags

```bash
# Enable general events system
go run cmd/companion/main.go -events -character assets/characters/examples/interactive_events.json

# Manually trigger specific event
go run cmd/companion/main.go -trigger-event daily_check_in -character character.json
```

### Keyboard Shortcuts

- **Ctrl+E**: Open events menu (show available general events)
- **Ctrl+R**: Quick-start random roleplay scenario
- **Ctrl+G**: Start mini-game or trivia session  
- **Ctrl+H**: Trigger humor/joke session
- **Number Keys (1-9)**: Select choice during interactive events

## Best Practices

### Event Design

1. **Meaningful Choices**: Ensure each choice has clear consequences
2. **Stat Balance**: Don't make events too rewarding or punishing
3. **Cooldown Management**: Set appropriate cooldowns to prevent spam
4. **Progressive Difficulty**: Use requirements to gate advanced content

### Content Creation

1. **Consistent Tone**: Match event personality to character
2. **Rich Descriptions**: Use emojis and descriptive text for immersion
3. **Branching Narratives**: Create follow-up events for complex stories
4. **Cultural Sensitivity**: Ensure content is appropriate for all users

### Performance

1. **Event Limits**: Keep individual character cards under 50 events
2. **Choice Limits**: Limit 2-5 choices per interactive event
3. **Memory Management**: System automatically limits choice history
4. **Cooldown Strategy**: Use longer cooldowns for complex events

## Validation

The system includes comprehensive validation:

```go
// Validate individual events
func ValidateGeneralEvent(event GeneralDialogEvent) error

// Validate choice structures  
func validateEventChoice(choice EventChoice, index int) error
```

**Validation Checks**:
- Required fields presence
- Valid category values
- Interactive events have choices
- Choice requirements reference valid stats
- Cooldown values are reasonable
- Follow-up events exist in the character card

## Error Handling

**Graceful Degradation**:
- Invalid events are skipped during initialization
- Failed event triggers return empty responses
- Invalid choices are ignored silently
- Missing follow-up events break chains gracefully

**Debug Mode**:
Enable debug logging to troubleshoot event issues:

```json
{
  "dialogBackend": {
    "debugMode": true
  }
}
```

## Migration Guide

### Adding General Events to Existing Characters

1. **Add `generalEvents` array to character card**
2. **Start with simple conversation events**
3. **Test with `-events` flag enabled**
4. **Gradually add more complex interactive scenarios**

### Converting Random Events to General Events

```json
// Old random event
{
  "randomEvents": [
    {
      "name": "daily_surprise",
      "probability": 0.1,
      "responses": ["Surprise!"]
    }
  ]
}

// New general event (user-triggered)
{
  "generalEvents": [
    {
      "name": "daily_surprise",
      "category": "humor", 
      "trigger": "surprise_me",
      "probability": 1.0,
      "interactive": false,
      "responses": ["Surprise!"],
      "cooldown": 3600
    }
  ]
}
```

This comprehensive system provides a powerful foundation for creating rich, interactive companion experiences while maintaining full backward compatibility with existing DDS functionality.
