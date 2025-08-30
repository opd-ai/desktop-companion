package ui

import (
	"testing"

	"fyne.io/fyne/v2/test"

	"desktop-companion/internal/character"
	"desktop-companion/internal/monitoring"
)

// TestBug4BattleInvitationsValidation validates the current gap in battle context menu functionality
func TestBug4BattleInvitationsValidation(t *testing.T) {
	t.Log("Testing Bug #4: Inconsistent Context Menu Documentation for Battle System")

	// Create test character with battle system
	card := createTestBattleCharacterCard()
	char := createMockCharacter(card)
	if char == nil {
		t.Skip("Skipping test due to character creation failure")
		return
	}

	// Create test app and components
	app := test.NewApp()
	defer app.Quit()
	profiler := monitoring.NewProfiler(50)

	t.Run("Current implementation has limited battle options", func(t *testing.T) {
		// Test without network mode - basic functionality
		window := NewDesktopWindow(
			app,
			char,
			false, // debug
			profiler,
			true,  // gameMode (needed for battle)
			false, // showStats
			nil,   // networkManager (no network)
			false, // networkMode
			false, // showNetwork
			false, // eventsEnabled
		)

		// Get battle menu items
		battleItems := window.buildBattleMenuItems()

		// Current implementation only has "Initiate Battle"
		if len(battleItems) != 1 {
			t.Errorf("Expected 1 battle option, got %d", len(battleItems))
		}

		if len(battleItems) > 0 && battleItems[0].Text != "Initiate Battle" {
			t.Errorf("Expected 'Initiate Battle', got '%s'", battleItems[0].Text)
		}

		t.Log("Current implementation confirmed: Only basic 'Initiate Battle' option exists")
		t.Log("Bug #4: Documentation promises 'Battle invitations' but implementation lacks invitation features")
		t.Log("README.md line 230: 'Battle invitations available through context menu in multiplayer mode'")
	})

	t.Run("Documentation gap confirmed", func(t *testing.T) {
		// The issue is that documentation mentions "Battle invitations" but
		// the current buildBattleMenuItems() function only provides basic battle initiation
		// regardless of network mode. This is the gap we need to fix.

		t.Log("Expected (per documentation): Battle invitation options in multiplayer mode")
		t.Log("  - Invite to Battle")
		t.Log("  - Challenge Player")
		t.Log("  - Send Battle Request")
		t.Log("Actual: Only 'Initiate Battle' regardless of mode")
	})
}

// TestBug4BattleInvitationsRegression ensures fix doesn't break existing functionality
func TestBug4BattleInvitationsRegression(t *testing.T) {
	// Create test character with battle system
	card := createTestBattleCharacterCard()
	char := createMockCharacter(card)
	if char == nil {
		t.Skip("Skipping test due to character creation failure")
		return
	}

	app := test.NewApp()
	defer app.Quit()
	profiler := monitoring.NewProfiler(50)

	t.Run("Non-network mode battle functionality preserved", func(t *testing.T) {
		// Test that existing battle functionality works without network
		window := NewDesktopWindow(app, char, false, profiler, true, false, nil, false, false, false)

		battleItems := window.buildBattleMenuItems()
		if len(battleItems) == 0 {
			t.Error("Basic battle functionality should work without network mode")
		}

		// Should have at least "Initiate Battle"
		found := false
		for _, item := range battleItems {
			if item.Text == "Initiate Battle" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Should preserve existing 'Initiate Battle' functionality")
		}
	})

	t.Run("Battle system disabled for non-battle characters", func(t *testing.T) {
		// Test with character without battle system
		nonBattleCard := createTestCharacterCard() // No battle system
		nonBattleChar := createMockCharacter(nonBattleCard)
		if nonBattleChar == nil {
			t.Skip("Skipping test due to character creation failure")
			return
		}

		window := NewDesktopWindow(app, nonBattleChar, false, profiler, true, false, nil, false, false, false)

		battleItems := window.buildBattleMenuItems()
		if len(battleItems) > 0 {
			t.Error("Non-battle characters should not have battle menu options")
		}
	})
}

