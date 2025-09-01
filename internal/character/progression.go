package character

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// ProgressionState manages character level progression and achievements
// Uses only Go standard library following "lazy programmer" principles
type ProgressionState struct {
	mu                  sync.RWMutex
	CurrentLevel        string              `json:"currentLevel"`
	Age                 time.Duration       `json:"age"`
	TotalCareTime       time.Duration       `json:"totalCareTime"`
	Achievements        []string            `json:"achievements"`
	InteractionCounts   map[string]int      `json:"interactionCounts"`
	LastLevelCheck      time.Time           `json:"lastLevelCheck"`
	AchievementProgress map[string]Progress `json:"achievementProgress"`
	Config              *ProgressionConfig  `json:"config,omitempty"`
}

// Progress tracks achievement progress over time
type Progress struct {
	StartTime    time.Time     `json:"startTime"`
	Duration     time.Duration `json:"duration"`
	RequiredTime time.Duration `json:"requiredTime"`
	MetCriteria  bool          `json:"metCriteria"`
}

// ProgressionConfig defines progression rules from character card JSON
type ProgressionConfig struct {
	Levels       []LevelConfig       `json:"levels"`
	Achievements []AchievementConfig `json:"achievements"`
}

// LevelConfig defines a character level with requirements and changes
type LevelConfig struct {
	Name        string            `json:"name"`
	Requirement map[string]int64  `json:"requirement"` // age in seconds, other criteria
	Size        int               `json:"size"`        // Character size at this level
	Animations  map[string]string `json:"animations"`  // Animation overrides for this level
}

// AchievementConfig defines an achievement with stat-based requirements
type AchievementConfig struct {
	Name        string                            `json:"name"`
	Requirement map[string]map[string]interface{} `json:"requirement"` // Complex stat requirements
	Reward      *AchievementReward                `json:"reward,omitempty"`
}

// AchievementReward defines what the character gets for achieving something
type AchievementReward struct {
	StatBoosts map[string]float64 `json:"statBoosts,omitempty"` // Permanent stat increases
	Animations map[string]string  `json:"animations,omitempty"` // Unlocked animations
	Size       int                `json:"size,omitempty"`       // Size change
}

// AchievementDetails contains detailed information about a newly earned achievement
// Used for displaying notifications with rich information about what was unlocked
type AchievementDetails struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Timestamp   time.Time          `json:"timestamp"`
	Reward      *AchievementReward `json:"reward,omitempty"`
}

// NewProgressionState creates a new progression state with configuration
func NewProgressionState(config *ProgressionConfig) *ProgressionState {
	ps := &ProgressionState{
		CurrentLevel:        "Baby", // Default starting level
		Age:                 0,
		TotalCareTime:       0,
		Achievements:        make([]string, 0),
		InteractionCounts:   make(map[string]int),
		LastLevelCheck:      time.Now(),
		AchievementProgress: make(map[string]Progress),
		Config:              config,
	}

	// Initialize achievement progress tracking
	if config != nil {
		for _, achievement := range config.Achievements {
			ps.AchievementProgress[achievement.Name] = Progress{
				StartTime:   time.Now(),
				MetCriteria: false,
			}
		}
	}

	return ps
}

// Update progression state based on current game state and elapsed time
// Returns true if level changed, list of newly earned achievement details
func (ps *ProgressionState) Update(gameState *GameState, elapsed time.Duration) (bool, []AchievementDetails) {
	if ps == nil || ps.Config == nil {
		return false, nil
	}

	ps.mu.Lock()
	defer ps.mu.Unlock()

	ps.Age += elapsed
	ps.TotalCareTime += elapsed

	levelChanged := ps.checkLevelProgression()
	newAchievementNames := ps.checkAchievements(gameState, elapsed)

	// Convert achievement names to detailed achievement information
	newAchievementDetails := make([]AchievementDetails, 0, len(newAchievementNames))
	for _, achievementName := range newAchievementNames {
		details := ps.createAchievementDetails(achievementName)
		newAchievementDetails = append(newAchievementDetails, details)
	}

	ps.LastLevelCheck = time.Now()

	return levelChanged, newAchievementDetails
}

// createAchievementDetails creates detailed achievement information from achievement name
func (ps *ProgressionState) createAchievementDetails(achievementName string) AchievementDetails {
	// Find the achievement config to get description and reward info
	var achievementConfig *AchievementConfig
	for _, config := range ps.Config.Achievements {
		if config.Name == achievementName {
			achievementConfig = &config
			break
		}
	}

	details := AchievementDetails{
		Name:      achievementName,
		Timestamp: time.Now(),
	}

	// Add description and reward if available
	if achievementConfig != nil {
		// For now, use the name as description. In a full implementation,
		// the AchievementConfig could include a separate description field
		details.Description = generateAchievementDescription(achievementConfig)
		details.Reward = achievementConfig.Reward
	}

	return details
}

