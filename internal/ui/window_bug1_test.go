package ui

import (
	"testing"

	"desktop-companion/internal/character"
)

// TestBug1GiftContextMenuFix validates the fix for missing gift system context menu integration
// This test verifies that characters with gift system enabled show "Give Gift" option in context menu
// This was Bug #1 in AUDIT.md: Missing Gift System Context Menu Integration
func TestBug1GiftContextMenuFix(t *testing.T) {
	// Create a character with gift system enabled and game features
	card := &character.CharacterCard{
		Name: "Test Character",
		GiftSystem: &character.GiftSystemConfig{
			Enabled: true,
		},
		// Add stats to enable game features
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

	// Create DesktopWindow
	dw := &DesktopWindow{
		character: char,
		gameMode:  true,
	}

	// Manually initialize gift dialog (normally done in NewDesktopWindow)
	if char.GetCard().HasGiftSystem() && char.GetGameState() != nil {
		giftManager := character.NewGiftManager(char.GetCard(), char.GetGameState())
		dw.giftDialog = NewGiftSelectionDialog(giftManager)
	}

	// Debug: Check what's available
	t.Logf("Character has gift system: %v", char.GetCard().HasGiftSystem())
	t.Logf("Character has game features: %v", char.GetCard().HasGameFeatures())
	t.Logf("Character has game state: %v", char.GetGameState() != nil)
	t.Logf("Gift dialog created: %v", dw.giftDialog != nil) // Verify character has gift system
	if !char.GetCard().HasGiftSystem() {
		t.Fatal("Character should have gift system enabled")
	}

	// Verify character has game features (required for game mode menu)
	if !char.GetCard().HasGameFeatures() {
		t.Fatal("Character should have game features enabled")
	} // Get game mode menu items
	menuItems := dw.buildGameModeMenuItems()

	// Verify that gift option IS present (confirming the fix works)
	var hasGiftOption bool
	for _, item := range menuItems {
		if item.Text == "Give Gift" {
			hasGiftOption = true
			break
		}
	}

	// This assertion should FAIL when bug exists, PASS when bug is fixed
	if !hasGiftOption {
		t.Error("EXPECTED: Gift option should be found in menu after fix")
	} // Verify other expected options are present
	expectedOptions := []string{"Feed", "Play"}
	for _, expected := range expectedOptions {
		found := false
		for _, item := range menuItems {
			if item.Text == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected menu option '%s' not found", expected)
		}
	}
}
