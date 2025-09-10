package bot

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// PersonalityTraits defines standardized personality trait names for consistency.
// These traits are used across character cards and bot behavior algorithms.
var PersonalityTraits = struct {
	// Social interaction traits
	Chattiness  string
	Helpfulness string
	Playfulness string
	Curiosity   string

	// Emotional characteristics
	Empathy       string
	Assertiveness string
	Patience      string
	Enthusiasm    string

	// Behavioral tendencies
	Independence string
	Creativity   string
	Analytical   string
	Spontaneity  string
}{
	// Social traits
	Chattiness:  "chattiness",
	Helpfulness: "helpfulness",
	Playfulness: "playfulness",
	Curiosity:   "curiosity",

	// Emotional traits
	Empathy:       "empathy",
	Assertiveness: "assertiveness",
	Patience:      "patience",
	Enthusiasm:    "enthusiasm",

	// Behavioral traits
	Independence: "independence",
	Creativity:   "creativity",
	Analytical:   "analytical",
	Spontaneity:  "spontaneity",
}

// PersonalityArchetype represents a predefined personality template.
// These provide consistent starting points for character creation.
type PersonalityArchetype struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Traits      map[string]float64  `json:"traits"`
	Behavior    PersonalityBehavior `json:"behavior"`
	Examples    []string            `json:"examples,omitempty"`
}

// PersonalityBehavior defines behavioral parameters driven by personality traits.
// These values control timing, interaction patterns, and decision making.
type PersonalityBehavior struct {
	ResponseDelay       string   `json:"responseDelay"`       // "1-3s" format for delay ranges
	InteractionRate     float64  `json:"interactionRate"`     // Actions per minute (0.1-10.0)
	Attention           float64  `json:"attention"`           // How quickly bot notices events (0.0-1.0)
	MaxActionsPerMinute int      `json:"maxActionsPerMinute"` // Rate limiting
	MinTimeBetweenSame  int      `json:"minTimeBetweenSame"`  // Seconds between same action type
	PreferredActions    []string `json:"preferredActions"`    // Actions this bot prefers
}

// ParseResponseDelay converts string delay specification to time.Duration range.
// Supports formats like "2s", "1s-3s", "500ms-2s" for flexible timing configuration.
func (pb *PersonalityBehavior) ParseResponseDelay() (min, max time.Duration, err error) {
	delay := strings.TrimSpace(pb.ResponseDelay)
	if delay == "" {
		return pb.getDefaultDelayRange()
	}

	// Handle range format "1s-3s" or "500ms-2s"
	if strings.Contains(delay, "-") {
		return pb.parseDelayRange(delay)
	}

	// Handle single value "2s" - use as base with ±25% variation
	return pb.parseSingleDelay(delay)
}

// getDefaultDelayRange returns the default delay range when no delay is specified.
func (pb *PersonalityBehavior) getDefaultDelayRange() (min, max time.Duration, err error) {
	return 2 * time.Second, 4 * time.Second, nil
}

// parseDelayRange parses range format delays like "1s-3s" or "500ms-2s".
func (pb *PersonalityBehavior) parseDelayRange(delay string) (min, max time.Duration, err error) {
	parts := strings.Split(delay, "-")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid delay range format: %s", delay)
	}

	minStr := strings.TrimSpace(parts[0])
	maxStr := strings.TrimSpace(parts[1])

	// If minStr doesn't have a unit, extract it from maxStr
	if !strings.ContainsAny(minStr, "msnhμ") {
		unit := pb.extractTimeUnit(maxStr)
		if unit != "" {
			minStr += unit
		}
	}

	min, err = time.ParseDuration(minStr)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid minimum delay: %w", err)
	}

	max, err = time.ParseDuration(maxStr)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid maximum delay: %w", err)
	}

	if min > max {
		return 0, 0, fmt.Errorf("minimum delay (%v) cannot be greater than maximum (%v)", min, max)
	}

	return min, max, nil
}

// extractTimeUnit extracts the unit portion from a duration string like "3s" -> "s".
func (pb *PersonalityBehavior) extractTimeUnit(durationStr string) string {
	for i := len(durationStr) - 1; i >= 0; i-- {
		if durationStr[i] >= '0' && durationStr[i] <= '9' {
			return durationStr[i+1:]
		}
	}
	return ""
}

