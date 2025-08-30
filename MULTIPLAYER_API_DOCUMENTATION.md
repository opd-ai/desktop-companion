# DDS Multiplayer API Documentation

## Table of Contents

1. [Overview](#overview)
2. [Core Interfaces](#core-interfaces)
3. [Network Manager API](#network-manager-api)
4. [Group Events API](#group-events-api)
5. [Bot Controller API](#bot-controller-api)
6. [State Synchronization API](#state-synchronization-api)
7. [Protocol Specification](#protocol-specification)
8. [Integration Examples](#integration-examples)
9. [Error Handling](#error-handling)
10. [Performance Considerations](#performance-considerations)

## Overview

The DDS Multiplayer API provides a comprehensive interface for peer-to-peer networking, autonomous bot behavior, group events, and real-time state synchronization. The API follows Go's interface-based design principles and uses only standard library dependencies.

### Architecture Principles

- **Interface-driven**: All major components expose clear interfaces for testability
- **Standard library only**: No external dependencies beyond Go standard library
- **Backward compatible**: All existing Character functionality preserved
- **Thread-safe**: Concurrent access protected with appropriate synchronization
- **Performance optimized**: Sub-microsecond operation latencies for real-time use

### Core Components

| Component | Purpose | Performance |
|-----------|---------|-------------|
| `NetworkManager` | Peer discovery and communication | 2-5s discovery time |
| `GroupEventManager` | Multi-character scenarios | 4.4μs event creation |
| `BotController` | Autonomous AI behavior | 49ns per update |
| `StateSynchronizer` | Real-time state sync | 3-4μs per state update |
| `NetworkDialogBackend` | Network-aware dialog | 232ns per operation |

## Core Interfaces

### NetworkManagerInterface

Primary interface for all network operations.

```go
type NetworkManagerInterface interface {
    // Peer management
    GetPeerCount() int
    GetPeers() []Peer
    GetConnectedPeers() []string
    GetLocalPeerID() string
    GetNetworkID() string
    
    // Message handling
    SendMessage(msgType MessageType, payload []byte, targetPeerID string) error
    BroadcastMessage(msgType MessageType, payload []byte) error
    RegisterMessageHandler(msgType MessageType, handler MessageHandler)
    
    // Lifecycle management
    Start() error
    Stop() error
    IsRunning() bool
}
```

#### GetPeerCount

Returns the number of currently connected peers.

**Returns**: `int` - Number of connected peers (0-15)

**Example**:
```go
peerCount := networkManager.GetPeerCount()
fmt.Printf("Connected to %d peers\n", peerCount)
```

#### GetPeers

Returns detailed information about all connected peers.

**Returns**: `[]Peer` - Array of peer information structures

**Example**:
```go
peers := networkManager.GetPeers()
for _, peer := range peers {
    fmt.Printf("Peer %s: %s\n", peer.ID, peer.CharacterID)
}
```

#### SendMessage

Sends a message to a specific peer.

**Parameters**:
- `msgType MessageType`: Type of message (see Protocol Specification)
- `payload []byte`: JSON-encoded message payload
- `targetPeerID string`: Target peer identifier

**Returns**: `error` - nil on success, error on failure

**Example**:
```go
payload, _ := json.Marshal(map[string]interface{}{
    "action": "click",
    "timestamp": time.Now(),
})
err := networkManager.SendMessage(MessageTypeCharacterAction, payload, "peer123")
if err != nil {
    log.Printf("Failed to send message: %v", err)
}
```

#### BroadcastMessage

Sends a message to all connected peers.

**Parameters**:
- `msgType MessageType`: Type of message
- `payload []byte`: JSON-encoded message payload

**Returns**: `error` - nil on success, error on failure

**Example**:
```go
payload, _ := json.Marshal(map[string]interface{}{
    "event": "peer_status_update",
    "status": "active",
})
err := networkManager.BroadcastMessage(MessageTypePeerUpdate, payload)
```

### CharacterControllerInterface

Interface for character state and behavior management.

```go
type CharacterControllerInterface interface {
    // Character actions
    HandleClick() (string, error)
    HandleFeed() (string, error)
    HandlePlay() (string, error)
    
    // State management
    GetStats() *character.Stats
    UpdateStats(stats *character.Stats) error
    
    // Animation control
    SetAnimation(name string) error
    GetCurrentAnimation() string
    
    // Persistence
    Save() error
    Load() error
}
```

### GroupNetworkManagerInterface

Specialized interface for group event network operations.

```go
type GroupNetworkManagerInterface interface {
    BroadcastMessage(msgType string, payload []byte) error
    SendMessage(msgType string, payload []byte, targetPeerID string) error
    RegisterMessageHandler(msgType string, handler func([]byte, string) error)
    GetConnectedPeers() []string
    GetLocalPeerID() string
}
```

## Network Manager API

### Initialization

```go
// Create new network manager
config := NetworkConfig{
    NetworkID:     "my_network",
    DiscoveryPort: 8080,
    MaxPeers:      8,
}

networkManager, err := NewNetworkManager(config)
if err != nil {
    log.Fatal("Failed to create network manager:", err)
}

// Start networking
err = networkManager.Start()
if err != nil {
    log.Fatal("Failed to start networking:", err)
}
defer networkManager.Stop()
```

### Message Handling

```go
// Register message handler
networkManager.RegisterMessageHandler(MessageTypeCharacterAction, func(message Message, peer *Peer) error {
    var actionData struct {
        Action    string    `json:"action"`
        Timestamp time.Time `json:"timestamp"`
    }
    
    if err := json.Unmarshal(message.Payload, &actionData); err != nil {
        return fmt.Errorf("failed to unmarshal action: %w", err)
    }
    
    fmt.Printf("Peer %s performed action: %s\n", peer.ID, actionData.Action)
    return nil
})
```

### Peer Discovery

```go
// Monitor peer events
networkManager.RegisterMessageHandler(MessageTypePeerDiscovery, func(message Message, peer *Peer) error {
    fmt.Printf("Discovered peer: %s (%s)\n", peer.ID, peer.CharacterID)
    
    // Send welcome message
    welcomePayload, _ := json.Marshal(map[string]interface{}{
        "message": "Welcome to the network!",
    })
    
    return networkManager.SendMessage(MessageTypeCharacterAction, welcomePayload, peer.ID)
})
```

## Group Events API

### GroupEventManager

Main interface for managing group events and collaborative activities.

```go
type GroupEventManager struct {
    // Public methods
    StartGroupEvent(templateID string, initiatorID string) (string, error)
    JoinGroupEvent(sessionID, participantID string) error
    SubmitVote(sessionID, participantID, choiceID string) error
    GetActiveEvents() []GroupEvent
    GetEventTemplates() []GroupEventTemplate
    GetParticipantHistory(participantID string) []CompletedEvent
}
```

#### StartGroupEvent

Initiates a new group event based on a template.

**Parameters**:
- `templateID string`: Identifier of the event template to use
- `initiatorID string`: Peer ID of the event initiator

**Returns**:
- `string`: Unique session ID for the event
- `error`: nil on success, error describing failure

**Example**:
```go
sessionID, err := groupEventManager.StartGroupEvent("trivia_game", myPeerID)
if err != nil {
    log.Printf("Failed to start group event: %v", err)
    return
}
fmt.Printf("Started group event with session ID: %s\n", sessionID)
```

#### JoinGroupEvent

Allows a participant to join an active group event.

**Parameters**:
- `sessionID string`: Session ID of the event to join
- `participantID string`: Peer ID of the joining participant

**Returns**: `error` - nil on success, error on failure

**Example**:
```go
err := groupEventManager.JoinGroupEvent(sessionID, myPeerID)
if err != nil {
    log.Printf("Failed to join group event: %v", err)
    return
}
fmt.Println("Successfully joined group event")
```

#### SubmitVote

Submits a vote for the current phase of a group event.

**Parameters**:
- `sessionID string`: Session ID of the event
- `participantID string`: Peer ID of the voting participant  
- `choiceID string`: ID of the selected choice

**Returns**: `error` - nil on success, error on failure

**Example**:
```go
err := groupEventManager.SubmitVote(sessionID, myPeerID, "choice_a")
if err != nil {
    log.Printf("Failed to submit vote: %v", err)
    return
}
fmt.Println("Vote submitted successfully")
```

### Group Event Templates

Define the structure and phases of group events.

```go
type GroupEventTemplate struct {
    ID              string        `json:"id"`
    Name            string        `json:"name"`
    Description     string        `json:"description"`
    Category        string        `json:"category"`        // "scenario", "minigame", "decision"
    MinParticipants int           `json:"minParticipants"` // 2-8 participants
    MaxParticipants int           `json:"maxParticipants"`
    EstimatedTime   time.Duration `json:"estimatedTime"`
    Phases          []EventPhase  `json:"phases"`
}

type EventPhase struct {
    Name        string        `json:"name"`
    Description string        `json:"description"`
    Type        string        `json:"type"`        // "intro", "choice", "vote", "result"
    Duration    time.Duration `json:"duration"`    // Maximum phase duration
    Choices     []EventChoice `json:"choices"`     // Available choices
    MinVotes    int           `json:"minVotes"`    // Minimum votes to proceed
    AutoAdvance bool          `json:"autoAdvance"` // Auto-advance when conditions met
}

type EventChoice struct {
    ID          string `json:"id"`
    Text        string `json:"text"`
    Description string `json:"description,omitempty"`
    Points      int    `json:"points"` // Points awarded for this choice
}
```

**Example Template**:
```json
{
  "id": "simple_trivia",
  "name": "Quick Trivia",
  "description": "A simple trivia game for 2-4 players",
  "category": "minigame",
  "minParticipants": 2,
  "maxParticipants": 4,
  "estimatedTime": "3m",
  "phases": [
    {
      "name": "question1",
      "description": "What's the capital of France?",
      "type": "choice",
      "duration": "30s",
      "minVotes": 2,
      "autoAdvance": true,
      "choices": [
        {"id": "paris", "text": "Paris", "points": 10},
        {"id": "london", "text": "London", "points": 0},
        {"id": "berlin", "text": "Berlin", "points": 0}
      ]
    }
  ]
}
```

## Bot Controller API

### BotController

Manages autonomous AI behavior for character bots.

```go
type BotController struct {
    // Public methods
    Update() error
    SetPersonality(personality BotPersonality) error
    GetPersonality() BotPersonality
    EnableLearning(enabled bool)
    GetStats() BotStats
}
```

#### Update

Called regularly (typically 60 FPS) to update bot behavior and decision making.

**Returns**: `error` - nil on success, error on failure

**Performance**: 49ns per call (optimized for real-time use)

**Example**:
```go
// In main update loop
err := botController.Update()
if err != nil {
    log.Printf("Bot update failed: %v", err)
}
```

#### SetPersonality

Configures the bot's personality traits.

**Parameters**:
- `personality BotPersonality`: Personality configuration

**Returns**: `error` - nil on success, error on validation failure

**Example**:
```go
personality := BotPersonality{
    Chattiness:    0.8,
    Helpfulness:  0.7,
    Playfulness:  0.6,
    ResponseDelay: "2-5s",
}

err := botController.SetPersonality(personality)
if err != nil {
    log.Printf("Failed to set personality: %v", err)
}
```

### BotPersonality

Configuration structure for bot personality traits.

```go
type BotPersonality struct {
    Chattiness    float64 `json:"chattiness"`    // 0.0-1.0: Communication frequency
    Helpfulness  float64 `json:"helpfulness"`  // 0.0-1.0: Assistance likelihood  
    Playfulness  float64 `json:"playfulness"`  // 0.0-1.0: Fun activity suggestion
    ResponseDelay string  `json:"responseDelay"` // e.g., "1-3s", "500ms-2s"
}
```

### Action Execution

```go
type ActionExecutor struct {
    // Public methods
    ExecuteAction(action BotAction) error
    GetSupportedActions() []string
    GetActionHistory() []ExecutedAction
    GetLearningRecommendations() []ActionRecommendation
}
```

**Supported Actions**:
- `click`: Click on character
- `feed`: Feed character (game mode)
- `play`: Play with character (game mode)
- `chat`: Send chat message
- `wait`: Pause for specified duration
- `observe`: Watch other characters without action

**Example**:
```go
action := BotAction{
    Type:     "click",
    Target:   "character123",
    Duration: 0,
    Message:  "",
}

err := actionExecutor.ExecuteAction(action)
if err != nil {
    log.Printf("Failed to execute action: %v", err)
}
```

## State Synchronization API

### StateSynchronizer

Manages real-time character state synchronization across peers.

```go
type StateSynchronizer struct {
    // Public methods
    SynchronizeState(state CharacterState) error
    GetPeerState(peerID string) (*CharacterState, error)
    SetConflictResolution(strategy ConflictResolutionStrategy)
    GetSyncStats() SyncStatistics
}
```

#### SynchronizeState

Synchronizes character state with connected peers.

**Parameters**:
- `state CharacterState`: Current character state to synchronize

**Returns**: `error` - nil on success, error on failure

**Performance**: 3-4μs per operation

**Example**:
```go
state := CharacterState{
    Position:  Position{X: 100, Y: 200},
    Animation: "happy",
    Stats: &Stats{
        Hunger:    80,
        Happiness: 90,
        Health:    85,
        Energy:    75,
    },
    Timestamp: time.Now(),
}

err := stateSynchronizer.SynchronizeState(state)
if err != nil {
    log.Printf("Failed to sync state: %v", err)
}
```

#### GetPeerState

Retrieves the current state of a specific peer.

**Parameters**:
- `peerID string`: Peer identifier

**Returns**:
- `*CharacterState`: Peer's current state (nil if not available)
- `error`: nil on success, error on failure

**Example**:
```go
peerState, err := stateSynchronizer.GetPeerState("peer123")
if err != nil {
    log.Printf("Failed to get peer state: %v", err)
    return
}

if peerState != nil {
    fmt.Printf("Peer position: %d, %d\n", peerState.Position.X, peerState.Position.Y)
}
```

### Conflict Resolution

```go
type ConflictResolutionStrategy string

const (
    ConflictResolutionTimestamp ConflictResolutionStrategy = "timestamp"
    ConflictResolutionPriority  ConflictResolutionStrategy = "priority"
    ConflictResolutionLastWrite ConflictResolutionStrategy = "last_write"
)
```

**Strategies**:
- **Timestamp**: Use the change with the latest timestamp
- **Priority**: Use peer priority to resolve conflicts
- **LastWrite**: Always accept the most recent write

**Example**:
```go
stateSynchronizer.SetConflictResolution(ConflictResolutionTimestamp)
```

## Protocol Specification

### Message Types

```go
type MessageType string

const (
    MessageTypePeerDiscovery    MessageType = "peer_discovery"
    MessageTypeCharacterAction  MessageType = "character_action"
    MessageTypeStateSync        MessageType = "state_sync"
    MessageTypeGroupEvent       MessageType = "group_event"
    MessageTypePeerUpdate      MessageType = "peer_update"
    MessageTypeNetworkEvent    MessageType = "network_event"
)
```

### Message Structure

All network messages follow this structure:

```go
type Message struct {
    Type      MessageType            `json:"type"`
    Sender    string                 `json:"sender"`
    Payload   []byte                 `json:"payload"`
    Timestamp time.Time              `json:"timestamp"`
    Signature []byte                 `json:"signature"`
}
```

### Character Action Message

```json
{
  "type": "character_action",
  "sender": "peer123",
  "payload": {
    "action": "click",
    "target": "character456",
    "response": "Hello there!",
    "animation": "happy"
  },
  "timestamp": "2025-08-30T12:34:56Z",
  "signature": "..."
}
```

### State Sync Message

```json
{
  "type": "state_sync",
  "sender": "peer123", 
  "payload": {
    "characterState": {
      "position": {"x": 100, "y": 200},
      "animation": "idle",
      "stats": {
        "hunger": 80,
        "happiness": 90,
        "health": 85,
        "energy": 75
      }
    },
    "checksum": "sha256:abc123..."
  },
  "timestamp": "2025-08-30T12:34:56Z",
  "signature": "..."
}
```

### Group Event Message

```json
{
  "type": "group_event",
  "sender": "peer123",
  "payload": {
    "action": "vote",
    "sessionId": "group_12345",
    "choiceId": "choice_a",
    "participantId": "peer123"
  },
  "timestamp": "2025-08-30T12:34:56Z",
  "signature": "..."
}
```

## Integration Examples

### Basic Multiplayer Character

```go
package main

import (
    "log"
    "desktop-companion/internal/character"
    "desktop-companion/internal/network"
)

func main() {
    // Load character
    char, err := character.LoadFromFile("multiplayer_character.json")
    if err != nil {
        log.Fatal("Failed to load character:", err)
    }

    // Create network manager
    networkManager, err := network.NewNetworkManager(network.Config{
        NetworkID:     char.Multiplayer.NetworkID,
        DiscoveryPort: char.Multiplayer.DiscoveryPort,
        MaxPeers:      char.Multiplayer.MaxPeers,
    })
    if err != nil {
        log.Fatal("Failed to create network manager:", err)
    }

    // Start networking
    if err := networkManager.Start(); err != nil {
        log.Fatal("Failed to start networking:", err)
    }
    defer networkManager.Stop()

    // Create multiplayer character wrapper
    multiplayerChar := character.NewMultiplayerCharacter(char, networkManager)

    // Set up message handlers
    networkManager.RegisterMessageHandler(network.MessageTypeCharacterAction, 
        multiplayerChar.HandleNetworkAction)

    // Main application loop
    for {
        // Handle user interactions
        multiplayerChar.Update()
        
        // Sleep or handle events
        time.Sleep(16 * time.Millisecond) // ~60 FPS
    }
}
```

### Bot with Group Events

```go
func createGroupEventBot() {
    // Load bot character
    char, err := character.LoadFromFile("group_bot.json")
    if err != nil {
        log.Fatal("Failed to load character:", err)
    }

    // Create network manager
    networkManager, err := network.NewNetworkManager(network.Config{
        NetworkID: char.Multiplayer.NetworkID,
    })
    if err != nil {
        log.Fatal("Failed to create network manager:", err)
    }

    // Create bot controller
    botController, err := bot.NewBotController(char, networkManager)
    if err != nil {
        log.Fatal("Failed to create bot controller:", err)
    }

    // Create group event manager
    groupEventManager := network.NewGroupEventManager(networkManager, char.GroupEvents)

    // Set up bot to start group events
    networkManager.RegisterMessageHandler(network.MessageTypePeerUpdate, 
        func(message network.Message, peer *network.Peer) error {
            // When new peer joins, consider starting a group event
            if botController.ShouldStartGroupEvent() {
                sessionID, err := groupEventManager.StartGroupEvent("icebreaker", networkManager.GetLocalPeerID())
                if err != nil {
                    log.Printf("Failed to start group event: %v", err)
                    return err
                }
                log.Printf("Started group event: %s", sessionID)
            }
            return nil
        })

    // Start networking
    networkManager.Start()
    defer networkManager.Stop()

    // Main bot loop
    for {
        botController.Update()
        time.Sleep(16 * time.Millisecond)
    }
}
```

### State Synchronization

```go
func setupStateSynchronization(char *character.Character, networkManager *network.NetworkManager) {
    // Create state synchronizer
    stateSynchronizer := network.NewStateSynchronizer(networkManager, networkManager)

    // Set conflict resolution strategy
    stateSynchronizer.SetConflictResolution(network.ConflictResolutionTimestamp)

    // Register for character state changes
    char.OnStateChange(func(newState character.State) {
        characterState := network.CharacterState{
            Position:  network.Position{X: newState.X, Y: newState.Y},
            Animation: newState.CurrentAnimation,
            Stats:     newState.Stats,
            Timestamp: time.Now(),
        }

        if err := stateSynchronizer.SynchronizeState(characterState); err != nil {
            log.Printf("Failed to synchronize state: %v", err)
        }
    })

    // Handle incoming state updates
    networkManager.RegisterMessageHandler(network.MessageTypeStateSync,
        func(message network.Message, peer *network.Peer) error {
            peerState, err := stateSynchronizer.GetPeerState(peer.ID)
            if err != nil {
                return err
            }

            if peerState != nil {
                // Update UI to show peer character position/animation
                updatePeerCharacterDisplay(peer.ID, *peerState)
            }
            return nil
        })
}
```

## Error Handling

### Network Errors

```go
// Handle network connection errors
networkManager.RegisterErrorHandler(func(err error, context string) {
    switch {
    case errors.Is(err, network.ErrPeerDisconnected):
        log.Printf("Peer disconnected: %s", context)
        // Gracefully handle peer disconnection
        
    case errors.Is(err, network.ErrMessageTooLarge):
        log.Printf("Message too large: %s", context)
        // Split message or reduce payload size
        
    case errors.Is(err, network.ErrSignatureVerification):
        log.Printf("Security error: %s", context)
        // Reject message, possibly block peer
        
    default:
        log.Printf("Network error: %v (context: %s)", err, context)
    }
})
```

### Common Error Types

```go
var (
    ErrPeerNotFound           = errors.New("peer not found")
    ErrInvalidMessage        = errors.New("invalid message format")
    ErrSignatureVerification = errors.New("signature verification failed") 
    ErrNetworkTimeout        = errors.New("network operation timeout")
    ErrPeerDisconnected      = errors.New("peer disconnected")
    ErrMessageTooLarge       = errors.New("message exceeds size limit")
    ErrInvalidConfiguration  = errors.New("invalid network configuration")
    ErrPortInUse            = errors.New("network port already in use")
)
```

### Error Recovery

```go
func handleNetworkError(err error) {
    switch {
    case errors.Is(err, network.ErrNetworkTimeout):
        // Retry with exponential backoff
        time.Sleep(time.Second * 2)
        // Retry operation
        
    case errors.Is(err, network.ErrPeerDisconnected):
        // Clean up peer state
        // Continue with remaining peers
        
    case errors.Is(err, network.ErrSignatureVerification):
        // Security issue - log and ignore message
        log.Println("Security warning: invalid message signature")
        
    default:
        // Generic error handling
        log.Printf("Unhandled network error: %v", err)
    }
}
```

## Performance Considerations

### Optimization Guidelines

1. **Message Frequency**: Limit network messages to essential updates only
2. **Payload Size**: Keep message payloads under 1KB when possible
3. **Batch Operations**: Combine multiple small updates into single messages
4. **Memory Management**: Use object pools for frequently allocated structs
5. **Concurrency**: Leverage Go's goroutines for non-blocking operations

### Performance Metrics

Monitor these metrics for optimal performance:

| Metric | Target | Monitoring |
|--------|--------|------------|
| Peer discovery | <5 seconds | Network latency |
| Message delivery | <50ms local | Round-trip time |
| Bot update cycle | <100ns | CPU profiling |
| State sync | <5μs | Memory allocation |
| Group event start | <10ms | Operation latency |

### Memory Usage

Expected memory overhead per peer connection:
- **Basic connection**: ~1KB per peer
- **With state sync**: ~2KB per peer
- **With group events**: ~3KB per peer
- **Bot personality**: ~0.5KB per bot

### Network Bandwidth

Typical bandwidth usage:
- **Idle networking**: ~100 bytes/minute per peer
- **Active interactions**: ~1KB/minute per peer
- **Group events**: ~5KB/event
- **State synchronization**: ~500 bytes/update

### Scaling Considerations

The DDS multiplayer system is designed for small-group interactions:

- **Recommended**: 2-6 peers for optimal experience
- **Maximum**: 8 peers per network (16 theoretical limit)
- **Performance**: Linear degradation with peer count
- **Network**: Local network only, no internet relay

This API documentation provides comprehensive guidance for integrating with and extending the DDS multiplayer system. The interface-based design ensures flexibility while maintaining performance and reliability for real-time multiplayer interactions.
