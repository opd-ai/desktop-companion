package character

import (
	"desktop-companion/internal/dialog"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CharacterCard represents the JSON configuration for a desktop companion character
// This follows the "lazy programmer" approach - leveraging Go's built-in JSON package
// instead of writing custom parsers
type CharacterCard struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Animations  map[string]string `json:"animations"`
	Dialogs     []Dialog          `json:"dialogs"`
	Behavior    Behavior          `json:"behavior"`
	// Game feature extensions (Phase 1 implementation)
	Stats        map[string]StatConfig        `json:"stats,omitempty"`
	GameRules    *GameRulesConfig             `json:"gameRules,omitempty"`
	Interactions map[string]InteractionConfig `json:"interactions,omitempty"`
	// Progression features (Phase 3 implementation)
	Progression *ProgressionConfig `json:"progression,omitempty"`
	// Random events (Phase 3 implementation)
	RandomEvents []RandomEventConfig `json:"randomEvents,omitempty"`
	// Romance feature extensions (Dating Simulator Phase 1)
	Personality    *PersonalityConfig  `json:"personality,omitempty"`
	RomanceDialogs []DialogExtended    `json:"romanceDialogs,omitempty"`
	RomanceEvents  []RandomEventConfig `json:"romanceEvents,omitempty"`
	// Advanced dialog system (Phase 1)
	DialogBackend *dialog.DialogBackendConfig `json:"dialogBackend,omitempty"`
	// General dialog events system (Phase 4)
	GeneralEvents []GeneralDialogEvent `json:"generalEvents,omitempty"`
	// Gift system (optional feature - maintains backward compatibility)
	GiftSystem *GiftSystemConfig `json:"giftSystem,omitempty"`
}

// Dialog represents an interaction trigger and response configuration
type Dialog struct {
	Trigger   string   `json:"trigger"`   // "click", "rightclick", "hover"
	Responses []string `json:"responses"` // 1-10 response strings
	Animation string   `json:"animation"` // Must match an animation key
	Cooldown  int      `json:"cooldown"`  // Seconds between triggers (default: 5)
}

// Behavior defines character behavior settings
type Behavior struct {
	IdleTimeout     int  `json:"idleTimeout"`     // Seconds before returning to idle
	MovementEnabled bool `json:"movementEnabled"` // Allow dragging
	DefaultSize     int  `json:"defaultSize"`     // Character size in pixels
}

// GameRulesConfig defines game-wide settings for Tamagotchi-style features
type GameRulesConfig struct {
	StatsDecayInterval             int  `json:"statsDecayInterval"`             // Seconds between stat degradation
	AutoSaveInterval               int  `json:"autoSaveInterval"`               // Seconds between auto-saves
	CriticalStateAnimationPriority bool `json:"criticalStateAnimationPriority"` // Priority for critical animations
	DeathEnabled                   bool `json:"deathEnabled"`                   // Whether character can die
	EvolutionEnabled               bool `json:"evolutionEnabled"`               // Whether character evolves
	MoodBasedAnimations            bool `json:"moodBasedAnimations"`            // Use mood for animation selection
}

// InteractionConfig defines a game interaction (feed, play, etc.)
type InteractionConfig struct {
	Triggers     []string                      `json:"triggers"`     // Input triggers (rightclick, doubleclick, etc.)
	Effects      map[string]float64            `json:"effects"`      // Stat changes to apply
	Animations   []string                      `json:"animations"`   // Animations to play
	Responses    []string                      `json:"responses"`    // Dialog responses
	Cooldown     int                           `json:"cooldown"`     // Seconds between uses
	Duration     int                           `json:"duration"`     // Duration of effect (for sleep, etc.)
	Requirements map[string]map[string]float64 `json:"requirements"` // Stat requirements to use interaction
}

// RandomEventConfig defines a random event that can affect character stats
// Events are triggered based on probability and conditions, following "lazy programmer" approach
type RandomEventConfig struct {
	Name        string                        `json:"name"`        // Event name for identification
	Description string                        `json:"description"` // Human-readable description
	Probability float64                       `json:"probability"` // 0.0-1.0 chance of triggering per check
	Effects     map[string]float64            `json:"effects"`     // Stat changes to apply when triggered
	Animations  []string                      `json:"animations"`  // Animations to play when triggered
	Responses   []string                      `json:"responses"`   // Dialog responses to show
	Cooldown    int                           `json:"cooldown"`    // Minimum seconds between triggers
	Duration    int                           `json:"duration"`    // Duration in seconds (0 = instant)
	Conditions  map[string]map[string]float64 `json:"conditions"`  // Stat conditions required to trigger
}

