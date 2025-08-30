# Group Events Documentation

## Overview

The Group Events system enables multi-character scenarios and collaborative activities in DDS multiplayer mode. This system supports collaborative mini-games, group decision events, and multi-character scenarios that can involve 2-8 participants.

## Architecture

### Core Components

- **GroupEventManager**: Main coordinator for group events
- **GroupEventTemplate**: Defines the structure and phases of group events
- **GroupEvent**: Active instance of a running group event
- **EventPhase**: Individual phases within a group event
- **EventChoice**: Voting options within each phase

### Key Features

- **Multi-character scenarios**: Story building, roleplay, group conversations
- **Collaborative mini-games**: Trivia, puzzles, team challenges
- **Group decision events**: Voting on activities, choices, preferences
- **Real-time synchronization**: All participants see updates immediately
- **Conflict resolution**: Handles simultaneous actions gracefully
- **Event history**: Tracks completed events and participant scores

## Implementation Details

### Standard Library Usage

Following the project's "library-first" philosophy, the Group Events system uses only Go standard library:

- `encoding/json`: Message serialization and template parsing
- `sync`: Concurrent access protection with RWMutex
- `time`: Event timing, cooldowns, and duration tracking
- `fmt`: Error formatting and string operations
- `math/rand`: Session ID generation

### Performance Characteristics

Benchmarked performance on AMD EPYC 7763:
- **StartGroupEvent**: 4.4μs per operation
- **SubmitVote**: 287ns per operation
- **Test Coverage**: 81.2% (exceeds 80% requirement)

### Interface Design

Uses interface-based design for testability:

```go
type GroupNetworkManagerInterface interface {
    BroadcastMessage(msgType string, payload []byte) error
    SendMessage(msgType string, payload []byte, targetPeerID string) error
    RegisterMessageHandler(msgType string, handler func([]byte, string) error)
    GetConnectedPeers() []string
    GetLocalPeerID() string
}
```

## Usage Examples

### Basic Group Event Creation

```go
// Create group event manager
networkManager := // ... existing network manager
templates := loadGroupEventTemplates()
gem := NewGroupEventManager(networkManager, templates)

// Start a group event
sessionID, err := gem.StartGroupEvent("trivia_game", "initiator_peer_id")
if err != nil {
    log.Printf("Failed to start group event: %v", err)
    return
}

// Other peers join
err = gem.JoinGroupEvent(sessionID, "participant_peer_id")
if err != nil {
    log.Printf("Failed to join event: %v", err)
    return
}

// Participants vote
err = gem.SubmitVote(sessionID, "participant_peer_id", "choice_a")
if err != nil {
    log.Printf("Failed to submit vote: %v", err)
    return
}
```

### Group Event Template Structure

```json
{
  "id": "trivia_challenge",
  "name": "Group Trivia Challenge",
  "description": "Collaborative trivia game for 2-6 players",
  "category": "minigame",
  "minParticipants": 2,
  "maxParticipants": 6,
  "estimatedTime": "5m",
  "phases": [
    {
      "name": "question1",
      "description": "First trivia question",
      "type": "choice",
      "duration": "30s",
      "minVotes": 2,
      "autoAdvance": true,
      "choices": [
        {"id": "a", "text": "Answer A", "points": 10},
        {"id": "b", "text": "Answer B", "points": 5},
        {"id": "c", "text": "Answer C", "points": 0}
      ]
    }
  ]
}
```

## Event Categories

### Minigames
- **Trivia Challenges**: Question-and-answer sessions with scoring
- **Word Games**: Collaborative word building or guessing games
- **Puzzles**: Group problem-solving activities

### Scenarios
- **Story Building**: Collaborative narrative creation
- **Roleplay**: Character-based interaction scenarios
- **Adventure**: Choose-your-own-adventure style events

### Decision Events
- **Group Planning**: Deciding on activities or preferences
- **Voting**: Democratic decision making on topics
- **Preferences**: Learning about group member preferences

## Integration with Existing Systems

### Network Layer Integration

The Group Events system integrates seamlessly with the existing network infrastructure:

- Uses existing message routing and delivery
- Leverages Ed25519 signature verification
- Follows existing JSON message protocol
- Compatible with peer discovery system

### Character System Integration

