package character

import (
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"
)

// GameState manages Tamagotchi-style stats and progression for a character
// This follows the "lazy programmer" approach using only Go standard library
// All game mechanics are configurable via JSON without custom code
type GameState struct {
	mu              sync.RWMutex
	Stats           map[string]*Stat  `json:"stats"`
	LastDecayUpdate time.Time         `json:"lastDecayUpdate"`
	CreationTime    time.Time         `json:"creationTime"`
	TotalPlayTime   time.Duration     `json:"totalPlayTime"`
	Config          *GameConfig       `json:"config,omitempty"`
	Progression     *ProgressionState `json:"progression,omitempty"`
	// Romance feature extensions (Dating Simulator Phase 1)
	RelationshipLevel  string                 `json:"relationshipLevel,omitempty"`
	InteractionHistory map[string][]time.Time `json:"interactionHistory,omitempty"`
	RomanceMemories    []RomanceMemory        `json:"romanceMemories,omitempty"`
}

// Stat represents a game statistic with boundaries and degradation rules
// All values are float64 to support precise calculations and gradual changes
type Stat struct {
	Current           float64 `json:"current"`
	Max               float64 `json:"max"`
	DegradationRate   float64 `json:"degradationRate"`   // Points per minute of decay
	CriticalThreshold float64 `json:"criticalThreshold"` // Threshold for critical state
}

// GameConfig holds game-wide settings that affect stat behavior
// These settings come from the character card's gameRules section
type GameConfig struct {
	StatsDecayInterval             time.Duration `json:"statsDecayInterval"` // How often stats decay
	CriticalStateAnimationPriority bool          `json:"criticalStateAnimationPriority"`
	MoodBasedAnimations            bool          `json:"moodBasedAnimations"`
}

// StatConfig represents the configuration for a stat from JSON
// This is used during character card loading to initialize stats
type StatConfig struct {
	Initial           float64 `json:"initial"`
	Max               float64 `json:"max"`
	DegradationRate   float64 `json:"degradationRate"`
	CriticalThreshold float64 `json:"criticalThreshold"`
}

// RomanceMemory represents a recorded romance interaction for the memory system
// Used to track relationship history and enable future complex romance features
type RomanceMemory struct {
	Timestamp       time.Time          `json:"timestamp"`
	InteractionType string             `json:"interactionType"`
	StatsBefore     map[string]float64 `json:"statsBefore"`
	StatsAfter      map[string]float64 `json:"statsAfter"`
	Response        string             `json:"response"`
}

// NewGameState creates a new game state from stat configurations
// Uses current time as baseline for all time-based calculations
func NewGameState(statConfigs map[string]StatConfig, config *GameConfig) *GameState {
	gs := &GameState{
		Stats:              make(map[string]*Stat),
		LastDecayUpdate:    time.Now(),
		CreationTime:       time.Now(),
		TotalPlayTime:      0,
		Config:             config,
		Progression:        nil, // Will be set separately if progression is enabled
		RelationshipLevel:  "Stranger", // Default relationship level
		InteractionHistory: make(map[string][]time.Time),
		RomanceMemories:    make([]RomanceMemory, 0),
	}

	// Initialize stats from configuration
	for name, config := range statConfigs {
		gs.Stats[name] = &Stat{
			Current:           config.Initial,
			Max:               config.Max,
			DegradationRate:   config.DegradationRate,
			CriticalThreshold: config.CriticalThreshold,
		}
	}

	return gs
}

// SetProgression sets the progression system for this game state
func (gs *GameState) SetProgression(progressionConfig *ProgressionConfig) {
	if gs == nil || progressionConfig == nil {
		return
	}

	gs.mu.Lock()
	defer gs.mu.Unlock()

	gs.Progression = NewProgressionState(progressionConfig)
}