// Romance-specific configuration structures (Dating Simulator Phase 1)
// PersonalityConfig defines character personality traits that affect romance interactions
type PersonalityConfig struct {
	Traits        map[string]float64 `json:"traits"`        // Personality traits (0.0-1.0 values)
	Compatibility map[string]float64 `json:"compatibility"` // Behavior compatibility modifiers
}

// RomanceRequirement defines complex requirements for romance features
type RomanceRequirement struct {
	Stats               map[string]map[string]float64 `json:"stats,omitempty"`               // Stat-based requirements
	RelationshipLevel   string                        `json:"relationshipLevel,omitempty"`   // Required relationship level
	InteractionCount    map[string]map[string]int     `json:"interactionCount,omitempty"`    // Interaction count requirements
	AchievementUnlocked []string                      `json:"achievementUnlocked,omitempty"` // Required achievements
}

// DialogExtended extends the basic Dialog with romance-specific features
type DialogExtended struct {
	Dialog                           // Embed existing Dialog struct
	Requirements *RomanceRequirement `json:"requirements,omitempty"` // Romance-specific requirements
	RomanceLevel string              `json:"romanceLevel,omitempty"` // Associated romance level
}

// InteractionConfigExtended extends basic InteractionConfig with romance features
type InteractionConfigExtended struct {
	InteractionConfig                      // Embed existing InteractionConfig struct
	UnlockRequirements *RomanceRequirement `json:"unlockRequirements,omitempty"` // Romance unlock requirements
	RomanceCategory    string              `json:"romanceCategory,omitempty"`    // Romance interaction category
}

// LoadCard loads and validates a character card from a JSON file
// Uses standard library encoding/json - no external dependencies needed
func LoadCard(path string) (*CharacterCard, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read character card %s: %w", path, err)
	}

	var card CharacterCard
	if err := json.Unmarshal(data, &card); err != nil {
		return nil, fmt.Errorf("failed to parse character card %s: %w", path, err)
	}

	// Get character directory for animation file validation
	characterDir := filepath.Dir(path)
	if err := card.ValidateWithBasePath(characterDir); err != nil {
		return nil, fmt.Errorf("invalid character card %s: %w", path, err)
	}

	return &card, nil
}

// Validate ensures the character card has valid configuration
// Implements comprehensive validation to prevent runtime errors
// Validate ensures the character card has valid configuration
func (c *CharacterCard) Validate() error {
	if err := c.validateCoreElements(); err != nil {
		return err
	}

	if err := c.validateGameSystems(); err != nil {
		return err
	}

	if err := c.validateRomanceSystems(); err != nil {
		return err
	}

	return nil
}

// validateCoreElements validates basic character configuration and assets
func (c *CharacterCard) validateCoreElements() error {
	if err := c.validateBasicFields(); err != nil {
		return err
	}

	if err := c.validateAnimations(); err != nil {
		return err
	}

	if err := c.validateDialogs(); err != nil {
		return err
	}

	if err := c.Behavior.Validate(); err != nil {
		return fmt.Errorf("behavior: %w", err)
	}

	return nil
}

// validateGameSystems validates game mechanics and progression systems
func (c *CharacterCard) validateGameSystems() error {
	if err := c.validateGameFeatures(); err != nil {
		return fmt.Errorf("game features: %w", err)
	}

	if err := c.validateProgression(); err != nil {
		return fmt.Errorf("progression: %w", err)
	}

	if err := c.validateRandomEvents(); err != nil {
		return fmt.Errorf("random events: %w", err)
	}

	return nil
}

// validateRomanceSystems validates romance features and dialog backend
func (c *CharacterCard) validateRomanceSystems() error {
	if err := c.validateRomanceFeatures(); err != nil {
		return fmt.Errorf("romance features: %w", err)
	}

	if err := c.validateDialogBackend(); err != nil {
		return fmt.Errorf("dialog backend: %w", err)
	}

	if err := c.validateGeneralEvents(); err != nil {
		return fmt.Errorf("general events: %w", err)
	}

	return nil
}

// ValidateWithBasePath ensures the character card has valid configuration including file existence checks
func (c *CharacterCard) ValidateWithBasePath(basePath string) error {
	if err := c.validateCoreFields(basePath); err != nil {
		return err
	}

	return c.validateFeatureSections()
}

