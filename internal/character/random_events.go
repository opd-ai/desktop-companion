package character

import (
	"math/rand"
	"sync"
	"time"
)

// RandomEventManager handles automatic triggering of random events that affect character stats
// This follows the "lazy programmer" approach using only Go standard library
// Events are JSON-configurable and integrate with existing game state management
type RandomEventManager struct {
	mu             sync.RWMutex
	events         []RandomEventConfig  // Event configurations from character card
	lastCheck      time.Time            // Last time events were checked
	eventCooldowns map[string]time.Time // Per-event cooldown tracking
	enabled        bool                 // Whether random events are enabled
	checkInterval  time.Duration        // How often to check for events
	randomSource   *rand.Rand           // Random number generator for event probability
}

// NewRandomEventManager creates a new random event manager from character card configuration
// Uses current time as seed for pseudo-random event generation
func NewRandomEventManager(events []RandomEventConfig, enabled bool, interval time.Duration) *RandomEventManager {
	// Initialize with time-based seed for pseudo-randomness
	source := rand.NewSource(time.Now().UnixNano())

	rem := &RandomEventManager{
		events:         events,
		lastCheck:      time.Now().Add(-interval), // Set lastCheck in the past so first update can trigger
		eventCooldowns: make(map[string]time.Time),
		enabled:        enabled && len(events) > 0, // Only enable if we have events and enabled is true
		checkInterval:  interval,
		randomSource:   rand.New(source),
	}

	return rem
}

// Update checks for and triggers random events based on elapsed time and probability
// Returns a TriggeredEvent if an event occurred, nil otherwise
// This method is called from the main game state update loop
func (rem *RandomEventManager) Update(elapsed time.Duration, gameState *GameState) *TriggeredEvent {
	if !rem.shouldProcessEvents(gameState) {
		return nil
	}

	rem.mu.Lock()
	defer rem.mu.Unlock()

	if !rem.isCheckIntervalReached() {
		return nil
	}

	return rem.processEventTriggers(gameState)
}

// shouldProcessEvents checks if event processing should proceed
func (rem *RandomEventManager) shouldProcessEvents(gameState *GameState) bool {
	return rem.enabled && len(rem.events) > 0 && gameState != nil
}

// isCheckIntervalReached determines if enough time has passed since last check
func (rem *RandomEventManager) isCheckIntervalReached() bool {
	now := time.Now()
	timeSinceLastCheck := now.Sub(rem.lastCheck)
	return timeSinceLastCheck >= rem.checkInterval
}

// processEventTriggers iterates through events and attempts to trigger one
func (rem *RandomEventManager) processEventTriggers(gameState *GameState) *TriggeredEvent {
	now := time.Now()

	for _, event := range rem.events {
		if triggeredEvent := rem.attemptEventTrigger(event, now, gameState); triggeredEvent != nil {
			rem.lastCheck = now
			return triggeredEvent
		}
	}

	rem.lastCheck = now
	return nil
}

// attemptEventTrigger tries to trigger a single event and returns result
func (rem *RandomEventManager) attemptEventTrigger(event RandomEventConfig, now time.Time, gameState *GameState) *TriggeredEvent {
	if !rem.canTriggerEvent(event, now, gameState) {
		return nil
	}

	if !rem.rollEventProbability(event.Probability) {
		return nil
	}

	return rem.createTriggeredEvent(event, now)
}

// rollEventProbability performs probability check for event triggering
func (rem *RandomEventManager) rollEventProbability(probability float64) bool {
	randomValue := rem.randomSource.Float64()
	return randomValue <= probability
}

// createTriggeredEvent creates and records a triggered event
func (rem *RandomEventManager) createTriggeredEvent(event RandomEventConfig, now time.Time) *TriggeredEvent {
	rem.eventCooldowns[event.Name] = now

	return &TriggeredEvent{
		Name:        event.Name,
		Description: event.Description,
		Effects:     event.Effects,
		Animations:  event.Animations,
		Responses:   event.Responses,
		Duration:    time.Duration(event.Duration) * time.Second,
	}
}

// canTriggerEvent checks if an event can currently trigger based on cooldowns and conditions
func (rem *RandomEventManager) canTriggerEvent(event RandomEventConfig, now time.Time, gameState *GameState) bool {
	// Check event-specific cooldown
	if lastTrigger, exists := rem.eventCooldowns[event.Name]; exists {
		cooldownDuration := time.Duration(event.Cooldown) * time.Second
		if now.Sub(lastTrigger) < cooldownDuration {
			return false
		}
	}

	// Check stat conditions if specified
	if len(event.Conditions) > 0 {
		return gameState.CanSatisfyRequirements(event.Conditions)
	}

	return true
}

// GetRandomResponse returns a random response from the event's response list
// Uses the same random source as event triggering for consistency
func (rem *RandomEventManager) GetRandomResponse(responses []string) string {
	if len(responses) == 0 {
		return ""
	}

	rem.mu.RLock()
	defer rem.mu.RUnlock()

	if !rem.enabled {
		return ""
	}

	index := rem.randomSource.Intn(len(responses))
	return responses[index]
}

// IsEnabled returns whether random events are currently enabled
func (rem *RandomEventManager) IsEnabled() bool {
	rem.mu.RLock()
	defer rem.mu.RUnlock()
	return rem.enabled
}

// SetEnabled allows enabling/disabling random events at runtime
func (rem *RandomEventManager) SetEnabled(enabled bool) {
	rem.mu.Lock()
	defer rem.mu.Unlock()
	rem.enabled = enabled
}

// GetEventCount returns the number of configured events
func (rem *RandomEventManager) GetEventCount() int {
	rem.mu.RLock()
	defer rem.mu.RUnlock()
	return len(rem.events)
}

// GetLastCheckTime returns when events were last checked (for debugging)
func (rem *RandomEventManager) GetLastCheckTime() time.Time {
	rem.mu.RLock()
	defer rem.mu.RUnlock()
	return rem.lastCheck
}

// TriggeredEvent represents an event that has been triggered and should be processed
// Contains all information needed to apply effects and show animations/dialogs
type TriggeredEvent struct {
	Name        string             // Event name for logging
	Description string             // Human readable description
	Effects     map[string]float64 // Stat changes to apply
	Animations  []string           // Animations to play
	Responses   []string           // Dialog responses to show
	Duration    time.Duration      // How long the event effect lasts
}

// HasEffects returns true if this event modifies character stats
func (te *TriggeredEvent) HasEffects() bool {
	return len(te.Effects) > 0
}

// HasAnimations returns true if this event should trigger animations
func (te *TriggeredEvent) HasAnimations() bool {
	return len(te.Animations) > 0
}

// HasResponses returns true if this event should show dialog responses
func (te *TriggeredEvent) HasResponses() bool {
	return len(te.Responses) > 0
}

// GetRandomAnimation returns a random animation from the event's animation list
// Returns empty string if no animations are configured
func (te *TriggeredEvent) GetRandomAnimation() string {
	if len(te.Animations) == 0 {
		return ""
	}

	// Use time-based pseudo-random selection for simplicity
	index := int(time.Now().UnixNano()) % len(te.Animations)
	return te.Animations[index]
}

// GetRandomResponse returns a random response from the event's response list
// Returns empty string if no responses are configured
func (te *TriggeredEvent) GetRandomResponse() string {
	if len(te.Responses) == 0 {
		return ""
	}

	// Use time-based pseudo-random selection for simplicity
	index := int(time.Now().UnixNano()) % len(te.Responses)
	return te.Responses[index]
}
