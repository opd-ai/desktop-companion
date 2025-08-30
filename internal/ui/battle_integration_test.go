package ui

import (
	"testing"

	"desktop-companion/internal/character"
	"desktop-companion/internal/monitoring"

	"fyne.io/fyne/v2/test"
)

// TestBattleMenuIntegration tests battle menu integration with context menu
func TestBattleMenuIntegration(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	// Create a test character card with battle system enabled
	card := &character.CharacterCard{
		Name:        "Battle Character",
		Description: "A character for testing battle functionality",
		Animations: map[string]string{
			"idle":    "idle.gif",
			"talking": "talking.gif",
		},
		Dialogs: []character.Dialog{
			{
				Trigger:   "click",
				Responses: []string{"Hello!"},
				Animation: "talking",
			},
		},
		Behavior: character.Behavior{
			IdleTimeout:     10,
			MovementEnabled: true,
			DefaultSize:     100,
		},
		BattleSystem: &character.BattleSystemConfig{
			Enabled: true,
			BattleStats: map[string]character.BattleStat{
				"hp":      {Base: 100, Max: 100},
				"attack":  {Base: 20, Max: 25},
				"defense": {Base: 15, Max: 20},
				"speed":   {Base: 10, Max: 15},
			},
			AIDifficulty:      "normal",
			PreferredActions:  []string{"attack", "heal", "defend"},
			RequireAnimations: false,
		},
	}

	// Create character instance
	char, err := character.New(card, "../../testdata")
	if err != nil {
		t.Fatalf("Failed to create character: %v", err)
	}

	// Create desktop window
	window := NewDesktopWindow(app, char, false, monitoring.NewProfiler(50), false, false, nil, false, false, false)

	t.Run("should_show_battle_options", func(t *testing.T) {
		shouldShow := window.shouldShowBattleOptions()
		if !shouldShow {
			t.Error("Should show battle options for character with battle system enabled")
		}
	})

	t.Run("build_battle_menu_items", func(t *testing.T) {
		menuItems := window.buildBattleMenuItems()
		if len(menuItems) == 0 {
			t.Error("Should have battle menu items for character with battle system enabled")
		}

		if len(menuItems) < 1 {
			t.Fatal("Expected at least one battle menu item")
		}

		found := false
		for _, item := range menuItems {
			if item.Text == "Initiate Battle" {
				found = true
				break
			}
		}

		if !found {
			t.Error("Should have 'Initiate Battle' menu item")
		}
	})

	t.Run("battle_menu_callback", func(t *testing.T) {
		menuItems := window.buildBattleMenuItems()
		if len(menuItems) == 0 {
			t.Fatal("No battle menu items returned")
		}

		battleItem := menuItems[0]
		if battleItem.Callback == nil {
			t.Error("Battle menu item should have callback")
		}

		// Call the callback - should not panic
		battleItem.Callback()
	})
}

// TestBattleMenuWithoutBattleSystem tests that battle menu is not shown for non-battle characters
func TestBattleMenuWithoutBattleSystem(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	// Create a test character card WITHOUT battle system
	card := &character.CharacterCard{
		Name:        "Non-Battle Character",
		Description: "A character without battle functionality",
		Animations: map[string]string{
			"idle":    "idle.gif",
			"talking": "talking.gif",
		},
		Dialogs: []character.Dialog{
			{
				Trigger:   "click",
				Responses: []string{"Hello!"},
				Animation: "talking",
			},
		},
		Behavior: character.Behavior{
			IdleTimeout:     10,
			MovementEnabled: true,
			DefaultSize:     100,
		},
		// No BattleSystem field or it's nil
	}

	// Create character instance
	char, err := character.New(card, "../../testdata")
	if err != nil {
		t.Fatalf("Failed to create character: %v", err)
	}

	// Create desktop window
	window := NewDesktopWindow(app, char, false, monitoring.NewProfiler(50), false, false, nil, false, false, false)

	t.Run("should_not_show_battle_options", func(t *testing.T) {
		shouldShow := window.shouldShowBattleOptions()
		if shouldShow {
			t.Error("Should not show battle options for character without battle system")
		}
	})

	t.Run("no_battle_menu_items", func(t *testing.T) {
		menuItems := window.buildBattleMenuItems()
		if len(menuItems) != 0 {
			t.Error("Should have no battle menu items for character without battle system")
		}
	})
}

// TestBattleInitiationHandler tests the battle initiation handler
func TestBattleInitiationHandler(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	// Create a test character card with battle system enabled
	card := &character.CharacterCard{
		Name:        "Battle Character",
		Description: "A character for testing battle functionality",
		Animations: map[string]string{
			"idle":    "idle.gif",
			"talking": "talking.gif",
		},
		Dialogs: []character.Dialog{
			{
				Trigger:   "click",
				Responses: []string{"Hello!"},
				Animation: "talking",
			},
		},
		Behavior: character.Behavior{
			IdleTimeout:     10,
			MovementEnabled: true,
			DefaultSize:     100,
		},
		BattleSystem: &character.BattleSystemConfig{
			Enabled: true,
		},
	}

	// Create character instance
	char, err := character.New(card, "../../testdata")
	if err != nil {
		t.Fatalf("Failed to create character: %v", err)
	}

	// Create desktop window
	window := NewDesktopWindow(app, char, false, monitoring.NewProfiler(50), false, false, nil, false, false, false)

	t.Run("handle_battle_initiation", func(t *testing.T) {
		// Should not panic when called
		window.handleBattleInitiation()

		// Since this currently just shows a dialog, we can't easily test the content
		// but at least we verify it doesn't crash
	})
}