// validateCoreFields validates essential card fields and animations with file system checks
func (c *CharacterCard) validateCoreFields(basePath string) error {
	if err := c.validateBasicFields(); err != nil {
		return err
	}

	if err := c.validateAnimationsWithBasePath(basePath); err != nil {
		return err
	}

	if err := c.validateDialogs(); err != nil {
		return err
	}

	if err := c.Behavior.Validate(); err != nil {
		return fmt.Errorf("behavior: %w", err)
	}

	return nil
}

// validateFeatureSections validates all optional feature configurations
func (c *CharacterCard) validateFeatureSections() error {
	if err := c.validateGameFeatures(); err != nil {
		return fmt.Errorf("game features: %w", err)
	}

	if err := c.validateProgression(); err != nil {
		return fmt.Errorf("progression: %w", err)
	}

	if err := c.validateRandomEvents(); err != nil {
		return fmt.Errorf("random events: %w", err)
	}

	if err := c.validateRomanceFeatures(); err != nil {
		return fmt.Errorf("romance features: %w", err)
	}

	if err := c.validateDialogBackend(); err != nil {
		return fmt.Errorf("dialog backend: %w", err)
	}

	if err := c.validateGiftSystem(); err != nil {
		return fmt.Errorf("gift system: %w", err)
	}

	return nil
}

// validateBasicFields checks name and description field constraints
func (c *CharacterCard) validateBasicFields() error {
	if len(c.Name) == 0 || len(c.Name) > 50 {
		return fmt.Errorf("name must be 1-50 characters, got %d", len(c.Name))
	}

	if len(c.Description) == 0 || len(c.Description) > 200 {
		return fmt.Errorf("description must be 1-200 characters, got %d", len(c.Description))
	}

	return nil
}

// validateAnimations ensures required animations exist and have valid file paths
func (c *CharacterCard) validateAnimations() error {
	if c.Animations == nil {
		return fmt.Errorf("animations map is required")
	}

	if err := c.validateRequiredAnimations(); err != nil {
		return err
	}

	return c.validateAnimationPaths()
}

// validateAnimationsWithBasePath ensures required animations exist and all animation files are accessible
func (c *CharacterCard) validateAnimationsWithBasePath(basePath string) error {
	if c.Animations == nil {
		return fmt.Errorf("animations map is required")
	}

	if err := c.validateRequiredAnimations(); err != nil {
		return err
	}

	return c.validateAnimationPathsWithBasePath(basePath)
}

// validateRequiredAnimations checks that mandatory animation keys are present
func (c *CharacterCard) validateRequiredAnimations() error {
	requiredAnimations := []string{"idle", "talking"}
	for _, required := range requiredAnimations {
		if _, exists := c.Animations[required]; !exists {
			return fmt.Errorf("required animation '%s' not found", required)
		}
	}
	return nil
}

// validateAnimationPaths ensures all animation files have proper GIF extensions
func (c *CharacterCard) validateAnimationPaths() error {
	for name, path := range c.Animations {
		if !strings.HasSuffix(strings.ToLower(path), ".gif") {
			return fmt.Errorf("animation '%s' must be a GIF file, got: %s", name, path)
		}
	}
	return nil
}

// validateAnimationPathsWithBasePath ensures all animation files exist and are accessible
func (c *CharacterCard) validateAnimationPathsWithBasePath(basePath string) error {
	for name, path := range c.Animations {
		if !strings.HasSuffix(strings.ToLower(path), ".gif") {
			return fmt.Errorf("animation '%s' must be a GIF file, got: %s", name, path)
		}

		// Check if the animation file actually exists and is readable
		fullPath := filepath.Join(basePath, path)
		if _, err := os.Stat(fullPath); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("animation file '%s' not found: %s", name, fullPath)
			}
			return fmt.Errorf("animation file '%s' not accessible: %s (%v)", name, fullPath, err)
		}
	}
	return nil
}

// validateDialogs ensures dialog configurations are valid and reference existing animations
func (c *CharacterCard) validateDialogs() error {
	if len(c.Dialogs) == 0 {
		return fmt.Errorf("at least one dialog configuration is required")
	}

	for i, dialog := range c.Dialogs {
		if err := dialog.Validate(c.Animations); err != nil {
			return fmt.Errorf("dialog %d: %w", i, err)
		}
	}

	return nil
}

