package character

import (
	"fmt"
	"time"
)

// GeneralDialogEvent represents a user-initiated interactive scenario
// This extends the existing RandomEventConfig structure for user-triggered events
type GeneralDialogEvent struct {
	RandomEventConfig               // Embed existing random event structure
	Category          string        `json:"category"`                  // "conversation", "roleplay", "game", "humor"
	Trigger           string        `json:"trigger"`                   // Custom trigger identifier
	Interactive       bool          `json:"interactive"`               // Whether event supports user choices
	Choices           []EventChoice `json:"choices,omitempty"`         // User interaction choices
	FollowUpEvents    []string      `json:"followUpEvents,omitempty"`  // Chain to other events
	Keywords          []string      `json:"keywords,omitempty"`        // Keywords for event discovery
	Difficulty        string        `json:"difficulty,omitempty"`      // "easy", "normal", "hard"
	MinRelationship   string        `json:"minRelationship,omitempty"` // Minimum relationship level required
}

// EventChoice represents a user choice within an interactive event
type EventChoice struct {
	Text         string                        `json:"text"`                   // Choice text to display to user
	Effects      map[string]float64            `json:"effects"`                // Stat effects of selecting this choice
	NextEvent    string                        `json:"nextEvent,omitempty"`    // Next event to trigger after this choice
	Requirements map[string]map[string]float64 `json:"requirements,omitempty"` // Requirements to show this choice
	Animation    string                        `json:"animation,omitempty"`    // Override animation for this choice
	Responses    []string                      `json:"responses,omitempty"`    // Custom responses for this choice
	Disabled     bool                          `json:"disabled,omitempty"`     // Whether choice is temporarily disabled
}

// GeneralEventManager handles user-initiated interactive scenarios
// Operates alongside existing RandomEventManager for automatic events
type GeneralEventManager struct {
	events            []GeneralDialogEvent // Available general events
	activeEvent       *GeneralDialogEvent  // Currently active event
	eventCooldowns    map[string]time.Time // Cooldown tracking per event
	userChoiceHistory map[string][]int     // Track user choices for learning
	enabled           bool                 // Whether general events are enabled
}

// NewGeneralEventManager creates a new manager for general dialog events
func NewGeneralEventManager(events []GeneralDialogEvent, enabled bool) *GeneralEventManager {
	return &GeneralEventManager{
		events:            events,
		eventCooldowns:    make(map[string]time.Time),
		userChoiceHistory: make(map[string][]int),
		enabled:           enabled && len(events) > 0,
	}
}

// GetAvailableEvents returns events that can currently be triggered
func (gem *GeneralEventManager) GetAvailableEvents(gameState *GameState) []GeneralDialogEvent {
	if !gem.enabled {
		return nil
	}

	var available []GeneralDialogEvent
	now := time.Now()

	for _, event := range gem.events {
		if gem.canTriggerEvent(event, now, gameState) {
			available = append(available, event)
		}
	}

	return available
}

// GetEventsByCategory returns available events filtered by category
func (gem *GeneralEventManager) GetEventsByCategory(category string, gameState *GameState) []GeneralDialogEvent {
	available := gem.GetAvailableEvents(gameState)
	var filtered []GeneralDialogEvent

	for _, event := range available {
		if event.Category == category {
			filtered = append(filtered, event)
		}
	}

	return filtered
}

// TriggerEvent initiates a specific general event by name
func (gem *GeneralEventManager) TriggerEvent(eventName string, gameState *GameState) (*GeneralDialogEvent, error) {
	if !gem.enabled {
		return nil, fmt.Errorf("general events are disabled")
	}

	event := gem.findEventByName(eventName)
	if event == nil {
		return nil, fmt.Errorf("event '%s' not found", eventName)
	}

	now := time.Now()
	if !gem.canTriggerEvent(*event, now, gameState) {
		return nil, fmt.Errorf("event '%s' cannot be triggered right now", eventName)
	}

	// Record cooldown
	gem.eventCooldowns[eventName] = now

	// Set as active event if interactive
	if event.Interactive {
		gem.activeEvent = event
	}

	return event, nil
}

