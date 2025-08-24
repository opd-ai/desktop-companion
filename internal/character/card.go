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

	if err := card.Validate(); err != nil {
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
	// Validate trigger
	validTriggers := []string{"click", "rightclick", "hover"}
	valid := false
	for _, trigger := range validTriggers {
		if d.Trigger == trigger {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("trigger must be one of %v, got: %s", validTriggers, d.Trigger)
	}

	// Validate responses
	if len(d.Responses) == 0 || len(d.Responses) > 10 {
		return fmt.Errorf("must have 1-10 responses, got %d", len(d.Responses))
	}

	for i, response := range d.Responses {
		if len(strings.TrimSpace(response)) == 0 {
			return fmt.Errorf("response %d cannot be empty", i)
		}
	}

	// Validate animation reference
	if _, exists := animations[d.Animation]; !exists {
		return fmt.Errorf("animation '%s' not found in animations map", d.Animation)
	}

	// Set default cooldown if not specified
	if d.Cooldown <= 0 {
		d.Cooldown = 5 // 5 second default
	}

	return nil
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
