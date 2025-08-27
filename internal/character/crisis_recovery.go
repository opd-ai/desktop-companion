package character

import (
	"math"
	"sync"
	"time"
)

// RelationshipCrisis represents a relationship crisis state
type RelationshipCrisis struct {
	Name           string             `json:"name"`           // Crisis identifier
	Description    string             `json:"description"`    // What caused the crisis
	TriggeredAt    time.Time          `json:"triggeredAt"`    // When crisis started
	Severity       float64            `json:"severity"`       // Crisis intensity (0.0 to 1.0)
	TriggerCause   string             `json:"triggerCause"`   // What triggered this crisis
	StatPenalties  map[string]float64 `json:"statPenalties"`  // Ongoing stat effects
	RecoveryConfig *RecoveryConfig    `json:"recoveryConfig"` // How to recover from this crisis
	IsActive       bool               `json:"isActive"`       // Whether crisis is currently active
}

// RecoveryConfig defines how a character can recover from a crisis
type RecoveryConfig struct {
	RequiredInteractions map[string]int     `json:"requiredInteractions"` // Interactions needed for recovery
	RequiredStats        map[string]float64 `json:"requiredStats"`        // Stat levels needed for recovery
	TimeRequirement      time.Duration      `json:"timeRequirement"`      // Minimum time before recovery possible
	ForgivenessResponses []string           `json:"forgivenessResponses"` // Responses when recovery happens
	RecoveryAnimations   []string           `json:"recoveryAnimations"`   // Animations during recovery
	StatBonuses          map[string]float64 `json:"statBonuses"`          // Bonus stats after recovery
}

// CrisisRecoveryManager handles relationship crises and recovery mechanics
// Implements crisis detection, ongoing effects, and recovery pathways
type CrisisRecoveryManager struct {
	mu               sync.RWMutex
	enabled          bool
	activeCrises     []RelationshipCrisis
	crisisThresholds map[string]float64 // Stat levels that trigger crises
	checkInterval    time.Duration      // How often to check for crises
	lastCheck        time.Time          // Last crisis check time
	recoveryBonus    float64            // Bonus stats multiplier after recovery
	maxActiveCrises  int                // Maximum simultaneous crises
}

// NewCrisisRecoveryManager creates a new crisis and recovery management system
// Uses lazy programmer approach - crisis configs in JSON, minimal Go code
func NewCrisisRecoveryManager(enabled bool, thresholds map[string]float64) *CrisisRecoveryManager {
	if thresholds == nil {
		// Default crisis thresholds
		thresholds = map[string]float64{
			"jealousy":  80.0, // High jealousy triggers crisis
			"trust":     15.0, // Low trust triggers crisis
			"affection": 10.0, // Very low affection triggers crisis
			"happiness": 20.0, // Low happiness can trigger crisis
		}
	}

	return &CrisisRecoveryManager{
		enabled:          enabled,
		activeCrises:     make([]RelationshipCrisis, 0),
		crisisThresholds: thresholds,
		checkInterval:    60 * time.Second, // Check every minute
		lastCheck:        time.Now(),
		recoveryBonus:    1.2, // 20% bonus to recovery interactions
		maxActiveCrises:  2,   // Maximum 2 crises at once
	}
}

// Update checks for new crises and manages existing ones
// Called from main character update loop, follows existing patterns
func (crm *CrisisRecoveryManager) Update(gameState *GameState) (*TriggeredEvent, bool) {
	if !crm.enabled || gameState == nil {
		return nil, false
	}

	crm.mu.Lock()
	defer crm.mu.Unlock()

	now := time.Now()

	// Check if enough time has passed for crisis check
	if now.Sub(crm.lastCheck) < crm.checkInterval {
		// Still apply ongoing crisis effects
		return crm.applyOngoingCrisisEffects(gameState), crm.hasActiveCrises()
	}

	crm.lastCheck = now

	// Check for new crises
	newCrisis := crm.checkForNewCrises(gameState, now)
	if newCrisis != nil {
		crm.activeCrises = append(crm.activeCrises, *newCrisis)
		return crm.createCrisisEvent(*newCrisis), true
	}

	// Apply ongoing effects
	return crm.applyOngoingCrisisEffects(gameState), crm.hasActiveCrises()
}