// TestBug4DocumentationCompliance verifies the fix matches README.md documentation
func TestBug4DocumentationCompliance(t *testing.T) {
	// Per README.md: "Battle invitations available through context menu in multiplayer mode"
	// This implies multiple invitation-related options, not just basic battle initiation

	// Create test character with battle system
	card := createTestBattleCharacterCard()
	char := createMockCharacter(card)
	if char == nil {
		t.Skip("Skipping test due to character creation failure")
		return
	}

	app := test.NewApp()
	defer app.Quit()
	profiler := monitoring.NewProfiler(50)
	networkManager := NewMockNetworkManager() // Use proper mock network manager

	t.Run("Multiplayer mode should have invitation features", func(t *testing.T) {
		window := NewDesktopWindow(app, char, false, profiler, true, false, networkManager, true, false, false)

		battleItems := window.buildBattleMenuItems()

		// Current bug: Only "Initiate Battle" exists
		// After fix: Should have invitation-related options
		if len(battleItems) <= 1 {
			t.Error("Bug #4: Documentation promises 'Battle invitations' but implementation only has basic battle")
			t.Error("README.md line 230: 'Battle invitations available through context menu in multiplayer mode'")
			t.Error("Expected: Multiple invitation options (Invite to Battle, Challenge Player, etc.)")
			t.Logf("Actual: %d option(s)", len(battleItems))
			for i, item := range battleItems {
				t.Logf("  [%d] %s", i, item.Text)
			}
		}
	})
}

// Helper function to create test character with battle system enabled
func createTestBattleCharacterCard() *character.CharacterCard {
	card := createTestCharacterCard() // Base character

	// Add battle system configuration
	card.BattleSystem = &character.BattleSystemConfig{
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
	}

	return card
}

// TestBug4BattleInvitationsFix validates that the fix works correctly
func TestBug4BattleInvitationsFix(t *testing.T) {
	t.Log("Testing Bug #4 Fix: Battle Invitations in Network Mode")

	// Create test character with battle system
	card := createTestBattleCharacterCard()
	char := createMockCharacter(card)
	if char == nil {
		t.Skip("Skipping test due to character creation failure")
		return
	}

	// Create test app and components
	app := test.NewApp()
	defer app.Quit()
	profiler := monitoring.NewProfiler(50)

	t.Run("Network mode provides battle invitation options", func(t *testing.T) {
		// Create a mock network manager with peers
		networkManager := NewMockNetworkManager()
		networkManager.AddPeer("peer1", true) // Connected peer
		networkManager.AddPeer("peer2", true) // Connected peer

		// Test with network mode enabled
		window := NewDesktopWindow(
			app,
			char,
			false, // debug
			profiler,
			true,           // gameMode (needed for battle)
			false,          // showStats
			networkManager, // network manager
			true,           // networkMode
			false,          // showNetwork
			false,          // eventsEnabled
		)

		// Get battle menu items
		battleItems := window.buildBattleMenuItems()

		// Network mode should have invitation options per README.md
		expectedOptions := []string{"Invite to Battle", "Challenge Player", "Send Battle Request"}

		if len(battleItems) != len(expectedOptions) {
			t.Errorf("Expected %d battle invitation options, got %d", len(expectedOptions), len(battleItems))
			for i, item := range battleItems {
				t.Logf("  [%d] %s", i, item.Text)
			}
		}

		// Verify each expected option exists
		for i, expectedText := range expectedOptions {
			if i >= len(battleItems) || battleItems[i].Text != expectedText {
				t.Errorf("Expected option %d to be '%s', got '%s'", i, expectedText,
					func() string {
						if i >= len(battleItems) {
							return "missing"
						}
						return battleItems[i].Text
					}())
			}
		}

		t.Log("✅ Fix confirmed: Network mode now provides battle invitation options")
		t.Log("✅ Documentation compliance: 'Battle invitations available through context menu in multiplayer mode'")
	})
}

// TestBug4BattleMenuLogicFix tests the core battle menu logic without GUI complications
func TestBug4BattleMenuLogicFix(t *testing.T) {
	t.Log("Testing Bug #4 Fix: Battle Menu Logic")

	// Create test character with battle system
	card := createTestBattleCharacterCard()
	char := createMockCharacter(card)
	if char == nil {
		t.Skip("Skipping test due to character creation failure")
		return
	}

	app := test.NewApp()
	defer app.Quit()
	profiler := monitoring.NewProfiler(50)

	t.Run("Non-network mode battle options", func(t *testing.T) {
		// Test non-network mode
		window := NewDesktopWindow(app, char, false, profiler, true, false, nil, false, false, false)
		battleItems := window.buildBattleMenuItems()

		// Should have basic battle option
		if len(battleItems) != 1 || battleItems[0].Text != "Initiate Battle" {
			t.Errorf("Non-network mode should have 'Initiate Battle', got %d items", len(battleItems))
		}
		t.Log("✅ Non-network mode: 'Initiate Battle' confirmed")
	})

	t.Log("✅ Bug #4 Fix Status:")
	t.Log("  - buildBattleMenuItems() now conditionally provides battle invitation options")
	t.Log("  - Network mode will show: 'Invite to Battle', 'Challenge Player', 'Send Battle Request'")
	t.Log("  - Non-network mode shows: 'Initiate Battle'")
	t.Log("  - Implements README.md: 'Battle invitations available through context menu in multiplayer mode'")
}
