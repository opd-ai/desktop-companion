package character

import (
	"fmt"
	"testing"
	"time"
)

// TestRomanceConfigValidation tests that all romance configurations are valid
func TestRomanceConfigValidation(t *testing.T) {
	tests := []struct {
		name     string
		card     CharacterCard
		wantErr  bool
		errorMsg string
	}{
		{
			name:    "valid comprehensive romance character",
			card:    createValidRomanceCard(),
			wantErr: false,
		},
		{
			name:     "invalid personality trait out of bounds",
			card:     createInvalidPersonalityCard(),
			wantErr:  true,
			errorMsg: "personality trait",
		},
		{
			name:     "invalid compatibility modifier too high",
			card:     createInvalidCompatibilityCard(),
			wantErr:  true,
			errorMsg: "compatibility modifier",
		},
		{
			name:     "invalid romance dialog requirements",
			card:     createInvalidRomanceDialogCard(),
			wantErr:  true,
			errorMsg: "romance dialog",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.card.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error containing '%s', got none", tt.errorMsg)
				} else if !stringContains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got '%v'", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestRomanceInteractionLogic tests romance interaction processing without animation files
func TestRomanceInteractionLogic(t *testing.T) {
	card := createValidRomanceCard()

	// Create character manually to bypass animation loading
	character := &Character{
		card:                     &card,
		currentState:             "idle",
		lastStateChange:          time.Now(),
		lastInteraction:          time.Now(),
		dialogCooldowns:          make(map[string]time.Time),
		gameInteractionCooldowns: make(map[string]time.Time),
		romanceEventCooldowns:    make(map[string]time.Time),
		lastRomanceEventCheck:    time.Now().Add(-30 * time.Second),
		idleTimeout:              time.Duration(card.Behavior.IdleTimeout) * time.Second,
		movementEnabled:          card.Behavior.MovementEnabled,
		size:                     card.Behavior.DefaultSize,
	}

	// Initialize game features
	character.initializeGameFeatures()

	if character.gameState == nil {
		t.Fatal("Game state should be initialized for romance character")
	}

	// Test romance interaction without animation dependency
	response := character.HandleRomanceInteraction("compliment")
	if response == "" {
		t.Error("Expected non-empty response from romance interaction")
	}

	// Test cooldown functionality
	response2 := character.HandleRomanceInteraction("compliment")
	if response2 != "" {
		t.Error("Expected empty response due to cooldown")
	}

	// Test stat effects
	affection := character.gameState.GetStat("affection")
	if affection == 0 {
		t.Error("Expected affection to increase after compliment")
	}

	t.Logf("Romance interaction test passed - affection: %.1f, response: %q", affection, response)
}

// TestRomanceMemorySystemBounds tests romance memory management limits
func TestRomanceMemorySystemBounds(t *testing.T) {
	card := createValidRomanceCard()

	// Create minimal character for testing
	character := &Character{
		card:                     &card,
		gameInteractionCooldowns: make(map[string]time.Time),
	}
	character.initializeGameFeatures()

	if character.gameState == nil {
		t.Fatal("Game state should be initialized")
	}

	// Test memory limit enforcement (should cap at 50 memories)
	for i := 0; i < 60; i++ {
		character.gameState.RecordRomanceInteraction(
			"test_interaction",
			fmt.Sprintf("Response %d", i),
			map[string]float64{"affection": 10},
			map[string]float64{"affection": 15},
		)
	}

	if len(character.gameState.RomanceMemories) > 50 {
		t.Errorf("Romance memories should be limited to 50, got %d", len(character.gameState.RomanceMemories))
	}

	// Verify newest memories are kept
	latestMemory := character.gameState.RomanceMemories[len(character.gameState.RomanceMemories)-1]
	if latestMemory.Response != "Response 59" {
		t.Error("Latest memory should be preserved")
	}

	t.Logf("Memory bounds test passed - memories capped at %d", len(character.gameState.RomanceMemories))
}

// BenchmarkRomanceMemoryOperations benchmarks memory system performance
func BenchmarkRomanceMemoryOperations(b *testing.B) {
	card := createValidRomanceCard()
	character := &Character{
		card:                     &card,
		gameInteractionCooldowns: make(map[string]time.Time),
	}
	character.initializeGameFeatures()

	if character.gameState == nil {
		b.Fatal("Game state should be initialized")
	}

	// Benchmark memory recording performance
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		character.gameState.RecordRomanceInteraction(
			"benchmark_interaction",
			"Test response",
			map[string]float64{"affection": 10},
			map[string]float64{"affection": 15},
		)
	}
}

// BenchmarkPersonalityCalculations benchmarks personality modifier calculations
func BenchmarkPersonalityCalculations(b *testing.B) {
	card := createValidRomanceCard()
	character := &Character{
		card: &card,
	}

	interactions := []string{"compliment", "gift", "deep_conversation", "romantic_gesture"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		interaction := interactions[i%len(interactions)]
		character.calculatePersonalityModifier(interaction)
	}
}

// TestRomanceStressTest performs basic stress testing without heavy memory usage
func TestRomanceStressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	card := createValidRomanceCard()
	character := &Character{
		card:                     &card,
		gameInteractionCooldowns: make(map[string]time.Time),
		romanceEventCooldowns:    make(map[string]time.Time),
	}
	character.initializeGameFeatures()

	// Perform 1000 operations quickly
	for i := 0; i < 1000; i++ {
		// Test different operations
		switch i % 5 {
		case 0:
			character.gameState.GetRelationshipLevel()
		case 1:
			character.gameState.UpdateRelationshipLevel(card.Progression)
		case 2:
			character.gameState.GetStats()
		case 3:
			character.calculatePersonalityModifier("compliment")
		case 4:
			// Reset cooldowns occasionally to allow interactions
			if i%100 == 0 {
				character.gameInteractionCooldowns = make(map[string]time.Time)
			}
			character.HandleRomanceInteraction("compliment")
		}
	}

	// Verify system state remains healthy
	if character.gameState.GetRelationshipLevel() == "" {
		t.Error("Relationship level was lost during stress test")
	}

	stats := character.gameState.GetStats()
	if len(stats) == 0 {
		t.Error("Stats were lost during stress test")
	}

	t.Logf("Stress test completed - 1000 operations successful")
}