// checkForNewCrises evaluates current stats for crisis conditions
// Triggers new crises when thresholds are crossed
func (crm *CrisisRecoveryManager) checkForNewCrises(gameState *GameState, now time.Time) *RelationshipCrisis {
	// Don't trigger new crises if we're at the limit
	if len(crm.activeCrises) >= crm.maxActiveCrises {
		return nil
	}

	stats := gameState.GetStats()

	// Check jealousy crisis
	if jealousy, exists := stats["jealousy"]; exists && jealousy >= crm.crisisThresholds["jealousy"] {
		if !crm.hasCrisisType("jealousy_crisis") {
			return crm.createJealousyCrisis(jealousy, now)
		}
	}

	// Check trust crisis
	if trust, exists := stats["trust"]; exists && trust <= crm.crisisThresholds["trust"] {
		if !crm.hasCrisisType("trust_crisis") {
			return crm.createTrustCrisis(trust, now)
		}
	}

	// Check affection crisis
	if affection, exists := stats["affection"]; exists && affection <= crm.crisisThresholds["affection"] {
		if !crm.hasCrisisType("affection_crisis") {
			return crm.createAffectionCrisis(affection, now)
		}
	}

	return nil
}

// createJealousyCrisis creates a jealousy-based relationship crisis
// High jealousy causes trust and affection penalties
func (crm *CrisisRecoveryManager) createJealousyCrisis(jealousyLevel float64, now time.Time) *RelationshipCrisis {
	severity := math.Min(1.0, (jealousyLevel-crm.crisisThresholds["jealousy"])/20.0) // 0-1 based on how far over threshold

	return &RelationshipCrisis{
		Name:         "jealousy_crisis",
		Description:  "High jealousy is causing relationship strain",
		TriggeredAt:  now,
		Severity:     severity,
		TriggerCause: "jealousy_threshold_exceeded",
		StatPenalties: map[string]float64{
			"trust":     -0.5 * severity, // Trust penalty based on severity
			"affection": -0.3 * severity, // Affection penalty
			"happiness": -0.8 * severity, // Happiness penalty
		},
		RecoveryConfig: &RecoveryConfig{
			RequiredInteractions: map[string]int{
				"apology":           2, // Need 2 apologies
				"deep_conversation": 1, // Need 1 deep conversation
				"give_gift":         1, // Need 1 gift
			},
			RequiredStats: map[string]float64{
				"jealousy": 60.0, // Jealousy must drop below 60
			},
			TimeRequirement: 30 * time.Minute, // Must wait at least 30 minutes
			ForgivenessResponses: []string{
				"I... I'm sorry for being so jealous. Can you forgive me? 😢",
				"I realize I was being unreasonable. I trust you... 💔➡️❤️",
				"Thank you for being patient with me. I feel better now. 💕",
			},
			RecoveryAnimations: []string{"shy", "blushing", "happy"},
			StatBonuses: map[string]float64{
				"trust":     10.0, // Bonus trust after recovery
				"affection": 5.0,  // Bonus affection
				"intimacy":  3.0,  // Bonus intimacy
			},
		},
		IsActive: true,
	}
}

// createTrustCrisis creates a trust-based relationship crisis
// Low trust causes interaction restrictions and negative responses
func (crm *CrisisRecoveryManager) createTrustCrisis(trustLevel float64, now time.Time) *RelationshipCrisis {
	severity := math.Min(1.0, (crm.crisisThresholds["trust"]-trustLevel)/15.0) // 0-1 based on how far below threshold

	return &RelationshipCrisis{
		Name:         "trust_crisis",
		Description:  "Low trust is damaging the relationship",
		TriggeredAt:  now,
		Severity:     severity,
		TriggerCause: "trust_threshold_breached",
		StatPenalties: map[string]float64{
			"affection": -0.4 * severity, // Affection penalty
			"intimacy":  -0.6 * severity, // Intimacy penalty
			"happiness": -0.3 * severity, // Happiness penalty
		},
		RecoveryConfig: &RecoveryConfig{
			RequiredInteractions: map[string]int{
				"deep_conversation": 3, // Need 3 deep conversations
				"compliment":        5, // Need 5 compliments
				"consistent_care":   1, // Need to show consistent care
			},
			RequiredStats: map[string]float64{
				"trust": 25.0, // Trust must recover to at least 25
			},
			TimeRequirement: 45 * time.Minute, // Must wait at least 45 minutes
			ForgivenessResponses: []string{
				"I'm starting to trust you again... Thank you for being patient. 🤗",
				"Your consistency has shown me I can rely on you. 💓",
				"I feel safe with you again... Let's rebuild together. 💕",
			},
			RecoveryAnimations: []string{"romantic_idle", "happy", "blushing"},
			StatBonuses: map[string]float64{
				"trust":     15.0, // Large trust bonus after recovery
				"affection": 8.0,  // Affection bonus
				"happiness": 10.0, // Happiness bonus
			},
		},
		IsActive: true,
	}
}