// parseSingleDelay parses single value delays like "2s" with ±25% variation.
func (pb *PersonalityBehavior) parseSingleDelay(delay string) (min, max time.Duration, err error) {
	base, err := time.ParseDuration(delay)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid delay format: %w", err)
	}

	variation := time.Duration(float64(base) * 0.25)
	return base - variation, base + variation, nil
}

// PersonalityManager handles personality loading, validation, and archetype management.
// Follows the project's standard library approach using encoding/json for configuration.
type PersonalityManager struct {
	archetypes map[string]PersonalityArchetype
}

// NewPersonalityManager creates a personality manager with built-in archetypes.
// Provides immediate functionality without requiring external configuration files.
func NewPersonalityManager() *PersonalityManager {
	pm := &PersonalityManager{
		archetypes: make(map[string]PersonalityArchetype),
	}

	// Load built-in personality archetypes
	pm.loadBuiltinArchetypes()

	return pm
}

// GetArchetype returns a personality archetype by name.
// Returns error if archetype doesn't exist to prevent misconfiguration.
func (pm *PersonalityManager) GetArchetype(name string) (*PersonalityArchetype, error) {
	archetype, exists := pm.archetypes[strings.ToLower(name)]
	if !exists {
		return nil, fmt.Errorf("personality archetype '%s' not found", name)
	}

	// Return copy to prevent external modification
	result := archetype
	return &result, nil
}

// ListArchetypes returns all available personality archetype names.
// Useful for configuration validation and user interface display.
func (pm *PersonalityManager) ListArchetypes() []string {
	names := make([]string, 0, len(pm.archetypes))
	for name := range pm.archetypes {
		names = append(names, name)
	}
	return names
}

// CreatePersonality converts a PersonalityArchetype to BotPersonality.
// Handles the translation between JSON configuration and runtime behavior.
func (pm *PersonalityManager) CreatePersonality(archetype *PersonalityArchetype) (*BotPersonality, error) {
	if archetype == nil {
		return nil, fmt.Errorf("archetype cannot be nil")
	}

	// Parse response delay range
	minDelay, maxDelay, err := archetype.Behavior.ParseResponseDelay()
	if err != nil {
		return nil, fmt.Errorf("invalid response delay: %w", err)
	}

	// Use average of min/max for ResponseDelay field
	avgDelay := minDelay + (maxDelay-minDelay)/2

	personality := &BotPersonality{
		ResponseDelay:       avgDelay,
		InteractionRate:     archetype.Behavior.InteractionRate,
		Attention:           archetype.Behavior.Attention,
		SocialTendencies:    make(map[string]float64),
		EmotionalProfile:    make(map[string]float64),
		MaxActionsPerMinute: archetype.Behavior.MaxActionsPerMinute,
		MinTimeBetweenSame:  archetype.Behavior.MinTimeBetweenSame,
		PreferredActions:    make([]string, len(archetype.Behavior.PreferredActions)),
	}

	// Copy traits to appropriate categories
	for trait, value := range archetype.Traits {
		switch trait {
		case PersonalityTraits.Chattiness, PersonalityTraits.Helpfulness,
			PersonalityTraits.Playfulness, PersonalityTraits.Curiosity:
			personality.SocialTendencies[trait] = value
		case PersonalityTraits.Empathy, PersonalityTraits.Assertiveness,
			PersonalityTraits.Patience, PersonalityTraits.Enthusiasm:
			personality.EmotionalProfile[trait] = value
		default:
			// Store unknown traits in SocialTendencies for backward compatibility
			personality.SocialTendencies[trait] = value
		}
	}

	// Copy preferred actions
	copy(personality.PreferredActions, archetype.Behavior.PreferredActions)

	return personality, nil
}

// LoadFromJSON creates a PersonalityArchetype from JSON data.
// Enables character cards to define custom personality configurations.
func (pm *PersonalityManager) LoadFromJSON(data []byte) (*PersonalityArchetype, error) {
	var archetype PersonalityArchetype
	if err := json.Unmarshal(data, &archetype); err != nil {
		return nil, fmt.Errorf("failed to parse personality JSON: %w", err)
	}

	if err := pm.validateArchetype(&archetype); err != nil {
		return nil, fmt.Errorf("invalid personality archetype: %w", err)
	}

	return &archetype, nil
}

