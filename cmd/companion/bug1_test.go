package main

import (
	"os"
	"path/filepath"
	"testing"

	"desktop-companion/internal/character"
)

// TestBug1AnimationLoadingGracefulDegradation is a regression test for
// the critical bug where character creation failed entirely if any required
// animation files were missing. Now it should succeed with graceful degradation.
func TestBug1AnimationLoadingGracefulDegradation(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir := t.TempDir()

	// Create character card directory
	cardDir := filepath.Join(tempDir, "test_character")
	if err := os.MkdirAll(cardDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create animations directory and invalid animation files
	animDir := filepath.Join(cardDir, "animations")
	if err := os.MkdirAll(animDir, 0755); err != nil {
		t.Fatalf("Failed to create animations directory: %v", err)
	}

	// Create animation files that exist but are empty (will fail to load as GIFs)
	idlePath := filepath.Join(animDir, "idle.gif")
	talkingPath := filepath.Join(animDir, "talking.gif")

	// Create invalid GIF files (empty files that will fail to load)
	if err := os.WriteFile(idlePath, []byte("invalid gif data"), 0644); err != nil {
		t.Fatalf("Failed to create idle animation file: %v", err)
	}
	if err := os.WriteFile(talkingPath, []byte("invalid gif data"), 0644); err != nil {
		t.Fatalf("Failed to create talking animation file: %v", err)
	}

	// Create a basic character card file
	cardPath := filepath.Join(cardDir, "character.json")
	cardContent := `{
		"name": "Test Character",
		"description": "A test character for bug reproduction",
		"animations": {
			"idle": "animations/idle.gif",
			"talking": "animations/talking.gif"
		},
		"dialogs": [
			{
				"trigger": "click",
				"responses": ["Hello!"],
				"animation": "talking",
				"cooldown": 5
			}
		],
		"behavior": {
			"defaultSize": 200,
			"idleTimeout": 30,
			"actionsEnabled": true
		}
	}`

	if err := os.WriteFile(cardPath, []byte(cardContent), 0644); err != nil {
		t.Fatalf("Failed to create character card: %v", err)
	}

	// Load the character card
	card, err := character.LoadCard(cardPath)
	if err != nil {
		t.Fatalf("Failed to load character card: %v", err)
	}

	// Test character creation - this should now succeed with graceful degradation
	char, err := character.New(card, cardDir)

	// FIXED: Character creation should succeed with graceful degradation
	// even when all animations fail to load
	if err != nil {
		t.Errorf("Character creation should succeed with graceful degradation, but failed: %v", err)
	} else {
		// After the fix, this should succeed with graceful degradation
		t.Logf("Character creation succeeded with graceful degradation")
		if char == nil {
			t.Errorf("Character should not be nil after successful creation")
		}
	}
}
