package ui

import (
	"testing"

	"github.com/opd-ai/desktop-companion/lib/character"
)

// TestFeature4DebugRomanceHistory is a simple test to debug the romance history feature
func TestFeature4DebugRomanceHistory(t *testing.T) {
	// Create a character with personality (which makes it have romance features)
	// AND with stats (which initializes gameState)
	card := &character.CharacterCard{
		Name: "TestCharacter",
		Personality: &character.PersonalityConfig{
			Traits: map[string]float64{
				"affection": 50.0,
			},
		},
		Stats: map[string]character.StatConfig{
			"affection": {
				Initial:           50.0,
				Max:               100.0,
				DegradationRate:   0.0,
				CriticalThreshold: 10.0,
			},
		},
	}

	char, err := character.New(card, "/tmp")
	if err != nil {
		t.Fatalf("Failed to create character: %v", err)
	}

	// Check if character has romance features
	hasRomanceFeatures := char.GetCard().HasRomanceFeatures()
	t.Logf("HasRomanceFeatures: %v", hasRomanceFeatures)

	// Check game state
	gameState := char.GetGameState()
	t.Logf("GameState is nil: %v", gameState == nil)

	if gameState != nil {
		// Check initial romance memories
		initialMemories := gameState.GetRomanceMemories()
		t.Logf("Initial romance memories count: %d", len(initialMemories))

		// Add a romance memory
		gameState.RecordRomanceInteraction(
			"compliment",
			"Thank you!",
			map[string]float64{"affection": 50.0},
			map[string]float64{"affection": 55.0},
		)

		// Check memories after adding
		afterMemories := gameState.GetRomanceMemories()
		t.Logf("Romance memories after adding: %d", len(afterMemories))

		// Test shouldShowRomanceHistory manually
		window := &DesktopWindow{character: char}
		shouldShow := window.shouldShowRomanceHistory()
		t.Logf("shouldShowRomanceHistory: %v", shouldShow)
	}
}
