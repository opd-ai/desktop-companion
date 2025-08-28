package character

import (
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// TestRomanceEventsIntegration tests the romance events system integration
func TestRomanceEventsIntegration(t *testing.T) {
	// Get the project root dynamically
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(filename), "..", "..")
	romanceCardPath := filepath.Join(projectRoot, "assets", "characters", "romance", "character.json")

	// Load the actual romance character with enhanced events
	card, err := LoadCard(romanceCardPath)
	if err != nil {
		t.Fatalf("failed to load romance character: %v", err)
	}

	romanceAssetsPath := filepath.Join(projectRoot, "assets", "characters", "romance")
	char, err := New(card, romanceAssetsPath)
	if err != nil {
		t.Fatalf("failed to create character: %v", err)
	}

	t.Run("romance events loaded correctly", func(t *testing.T) {
		if len(char.card.RomanceEvents) == 0 {
			t.Fatal("expected romance events to be loaded from character JSON")
		}

		t.Logf("Loaded %d romance events", len(char.card.RomanceEvents))

		// Verify some key events exist
		eventNames := make(map[string]bool)
		for _, event := range char.card.RomanceEvents {
			eventNames[event.Name] = true
		}

		expectedEvents := []string{
			"Love Letter Memory",
			"Romantic Daydream",
			"Sweet Memory Flashback",
			"Relationship Milestone Reflection",
			"Frequent Compliment Appreciation",
			"Gift Memory Warmth",
			"Trust Building Reflection",
		}

		for _, expectedEvent := range expectedEvents {
			if !eventNames[expectedEvent] {
				t.Errorf("expected romance event '%s' not found", expectedEvent)
			}
		}
	})

	t.Run("memory-based conditions work", func(t *testing.T) {
		if char.gameState == nil {
			t.Fatal("expected game state to be initialized")
		}

		// Build up some interaction history
		char.gameState.ApplyInteractionEffects(map[string]float64{"affection": 25})
		char.gameState.RecordRomanceInteraction("compliment", "Thank you!",
			map[string]float64{"affection": 10},
			map[string]float64{"affection": 35})
		char.gameState.RecordRomanceInteraction("compliment", "So sweet!",
			map[string]float64{"affection": 35},
			map[string]float64{"affection": 40})

		// Test memory-based conditions
		conditions := map[string]map[string]float64{
			"memoryCount": {"recent_positive_min": 1},
		}

		canSatisfy := char.gameState.CanSatisfyRomanceRequirements(conditions)
		if !canSatisfy {
			t.Error("expected to satisfy memory-based conditions")
		}
	})

	t.Run("romance events system functioning", func(t *testing.T) {
		// Set up favorable conditions
		char.gameState.ApplyInteractionEffects(map[string]float64{"affection": 20})

		// Reset event timing to allow immediate checks
		char.lastRomanceEventCheck = time.Now().Add(-60 * time.Second)
		char.romanceEventCooldowns = make(map[string]time.Time)

		// Try to trigger romance events (probability-based, so might not trigger)
		triggeredEvent := char.checkAndTriggerRomanceEvent(time.Minute)

		// Just verify the system doesn't crash and can return events
		if triggeredEvent != nil {
			t.Logf("Successfully triggered romance event: %s - %s",
				triggeredEvent.Name, triggeredEvent.Description)

			if len(triggeredEvent.Responses) == 0 {
				t.Error("triggered event should have responses")
			}
		} else {
			t.Log("No romance event triggered this time (probability-based)")
		}
	})

	t.Run("enhanced condition checking", func(t *testing.T) {
		// Test relationship level conditions
		char.gameState.UpdateRelationshipLevel(char.card.Progression)
		currentLevel := char.gameState.GetRelationshipLevel()
		t.Logf("Current relationship level: %s", currentLevel)

		// Test interaction count based conditions
		interactionCount := char.gameState.GetInteractionCount("compliment")
		t.Logf("Compliment interaction count: %d", interactionCount)

		// Test memory count
		memoryCount := len(char.gameState.GetRomanceMemories())
		t.Logf("Romance memory count: %d", memoryCount)

		if memoryCount == 0 {
			t.Error("expected some romance memories to be recorded")
		}
	})
}