// Helper functions for creating test character cards

func createValidRomanceCard() CharacterCard {
	return CharacterCard{
		Name:        "Valid Romance Test",
		Description: "Test character with valid romance features",
		Animations: map[string]string{
			"idle":    "test.gif",
			"talking": "test.gif",
		},
		Dialogs: []Dialog{
			{Trigger: "click", Responses: []string{"Hello!"}, Animation: "talking"},
		},
		Behavior: Behavior{IdleTimeout: 30, DefaultSize: 128},
		Stats: map[string]StatConfig{
			"affection": {Initial: 0, Max: 100, DegradationRate: 0.1, CriticalThreshold: 10},
			"trust":     {Initial: 20, Max: 100, DegradationRate: 0.05, CriticalThreshold: 5},
			"intimacy":  {Initial: 0, Max: 100, DegradationRate: 0.2, CriticalThreshold: 0},
			"jealousy":  {Initial: 0, Max: 100, DegradationRate: 2.0, CriticalThreshold: 80},
		},
		Interactions: map[string]InteractionConfig{
			"compliment": {
				Triggers:  []string{"shift+click"},
				Effects:   map[string]float64{"affection": 5, "trust": 1},
				Responses: []string{"Thank you! 💕"},
				Cooldown:  5, // Short cooldown for testing
			},
		},
		Personality: &PersonalityConfig{
			Traits: map[string]float64{
				"shyness":     0.6,
				"romanticism": 0.8,
			},
			Compatibility: map[string]float64{
				"gift_appreciation": 1.5,
			},
		},
		Progression: &ProgressionConfig{
			Levels: []LevelConfig{
				{
					Name:        "Stranger",
					Requirement: map[string]int64{"affection": 0},
					Size:        128,
				},
				{
					Name:        "Friend",
					Requirement: map[string]int64{"affection": 15, "trust": 10},
					Size:        132,
				},
			},
		},
	}
}

func createInvalidPersonalityCard() CharacterCard {
	card := createValidRomanceCard()
	card.Personality.Traits["shyness"] = 1.5 // Invalid: > 1.0
	return card
}

func createInvalidCompatibilityCard() CharacterCard {
	card := createValidRomanceCard()
	card.Personality.Compatibility["gift_appreciation"] = 6.0 // Invalid: > 5.0
	return card
}

func createInvalidRomanceDialogCard() CharacterCard {
	card := createValidRomanceCard()
	card.RomanceDialogs = []DialogExtended{
		{
			Dialog: Dialog{
				Trigger:   "click",
				Responses: []string{"Hello sweetheart!"},
				Animation: "talking",
			},
			Requirements: &RomanceRequirement{
				Stats: map[string]map[string]float64{
					"nonexistent_stat": {"min": 30}, // Invalid stat reference
				},
			},
		},
	}
	return card
}

// stringContains checks if a string contains a substring (helper for tests)
func stringContains(s, substr string) bool {
	return len(s) >= len(substr) && s[len(s)-len(substr):] == substr ||
		len(substr) == 0 ||
		(len(s) > len(substr) && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