// Validate ensures dialog configuration is valid
func (d *Dialog) Validate(animations map[string]string) error {
	if err := d.validateTrigger(); err != nil {
		return err
	}

	if err := d.validateResponses(); err != nil {
		return err
	}

	if err := d.validateAnimationReference(animations); err != nil {
		return err
	}

	d.setDefaultCooldown()
	return nil
}

// validateTrigger checks if the trigger type is valid
func (d *Dialog) validateTrigger() error {
	validTriggers := []string{"click", "rightclick", "hover"}
	for _, trigger := range validTriggers {
		if d.Trigger == trigger {
			return nil
		}
	}
	return fmt.Errorf("trigger must be one of %v, got: %s", validTriggers, d.Trigger)
}

// validateResponses ensures responses are within limits and not empty
func (d *Dialog) validateResponses() error {
	if len(d.Responses) == 0 || len(d.Responses) > 10 {
		return fmt.Errorf("must have 1-10 responses, got %d", len(d.Responses))
	}

	for i, response := range d.Responses {
		if len(strings.TrimSpace(response)) == 0 {
			return fmt.Errorf("response %d cannot be empty", i)
		}
	}

	return nil
}

// validateAnimationReference checks if the referenced animation exists
func (d *Dialog) validateAnimationReference(animations map[string]string) error {
	if _, exists := animations[d.Animation]; !exists {
		return fmt.Errorf("animation '%s' not found in animations map", d.Animation)
	}
	return nil
}

// setDefaultCooldown applies default cooldown value if not specified
func (d *Dialog) setDefaultCooldown() {
	if d.Cooldown <= 0 {
		d.Cooldown = 5 // 5 second default
	}
}

// Validate ensures behavior settings are within acceptable ranges
func (b *Behavior) Validate() error {
	if b.IdleTimeout < 10 || b.IdleTimeout > 300 {
		return fmt.Errorf("idleTimeout must be 10-300 seconds, got %d", b.IdleTimeout)
	}

	if b.DefaultSize < 64 || b.DefaultSize > 512 {
		return fmt.Errorf("defaultSize must be 64-512 pixels, got %d", b.DefaultSize)
	}

	return nil
}

// GetAnimationPath returns the full path to an animation file
// Resolves relative paths from the character card directory
func (c *CharacterCard) GetAnimationPath(basePath, animationName string) (string, error) {
	animationFile, exists := c.Animations[animationName]
	if !exists {
		return "", fmt.Errorf("animation '%s' not found", animationName)
	}

	fullPath := filepath.Join(basePath, animationFile)

	// Verify file exists
	if _, err := os.Stat(fullPath); err != nil {
		return "", fmt.Errorf("animation file not found: %s", fullPath)
	}

	return fullPath, nil
}

// validateGameFeatures validates game-specific configuration fields
// This ensures game stats, rules, and interactions are properly configured
func (c *CharacterCard) validateGameFeatures() error {
	// Stats validation (optional)
	if c.Stats != nil {
		for name, stat := range c.Stats {
			if err := c.validateStatConfig(name, stat); err != nil {
				return fmt.Errorf("stat '%s': %w", name, err)
			}
		}
	}

	// Game rules validation (optional)
	if c.GameRules != nil {
		if err := c.validateGameRules(); err != nil {
			return fmt.Errorf("game rules: %w", err)
		}
	}

	// Interactions validation (optional)
	if c.Interactions != nil {
		for name, interaction := range c.Interactions {
			if err := c.validateInteractionConfig(name, interaction); err != nil {
				return fmt.Errorf("interaction '%s': %w", name, err)
			}
		}
	}

	return nil
}

// validateStatConfig ensures a stat configuration is valid
func (c *CharacterCard) validateStatConfig(name string, stat StatConfig) error {
	if stat.Max <= 0 {
		return fmt.Errorf("max value must be positive, got %f", stat.Max)
	}

	if stat.Initial < 0 || stat.Initial > stat.Max {
		return fmt.Errorf("initial value (%f) must be between 0 and max (%f)", stat.Initial, stat.Max)
	}

	if stat.DegradationRate < 0 {
		return fmt.Errorf("degradation rate cannot be negative, got %f", stat.DegradationRate)
	}

	if stat.CriticalThreshold < 0 || stat.CriticalThreshold > stat.Max {
		return fmt.Errorf("critical threshold (%f) must be between 0 and max (%f)", stat.CriticalThreshold, stat.Max)
	}

	return nil
}

