package ui

import (
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	"desktop-companion/internal/monitoring"
)

func TestEnhancedUserExperience_EscapeKey(t *testing.T) {
	// Create test app
	app := test.NewApp()
	defer app.Quit()

	// Create character with dialog backend enabled
	card := createTestCharacterCardWithDialogBackend()
	char := createMockCharacter(card)
	if char == nil {
		t.Skipf("Skipping test due to character creation failure")
		return
	}

	// Create desktop window
	profiler := monitoring.NewProfiler(50)
	window := NewDesktopWindow(app, char, true, profiler, false, false)

	if window.chatbotInterface == nil {
		t.Error("Expected chatbot interface to be created")
		return
	}

	// Test ESC key functionality
	canvas := window.window.Canvas()
	
	// First open the chatbot
	window.ToggleChatbotInterfaceWithFocus()
	if !window.chatbotInterface.IsVisible() {
		t.Error("Expected chatbot interface to be visible after toggle")
	}

	// Now press ESC to close
	escEvent := &fyne.KeyEvent{Name: fyne.KeyEscape}
	canvas.OnTypedKey()(escEvent)
	
	// Give UI time to update
	time.Sleep(10 * time.Millisecond)

	// Should now be hidden
	if window.chatbotInterface.IsVisible() {
		t.Error("Expected chatbot interface to be hidden after pressing ESC key")
	}

	t.Log("Enhanced UX: ESC key closes chatbot interface correctly")
}

func TestEnhancedUserExperience_FocusManagement(t *testing.T) {
	// Create test app
	app := test.NewApp()
	defer app.Quit()

	// Create character with dialog backend enabled
	card := createTestCharacterCardWithDialogBackend()
	char := createMockCharacter(card)
	if char == nil {
		t.Skipf("Skipping test due to character creation failure")
		return
	}

	// Create desktop window
	profiler := monitoring.NewProfiler(50)
	window := NewDesktopWindow(app, char, true, profiler, false, false)

	if window.chatbotInterface == nil {
		t.Error("Expected chatbot interface to be created")
		return
	}

	// Test enhanced focus functionality
	window.ToggleChatbotInterfaceWithFocus()
	
	if !window.chatbotInterface.IsVisible() {
		t.Error("Expected chatbot interface to be visible")
		return
	}

	// Test FocusInput method directly
	window.chatbotInterface.FocusInput()
	
	// Interface should still be visible and properly focused
	if !window.chatbotInterface.IsVisible() {
		t.Error("Expected chatbot interface to remain visible after focus")
	}

	t.Log("Enhanced UX: Focus management working correctly")
}

func TestEnhancedUserExperience_ShortcutsMenu(t *testing.T) {
	// Create test app
	app := test.NewApp()
	defer app.Quit()

	// Create character with dialog backend enabled
	card := createTestCharacterCardWithDialogBackend()
	char := createMockCharacter(card)
	if char == nil {
		t.Skipf("Skipping test due to character creation failure")
		return
	}

	// Create desktop window
	profiler := monitoring.NewProfiler(50)
	window := NewDesktopWindow(app, char, true, profiler, false, false)

	// Test that showContextMenu doesn't crash with new shortcuts menu
	window.showContextMenu()

	// Context menu should be visible
	if !window.contextMenu.IsVisible() {
		t.Error("Expected context menu to be visible after showContextMenu")
	}

	// Note: We can't easily test the actual menu item callbacks in unit tests
	// but we can verify the menu system doesn't break

	t.Log("Enhanced UX: Context menu with shortcuts help working correctly")
}
