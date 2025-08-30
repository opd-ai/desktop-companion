package ui

import (
	"testing"

	"fyne.io/fyne/v2/app"

	"desktop-companion/internal/character"
	"desktop-companion/internal/dialog"
	"desktop-companion/internal/monitoring"
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

		// Create test app and window
		testApp := app.New()
		profiler := monitoring.NewProfiler(50)

		window := NewDesktopWindow(testApp, char, false, profiler, false, false, nil, false, false, false)

		// Check if chatbot interface was created
		hasChatbot := window.chatbotInterface != nil

		if hasChatbot {
			t.Log("Character without dialog backend has chatbot interface (unexpected)")
		} else {
			t.Log("Bug confirmed: Character without dialog backend has no chatbot interface")
			t.Log("This means 'Open Chat' won't appear in context menu")
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

		// Create test app and window
		testApp := app.New()
		profiler := monitoring.NewProfiler(50)

		window := NewDesktopWindow(testApp, char, false, profiler, false, false, nil, false, false, false)

		// Check if chatbot interface was created
		hasChatbot := window.chatbotInterface != nil

		if hasChatbot {
			t.Log("Character with disabled dialog backend has chatbot interface (unexpected)")
		} else {
			t.Log("Bug confirmed: Character with disabled dialog backend has no chatbot interface")
			t.Log("This means 'Open Chat' won't appear in context menu even though backend exists")
		}

		// Verify HasDialogBackend returns false (because Enabled is false)
		hasBackend := char.GetCard().HasDialogBackend()
		if hasBackend {
			t.Error("HasDialogBackend should return false for disabled dialog backend")
		}

		t.Log("Expected behavior: 'Open Chat' should be available but show appropriate message when disabled")
	})

	t.Log("Bug #3 analysis: Context menu 'Open Chat' availability depends strictly on HasDialogBackend()")
	t.Log("This may be too restrictive and could confuse users who expect to see the option")
}
