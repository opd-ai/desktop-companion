package character

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

// GiftManager extends existing character system to handle gift interactions
// Follows the "lazy programmer" approach by reusing existing patterns for
// stat effects, memory management, and personality-based responses
type GiftManager struct {
	character   *CharacterCard
	gameState   *GameState
	giftCatalog map[string]*GiftDefinition
	mu          sync.RWMutex
}

// NewGiftManager creates a gift manager that integrates with existing systems
// Reuses existing CharacterCard and GameState without modification
func NewGiftManager(character *CharacterCard, gameState *GameState) *GiftManager {
	return &GiftManager{
		character:   character,
		gameState:   gameState,
		giftCatalog: make(map[string]*GiftDefinition),
	}
}

// LoadGiftCatalog loads gift definitions from assets/gifts/ directory
// Reuses existing LoadGiftCatalog function from gift_definition.go
func (gm *GiftManager) LoadGiftCatalog(giftsPath string) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	catalog, err := LoadGiftCatalog(giftsPath)
	if err != nil {
		return fmt.Errorf("failed to load gift catalog: %w", err)
	}

	gm.giftCatalog = catalog
	return nil
}

// GetGiftCatalog returns a copy of the loaded gift catalog
// Provides thread-safe access to gift definitions
func (gm *GiftManager) GetGiftCatalog() map[string]*GiftDefinition {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	// Return a copy to prevent external modification
	catalog := make(map[string]*GiftDefinition, len(gm.giftCatalog))
	for id, gift := range gm.giftCatalog {
		catalog[id] = gift
	}
	return catalog
}

// GetAvailableGifts returns gifts user can currently give based on relationship and stats
// Filters gifts based on unlock requirements following existing requirement patterns
func (gm *GiftManager) GetAvailableGifts() []*GiftDefinition {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	var availableGifts []*GiftDefinition
	for _, gift := range gm.giftCatalog {
		if gm.canGiveGift(gift) {
			availableGifts = append(availableGifts, gift)
		}
	}
	return availableGifts
}

// canGiveGift checks if a gift can be given based on unlock requirements
// Reuses existing requirement checking patterns from interaction system
func (gm *GiftManager) canGiveGift(gift *GiftDefinition) bool {
	// Check relationship level requirement if specified
	if relationshipReq, exists := gift.Properties.UnlockRequirements["relationshipLevel"]; exists {
		if relationshipStr, ok := relationshipReq.(string); ok {
			if !gm.meetsRelationshipLevel(relationshipStr) {
				return false
			}
		}
	}

	// Check stat requirements if specified
	if statsReq, exists := gift.Properties.UnlockRequirements["stats"]; exists {
		if statsMap, ok := statsReq.(map[string]interface{}); ok {
			if !gm.meetsStatRequirements(statsMap) {
				return false
			}
		}
	}

	// Check cooldown if gift has cooldown properties
	if gift.Properties.CooldownSeconds > 0 {
		if gm.IsGiftOnCooldown(gift.ID) {
			return false
		}
	}

	return true
}

// meetsRelationshipLevel checks if current relationship level meets requirement
// Reuses existing relationship progression system
func (gm *GiftManager) meetsRelationshipLevel(required string) bool {
	if gm.gameState == nil {
		return required == "Stranger" // Default level
	}

	gm.gameState.mu.RLock()
	current := gm.gameState.RelationshipLevel
	gm.gameState.mu.RUnlock()

	if current == "" {
		current = "Stranger"
	}

	// Define relationship hierarchy (reuses existing progression order)
	levels := map[string]int{
		"Stranger":          0,
		"Acquaintance":      1,
		"Friend":            2,
		"Close Friend":      3,
		"Romantic Interest": 4,
		"Partner":           5,
	}

	currentLevel, exists := levels[current]
	if !exists {
		currentLevel = 0
	}

	requiredLevel, exists := levels[required]
	if !exists {
		return false
	}

	return currentLevel >= requiredLevel
}

// meetsStatRequirements checks if current stats meet the gift requirements
// Reuses existing stat checking patterns from interaction system
func (gm *GiftManager) meetsStatRequirements(requirements map[string]interface{}) bool {
	if !gm.validateGameState() {
		return false
	}

	gm.gameState.mu.RLock()
	defer gm.gameState.mu.RUnlock()

	for statName, requirement := range requirements {
		if !gm.validateStatExists(statName) {
			return false
		}

		if !gm.checkStatRequirement(statName, requirement) {
			return false
		}
	}

	return true
}

// validateGameState checks if the game state is available for stat validation
func (gm *GiftManager) validateGameState() bool {
	return gm.gameState != nil
}

// validateStatExists checks if the required stat exists in the game state
func (gm *GiftManager) validateStatExists(statName string) bool {
	stat := gm.gameState.Stats[statName]
	return stat != nil
}

// checkStatRequirement validates a single stat against its requirement criteria
func (gm *GiftManager) checkStatRequirement(statName string, requirement interface{}) bool {
	stat := gm.gameState.Stats[statName]
	reqMap, ok := requirement.(map[string]interface{})
	if !ok {
		return true // No specific requirements to check
	}

	return gm.checkMinRequirement(stat, reqMap) && gm.checkMaxRequirement(stat, reqMap)
}