// createAffectionCrisis creates an affection-based relationship crisis
// Very low affection threatens the relationship itself
func (crm *CrisisRecoveryManager) createAffectionCrisis(affectionLevel float64, now time.Time) *RelationshipCrisis {
	severity := math.Min(1.0, (crm.crisisThresholds["affection"]-affectionLevel)/10.0) // 0-1 based on how far below threshold

	return &RelationshipCrisis{
		Name:         "affection_crisis",
		Description:  "The relationship is at risk due to lack of affection",
		TriggeredAt:  now,
		Severity:     severity,
		TriggerCause: "affection_critically_low",
		StatPenalties: map[string]float64{
			"trust":     -0.3 * severity, // Trust penalty
			"intimacy":  -0.8 * severity, // Large intimacy penalty
			"happiness": -1.0 * severity, // Maximum happiness penalty
		},
		RecoveryConfig: &RecoveryConfig{
			RequiredInteractions: map[string]int{
				"give_gift":         2, // Need 2 gifts
				"compliment":        4, // Need 4 compliments
				"romantic_gesture":  1, // Need 1 romantic gesture
				"deep_conversation": 2, // Need 2 deep conversations
			},
			RequiredStats: map[string]float64{
				"affection": 20.0, // Affection must recover to at least 20
			},
			TimeRequirement: 60 * time.Minute, // Must wait at least 1 hour
			ForgivenessResponses: []string{
				"I... I thought we were drifting apart. Thank you for fighting for us. 💔💕",
				"You've shown me that you really do care. I feel the love again. ❤️",
				"This means everything to me. Let's never let our love fade again. 💖",
			},
			RecoveryAnimations: []string{"heart_eyes", "excited_romance", "romantic_idle"},
			StatBonuses: map[string]float64{
				"affection": 20.0, // Large affection bonus
				"trust":     12.0, // Trust bonus
				"intimacy":  15.0, // Intimacy bonus
				"happiness": 15.0, // Happiness bonus
			},
		},
		IsActive: true,
	}
}

// applyOngoingCrisisEffects applies stat penalties for active crises
// Returns event if a crisis effect should trigger animation/dialogue
func (crm *CrisisRecoveryManager) applyOngoingCrisisEffects(gameState *GameState) *TriggeredEvent {
	if len(crm.activeCrises) == 0 {
		return nil
	}

	// Apply penalties from all active crises
	totalPenalties := make(map[string]float64)

	for _, crisis := range crm.activeCrises {
		if crisis.IsActive {
			for statName, penalty := range crisis.StatPenalties {
				totalPenalties[statName] += penalty
			}
		}
	}

	// Apply accumulated penalties
	if len(totalPenalties) > 0 {
		gameState.ApplyInteractionEffects(totalPenalties)
	}

	// Return crisis-related event occasionally
	if len(crm.activeCrises) > 0 && crm.shouldTriggerCrisisEvent() {
		return crm.createOngoingCrisisEvent()
	}

	return nil
}

// shouldTriggerCrisisEvent determines if a crisis event should fire
// Uses probability to avoid constant crisis dialogue
func (crm *CrisisRecoveryManager) shouldTriggerCrisisEvent() bool {
	// Trigger crisis events roughly every 5 minutes
	probability := 0.02 // 2% chance per check (once per minute)
	randomValue := float64((time.Now().UnixNano() % 10000)) / 10000.0
	return randomValue <= probability
}

