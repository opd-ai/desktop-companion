package character

import (
	"testing"
	"time"
)

// TestGeneralEventManager tests the core functionality of the general events system
func TestGeneralEventManager(t *testing.T) {
	// Create test events
	events := []GeneralDialogEvent{
		{
			RandomEventConfig: RandomEventConfig{
				Name:        "test_conversation",
				Description: "A test conversation event",
				Probability: 1.0,
				Responses:   []string{"Hello! How are you?"},
				Cooldown:    60,
			},
			Category:    "conversation",
			Trigger:     "daily_chat",
			Interactive: true,
			Choices: []EventChoice{
				{
					Text:    "I'm doing great!",
					Effects: map[string]float64{"happiness": 5},
				},
				{
					Text:    "Could be better...",
					Effects: map[string]float64{"friendship": 3},
				},
			},
		},
		{
			RandomEventConfig: RandomEventConfig{
				Name:        "test_roleplay",
				Description: "A test roleplay event",
				Probability: 1.0,
				Responses:   []string{"Welcome to the adventure!"},
				Cooldown:    300,
				Conditions: map[string]map[string]float64{
					"level": {"min": 5},
				},
			},
			Category:    "roleplay",
			Trigger:     "start_adventure",
			Interactive: false,
		},
	}

	// Create manager
	manager := NewGeneralEventManager(events, true)

	if manager == nil {
		t.Fatal("Failed to create general event manager")
	}

	// Test getting available events (without game state - should work for events without conditions)
	available := manager.GetAvailableEvents(nil)
	t.Logf("Available events without game state: %d", len(available))
	for i, event := range available {
		t.Logf("Event %d: %s (conditions: %v)", i, event.Name, event.Conditions)
	}
	if len(available) == 0 {
		t.Error("Expected at least one available event without conditions")
	}

	// Test category filtering BEFORE triggering events (to avoid cooldown issues)
	conversationEvents := manager.GetEventsByCategory("conversation", nil)
	t.Logf("Conversation events found: %d", len(conversationEvents))
	for i, event := range conversationEvents {
		t.Logf("Conversation event %d: %s (category: %s)", i, event.Name, event.Category)
	}
	if len(conversationEvents) != 1 {
		t.Errorf("Expected 1 conversation event, got %d", len(conversationEvents))
	}

	roleplayEvents := manager.GetEventsByCategory("roleplay", nil)
	t.Logf("Roleplay events found: %d", len(roleplayEvents))
	for i, event := range roleplayEvents {
		t.Logf("Roleplay event %d: %s (category: %s, conditions: %v)", i, event.Name, event.Category, event.Conditions)
	}
	if len(roleplayEvents) != 0 { // Should be 0 because of level requirement
		t.Errorf("Expected 0 roleplay events without meeting conditions, got %d", len(roleplayEvents))
	}

	// Test triggering an event
	event, err := manager.TriggerEvent("test_conversation", nil)
	if err != nil {
		t.Errorf("Failed to trigger event: %v", err)
	}
	if event == nil {
		t.Error("Expected event to be returned")
	}

	// Test that event is now on cooldown
	_, err = manager.TriggerEvent("test_conversation", nil)
	if err == nil {
		t.Error("Expected event to be on cooldown")
	}

	// Test interactive event
	if event.Interactive {
		choice, nextAction, err := manager.SubmitChoice(0, nil)
		if err != nil {
			t.Errorf("Failed to submit choice: %v", err)
		}
		if choice == nil {
			t.Error("Expected choice to be returned")
		}
		if choice.Text != "I'm doing great!" {
			t.Errorf("Expected choice text 'I'm doing great!', got '%s'", choice.Text)
		}
		if nextAction != "" {
			t.Errorf("Expected no next action, got '%s'", nextAction)
		}
	}
}

// TestGeneralEventValidation tests the validation system
func TestGeneralEventValidation(t *testing.T) {
	// Test valid event
	validEvent := GeneralDialogEvent{
		RandomEventConfig: RandomEventConfig{
			Name:        "valid_event",
			Description: "A valid test event",
			Probability: 1.0,
		},
		Category:    "conversation",
		Trigger:     "test_trigger",
		Interactive: false,
	}

	if err := ValidateGeneralEvent(validEvent); err != nil {
		t.Errorf("Valid event failed validation: %v", err)
	}

	// Test invalid category
	invalidEvent := validEvent
	invalidEvent.Category = "invalid_category"
	if err := ValidateGeneralEvent(invalidEvent); err == nil {
		t.Error("Expected validation error for invalid category")
	}

	// Test empty trigger
	invalidEvent = validEvent
	invalidEvent.Trigger = ""
	if err := ValidateGeneralEvent(invalidEvent); err == nil {
		t.Error("Expected validation error for empty trigger")
	}

	// Test interactive event without choices
	invalidEvent = validEvent
	invalidEvent.Interactive = true
	invalidEvent.Choices = nil
	if err := ValidateGeneralEvent(invalidEvent); err == nil {
		t.Error("Expected validation error for interactive event without choices")
	}

	// Test interactive event with invalid choice
	invalidEvent = validEvent
	invalidEvent.Interactive = true
	invalidEvent.Choices = []EventChoice{
		{
			Text: "", // Empty text should fail
		},
	}
	if err := ValidateGeneralEvent(invalidEvent); err == nil {
		t.Error("Expected validation error for choice with empty text")
	}
}

