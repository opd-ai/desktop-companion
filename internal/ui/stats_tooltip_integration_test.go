package ui

import (
	"testing"

	"fyne.io/fyne/v2/test"
)

// TestFeature3QuickStatsPeek tests the hover tooltip functionality
func TestFeature3QuickStatsPeek(t *testing.T) {
	// Create test character with game state
	char := createTestCharacterWithGame(t, t.TempDir())

	// Create a desktop window with tooltip functionality
	app := test.NewApp()
	defer test.NewApp() // Reset test app
	window := NewDesktopWindow(app, char, false, nil, true, false, nil, false, false, false)

	// Test tooltip creation
	if window.statsTooltip == nil {
		t.Error("Expected statsTooltip to be created for game mode character")
	}

	// Test tooltip show/hide methods
	window.ShowStatsTooltip()
	if !window.statsTooltip.IsVisible() {
		t.Error("Expected tooltip to be visible after ShowStatsTooltip()")
	}

	window.HideStatsTooltip()
	if window.statsTooltip.IsVisible() {
		t.Error("Expected tooltip to be hidden after HideStatsTooltip()")
	}
}

// TestFeature3HoverDetection tests the draggable character hover detection
func TestFeature3HoverDetection(t *testing.T) {
	// Create test character with game state
	char := createTestCharacterWithGame(t, t.TempDir())

	// Create a desktop window
	app := test.NewApp()
	defer test.NewApp() // Reset test app
	window := NewDesktopWindow(app, char, false, nil, true, false, nil, false, false, false)

	// Create draggable character
	draggable := NewDraggableCharacter(window, char, false)

	// Test that draggable character has hover fields
	if !draggable.isHovering == true { // Initial state should be false
		// This test just verifies the field exists
	}

	// Simulate mouse in event
	draggable.MouseIn(nil)
	if !draggable.isHovering {
		t.Error("Expected isHovering to be true after MouseIn")
	}

	// Simulate mouse out event
	draggable.MouseOut()
	if draggable.isHovering {
		t.Error("Expected isHovering to be false after MouseOut")
	}
}

// TestFeature3IntegrationWithExistingOverlay tests tooltip works with existing stats overlay
func TestFeature3IntegrationWithExistingOverlay(t *testing.T) {
	// Create test character with game state
	char := createTestCharacterWithGame(t, t.TempDir())

	// Create a desktop window with stats overlay shown
	app := test.NewApp()
	defer test.NewApp() // Reset test app
	window := NewDesktopWindow(app, char, false, nil, true, true, nil, false, false, false)

	// Both stats overlay and tooltip should exist
	if window.statsOverlay == nil {
		t.Error("Expected statsOverlay to be created")
	}
	if window.statsTooltip == nil {
		t.Error("Expected statsTooltip to be created")
	}

	// Both can be visible at the same time
	window.ShowStatsTooltip()
	if !window.statsOverlay.IsVisible() {
		t.Error("Expected stats overlay to remain visible")
	}
	if !window.statsTooltip.IsVisible() {
		t.Error("Expected tooltip to be visible")
	}
}

// TestFeature3WithoutGameMode tests tooltip not created without game mode
func TestFeature3WithoutGameMode(t *testing.T) {
	// Create test character without game state
	char := createTestCharacterWithoutGame(t, t.TempDir())

	// Create a desktop window without game mode
	app := test.NewApp()
	defer test.NewApp() // Reset test app
	window := NewDesktopWindow(app, char, false, nil, false, false, nil, false, false, false)

	// Stats tooltip should not be created
	if window.statsTooltip != nil {
		t.Error("Expected statsTooltip to be nil for non-game mode character")
	}
}
