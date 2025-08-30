package ui

import (
	"testing"

	"desktop-companion/internal/character"
)

// TestBug1GiftContextMenuRegression validates that the gift context menu fix continues to work
// This is a regression test for Bug #1: Missing Gift System Context Menu Integration
func TestBug1GiftContextMenuRegression(t *testing.T) {
	t.Run("CharacterWithGiftSystem", func(t *testing.T) {
		// Create a character with gift system enabled and game features
		card := &character.CharacterCard{
			Name: "Gift Character",
			GiftSystem: &character.GiftSystemConfig{
				Enabled: true,
			},
			Stats: map[string]character.StatConfig{
				"happiness": {
					Initial:           50,
					Max:               100,
					DegradationRate:   1.0,
					CriticalThreshold: 20,
				},
			},
		}

		char, err := character.New(card, "")
		if err != nil {
			t.Fatalf("Failed to create character: %v", err)
		}

		// Create DesktopWindow with gift system
		dw := &DesktopWindow{
			character: char,
			gameMode:  true,
		}

		// Initialize gift dialog (normally done in NewDesktopWindow)
		if char.GetCard().HasGiftSystem() && char.GetGameState() != nil {
			giftManager := character.NewGiftManager(char.GetCard(), char.GetGameState())
			dw.giftDialog = NewGiftSelectionDialog(giftManager)
		}

		// Verify preconditions
		if !char.GetCard().HasGiftSystem() {
			t.Fatal("Character should have gift system enabled")
		}

		// Get game mode menu items
		menuItems := dw.buildGameModeMenuItems()

		// Verify gift option is present
		hasGiftOption := false
		for _, item := range menuItems {
			if item.Text == "Give Gift" {
				hasGiftOption = true
				break
			}
		}

		if !hasGiftOption {
			t.Error("REGRESSION: Gift option should be present in context menu for characters with gift system")
		}

		// Verify the callback is not nil (basic sanity check)
		for _, item := range menuItems {
			if item.Text == "Give Gift" && item.Callback == nil {
				t.Error("Gift option callback should not be nil")
			}
		}
	})

	t.Run("CharacterWithoutGiftSystem", func(t *testing.T) {
		// Create a character without gift system
		card := &character.CharacterCard{
			Name: "No Gift Character",
			// No GiftSystem configured
			Stats: map[string]character.StatConfig{
				"happiness": {
					Initial:           50,
					Max:               100,
					DegradationRate:   1.0,
					CriticalThreshold: 20,
				},
			},
		}

		char, err := character.New(card, "")
		if err != nil {
			t.Fatalf("Failed to create character: %v", err)
		}

		dw := &DesktopWindow{
			character: char,
			gameMode:  true,
		}

		// Verify character does NOT have gift system
		if char.GetCard().HasGiftSystem() {
			t.Fatal("Character should NOT have gift system enabled")
		}

		// Get game mode menu items
		menuItems := dw.buildGameModeMenuItems()

		// Verify gift option is NOT present
		for _, item := range menuItems {
			if item.Text == "Give Gift" {
				t.Error("Gift option should NOT be present for characters without gift system")
			}
		}

		// Verify standard options are still there
		hasPlay := false
		hasFeed := false
		for _, item := range menuItems {
			if item.Text == "Play" {
				hasPlay = true
			}
			if item.Text == "Feed" {
				hasFeed = true
			}
		}

		if !hasPlay {
			t.Error("Play option should be present in game mode")
		}
		if !hasFeed {
			t.Error("Feed option should be present in game mode")
		}
	})

	t.Run("NonGameMode", func(t *testing.T) {
		// Create a character with gift system but not in game mode
		card := &character.CharacterCard{
			Name: "Gift Character Non-Game",
			GiftSystem: &character.GiftSystemConfig{
				Enabled: true,
			},
			Stats: map[string]character.StatConfig{
				"happiness": {
					Initial:           50,
					Max:               100,
					DegradationRate:   1.0,
					CriticalThreshold: 20,
				},
			},
		}

		char, err := character.New(card, "")
		if err != nil {
			t.Fatalf("Failed to create character: %v", err)
		}

		dw := &DesktopWindow{
			character: char,
			gameMode:  false, // Not in game mode
		}

		// Get game mode menu items (should return nil/empty)
		menuItems := dw.buildGameModeMenuItems()

		// Should have no menu items since not in game mode
		if len(menuItems) > 0 {
			t.Error("Should have no game mode menu items when not in game mode")
		}
	})
}
