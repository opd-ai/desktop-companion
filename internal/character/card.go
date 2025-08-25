package character

import (
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
	Progression  *ProgressionConfig           `json:"progression,omitempty"`
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
func (c *CharacterCard) Validate() error {
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

	if err := c.validateGameFeatures(); err != nil {
		return fmt.Errorf("game features: %w", err)
	}

	if err := c.validateProgression(); err != nil {
		return fmt.Errorf("progression: %w", err)
	}

	return nil
}

// ValidateWithBasePath ensures the character card has valid configuration including file existence checks
func (c *CharacterCard) ValidateWithBasePath(basePath string) error {
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

	if err := c.validateGameFeatures(); err != nil {
		return fmt.Errorf("game features: %w", err)
	}

	if err := c.validateProgression(); err != nil {
		return fmt.Errorf("progression: %w", err)
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
	if len(interaction.Triggers) == 0 {
		return fmt.Errorf("must have at least one trigger")
	}

	// Validate triggers
	validTriggers := []string{"click", "rightclick", "doubleclick", "shift+click", "hover"}
	for _, trigger := range interaction.Triggers {
		if !c.isValidTrigger(trigger, validTriggers) {
			return fmt.Errorf("invalid trigger '%s', must be one of %v", trigger, validTriggers)
		}
	}

	// Validate animations exist
	for _, animation := range interaction.Animations {
		if _, exists := c.Animations[animation]; !exists {
			return fmt.Errorf("animation '%s' not found in animations map", animation)
		}
	}

	// Validate responses
	if len(interaction.Responses) == 0 || len(interaction.Responses) > 10 {
		return fmt.Errorf("must have 1-10 responses, got %d", len(interaction.Responses))
	}

	// Validate cooldown
	if interaction.Cooldown < 0 || interaction.Cooldown > 3600 {
		return fmt.Errorf("cooldown must be 0-3600 seconds, got %d", interaction.Cooldown)
	}

	return nil
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