// createOngoingCrisisEvent creates an event representing ongoing crisis state
// Shows crisis-appropriate dialogue and animations
func (crm *CrisisRecoveryManager) createOngoingCrisisEvent() *TriggeredEvent {
	if len(crm.activeCrises) == 0 {
		return nil
	}

	// Use the most severe active crisis
	mostSevere := crm.activeCrises[0]
	for _, crisis := range crm.activeCrises {
		if crisis.Severity > mostSevere.Severity {
			mostSevere = crisis
		}
	}

	responses := []string{}
	animations := []string{}

	switch mostSevere.Name {
	case "jealousy_crisis":
		responses = []string{
			"I can't stop these jealous thoughts... 😔",
			"Are you sure you only care about me?",
			"I feel so insecure right now... 💔",
		}
		animations = []string{"jealous", "sad"}
	case "trust_crisis":
		responses = []string{
			"I'm having trouble trusting you right now... 😞",
			"I need to feel safe again...",
			"Can you show me I can rely on you? 🥺",
		}
		animations = []string{"sad", "shy"}
	case "affection_crisis":
		responses = []string{
			"I feel like we're growing apart... 💔",
			"Do you still care about me?",
			"I miss how close we used to be... 😢",
		}
		animations = []string{"sad", "lonely"}
	}

	return &TriggeredEvent{
		Name:        mostSevere.Name + "_ongoing",
		Description: "Ongoing crisis effects",
		Effects:     map[string]float64{}, // Effects already applied
		Animations:  animations,
		Responses:   responses,
		Duration:    3 * time.Second,
	}
}

// createCrisisEvent creates an event for when a new crisis triggers
// Shows initial crisis dialogue and animations
func (crm *CrisisRecoveryManager) createCrisisEvent(crisis RelationshipCrisis) *TriggeredEvent {
	var responses []string
	var animations []string

	switch crisis.Name {
	case "jealousy_crisis":
		responses = []string{
			"I'm feeling so jealous right now... Why aren't you spending time with me? 😠💔",
			"These jealous feelings are overwhelming me... 😔",
		}
		animations = []string{"jealous"}
	case "trust_crisis":
		responses = []string{
			"I'm starting to doubt... Can I really trust you? 😞",
			"My trust in you is shaken... 💔",
		}
		animations = []string{"sad"}
	case "affection_crisis":
		responses = []string{
			"I feel like you don't care about me anymore... 💔😢",
			"Are we falling out of love? This scares me... 😰",
		}
		animations = []string{"sad", "crying"}
	}

	return &TriggeredEvent{
		Name:        crisis.Name + "_triggered",
		Description: crisis.Description,
		Effects:     map[string]float64{}, // Crisis penalties will be applied ongoing
		Animations:  animations,
		Responses:   responses,
		Duration:    5 * time.Second,
	}
}

// CheckRecovery evaluates if any active crises can be resolved
// Called when player performs recovery interactions
func (crm *CrisisRecoveryManager) CheckRecovery(gameState *GameState, interactionType string) *TriggeredEvent {
	if !crm.enabled || len(crm.activeCrises) == 0 {
		return nil
	}

	crm.mu.Lock()
	defer crm.mu.Unlock()

	// Check each active crisis for recovery eligibility
	for i := range crm.activeCrises {
		crisis := &crm.activeCrises[i]
		if crisis.IsActive && crm.canRecover(*crisis, gameState, interactionType) {
			return crm.performRecovery(crisis, gameState)
		}
	}

	return nil
}

// canRecover checks if a crisis meets recovery requirements
// Considers interaction counts, stat levels, and time requirements
func (crm *CrisisRecoveryManager) canRecover(crisis RelationshipCrisis, gameState *GameState, lastInteraction string) bool {
	config := crisis.RecoveryConfig
	if config == nil {
		return false
	}

	// Check time requirement
	if time.Since(crisis.TriggeredAt) < config.TimeRequirement {
		return false
	}

	// Check stat requirements
	stats := gameState.GetStats()
	for statName, requiredValue := range config.RequiredStats {
		if currentValue, exists := stats[statName]; !exists || currentValue < requiredValue {
			return false
		}
	}

	// Check interaction requirements
	interactionHistory := gameState.GetInteractionHistory()
	for requiredInteraction, requiredCount := range config.RequiredInteractions {
		actualCount := 0
		if interactions, exists := interactionHistory[requiredInteraction]; exists {
			actualCount = len(interactions)
		}
		if actualCount < requiredCount {
			return false
		}
	}

	return true
}

