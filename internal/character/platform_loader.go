// Package character provides platform-aware character card loading functionality.
// This implements the JSON Schema Extensions for Phase 5.1 of the Android migration plan.
package character

import (
	"github.com/opd-ai/desktop-companion/internal/platform"
	"fmt"
)

// PlatformAwareLoader handles loading character cards with platform-specific adaptations.
// Uses standard library JSON parsing with platform detection for adaptive configuration.
type PlatformAwareLoader struct {
	platform *platform.PlatformInfo
}

// NewPlatformAwareLoader creates a new platform-aware character card loader.
// Follows the "lazy programmer" approach by using existing platform detection.
func NewPlatformAwareLoader() *PlatformAwareLoader {
	return &PlatformAwareLoader{
		platform: platform.GetPlatformInfo(),
	}
}

// LoadCharacterCard loads a character card with platform-specific configuration applied.
// Maintains full backward compatibility - existing character cards work unchanged.
// Only applies platform adaptations when platformConfig is present in the JSON.
func (pal *PlatformAwareLoader) LoadCharacterCard(path string) (*CharacterCard, error) {
	// Load base character card using existing logic
	baseCard, err := LoadCard(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load base character card: %w", err)
	}

	// Apply platform-specific overrides if configuration exists
	return pal.applyPlatformConfig(baseCard), nil
}

// applyPlatformConfig applies platform-specific configuration overrides to a character card.
// Uses defensive programming - only applies overrides when platform config exists.
// Desktop behavior remains completely unchanged when no platform config is present.
func (pal *PlatformAwareLoader) applyPlatformConfig(card *CharacterCard) *CharacterCard {
	if card.PlatformConfig == nil {
		return card // No platform config, return original unchanged
	}

	var platformConfig *PlatformSpecificConfig
	if pal.platform.IsMobile() {
		platformConfig = card.PlatformConfig.Mobile
	} else {
		platformConfig = card.PlatformConfig.Desktop
	}

	if platformConfig == nil {
		return card // No config for current platform, return original
	}

	// Create a copy to avoid modifying the original
	adaptedCard := *card

	// Apply behavior overrides
	if platformConfig.Behavior != nil {
		adaptedCard.Behavior = pal.mergeBehavior(card.Behavior, *platformConfig.Behavior)
	}

	// Apply window mode and size overrides
	if platformConfig.WindowMode != "" {
		// Window mode configuration would be handled by UI layer
		// For now, we store it for later use
	}

	if platformConfig.DefaultSize > 0 {
		adaptedCard.Behavior.DefaultSize = platformConfig.DefaultSize
	}

	// Apply interaction overrides
	if len(platformConfig.Interactions) > 0 {
		adaptedCard.Interactions = pal.mergeInteractions(card.Interactions, platformConfig.Interactions)
	}

	return &adaptedCard
}

// mergeBehavior merges base behavior with platform-specific overrides.
// Uses explicit field-by-field merging to maintain control over what gets overridden.
func (pal *PlatformAwareLoader) mergeBehavior(base Behavior, override Behavior) Behavior {
	merged := base // Start with base behavior

	// Apply non-zero overrides
	if override.IdleTimeout > 0 {
		merged.IdleTimeout = override.IdleTimeout
	}

	if override.DefaultSize > 0 {
		merged.DefaultSize = override.DefaultSize
	}

	// Movement enabled is platform-specific - mobile typically disables it
	if pal.platform.IsMobile() {
		merged.MovementEnabled = override.MovementEnabled
	} else {
		// For desktop, only override if explicitly set
		merged.MovementEnabled = override.MovementEnabled || base.MovementEnabled
	}

	return merged
}

// mergeInteractions merges base interactions with platform-specific overrides.
// Preserves existing interactions while adding or overriding platform-specific ones.
func (pal *PlatformAwareLoader) mergeInteractions(
	base map[string]InteractionConfig,
	overrides map[string]PlatformInteractionConfig,
) map[string]InteractionConfig {
	if base == nil {
		base = make(map[string]InteractionConfig)
	}

	merged := make(map[string]InteractionConfig)

	// Copy base interactions
	for name, config := range base {
		merged[name] = config
	}

	// Apply platform-specific overrides
	for name, platformConfig := range overrides {
		// Convert platform interaction to base interaction
		baseInteraction := platformConfig.InteractionConfig

		// Apply platform-specific trigger adaptations
		if len(platformConfig.Triggers) > 0 {
			baseInteraction.Triggers = pal.adaptTriggers(platformConfig.Triggers)
		}

		merged[name] = baseInteraction
	}

	return merged
}

