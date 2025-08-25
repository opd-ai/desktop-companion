package character

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Helper function to get a valid character card for testing
func getValidCharacterCard() CharacterCard {
	return CharacterCard{
		Name:        "Test Character",
		Description: "A test character for validation testing",
		Animations: map[string]string{
			"idle":    "idle.gif",
			"talking": "talking.gif",
			"happy":   "happy.gif",
			"sad":     "sad.gif",
		},
		Dialogs: []Dialog{
			{
				Trigger:   "click",
				Responses: []string{"Hello!", "Hi there!"},
				Animation: "talking",
				Cooldown:  5,
			},
		},
		Behavior: Behavior{
			IdleTimeout: 30,
			DefaultSize: 128,
		},
	}
}

// Helper function to check if string contains substring
func containsSubstring(str, substr string) bool {
	return strings.Contains(str, substr)
}

// TestLoadCard verifies character card loading and validation
func TestLoadCard(t *testing.T) {
	// Create temporary test character card
	testCard := `{
		"name": "Test Character",
		"description": "A test character for unit testing",
		"animations": {
			"idle": "test_idle.gif",
			"talking": "test_talking.gif"
		},
		"dialogs": [
			{
				"trigger": "click",
				"responses": ["Hello test!"],
				"animation": "talking",
				"cooldown": 5
			}
		],
		"behavior": {
			"idleTimeout": 30,
			"movementEnabled": true,
			"defaultSize": 128
		}
	}`

	// Create temporary file
	tmpDir, err := os.MkdirTemp("", "character_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create valid GIF files that the character card references
	validGIF := []byte{71, 73, 70, 56, 57, 97, 1, 0, 1, 0, 128, 0, 0, 255, 255, 255, 0, 0, 0, 44, 0, 0, 0, 0, 1, 0, 1, 0, 0, 2, 2, 68, 1, 0, 59}

	err = os.WriteFile(filepath.Join(tmpDir, "test_idle.gif"), validGIF, 0644)
	if err != nil {
		t.Fatalf("Failed to create test_idle.gif: %v", err)
	}

	err = os.WriteFile(filepath.Join(tmpDir, "test_talking.gif"), validGIF, 0644)
	if err != nil {
		t.Fatalf("Failed to create test_talking.gif: %v", err)
	}

	cardPath := filepath.Join(tmpDir, "character.json")
	err = os.WriteFile(cardPath, []byte(testCard), 0644)
	if err != nil {
		t.Fatalf("Failed to write test card: %v", err)
	}

	// Test loading
	card, err := LoadCard(cardPath)
	if err != nil {
		t.Fatalf("Failed to load character card: %v", err)
	}

	// Verify card contents
	if card.Name != "Test Character" {
		t.Errorf("Expected name 'Test Character', got '%s'", card.Name)
	}

	if card.Behavior.DefaultSize != 128 {
		t.Errorf("Expected default size 128, got %d", card.Behavior.DefaultSize)
	}

	if len(card.Dialogs) != 1 {
		t.Errorf("Expected 1 dialog, got %d", len(card.Dialogs))
	}
}

// TestCharacterCardValidation tests validation logic
func TestCharacterCardValidation(t *testing.T) {
	tests := []struct {
		name        string
		card        CharacterCard
		expectError bool
	}{
		{
			name: "valid card",
			card: CharacterCard{
				Name:        "Valid Character",
				Description: "A valid test character",
				Animations: map[string]string{
					"idle":    "idle.gif",
					"talking": "talking.gif",
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
					IdleTimeout: 30,
					DefaultSize: 128,
				},
			},
			expectError: false,
		},
		{
			name: "missing name",
			card: CharacterCard{
				Description: "Character without name",
				Animations: map[string]string{
					"idle":    "idle.gif",
					"talking": "talking.gif",
				},
				Dialogs: []Dialog{
					{
						Trigger:   "click",
						Responses: []string{"Hello!"},
						Animation: "talking",
					},
				},
				Behavior: Behavior{
					IdleTimeout: 30,
					DefaultSize: 128,
				},
			},
			expectError: true,
		},
		{
			name: "missing required animation",
			card: CharacterCard{
				Name:        "Test Character",
				Description: "Missing required animation",
				Animations: map[string]string{
					"idle": "idle.gif",
					// Missing "talking" animation
				},
				Dialogs: []Dialog{
					{
						Trigger:   "click",
						Responses: []string{"Hello!"},
						Animation: "idle",
					},
				},
				Behavior: Behavior{
					IdleTimeout: 30,
					DefaultSize: 128,
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.card.Validate()
			if tt.expectError && err == nil {
				t.Error("Expected validation error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}
		})
	}
}

// TestDialogCooldown tests dialog cooldown functionality
func TestDialogCooldown(t *testing.T) {
	dialog := Dialog{
		Trigger:   "click",
		Responses: []string{"Test response"},
		Animation: "talking",
		Cooldown:  2, // 2 second cooldown
	}

	// Should be able to trigger immediately
	if !dialog.CanTrigger(time.Time{}) {
		t.Error("Should be able to trigger dialog initially")
	}

	// Should not be able to trigger within cooldown period
	recentTime := time.Now().Add(-1 * time.Second) // 1 second ago
	if dialog.CanTrigger(recentTime) {
		t.Error("Should not be able to trigger dialog within cooldown")
	}

	// Should be able to trigger after cooldown period
	oldTime := time.Now().Add(-3 * time.Second) // 3 seconds ago
	if !dialog.CanTrigger(oldTime) {
		t.Error("Should be able to trigger dialog after cooldown")
	}
}

// TestDialogRandomResponse tests response randomization
func TestDialogRandomResponse(t *testing.T) {
	dialog := Dialog{
		Responses: []string{"Response 1", "Response 2", "Response 3"},
	}

	// Get multiple responses to check randomization
	responses := make(map[string]bool)
	for i := 0; i < 50; i++ {
		response := dialog.GetRandomResponse()
		if response == "" {
			t.Error("Got empty response")
		}
		responses[response] = true
	}

	// Should have gotten at least some variety (not just one response)
	if len(responses) < 2 {
		t.Error("Expected some variety in random responses")
	}
}

// BenchmarkCharacterCardValidation benchmarks the validation performance
func BenchmarkCharacterCardValidation(b *testing.B) {
	card := CharacterCard{
		Name:        "Benchmark Character",
		Description: "A character for benchmarking validation performance",
		Animations: map[string]string{
			"idle":    "idle.gif",
			"talking": "talking.gif",
			"happy":   "happy.gif",
			"sad":     "sad.gif",
		},
		Dialogs: []Dialog{
			{
				Trigger:   "click",
				Responses: []string{"Hello!", "Hi there!", "How are you?"},
				Animation: "talking",
				Cooldown:  5,
			},
			{
				Trigger:   "rightclick",
				Responses: []string{"Right click!", "That tickles!"},
				Animation: "happy",
				Cooldown:  8,
			},
		},
		Behavior: Behavior{
			IdleTimeout: 30,
			DefaultSize: 128,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = card.Validate()
	}
}

// TestCharacterCardRandomEventsValidation tests validation of random events configuration
func TestCharacterCardRandomEventsValidation(t *testing.T) {
	baseCard := getValidCharacterCard()

	tests := []struct {
		name          string
		randomEvents  []RandomEventConfig
		expectError   bool
		errorContains string
	}{
		{
			name:         "no random events (valid)",
			randomEvents: []RandomEventConfig{},
			expectError:  false,
		},
		{
			name: "valid random event",
			randomEvents: []RandomEventConfig{
				{
					Name:        "good_event",
					Description: "A good event",
					Probability: 0.1,
					Effects:     map[string]float64{"hunger": 10},
					Animations:  []string{"happy"},
					Responses:   []string{"Yay!"},
					Cooldown:    60,
					Duration:    0,
				},
			},
			expectError: false,
		},
		{
			name: "empty event name",
			randomEvents: []RandomEventConfig{
				{
					Name:        "",
					Description: "No name event",
					Probability: 0.1,
				},
			},
			expectError:   true,
			errorContains: "name cannot be empty",
		},
		{
			name: "empty description",
			randomEvents: []RandomEventConfig{
				{
					Name:        "test_event",
					Description: "",
					Probability: 0.1,
				},
			},
			expectError:   true,
			errorContains: "description cannot be empty",
		},
		{
			name: "invalid probability negative",
			randomEvents: []RandomEventConfig{
				{
					Name:        "bad_prob_event",
					Description: "Bad probability",
					Probability: -0.1,
				},
			},
			expectError:   true,
			errorContains: "probability must be 0.0-1.0",
		},
		{
			name: "invalid probability too high",
			randomEvents: []RandomEventConfig{
				{
					Name:        "bad_prob_event",
					Description: "Bad probability",
					Probability: 1.5,
				},
			},
			expectError:   true,
			errorContains: "probability must be 0.0-1.0",
		},
		{
			name: "invalid cooldown negative",
			randomEvents: []RandomEventConfig{
				{
					Name:        "bad_cooldown_event",
					Description: "Bad cooldown",
					Probability: 0.1,
					Cooldown:    -1,
				},
			},
			expectError:   true,
			errorContains: "cooldown must be 0-86400 seconds",
		},
		{
			name: "invalid cooldown too high",
			randomEvents: []RandomEventConfig{
				{
					Name:        "bad_cooldown_event",
					Description: "Bad cooldown",
					Probability: 0.1,
					Cooldown:    90000, // > 86400 (24 hours)
				},
			},
			expectError:   true,
			errorContains: "cooldown must be 0-86400 seconds",
		},
		{
			name: "invalid duration negative",
			randomEvents: []RandomEventConfig{
				{
					Name:        "bad_duration_event",
					Description: "Bad duration",
					Probability: 0.1,
					Duration:    -1,
				},
			},
			expectError:   true,
			errorContains: "duration must be 0-3600 seconds",
		},
		{
			name: "invalid duration too high",
			randomEvents: []RandomEventConfig{
				{
					Name:        "bad_duration_event",
					Description: "Bad duration",
					Probability: 0.1,
					Duration:    4000, // > 3600 (1 hour)
				},
			},
			expectError:   true,
			errorContains: "duration must be 0-3600 seconds",
		},
		{
			name: "non-existent animation",
			randomEvents: []RandomEventConfig{
				{
					Name:        "bad_animation_event",
					Description: "Bad animation",
					Probability: 0.1,
					Animations:  []string{"nonexistent"},
				},
			},
			expectError:   true,
			errorContains: "animation 'nonexistent' not found",
		},
		{
			name: "too many responses",
			randomEvents: []RandomEventConfig{
				{
					Name:        "too_many_responses_event",
					Description: "Too many responses",
					Probability: 0.1,
					Responses:   []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11"},
				},
			},
			expectError:   true,
			errorContains: "must have 0-10 responses",
		},
		{
			name: "effects reference non-existent stat (no stats defined)",
			randomEvents: []RandomEventConfig{
				{
					Name:        "bad_effect_event",
					Description: "Bad effect",
					Probability: 0.1,
					Effects:     map[string]float64{"nonexistent": 10},
				},
			},
			expectError:   true,
			errorContains: "event effects reference stat 'nonexistent' which is not defined",
		},
		{
			name: "conditions reference non-existent stat (no stats defined)",
			randomEvents: []RandomEventConfig{
				{
					Name:        "bad_condition_event",
					Description: "Bad condition",
					Probability: 0.1,
					Conditions:  map[string]map[string]float64{"nonexistent": {"min": 50}},
				},
			},
			expectError:   true,
			errorContains: "event conditions reference stat 'nonexistent' which is not defined",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := baseCard
			card.RandomEvents = tt.randomEvents

			err := card.Validate()

			if tt.expectError {
				if err == nil {
					t.Error("expected validation error, but got none")
				} else if !containsSubstring(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain '%s', got: %v", tt.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no validation error, got: %v", err)
				}
			}
		})
	}
}

// TestCharacterCardRandomEventsWithStats tests random events validation when stats are defined
func TestCharacterCardRandomEventsWithStats(t *testing.T) {
	baseCard := getValidCharacterCard()
	baseCard.Stats = map[string]StatConfig{
		"hunger": {
			Initial:           100,
			Max:               100,
			DegradationRate:   1.0,
			CriticalThreshold: 20,
		},
		"happiness": {
			Initial:           100,
			Max:               100,
			DegradationRate:   0.8,
			CriticalThreshold: 15,
		},
	}

	tests := []struct {
		name          string
		randomEvents  []RandomEventConfig
		expectError   bool
		errorContains string
	}{
		{
			name: "valid event with existing stats",
			randomEvents: []RandomEventConfig{
				{
					Name:        "stat_event",
					Description: "Affects stats",
					Probability: 0.1,
					Effects:     map[string]float64{"hunger": 10, "happiness": 5},
					Conditions:  map[string]map[string]float64{"hunger": {"max": 50}},
					Cooldown:    60,
				},
			},
			expectError: false,
		},
		{
			name: "effect references non-existent stat (with stats defined)",
			randomEvents: []RandomEventConfig{
				{
					Name:        "bad_effect_event",
					Description: "Bad effect",
					Probability: 0.1,
					Effects:     map[string]float64{"energy": 10}, // energy not defined
				},
			},
			expectError:   true,
			errorContains: "event effects reference stat 'energy' which is not defined",
		},
		{
			name: "condition references non-existent stat (with stats defined)",
			randomEvents: []RandomEventConfig{
				{
					Name:        "bad_condition_event",
					Description: "Bad condition",
					Probability: 0.1,
					Conditions:  map[string]map[string]float64{"energy": {"min": 50}}, // energy not defined
				},
			},
			expectError:   true,
			errorContains: "event conditions reference stat 'energy' which is not defined",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := baseCard
			card.RandomEvents = tt.randomEvents

			err := card.Validate()

			if tt.expectError {
				if err == nil {
					t.Error("expected validation error, but got none")
				} else if !containsSubstring(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain '%s', got: %v", tt.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no validation error, got: %v", err)
				}
			}
		})
	}
}