// generateAchievementDescription creates a user-friendly description from achievement config
func generateAchievementDescription(config *AchievementConfig) string {
	if config == nil {
		return "Achievement unlocked!"
	}

	// Create a basic description based on requirements
	// This follows the "lazy programmer" principle - simple but functional
	switch {
	case len(config.Requirement) == 0:
		return "Achievement unlocked!"
	case config.Requirement["happiness"] != nil:
		return "Maintained excellent happiness level!"
	case config.Requirement["hunger"] != nil:
		return "Kept your companion well-fed!"
	case config.Requirement["health"] != nil:
		return "Maintained excellent health!"
	case config.Requirement["energy"] != nil:
		return "Kept your companion energized!"
	default:
		return "Special achievement unlocked!"
	}
}

// checkLevelProgression evaluates if character should advance to next level
func (ps *ProgressionState) checkLevelProgression() bool {
	if ps.Config == nil {
		return false
	}

	currentLevelIndex := ps.getCurrentLevelIndex()
	ageSeconds := int64(ps.Age.Seconds())

	// Check each level after current level
	for i := currentLevelIndex + 1; i < len(ps.Config.Levels); i++ {
		level := ps.Config.Levels[i]

		// Check age requirement
		if ageReq, hasAge := level.Requirement["age"]; hasAge {
			if ageSeconds >= ageReq {
				ps.CurrentLevel = level.Name
				return true
			}
		}
	}

	return false
}

// getCurrentLevelIndex finds the index of current level in config
func (ps *ProgressionState) getCurrentLevelIndex() int {
	if ps.Config == nil {
		return 0
	}

	for i, level := range ps.Config.Levels {
		if level.Name == ps.CurrentLevel {
			return i
		}
	}

	return 0 // Default to first level if current not found
}

// checkAchievements evaluates achievement progress and completion
func (ps *ProgressionState) checkAchievements(gameState *GameState, elapsed time.Duration) []string {
	if ps.Config == nil || gameState == nil {
		return nil
	}

	newAchievements := make([]string, 0)

	for _, achievement := range ps.Config.Achievements {
		if ps.hasAchievement(achievement.Name) {
			continue
		}

		progress := ps.AchievementProgress[achievement.Name]
		metCriteria := ps.evaluateAchievementRequirement(achievement.Requirement, gameState)

		if ps.shouldStartProgress(metCriteria, progress) {
			progress = ps.initializeAchievementProgress(achievement, progress)
			if earnedImmediately := ps.processInstantAchievement(achievement.Name, achievement.Reward, gameState, progress); earnedImmediately {
				newAchievements = append(newAchievements, achievement.Name)
			}
		} else if ps.shouldResetProgress(metCriteria, progress) {
			progress = ps.resetAchievementProgress(progress)
		}

		var earnedAfterDuration bool
		progress, earnedAfterDuration = ps.processDurationAchievement(achievement.Name, achievement.Reward, gameState, progress, elapsed)
		if earnedAfterDuration {
			newAchievements = append(newAchievements, achievement.Name)
		}

		ps.AchievementProgress[achievement.Name] = progress
	}

	return newAchievements
}

// shouldStartProgress determines if achievement progress should begin
func (ps *ProgressionState) shouldStartProgress(metCriteria bool, progress Progress) bool {
	return metCriteria && !progress.MetCriteria
}

// shouldResetProgress determines if achievement progress should be reset
func (ps *ProgressionState) shouldResetProgress(metCriteria bool, progress Progress) bool {
	return !metCriteria && progress.MetCriteria
}

// initializeAchievementProgress sets up progress tracking for an achievement
func (ps *ProgressionState) initializeAchievementProgress(achievement AchievementConfig, progress Progress) Progress {
	progress.StartTime = time.Now()
	progress.MetCriteria = true
	progress.Duration = 0
	progress.RequiredTime = ps.extractRequiredDuration(achievement.Requirement)
	return progress
}

// extractRequiredDuration gets the required duration from achievement requirements
func (ps *ProgressionState) extractRequiredDuration(requirement map[string]map[string]interface{}) time.Duration {
	if maintainAbove, hasMA := requirement["maintainAbove"]; hasMA {
		if duration, hasDuration := maintainAbove["duration"].(float64); hasDuration {
			return time.Duration(duration) * time.Second
		}
	}
	return 0
}

// processInstantAchievement handles achievements that are earned immediately
func (ps *ProgressionState) processInstantAchievement(name string, reward *AchievementReward, gameState *GameState, progress Progress) bool {
	if progress.RequiredTime == 0 {
		ps.Achievements = append(ps.Achievements, name)
		ps.applyAchievementReward(reward, gameState)
		return true
	}
	return false
}

