package ui

import (
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	"desktop-companion/internal/monitoring"
)

func TestDesktopWindow_ChatbotIntegration(t *testing.T) {
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

	// Create desktop window with debug mode
	profiler := monitoring.NewProfiler(50)
	window := NewDesktopWindow(app, char, true, profiler, false, false, nil, false, false, false)

	// Test 1: Verify chatbot interface is created for AI-enabled characters
	if window.chatbotInterface == nil {
		t.Error("Expected chatbot interface to be created for AI-enabled character, but it was nil")
		return
	}

	// Test 2: Test keyboard shortcut for toggling chatbot interface
	canvas := window.window.Canvas()

	// Initially should be hidden
	if window.chatbotInterface.IsVisible() {
		t.Error("Expected chatbot interface to be initially hidden")
	}

	// Simulate 'C' key press to toggle chatbot
	keyEvent := &fyne.KeyEvent{Name: fyne.KeyC}
	canvas.OnTypedKey()(keyEvent)

	// Give UI time to update
	time.Sleep(10 * time.Millisecond)

	// Should now be visible
	if !window.chatbotInterface.IsVisible() {
		t.Error("Expected chatbot interface to be visible after pressing 'C' key")
	}

	// Press 'C' again to hide
	canvas.OnTypedKey()(keyEvent)
	time.Sleep(10 * time.Millisecond)

	// Should be hidden again
	if window.chatbotInterface.IsVisible() {
		t.Error("Expected chatbot interface to be hidden after pressing 'C' key again")
	}

	t.Log("Chatbot interface keyboard shortcut integration working correctly")
}

func TestDesktopWindow_ChatbotToggleMethod(t *testing.T) {
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
	window := NewDesktopWindow(app, char, true, profiler, false, false, nil, false, false, false)

	if window.chatbotInterface == nil {
		t.Error("Expected chatbot interface to be created")
		return
	}

	// Test direct toggle method
	initialState := window.chatbotInterface.IsVisible()

	window.ToggleChatbotInterface()

	if window.chatbotInterface.IsVisible() == initialState {
		t.Error("Expected chatbot interface visibility to change after toggle")
	}

	// Toggle again
	window.ToggleChatbotInterface()

	if window.chatbotInterface.IsVisible() != initialState {
		t.Error("Expected chatbot interface to return to initial state after second toggle")
	}

	t.Log("Chatbot interface toggle method working correctly")
}

func TestDesktopWindow_WithoutAI(t *testing.T) {
	// Create test app
	app := test.NewApp()
	defer app.Quit()

	// Create character without dialog backend (regular character)
	card := createTestCharacterCard()
	char := createMockCharacter(card)
	if char == nil {
		t.Skipf("Skipping test due to character creation failure")
		return
	}

	// Create desktop window
	profiler := monitoring.NewProfiler(50)
	window := NewDesktopWindow(app, char, true, profiler, false, false, nil, false, false, false)

	// Test: Verify chatbot interface is NOT created for non-AI characters
	if window.chatbotInterface != nil {
		t.Error("Expected chatbot interface to be nil for non-AI character, but it was created")
	}

	// Test that toggle method doesn't panic when chatbot interface is nil
	window.ToggleChatbotInterface() // Should not panic

	t.Log("Non-AI character correctly has no chatbot interface")
}

func TestDesktopWindow_ContextMenuChatOption(t *testing.T) {
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
	window := NewDesktopWindow(app, char, true, profiler, false, false, nil, false, false, false)

	if window.chatbotInterface == nil {
		t.Error("Expected chatbot interface to be created")
		return
	}

	// Test context menu creation
	window.showContextMenu()

	// Context menu should be visible
	if !window.contextMenu.IsVisible() {
		t.Error("Expected context menu to be visible after showContextMenu")
	}

	// Note: Testing the actual menu item callback would require more complex UI simulation
	// For now, we verify that the context menu system doesn't break with chatbot integration

	t.Log("Context menu integration with chatbot working correctly")
}