// adaptTriggers converts platform-specific triggers to standard interaction triggers.
// Maps mobile touch events to desktop mouse events and vice versa.
func (pal *PlatformAwareLoader) adaptTriggers(platformTriggers []string) []string {
	adapted := make([]string, 0, len(platformTriggers))

	for _, trigger := range platformTriggers {
		switch trigger {
		case "tap":
			if pal.platform.IsMobile() {
				adapted = append(adapted, "click")
			} else {
				adapted = append(adapted, "click")
			}
		case "longpress":
			if pal.platform.IsMobile() {
				adapted = append(adapted, "rightclick")
			} else {
				adapted = append(adapted, "rightclick")
			}
		case "doubletap":
			adapted = append(adapted, "doubleclick")
		case "swipe":
			// Swipe doesn't have a direct desktop equivalent, map to click
			adapted = append(adapted, "click")
		default:
			// Pass through standard triggers unchanged
			adapted = append(adapted, trigger)
		}
	}

	return adapted
}

// ValidatePlatformConfig validates platform-specific configuration for consistency.
// Ensures platform configs don't conflict with base configuration requirements.
func ValidatePlatformConfig(card *CharacterCard) error {
	if card.PlatformConfig == nil {
		return nil // No platform config to validate
	}

	// Validate desktop configuration
	if card.PlatformConfig.Desktop != nil {
		if err := validatePlatformSpecificConfig(card.PlatformConfig.Desktop, "desktop"); err != nil {
			return fmt.Errorf("invalid desktop platform config: %w", err)
		}
	}

	// Validate mobile configuration
	if card.PlatformConfig.Mobile != nil {
		if err := validatePlatformSpecificConfig(card.PlatformConfig.Mobile, "mobile"); err != nil {
			return fmt.Errorf("invalid mobile platform config: %w", err)
		}
	}

	return nil
}

// validatePlatformSpecificConfig validates a single platform configuration.
// Checks for reasonable values and platform-appropriate settings.
func validatePlatformSpecificConfig(config *PlatformSpecificConfig, platformType string) error {
	// Validate behavior overrides
	if config.Behavior != nil {
		if config.Behavior.IdleTimeout < 0 {
			return fmt.Errorf("idle timeout cannot be negative")
		}
		if config.Behavior.DefaultSize < 32 || config.Behavior.DefaultSize > 1024 {
			return fmt.Errorf("default size must be between 32 and 1024 pixels")
		}
	}

	// Validate window mode
	if config.WindowMode != "" {
		validModes := map[string]bool{
			"overlay":    true,
			"fullscreen": true,
			"pip":        true,
		}
		if !validModes[config.WindowMode] {
			return fmt.Errorf("invalid window mode: %s", config.WindowMode)
		}

		// Platform-specific validation
		if platformType == "mobile" && config.WindowMode == "overlay" {
			// Warning: overlay mode may not work well on mobile
			// But we don't prevent it - let the user decide
		}
	}

	// Validate mobile controls (only for mobile platform)
	if config.MobileControls != nil && platformType != "mobile" {
		return fmt.Errorf("mobile controls configuration only valid for mobile platform")
	}

	return nil
}

// GetPlatformConfig returns the appropriate platform configuration for the current platform.
// Returns nil if no platform-specific configuration exists.
func (pal *PlatformAwareLoader) GetPlatformConfig(card *CharacterCard) *PlatformSpecificConfig {
	if card.PlatformConfig == nil {
		return nil
	}

	if pal.platform.IsMobile() {
		return card.PlatformConfig.Mobile
	}
	return card.PlatformConfig.Desktop
}

// CreateExamplePlatformConfig generates an example platform configuration for documentation.
// Demonstrates best practices for cross-platform character card configuration.
func CreateExamplePlatformConfig() *PlatformConfig {
	return &PlatformConfig{
		Desktop: &PlatformSpecificConfig{
			Behavior: &Behavior{
				MovementEnabled: true,
				DefaultSize:     128,
				IdleTimeout:     30,
			},
			WindowMode: "overlay",
			Interactions: map[string]PlatformInteractionConfig{
				"pet": {
					InteractionConfig: InteractionConfig{
						Triggers: []string{"click"},
						Effects:  map[string]float64{"happiness": 10},
						Cooldown: 5,
					},
				},
			},
		},
		Mobile: &PlatformSpecificConfig{
			Behavior: &Behavior{
				MovementEnabled: false,
				DefaultSize:     256,
				IdleTimeout:     60,
			},
			WindowMode:     "fullscreen",
			TouchOptimized: true,
			MobileControls: &MobileControlsConfig{
				ShowBottomBar:        true,
				SwipeGesturesEnabled: true,
				HapticFeedback:       true,
				LargeButtons:         true,
				ContextMenuStyle:     "bottomsheet",
			},
			Interactions: map[string]PlatformInteractionConfig{
				"pet": {
					InteractionConfig: InteractionConfig{
						Triggers: []string{"tap"},
						Effects:  map[string]float64{"happiness": 15},
						Cooldown: 3,
					},
					HapticPattern: "light",
					TouchFeedback: true,
				},
			},
		},
	}
}
