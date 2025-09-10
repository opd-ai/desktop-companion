package character

import (
	"testing"
	"time"
)

// TestRelationshipProgressionIntegration demonstrates the complete relationship level system
// This test validates Phase 3 Task 1: Relationship Progression implementation
func TestRelationshipProgressionIntegration(t *testing.T) {
	// Load the romance character card
	card, err := LoadCard("../../assets/characters/romance/character.json")
	if err != nil {
		t.Fatalf("Failed to load romance character: %v", err)
	}

	// Disable news features for this test to avoid network timeouts
	if card.NewsFeatures != nil {
		card.NewsFeatures.Enabled = false
	}

	// Create character instance
	character, err := New(card, "../../assets/characters/romance")
	if err != nil {
		t.Fatalf("Failed to create character: %v", err)
	}

	// Enable game mode to activate romance features
	err = character.EnableGameMode(nil, "")
	if err != nil {
		t.Fatalf("Failed to enable game mode: %v", err)
	}

	gameState := character.GetGameState()
	if gameState == nil {
		t.Fatal("Game state should not be nil after enabling game mode")
	}

	t.Logf("Initial relationship level: %s", gameState.GetRelationshipLevel())

	// Verify initial state
	if gameState.GetRelationshipLevel() != "Stranger" {
		t.Errorf("Expected initial relationship level 'Stranger', got '%s'", gameState.GetRelationshipLevel())
	}

	// Check initial romance stats
	romanceStats := gameState.GetRomanceStats()
	t.Logf("Initial romance stats: %+v", romanceStats)

	if romanceStats["affection"] != 0 {
		t.Errorf("Expected initial affection=0, got %f", romanceStats["affection"])
	}
	if romanceStats["trust"] != 20 {
		t.Errorf("Expected initial trust=20, got %f", romanceStats["trust"])
	}

	// Simulate romance interactions to build relationship
	t.Log("\n=== Starting Romance Interaction Simulation ===")

	// Perform multiple compliments to build affection and trust
	for i := 0; i < 5; i++ {
		response := character.HandleRomanceInteraction("compliment")
		t.Logf("Compliment %d response: %s", i+1, response)

		// Wait for cooldown to expire (reduced for automated testing)
		time.Sleep(time.Millisecond * 100)

		stats := gameState.GetRomanceStats()
		level := gameState.GetRelationshipLevel()
		t.Logf("After compliment %d - Affection: %.1f, Trust: %.1f, Level: %s",
			i+1, stats["affection"], stats["trust"], level)

		// Check if compliment was successful (response should not be a failure message)
		if response == "" || response == "Maybe when I know you better?" {
			t.Logf("Compliment %d failed requirements or hit cooldown", i+1)
		}
	}

	// Check interaction history
	complimentCount := gameState.GetInteractionCount("compliment")
	t.Logf("Total compliments given: %d", complimentCount)

	// Should have successful compliments recorded
	if complimentCount == 0 {
		t.Error("No compliments were successfully recorded")
	}

	// Simulate time passing and age progression to meet Friend requirements
	if gameState.Progression != nil {
		// Advance age to meet Friend level requirement (1 day)
		gameState.Progression.Age = time.Hour * 25 // More than 1 day

		// Check if we can advance to Friend level
		levelChanged := gameState.UpdateRelationshipLevel(card.Progression)

		currentStats := gameState.GetRomanceStats()
		currentLevel := gameState.GetRelationshipLevel()

		t.Logf("After age advancement - Affection: %.1f, Trust: %.1f, Age: %v, Level: %s, Changed: %v",
			currentStats["affection"], currentStats["trust"], gameState.Progression.Age, currentLevel, levelChanged)

		// Check if we meet Friend requirements
		friendReq := map[string]int64{"age": 86400, "affection": 15, "trust": 10}
		canBecomeFriend := gameState.meetsRelationshipRequirements(friendReq, int64(gameState.Progression.Age.Seconds()))
		t.Logf("Can become Friend? %v (affection>=15: %v, trust>=10: %v, age>=1day: %v)",
			canBecomeFriend,
			currentStats["affection"] >= 15,
			currentStats["trust"] >= 10,
			gameState.Progression.Age >= time.Hour*24)
	}

	// Perform deep conversations to build trust (since gift requires higher affection)
	for i := 0; i < 3; i++ {
		response := character.HandleRomanceInteraction("deep_conversation")
		t.Logf("Deep conversation %d response: %s", i+1, response)

		time.Sleep(time.Millisecond * 100)

		stats := gameState.GetRomanceStats()
		level := gameState.GetRelationshipLevel()
		t.Logf("After conversation %d - Affection: %.1f, Trust: %.1f, Intimacy: %.1f, Level: %s",
			i+1, stats["affection"], stats["trust"], stats["intimacy"], level)

		// Check if conversation was successful
		if response == "" || response == "I'm not ready for such deep talks yet." {
			t.Logf("Conversation %d failed requirements or hit cooldown", i+1)
		}
	}

	// Now try gift giving once affection is higher
	currentStats := gameState.GetRomanceStats()
	t.Logf("Before gift attempt - Affection: %.1f (need 15)", currentStats["affection"])

	response := character.HandleRomanceInteraction("give_gift")
	t.Logf("Gift response: %s", response)

	if response != "" && response != "Perhaps when we're closer?" {
		stats := gameState.GetRomanceStats()
		level := gameState.GetRelationshipLevel()
		t.Logf("After gift - Affection: %.1f, Trust: %.1f, Level: %s",
			stats["affection"], stats["trust"], level)
	}

	// Force relationship level update
	if card.Progression != nil {
		levelChanged := gameState.UpdateRelationshipLevel(card.Progression)
		finalLevel := gameState.GetRelationshipLevel()
		finalStats := gameState.GetRomanceStats()

		t.Logf("Final state - Level: %s (changed: %v), Affection: %.1f, Trust: %.1f, Intimacy: %.1f",
			finalLevel, levelChanged, finalStats["affection"], finalStats["trust"], finalStats["intimacy"])
	}

	// Test romance memory system
	memories := gameState.GetRomanceMemories()
	t.Logf("Romance memories recorded: %d", len(memories))

	if len(memories) > 0 {
		lastMemory := memories[len(memories)-1]
		t.Logf("Last memory: %s -> %s", lastMemory.InteractionType, lastMemory.Response)
	}

	// Test interaction history
	history := gameState.GetInteractionHistory()
	t.Logf("Interaction history: %+v", history)

	// Verify that romance features are working
	if !card.HasRomanceFeatures() {
		t.Error("Romance character should have romance features")
	}

	t.Log("\n=== Romance System Integration Test Complete ===")
}