// SubmitChoice handles user choice selection in the active interactive event
func (gem *GeneralEventManager) SubmitChoice(choiceIndex int, gameState *GameState) (*EventChoice, string, error) {
	if gem.activeEvent == nil {
		return nil, "", fmt.Errorf("no active interactive event")
	}

	if !gem.activeEvent.Interactive {
		return nil, "", fmt.Errorf("active event is not interactive")
	}

	if choiceIndex < 0 || choiceIndex >= len(gem.activeEvent.Choices) {
		return nil, "", fmt.Errorf("invalid choice index: %d", choiceIndex)
	}

	choice := gem.activeEvent.Choices[choiceIndex]

	// Check choice requirements
	if len(choice.Requirements) > 0 && gameState != nil {
		if !gameState.CanSatisfyRequirements(choice.Requirements) {
			return nil, "", fmt.Errorf("choice requirements not met")
		}
	}

	// Record choice for learning
	gem.recordUserChoice(gem.activeEvent.Name, choiceIndex)

	// Apply choice effects
	if gameState != nil && len(choice.Effects) > 0 {
		gameState.ApplyInteractionEffects(choice.Effects)
	}

	// Determine next action
	nextAction := ""
	if choice.NextEvent != "" {
		nextAction = choice.NextEvent
		gem.activeEvent = nil // Clear active event before transitioning
	} else {
		gem.activeEvent = nil // Event complete
	}

	return &choice, nextAction, nil
}

// GetActiveEvent returns the currently active interactive event
func (gem *GeneralEventManager) GetActiveEvent() *GeneralDialogEvent {
	return gem.activeEvent
}

// ClearActiveEvent clears the currently active event (for cancellation)
func (gem *GeneralEventManager) ClearActiveEvent() {
	gem.activeEvent = nil
}

// IsEventAvailable checks if a specific event can be triggered
func (gem *GeneralEventManager) IsEventAvailable(eventName string, gameState *GameState) bool {
	if !gem.enabled {
		return false
	}

	event := gem.findEventByName(eventName)
	if event == nil {
		return false
	}

	now := time.Now()
	return gem.canTriggerEvent(*event, now, gameState)
}

// GetUserChoiceHistory returns the user's choice history for an event (for learning)
func (gem *GeneralEventManager) GetUserChoiceHistory(eventName string) []int {
	return gem.userChoiceHistory[eventName]
}

// Private helper methods

// findEventByName locates an event by its name
func (gem *GeneralEventManager) findEventByName(name string) *GeneralDialogEvent {
	for i := range gem.events {
		if gem.events[i].Name == name {
			return &gem.events[i]
		}
	}
	return nil
}

// canTriggerEvent checks if an event can currently be triggered
func (gem *GeneralEventManager) canTriggerEvent(event GeneralDialogEvent, now time.Time, gameState *GameState) bool {
	// Check cooldown
	if lastTrigger, exists := gem.eventCooldowns[event.Name]; exists {
		cooldownDuration := time.Duration(event.Cooldown) * time.Second
		if now.Sub(lastTrigger) < cooldownDuration {
			return false
		}
	}

	// Check conditions (stat requirements) - only if conditions exist and gameState is available
	if len(event.Conditions) > 0 {
		if gameState == nil {
			return false // Cannot check conditions without game state
		}
		if !gameState.CanSatisfyRequirements(event.Conditions) {
			return false
		}
	}

	// Check minimum relationship level - only if specified and gameState is available
	if event.MinRelationship != "" {
		if gameState == nil {
			return false // Cannot check relationship without game state
		}
		currentLevel := gameState.GetRelationshipLevel()
		if !gem.meetsRelationshipRequirement(currentLevel, event.MinRelationship) {
			return false
		}
	}

	return true
}

