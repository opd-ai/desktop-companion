# Bot Package

The bot package provides autonomous AI-controlled behavior for DDS characters through personality-driven decision making and seamless integration with the existing Character system.

## Overview

The BotController implements autonomous character behavior that feels natural and human-like while maintaining clean separation from the core Character implementation. It follows the project's philosophy of using standard library features and simple, maintainable patterns.

## Key Features

- **Personality-Driven Behavior**: Configurable personality traits drive all decision making
- **Natural Timing**: Human-like delays and variations prevent mechanical behavior  
- **Rate Limiting**: Prevents excessive actions that would feel unnatural
- **Network Integration**: Coordinates with multiplayer features when available
- **Zero Breaking Changes**: Integrates via interfaces without modifying existing Character code
- **High Performance**: <50ns per Update() call, suitable for 60 FPS integration

## Architecture

### Core Components

1. **BotController**: Main autonomous behavior engine
2. **BotPersonality**: Configuration for behavior characteristics
3. **BotDecision**: Represents planned actions with timing and probability
4. **CharacterController Interface**: Clean separation from Character implementation
5. **NetworkController Interface**: Optional multiplayer coordination

### Integration Pattern

The bot integrates with the existing Character.Update() cycle through a simple wrapper pattern:

```go
// In main application loop
if bot.Update() {
    // Bot performed an action - may trigger visual changes
}
character.Update() // Normal character update cycle
```

## Usage Examples

### Basic Bot Setup

```go
// Create personality configuration
personality := bot.DefaultPersonality()
personality.SocialTendencies["chattiness"] = 0.9 // More talkative
personality.InteractionRate = 3.0               // 3 actions per minute

// Create bot with character and network controllers
botController, err := bot.NewBotController(
    personality,
    characterController, // Implements CharacterController interface
    networkController,   // Implements NetworkController interface (optional)
)
if err != nil {
    log.Fatal(err)
}

// In main loop (60 FPS)
for {
    actionTaken := botController.Update()
    if actionTaken {
        // Bot performed an action - refresh UI if needed
    }
    
    time.Sleep(16 * time.Millisecond) // ~60 FPS
}
```

### Personality Customization

```go
// Shy, helpful bot
shyPersonality := bot.DefaultPersonality()
shyPersonality.SocialTendencies["chattiness"] = 0.2
shyPersonality.SocialTendencies["helpfulness"] = 0.9
shyPersonality.ResponseDelay = 5 * time.Second
shyPersonality.InteractionRate = 1.0

// Energetic, playful bot  
energeticPersonality := bot.DefaultPersonality()
energeticPersonality.SocialTendencies["playfulness"] = 0.9
energeticPersonality.EmotionalProfile["enthusiasm"] = 0.95
energeticPersonality.ResponseDelay = 1 * time.Second
energeticPersonality.InteractionRate = 4.0
```

### Monitoring and Control

```go
// Get performance statistics
stats := botController.GetStats()
fmt.Printf("Actions in history: %d\n", stats["actionsInHistory"])
fmt.Printf("Time since last action: %v\n", stats["timeSinceLastAction"])

// Get recent action history for debugging
history := botController.GetActionHistory()
for _, action := range history {
    fmt.Printf("Action: %s, Probability: %.2f\n", action.Action, action.Probability)
}

// Temporarily disable bot
botController.Disable()
// ... later ...
botController.Enable()
```

## Personality Configuration

### Social Tendencies

- **chattiness** (0.0-1.0): How likely to initiate network conversations
- **helpfulness** (0.0-1.0): How likely to perform caring actions (feeding, etc.)
- **playfulness** (0.0-1.0): How likely to perform interactive actions (clicking, playing)
- **curiosity** (0.0-1.0): How likely to explore new interactions

### Emotional Profile

- **empathy** (0.0-1.0): How much character state affects bot behavior
- **assertiveness** (0.0-1.0): How quickly bot acts vs. waiting
- **patience** (0.0-1.0): How long bot waits between actions
- **enthusiasm** (0.0-1.0): How likely to perform energetic actions

### Timing Configuration

- **ResponseDelay**: Average time between bot actions
- **InteractionRate**: Target actions per minute (0.1-10.0)
- **Attention**: How quickly bot notices events (0.0-1.0)
- **MaxActionsPerMinute**: Hard limit to prevent spam (1-30)
- **MinTimeBetweenSame**: Minimum seconds between same action type (1-300)

## Action Types

The bot can perform these autonomous actions:

- **click**: Basic character interaction (HandleClick)
- **feed**: Care action (HandleRightClick) - considers character hunger
- **play**: Energy action (HandleDoubleClick) - considers character energy/mood
- **chat**: Network message to peers (requires NetworkController)
- **wait**: Deliberate pause in activity

## Performance Characteristics

- **Memory Usage**: <5KB per bot instance
- **CPU Usage**: <50ns per Update() call
- **Action Generation**: ~2μs per decision cycle
- **Thread Safety**: Full concurrent access support
- **Integration Cost**: Zero overhead when bot is disabled

## Error Handling

The bot package follows Go's standard error handling patterns:

- All initialization functions return explicit errors
- Invalid personality configurations are rejected at creation time
- Network errors are handled gracefully without crashing
- Missing controllers are handled via nil checks

## Testing

Comprehensive test suite with >75% coverage includes:

- Unit tests for all public methods
- Mock implementations for clean testing
- Personality validation edge cases
- Rate limiting and timing behavior
- Concurrent access safety
- Performance benchmarks

Run tests:
```bash
go test ./lib/bot/... -v -cover
```

Run benchmarks:
```bash
go test ./lib/bot/... -bench=. -benchmem
```

## Integration with Existing Systems

### Character System Integration

The bot uses the CharacterController interface to interact with characters:

```go
type CharacterController interface {
    HandleClick() string
    HandleRightClick() string  
    HandleDoubleClick() string
    GetCurrentState() string
    GetLastInteractionTime() time.Time
    GetStats() map[string]float64
    GetMood() float64
    IsGameMode() bool
}
```

### Network System Integration

For multiplayer features, the bot uses the NetworkController interface:

```go
type NetworkController interface {
    GetPeerCount() int
    GetPeerIDs() []string
    SendMessage(peerID string, message interface{}) error
    IsNetworkEnabled() bool
}
```

### Character Card Configuration

Bot personality can be configured in character cards:

```json
{
  "name": "Autonomous Bot Character",
  "botPersonality": {
    "responseDelay": "3s",
    "interactionRate": 2.0,
    "attention": 0.7,
    "socialTendencies": {
      "chattiness": 0.8,
      "helpfulness": 0.9,
      "playfulness": 0.6
    },
    "emotionalProfile": {
      "empathy": 0.8,
      "assertiveness": 0.4,
      "enthusiasm": 0.7
    },
    "maxActionsPerMinute": 5,
    "minTimeBetweenSame": 10,
    "preferredActions": ["click", "chat", "feed"]
  }
}
```

## Design Philosophy

The bot package follows the DDS project's core principles:

1. **Library-First Development**: Uses only Go standard library
2. **Interface-Based Design**: Clean separation via interfaces
3. **Minimal Custom Code**: Simple, maintainable implementation
4. **Zero Breaking Changes**: Integrates without modifying existing code
5. **Performance First**: Optimized for 60 FPS real-time operation

## Future Extensions

The bot system is designed for easy extension:

- Additional action types can be added to the decision engine
- New personality traits can be incorporated into behavior algorithms
- Learning systems can be built on top of the action history
- Complex group behaviors can be implemented via network coordination

The simple, interface-based design ensures these extensions won't require changes to the core bot logic.