Group events can be triggered by:
- Bot personality systems (autonomous bots can start events)
- User interactions (keyboard shortcuts, menu selections)
- Network events (peer joining/leaving can trigger activities)
- Timer-based events (scheduled group activities)

### Dialog System Integration

Group events can trigger:
- Character dialog responses based on event outcomes
- Mood changes based on group activity success
- Relationship stat changes between participants
- Memory formation for future interactions

## Error Handling

The system handles common error scenarios gracefully:

- **Network disconnections**: Events continue with remaining participants
- **Invalid votes**: Rejected with clear error messages
- **Phase timeouts**: Automatic advancement after duration expires
- **Participant limits**: Clear feedback when events are full
- **Template validation**: Comprehensive validation at startup

## Testing Strategy

### Unit Tests (81.2% coverage)

- **Happy path scenarios**: Normal event flow from start to completion
- **Error conditions**: Invalid inputs, network failures, edge cases
- **Concurrent access**: Race condition testing with multiple goroutines
- **Message handling**: Network message processing and validation
- **Performance**: Benchmarks for critical operations

### Integration Tests

- **Network integration**: Real network manager integration
- **Character integration**: Bot behavior with group events
- **UI integration**: User interface for group event management
- **End-to-end**: Complete multiplayer scenarios

## Configuration

### Character Card Integration

Characters can specify group event templates in their configuration:

```json
{
  "name": "Group Event Host",
  "multiplayer": {
    "enabled": true,
    "botCapable": true,
    "networkID": "group_host_v1"
  },
  "groupEvents": [
    {
      "id": "icebreaker",
      "name": "Icebreaker Questions",
      "category": "scenario",
      "minParticipants": 3,
      "maxParticipants": 8,
      "phases": [...]
    }
  ]
}
```

### Network Message Types

The system introduces three new message types:

- `group_event_invite`: Invitation to join a group event
- `group_event_action`: Vote submissions and participant actions
- `group_event_update`: Phase changes and event status updates

## Performance Optimization

### Memory Management

- Efficient use of maps for participant tracking
- Cleanup of completed events from memory
- Garbage-friendly struct design with value types
- Minimal allocation during vote processing

### Network Efficiency

- JSON message compression opportunities
- Batched state updates when possible
- Efficient peer filtering for relevant messages
- Minimal network chatter during quiet periods

## Future Enhancements

### Planned Features

- **Voice chat integration**: Audio communication during group events
- **Custom event creation**: User-defined event templates
- **Achievement system**: Rewards for group event participation
- **Advanced scoring**: Complex scoring algorithms for competitive events

### Extensibility

The system is designed for easy extension:
- Plugin architecture for new event types
- Configurable scoring systems
- Custom phase types beyond voting
- Integration hooks for external event sources

## Troubleshooting

### Common Issues

1. **Event won't start**: Check participant count meets minimum requirements
2. **Votes not registering**: Verify choice ID matches template choices
3. **Phase not advancing**: Check minVotes and autoAdvance settings
4. **Network sync issues**: Verify all peers have compatible network IDs

### Debug Information

Enable debug logging to see:
- Event lifecycle state changes
- Network message flow
- Participant join/leave events
- Vote counting and phase advancement logic

### Performance Monitoring

Monitor these metrics:
- Event start latency (target: <5ms)
- Vote processing time (target: <1ms)
- Network message delivery time (target: <50ms local)
- Memory usage during large events (target: <10MB per event)

## Best Practices

### Event Design

- Keep phases short (30-60 seconds) for engagement
- Provide clear, concise choice descriptions
- Balance scoring to avoid dominant strategies
- Test events with minimum and maximum participants

### Network Considerations

- Design events to handle participant disconnections
- Avoid requiring all participants for advancement when possible
- Provide clear feedback on event status to all participants
- Handle network partitions gracefully

### User Experience

- Clear visual feedback on voting status
- Timeout warnings before phase advancement
- Score displays that encourage participation
- Easy-to-understand event progression

## Conclusion

The Group Events system provides a robust foundation for multiplayer interactive experiences in DDS. By leveraging Go's standard library and following the project's interface-based design principles, it integrates seamlessly with existing systems while providing extensible functionality for future enhancements.

The system's performance characteristics (4.4μs event creation, 287ns vote processing) and high test coverage (81.2%) ensure production readiness while maintaining the project's commitment to simplicity and maintainability.