// validateGameRules ensures game rules configuration is valid
func (c *CharacterCard) validateGameRules() error {
	if c.GameRules.StatsDecayInterval < 10 || c.GameRules.StatsDecayInterval > 3600 {
		return fmt.Errorf("stats decay interval must be 10-3600 seconds, got %d", c.GameRules.StatsDecayInterval)
	}

	if c.GameRules.AutoSaveInterval < 60 || c.GameRules.AutoSaveInterval > 7200 {
		return fmt.Errorf("auto save interval must be 60-7200 seconds, got %d", c.GameRules.AutoSaveInterval)
	}

	return nil
}

// validateInteractionConfig ensures an interaction configuration is valid
func (c *CharacterCard) validateInteractionConfig(name string, interaction InteractionConfig) error {
	if err := c.validateInteractionTriggers(interaction.Triggers); err != nil {
		return err
	}

	if err := c.validateInteractionAnimations(interaction.Animations); err != nil {
		return err
	}

	if err := c.validateInteractionResponses(interaction.Responses); err != nil {
		return err
	}

	if err := c.validateInteractionCooldown(interaction.Cooldown, interaction.Triggers); err != nil {
		return err
	}

	return nil
}

// validateInteractionTriggers validates that interaction triggers are valid and non-empty
func (c *CharacterCard) validateInteractionTriggers(triggers []string) error {
	if len(triggers) == 0 {
		return fmt.Errorf("must have at least one trigger")
	}

	validTriggers := []string{
		"click", "rightclick", "doubleclick", "shift+click", "hover",
		"ctrl+shift+click", "alt+shift+click", "daily_interaction_bonus",
	}

	for _, trigger := range triggers {
		if !c.isValidTrigger(trigger, validTriggers) {
			return fmt.Errorf("invalid trigger '%s', must be one of %v", trigger, validTriggers)
		}
	}

	return nil
}

// validateInteractionAnimations validates that all referenced animations exist in the animations map
func (c *CharacterCard) validateInteractionAnimations(animations []string) error {
	for _, animation := range animations {
		if _, exists := c.Animations[animation]; !exists {
			return fmt.Errorf("animation '%s' not found in animations map", animation)
		}
	}
	return nil
}

// validateInteractionResponses validates that responses count is within acceptable limits
func (c *CharacterCard) validateInteractionResponses(responses []string) error {
	if len(responses) == 0 || len(responses) > 10 {
		return fmt.Errorf("must have 1-10 responses, got %d", len(responses))
	}
	return nil
}

// validateInteractionCooldown validates cooldown duration based on trigger types
func (c *CharacterCard) validateInteractionCooldown(cooldown int, triggers []string) error {
	maxCooldown := c.calculateMaxCooldown(triggers)

	if cooldown < 0 || cooldown > maxCooldown {
		return fmt.Errorf("cooldown must be 0-%d seconds, got %d", maxCooldown, cooldown)
	}

	return nil
}

// calculateMaxCooldown determines the maximum allowed cooldown based on trigger types
func (c *CharacterCard) calculateMaxCooldown(triggers []string) int {
	maxCooldown := 3600 // Default 1 hour for most interactions

	for _, trigger := range triggers {
		if trigger == "daily_interaction_bonus" {
			return 86400 // Allow 24 hours for daily interactions
		}
	}

	return maxCooldown
}

// isValidTrigger checks if a trigger is in the valid list
func (c *CharacterCard) isValidTrigger(trigger string, validTriggers []string) bool {
	for _, valid := range validTriggers {
		if trigger == valid {
			return true
		}
	}
	return false
}

// validateProgression validates progression configuration (levels and achievements)
func (c *CharacterCard) validateProgression() error {
	if c.Progression == nil {
		return nil // Progression is optional
	}

	// Validate levels
	if len(c.Progression.Levels) == 0 {
		return fmt.Errorf("must have at least one level")
	}

	for i, level := range c.Progression.Levels {
		if err := c.validateProgressionLevel(level, i); err != nil {
			return fmt.Errorf("level %d (%s): %w", i, level.Name, err)
		}
	}

	// Validate achievements
	for i, achievement := range c.Progression.Achievements {
		if err := c.validateProgressionAchievement(achievement, i); err != nil {
			return fmt.Errorf("achievement %d (%s): %w", i, achievement.Name, err)
		}
	}

	return nil
}

// validateProgressionLevel validates a single level configuration
func (c *CharacterCard) validateProgressionLevel(level LevelConfig, index int) error {
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

	// Validate level-specific animations exist in main animations map
	for animName := range level.Animations {
		if _, exists := c.Animations[animName]; !exists {
			return fmt.Errorf("level animation '%s' not found in main animations map", animName)
		}
	}

	return nil
}

