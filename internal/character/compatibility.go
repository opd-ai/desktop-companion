package character

import (
	"math"
	"sync"
	"time"
)

// PlayerBehaviorPattern represents analyzed player interaction patterns
type PlayerBehaviorPattern struct {
	InteractionFrequency map[string]float64 `json:"interactionFrequency"` // How often each interaction is used
	PreferredTimeGaps    map[string]float64 `json:"preferredTimeGaps"`    // Average time between interactions
	SessionLength        float64            `json:"sessionLength"`        // Average session duration
	ConsistencyScore     float64            `json:"consistencyScore"`     // How consistent the player is
	VarietyScore         float64            `json:"varietyScore"`         // How much variety in interactions
	TotalInteractions    int                `json:"totalInteractions"`    // Total interactions recorded
	LastAnalysisUpdate   time.Time          `json:"lastAnalysisUpdate"`   // When analysis was last updated
}

// CompatibilityModifier represents dynamic personality adjustments
type CompatibilityModifier struct {
	StatName      string    `json:"statName"`      // Which stat to modify
	ModifierValue float64   `json:"modifierValue"` // How much to modify (multiplier)
	Reason        string    `json:"reason"`        // Why this modifier exists
	CreatedAt     time.Time `json:"createdAt"`     // When modifier was created
	DecayRate     float64   `json:"decayRate"`     // How fast modifier fades (per hour)
	MinValue      float64   `json:"minValue"`      // Minimum modifier value
	MaxValue      float64   `json:"maxValue"`      // Maximum modifier value
}

// CompatibilityAnalyzer analyzes player behavior and adapts character personality
// Implements advanced compatibility algorithms through behavior pattern recognition
type CompatibilityAnalyzer struct {
	mu                    sync.RWMutex
	enabled               bool
	playerPattern         *PlayerBehaviorPattern
	activeModifiers       []CompatibilityModifier
	analysisInterval      time.Duration
	lastUpdate            time.Time
	personalityThresholds map[string]float64 // Thresholds for triggering adaptations
	adaptationStrength    float64            // How strongly to adapt (0.0 to 1.0)
}

// NewCompatibilityAnalyzer creates a new behavior analysis system
// Uses lazy programmer approach - analyzes existing interaction data
func NewCompatibilityAnalyzer(enabled bool, adaptationStrength float64) *CompatibilityAnalyzer {
	return &CompatibilityAnalyzer{
		enabled: enabled,
		playerPattern: &PlayerBehaviorPattern{
			InteractionFrequency: make(map[string]float64),
			PreferredTimeGaps:    make(map[string]float64),
			LastAnalysisUpdate:   time.Now(),
		},
		activeModifiers:  make([]CompatibilityModifier, 0),
		analysisInterval: 5 * time.Minute, // Analyze every 5 minutes
		lastUpdate:       time.Now(),
		personalityThresholds: map[string]float64{
			"consistent_player":   0.8, // Threshold for considering player consistent
			"variety_lover":       0.7, // Threshold for considering player variety-seeking
			"frequent_interactor": 0.5, // Interactions per minute to be considered frequent
		},
		adaptationStrength: math.Min(1.0, math.Max(0.0, adaptationStrength)),
	}
}

// Update analyzes recent player behavior and adjusts compatibility modifiers
// Called from main character update loop, follows existing patterns
func (ca *CompatibilityAnalyzer) Update(gameState *GameState) []CompatibilityModifier {
	if !ca.enabled || gameState == nil {
		return nil
	}

	ca.mu.Lock()
	defer ca.mu.Unlock()

	now := time.Now()

	// Check if enough time has passed for analysis update
	if now.Sub(ca.lastUpdate) < ca.analysisInterval {
		// Still decay existing modifiers
		ca.decayModifiers()
		return ca.getActiveModifiers()
	}

	ca.lastUpdate = now

	// Analyze player behavior patterns
	ca.analyzePlayerBehavior(gameState)

	// Generate new compatibility modifiers based on analysis
	newModifiers := ca.generateCompatibilityModifiers()

	// Add new modifiers to active list
	ca.activeModifiers = append(ca.activeModifiers, newModifiers...)

	// Clean up expired modifiers
	ca.cleanupExpiredModifiers()

	return ca.getActiveModifiers()
}