// TestRelationshipLevelProgression tests the specific progression mechanics
func TestRelationshipLevelProgression(t *testing.T) {
	// Create a simplified test scenario
	statConfigs := map[string]StatConfig{
		"affection": {Initial: 0, Max: 100, DegradationRate: 0.1, CriticalThreshold: 10},
		"trust":     {Initial: 20, Max: 100, DegradationRate: 0.05, CriticalThreshold: 5},
		"intimacy":  {Initial: 0, Max: 100, DegradationRate: 0.2, CriticalThreshold: 0},
	}

	gameState := NewGameState(statConfigs, nil)

	// Create progression config matching the romance character
	progressionConfig := &ProgressionConfig{
		Levels: []LevelConfig{
			{Name: "Stranger", Requirement: map[string]int64{"age": 0}, Size: 128},
			{Name: "Friend", Requirement: map[string]int64{"age": 86400, "affection": 15, "trust": 10}, Size: 132},
			{Name: "Close Friend", Requirement: map[string]int64{"age": 172800, "affection": 30, "trust": 25}, Size: 136},
			{Name: "Romantic Interest", Requirement: map[string]int64{"age": 259200, "affection": 50, "trust": 40, "intimacy": 20}, Size: 140},
		},
	}

	gameState.SetProgression(progressionConfig)

	// Test progression through each level
	levels := []struct {
		name          string
		affection     float64
		trust         float64
		intimacy      float64
		age           time.Duration
		shouldAdvance bool
	}{
		{"Stranger", 5, 15, 0, time.Hour * 12, false},           // Not enough affection or age
		{"Friend", 20, 15, 0, time.Hour * 25, true},             // Meets Friend requirements
		{"Close Friend", 35, 30, 5, time.Hour * 49, true},       // Meets Close Friend requirements
		{"Romantic Interest", 55, 45, 25, time.Hour * 73, true}, // Meets Romantic Interest requirements
	}

	for _, test := range levels {
		// Set stats
		gameState.Stats["affection"].Current = test.affection
		gameState.Stats["trust"].Current = test.trust
		gameState.Stats["intimacy"].Current = test.intimacy

		// Set age
		if gameState.Progression != nil {
			gameState.Progression.Age = test.age
		}

		// Try to update relationship level
		levelChanged := gameState.UpdateRelationshipLevel(progressionConfig)
		currentLevel := gameState.GetRelationshipLevel()

		t.Logf("Test case '%s': affection=%.1f, trust=%.1f, intimacy=%.1f, age=%v -> level='%s', changed=%v",
			test.name, test.affection, test.trust, test.intimacy, test.age, currentLevel, levelChanged)

		if test.shouldAdvance && !levelChanged {
			t.Errorf("Expected level to advance to '%s' but it didn't change from '%s'", test.name, currentLevel)
		}

		if levelChanged && currentLevel != test.name {
			t.Logf("Level advanced to '%s' (expected '%s')", currentLevel, test.name)
		}
	}
}

