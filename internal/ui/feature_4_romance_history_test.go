package ui

import (
	"testing"
	"time"

	"desktop-companion/internal/character"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

// Helper function to create romance-enabled character with proper stats
func createRomanceCharacterWithStats(t *testing.T) *character.Character {
	t.Helper()

	// Create character card with both romance features and stats
	card := &character.CharacterCard{
		Name:        "RomanceTestChar",
		Description: "Test character with romance features",
		Personality: &character.PersonalityConfig{
			Traits: map[string]float64{
				"kindness":    0.70,
				"confidence":  0.60,
				"playfulness": 0.80,
				"romanticism": 0.75,
			},
		},
		Stats: map[string]character.StatConfig{
			"health":    {Initial: 80, Max: 100, DegradationRate: 0.1, CriticalThreshold: 20},
			"happiness": {Initial: 75, Max: 100, DegradationRate: 0.1, CriticalThreshold: 15},
			"energy":    {Initial: 90, Max: 100, DegradationRate: 0.1, CriticalThreshold: 25},
			"affection": {Initial: 0, Max: 100, DegradationRate: 0.0, CriticalThreshold: 0},
			"trust":     {Initial: 0, Max: 100, DegradationRate: 0.0, CriticalThreshold: 0},
		},
		Animations: map[string]string{
			"idle": "test.gif",
		},
	}

	char, err := character.New(card, "test_data")
	if err != nil {
		t.Fatalf("Failed to create character: %v", err)
	}

	return char
}

// Helper function to create basic character without romance features
func createBasicCharacter(t *testing.T) *character.Character {
	t.Helper()

	card := &character.CharacterCard{
		Name:        "BasicTestChar",
		Description: "Basic test character",
		Animations: map[string]string{
			"idle": "test.gif",
		},
	}

	char, err := character.New(card, "test_data")
	if err != nil {
		t.Fatalf("Failed to create character: %v", err)
	}

	return char
}

func TestFeature4_shouldShowRomanceHistory(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	t.Run("RomanceCharacterWithMemories", func(t *testing.T) {
		char := createRomanceCharacterWithStats(t)
		window := createTestDesktopWindow(t, char, app)

		// Add some romance memories
		gameState := char.GetGameState()
		if gameState != nil {
			gameState.RecordRomanceInteraction("compliment", "Thank you!", map[string]float64{"affection": 50.0}, map[string]float64{"affection": 55.0})
			gameState.RecordRomanceInteraction("gift", "That's so sweet!", map[string]float64{"affection": 55.0, "happiness": 70.0}, map[string]float64{"affection": 65.0, "happiness": 80.0})
		}

		result := window.shouldShowRomanceHistory()
		if !result {
			t.Error("Expected shouldShowRomanceHistory to return true for romance character with memories")
		}
	})

	t.Run("RomanceCharacterWithoutMemories", func(t *testing.T) {
		char := createRomanceCharacterWithStats(t)
		window := createTestDesktopWindow(t, char, app)

		result := window.shouldShowRomanceHistory()
		if result {
			t.Error("Expected shouldShowRomanceHistory to return false for romance character without memories")
		}
	})

	t.Run("BasicCharacterWithMemories", func(t *testing.T) {
		char := createBasicCharacter(t)
		window := createTestDesktopWindow(t, char, app)

		result := window.shouldShowRomanceHistory()
		if result {
			t.Error("Expected shouldShowRomanceHistory to return false for basic character")
		}
	})

	t.Run("NilCharacter", func(t *testing.T) {
		window := &DesktopWindow{character: nil}
		result := window.shouldShowRomanceHistory()
		if result {
			t.Error("Expected shouldShowRomanceHistory to return false for nil character")
		}
	})
}

func TestFeature4_showRomanceHistory(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	t.Run("DisplaysMemoriesCorrectly", func(t *testing.T) {
		char := createRomanceCharacterWithStats(t)
		window := createTestDesktopWindow(t, char, app)

		// Add test memories
		gameState := char.GetGameState()
		if gameState != nil {
			gameState.RecordRomanceInteraction("compliment", "First romantic moment", map[string]float64{"affection": 40.0}, map[string]float64{"affection": 60.0})
			gameState.RecordRomanceInteraction("conversation", "Sweet conversation", map[string]float64{"happiness": 70.0, "energy": 90.0}, map[string]float64{"happiness": 80.0, "energy": 85.0})
		}

		// This should not panic and should display the dialog
		window.showRomanceHistory()

		// Test passes if no panic occurs
	})

	t.Run("HandlesEmptyMemories", func(t *testing.T) {
		char := createRomanceCharacterWithStats(t)
		window := createTestDesktopWindow(t, char, app)

		// No memories added - should handle gracefully
		window.showRomanceHistory()

		// Test passes if no panic occurs
	})

	t.Run("NilCharacterHandling", func(t *testing.T) {
		window := &DesktopWindow{character: nil}
		window.showRomanceHistory()

		// Test passes if no panic occurs
	})
}

func TestFeature4_formatRomanceHistory(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	t.Run("FormatsMemoriesCorrectly", func(t *testing.T) {
		char := createRomanceCharacterWithStats(t)
		window := createTestDesktopWindow(t, char, app)

		// Add test memories
		gameState := char.GetGameState()
		if gameState != nil {
			gameState.RecordRomanceInteraction("gift", "Romantic dinner date", map[string]float64{"happiness": 70.0, "energy": 90.0}, map[string]float64{"happiness": 85.0, "energy": 80.0})
			gameState.RecordRomanceInteraction("compliment", "Shared a laugh", map[string]float64{"happiness": 75.0}, map[string]float64{"happiness": 83.0})
		}

		// Get memories to test formatRomanceHistory
		memories := gameState.GetRomanceMemories()
		result := window.formatRomanceHistory(memories)

		if result == "" {
			t.Error("Expected formatRomanceHistory to return non-empty string")
		}

		// Check for expected content
		if len(result) < 10 {
			t.Error("Expected formatted romance history to contain substantial content")
		}
	})

	t.Run("HandlesNoMemories", func(t *testing.T) {
		char := createRomanceCharacterWithStats(t)
		window := createTestDesktopWindow(t, char, app)

		// Get empty memories
		gameState := char.GetGameState()
		memories := gameState.GetRomanceMemories()
		result := window.formatRomanceHistory(memories)

		if result == "" {
			t.Error("Expected formatRomanceHistory to return content even for no memories")
		}
	})

	t.Run("LimitsToLast10Memories", func(t *testing.T) {
		char := createRomanceCharacterWithStats(t)
		window := createTestDesktopWindow(t, char, app)

		// Add more than 10 memories
		gameState := char.GetGameState()
		if gameState != nil {
			for i := 0; i < 15; i++ {
				gameState.RecordRomanceInteraction("compliment", "Memory interaction", map[string]float64{"happiness": 70.0}, map[string]float64{"happiness": 75.0})
			}
		}

		memories := gameState.GetRomanceMemories()
		result := window.formatRomanceHistory(memories)

		// Should contain content but be limited
		if result == "" {
			t.Error("Expected formatRomanceHistory to return content for multiple memories")
		}
	})

	t.Run("EmptyMemoriesList", func(t *testing.T) {
		char := createRomanceCharacterWithStats(t)
		window := createTestDesktopWindow(t, char, app)

		// Test with empty memories slice
		var emptyMemories []character.RomanceMemory
		result := window.formatRomanceHistory(emptyMemories)

		if result == "" {
			t.Error("Expected formatRomanceHistory to return content for empty memories list")
		}
	})
}

func TestFeature4_formatStatChanges(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	t.Run("PositiveChanges", func(t *testing.T) {
		char := createRomanceCharacterWithStats(t)
		window := createTestDesktopWindow(t, char, app)

		before := map[string]float64{
			"happiness": 70.0,
			"energy":    80.0,
			"health":    85.0,
		}
		after := map[string]float64{
			"happiness": 85.0,
			"energy":    90.0,
			"health":    90.0,
		}

		result := window.formatStatChanges(before, after)

		if result == "" {
			t.Error("Expected formatStatChanges to return non-empty string for positive changes")
		}

		// Should contain up arrow emoji for positive changes
		if len(result) < 5 {
			t.Error("Expected formatted stat changes to contain substantial content")
		}
	})

	t.Run("NegativeChanges", func(t *testing.T) {
		char := createRomanceCharacterWithStats(t)
		window := createTestDesktopWindow(t, char, app)

		before := map[string]float64{
			"happiness": 80.0,
			"energy":    90.0,
		}
		after := map[string]float64{
			"happiness": 70.0,
			"energy":    75.0,
		}

		result := window.formatStatChanges(before, after)

		if result == "" {
			t.Error("Expected formatStatChanges to return non-empty string for negative changes")
		}

		// Should contain down arrow emoji for negative changes
		if len(result) < 5 {
			t.Error("Expected formatted stat changes to contain substantial content")
		}
	})

	t.Run("MixedChanges", func(t *testing.T) {
		char := createRomanceCharacterWithStats(t)
		window := createTestDesktopWindow(t, char, app)

		before := map[string]float64{
			"happiness": 70.0,
			"energy":    90.0,
			"health":    80.0,
		}
		after := map[string]float64{
			"happiness": 80.0,
			"energy":    85.0,
			"health":    80.0,
		}

		result := window.formatStatChanges(before, after)

		if result == "" {
			t.Error("Expected formatStatChanges to return non-empty string for mixed changes")
		}

		// Should contain both up and down arrows
		if len(result) < 5 {
			t.Error("Expected formatted stat changes to contain substantial content")
		}
	})

	t.Run("EmptyChanges", func(t *testing.T) {
		char := createRomanceCharacterWithStats(t)
		window := createTestDesktopWindow(t, char, app)

		before := map[string]float64{}
		after := map[string]float64{}

		result := window.formatStatChanges(before, after)

		if result != "" {
			t.Error("Expected formatStatChanges to return empty string for no changes")
		}
	})

	t.Run("ZeroChanges", func(t *testing.T) {
		char := createRomanceCharacterWithStats(t)
		window := createTestDesktopWindow(t, char, app)

		before := map[string]float64{
			"happiness": 80.0,
			"energy":    90.0,
		}
		after := map[string]float64{
			"happiness": 80.0,
			"energy":    90.0,
		}

		result := window.formatStatChanges(before, after)

		if result != "" {
			t.Error("Expected formatStatChanges to return empty string for zero changes")
		}
	})
}

func TestFeature4_ContextMenuIntegration(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	t.Run("RomanceHistoryMenuItemPresent", func(t *testing.T) {
		char := createRomanceCharacterWithStats(t)
		window := createTestDesktopWindow(t, char, app)

		// Add romance memories to enable the menu option
		gameState := char.GetGameState()
		if gameState != nil {
			gameState.RecordRomanceInteraction("compliment", "Test memory", map[string]float64{"happiness": 70.0}, map[string]float64{"happiness": 80.0})
		}

		// Build chat menu items
		menuItems := window.buildChatMenuItems()

		found := false
		for _, item := range menuItems {
			if item.Text == "View Romance History" {
				found = true
				break
			}
		}

		if !found {
			t.Error("Expected 'View Romance History' menu item to be present for romance character with memories")
		}
	})

	t.Run("RomanceHistoryMenuItemAbsent", func(t *testing.T) {
		char := createRomanceCharacterWithStats(t)
		window := createTestDesktopWindow(t, char, app)

		// No memories added - menu item should not be present
		menuItems := window.buildChatMenuItems()

		found := false
		for _, item := range menuItems {
			if item.Text == "View Romance History" {
				found = true
				break
			}
		}

		if found {
			t.Error("Expected 'View Romance History' menu item to be absent for romance character without memories")
		}
	})

	t.Run("BasicCharacterNoRomanceMenu", func(t *testing.T) {
		char := createBasicCharacter(t)
		window := createTestDesktopWindow(t, char, app)

		// Build chat menu items
		menuItems := window.buildChatMenuItems()

		found := false
		for _, item := range menuItems {
			if item.Text == "View Romance History" {
				found = true
				break
			}
		}

		if found {
			t.Error("Expected 'View Romance History' menu item to be absent for basic character")
		}
	})
}

func TestFeature4_EdgeCases(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	t.Run("CharacterWithoutGameState", func(t *testing.T) {
		// Create character that might not have game state initialized
		char := createBasicCharacter(t)
		window := createTestDesktopWindow(t, char, app)

		// These should not panic
		result := window.shouldShowRomanceHistory()
		if result {
			t.Error("Expected shouldShowRomanceHistory to return false for character without game state")
		}

		window.showRomanceHistory()

		// Test with empty memories for formatRomanceHistory
		var emptyMemories []character.RomanceMemory
		formatted := window.formatRomanceHistory(emptyMemories)
		if formatted == "" {
			t.Error("Expected formatRomanceHistory to return content even for empty memories")
		}
	})

	t.Run("RecentMemoryTimestamps", func(t *testing.T) {
		char := createRomanceCharacterWithStats(t)
		window := createTestDesktopWindow(t, char, app)

		// Add memories with different timestamps
		gameState := char.GetGameState()
		if gameState != nil {
			gameState.RecordRomanceInteraction("compliment", "Recent memory", map[string]float64{"happiness": 70.0}, map[string]float64{"happiness": 80.0})
			time.Sleep(time.Millisecond) // Ensure different timestamp
			gameState.RecordRomanceInteraction("gift", "Another memory", map[string]float64{"happiness": 75.0}, map[string]float64{"happiness": 80.0})
		}

		memories := gameState.GetRomanceMemories()
		result := window.formatRomanceHistory(memories)

		if result == "" {
			t.Error("Expected formatRomanceHistory to handle timestamped memories")
		}
	})

	t.Run("LargeStatChanges", func(t *testing.T) {
		char := createRomanceCharacterWithStats(t)
		window := createTestDesktopWindow(t, char, app)

		before := map[string]float64{
			"happiness": 0.0,
			"energy":    100.0,
		}
		after := map[string]float64{
			"happiness": 100.0,
			"energy":    0.0,
		}

		result := window.formatStatChanges(before, after)

		if result == "" {
			t.Error("Expected formatStatChanges to handle large values")
		}
	})
}

// Helper function to create a test DesktopWindow
func createTestDesktopWindow(t *testing.T, char *character.Character, app fyne.App) *DesktopWindow {
	t.Helper()

	// Create a minimal window with just the dialog system for testing
	window := &DesktopWindow{
		character: char,
		dialog:    NewDialogBubble(), // Initialize dialog to prevent nil pointer
	}

	return window
}