// validateArchetype ensures personality archetype values are within acceptable ranges.
// Prevents configuration errors that could cause poor bot behavior.
func (pm *PersonalityManager) validateArchetype(archetype *PersonalityArchetype) error {
	if err := pm.validateArchetypeBasics(archetype); err != nil {
		return err
	}

	if err := pm.validateArchetypeTraits(archetype); err != nil {
		return err
	}

	if err := pm.validateBehaviorParameters(&archetype.Behavior); err != nil {
		return err
	}

	if err := pm.validateResponseDelayFormat(&archetype.Behavior); err != nil {
		return err
	}

	return nil
}

// validateArchetypeBasics checks fundamental archetype properties like name.
func (pm *PersonalityManager) validateArchetypeBasics(archetype *PersonalityArchetype) error {
	if archetype.Name == "" {
		return fmt.Errorf("personality name cannot be empty")
	}
	return nil
}

// validateArchetypeTraits ensures all trait values are within the valid 0.0-1.0 range.
func (pm *PersonalityManager) validateArchetypeTraits(archetype *PersonalityArchetype) error {
	for trait, value := range archetype.Traits {
		if value < 0.0 || value > 1.0 {
			return fmt.Errorf("trait '%s' value %.2f must be between 0.0 and 1.0", trait, value)
		}
	}
	return nil
}

// validateBehaviorParameters checks behavior configuration values are within acceptable ranges.
func (pm *PersonalityManager) validateBehaviorParameters(behavior *PersonalityBehavior) error {
	if behavior.InteractionRate < 0.1 || behavior.InteractionRate > 10.0 {
		return fmt.Errorf("interaction rate %.2f must be between 0.1 and 10.0", behavior.InteractionRate)
	}

	if behavior.Attention < 0.0 || behavior.Attention > 1.0 {
		return fmt.Errorf("attention %.2f must be between 0.0 and 1.0", behavior.Attention)
	}

	if behavior.MaxActionsPerMinute < 1 || behavior.MaxActionsPerMinute > 60 {
		return fmt.Errorf("max actions per minute %d must be between 1 and 60", behavior.MaxActionsPerMinute)
	}

	if behavior.MinTimeBetweenSame < 0 || behavior.MinTimeBetweenSame > 300 {
		return fmt.Errorf("min time between same action %d must be between 0 and 300 seconds", behavior.MinTimeBetweenSame)
	}

	return nil
}

// validateResponseDelayFormat ensures the response delay configuration is parseable and valid.
func (pm *PersonalityManager) validateResponseDelayFormat(behavior *PersonalityBehavior) error {
	_, _, err := behavior.ParseResponseDelay()
	if err != nil {
		return fmt.Errorf("invalid response delay: %w", err)
	}
	return nil
}

