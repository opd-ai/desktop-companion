package ui

import (
	"testing"

	"github.com/opd-ai/desktop-companion/internal/character"
	"github.com/opd-ai/desktop-companion/internal/dialog"
)

// TestBug3MissingChatContextMenu tests that "Open Chat" context menu is missing for some AI characters
// This test reproduces the bug where characters without dialog backend don't show "Open Chat" option
func TestBug3MissingChatContextMenu(t *testing.T) {
	t.Log("Testing Bug #3: Chatbot Context Menu Access Inconsistency")

	// Test 1: Character with no dialog backend (should this show Open Chat?)
	t.Run("CharacterWithoutDialogBackend", func(t *testing.T) {
		card := &character.CharacterCard{
			Name:        "TestCharacter",
			Description: "Test character without dialog backend",
			// No DialogBackend field set
		}

		// Create character instance
		char, err := character.New(card, "/tmp")
		if err != nil {
			t.Fatalf("Failed to create character: %v", err)
		}

		// Create a minimal DesktopWindow for testing logic without GUI
		window := &DesktopWindow{
			character: char,
		}

		// Check if character should show chat option (using new logic)
		shouldShowChat := window.shouldShowChatOption()

		if shouldShowChat {
			t.Log("FIXED: Character without dialog backend now shows 'Open Chat' option")
		} else {
			t.Log("Character without dialog backend has no 'Open Chat' option")
		}

		// Verify HasDialogBackend returns false
		hasBackend := char.GetCard().HasDialogBackend()
		if hasBackend {
			t.Error("HasDialogBackend should return false for character without dialog backend")
		}

		t.Log("Expected behavior: Some AI characters should show 'Open Chat' even without full dialog backend")
	})

	// Test 2: Character with dialog backend but disabled
	t.Run("CharacterWithDisabledDialogBackend", func(t *testing.T) {
		card := &character.CharacterCard{
			Name:        "TestCharacter",
			Description: "Test character with disabled dialog backend",
			DialogBackend: &dialog.DialogBackendConfig{
				DefaultBackend: "test",
				Enabled:        false, // Explicitly disabled
			},
		}

		// Create character instance
		char, err := character.New(card, "/tmp")
		if err != nil {
			t.Fatalf("Failed to create character: %v", err)
		}

		// Create a minimal DesktopWindow for testing logic without GUI
		window := &DesktopWindow{
			character: char,
		}

		// Check if character should show chat option (using new logic)
		shouldShowChat := window.shouldShowChatOption()

		if shouldShowChat {
			t.Log("FIXED: Character with disabled dialog backend now shows 'Open Chat' option")
		} else {
			t.Log("Character with disabled dialog backend has no 'Open Chat' option")
		}

		// Verify HasDialogBackend returns false (because Enabled is false)
		hasBackend := char.GetCard().HasDialogBackend()
		if hasBackend {
			t.Error("HasDialogBackend should return false for disabled dialog backend")
		}

		t.Log("Expected behavior: 'Open Chat' should be available but show appropriate message when disabled")
	})

	t.Log("Bug #3 historical test: Shows behavior before and after fix")
	t.Log("The shouldShowChatOption() method now handles both dialog backend config AND romance features")
}
