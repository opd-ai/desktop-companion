package character

import (
	"desktop-companion/internal/dialog"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
)

// RomanceMemory represents a memory of a romance interaction for tracking
type RomanceMemory struct {
	Timestamp       time.Time          `json:"timestamp"`
	InteractionType string             `json:"interactionType"`
	StatsBefore     map[string]float64 `json:"statsBefore"`
	StatsAfter      map[string]float64 `json:"statsAfter"`
	Response        string             `json:"response"`
}

// DialogMemory represents a memory of a dialog interaction for learning and adaptation
type DialogMemory struct {
	Timestamp        time.Time            `json:"timestamp"`
	Trigger          string               `json:"trigger"`
	Response         string               `json:"response"`
	EmotionalTone    string               `json:"emotionalTone"`
	Topics           []string             `json:"topics"`
	MemoryImportance float64              `json:"memoryImportance"`
	BackendUsed      string               `json:"backendUsed"`
	Confidence       float64              `json:"confidence"`
	UserFeedback     *dialog.UserFeedback `json:"userFeedback,omitempty"`
	IsFavorite       bool                 `json:"isFavorite,omitempty"`     // Whether user marked this response as favorite
	FavoriteRating   float64              `json:"favoriteRating,omitempty"` // User rating for this response (1-5 stars)
}