// performRecovery resolves a crisis and applies recovery bonuses
// Returns recovery event with forgiveness dialogue
func (crm *CrisisRecoveryManager) performRecovery(crisis *RelationshipCrisis, gameState *GameState) *TriggeredEvent {
	// Mark crisis as resolved
	crisis.IsActive = false

	// Apply recovery bonuses
	if crisis.RecoveryConfig != nil && len(crisis.RecoveryConfig.StatBonuses) > 0 {
		// Apply bonuses with recovery multiplier
		bonuses := make(map[string]float64)
		for statName, bonus := range crisis.RecoveryConfig.StatBonuses {
			bonuses[statName] = bonus * crm.recoveryBonus
		}
		gameState.ApplyInteractionEffects(bonuses)
	}

	// Clean up resolved crises
	crm.cleanupResolvedCrises()

	// Return recovery event
	return &TriggeredEvent{
		Name:        crisis.Name + "_recovered",
		Description: "Crisis resolved through player care",
		Effects:     crisis.RecoveryConfig.StatBonuses,
		Animations:  crisis.RecoveryConfig.RecoveryAnimations,
		Responses:   crisis.RecoveryConfig.ForgivenessResponses,
		Duration:    8 * time.Second, // Longer duration for important recovery moments
	}
}

// cleanupResolvedCrises removes inactive crises from the active list
// Keeps crisis list clean and prevents memory leaks
func (crm *CrisisRecoveryManager) cleanupResolvedCrises() {
	var activeCrises []RelationshipCrisis
	for _, crisis := range crm.activeCrises {
		if crisis.IsActive {
			activeCrises = append(activeCrises, crisis)
		}
	}
	crm.activeCrises = activeCrises
}

// hasCrisisType checks if a specific type of crisis is already active
// Prevents duplicate crises of the same type
func (crm *CrisisRecoveryManager) hasCrisisType(crisisType string) bool {
	for _, crisis := range crm.activeCrises {
		if crisis.IsActive && crisis.Name == crisisType {
			return true
		}
	}
	return false
}

// hasActiveCrises returns true if any crises are currently active
// Used to determine if crisis mode should be active
func (crm *CrisisRecoveryManager) hasActiveCrises() bool {
	for _, crisis := range crm.activeCrises {
		if crisis.IsActive {
			return true
		}
	}
	return false
}

// GetActiveCrises returns a copy of all active crises
// Safe for concurrent access
func (crm *CrisisRecoveryManager) GetActiveCrises() []RelationshipCrisis {
	if !crm.enabled {
		return nil
	}

	crm.mu.RLock()
	defer crm.mu.RUnlock()

	var activeCrises []RelationshipCrisis
	for _, crisis := range crm.activeCrises {
		if crisis.IsActive {
			activeCrises = append(activeCrises, crisis)
		}
	}

	return activeCrises
}

// GetCrisisStatus returns debug information about crisis management
// Used for testing and debugging crisis mechanics
func (crm *CrisisRecoveryManager) GetCrisisStatus() map[string]interface{} {
	crm.mu.RLock()
	defer crm.mu.RUnlock()

	status := map[string]interface{}{
		"enabled":       crm.enabled,
		"activeCrises":  len(crm.activeCrises),
		"maxCrises":     crm.maxActiveCrises,
		"thresholds":    crm.crisisThresholds,
		"recoveryBonus": crm.recoveryBonus,
		"lastCheck":     crm.lastCheck,
		"checkInterval": crm.checkInterval,
	}

	// Add details about active crises
	if len(crm.activeCrises) > 0 {
		crisisDetails := make([]map[string]interface{}, 0, len(crm.activeCrises))
		for _, crisis := range crm.activeCrises {
			if crisis.IsActive {
				detail := map[string]interface{}{
					"name":        crisis.Name,
					"severity":    crisis.Severity,
					"triggeredAt": crisis.TriggeredAt,
					"duration":    time.Since(crisis.TriggeredAt),
					"cause":       crisis.TriggerCause,
				}
				crisisDetails = append(crisisDetails, detail)
			}
		}
		status["crisisDetails"] = crisisDetails
	}

	return status
}

// SetEnabled allows runtime enabling/disabling of crisis management
// Useful for different character personality configurations
func (crm *CrisisRecoveryManager) SetEnabled(enabled bool) {
	crm.mu.Lock()
	defer crm.mu.Unlock()
	crm.enabled = enabled
}

// SetRecoveryBonus adjusts the bonus multiplier for recovery interactions
// Higher values make recovery more rewarding
func (crm *CrisisRecoveryManager) SetRecoveryBonus(bonus float64) {
	crm.mu.Lock()
	defer crm.mu.Unlock()
	crm.recoveryBonus = math.Max(1.0, bonus) // Minimum 1.0 (no penalty)
}