// meetsRelationshipRequirement checks if current relationship meets minimum
func (gem *GeneralEventManager) meetsRelationshipRequirement(current, required string) bool {
	levels := []string{"Stranger", "Friend", "Close Friend", "Romantic Interest", "Partner"}

	currentIndex := gem.findRelationshipIndex(current, levels)
	requiredIndex := gem.findRelationshipIndex(required, levels)

	return currentIndex >= requiredIndex
}

// findRelationshipIndex finds the index of a relationship level
func (gem *GeneralEventManager) findRelationshipIndex(level string, levels []string) int {
	for i, l := range levels {
		if l == level {
			return i
		}
	}
	return 0 // Default to lowest level
}

// recordUserChoice records a user's choice for learning purposes
func (gem *GeneralEventManager) recordUserChoice(eventName string, choiceIndex int) {
	if gem.userChoiceHistory[eventName] == nil {
		gem.userChoiceHistory[eventName] = make([]int, 0)
	}

	gem.userChoiceHistory[eventName] = append(gem.userChoiceHistory[eventName], choiceIndex)

	// Keep only recent choices (last 10) to prevent memory growth
	if len(gem.userChoiceHistory[eventName]) > 10 {
		gem.userChoiceHistory[eventName] = gem.userChoiceHistory[eventName][1:]
	}
}

// validateBasicEventFields validates name and description fields of a general event.
func validateBasicEventFields(event GeneralDialogEvent) error {
	if event.Name == "" {
		return fmt.Errorf("event name cannot be empty")
	}

	if event.Description == "" {
		return fmt.Errorf("event description cannot be empty")
	}

	return nil
}

// validateEventCategory validates the category field against allowed values.
func validateEventCategory(category string) error {
	validCategories := []string{"conversation", "roleplay", "game", "humor", "romance"}
	categoryValid := false
	for _, cat := range validCategories {
		if category == cat {
			categoryValid = true
			break
		}
	}
	if !categoryValid {
		return fmt.Errorf("invalid category '%s', must be one of: %v", category, validCategories)
	}

	return nil
}

// validateEventTrigger validates the trigger field is not empty.
func validateEventTrigger(trigger string) error {
	if trigger == "" {
		return fmt.Errorf("trigger cannot be empty")
	}

	return nil
}

// validateInteractiveChoices validates choices for interactive events.
func validateInteractiveChoices(event GeneralDialogEvent) error {
	if !event.Interactive {
		return nil
	}

	if len(event.Choices) == 0 {
		return fmt.Errorf("interactive events must have at least one choice")
	}

	for i, choice := range event.Choices {
		if err := validateEventChoice(choice, i); err != nil {
			return fmt.Errorf("choice %d: %w", i, err)
		}
	}

	return nil
}

// ValidateGeneralEvent validates a general event configuration
func ValidateGeneralEvent(event GeneralDialogEvent) error {
	if err := validateBasicEventFields(event); err != nil {
		return err
	}

	if err := validateEventCategory(event.Category); err != nil {
		return err
	}

	if err := validateEventTrigger(event.Trigger); err != nil {
		return err
	}

	if err := validateInteractiveChoices(event); err != nil {
		return err
	}

	return nil
}

// validateEventChoice validates an individual event choice
func validateEventChoice(choice EventChoice, index int) error {
	if choice.Text == "" {
		return fmt.Errorf("choice text cannot be empty")
	}

	// Validate effects (reuse existing stat validation logic)
	validStats := []string{"hunger", "happiness", "health", "energy", "affection", "trust", "intimacy", "jealousy"}
	for stat := range choice.Effects {
		statValid := false
		for _, validStat := range validStats {
			if stat == validStat {
				statValid = true
				break
			}
		}
		if !statValid {
			return fmt.Errorf("invalid stat '%s' in effects", stat)
		}
	}

	return nil
}