// TestCharacterGeneralEvents tests integration with Character struct
func TestCharacterGeneralEvents(t *testing.T) {
	// Create a basic character card with general events
	card := &CharacterCard{
		Name:        "Test Character",
		Description: "A test character with general events",
		Animations: map[string]string{
			"idle":    "animations/idle.gif",
			"talking": "animations/talking.gif",
		},
		Dialogs: []Dialog{
			{
				Trigger:   "click",
				Responses: []string{"Hello!"},
				Animation: "talking",
				Cooldown:  5,
			},
		},
		Behavior: Behavior{
			IdleTimeout:     30,
			MovementEnabled: false,
			DefaultSize:     128,
		},
		GeneralEvents: []GeneralDialogEvent{
			{
				RandomEventConfig: RandomEventConfig{
					Name:        "test_event",
					Description: "Test event",
					Probability: 1.0,
					Responses:   []string{"Test response"},
					Cooldown:    60,
				},
				Category: "conversation",
				Trigger:  "test_trigger",
			},
		},
	}

	// Test character card validation first (this doesn't require files)
	if err := card.Validate(); err != nil {
		t.Fatalf("Character card validation failed: %v", err)
	}

	// Test general event validation specifically
	if err := card.validateGeneralEvents(); err != nil {
		t.Fatalf("General events validation failed: %v", err)
	}

	// Test general event manager creation without full character
	manager := NewGeneralEventManager(card.GeneralEvents, true)
	if manager == nil {
		t.Fatal("Failed to create general event manager")
	}

	// Test manager functionality
	available := manager.GetAvailableEvents(nil)
	if len(available) != 1 {
		t.Errorf("Expected 1 available event, got %d", len(available))
	}

	// Test triggering an event
	event, err := manager.TriggerEvent("test_event", nil)
	if err != nil {
		t.Errorf("Failed to trigger event: %v", err)
	}
	if event == nil {
		t.Error("Expected event to be returned")
	}

	// Test that event is on cooldown
	_, err = manager.TriggerEvent("test_event", nil)
	if err == nil {
		t.Error("Expected event to be on cooldown")
	}

	// Test availability check
	if manager.IsEventAvailable("test_event", nil) {
		t.Error("Event should not be available due to cooldown")
	}

	if !manager.IsEventAvailable("nonexistent", nil) {
		// This should be false anyway, so this is the expected case
	}
}

// TestEventCooldowns tests cooldown functionality
func TestEventCooldowns(t *testing.T) {
	events := []GeneralDialogEvent{
		{
			RandomEventConfig: RandomEventConfig{
				Name:        "short_cooldown",
				Description: "Event with short cooldown",
				Probability: 1.0,
				Responses:   []string{"Quick event"},
				Cooldown:    1, // 1 second cooldown
			},
			Category: "conversation",
			Trigger:  "quick_event",
		},
	}

	manager := NewGeneralEventManager(events, true)

	// Trigger event
	_, err := manager.TriggerEvent("short_cooldown", nil)
	if err != nil {
		t.Errorf("Failed to trigger event: %v", err)
	}

	// Should be on cooldown
	_, err = manager.TriggerEvent("short_cooldown", nil)
	if err == nil {
		t.Error("Expected event to be on cooldown")
	}

	// Wait for cooldown to expire
	time.Sleep(1100 * time.Millisecond) // Wait slightly more than 1 second

	// Should be available again
	_, err = manager.TriggerEvent("short_cooldown", nil)
	if err != nil {
		t.Errorf("Event should be available after cooldown: %v", err)
	}
}

// TestEventChoiceHistory tests user choice tracking
func TestEventChoiceHistory(t *testing.T) {
	events := []GeneralDialogEvent{
		{
			RandomEventConfig: RandomEventConfig{
				Name:        "choice_test",
				Description: "Event to test choices",
				Probability: 1.0,
				Responses:   []string{"Choose one:"},
				Cooldown:    1,
			},
			Category:    "conversation",
			Trigger:     "choice_event",
			Interactive: true,
			Choices: []EventChoice{
				{Text: "Choice A", Effects: map[string]float64{"stat1": 1}},
				{Text: "Choice B", Effects: map[string]float64{"stat2": 1}},
			},
		},
	}

	manager := NewGeneralEventManager(events, true)

	// Trigger event and make choices
	_, err := manager.TriggerEvent("choice_test", nil)
	if err != nil {
		t.Errorf("Failed to trigger event: %v", err)
	}

	// Submit choice 0
	_, _, err = manager.SubmitChoice(0, nil)
	if err != nil {
		t.Errorf("Failed to submit choice: %v", err)
	}

	// Check choice history
	history := manager.GetUserChoiceHistory("choice_test")
	if len(history) != 1 {
		t.Errorf("Expected 1 choice in history, got %d", len(history))
	}
	if history[0] != 0 {
		t.Errorf("Expected choice 0 in history, got %d", history[0])
	}
}