// GameState manages Tamagotchi-style stats and progression for a character
// This follows the "lazy programmer" approach using only Go standard library
// All game mechanics are configurable via JSON without custom code
type GameState struct {
	mu                 sync.RWMutex
	Stats              map[string]*Stat       `json:"stats"`
	LastDecayUpdate    time.Time              `json:"lastDecayUpdate"`
	CreationTime       time.Time              `json:"creationTime"`
	TotalPlayTime      time.Duration          `json:"totalPlayTime"`
	Config             *GameConfig            `json:"config,omitempty"`
	Progression        *ProgressionState      `json:"progression,omitempty"`
	RelationshipLevel  string                 `json:"relationshipLevel,omitempty"`
	InteractionHistory map[string][]time.Time `json:"interactionHistory,omitempty"`
	RomanceMemories    []RomanceMemory        `json:"romanceMemories,omitempty"`
	DialogMemories     []DialogMemory         `json:"dialogMemories,omitempty"`
	GiftMemories       []GiftMemory           `json:"giftMemories,omitempty"`
	recentAchievements []AchievementDetails   // Non-persistent field for UI notifications
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

// NewGameState creates a new game state from stat configurations
// Uses current time as baseline for all time-based calculations
func NewGameState(statConfigs map[string]StatConfig, config *GameConfig) *GameState {
	gs := &GameState{
		Stats:              make(map[string]*Stat),
		LastDecayUpdate:    time.Now(),
		CreationTime:       time.Now(),
		TotalPlayTime:      0,
		Config:             config,
		Progression:        nil,        // Will be set separately if progression is enabled
		RelationshipLevel:  "Stranger", // Default relationship level
		InteractionHistory: make(map[string][]time.Time),
		RomanceMemories:    make([]RomanceMemory, 0),
		DialogMemories:     make([]DialogMemory, 0),
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
	levelChanged, newAchievements := gs.updateProgression(elapsed)

	// Store achievement details for UI retrieval
	gs.recentAchievements = newAchievements

	// Check if enough time has passed for degradation
	decayInterval := gs.calculateDecayInterval()
	if timeSinceLastDecay < decayInterval {
		return gs.buildProgressionStates(levelChanged, newAchievements)
	}

	// Apply stat degradation and collect triggered states
	triggeredStates := gs.applyStatDegradation(timeSinceLastDecay)

	// Add progression-based states
	triggeredStates = append(triggeredStates, gs.buildProgressionStates(levelChanged, newAchievements)...)

	gs.LastDecayUpdate = now
	return triggeredStates
}

// updateProgression processes progression updates and returns whether level changed and new achievements
func (gs *GameState) updateProgression(elapsed time.Duration) (bool, []AchievementDetails) {
	var levelChanged bool
	var newAchievements []AchievementDetails
	if gs.Progression != nil {
		levelChanged, newAchievements = gs.Progression.Update(gs, elapsed)
	}
	return levelChanged, newAchievements
}

// GetRecentAchievements returns and clears any recently earned achievements
// This allows the UI to retrieve new achievements for notifications
func (gs *GameState) GetRecentAchievements() []AchievementDetails {
	if gs == nil {
		return nil
	}

	gs.mu.Lock()
	defer gs.mu.Unlock()

	achievements := gs.recentAchievements
	gs.recentAchievements = nil // Clear after retrieval
	return achievements
}

// calculateDecayInterval determines the time interval for stat degradation
func (gs *GameState) calculateDecayInterval() time.Duration {
	decayInterval := time.Minute
	if gs.Config != nil && gs.Config.StatsDecayInterval > 0 {
		decayInterval = gs.Config.StatsDecayInterval
	}
	return decayInterval
}

// buildProgressionStates creates triggered states from progression events
func (gs *GameState) buildProgressionStates(levelChanged bool, newAchievements []AchievementDetails) []string {
	triggeredStates := make([]string, 0)
	if levelChanged {
		triggeredStates = append(triggeredStates, "level_up")
	}
	for _, achievement := range newAchievements {
		triggeredStates = append(triggeredStates, fmt.Sprintf("achievement_%s", achievement.Name))
	}
	return triggeredStates
}

// applyStatDegradation processes stat degradation and returns triggered states
func (gs *GameState) applyStatDegradation(timeSinceLastDecay time.Duration) []string {
	minutesElapsed := timeSinceLastDecay.Minutes()
	triggeredStates := make([]string, 0)

	for name, stat := range gs.Stats {
		if stat.DegradationRate > 0 {
			statStates := gs.processStatDegradation(name, stat, minutesElapsed)
			triggeredStates = append(triggeredStates, statStates...)
		}
	}

	return triggeredStates
}

// processStatDegradation handles degradation for a single stat and returns triggered states
func (gs *GameState) processStatDegradation(name string, stat *Stat, minutesElapsed float64) []string {
	triggeredStates := make([]string, 0)

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
		stateMapping := gs.getStatStateMapping(name)
		if stateMapping != "" {
			triggeredStates = append(triggeredStates, stateMapping)
		}
	}

	return triggeredStates
}

// getStatStateMapping returns the animation state for a critical stat
func (gs *GameState) getStatStateMapping(statName string) string {
	switch statName {
	case "hunger":
		return "hungry"
	case "happiness":
		return "sad"
	case "health":
		return "sick"
	case "energy":
		return "tired"
	default:
		return ""
	}
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

// CanSatisfyRomanceRequirements checks if current state meets romance event requirements
// Supports enhanced conditions like relationship level, interaction history, and memory patterns
func (gs *GameState) CanSatisfyRomanceRequirements(conditions map[string]map[string]float64) bool {
	if gs == nil || len(conditions) == 0 {
		return true
	}

	gs.mu.RLock()
	defer gs.mu.RUnlock()

	// First check standard stat requirements
	if !gs.canSatisfyStatRequirements(conditions) {
		return false
	}

	// Check relationship-specific requirements
	return gs.canSatisfyRelationshipRequirements(conditions)
}

// canSatisfyStatRequirements checks standard stat-based conditions
func (gs *GameState) canSatisfyStatRequirements(conditions map[string]map[string]float64) bool {
	for statName, constraints := range conditions {
		// Skip special relationship conditions
		if statName == "relationshipLevel" || statName == "interactionCount" || statName == "memoryCount" {
			continue
		}

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

// canSatisfyRelationshipRequirements checks relationship and memory-based conditions
func (gs *GameState) canSatisfyRelationshipRequirements(conditions map[string]map[string]float64) bool {
	// Check relationship level requirements
	if levelConditions, hasLevelConditions := conditions["relationshipLevel"]; hasLevelConditions {
		if !gs.checkRelationshipLevelConditions(levelConditions) {
			return false
		}
	}

	// Check interaction count requirements
	if interactionConditions, hasInteractionConditions := conditions["interactionCount"]; hasInteractionConditions {
		if !gs.checkInteractionCountConditions(interactionConditions) {
			return false
		}
	}

	// Check memory-based requirements
	if memoryConditions, hasMemoryConditions := conditions["memoryCount"]; hasMemoryConditions {
		if !gs.checkMemoryCountConditions(memoryConditions) {
			return false
		}
	}

	return true
}

// checkRelationshipLevelConditions validates relationship level requirements
func (gs *GameState) checkRelationshipLevelConditions(conditions map[string]float64) bool {
	currentLevel := gs.GetRelationshipLevel()

	// Map relationship levels to numeric values for comparison
	levelValues := map[string]float64{
		"Stranger":          0,
		"Friend":            1,
		"Close Friend":      2,
		"Romantic Interest": 3,
		"Partner":           4,
	}

	currentValue, exists := levelValues[currentLevel]
	if !exists {
		return false
	}

	// Check minimum level requirement
	if minLevel, hasMin := conditions["min"]; hasMin {
		if currentValue < minLevel {
			return false
		}
	}

	// Check maximum level requirement
	if maxLevel, hasMax := conditions["max"]; hasMax {
		if currentValue > maxLevel {
			return false
		}
	}

	return true
}

// checkInteractionCountConditions validates interaction count requirements
func (gs *GameState) checkInteractionCountConditions(conditions map[string]float64) bool {
	// Check total interaction count
	if minTotal, hasMinTotal := conditions["total_min"]; hasMinTotal {
		totalInteractions := 0
		for _, interactions := range gs.InteractionHistory {
			totalInteractions += len(interactions)
		}
		if float64(totalInteractions) < minTotal {
			return false
		}
	}

	// Check specific interaction type counts (e.g., compliment_min: 5)
	for conditionKey, requiredCount := range conditions {
		if strings.HasSuffix(conditionKey, "_min") {
			interactionType := strings.TrimSuffix(conditionKey, "_min")
			actualCount := gs.GetInteractionCount(interactionType)
			if float64(actualCount) < requiredCount {
				return false
			}
		}
	}

	return true
}

// checkMemoryCountConditions validates memory-based requirements
func (gs *GameState) checkMemoryCountConditions(conditions map[string]float64) bool {
	// Check total memory count
	if minMemories, hasMinMemories := conditions["total_min"]; hasMinMemories {
		if float64(len(gs.RomanceMemories)) < minMemories {
			return false
		}
	}

	// Check recent memory patterns (e.g., recent_positive_min: 2)
	if recentPositive, hasRecentPositive := conditions["recent_positive_min"]; hasRecentPositive {
		recentCount := gs.countRecentPositiveMemories(24 * time.Hour) // Last 24 hours
		if float64(recentCount) < recentPositive {
			return false
		}
	}

	return true
}

// countRecentPositiveMemories counts positive interactions in the recent time period
func (gs *GameState) countRecentPositiveMemories(duration time.Duration) int {
	now := time.Now()
	cutoff := now.Add(-duration)
	count := 0

	for _, memory := range gs.RomanceMemories {
		if memory.Timestamp.After(cutoff) {
			// Consider interactions that increased affection or trust as positive
			if affectionGain, hasAffection := memory.StatsAfter["affection"]; hasAffection {
				if affectionBefore, hadBefore := memory.StatsBefore["affection"]; hadBefore {
					if affectionGain > affectionBefore {
						count++
					}
				}
			}
		}
	}

	return count
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

// GetMoodCategory returns the mood category based on overall mood calculation
// Used for mood-based animation preferences
func (gs *GameState) GetMoodCategory() string {
	mood := gs.GetOverallMood()
	switch {
	case mood >= 80:
		return "happy"
	case mood >= 60:
		return "content"
	case mood >= 40:
		return "neutral"
	case mood >= 20:
		return "sad"
	default:
		return "depressed"
	}
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

// RecordRomanceInteraction records an interaction for romance memory system
func (gs *GameState) RecordRomanceInteraction(interactionType, response string, statsBefore, statsAfter map[string]float64) {
	if gs == nil {
		return
	}

	gs.mu.Lock()
	defer gs.mu.Unlock()

	// Initialize interaction history if needed
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

// GetInteractionCount returns the number of times an interaction has been performed
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

// UpdateRelationshipLevel checks and updates relationship level based on progression
func (gs *GameState) UpdateRelationshipLevel(progressionConfig *ProgressionConfig) bool {
	if gs == nil || progressionConfig == nil {
		return false
	}

	gs.mu.Lock()
	defer gs.mu.Unlock()

	// Get current age in seconds for level requirements
	var currentAge int64 = 0
	if gs.Progression != nil {
		currentAge = int64(gs.Progression.GetAge().Seconds())
	}

	// Find the highest level we can achieve
	var newLevel string = "Stranger"
	for _, level := range progressionConfig.Levels {
		// Check if we meet all requirements for this level
		if gs.meetsRelationshipRequirements(level.Requirement, currentAge) {
			newLevel = level.Name
		}
	}

	// Check if level changed
	oldLevel := gs.RelationshipLevel
	if oldLevel == "" {
		oldLevel = "Stranger"
	}

	if newLevel != oldLevel {
		gs.RelationshipLevel = newLevel
		return true
	}

	return false
}

// meetsRelationshipRequirements checks if current stats meet level requirements
func (gs *GameState) meetsRelationshipRequirements(requirements map[string]int64, currentAge int64) bool {
	for statName, threshold := range requirements {
		if statName == "age" {
			if currentAge < threshold {
				return false
			}
			continue
		}

		// Check romance stats
		if stat, exists := gs.Stats[statName]; exists {
			if stat.Current < float64(threshold) {
				return false
			}
		} else {
			// Required stat doesn't exist
			return false
		}
	}
	return true
}

// GetRomanceStats returns a copy of romance-related stats
func (gs *GameState) GetRomanceStats() map[string]float64 {
	if gs == nil {
		return nil
	}

	gs.mu.RLock()
	defer gs.mu.RUnlock()

	romanceStats := make(map[string]float64)
	romanceStatNames := []string{"affection", "trust", "intimacy", "jealousy"}

	for _, statName := range romanceStatNames {
		if stat, exists := gs.Stats[statName]; exists {
			romanceStats[statName] = stat.Current
		}
	}

	return romanceStats
}

// GetInteractionHistory returns a copy of interaction history
func (gs *GameState) GetInteractionHistory() map[string][]time.Time {
	if gs == nil || gs.InteractionHistory == nil {
		return nil
	}

	gs.mu.RLock()
	defer gs.mu.RUnlock()

	history := make(map[string][]time.Time)
	for interactionType, timestamps := range gs.InteractionHistory {
		// Create a copy of the slice
		history[interactionType] = make([]time.Time, len(timestamps))
		copy(history[interactionType], timestamps)
	}

	return history
}

// GetRomanceMemories returns a copy of romance memories
func (gs *GameState) GetRomanceMemories() []RomanceMemory {
	if gs == nil || gs.RomanceMemories == nil {
		return nil
	}

	gs.mu.RLock()
	defer gs.mu.RUnlock()

	memories := make([]RomanceMemory, len(gs.RomanceMemories))
	copy(memories, gs.RomanceMemories)
	return memories
}

// RecordDialogMemory records a dialog interaction for learning and adaptation
func (gs *GameState) RecordDialogMemory(memory DialogMemory) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	// Initialize if nil
	if gs.DialogMemories == nil {
		gs.DialogMemories = make([]DialogMemory, 0)
	}

	gs.DialogMemories = append(gs.DialogMemories, memory)

	// Limit memory storage to prevent excessive growth
	const maxDialogMemories = 100
	if len(gs.DialogMemories) > maxDialogMemories {
		// Remove oldest memories, keeping the most recent ones
		start := len(gs.DialogMemories) - maxDialogMemories
		gs.DialogMemories = gs.DialogMemories[start:]
	}
}

// GetDialogMemories returns a copy of dialog memories
func (gs *GameState) GetDialogMemories() []DialogMemory {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	if gs.DialogMemories == nil {
		return []DialogMemory{}
	}

	memories := make([]DialogMemory, len(gs.DialogMemories))
	copy(memories, gs.DialogMemories)
	return memories
}

// GetRecentDialogMemories returns recent dialog memories (last N)
func (gs *GameState) GetRecentDialogMemories(count int) []DialogMemory {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	if gs.DialogMemories == nil || len(gs.DialogMemories) == 0 {
		return []DialogMemory{}
	}

	start := len(gs.DialogMemories) - count
	if start < 0 {
		start = 0
	}

	memories := make([]DialogMemory, len(gs.DialogMemories[start:]))
	copy(memories, gs.DialogMemories[start:])
	return memories
}

// GetDialogMemoriesByTrigger returns dialog memories filtered by trigger type
func (gs *GameState) GetDialogMemoriesByTrigger(trigger string) []DialogMemory {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	if gs.DialogMemories == nil {
		return []DialogMemory{}
	}

	var filtered []DialogMemory
	for _, memory := range gs.DialogMemories {
		if memory.Trigger == trigger {
			filtered = append(filtered, memory)
		}
	}

	return filtered
}

// GetHighImportanceDialogMemories returns dialog memories with high importance scores
func (gs *GameState) GetHighImportanceDialogMemories(minImportance float64) []DialogMemory {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	if gs.DialogMemories == nil {
		return []DialogMemory{}
	}

	var important []DialogMemory
	for _, memory := range gs.DialogMemories {
		if memory.MemoryImportance >= minImportance {
			important = append(important, memory)
		}
	}

	return important
}

// MarkDialogResponseFavorite marks a dialog response as favorite and sets rating
func (gs *GameState) MarkDialogResponseFavorite(response string, rating float64) bool {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	if gs.DialogMemories == nil {
		return false
	}

	// Find the most recent matching response and mark as favorite
	for i := len(gs.DialogMemories) - 1; i >= 0; i-- {
		if gs.DialogMemories[i].Response == response {
			gs.DialogMemories[i].IsFavorite = true
			gs.DialogMemories[i].FavoriteRating = rating
			return true
		}
	}

	return false
}

// UnmarkDialogResponseFavorite removes favorite status from a dialog response
func (gs *GameState) UnmarkDialogResponseFavorite(response string) bool {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	if gs.DialogMemories == nil {
		return false
	}

	// Find the most recent matching response and unmark as favorite
	for i := len(gs.DialogMemories) - 1; i >= 0; i-- {
		if gs.DialogMemories[i].Response == response {
			gs.DialogMemories[i].IsFavorite = false
			gs.DialogMemories[i].FavoriteRating = 0
			return true
		}
	}

	return false
}

// GetFavoriteDialogResponses returns all dialog memories marked as favorites
func (gs *GameState) GetFavoriteDialogResponses() []DialogMemory {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	if gs.DialogMemories == nil {
		return []DialogMemory{}
	}

	var favorites []DialogMemory
	for _, memory := range gs.DialogMemories {
		if memory.IsFavorite {
			favorites = append(favorites, memory)
		}
	}

	return favorites
}

// IsDialogResponseFavorite checks if a specific response is marked as favorite
func (gs *GameState) IsDialogResponseFavorite(response string) (bool, float64) {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	if gs.DialogMemories == nil {
		return false, 0
	}

	// Find the most recent matching response
	for i := len(gs.DialogMemories) - 1; i >= 0; i-- {
		if gs.DialogMemories[i].Response == response {
			return gs.DialogMemories[i].IsFavorite, gs.DialogMemories[i].FavoriteRating
		}
	}

	return false, 0
}

// GetFavoriteResponsesByRating returns favorite responses filtered by minimum rating
func (gs *GameState) GetFavoriteResponsesByRating(minRating float64) []DialogMemory {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	if gs.DialogMemories == nil {
		return []DialogMemory{}
	}

	var favorites []DialogMemory
	for _, memory := range gs.DialogMemories {
		if memory.IsFavorite && memory.FavoriteRating >= minRating {
			favorites = append(favorites, memory)
		}
	}

	return favorites
}