// loadBuiltinArchetypes initializes standard personality archetypes.
// Provides commonly used personality types without external configuration.
func (pm *PersonalityManager) loadBuiltinArchetypes() {
	// Social and Outgoing - High interaction, chatty, helpful
	pm.archetypes["social"] = PersonalityArchetype{
		Name:        "Social",
		Description: "Outgoing, chatty companion who loves interaction and helping others",
		Traits: map[string]float64{
			PersonalityTraits.Chattiness:    0.9,
			PersonalityTraits.Helpfulness:   0.8,
			PersonalityTraits.Playfulness:   0.7,
			PersonalityTraits.Curiosity:     0.8,
			PersonalityTraits.Empathy:       0.9,
			PersonalityTraits.Enthusiasm:    0.8,
			PersonalityTraits.Assertiveness: 0.6,
			PersonalityTraits.Patience:      0.7,
		},
		Behavior: PersonalityBehavior{
			ResponseDelay:       "1s-3s",
			InteractionRate:     4.0,
			Attention:           0.9,
			MaxActionsPerMinute: 8,
			MinTimeBetweenSame:  5,
			PreferredActions:    []string{"chat", "click", "play"},
		},
		Examples: []string{"Always ready to chat", "Offers help frequently", "Initiates conversations"},
	}

	// Shy and Reserved - Lower interaction, thoughtful responses
	pm.archetypes["shy"] = PersonalityArchetype{
		Name:        "Shy",
		Description: "Quiet, thoughtful companion who prefers observation to interaction",
		Traits: map[string]float64{
			PersonalityTraits.Chattiness:    0.3,
			PersonalityTraits.Helpfulness:   0.7,
			PersonalityTraits.Playfulness:   0.4,
			PersonalityTraits.Curiosity:     0.6,
			PersonalityTraits.Empathy:       0.8,
			PersonalityTraits.Enthusiasm:    0.4,
			PersonalityTraits.Assertiveness: 0.2,
			PersonalityTraits.Patience:      0.9,
		},
		Behavior: PersonalityBehavior{
			ResponseDelay:       "3s-7s",
			InteractionRate:     1.0,
			Attention:           0.8,
			MaxActionsPerMinute: 3,
			MinTimeBetweenSame:  15,
			PreferredActions:    []string{"click"},
		},
		Examples: []string{"Takes time to respond", "Observes before acting", "Thoughtful interactions"},
	}

	// Playful and Energetic - High energy, loves games and fun
	pm.archetypes["playful"] = PersonalityArchetype{
		Name:        "Playful",
		Description: "Energetic, fun-loving companion who enjoys games and playful interactions",
		Traits: map[string]float64{
			PersonalityTraits.Chattiness:    0.7,
			PersonalityTraits.Helpfulness:   0.6,
			PersonalityTraits.Playfulness:   0.9,
			PersonalityTraits.Curiosity:     0.8,
			PersonalityTraits.Empathy:       0.6,
			PersonalityTraits.Enthusiasm:    0.9,
			PersonalityTraits.Assertiveness: 0.7,
			PersonalityTraits.Patience:      0.4,
			PersonalityTraits.Spontaneity:   0.8,
		},
		Behavior: PersonalityBehavior{
			ResponseDelay:       "500ms-2s",
			InteractionRate:     5.0,
			Attention:           0.8,
			MaxActionsPerMinute: 10,
			MinTimeBetweenSame:  3,
			PreferredActions:    []string{"play", "click", "chat"},
		},
		Examples: []string{"Quick to initiate play", "Energetic responses", "Loves interactive games"},
	}

	// Helper and Supportive - Focused on helping and supporting users
	pm.archetypes["helper"] = PersonalityArchetype{
		Name:        "Helper",
		Description: "Supportive, caring companion focused on helping and assisting users",
		Traits: map[string]float64{
			PersonalityTraits.Chattiness:    0.6,
			PersonalityTraits.Helpfulness:   0.9,
			PersonalityTraits.Playfulness:   0.5,
			PersonalityTraits.Curiosity:     0.7,
			PersonalityTraits.Empathy:       0.9,
			PersonalityTraits.Enthusiasm:    0.7,
			PersonalityTraits.Assertiveness: 0.5,
			PersonalityTraits.Patience:      0.8,
			PersonalityTraits.Analytical:    0.7,
		},
		Behavior: PersonalityBehavior{
			ResponseDelay:       "2s-4s",
			InteractionRate:     3.0,
			Attention:           0.9,
			MaxActionsPerMinute: 6,
			MinTimeBetweenSame:  8,
			PreferredActions:    []string{"help", "chat", "click"},
		},
		Examples: []string{"Offers assistance proactively", "Patient with questions", "Focuses on user needs"},
	}

	// Balanced Default - Well-rounded personality for general use
	pm.archetypes["balanced"] = PersonalityArchetype{
		Name:        "Balanced",
		Description: "Well-rounded companion with moderate traits suitable for most users",
		Traits: map[string]float64{
			PersonalityTraits.Chattiness:    0.6,
			PersonalityTraits.Helpfulness:   0.7,
			PersonalityTraits.Playfulness:   0.6,
			PersonalityTraits.Curiosity:     0.7,
			PersonalityTraits.Empathy:       0.7,
			PersonalityTraits.Enthusiasm:    0.6,
			PersonalityTraits.Assertiveness: 0.5,
			PersonalityTraits.Patience:      0.7,
		},
		Behavior: PersonalityBehavior{
			ResponseDelay:       "2s-4s",
			InteractionRate:     2.5,
			Attention:           0.7,
			MaxActionsPerMinute: 5,
			MinTimeBetweenSame:  10,
			PreferredActions:    []string{"click", "chat"},
		},
		Examples: []string{"Adapts to user preferences", "Moderate in all behaviors", "Reliable companion"},
	}
}
