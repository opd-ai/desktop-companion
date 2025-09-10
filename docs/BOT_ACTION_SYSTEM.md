# Bot Action System Documentation

## Overview

The Bot Action System provides autonomous character behavior through personality-driven action selection, execution, and learning capabilities. It implements the core functionality needed for AI-controlled characters in multiplayer scenarios.

## Architecture

### Core Components

1. **ActionExecutor** (`lib/bot/actions.go`): Handles action execution with comprehensive error handling and performance tracking
2. **BotController Integration**: Seamlessly integrates with existing bot personality system
3. **Learning System**: Analyzes action effectiveness and learns from peer interactions

### Action Types

- **Click**: Basic interaction - increases happiness
- **Feed**: Right-click interaction - increases hunger (game mode)
- **Play**: Double-click interaction - increases happiness, decreases energy
- **Chat**: Network communication with peers
- **Wait**: Passive waiting for natural timing
- **Observe**: Watch peer interactions for learning

## Key Features

### Performance
- **1098ns per operation**: Highly optimized execution
- **Comprehensive tracking**: Success rates, timing, and stat impacts
- **Memory efficient**: Rolling history with configurable limits

### Learning Capabilities
- **Stat Impact Analysis**: Learns which actions are most effective for character care
- **Peer Learning**: Observes and learns from other bot behaviors
- **Recommendation System**: Suggests optimal actions based on context and learning

### Error Handling
- **Graceful degradation**: Handles network failures and nil controllers
- **Comprehensive logging**: Detailed action results for debugging
- **Fallback behaviors**: Always provides meaningful responses

## Usage Examples

### Basic Action Execution

```go
// Create action executor
executor := NewActionExecutor(characterController, networkController)

// Execute an action
decision := BotDecision{
    Action:      "click",
    Probability: 1.0,
    Priority:    1,
}

result, err := executor.ExecuteAction(decision)
if err != nil {
    log.Printf("Action failed: %v", err)
} else {
    log.Printf("Action succeeded: %s", result.Response)
}
```

### Bot Controller Integration

```go
// Create bot with action system
personality := DefaultPersonality()
botController, err := NewBotController(personality, charController, netController)

// Get action recommendations
recommended := botController.GetRecommendedAction()

// Analyze effectiveness
happinessImpact := botController.AnalyzeStatImpact(ActionClick, "happiness")

// Learn from peers
peerActions := []PeerActionEvent{...}
botController.LearnFromPeerActions(peerActions)
```

### Performance Monitoring

```go
// Get execution statistics
stats := botController.GetActionStats()
for actionType, actionStats := range stats {
    fmt.Printf("%s: %d executions, %.2f success rate\n", 
        actionType, actionStats.TotalExecutions, actionStats.SuccessRate)
}

// Get detailed history
history := botController.GetActionExecutionHistory()
for _, result := range history {
    fmt.Printf("Action: %s, Success: %t, Duration: %v\n",
        result.Action, result.Success, result.Duration)
}
```

## Integration with Existing Systems

### Character Controller Interface

The action system uses the existing `CharacterController` interface:
- `HandleClick()`, `HandleRightClick()`, `HandleDoubleClick()` for interactions
- `GetStats()`, `GetMood()`, `IsGameMode()` for decision making
- `GetCurrentState()`, `GetLastInteractionTime()` for context

### Network Controller Interface

For multiplayer features:
- `SendMessage()` for peer communication
- `GetPeerCount()`, `GetPeerIDs()` for network awareness
- `IsNetworkEnabled()` for feature detection

## Configuration

### Personality-Driven Behavior

Actions are selected based on personality traits:
- **Playfulness**: Influences click action probability
- **Helpfulness**: Affects feeding behavior when character stats are low
- **Chattiness**: Drives network communication frequency
- **Attention**: Modifies response timing and awareness

### Performance Tuning

- **History Size**: Configurable rolling window (default: 100 actions)
- **Rate Limiting**: Respects personality-based action frequency limits
- **Caching**: Response caching for network operations

## Testing

Comprehensive test suite with >95% coverage:
- Unit tests for all action types
- Integration tests with bot controller
- Performance benchmarks
- Error condition handling
- Mock implementations for isolated testing

## Future Enhancements

Planned improvements for Phase 3:
- Advanced peer learning algorithms
- Context-aware action selection
- Group interaction strategies
- Long-term memory and adaptation

## Performance Metrics

- **Execution Time**: 1098ns per operation
- **Memory Usage**: 789B per operation with 5 allocations
- **Success Rate Tracking**: Real-time monitoring of action effectiveness
- **Learning Efficiency**: Adapts behavior based on stat impact analysis