// Update applies time-based stat degradation and returns triggered states
// This method is called from the main character update loop
// Returns list of states that should trigger animations (e.g., "hungry", "critical")
func (gs *GameState) Update(elapsed time.Duration) []string {
	if gs == nil {
		return nil
	}

	gs.mu.Lock()
	defer gs.mu.Unlock()

	now := time.Now()
	timeSinceLastDecay := now.Sub(gs.LastDecayUpdate)

	// Update total play time
	gs.TotalPlayTime += elapsed

	// Update progression if enabled
	var levelChanged bool
	var newAchievements []string
	if gs.Progression != nil {
		levelChanged, newAchievements = gs.Progression.Update(gs, elapsed)
	}

	// Only apply degradation if enough time has passed
	decayInterval := time.Minute
	if gs.Config != nil && gs.Config.StatsDecayInterval > 0 {
		decayInterval = gs.Config.StatsDecayInterval
	}

	if timeSinceLastDecay < decayInterval {
		// Even if no stat degradation, return progression-based states
		triggeredStates := make([]string, 0)
		if levelChanged {
			triggeredStates = append(triggeredStates, "level_up")
		}
		for _, achievement := range newAchievements {
			triggeredStates = append(triggeredStates, fmt.Sprintf("achievement_%s", achievement))
		}
		return triggeredStates
	}

	// Apply degradation to all stats
	minutesElapsed := timeSinceLastDecay.Minutes()
	triggeredStates := make([]string, 0)

	for name, stat := range gs.Stats {
		if stat.DegradationRate > 0 {
			// Calculate degradation amount
			degradationAmount := stat.DegradationRate * minutesElapsed
			oldValue := stat.Current

			// Apply degradation with minimum bound of 0
			stat.Current = math.Max(0, stat.Current-degradationAmount)

			// Check if we crossed the critical threshold
			if oldValue > stat.CriticalThreshold && stat.Current <= stat.CriticalThreshold {
				triggeredStates = append(triggeredStates, fmt.Sprintf("%s_critical", name))
			}

			// Check for specific stat-based states
			if stat.Current <= stat.CriticalThreshold {
				switch name {
				case "hunger":
					triggeredStates = append(triggeredStates, "hungry")
				case "happiness":
					triggeredStates = append(triggeredStates, "sad")
				case "health":
					triggeredStates = append(triggeredStates, "sick")
				case "energy":
					triggeredStates = append(triggeredStates, "tired")
				}
			}
		}
	}

	// Add progression-based states
	if levelChanged {
		triggeredStates = append(triggeredStates, "level_up")
	}
	for _, achievement := range newAchievements {
		triggeredStates = append(triggeredStates, fmt.Sprintf("achievement_%s", achievement))
	}

	gs.LastDecayUpdate = now
	return triggeredStates
}

// ApplyInteractionEffects modifies stats based on character interactions
// Effects map contains stat names and the amount to modify (can be positive or negative)
// All modifications respect stat boundaries (0 to Max)
func (gs *GameState) ApplyInteractionEffects(effects map[string]float64) {
	if gs == nil || len(effects) == 0 {
		return
	}

	gs.mu.Lock()
	defer gs.mu.Unlock()

	for statName, change := range effects {
		if stat, exists := gs.Stats[statName]; exists {
			// Apply change with bounds checking
			newValue := stat.Current + change
			stat.Current = math.Max(0, math.Min(stat.Max, newValue))
		}
	}
}

// GetStats returns a copy of current stat values for external access
// Returns map is safe to read without affecting the game state
func (gs *GameState) GetStats() map[string]float64 {
	if gs == nil {
		return nil
	}

	gs.mu.RLock()
	defer gs.mu.RUnlock()

	stats := make(map[string]float64)
	for name, stat := range gs.Stats {
		stats[name] = stat.Current
	}

	return stats
}

// GetStat returns the current value of a specific stat
// Returns 0 if the stat doesn't exist
func (gs *GameState) GetStat(name string) float64 {
	if gs == nil {
		return 0
	}

	gs.mu.RLock()
	defer gs.mu.RUnlock()

	if stat, exists := gs.Stats[name]; exists {
		return stat.Current
	}

	return 0
}

// GetCriticalStates returns a list of stats that are below their critical thresholds
// Used to determine if character should show critical animations
func (gs *GameState) GetCriticalStates() []string {
	if gs == nil {
		return nil
	}

	gs.mu.RLock()
	defer gs.mu.RUnlock()

	criticalStates := make([]string, 0)
	for name, stat := range gs.Stats {
		if stat.Current <= stat.CriticalThreshold {
			criticalStates = append(criticalStates, name)
		}
	}

	return criticalStates
}

// GetAge returns how long the character has existed
func (gs *GameState) GetAge() time.Duration {
	if gs == nil {
		return 0
	}

	gs.mu.RLock()
	defer gs.mu.RUnlock()

	return time.Since(gs.CreationTime)
}

// GetPlayTime returns total time the character has been active
func (gs *GameState) GetPlayTime() time.Duration {
	if gs == nil {
		return 0
	}

	gs.mu.RLock()
	defer gs.mu.RUnlock()

	return gs.TotalPlayTime
}

// CanSatisfyRequirements checks if current stats meet interaction requirements
// Requirements map specifies min/max values that stats must satisfy
// Used to gate interactions behind stat conditions (e.g., can't play if too tired)
func (gs *GameState) CanSatisfyRequirements(requirements map[string]map[string]float64) bool {
	if gs == nil || len(requirements) == 0 {
		return true
	}

	gs.mu.RLock()
	defer gs.mu.RUnlock()

	for statName, constraints := range requirements {
		stat, exists := gs.Stats[statName]
		if !exists {
			return false
		}

		// Check minimum requirement
		if minVal, hasMin := constraints["min"]; hasMin {
			if stat.Current < minVal {
				return false
			}
		}

		// Check maximum requirement
		if maxVal, hasMax := constraints["max"]; hasMax {
			if stat.Current > maxVal {
				return false
			}
		}
	}

	return true
}

