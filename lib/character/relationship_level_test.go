package character

import (
	"testing"
	"time"
)

// TestRelationshipLevelSystem tests the complete relationship level functionality
func TestRelationshipLevelSystem(t *testing.T) {
	// Create a game state with romance stats
	statConfigs := map[string]StatConfig{
		"affection": {Initial: 0, Max: 100, DegradationRate: 0.1, CriticalThreshold: 10},
		"trust":     {Initial: 20, Max: 100, DegradationRate: 0.05, CriticalThreshold: 5},
		"intimacy":  {Initial: 0, Max: 100, DegradationRate: 0.2, CriticalThreshold: 0},
	}

	gameConfig := &GameConfig{
		StatsDecayInterval: time.Minute,
	}

	gs := NewGameState(statConfigs, gameConfig)

	// Test initial relationship level
	if gs.GetRelationshipLevel() != "Stranger" {
		t.Errorf("Expected initial relationship level 'Stranger', got '%s'", gs.GetRelationshipLevel())
	}

	// Create progression config matching the romance character
	progressionConfig := &ProgressionConfig{
		Levels: []LevelConfig{
			{
				Name:        "Stranger",
				Requirement: map[string]int64{"age": 0},
				Size:        128,
			},
			{
				Name: "Friend",
				Requirement: map[string]int64{
					"age":       86400, // 1 day
					"affection": 15,
					"trust":     10,
				},
				Size: 132,
			},
			{
				Name: "Close Friend",
				Requirement: map[string]int64{
					"age":       172800, // 2 days
					"affection": 30,
					"trust":     25,
				},
				Size: 136,
			},
			{
				Name: "Romantic Interest",
				Requirement: map[string]int64{
					"age":       259200, // 3 days
					"affection": 50,
					"trust":     40,
					"intimacy":  20,
				},
				Size: 140,
			},
		},
	}

	// Set progression system
	gs.SetProgression(progressionConfig)

	// Test that relationship level doesn't change without meeting requirements
	levelChanged := gs.UpdateRelationshipLevel(progressionConfig)
	if levelChanged {
		t.Error("Relationship level should not change without meeting requirements")
	}
	if gs.GetRelationshipLevel() != "Stranger" {
		t.Errorf("Should still be 'Stranger', got '%s'", gs.GetRelationshipLevel())
	}

	// Meet requirements for Friend level (affection and trust, but not age)
	gs.Stats["affection"].Current = 20
	gs.Stats["trust"].Current = 15

	levelChanged = gs.UpdateRelationshipLevel(progressionConfig)
	if levelChanged {
		t.Error("Should not advance to Friend without meeting age requirement")
	}

	// Advance age to meet Friend requirements
	gs.Progression.Age = time.Hour * 25 // More than 1 day

	levelChanged = gs.UpdateRelationshipLevel(progressionConfig)
	if !levelChanged {
		t.Error("Should advance to Friend level when all requirements are met")
	}
	if gs.GetRelationshipLevel() != "Friend" {
		t.Errorf("Expected 'Friend', got '%s'", gs.GetRelationshipLevel())
	}

	// Test advancing to Close Friend
	gs.Stats["affection"].Current = 35
	gs.Stats["trust"].Current = 30
	gs.Progression.Age = time.Hour * 49 // More than 2 days

	levelChanged = gs.UpdateRelationshipLevel(progressionConfig)
	if !levelChanged {
		t.Error("Should advance to Close Friend level")
	}
	if gs.GetRelationshipLevel() != "Close Friend" {
		t.Errorf("Expected 'Close Friend', got '%s'", gs.GetRelationshipLevel())
	}

	// Test advancing to Romantic Interest (requires intimacy)
	gs.Stats["affection"].Current = 55
	gs.Stats["trust"].Current = 45
	gs.Stats["intimacy"].Current = 25
	gs.Progression.Age = time.Hour * 73 // More than 3 days

	levelChanged = gs.UpdateRelationshipLevel(progressionConfig)
	if !levelChanged {
		t.Error("Should advance to Romantic Interest level")
	}
	if gs.GetRelationshipLevel() != "Romantic Interest" {
		t.Errorf("Expected 'Romantic Interest', got '%s'", gs.GetRelationshipLevel())
	}
}

// TestRomanceMemorySystem tests the romance interaction memory functionality
func TestRomanceMemorySystem(t *testing.T) {
	statConfigs := map[string]StatConfig{
		"affection": {Initial: 0, Max: 100, DegradationRate: 0.1, CriticalThreshold: 10},
		"happiness": {Initial: 50, Max: 100, DegradationRate: 0.5, CriticalThreshold: 20},
	}

	gs := NewGameState(statConfigs, nil)

	// Test initial state
	if len(gs.GetRomanceMemories()) != 0 {
		t.Error("Should start with no romance memories")
	}
	if gs.GetInteractionCount("compliment") != 0 {
		t.Error("Should start with no interactions")
	}

	// Record some romance interactions
	statsBefore := map[string]float64{"affection": 0, "happiness": 50}
	statsAfter := map[string]float64{"affection": 5, "happiness": 55}

	gs.RecordRomanceInteraction("compliment", "Thank you! 💕", statsBefore, statsAfter)

	// Verify memory was recorded
	memories := gs.GetRomanceMemories()
	if len(memories) != 1 {
		t.Errorf("Expected 1 memory, got %d", len(memories))
	}
	if memories[0].InteractionType != "compliment" {
		t.Errorf("Expected interaction type 'compliment', got '%s'", memories[0].InteractionType)
	}
	if memories[0].Response != "Thank you! 💕" {
		t.Errorf("Expected response 'Thank you! 💕', got '%s'", memories[0].Response)
	}

	// Verify interaction count
	if gs.GetInteractionCount("compliment") != 1 {
		t.Errorf("Expected 1 compliment interaction, got %d", gs.GetInteractionCount("compliment"))
	}

	// Record many more interactions to test memory limit
	for i := 0; i < 60; i++ {
		gs.RecordRomanceInteraction("test", "response", statsBefore, statsAfter)
	}

	// Should be limited to 50 memories
	memories = gs.GetRomanceMemories()
	if len(memories) > 50 {
		t.Errorf("Memory limit exceeded: expected max 50, got %d", len(memories))
	}
}

