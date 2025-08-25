package character

import (
	"os"
	"path/filepath"
	"testing"
)

// TestCharacterCardGameFeatures verifies game feature validation and loading
func TestCharacterCardGameFeatures(t *testing.T) {
	// Create temporary test character card with game features
	testCardWithGame := `{
		"name": "Game Pet",
		"description": "A pet with Tamagotchi-style game features",
		"animations": {
			"idle": "test_idle.gif",
			"talking": "test_talking.gif",
			"hungry": "test_hungry.gif",
			"happy": "test_happy.gif"
		},
		"dialogs": [
			{
				"trigger": "click",
				"responses": ["Hello game test!"],
				"animation": "talking",
				"cooldown": 5
			}
		],
		"behavior": {
			"idleTimeout": 30,
			"movementEnabled": true,
			"defaultSize": 128
		},
		"stats": {
			"hunger": {
				"initial": 100,
				"max": 100,
				"degradationRate": 1.0,
				"criticalThreshold": 20
			},
			"happiness": {
				"initial": 80,
				"max": 100,
				"degradationRate": 0.5,
				"criticalThreshold": 15
			}
		},
		"gameRules": {
			"statsDecayInterval": 60,
			"autoSaveInterval": 300,
			"criticalStateAnimationPriority": true,
			"deathEnabled": false,
			"evolutionEnabled": true,
			"moodBasedAnimations": true
		},
		"interactions": {
			"feed": {
				"triggers": ["rightclick"],
				"effects": {"hunger": 25, "happiness": 5},
				"animations": ["happy"],
				"responses": ["Yum! Thank you!", "That was delicious!"],
				"cooldown": 30,
				"requirements": {"hunger": {"max": 80}}
			}
		}
	}`

	// Create temporary file
	tmpDir, err := os.MkdirTemp("", "game_character_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create valid GIF files
	validGIF := []byte{71, 73, 70, 56, 57, 97, 1, 0, 1, 0, 128, 0, 0, 255, 255, 255, 0, 0, 0, 44, 0, 0, 0, 0, 1, 0, 1, 0, 0, 2, 2, 68, 1, 0, 59}

	gifFiles := []string{"test_idle.gif", "test_talking.gif", "test_hungry.gif", "test_happy.gif"}
	for _, filename := range gifFiles {
		err = os.WriteFile(filepath.Join(tmpDir, filename), validGIF, 0644)
		if err != nil {
			t.Fatalf("Failed to create %s: %v", filename, err)
		}
	}

	cardPath := filepath.Join(tmpDir, "character.json")
	err = os.WriteFile(cardPath, []byte(testCardWithGame), 0644)
	if err != nil {
		t.Fatalf("Failed to create character card: %v", err)
	}

	// Load and validate the character card
	card, err := LoadCard(cardPath)
	if err != nil {
		t.Fatalf("Failed to load character card with game features: %v", err)
	}

	// Verify game features are detected
	if !card.HasGameFeatures() {
		t.Error("Expected character card to have game features")
	}

	// Verify stats configuration
	if len(card.Stats) != 2 {
		t.Errorf("Expected 2 stats, got %d", len(card.Stats))
	}

	hungerStat, exists := card.Stats["hunger"]
	if !exists {
		t.Error("Expected hunger stat to exist")
	} else {
		if hungerStat.Initial != 100 {
			t.Errorf("Expected hunger initial to be 100, got %f", hungerStat.Initial)
		}
		if hungerStat.DegradationRate != 1.0 {
			t.Errorf("Expected hunger degradation rate to be 1.0, got %f", hungerStat.DegradationRate)
		}
	}

	// Verify game rules configuration
	if card.GameRules == nil {
		t.Error("Expected game rules to be set")
	} else {
		if card.GameRules.StatsDecayInterval != 60 {
			t.Errorf("Expected stats decay interval to be 60, got %d", card.GameRules.StatsDecayInterval)
		}
		if !card.GameRules.CriticalStateAnimationPriority {
			t.Error("Expected critical state animation priority to be true")
		}
	}

	// Verify interactions configuration
	if len(card.Interactions) != 1 {
		t.Errorf("Expected 1 interaction, got %d", len(card.Interactions))
	}

	feedInteraction, exists := card.Interactions["feed"]
	if !exists {
		t.Error("Expected feed interaction to exist")
	} else {
		if len(feedInteraction.Triggers) != 1 || feedInteraction.Triggers[0] != "rightclick" {
			t.Errorf("Expected feed interaction to have rightclick trigger, got %v", feedInteraction.Triggers)
		}
		if feedInteraction.Effects["hunger"] != 25 {
			t.Errorf("Expected feed to increase hunger by 25, got %f", feedInteraction.Effects["hunger"])
		}
	}
}

// TestGameFeatureValidationErrors verifies proper error handling for invalid game configurations
func TestGameFeatureValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		cardJSON    string
		expectedErr string
	}{
		{
			name: "invalid stat max",
			cardJSON: `{
				"name": "Test Pet",
				"description": "Test description",
				"animations": {"idle": "test.gif", "talking": "test.gif"},
				"dialogs": [{"trigger": "click", "responses": ["hi"], "animation": "talking"}],
				"behavior": {"idleTimeout": 30, "defaultSize": 128},
				"stats": {
					"hunger": {"initial": 100, "max": 0, "degradationRate": 1.0, "criticalThreshold": 20}
				}
			}`,
			expectedErr: "max value must be positive",
		},
		{
			name: "invalid initial value",
			cardJSON: `{
				"name": "Test Pet",
				"description": "Test description",
				"animations": {"idle": "test.gif", "talking": "test.gif"},
				"dialogs": [{"trigger": "click", "responses": ["hi"], "animation": "talking"}],
				"behavior": {"idleTimeout": 30, "defaultSize": 128},
				"stats": {
					"hunger": {"initial": 150, "max": 100, "degradationRate": 1.0, "criticalThreshold": 20}
				}
			}`,
			expectedErr: "initial value (150.000000) must be between 0 and max (100.000000)",
		},
		{
			name: "invalid degradation rate",
			cardJSON: `{
				"name": "Test Pet",
				"description": "Test description",
				"animations": {"idle": "test.gif", "talking": "test.gif"},
				"dialogs": [{"trigger": "click", "responses": ["hi"], "animation": "talking"}],
				"behavior": {"idleTimeout": 30, "defaultSize": 128},
				"stats": {
					"hunger": {"initial": 100, "max": 100, "degradationRate": -1.0, "criticalThreshold": 20}
				}
			}`,
			expectedErr: "degradation rate cannot be negative",
		},
		{
			name: "invalid game rules decay interval",
			cardJSON: `{
				"name": "Test Pet",
				"description": "Test description",
				"animations": {"idle": "test.gif", "talking": "test.gif"},
				"dialogs": [{"trigger": "click", "responses": ["hi"], "animation": "talking"}],
				"behavior": {"idleTimeout": 30, "defaultSize": 128},
				"gameRules": {"statsDecayInterval": 5}
			}`,
			expectedErr: "stats decay interval must be 10-3600 seconds",
		},
		{
			name: "invalid interaction trigger",
			cardJSON: `{
				"name": "Test Pet",
				"description": "Test description",
				"animations": {"idle": "test.gif", "talking": "test.gif"},
				"dialogs": [{"trigger": "click", "responses": ["hi"], "animation": "talking"}],
				"behavior": {"idleTimeout": 30, "defaultSize": 128},
				"interactions": {
					"feed": {
						"triggers": ["invalid_trigger"],
						"effects": {"hunger": 25},
						"animations": ["talking"],
						"responses": ["yum"]
					}
				}
			}`,
			expectedErr: "invalid trigger 'invalid_trigger'",
		},
		{
			name: "invalid interaction animation reference",
			cardJSON: `{
				"name": "Test Pet",
				"description": "Test description",
				"animations": {"idle": "test.gif", "talking": "test.gif"},
				"dialogs": [{"trigger": "click", "responses": ["hi"], "animation": "talking"}],
				"behavior": {"idleTimeout": 30, "defaultSize": 128},
				"interactions": {
					"feed": {
						"triggers": ["rightclick"],
						"effects": {"hunger": 25},
						"animations": ["nonexistent"],
						"responses": ["yum"]
					}
				}
			}`,
			expectedErr: "animation 'nonexistent' not found in animations map",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpDir, err := os.MkdirTemp("", "validation_test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			// Create test GIF file
			validGIF := []byte{71, 73, 70, 56, 57, 97, 1, 0, 1, 0, 128, 0, 0, 255, 255, 255, 0, 0, 0, 44, 0, 0, 0, 0, 1, 0, 1, 0, 0, 2, 2, 68, 1, 0, 59}
			err = os.WriteFile(filepath.Join(tmpDir, "test.gif"), validGIF, 0644)
			if err != nil {
				t.Fatalf("Failed to create test.gif: %v", err)
			}

			cardPath := filepath.Join(tmpDir, "character.json")
			err = os.WriteFile(cardPath, []byte(tt.cardJSON), 0644)
			if err != nil {
				t.Fatalf("Failed to create character card: %v", err)
			}

			// Attempt to load the card - should fail with expected error
			_, err = LoadCard(cardPath)
			if err == nil {
				t.Errorf("Expected LoadCard to fail with error containing '%s', but it succeeded", tt.expectedErr)
				return
			}

			if !contains(err.Error(), tt.expectedErr) {
				t.Errorf("Expected error to contain '%s', got: %v", tt.expectedErr, err)
			}
		})
	}
}