// validateProgressionAchievement validates a single achievement configuration
func (c *CharacterCard) validateProgressionAchievement(achievement AchievementConfig, index int) error {
	if len(achievement.Name) == 0 {
		return fmt.Errorf("name cannot be empty")
	}

	if len(achievement.Requirement) == 0 {
		return fmt.Errorf("must have at least one requirement")
	}

	// Validate that required stats exist in character stats
	if c.Stats != nil {
		for statName := range achievement.Requirement {
			if statName == "maintainAbove" {
				continue // Special requirement type
			}
			if _, exists := c.Stats[statName]; !exists {
				return fmt.Errorf("achievement requires stat '%s' which is not defined", statName)
			}
		}
	}

	return nil
}

// HasGameFeatures returns true if this character card includes game features
func (c *CharacterCard) HasGameFeatures() bool {
	return len(c.Stats) > 0
}

// GetRandomResponse returns a random response from a dialog's response list
// Uses time-based seeding for pseudo-randomness - good enough for desktop pets
func (d *Dialog) GetRandomResponse() string {
	if len(d.Responses) == 0 {
		return ""
	}

	// Simple time-based pseudo-random selection
	// For a desktop pet, this provides sufficient randomness
	index := int(time.Now().UnixNano()) % len(d.Responses)
	return d.Responses[index]
}

// CanTrigger checks if enough time has passed since the last trigger
func (d *Dialog) CanTrigger(lastTriggerTime time.Time) bool {
	return time.Since(lastTriggerTime) >= time.Duration(d.Cooldown)*time.Second
}

// validateRandomEvents validates the random events configuration
func (c *CharacterCard) validateRandomEvents() error {
	if len(c.RandomEvents) == 0 {
		return nil // Random events are optional
	}

	for i, event := range c.RandomEvents {
		if err := c.validateRandomEventConfig(event, i); err != nil {
			return fmt.Errorf("event %d (%s): %w", i, event.Name, err)
		}
	}

	return nil
}

// validateRandomEventConfig validates a single random event configuration
func (c *CharacterCard) validateRandomEventConfig(event RandomEventConfig, index int) error {
	if err := c.validateEventBasicFields(event); err != nil {
		return err
	}

	if err := c.validateEventNumericRanges(event); err != nil {
		return err
	}

	if err := c.validateEventAnimations(event); err != nil {
		return err
	}

	if err := c.validateEventResponses(event); err != nil {
		return err
	}

	if err := c.validateEventStatReferences(event); err != nil {
		return err
	}

	return nil
}

// validateEventBasicFields validates required string fields are not empty
func (c *CharacterCard) validateEventBasicFields(event RandomEventConfig) error {
	if len(event.Name) == 0 {
		return fmt.Errorf("name cannot be empty")
	}

	if len(event.Description) == 0 {
		return fmt.Errorf("description cannot be empty")
	}

	return nil
}

// validateEventNumericRanges validates numeric fields are within acceptable ranges
func (c *CharacterCard) validateEventNumericRanges(event RandomEventConfig) error {
	if event.Probability < 0.0 || event.Probability > 1.0 {
		return fmt.Errorf("probability must be 0.0-1.0, got %f", event.Probability)
	}

	if event.Cooldown < 0 || event.Cooldown > 86400 {
		return fmt.Errorf("cooldown must be 0-86400 seconds, got %d", event.Cooldown)
	}

	if event.Duration < 0 || event.Duration > 3600 {
		return fmt.Errorf("duration must be 0-3600 seconds, got %d", event.Duration)
	}

	return nil
}

// validateEventAnimations validates that all referenced animations exist
func (c *CharacterCard) validateEventAnimations(event RandomEventConfig) error {
	for _, animation := range event.Animations {
		if _, exists := c.Animations[animation]; !exists {
			return fmt.Errorf("animation '%s' not found in animations map", animation)
		}
	}

	return nil
}

// validateEventResponses validates response count is within limits
func (c *CharacterCard) validateEventResponses(event RandomEventConfig) error {
	if len(event.Responses) > 10 {
		return fmt.Errorf("must have 0-10 responses, got %d", len(event.Responses))
	}

	return nil
}