// TestRomanceStatsAccess tests romance-specific stat access methods
func TestRomanceStatsAccess(t *testing.T) {
	statConfigs := map[string]StatConfig{
		"hunger":    {Initial: 100, Max: 100, DegradationRate: 1.0, CriticalThreshold: 20},
		"affection": {Initial: 25, Max: 100, DegradationRate: 0.1, CriticalThreshold: 10},
		"trust":     {Initial: 30, Max: 100, DegradationRate: 0.05, CriticalThreshold: 5},
		"intimacy":  {Initial: 15, Max: 100, DegradationRate: 0.2, CriticalThreshold: 0},
		"jealousy":  {Initial: 5, Max: 100, DegradationRate: 2.0, CriticalThreshold: 80},
	}

	gs := NewGameState(statConfigs, nil)

	// Test romance stats access
	romanceStats := gs.GetRomanceStats()

	// Should only include romance stats
	expectedStats := map[string]float64{
		"affection": 25,
		"trust":     30,
		"intimacy":  15,
		"jealousy":  5,
	}

	for statName, expectedValue := range expectedStats {
		if value, exists := romanceStats[statName]; !exists {
			t.Errorf("Romance stat '%s' not found", statName)
		} else if value != expectedValue {
			t.Errorf("Expected %s=%f, got %f", statName, expectedValue, value)
		}
	}

	// Should not include non-romance stats
	if _, exists := romanceStats["hunger"]; exists {
		t.Error("Non-romance stat 'hunger' should not be included in romance stats")
	}
}

// TestInteractionHistoryAccess tests interaction history functionality
func TestInteractionHistoryAccess(t *testing.T) {
	gs := NewGameState(map[string]StatConfig{}, nil)

	// Record different types of interactions
	statsBefore := map[string]float64{}
	statsAfter := map[string]float64{}

	gs.RecordRomanceInteraction("compliment", "Thanks!", statsBefore, statsAfter)
	gs.RecordRomanceInteraction("compliment", "Sweet!", statsBefore, statsAfter)
	gs.RecordRomanceInteraction("give_gift", "A gift!", statsBefore, statsAfter)

	// Test interaction counts
	if gs.GetInteractionCount("compliment") != 2 {
		t.Errorf("Expected 2 compliments, got %d", gs.GetInteractionCount("compliment"))
	}
	if gs.GetInteractionCount("give_gift") != 1 {
		t.Errorf("Expected 1 gift, got %d", gs.GetInteractionCount("give_gift"))
	}
	if gs.GetInteractionCount("nonexistent") != 0 {
		t.Errorf("Expected 0 for nonexistent interaction, got %d", gs.GetInteractionCount("nonexistent"))
	}

	// Test full interaction history
	history := gs.GetInteractionHistory()
	if len(history["compliment"]) != 2 {
		t.Errorf("Expected 2 compliment entries in history, got %d", len(history["compliment"]))
	}
	if len(history["give_gift"]) != 1 {
		t.Errorf("Expected 1 give_gift entry in history, got %d", len(history["give_gift"]))
	}
}

// TestRelationshipLevelRequirements tests the requirement checking logic
func TestRelationshipLevelRequirements(t *testing.T) {
	statConfigs := map[string]StatConfig{
		"affection": {Initial: 20, Max: 100, DegradationRate: 0.1, CriticalThreshold: 10},
		"trust":     {Initial: 15, Max: 100, DegradationRate: 0.05, CriticalThreshold: 5},
	}

	gs := NewGameState(statConfigs, nil)

	// Test meeting requirements
	requirements := map[string]int64{
		"age":       3600, // 1 hour
		"affection": 15,
		"trust":     10,
	}

	// Should not meet age requirement
	if gs.meetsRelationshipRequirements(requirements, 1800) { // 30 minutes
		t.Error("Should not meet requirements with insufficient age")
	}

	// Should meet all requirements
	if !gs.meetsRelationshipRequirements(requirements, 7200) { // 2 hours
		t.Error("Should meet all requirements")
	}

	// Test with insufficient stats
	gs.Stats["affection"].Current = 10 // Below requirement
	if gs.meetsRelationshipRequirements(requirements, 7200) {
		t.Error("Should not meet requirements with insufficient affection")
	}

	// Test with missing stat
	requirementsWithMissingStat := map[string]int64{
		"age":         3600,
		"nonexistent": 50,
	}

	if gs.meetsRelationshipRequirements(requirementsWithMissingStat, 7200) {
		t.Error("Should not meet requirements when required stat doesn't exist")
	}
}