// GetStatPercentage returns stat value as percentage (0-100) of maximum
// Useful for UI displays like progress bars
func (gs *GameState) GetStatPercentage(name string) float64 {
	if gs == nil {
		return 0
	}

	gs.mu.RLock()
	defer gs.mu.RUnlock()

	stat, exists := gs.Stats[name]
	if !exists || stat.Max == 0 {
		return 0
	}

	percentage := (stat.Current / stat.Max) * 100
	return math.Max(0, math.Min(100, percentage))
}

// GetOverallMood calculates character's general mood based on all stats
// Returns a value from 0 (critical) to 100 (excellent)
// Used for mood-based animation selection
func (gs *GameState) GetOverallMood() float64 {
	if gs == nil {
		return 50 // Neutral mood when no game state
	}

	gs.mu.RLock()
	defer gs.mu.RUnlock()

	if len(gs.Stats) == 0 {
		return 50
	}

	totalPercentage := 0.0
	for _, stat := range gs.Stats {
		if stat.Max > 0 {
			percentage := (stat.Current / stat.Max) * 100
			totalPercentage += percentage
		}
	}

	return totalPercentage / float64(len(gs.Stats))
}

// GetProgression returns a reference to the progression state
func (gs *GameState) GetProgression() *ProgressionState {
	if gs == nil {
		return nil
	}

	gs.mu.RLock()
	defer gs.mu.RUnlock()

	return gs.Progression
}

// RecordInteraction records an interaction for progression tracking
func (gs *GameState) RecordInteraction(interactionType string) {
	if gs == nil || gs.Progression == nil {
		return
	}

	gs.mu.RLock()
	progression := gs.Progression
	gs.mu.RUnlock()

	if progression != nil {
		progression.RecordInteraction(interactionType)
	}
}

// GetCurrentSize returns the character size based on progression level
func (gs *GameState) GetCurrentSize() int {
	if gs == nil || gs.Progression == nil {
		return 128 // Default size
	}

	gs.mu.RLock()
	defer gs.mu.RUnlock()

	return gs.Progression.GetCurrentSize()
}

// GetLevelAnimation returns level-specific animation if available
func (gs *GameState) GetLevelAnimation(animationName string) (string, bool) {
	if gs == nil || gs.Progression == nil {
		return "", false
	}

	gs.mu.RLock()
	defer gs.mu.RUnlock()

	return gs.Progression.GetLevelAnimation(animationName)
}