// validateEventStatReferences validates that referenced stats exist when character has stats defined
func (c *CharacterCard) validateEventStatReferences(event RandomEventConfig) error {
	// If no stats are defined, stat references are allowed but will be ignored at runtime
	if len(c.Stats) == 0 {
		return nil
	}

	for statName := range event.Effects {
		if _, exists := c.Stats[statName]; !exists {
			return fmt.Errorf("event effects reference stat '%s' which is not defined", statName)
		}
	}

	for statName := range event.Conditions {
		// Allow special romance condition types
		if c.isSpecialRomanceCondition(statName) {
			continue
		}

		if _, exists := c.Stats[statName]; !exists {
			return fmt.Errorf("event conditions reference stat '%s' which is not defined", statName)
		}
	}

	return nil
}

// isSpecialRomanceCondition checks if a condition name is a special romance condition type
func (c *CharacterCard) isSpecialRomanceCondition(conditionName string) bool {
	specialConditions := []string{
		"relationshipLevel",
		"interactionCount",
		"memoryCount",
	}

	for _, special := range specialConditions {
		if conditionName == special {
			return true
		}
	}

	return false
}

// Romance feature validation methods (Dating Simulator Phase 1)

// validateRomanceFeatures validates romance-specific configuration fields
func (c *CharacterCard) validateRomanceFeatures() error {
	// Personality validation (optional)
	if c.Personality != nil {
		if err := c.validatePersonalityConfig(); err != nil {
			return fmt.Errorf("personality: %w", err)
		}
	}

	// Romance dialogs validation (optional)
	if len(c.RomanceDialogs) > 0 {
		for i, dialog := range c.RomanceDialogs {
			if err := c.validateRomanceDialog(dialog, i); err != nil {
				return fmt.Errorf("romance dialog %d: %w", i, err)
			}
		}
	}

	// Romance events validation (optional)
	if len(c.RomanceEvents) > 0 {
		for i, event := range c.RomanceEvents {
			if err := c.validateRandomEventConfig(event, i); err != nil {
				return fmt.Errorf("romance event %d (%s): %w", i, event.Name, err)
			}
		}
	}

	return nil
}

// validatePersonalityConfig ensures personality configuration is valid
func (c *CharacterCard) validatePersonalityConfig() error {
	if c.Personality.Traits != nil {
		for name, value := range c.Personality.Traits {
			if value < 0.0 || value > 1.0 {
				return fmt.Errorf("trait '%s' must be 0.0-1.0, got %f", name, value)
			}
		}
	}

	if c.Personality.Compatibility != nil {
		for name, value := range c.Personality.Compatibility {
			if value < 0.0 || value > 5.0 {
				return fmt.Errorf("compatibility modifier '%s' must be 0.0-5.0, got %f", name, value)
			}
		}
	}

	return nil
}

// validateRomanceDialog validates an extended dialog configuration
func (c *CharacterCard) validateRomanceDialog(dialog DialogExtended, index int) error {
	// Validate base dialog
	if err := dialog.Dialog.Validate(c.Animations); err != nil {
		return err
	}

	// Validate romance requirements if present
	if dialog.Requirements != nil {
		if err := c.validateRomanceRequirements(dialog.Requirements); err != nil {
			return fmt.Errorf("requirements: %w", err)
		}
	}

	return nil
}

// validateRomanceRequirements validates romance requirement configuration
func (c *CharacterCard) validateRomanceRequirements(req *RomanceRequirement) error {
	// Validate stat requirements reference existing stats (if stats are defined)
	if len(c.Stats) > 0 && req.Stats != nil {
		for statName := range req.Stats {
			if _, exists := c.Stats[statName]; !exists {
				return fmt.Errorf("requirement references stat '%s' which is not defined", statName)
			}
		}
	}

	// Validate interaction count requirements
	if req.InteractionCount != nil {
		for interactionName := range req.InteractionCount {
			if c.Interactions != nil {
				if _, exists := c.Interactions[interactionName]; !exists {
					return fmt.Errorf("requirement references interaction '%s' which is not defined", interactionName)
				}
			}
		}
	}

	return nil
}

// validateDialogBackend ensures dialog backend configuration is valid when enabled
func (c *CharacterCard) validateDialogBackend() error {
	if c.DialogBackend == nil {
		return nil // Optional feature, validation not required when absent
	}

	return dialog.ValidateBackendConfig(*c.DialogBackend)
}

// HasRomanceFeatures returns true if this character card includes romance features
func (c *CharacterCard) HasRomanceFeatures() bool {
	return c.Personality != nil || len(c.RomanceDialogs) > 0 || len(c.RomanceEvents) > 0
}

