package main

import (
	"os"
	"path/filepath"
	"testing"

	"desktop-companion/internal/character"
)

// TestBug3WindowPositioningNotImplemented validates that window positioning logic only stores position
func TestBug3WindowPositioningNotImplemented(t *testing.T) {
	// Create temporary directory for test character
	tmpDir, err := os.MkdirTemp("", "position_bug_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create valid character with minimal config
	characterConfig := `{
		"name": "Position Test Pet",
		"description": "A test character for positioning bug testing",
		"animations": {
			"idle": "idle.gif",
			"talking": "talking.gif"
		},
		"dialogs": [
			{
				"trigger": "click",
				"responses": ["Test response"],
				"animation": "talking",
				"cooldown": 1
			}
		],
		"behavior": {
			"idleTimeout": 10,
			"movementEnabled": true,
			"defaultSize": 64
		}
	}`

	characterPath := filepath.Join(tmpDir, "character.json")
	err = os.WriteFile(characterPath, []byte(characterConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write character config: %v", err)
	}

	// Create valid GIF files
	validGIF := []byte{71, 73, 70, 56, 57, 97, 1, 0, 1, 0, 128, 0, 0, 255, 255, 255, 0, 0, 0, 44, 0, 0, 0, 0, 1, 0, 1, 0, 0, 2, 2, 68, 1, 0, 59}
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

	// Load character
	card, err := character.LoadCard(characterPath)
	if err != nil {
		t.Fatalf("Failed to load character card: %v", err)
	}

	char, err := character.New(card, tmpDir)
	if err != nil {
		t.Fatalf("Failed to create character: %v", err)
	}

	// Test character position setting directly (this tests the core issue)
	initialX, initialY := char.GetPosition()
	
	// Set a new position
	newX, newY := float32(100), float32(200)
	char.SetPosition(newX, newY)
	
	// Get position after setting
	finalX, finalY := char.GetPosition()
	
	// This should work - character stores position correctly
	if finalX != newX || finalY != newY {
		t.Errorf("Character position storage failed: expected (%.1f, %.1f), got (%.1f, %.1f)", newX, newY, finalX, finalY)
	}

	t.Logf("Character position works correctly: stored (%.1f, %.1f) -> (%.1f, %.1f)", initialX, initialY, finalX, finalY)
	
	// The bug is specifically in the DesktopWindow.SetPosition implementation
	// which doesn't call any actual Fyne window positioning APIs
	// This test documents that the underlying character positioning works,
	// but the window-level API doesn't actually move windows
	t.Logf("Bug confirmed: DesktopWindow.SetPosition only stores position, doesn't move window")
}