// analyzePlayerBehavior examines interaction history to identify patterns
// Uses existing interaction tracking data from game state
func (ca *CompatibilityAnalyzer) analyzePlayerBehavior(gameState *GameState) {
	interactionHistory := gameState.GetInteractionHistory()
	if len(interactionHistory) == 0 {
		return
	}

	// Calculate interaction frequency
	totalInteractions := 0
	for interactionType, timestamps := range interactionHistory {
		count := len(timestamps)
		totalInteractions += count
		ca.playerPattern.InteractionFrequency[interactionType] = float64(count)

		// Calculate average time gaps for this interaction type
		if count > 1 {
			totalGap := 0.0
			for i := 1; i < len(timestamps); i++ {
				gap := timestamps[i].Sub(timestamps[i-1]).Minutes()
				totalGap += gap
			}
			ca.playerPattern.PreferredTimeGaps[interactionType] = totalGap / float64(count-1)
		}
	}

	ca.playerPattern.TotalInteractions = totalInteractions

	// Calculate consistency score (how regular are interactions)
	ca.calculateConsistencyScore(interactionHistory)

	// Calculate variety score (how many different interactions used)
	ca.calculateVarietyScore(interactionHistory)

	ca.playerPattern.LastAnalysisUpdate = time.Now()
}

// calculateConsistencyScore determines how consistent player interaction timing is
// Higher score means more predictable interaction patterns
func (ca *CompatibilityAnalyzer) calculateConsistencyScore(interactionHistory map[string][]time.Time) {
	if len(interactionHistory) == 0 {
		ca.playerPattern.ConsistencyScore = 0.0
		return
	}

	// Collect all interaction timestamps
	var allTimestamps []time.Time
	for _, timestamps := range interactionHistory {
		allTimestamps = append(allTimestamps, timestamps...)
	}

	if len(allTimestamps) < 3 {
		ca.playerPattern.ConsistencyScore = 0.5 // Neutral score for insufficient data
		return
	}

	// Calculate variance in interaction gaps
	gaps := make([]float64, 0, len(allTimestamps)-1)
	for i := 1; i < len(allTimestamps); i++ {
		gap := allTimestamps[i].Sub(allTimestamps[i-1]).Minutes()
		gaps = append(gaps, gap)
	}

	// Calculate standard deviation of gaps
	mean := 0.0
	for _, gap := range gaps {
		mean += gap
	}
	mean /= float64(len(gaps))

	variance := 0.0
	for _, gap := range gaps {
		variance += math.Pow(gap-mean, 2)
	}
	variance /= float64(len(gaps))

	stdDev := math.Sqrt(variance)

	// Convert to consistency score (lower std dev = higher consistency)
	// Score between 0.0 and 1.0
	ca.playerPattern.ConsistencyScore = math.Max(0.0, 1.0-(stdDev/60.0)) // Normalize by hour
}

// calculateVarietyScore determines how diverse player interactions are
// Higher score means player uses many different interaction types
func (ca *CompatibilityAnalyzer) calculateVarietyScore(interactionHistory map[string][]time.Time) {
	if len(interactionHistory) == 0 {
		ca.playerPattern.VarietyScore = 0.0
		return
	}

	// Count unique interaction types used
	uniqueTypes := len(interactionHistory)

	// Calculate distribution evenness (how evenly distributed are interactions)
	totalInteractions := 0
	for _, timestamps := range interactionHistory {
		totalInteractions += len(timestamps)
	}

	if totalInteractions == 0 {
		ca.playerPattern.VarietyScore = 0.0
		return
	}

	// Shannon diversity index adapted for interactions
	entropy := 0.0
	for _, timestamps := range interactionHistory {
		if len(timestamps) > 0 {
			proportion := float64(len(timestamps)) / float64(totalInteractions)
			entropy -= proportion * math.Log2(proportion)
		}
	}

	// Normalize entropy to 0-1 scale
	maxEntropy := math.Log2(float64(uniqueTypes))
	if maxEntropy > 0 {
		ca.playerPattern.VarietyScore = entropy / maxEntropy
	} else {
		ca.playerPattern.VarietyScore = 0.0
	}
}