// checkMinRequirement validates the minimum value requirement for a stat
func (gm *GiftManager) checkMinRequirement(stat *Stat, reqMap map[string]interface{}) bool {
	minVal, exists := reqMap["min"]
	if !exists {
		return true // No minimum requirement
	}

	minFloat, ok := minVal.(float64)
	if !ok {
		return true // Invalid requirement format, skip check
	}

	return stat.Current >= minFloat
}

// checkMaxRequirement validates the maximum value requirement for a stat
func (gm *GiftManager) checkMaxRequirement(stat *Stat, reqMap map[string]interface{}) bool {
	maxVal, exists := reqMap["max"]
	if !exists {
		return true // No maximum requirement
	}

	maxFloat, ok := maxVal.(float64)
	if !ok {
		return true // Invalid requirement format, skip check
	}

	return stat.Current <= maxFloat
}

// GiveGift processes gift giving with personality-aware responses
// Integrates with existing stat system, memory system, and animation system
func (gm *GiftManager) GiveGift(giftID, notes string) (*GiftResponse, error) {
	gm.mu.RLock()
	gift, exists := gm.giftCatalog[giftID]
	gm.mu.RUnlock()

	if !exists {
		return &GiftResponse{
			ErrorMessage: fmt.Sprintf("Gift '%s' not found", giftID),
		}, fmt.Errorf("gift not found: %s", giftID)
	}

	if !gm.canGiveGift(gift) {
		return &GiftResponse{
			ErrorMessage: "Gift requirements not met",
		}, fmt.Errorf("gift requirements not met for: %s", giftID)
	}

	// Apply personality modifiers to stat effects (reuses existing personality system)
	modifiedEffects := gm.applyPersonalityModifiers(gift)

	// Apply stat effects using existing GameState methods
	var actualEffects map[string]float64
	if gm.gameState != nil {
		actualEffects = gm.applyStatEffects(modifiedEffects)
	} else {
		actualEffects = modifiedEffects
	}

	// Select response based on personality and gift preferences
	response := gm.selectResponse(gift, notes)

	// Select animation based on gift effects and personality
	animation := gm.selectAnimation(gift)

	// Record gift memory using existing memory patterns
	memoryCreated := gm.recordGiftMemory(gift, notes, response, actualEffects)

	return &GiftResponse{
		Response:      response,
		Animation:     animation,
		StatEffects:   actualEffects,
		MemoryCreated: memoryCreated,
	}, nil
}

// applyPersonalityModifiers modifies gift effects based on character personality
// Reuses existing personality trait system from romance features
func (gm *GiftManager) applyPersonalityModifiers(gift *GiftDefinition) map[string]float64 {
	effects := make(map[string]float64)
	for stat, value := range gift.GiftEffects.Immediate.Stats {
		effects[stat] = value
	}

	if gm.character.Personality == nil {
		return effects
	}

	// Apply personality modifiers based on character traits
	for traitName, traitValue := range gm.character.Personality.Traits {
		if modifiers, exists := gift.PersonalityModifiers[traitName]; exists {
			for stat, modifier := range modifiers {
				if baseValue, exists := effects[stat]; exists {
					// Apply modifier weighted by trait strength
					effects[stat] = baseValue * (1.0 + (modifier-1.0)*traitValue)
				}
			}
		}
	}

	return effects
}

// applyStatEffects applies stat changes using existing GameState mechanisms
// Reuses existing ApplyInteractionEffects patterns
func (gm *GiftManager) applyStatEffects(effects map[string]float64) map[string]float64 {
	gm.gameState.mu.Lock()
	defer gm.gameState.mu.Unlock()

	actualEffects := make(map[string]float64)
	for statName, change := range effects {
		stat := gm.gameState.Stats[statName]
		if stat != nil {
			oldValue := stat.Current
			stat.Current = math.Min(stat.Max, math.Max(0, stat.Current+change))
			actualEffects[statName] = stat.Current - oldValue
		}
	}

	return actualEffects
}

// selectResponse chooses an appropriate response based on personality and preferences
// Reuses existing response selection patterns from dialog system
func (gm *GiftManager) selectResponse(gift *GiftDefinition, notes string) string {
	responses := gift.GiftEffects.Immediate.Responses

	// Check for personality-specific responses
	if gm.character.GiftSystem != nil && gm.character.Personality != nil {
		for traitName, traitValue := range gm.character.Personality.Traits {
			if traitValue > 0.6 { // Strong trait influence
				if personalityResp, exists := gm.character.GiftSystem.Preferences.PersonalityResponses[traitName]; exists {
					if len(personalityResp.GiftReceived) > 0 {
						responses = personalityResp.GiftReceived
						break
					}
				}
			}
		}
	}

	if len(responses) == 0 {
		return "Thank you for the gift!"
	}

	// Use time-based seeding for pseudo-randomness (follows existing pattern)
	rand.Seed(time.Now().UnixNano())
	return responses[rand.Intn(len(responses))]
}

