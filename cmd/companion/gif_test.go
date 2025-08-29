package main

import (
	"os"
	"path/filepath"
	"testing"

	"desktop-companion/internal/character"
)

// TestBug2InvalidGIFData reproduces the original bug with malformed GIF data
func TestBug2InvalidGIFData(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "gif_bug_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Use the same malformed GIF data that was originally in integration test
	malformedGIF := []byte("GIF89a\x01\x00\x01\x00\x00\x00\x00\x21\xF9\x04\x01\x00\x00\x00\x00,\x00\x00\x00\x00\x01\x00\x01\x00\x00\x02\x02\x04\x01\x00;")

	// Create character configuration
	characterConfig := `{
		"name": "Test Pet",
		"description": "A test character for GIF bug testing",
		"animations": {
			"idle": "idle.gif",
			"talking": "talking.gif"
		},
		"dialogs": [
			{
				"trigger": "click",
				"responses": ["Hello test!"],
				"animation": "talking",
				"cooldown": 1
			}
		],
		"behavior": {
			"idleTimeout": 10,
			"movementEnabled": false,
			"defaultSize": 64
		}
	}`

	characterPath := filepath.Join(tmpDir, "character.json")
	err = os.WriteFile(characterPath, []byte(characterConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write character config: %v", err)
	}

	// Write malformed GIF files
	idlePath := filepath.Join(tmpDir, "idle.gif")
	talkingPath := filepath.Join(tmpDir, "talking.gif")

	err = os.WriteFile(idlePath, malformedGIF, 0644)
	if err != nil {
		t.Fatalf("Failed to write idle.gif: %v", err)
	}

	err = os.WriteFile(talkingPath, malformedGIF, 0644)
	if err != nil {
		t.Fatalf("Failed to write talking.gif: %v", err)
	}

	// Load character card
	card, err := character.LoadCard(characterPath)
	if err != nil {
		t.Fatalf("Failed to load character card: %v", err)
	}

	// This should fail due to malformed GIF data
	_, err = character.New(card, tmpDir)
	if err == nil {
		t.Fatal("Expected error when creating character with malformed GIF data, but succeeded")
	}

	// Verify it's the expected error format for animation loading failures
	expectedMsg := "failed to load any animations"
	if !contains(err.Error(), expectedMsg) {
		t.Fatalf("Expected error containing %q, got: %v", expectedMsg, err)
	}

	t.Logf("Confirmed malformed GIF still fails as expected: %v", err)
}

// TestBug2InvalidGIFDataFixed tests that character creation works with valid GIF data
func TestBug2InvalidGIFDataFixed(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "gif_fix_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Use valid GIF data (same as integration test now uses)
	validGIF := []byte{71, 73, 70, 56, 57, 97, 1, 0, 1, 0, 128, 0, 0, 255, 255, 255, 0, 0, 0, 44, 0, 0, 0, 0, 1, 0, 1, 0, 0, 2, 2, 68, 1, 0, 59}

	// Create character configuration
	characterConfig := `{
		"name": "Test Pet",
		"description": "A test character for GIF fix testing",
		"animations": {
			"idle": "idle.gif",
			"talking": "talking.gif"
		},
		"dialogs": [
			{
				"trigger": "click",
				"responses": ["Hello test!"],
				"animation": "talking",
				"cooldown": 1
			}
		],
		"behavior": {
			"idleTimeout": 10,
			"movementEnabled": false,
			"defaultSize": 64
		}
	}`

	characterPath := filepath.Join(tmpDir, "character.json")
	err = os.WriteFile(characterPath, []byte(characterConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write character config: %v", err)
	}

	// Write valid GIF files
	idlePath := filepath.Join(tmpDir, "idle.gif")
	talkingPath := filepath.Join(tmpDir, "talking.gif")

	err = os.WriteFile(idlePath, validGIF, 0644)
	if err != nil {
		t.Fatalf("Failed to write idle.gif: %v", err)
	}

	err = os.WriteFile(talkingPath, validGIF, 0644)
	if err != nil {
		t.Fatalf("Failed to write talking.gif: %v", err)
	}

	// Load character card
	card, err := character.LoadCard(characterPath)
	if err != nil {
		t.Fatalf("Failed to load character card: %v", err)
	}

	// This should now succeed with valid GIF data
	char, err := character.New(card, tmpDir)
	if err != nil {
		t.Fatalf("Expected character creation to succeed with valid GIF data, got error: %v", err)
	}

	// Verify character was created successfully
	if char.GetName() != "Test Pet" {
		t.Errorf("Expected character name 'Test Pet', got '%s'", char.GetName())
	}

	t.Logf("Character created successfully with valid GIF data")
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[len(s)-len(substr):] == substr ||
		len(s) > len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