// generateCompatibilityModifiers creates new modifiers based on behavior analysis
// Adapts character personality to better match player preferences
func (ca *CompatibilityAnalyzer) generateCompatibilityModifiers() []CompatibilityModifier {
	var modifiers []CompatibilityModifier
	now := time.Now()

	// Modifier for consistent players (boost consistent_interaction compatibility)
	if ca.playerPattern.ConsistencyScore >= ca.personalityThresholds["consistent_player"] {
		modifier := CompatibilityModifier{
			StatName:      "consistent_interaction_bonus",
			ModifierValue: 1.0 + (ca.adaptationStrength * 0.3), // Up to 30% bonus
			Reason:        "Player shows consistent interaction patterns",
			CreatedAt:     now,
			DecayRate:     0.02, // Decays 2% per hour
			MinValue:      1.0,
			MaxValue:      1.5,
		}
		modifiers = append(modifiers, modifier)
	}

	// Modifier for variety-seeking players (boost variety_preference compatibility)
	if ca.playerPattern.VarietyScore >= ca.personalityThresholds["variety_lover"] {
		modifier := CompatibilityModifier{
			StatName:      "variety_preference_bonus",
			ModifierValue: 1.0 + (ca.adaptationStrength * 0.25), // Up to 25% bonus
			Reason:        "Player enjoys interaction variety",
			CreatedAt:     now,
			DecayRate:     0.03, // Decays 3% per hour
			MinValue:      1.0,
			MaxValue:      1.4,
		}
		modifiers = append(modifiers, modifier)
	}

	// Modifier for frequent interactors (general interaction bonuses)
	averageFrequency := ca.calculateAverageInteractionFrequency()
	if averageFrequency >= ca.personalityThresholds["frequent_interactor"] {
		modifier := CompatibilityModifier{
			StatName:      "interaction_responsiveness_bonus",
			ModifierValue: 1.0 + (ca.adaptationStrength * 0.2), // Up to 20% bonus
			Reason:        "Player interacts frequently and actively",
			CreatedAt:     now,
			DecayRate:     0.015, // Decays 1.5% per hour
			MinValue:      1.0,
			MaxValue:      1.3,
		}
		modifiers = append(modifiers, modifier)
	}

	return modifiers
}

// calculateAverageInteractionFrequency determines overall interaction frequency
// Used to identify highly active players for adaptation
func (ca *CompatibilityAnalyzer) calculateAverageInteractionFrequency() float64 {
	if ca.playerPattern.TotalInteractions == 0 {
		return 0.0
	}

	// Calculate interactions per minute based on time since first interaction
	timeSinceFirstInteraction := time.Since(ca.playerPattern.LastAnalysisUpdate).Minutes()
	if timeSinceFirstInteraction == 0 {
		return 0.0
	}

	return float64(ca.playerPattern.TotalInteractions) / timeSinceFirstInteraction
}

// decayModifiers reduces the strength of existing modifiers over time
// Prevents permanent personality changes from temporary behavior patterns
func (ca *CompatibilityAnalyzer) decayModifiers() {
	now := time.Now()

	for i := range ca.activeModifiers {
		modifier := &ca.activeModifiers[i]
		hoursSinceCreation := now.Sub(modifier.CreatedAt).Hours()

		// Calculate decay
		decayAmount := modifier.DecayRate * hoursSinceCreation
		newValue := modifier.ModifierValue * (1.0 - decayAmount)

		// Apply bounds
		modifier.ModifierValue = math.Max(modifier.MinValue, math.Min(modifier.MaxValue, newValue))
	}
}

// cleanupExpiredModifiers removes modifiers that have decayed to minimum values
// Keeps modifier list clean and prevents performance issues
func (ca *CompatibilityAnalyzer) cleanupExpiredModifiers() {
	var activeModifiers []CompatibilityModifier

	for _, modifier := range ca.activeModifiers {
		// Keep modifiers that still have significant effect
		if modifier.ModifierValue > modifier.MinValue+0.01 {
			activeModifiers = append(activeModifiers, modifier)
		}
	}

	ca.activeModifiers = activeModifiers
}