// TestHasGameFeatures verifies the HasGameFeatures detection
func TestHasGameFeatures(t *testing.T) {
	// Test card without game features
	cardWithoutGame := &CharacterCard{
		Name:        "Regular Pet",
		Description: "A regular desktop pet",
		Animations:  map[string]string{"idle": "idle.gif"},
	}

	if cardWithoutGame.HasGameFeatures() {
		t.Error("Expected card without stats to not have game features")
	}

	// Test card with game features
	cardWithGame := &CharacterCard{
		Name:        "Game Pet",
		Description: "A game-enabled desktop pet",
		Animations:  map[string]string{"idle": "idle.gif"},
		Stats: map[string]StatConfig{
			"hunger": {Initial: 100, Max: 100, DegradationRate: 1.0, CriticalThreshold: 20},
		},
	}

	if !cardWithGame.HasGameFeatures() {
		t.Error("Expected card with stats to have game features")
	}

	// Test card with empty stats map
	cardWithEmptyStats := &CharacterCard{
		Name:        "Empty Stats Pet",
		Description: "A pet with empty stats",
		Animations:  map[string]string{"idle": "idle.gif"},
		Stats:       map[string]StatConfig{},
	}

	if cardWithEmptyStats.HasGameFeatures() {
		t.Error("Expected card with empty stats to not have game features")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			(len(s) > len(substr) && findSubstring(s, substr)))))
}

// Simple substring search helper
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
