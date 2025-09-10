package ui

import (
	"testing"

	"github.com/opd-ai/desktop-companion/lib/character"
)

// TestBug2NetworkContextMenuRegression validates that the network context menu fix continues to work
// This is a regression test for Bug #2: Missing Network Overlay Context Menu Access
func TestBug2NetworkContextMenuRegression(t *testing.T) {
	t.Run("NetworkModeEnabled", func(t *testing.T) {
		// Create a basic character
		card := &character.CharacterCard{
			Name: "Network Character",
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

		// Create DesktopWindow with network mode enabled
		dw := &DesktopWindow{
			character:   char,
			networkMode: true,
			showNetwork: true,
		}

		// Create network overlay (normally done in NewDesktopWindow)
		dw.networkOverlay = &NetworkOverlay{
			visible: false, // Initially hidden
		}

		// Get network menu items
		networkItems := dw.buildNetworkMenuItems()

		// Verify network menu items exist
		if len(networkItems) == 0 {
			t.Error("REGRESSION: Network menu items should be present when network mode enabled")
		}

		// Verify the network overlay option is present with correct text
		hasShowOption := false
		for _, item := range networkItems {
			if item.Text == "Show Network Overlay" {
				hasShowOption = true

				// Test the callback is not nil
				if item.Callback == nil {
					t.Error("Network overlay callback should not be nil")
				}
				break
			}
		}

		if !hasShowOption {
			t.Error("REGRESSION: 'Show Network Overlay' option should be present")
		}
	})

	t.Run("NetworkModeDisabled", func(t *testing.T) {
		// Create a basic character
		card := &character.CharacterCard{
			Name: "Non-Network Character",
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

		// Create DesktopWindow with network mode disabled
		dw := &DesktopWindow{
			character:   char,
			networkMode: false, // Network mode disabled
		}

		// Network overlay should be nil when network mode disabled
		if dw.networkOverlay != nil {
			// Even if present, buildNetworkMenuItems should return empty
		}

		// Get network menu items
		networkItems := dw.buildNetworkMenuItems()

		// Should return nil/empty when network mode is disabled
		if len(networkItems) > 0 {
			t.Error("Network menu items should be empty when network mode disabled")
		}
	})

	t.Run("NetworkOverlayVisible", func(t *testing.T) {
		// Create a basic character
		card := &character.CharacterCard{
			Name: "Network Character Visible",
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

		// Create DesktopWindow with network mode enabled
		dw := &DesktopWindow{
			character:   char,
			networkMode: true,
			showNetwork: true,
		}

		// Create network overlay that's visible
		dw.networkOverlay = &NetworkOverlay{
			visible: true, // Already visible
		}

		// Get network menu items
		networkItems := dw.buildNetworkMenuItems()

		// Verify the text changes when overlay is visible
		hasHideOption := false
		for _, item := range networkItems {
			if item.Text == "Hide Network Overlay" {
				hasHideOption = true
				break
			}
		}

		if !hasHideOption {
			t.Error("'Hide Network Overlay' option should be present when overlay is visible")
		}
	})

	t.Run("IntegrationWithFullContextMenu", func(t *testing.T) {
		// Create a basic character
		card := &character.CharacterCard{
			Name: "Full Menu Character",
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

		// Create DesktopWindow with network mode enabled
		dw := &DesktopWindow{
			character:   char,
			networkMode: true,
			gameMode:    true, // Also enable game mode
		}

		// Create network overlay
		dw.networkOverlay = &NetworkOverlay{
			visible: false,
		}

		// Get all menu items to verify integration
		basicItems := dw.buildBasicMenuItems()
		gameItems := dw.buildGameModeMenuItems()
		battleItems := dw.buildBattleMenuItems()
		chatItems := dw.buildChatMenuItems()
		networkItems := dw.buildNetworkMenuItems()
		utilityItems := dw.buildUtilityMenuItems()

		// Combine all items
		var allItems []ContextMenuItem
		allItems = append(allItems, basicItems...)
		allItems = append(allItems, gameItems...)
		allItems = append(allItems, battleItems...)
		allItems = append(allItems, chatItems...)
		allItems = append(allItems, networkItems...)
		allItems = append(allItems, utilityItems...)

		// Verify network option is present in full menu
		hasNetworkOption := false
		for _, item := range allItems {
			if item.Text == "Show Network Overlay" {
				hasNetworkOption = true
				break
			}
		}

		if !hasNetworkOption {
			t.Error("REGRESSION: Network option should be present in full context menu")
		}

		// Verify other standard options are still present
		standardOptions := []string{"Talk", "About", "Shortcuts"}
		for _, expected := range standardOptions {
			found := false
			for _, item := range allItems {
				if item.Text == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Standard option '%s' should be present", expected)
			}
		}
	})
}