// getActiveModifiers returns a copy of current active modifiers
// Safe for concurrent access
func (ca *CompatibilityAnalyzer) getActiveModifiers() []CompatibilityModifier {
	modifiers := make([]CompatibilityModifier, len(ca.activeModifiers))
	copy(modifiers, ca.activeModifiers)
	return modifiers
}

// GetPlayerPattern returns current analyzed player behavior pattern
// Used for debugging and UI display of player behavior insights
func (ca *CompatibilityAnalyzer) GetPlayerPattern() *PlayerBehaviorPattern {
	if !ca.enabled {
		return nil
	}

	ca.mu.RLock()
	defer ca.mu.RUnlock()

	// Return a copy to prevent external modification
	pattern := &PlayerBehaviorPattern{
		InteractionFrequency: make(map[string]float64),
		PreferredTimeGaps:    make(map[string]float64),
		SessionLength:        ca.playerPattern.SessionLength,
		ConsistencyScore:     ca.playerPattern.ConsistencyScore,
		VarietyScore:         ca.playerPattern.VarietyScore,
		TotalInteractions:    ca.playerPattern.TotalInteractions,
		LastAnalysisUpdate:   ca.playerPattern.LastAnalysisUpdate,
	}

	for k, v := range ca.playerPattern.InteractionFrequency {
		pattern.InteractionFrequency[k] = v
	}
	for k, v := range ca.playerPattern.PreferredTimeGaps {
		pattern.PreferredTimeGaps[k] = v
	}

	return pattern
}

// GetCompatibilityInsights returns analysis of current compatibility state
// Useful for debugging and understanding why certain modifiers are active
func (ca *CompatibilityAnalyzer) GetCompatibilityInsights() map[string]interface{} {
	if !ca.enabled {
		return map[string]interface{}{"enabled": false}
	}

	ca.mu.RLock()
	defer ca.mu.RUnlock()

	insights := map[string]interface{}{
		"enabled":            ca.enabled,
		"adaptationStrength": ca.adaptationStrength,
		"analysisInterval":   ca.analysisInterval,
		"activeModifiers":    len(ca.activeModifiers),
		"playerPattern":      ca.playerPattern,
		"lastUpdate":         ca.lastUpdate,
	}

	// Add modifier details
	modifierDetails := make([]map[string]interface{}, 0, len(ca.activeModifiers))
	for _, modifier := range ca.activeModifiers {
		detail := map[string]interface{}{
			"statName":      modifier.StatName,
			"modifierValue": modifier.ModifierValue,
			"reason":        modifier.Reason,
			"createdAt":     modifier.CreatedAt,
			"age":           time.Since(modifier.CreatedAt),
		}
		modifierDetails = append(modifierDetails, detail)
	}
	insights["modifierDetails"] = modifierDetails

	return insights
}

// SetEnabled allows runtime enabling/disabling of compatibility analysis
// Useful for different character personality configurations
func (ca *CompatibilityAnalyzer) SetEnabled(enabled bool) {
	ca.mu.Lock()
	defer ca.mu.Unlock()
	ca.enabled = enabled
}

// SetAdaptationStrength adjusts how strongly the character adapts to player behavior
// Value between 0.0 (no adaptation) and 1.0 (maximum adaptation)
func (ca *CompatibilityAnalyzer) SetAdaptationStrength(strength float64) {
	ca.mu.Lock()
	defer ca.mu.Unlock()
	ca.adaptationStrength = math.Min(1.0, math.Max(0.0, strength))
}

// ForceAnalysis forces immediate analysis regardless of timing intervals
// Used for testing to bypass the normal 5-minute analysis interval
func (ca *CompatibilityAnalyzer) ForceAnalysis(gameState *GameState) []CompatibilityModifier {
	if !ca.enabled || gameState == nil {
		return nil
	}

	ca.mu.Lock()
	// Force analysis by setting lastUpdate to past time
	ca.lastUpdate = time.Now().Add(-ca.analysisInterval - time.Minute)
	ca.mu.Unlock()

	// Call Update which will now proceed with analysis
	return ca.Update(gameState)
}