// selectAnimation chooses appropriate animation based on gift effects
// Reuses existing animation selection patterns
func (gm *GiftManager) selectAnimation(gift *GiftDefinition) string {
	animations := gift.GiftEffects.Immediate.Animations
	if len(animations) == 0 {
		return "happy" // Default fallback
	}

	// Use time-based seeding for pseudo-randomness (follows existing pattern)
	rand.Seed(time.Now().UnixNano())
	return animations[rand.Intn(len(animations))]
}

// recordGiftMemory creates a memory entry for the gift interaction
// Extends existing romance memory system with gift-specific memory type
func (gm *GiftManager) recordGiftMemory(gift *GiftDefinition, notes, response string, effects map[string]float64) bool {
	if gm.gameState == nil {
		return false
	}

	// Create gift memory using existing memory structure pattern
	giftMemory := GiftMemory{
		Timestamp:        time.Now(),
		GiftID:           gift.ID,
		GiftName:         gift.Name,
		Notes:            notes,
		Response:         response,
		StatEffects:      effects,
		MemoryImportance: gift.GiftEffects.Memory.Importance,
		Tags:             gift.GiftEffects.Memory.Tags,
		EmotionalTone:    gift.GiftEffects.Memory.EmotionalTone,
	}

	gm.gameState.mu.Lock()
	defer gm.gameState.mu.Unlock()

	// Initialize gift memories slice if needed
	if gm.gameState.GiftMemories == nil {
		gm.gameState.GiftMemories = make([]GiftMemory, 0)
	}

	// Add new memory entry
	gm.gameState.GiftMemories = append(gm.gameState.GiftMemories, giftMemory)

	// Limit memory size to prevent unbounded growth (follows existing pattern)
	maxMemories := 100
	if len(gm.gameState.GiftMemories) > maxMemories {
		// Keep most recent memories
		gm.gameState.GiftMemories = gm.gameState.GiftMemories[len(gm.gameState.GiftMemories)-maxMemories:]
	}

	return true
}

// GetGiftMemories returns a copy of gift memories for analysis
// Provides thread-safe access to gift interaction history
func (gm *GiftManager) GetGiftMemories() []GiftMemory {
	if gm.gameState == nil {
		return nil
	}

	gm.gameState.mu.RLock()
	defer gm.gameState.mu.RUnlock()

	if gm.gameState.GiftMemories == nil {
		return nil
	}

	// Return a copy to prevent external modification
	memories := make([]GiftMemory, len(gm.gameState.GiftMemories))
	copy(memories, gm.gameState.GiftMemories)
	return memories
}

// IsGiftSystemEnabled checks if the character has gift system enabled
// Helper method for UI integration
func (gm *GiftManager) IsGiftSystemEnabled() bool {
	return gm.character.GiftSystem != nil && gm.character.GiftSystem.Enabled
}

// GetGiftPreferences returns character gift preferences for UI hints
// Provides information for better gift selection UX
func (gm *GiftManager) GetGiftPreferences() *GiftPreferences {
	if gm.character.GiftSystem == nil {
		return nil
	}
	return &gm.character.GiftSystem.Preferences
}

// GetGiftCooldownRemaining returns the remaining cooldown time for a specific gift type
// Thread-safe method that checks gift memories for the last usage timestamp
func (gm *GiftManager) GetGiftCooldownRemaining(giftID string) time.Duration {
	if gm.gameState == nil {
		return 0 // No game state means no cooldown tracking
	}

	gm.gameState.mu.RLock()
	defer gm.gameState.mu.RUnlock()

	// Get the gift definition to check cooldown settings
	gm.mu.RLock()
	gift, exists := gm.giftCatalog[giftID]
	gm.mu.RUnlock()

	if !exists || gift.Properties.CooldownSeconds <= 0 {
		return 0 // No gift found or no cooldown configured
	}

	// Find the most recent use of this gift
	var lastUsed time.Time
	for i := len(gm.gameState.GiftMemories) - 1; i >= 0; i-- {
		memory := gm.gameState.GiftMemories[i]
		if memory.GiftID == giftID {
			lastUsed = memory.Timestamp
			break
		}
	}

	if lastUsed.IsZero() {
		return 0 // Gift never used, no cooldown
	}

	// Calculate remaining cooldown time
	cooldownDuration := time.Duration(gift.Properties.CooldownSeconds) * time.Second
	elapsed := time.Since(lastUsed)
	remaining := cooldownDuration - elapsed

	if remaining > 0 {
		return remaining
	}

	return 0 // Cooldown expired
}

// IsGiftOnCooldown checks if a gift is currently on cooldown
// Returns true if the gift cannot be given due to cooldown restrictions
func (gm *GiftManager) IsGiftOnCooldown(giftID string) bool {
	return gm.GetGiftCooldownRemaining(giftID) > 0
}

// AddGiftToTestCatalog adds a gift definition to the catalog for testing purposes
// This method is intended for use in unit tests only
func (gm *GiftManager) AddGiftToTestCatalog(gift *GiftDefinition) {
	gm.mu.Lock()
	defer gm.mu.Unlock()
	gm.giftCatalog[gift.ID] = gift
}