// TestRomanceInteractionFlow tests the complete interaction flow with relationship progression
func TestRomanceInteractionFlow(t *testing.T) {
	// Create a character card with romance features for testing
	card := &CharacterCard{
		Name: "Test Romance Character",
		Animations: map[string]string{
			"idle":     "../default/animations/idle.gif",
			"talking":  "../default/animations/talking.gif",
			"happy":    "../default/animations/happy.gif",
			"blushing": "../default/animations/happy.gif",
		},
		Stats: map[string]StatConfig{
			"affection": {Initial: 0, Max: 100, DegradationRate: 0.1, CriticalThreshold: 10},
			"trust":     {Initial: 20, Max: 100, DegradationRate: 0.05, CriticalThreshold: 5},
		},
		Interactions: map[string]InteractionConfig{
			"compliment": {
				Triggers:     []string{"hover"},
				Effects:      map[string]float64{"affection": 5, "trust": 1},
				Animations:   []string{"blushing"},
				Responses:    []string{"Thank you!"},
				Cooldown:     30,
				Requirements: map[string]map[string]float64{"trust": {"min": 5}},
			},
		},
		Progression: &ProgressionConfig{
			Levels: []LevelConfig{
				{Name: "Stranger", Requirement: map[string]int64{"age": 0}, Size: 128},
				{Name: "Friend", Requirement: map[string]int64{"affection": 10, "trust": 15}, Size: 132},
			},
		},
		Personality: &PersonalityConfig{
			Traits: map[string]float64{
				"shyness":     0.6,
				"romanticism": 0.8,
			},
		},
		GameRules: &GameRulesConfig{
			StatsDecayInterval:             60,
			AutoSaveInterval:               300,
			CriticalStateAnimationPriority: true,
			MoodBasedAnimations:            true,
		},
	}

	character, err := New(card, "../../assets/characters/romance")
	if err != nil {
		t.Fatalf("Failed to create character: %v", err)
	}

	err = character.EnableGameMode(nil, "")
	if err != nil {
		t.Fatalf("Failed to enable game mode: %v", err)
	}

	gameState := character.GetGameState()

	// Test initial state
	if gameState.GetRelationshipLevel() != "Stranger" {
		t.Errorf("Expected initial level 'Stranger', got '%s'", gameState.GetRelationshipLevel())
	}

	// Test romance interaction
	response := character.HandleRomanceInteraction("compliment")
	if response == "" {
		t.Error("Expected response from romance interaction")
	}

	t.Logf("Compliment response: %s", response)

	// Check stats changed
	stats := gameState.GetRomanceStats()
	if stats["affection"] <= 0 {
		t.Error("Affection should have increased")
	}

	// Check memory was recorded
	memories := gameState.GetRomanceMemories()
	if len(memories) != 1 {
		t.Errorf("Expected 1 memory, got %d", len(memories))
	}

	// Test interaction count
	count := gameState.GetInteractionCount("compliment")
	if count != 1 {
		t.Errorf("Expected 1 compliment interaction, got %d", count)
	}

	t.Logf("Romance interaction flow test completed successfully")
}