// MarshalJSON implements custom JSON marshaling for save files
// Ensures proper serialization of time.Duration fields
func (gs *GameState) MarshalJSON() ([]byte, error) {
	if gs == nil {
		return []byte("null"), nil
	}

	gs.mu.RLock()
	defer gs.mu.RUnlock()

	// Create a serializable version of the struct
	type Alias GameState
	return json.Marshal(&struct {
		*Alias
		TotalPlayTimeNanos int64 `json:"totalPlayTimeNanos"`
	}{
		Alias:              (*Alias)(gs),
		TotalPlayTimeNanos: int64(gs.TotalPlayTime),
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for save files
// Handles time.Duration deserialization properly
func (gs *GameState) UnmarshalJSON(data []byte) error {
	type Alias GameState
	aux := &struct {
		*Alias
		TotalPlayTimeNanos int64 `json:"totalPlayTimeNanos"`
	}{
		Alias: (*Alias)(gs),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	gs.TotalPlayTime = time.Duration(aux.TotalPlayTimeNanos)
	return nil
}

// Validate ensures the game state has consistent data
// Used during loading to verify save file integrity
func (gs *GameState) Validate() error {
	if gs == nil {
		return fmt.Errorf("game state is nil")
	}

	if len(gs.Stats) == 0 {
		return fmt.Errorf("game state must have at least one stat")
	}

	for name, stat := range gs.Stats {
		if err := gs.validateStat(name, stat); err != nil {
			return fmt.Errorf("stat '%s': %w", name, err)
		}
	}

	if gs.LastDecayUpdate.IsZero() {
		gs.LastDecayUpdate = time.Now()
	}

	if gs.CreationTime.IsZero() {
		gs.CreationTime = time.Now()
	}

	return nil
}

// validateStat checks individual stat consistency
func (gs *GameState) validateStat(name string, stat *Stat) error {
	if stat == nil {
		return fmt.Errorf("stat is nil")
	}

	if stat.Max <= 0 {
		return fmt.Errorf("max value must be positive, got %f", stat.Max)
	}

	if stat.Current < 0 {
		return fmt.Errorf("current value cannot be negative, got %f", stat.Current)
	}

	if stat.Current > stat.Max {
		return fmt.Errorf("current value (%f) cannot exceed max (%f)", stat.Current, stat.Max)
	}

	if stat.DegradationRate < 0 {
		return fmt.Errorf("degradation rate cannot be negative, got %f", stat.DegradationRate)
	}

	if stat.CriticalThreshold < 0 || stat.CriticalThreshold > stat.Max {
		return fmt.Errorf("critical threshold (%f) must be between 0 and max (%f)", stat.CriticalThreshold, stat.Max)
	}

	return nil
}

// Romance-related methods for Dating Simulator features (Phase 2 implementation)

// RecordRomanceInteraction records a romance interaction in the memory system
// Tracks interaction history for relationship progression and future features
func (gs *GameState) RecordRomanceInteraction(interactionType, response string, statsBefore, statsAfter map[string]float64) {
	if gs == nil {
		return
	}

	gs.mu.Lock()
	defer gs.mu.Unlock()

	// Initialize maps if they don't exist
	if gs.InteractionHistory == nil {
		gs.InteractionHistory = make(map[string][]time.Time)
	}
	if gs.RomanceMemories == nil {
		gs.RomanceMemories = make([]RomanceMemory, 0)
	}

	// Record in interaction history
	gs.InteractionHistory[interactionType] = append(
		gs.InteractionHistory[interactionType],
		time.Now(),
	)

	// Record detailed memory
	memory := RomanceMemory{
		Timestamp:       time.Now(),
		InteractionType: interactionType,
		StatsBefore:     statsBefore,
		StatsAfter:      statsAfter,
		Response:        response,
	}
	gs.RomanceMemories = append(gs.RomanceMemories, memory)

	// Keep only last 50 memories to prevent unbounded growth
	if len(gs.RomanceMemories) > 50 {
		gs.RomanceMemories = gs.RomanceMemories[len(gs.RomanceMemories)-50:]
	}
}

// GetInteractionCount returns the number of times a specific interaction has been used
// Used for progression requirements and achievement tracking
func (gs *GameState) GetInteractionCount(interactionType string) int {
	if gs == nil || gs.InteractionHistory == nil {
		return 0
	}

	gs.mu.RLock()
	defer gs.mu.RUnlock()

	if interactions, exists := gs.InteractionHistory[interactionType]; exists {
		return len(interactions)
	}
	return 0
}

// GetRelationshipLevel returns the current relationship level
// Returns "Stranger" as default if no level has been set
func (gs *GameState) GetRelationshipLevel() string {
	if gs == nil {
		return "Stranger"
	}

	gs.mu.RLock()
	defer gs.mu.RUnlock()

	if gs.RelationshipLevel == "" {
		return "Stranger" // Default level
	}
	return gs.RelationshipLevel
}

// UpdateRelationshipLevel checks progression levels and updates relationship level if requirements are met
// Returns true if the relationship level changed, false otherwise
func (gs *GameState) UpdateRelationshipLevel(progressionConfig *ProgressionConfig) bool {
	if gs == nil || progressionConfig == nil {
		return false
	}

	gs.mu.Lock()
	defer gs.mu.Unlock()

	oldLevel := gs.RelationshipLevel
	newLevel := oldLevel

	// Check each relationship level in order to find the highest level the character qualifies for
	for _, level := range progressionConfig.Levels {
		if gs.meetsRelationshipRequirements(level.Requirement) {
			newLevel = level.Name
		}
	}

	// Update the relationship level if it changed
	if newLevel != oldLevel {
		gs.RelationshipLevel = newLevel
		return true
	}

	return false
}

// meetsRelationshipRequirements checks if current stats meet the requirements for a relationship level
// Supports age-based, stat-based, and other progression requirements
func (gs *GameState) meetsRelationshipRequirements(requirements map[string]int64) bool {
	if gs == nil || len(requirements) == 0 {
		return true
	}

	for statName, threshold := range requirements {
		if statName == "age" {
			// Special handling for age requirement (in seconds)
			ageSeconds := int64(time.Since(gs.CreationTime).Seconds())
			if ageSeconds < threshold {
				return false
			}
			continue
		}

		// Check regular stat requirements
		if stat, exists := gs.Stats[statName]; exists {
			if int64(stat.Current) < threshold {
				return false
			}
		} else {
			// If stat doesn't exist, requirement can't be met
			return false
		}
	}

	return true
}