// HasDialogBackend returns true if this character card has dialog backend configuration enabled
func (c *CharacterCard) HasDialogBackend() bool {
	return c.DialogBackend != nil && c.DialogBackend.Enabled
}

// Bug #4 Fix: Additional methods for more granular dialog backend state checking

// HasDialogBackendConfig returns true if this character has dialog backend configuration
// regardless of whether it's enabled or disabled. Useful for determining if a character
// was intended to have AI capabilities.
func (c *CharacterCard) HasDialogBackendConfig() bool {
	return c.DialogBackend != nil
}

// IsDialogBackendEnabled returns true if dialog backend is both configured and enabled
// This is equivalent to HasDialogBackend() but with a more descriptive name.
func (c *CharacterCard) IsDialogBackendEnabled() bool {
	return c.DialogBackend != nil && c.DialogBackend.Enabled
}

// GetDialogBackendStatus returns detailed information about dialog backend availability
// Returns: (hasConfig, isEnabled, reason) where reason explains the current state
func (c *CharacterCard) GetDialogBackendStatus() (bool, bool, string) {
	if c.DialogBackend == nil {
		return false, false, "No dialog backend configured"
	}

	if !c.DialogBackend.Enabled {
		return true, false, "Dialog backend configured but disabled"
	}

	return true, true, "Dialog backend configured and enabled"
}

// GetPersonalityTrait returns the value of a personality trait, defaulting to 0.5 if not found
func (c *CharacterCard) GetPersonalityTrait(trait string) float64 {
	if c.Personality == nil || c.Personality.Traits == nil {
		return 0.5 // Default neutral value
	}
	if value, exists := c.Personality.Traits[trait]; exists {
		return value
	}
	return 0.5
}

// GetCompatibilityModifier returns the compatibility modifier for a behavior, defaulting to 1.0
func (c *CharacterCard) GetCompatibilityModifier(behavior string) float64 {
	if c.Personality == nil || c.Personality.Compatibility == nil {
		return 1.0 // Default no modifier
	}
	if modifier, exists := c.Personality.Compatibility[behavior]; exists {
		return modifier
	}
	return 1.0
}

// validateGeneralEvents validates general dialog events configuration
func (c *CharacterCard) validateGeneralEvents() error {
	if len(c.GeneralEvents) == 0 {
		return nil // No general events to validate
	}

	for i, event := range c.GeneralEvents {
		if err := ValidateGeneralEvent(event); err != nil {
			return fmt.Errorf("event %d (%s): %w", i, event.Name, err)
		}
	}

	return nil
}

// validateGiftSystem validates gift system configuration
// Maintains backward compatibility by treating nil GiftSystem as valid
func (c *CharacterCard) validateGiftSystem() error {
	if c.GiftSystem == nil {
		return nil // Optional feature, validation not required when absent
	}

	// Validate inventory settings
	if c.GiftSystem.InventorySettings.MaxSlots < 1 {
		return fmt.Errorf("inventory maxSlots must be at least 1, got %d", c.GiftSystem.InventorySettings.MaxSlots)
	}
	if c.GiftSystem.InventorySettings.MaxSlots > 100 {
		return fmt.Errorf("inventory maxSlots cannot exceed 100, got %d", c.GiftSystem.InventorySettings.MaxSlots)
	}

	// Validate category preferences
	validCategories := []string{"food", "flowers", "books", "jewelry", "toys", "electronics", "clothing", "art", "practical", "expensive"}
	for _, category := range c.GiftSystem.Preferences.FavoriteCategories {
		if !sliceContains(validCategories, category) {
			return fmt.Errorf("invalid favorite category '%s', must be one of: %v", category, validCategories)
		}
	}
	for _, category := range c.GiftSystem.Preferences.DislikedCategories {
		if !sliceContains(validCategories, category) {
			return fmt.Errorf("invalid disliked category '%s', must be one of: %v", category, validCategories)
		}
	}

	// Validate personality responses
	for personality, response := range c.GiftSystem.Preferences.PersonalityResponses {
		if len(response.GiftReceived) == 0 {
			return fmt.Errorf("personality '%s' must have at least one gift received response", personality)
		}
		if len(response.GiftReceived) > 10 {
			return fmt.Errorf("personality '%s' cannot have more than 10 gift received responses, got %d", personality, len(response.GiftReceived))
		}
	}

	return nil
}

// HasGiftSystem returns true if this character card includes gift system configuration
func (c *CharacterCard) HasGiftSystem() bool {
	return c.GiftSystem != nil && c.GiftSystem.Enabled
}
