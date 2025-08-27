package character

import (
	"math"
	"sync"
	"time"
)

// JealousyTrigger represents a condition that can trigger jealousy
type JealousyTrigger struct {
	Name              string             `json:"name"`
	Description       string             `json:"description"`
	InteractionGap    time.Duration      `json:"interactionGap"`    // Time since last interaction to trigger
	JealousyIncrement float64            `json:"jealousyIncrement"` // How much jealousy to add
	TrustPenalty      float64            `json:"trustPenalty"`      // Trust decrease when triggered
	Conditions        map[string]float64 `json:"conditions"`        // Stat conditions to check
	Responses         []string           `json:"responses"`         // Jealousy responses
	Animations        []string           `json:"animations"`        // Animation states to trigger
	Probability       float64            `json:"probability"`       // Chance to trigger when conditions met
}

// JealousyManager handles all jealousy-related mechanics for romance characters
// Implements automatic jealousy triggering, consequences, and recovery systems
type JealousyManager struct {
	mu                   sync.RWMutex
	enabled              bool
	lastJealousyCheck    time.Time
	jealousyTriggers     []JealousyTrigger
	checkInterval        time.Duration
	jealousyThreshold    float64            // Jealousy level that triggers consequences
	jealousyConsequences map[string]float64 // Stats affected when jealous
}

// NewJealousyManager creates a new jealousy management system
// Uses lazy programmer approach - minimal setup, maximum JSON configurability
func NewJealousyManager(triggers []JealousyTrigger, enabled bool, threshold float64) *JealousyManager {
	return &JealousyManager{
		enabled:           enabled,
		lastJealousyCheck: time.Time{}, // Initialize to zero time for immediate first check
		jealousyTriggers:  triggers,
		checkInterval:     30 * time.Second, // Check every 30 seconds
		jealousyThreshold: threshold,
		jealousyConsequences: map[string]float64{
			"affection": -2.0, // Jealousy reduces affection more noticeably
			"trust":     -1.5, // And trust
			"happiness": -3.0, // And happiness significantly
		},
	}
}

// Update processes jealousy mechanics and returns triggered events
// Called from main character update loop, follows existing pattern
func (jm *JealousyManager) Update(gameState *GameState, lastInteraction time.Time) *TriggeredEvent {
	if !jm.enabled || gameState == nil {
		return nil
	}

	jm.mu.Lock()
	defer jm.mu.Unlock()

	now := time.Now()

	// Always update jealousy consequences - interval checking can be added later if needed
	// This ensures immediate response for testing and game responsiveness
	jm.lastJealousyCheck = now

	// Apply jealousy consequences if above threshold
	jm.applyJealousyConsequences(gameState)

	// Check for jealousy triggers
	return jm.checkJealousyTriggers(gameState, lastInteraction, now)
}

// applyJealousyConsequences reduces stats when jealousy is high
// Implements the "consequences" part of jealousy mechanics
func (jm *JealousyManager) applyJealousyConsequences(gameState *GameState) {
	jealousyLevel := gameState.GetStat("jealousy")

	// Only apply consequences if above threshold
	if jealousyLevel < jm.jealousyThreshold {
		return
	}

	// Scale consequences based on how far above threshold we are
	intensity := math.Min(1.0, (jealousyLevel-jm.jealousyThreshold)/(100.0-jm.jealousyThreshold))

	// Apply scaled consequences
	consequences := make(map[string]float64)
	for statName, basePenalty := range jm.jealousyConsequences {
		consequences[statName] = basePenalty * intensity
	}

	gameState.ApplyInteractionEffects(consequences)
}

// checkJealousyTriggers evaluates triggers and returns event if one fires
// Uses probability-based triggering like existing random events
func (jm *JealousyManager) checkJealousyTriggers(gameState *GameState, lastInteraction time.Time, now time.Time) *TriggeredEvent {
	for _, trigger := range jm.jealousyTriggers {
		if jm.shouldTriggerJealousy(trigger, gameState, lastInteraction, now) {
			if jm.rollProbability(trigger.Probability) {
				return jm.createJealousyEvent(trigger, gameState)
			}
		}
	}
	return nil
}

