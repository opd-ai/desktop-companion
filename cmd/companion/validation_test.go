package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/opd-ai/desktop-companion/lib/character"
)

// test_animation_file_validation_missing reproduces the bug where LoadCard doesn't verify animation files exist
func TestAnimationFileValidationMissing(t *testing.T) {
	t.Log("Bug reproduction: Character file validation doesn't check animation file existence")
	t.Log("Description: LoadCard validates card format but doesn't verify referenced animation files exist")

	// Create a temporary character card with non-existent animation files
	tempDir := t.TempDir()
	characterPath := filepath.Join(tempDir, "character.json")

	// Character card that references non-existent animation files
	cardContent := `{
  "name": "Test Character",
  "version": "1.0.0", 
  "description": "Test character for bug reproduction",
  "size": 64,
  "movementEnabled": true,
  "animations": {
    "idle": "nonexistent_idle.gif",
    "talking": "nonexistent_talking.gif"
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
    "idleTimeout": 30,
    "movementEnabled": true,
    "defaultSize": 128
  }
}`

	err := os.WriteFile(characterPath, []byte(cardContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test character file: %v", err)
	}

	// This should now fail with the fix - demonstrating the bug is resolved
	_, err = character.LoadCard(characterPath)
	if err == nil {
		t.Fatalf("LoadCard should have failed due to missing animation files, but it succeeded")
	}

	// Verify the error is about missing animation files
	if !strings.Contains(err.Error(), "animation file") && !strings.Contains(err.Error(), "not found") {
		t.Fatalf("Expected error about missing animation files, got: %v", err)
	}

	t.Log("✅ FIX VERIFIED: LoadCard now correctly fails when animation files don't exist")
	t.Log("Expected behavior: LoadCard should validate that animation files exist")
	t.Log("Actual behavior: LoadCard now checks file existence during validation")
	t.Log("Impact: Users now get clear validation errors during card loading instead of confusing runtime errors")
} // test_animation_validation_requirements documents what the validation should check
func TestAnimationValidationRequirements(t *testing.T) {
	t.Log("Animation validation requirements documentation")

	t.Log("Requirement: All animation paths should reference existing files")
	t.Log("Requirement: Files should be readable")
	t.Log("Requirement: Files should be valid GIF format")
	t.Log("Requirement: Validation should happen during LoadCard, not character creation")

	t.Log("Current implementation: Only checks .gif extension")
	t.Log("Fix needed: Add file existence and readability checks to validateAnimationPaths")
}

// test_animation_validation_with_valid_files verifies the fix works with existing files
func TestAnimationValidationWithValidFiles(t *testing.T) {
	t.Log("VALIDATION FIX VERIFICATION: Testing that LoadCard works with valid animation files")

	// Create a temporary character card with valid animation files
	tempDir := t.TempDir()
	characterPath := filepath.Join(tempDir, "character.json")

	// Create valid GIF files
	validGIF := []byte{71, 73, 70, 56, 57, 97, 1, 0, 1, 0, 128, 0, 0, 255, 255, 255, 0, 0, 0, 44, 0, 0, 0, 0, 1, 0, 1, 0, 0, 2, 2, 68, 1, 0, 59}

	err := os.WriteFile(filepath.Join(tempDir, "idle.gif"), validGIF, 0644)
	if err != nil {
		t.Fatalf("Failed to create idle.gif: %v", err)
	}

	err = os.WriteFile(filepath.Join(tempDir, "talking.gif"), validGIF, 0644)
	if err != nil {
		t.Fatalf("Failed to create talking.gif: %v", err)
	}

	// Character card that references existing animation files
	cardContent := `{
  "name": "Valid Character",
  "version": "1.0.0", 
  "description": "Test character with valid files",
  "size": 64,
  "movementEnabled": true,
  "animations": {
    "idle": "idle.gif",
    "talking": "talking.gif"
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
    "idleTimeout": 30,
    "movementEnabled": true,
    "defaultSize": 128
  }
}`

	err = os.WriteFile(characterPath, []byte(cardContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test character file: %v", err)
	}

	// This should succeed with valid files
	card, err := character.LoadCard(characterPath)
	if err != nil {
		t.Fatalf("LoadCard failed with valid files: %v", err)
	}

	// Verify the card was loaded successfully
	if card.Name != "Valid Character" {
		t.Fatalf("Expected character name 'Valid Character', got %q", card.Name)
	}

	t.Log("✅ FIX VERIFIED: LoadCard succeeds with valid animation files")
	t.Log("✅ FIX VERIFIED: File existence validation works correctly")
}
