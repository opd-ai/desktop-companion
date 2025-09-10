package network

import (
	"fmt"
	"sync"
	"time"
)

// ActivityType represents the type of network activity event
type ActivityType int

const (
	ActivityJoined ActivityType = iota
	ActivityLeft
	ActivityInteraction
	ActivityStateChange
	ActivityChat
	ActivityBattle
	ActivityDiscovery
)

// String returns human-readable activity type names
func (at ActivityType) String() string {
	switch at {
	case ActivityJoined:
		return "joined"
	case ActivityLeft:
		return "left"
	case ActivityInteraction:
		return "interaction"
	case ActivityStateChange:
		return "state_change"
	case ActivityChat:
		return "chat"
	case ActivityBattle:
		return "battle"
	case ActivityDiscovery:
		return "discovery"
	default:
		return "unknown"
	}
}

// ActivityEvent represents a single network activity event for the feed
type ActivityEvent struct {
	Type          ActivityType `json:"type"`
	PeerID        string       `json:"peerID"`
	CharacterName string       `json:"characterName"`
	Description   string       `json:"description"`
	Timestamp     time.Time    `json:"timestamp"`
	Details       interface{}  `json:"details,omitempty"`
}

// ActivityTracker tracks and manages network activity events
// Uses standard library synchronization primitives following project patterns
type ActivityTracker struct {
	mu        sync.RWMutex
	events    []ActivityEvent
	maxEvents int
	listeners []func(ActivityEvent)
}

// NewActivityTracker creates a new activity tracker with specified max events
func NewActivityTracker(maxEvents int) *ActivityTracker {
	if maxEvents <= 0 {
		maxEvents = 100 // Default sensible limit
	}

	return &ActivityTracker{
		events:    make([]ActivityEvent, 0, maxEvents),
		maxEvents: maxEvents,
		listeners: make([]func(ActivityEvent), 0),
	}
}

// AddEvent adds a new activity event to the tracker
// Automatically adds timestamp and notifies listeners
func (at *ActivityTracker) AddEvent(event ActivityEvent) {
	at.mu.Lock()
	defer at.mu.Unlock()

	// Set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Add to events list
	at.events = append(at.events, event)

	// Keep only recent events to prevent memory growth
	if len(at.events) > at.maxEvents {
		// Remove oldest events (FIFO)
		copy(at.events, at.events[1:])
		at.events = at.events[:at.maxEvents]
	}

	// Notify all listeners asynchronously to avoid blocking
	for _, listener := range at.listeners {
		go func(l func(ActivityEvent)) {
			defer func() {
				// Recover from listener panics to maintain stability
				if r := recover(); r != nil {
					// Log error in real implementation
				}
			}()
			l(event)
		}(listener)
	}
}

// GetRecentEvents returns the most recent events up to the specified count
func (at *ActivityTracker) GetRecentEvents(count int) []ActivityEvent {
	at.mu.RLock()
	defer at.mu.RUnlock()

	if count <= 0 || len(at.events) == 0 {
		return []ActivityEvent{}
	}

	if count > len(at.events) {
		count = len(at.events)
	}

	// Return copy of most recent events
	start := len(at.events) - count
	result := make([]ActivityEvent, count)
	copy(result, at.events[start:])

	return result
}

// GetAllEvents returns a copy of all stored events
func (at *ActivityTracker) GetAllEvents() []ActivityEvent {
	at.mu.RLock()
	defer at.mu.RUnlock()

	result := make([]ActivityEvent, len(at.events))
	copy(result, at.events)

	return result
}

// AddListener registers a callback for new activity events
func (at *ActivityTracker) AddListener(listener func(ActivityEvent)) {
	if listener == nil {
		return
	}

	at.mu.Lock()
	defer at.mu.Unlock()

	at.listeners = append(at.listeners, listener)
}

// Clear removes all stored events and listeners
func (at *ActivityTracker) Clear() {
	at.mu.Lock()
	defer at.mu.Unlock()

	at.events = at.events[:0]
	at.listeners = at.listeners[:0]
}

// GetEventCount returns the current number of stored events
func (at *ActivityTracker) GetEventCount() int {
	at.mu.RLock()
	defer at.mu.RUnlock()

	return len(at.events)
}

// CreateCharacterActionEvent creates an activity event for character interactions
func CreateCharacterActionEvent(peerID, characterName, action string, details interface{}) ActivityEvent {
	description := fmt.Sprintf("%s performed %s", characterName, action)

	return ActivityEvent{
		Type:          ActivityInteraction,
		PeerID:        peerID,
		CharacterName: characterName,
		Description:   description,
		Details:       details,
	}
}

// CreatePeerJoinedEvent creates an activity event for peer joining
func CreatePeerJoinedEvent(peerID, characterName string) ActivityEvent {
	description := fmt.Sprintf("%s joined the network", characterName)

	return ActivityEvent{
		Type:          ActivityJoined,
		PeerID:        peerID,
		CharacterName: characterName,
		Description:   description,
	}
}

// CreatePeerLeftEvent creates an activity event for peer leaving
func CreatePeerLeftEvent(peerID, characterName string) ActivityEvent {
	description := fmt.Sprintf("%s left the network", characterName)

	return ActivityEvent{
		Type:          ActivityLeft,
		PeerID:        peerID,
		CharacterName: characterName,
		Description:   description,
	}
}

// CreateChatEvent creates an activity event for chat messages
func CreateChatEvent(peerID, characterName, message string) ActivityEvent {
	description := fmt.Sprintf("%s: %s", characterName, message)

	return ActivityEvent{
		Type:          ActivityChat,
		PeerID:        peerID,
		CharacterName: characterName,
		Description:   description,
		Details:       map[string]string{"message": message},
	}
}

// CreateBattleEvent creates an activity event for battle actions
func CreateBattleEvent(peerID, characterName, battleAction string) ActivityEvent {
	description := fmt.Sprintf("%s %s", characterName, battleAction)

	return ActivityEvent{
		Type:          ActivityBattle,
		PeerID:        peerID,
		CharacterName: characterName,
		Description:   description,
		Details:       map[string]string{"action": battleAction},
	}
}