// shouldTriggerJealousy checks if conditions are met for a jealousy trigger
// Considers interaction timing, stats, and personality traits
func (jm *JealousyManager) shouldTriggerJealousy(trigger JealousyTrigger, gameState *GameState, lastInteraction time.Time, now time.Time) bool {
	// Check interaction gap requirement
	if trigger.InteractionGap > 0 {
		timeSinceInteraction := now.Sub(lastInteraction)
		if timeSinceInteraction < trigger.InteractionGap {
			return false
		}
	}

	// Check stat conditions
	for statName, minValue := range trigger.Conditions {
		currentValue := gameState.GetStat(statName)
		if currentValue < minValue {
			return false
		}
	}

	return true
}

// rollProbability performs probability check for jealousy trigger
// Uses same pattern as existing random event system
func (jm *JealousyManager) rollProbability(probability float64) bool {
	if probability <= 0 {
		return false
	}
	if probability >= 1.0 {
		return true
	}

	// Time-based pseudo-random value
	randomValue := float64((time.Now().UnixNano() % 10000)) / 10000.0
	return randomValue <= probability
}

// createJealousyEvent creates a triggered event from jealousy trigger
// Applies stat changes and returns event for animation/response
func (jm *JealousyManager) createJealousyEvent(trigger JealousyTrigger, gameState *GameState) *TriggeredEvent {
	// Apply jealousy increment and trust penalty
	effects := map[string]float64{
		"jealousy": trigger.JealousyIncrement,
	}

	if trigger.TrustPenalty > 0 {
		effects["trust"] = -trigger.TrustPenalty
	}

	return &TriggeredEvent{
		Name:        trigger.Name,
		Description: trigger.Description,
		Effects:     effects,
		Animations:  trigger.Animations,
		Responses:   trigger.Responses,
		Duration:    5 * time.Second, // Default duration for jealousy events
	}
}

// GetJealousyLevel returns current jealousy intensity (0.0 to 1.0)
// Used by UI and other systems to understand jealousy state
func (jm *JealousyManager) GetJealousyLevel(gameState *GameState) float64 {
	if !jm.enabled || gameState == nil {
		return 0.0
	}

	jm.mu.RLock()
	defer jm.mu.RUnlock()

	jealousyValue := gameState.GetStat("jealousy")
	return math.Min(1.0, jealousyValue/100.0)
}

// IsJealousyCritical returns true if jealousy is above threshold
// Used to determine if relationship is in crisis
func (jm *JealousyManager) IsJealousyCritical(gameState *GameState) bool {
	if !jm.enabled || gameState == nil {
		return false
	}

	jm.mu.RLock()
	defer jm.mu.RUnlock()

	return gameState.GetStat("jealousy") >= jm.jealousyThreshold
}

// SetEnabled allows runtime enabling/disabling of jealousy mechanics
// Useful for characters with different personality configurations
func (jm *JealousyManager) SetEnabled(enabled bool) {
	jm.mu.Lock()
	defer jm.mu.Unlock()
	jm.enabled = enabled
}

// GetStatus returns debug information about jealousy state
// Used for testing and debugging jealousy mechanics
func (jm *JealousyManager) GetStatus(gameState *GameState) map[string]interface{} {
	jm.mu.RLock()
	defer jm.mu.RUnlock()

	status := map[string]interface{}{
		"enabled":       jm.enabled,
		"triggerCount":  len(jm.jealousyTriggers),
		"checkInterval": jm.checkInterval,
		"threshold":     jm.jealousyThreshold,
		"lastCheck":     jm.lastJealousyCheck,
	}

	if gameState != nil {
		status["currentJealousy"] = gameState.GetStat("jealousy")
		status["isCritical"] = jm.IsJealousyCritical(gameState)
		status["intensity"] = jm.GetJealousyLevel(gameState)
	}

	return status
}