// processDurationAchievement handles achievements that require sustained criteria
func (ps *ProgressionState) processDurationAchievement(name string, reward *AchievementReward, gameState *GameState, progress Progress, elapsed time.Duration) (Progress, bool) {
	if progress.MetCriteria && progress.RequiredTime > 0 {
		progress.Duration += elapsed
		if progress.Duration >= progress.RequiredTime {
			ps.Achievements = append(ps.Achievements, name)
			ps.applyAchievementReward(reward, gameState)
			return progress, true
		}
	}
	return progress, false
}

// resetAchievementProgress resets progress when criteria are no longer met
func (ps *ProgressionState) resetAchievementProgress(progress Progress) Progress {
	progress.MetCriteria = false
	progress.Duration = 0
	return progress
} // evaluateAchievementRequirement checks if current game state meets achievement criteria
func (ps *ProgressionState) evaluateAchievementRequirement(requirement map[string]map[string]interface{}, gameState *GameState) bool {
	for statName, criteria := range requirement {
		if ps.shouldSkipStatRequirement(statName) {
			continue
		}

		currentValue := gameState.GetStat(statName)
		if !ps.validateStatCriteria(criteria, currentValue) {
			return false
		}
	}

	return true
}

// shouldSkipStatRequirement determines if a stat requirement should be skipped during evaluation
func (ps *ProgressionState) shouldSkipStatRequirement(statName string) bool {
	return statName == "maintainAbove"
}

// validateStatCriteria checks if current value meets all criteria for a specific stat
func (ps *ProgressionState) validateStatCriteria(criteria map[string]interface{}, currentValue float64) bool {
	if !ps.checkMaintainAboveRequirement(criteria, currentValue) {
		return false
	}

	if !ps.checkMinimumValueRequirement(criteria, currentValue) {
		return false
	}

	if !ps.checkMaximumValueRequirement(criteria, currentValue) {
		return false
	}

	return true
}

// checkMaintainAboveRequirement validates maintainAbove threshold requirements
func (ps *ProgressionState) checkMaintainAboveRequirement(criteria map[string]interface{}, currentValue float64) bool {
	maintainAbove, hasMA := criteria["maintainAbove"]
	if !hasMA {
		return true
	}

	threshold, ok := maintainAbove.(float64)
	if !ok {
		return true
	}

	return currentValue >= threshold
}

// checkMinimumValueRequirement validates minimum value requirements
func (ps *ProgressionState) checkMinimumValueRequirement(criteria map[string]interface{}, currentValue float64) bool {
	minVal, hasMin := criteria["min"]
	if !hasMin {
		return true
	}

	threshold, ok := minVal.(float64)
	if !ok {
		return true
	}

	return currentValue >= threshold
}

// checkMaximumValueRequirement validates maximum value requirements
func (ps *ProgressionState) checkMaximumValueRequirement(criteria map[string]interface{}, currentValue float64) bool {
	maxVal, hasMax := criteria["max"]
	if !hasMax {
		return true
	}

	threshold, ok := maxVal.(float64)
	if !ok {
		return true
	}

	return currentValue <= threshold
}

// applyAchievementReward applies rewards from completing an achievement
func (ps *ProgressionState) applyAchievementReward(reward *AchievementReward, gameState *GameState) {
	if reward == nil {
		return
	}

	// Apply permanent stat boosts
	if len(reward.StatBoosts) > 0 {
		for statName, boost := range reward.StatBoosts {
			if stat, exists := gameState.Stats[statName]; exists {
				// Increase the maximum value of the stat
				stat.Max += boost
				// Also increase current value to match
				stat.Current += boost
			}
		}
	}

	// Note: Animation and size rewards would be applied at the UI level
	// This keeps the progression system focused on data management
}

// hasAchievement checks if an achievement has already been earned
func (ps *ProgressionState) hasAchievement(name string) bool {
	for _, achievement := range ps.Achievements {
		if achievement == name {
			return true
		}
	}
	return false
}

// RecordInteraction increments the count for a specific interaction type
func (ps *ProgressionState) RecordInteraction(interactionType string) {
	if ps == nil {
		return
	}

	ps.mu.Lock()
	defer ps.mu.Unlock()

	ps.InteractionCounts[interactionType]++
}

// GetCurrentLevel returns the current level configuration
func (ps *ProgressionState) GetCurrentLevel() *LevelConfig {
	if ps == nil || ps.Config == nil {
		return nil
	}

	ps.mu.RLock()
	defer ps.mu.RUnlock()

	for _, level := range ps.Config.Levels {
		if level.Name == ps.CurrentLevel {
			return &level
		}
	}

	return nil
}

// GetAge returns the character's current age
func (ps *ProgressionState) GetAge() time.Duration {
	if ps == nil {
		return 0
	}

	ps.mu.RLock()
	defer ps.mu.RUnlock()

	return ps.Age
}

