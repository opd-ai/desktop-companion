package ui

import (
	"testing"

	"desktop-companion/internal/character"
)

// TestBug2NetworkContextMenuFix validates the fix for missing network overlay context menu access
// This test verifies that when network mode is enabled, "Network Overlay" option appears in context menu
// This was Bug #2 in AUDIT.md: Missing Network Overlay Context Menu Access
func TestBug2NetworkContextMenuFix(t *testing.T) {
	// Create a basic character
	card := &character.CharacterCard{
		Name: "Test Character",
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

	// Manually create network overlay (normally done in NewDesktopWindow)
	// For testing, we'll create a minimal mock that satisfies the interface
	dw.networkOverlay = &NetworkOverlay{
		visible: false, // Initially hidden
	}

	// Verify preconditions
	if !dw.networkMode {
		t.Fatal("Network mode should be enabled")
	}

	if dw.networkOverlay == nil {
		t.Fatal("Network overlay should be available")
	}

	// Check if there's a buildNetworkMenuItems function (there should be after fix)
	// Get all context menu items by examining individual builders
	basicItems := dw.buildBasicMenuItems()
	gameItems := dw.buildGameModeMenuItems()
	battleItems := dw.buildBattleMenuItems()
	chatItems := dw.buildChatMenuItems()
	networkItems := dw.buildNetworkMenuItems() // This should now exist
	utilityItems := dw.buildUtilityMenuItems()

	// Combine all items (this is what showContextMenu does)
	var allItems []ContextMenuItem
	allItems = append(allItems, basicItems...)
	allItems = append(allItems, gameItems...)
	allItems = append(allItems, battleItems...)
	allItems = append(allItems, chatItems...)
	allItems = append(allItems, networkItems...)
	allItems = append(allItems, utilityItems...)

	// Verify that network overlay option IS present (confirming the fix works)
	var hasNetworkOption bool
	var networkOptionText string
	for _, item := range allItems {
		if item.Text == "Show Network Overlay" || item.Text == "Hide Network Overlay" || item.Text == "Network Overlay" {
			hasNetworkOption = true
			networkOptionText = item.Text
			break
		}
	}

	// This assertion should FAIL when bug exists, PASS when bug is fixed
	if !hasNetworkOption {
		t.Error("EXPECTED: Network Overlay option should be found in menu after fix")
	} else {
		t.Logf("SUCCESS: Found network option: %s", networkOptionText)
	}

	// Verify that other expected options are present
	var hasTalkOption bool
	for _, item := range allItems {
		if item.Text == "Talk" {
			hasTalkOption = true
			break
		}
	}

	if !hasTalkOption {
		t.Error("Expected 'Talk' option should be present in basic menu items")
	}

	t.Logf("Current menu items: %d total", len(allItems))
	for _, item := range allItems {
		t.Logf("  - %s", item.Text)
	}
}
