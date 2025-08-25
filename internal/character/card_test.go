package character

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

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