// GetAchievements returns a copy of earned achievements
func (ps *ProgressionState) GetAchievements() []string {
	if ps == nil {
		return nil
	}

	ps.mu.RLock()
	defer ps.mu.RUnlock()

	achievements := make([]string, len(ps.Achievements))
	copy(achievements, ps.Achievements)
	return achievements
}

// GetInteractionCounts returns a copy of interaction counts
func (ps *ProgressionState) GetInteractionCounts() map[string]int {
	if ps == nil {
		return nil
	}

	ps.mu.RLock()
	defer ps.mu.RUnlock()

	counts := make(map[string]int)
	for k, v := range ps.InteractionCounts {
		counts[k] = v
	}
	return counts
}

// GetCurrentSize returns the character size for the current level
func (ps *ProgressionState) GetCurrentSize() int {
	currentLevel := ps.GetCurrentLevel()
	if currentLevel == nil || currentLevel.Size == 0 {
		return 128 // Default size
	}
	return currentLevel.Size
}

// GetLevelAnimation returns level-specific animation if available
func (ps *ProgressionState) GetLevelAnimation(animationName string) (string, bool) {
	currentLevel := ps.GetCurrentLevel()
	if currentLevel == nil || len(currentLevel.Animations) == 0 {
		return "", false
	}

	ps.mu.RLock()
	defer ps.mu.RUnlock()

	animation, exists := currentLevel.Animations[animationName]
	return animation, exists
}

// MarshalJSON implements custom JSON marshaling for save files
func (ps *ProgressionState) MarshalJSON() ([]byte, error) {
	if ps == nil {
		return []byte("null"), nil
	}

	ps.mu.RLock()
	defer ps.mu.RUnlock()

	// Create a serializable version of the struct
	type Alias ProgressionState
	return json.Marshal(&struct {
		*Alias
		AgeNanos           int64 `json:"ageNanos"`
		TotalCareTimeNanos int64 `json:"totalCareTimeNanos"`
	}{
		Alias:              (*Alias)(ps),
		AgeNanos:           int64(ps.Age),
		TotalCareTimeNanos: int64(ps.TotalCareTime),
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for save files
func (ps *ProgressionState) UnmarshalJSON(data []byte) error {
	type Alias ProgressionState
	aux := &struct {
		*Alias
		AgeNanos           int64 `json:"ageNanos"`
		TotalCareTimeNanos int64 `json:"totalCareTimeNanos"`
	}{
		Alias: (*Alias)(ps),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	ps.Age = time.Duration(aux.AgeNanos)
	ps.TotalCareTime = time.Duration(aux.TotalCareTimeNanos)
	return nil
}

// Validate ensures the progression state has consistent data
func (ps *ProgressionState) Validate() error {
	if ps == nil {
		return fmt.Errorf("progression state is nil")
	}

	if ps.Age < 0 {
		return fmt.Errorf("age cannot be negative")
	}

	if ps.TotalCareTime < 0 {
		return fmt.Errorf("total care time cannot be negative")
	}

	if ps.Config != nil {
		if err := ps.validateConfig(); err != nil {
			return fmt.Errorf("invalid progression config: %w", err)
		}
	}

	return nil
}

// validateConfig validates the progression configuration
func (ps *ProgressionState) validateConfig() error {
	if len(ps.Config.Levels) == 0 {
		return fmt.Errorf("must have at least one level")
	}

	// Validate levels
	for i, level := range ps.Config.Levels {
		if err := ps.validateLevel(level, i); err != nil {
			return fmt.Errorf("level %d (%s): %w", i, level.Name, err)
		}
	}

	// Validate achievements
	for i, achievement := range ps.Config.Achievements {
		if err := ps.validateAchievement(achievement, i); err != nil {
			return fmt.Errorf("achievement %d (%s): %w", i, achievement.Name, err)
		}
	}

	return nil
}

// validateLevel validates a single level configuration
func (ps *ProgressionState) validateLevel(level LevelConfig, index int) error {
	if len(level.Name) == 0 {
		return fmt.Errorf("name cannot be empty")
	}

	if level.Size < 32 || level.Size > 1024 {
		return fmt.Errorf("size must be 32-1024 pixels, got %d", level.Size)
	}

	// First level should have age requirement of 0
	if index == 0 {
		if ageReq, hasAge := level.Requirement["age"]; hasAge && ageReq != 0 {
			return fmt.Errorf("first level must have age requirement of 0, got %d", ageReq)
		}
	}

	return nil
}

// validateAchievement validates a single achievement configuration
func (ps *ProgressionState) validateAchievement(achievement AchievementConfig, index int) error {
	if len(achievement.Name) == 0 {
		return fmt.Errorf("name cannot be empty")
	}

	if len(achievement.Requirement) == 0 {
		return fmt.Errorf("must have at least one requirement")
	}

	return nil
}
