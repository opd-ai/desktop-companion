package character

import (
	"encoding/json"
	"math"
	"testing"
	"time"
)

// floatEquals compares two float64 values with a small tolerance for floating point precision
func floatEquals(a, b float64) bool {
	return math.Abs(a-b) < 0.0001
}

// TestNewGameState verifies GameState creation and initialization
func TestNewGameState(t *testing.T) {
	config := map[string]StatConfig{
		"hunger": {
			Initial:           100,
			Max:               100,
			DegradationRate:   1.0,
			CriticalThreshold: 20,
		},
		"happiness": {
			Initial:           80,
			Max:               100,
			DegradationRate:   0.5,
			CriticalThreshold: 15,
		},
	}

	gameConfig := &GameConfig{
		StatsDecayInterval:             time.Minute,
		CriticalStateAnimationPriority: true,
		MoodBasedAnimations:            true,
	}

	gs := NewGameState(config, gameConfig)

	// Verify initialization
	if gs == nil {
		t.Fatal("NewGameState returned nil")
	}

	if len(gs.Stats) != 2 {
		t.Errorf("Expected 2 stats, got %d", len(gs.Stats))
	}

	// Check hunger stat
	hungerStat := gs.Stats["hunger"]
	if hungerStat.Current != 100 {
		t.Errorf("Expected hunger current to be 100, got %f", hungerStat.Current)
	}
	if hungerStat.Max != 100 {
		t.Errorf("Expected hunger max to be 100, got %f", hungerStat.Max)
	}
	if hungerStat.DegradationRate != 1.0 {
		t.Errorf("Expected hunger degradation rate to be 1.0, got %f", hungerStat.DegradationRate)
	}

	// Check happiness stat
	happinessStat := gs.Stats["happiness"]
	if happinessStat.Current != 80 {
		t.Errorf("Expected happiness current to be 80, got %f", happinessStat.Current)
	}

	// Verify config
	if gs.Config == nil {
		t.Error("Expected config to be set")
	}
	if gs.Config.StatsDecayInterval != time.Minute {
		t.Errorf("Expected decay interval to be 1 minute, got %v", gs.Config.StatsDecayInterval)
	}
}

// TestGameStateUpdate verifies time-based stat degradation
func TestGameStateUpdate(t *testing.T) {
	config := map[string]StatConfig{
		"hunger": {
			Initial:           100,
			Max:               100,
			DegradationRate:   60.0, // 1 point per second for fast testing
			CriticalThreshold: 20,
		},
	}

	gs := NewGameState(config, &GameConfig{StatsDecayInterval: time.Second})

	// Fast-forward time by setting last update to past
	gs.LastDecayUpdate = time.Now().Add(-2 * time.Second)

	// Update should apply 2 seconds worth of degradation
	triggeredStates := gs.Update(time.Second)

	// Check degradation applied
	hunger := gs.GetStat("hunger")
	expectedHunger := 100.0 - (60.0 * (2.0 / 60.0)) // 60 points/min * 2 seconds = 2 points
	if !floatEquals(hunger, expectedHunger) {
		t.Errorf("Expected hunger to be %f after degradation, got %f", expectedHunger, hunger)
	}

	// No critical states should be triggered yet
	if len(triggeredStates) != 0 {
		t.Errorf("Expected no triggered states, got %v", triggeredStates)
	}
}

// TestGameStateCriticalStates verifies critical state detection
func TestGameStateCriticalStates(t *testing.T) {
	config := map[string]StatConfig{
		"hunger": {
			Initial:           25,
			Max:               100,
			DegradationRate:   60.0, // Fast degradation for testing
			CriticalThreshold: 20,
		},
	}

	gs := NewGameState(config, &GameConfig{StatsDecayInterval: time.Second})

	// Fast-forward time to trigger critical state
	gs.LastDecayUpdate = time.Now().Add(-10 * time.Second)

	triggeredStates := gs.Update(time.Second)

	// Should trigger hungry and critical states
	if len(triggeredStates) == 0 {
		t.Error("Expected critical states to be triggered")
	}

	// Check that we got the expected states
	foundHungry := false
	foundCritical := false
	for _, state := range triggeredStates {
		if state == "hungry" {
			foundHungry = true
		}
		if state == "hunger_critical" {
			foundCritical = true
		}
	}

	if !foundHungry {
		t.Error("Expected 'hungry' state to be triggered")
	}
	if !foundCritical {
		t.Error("Expected 'hunger_critical' state to be triggered")
	}

	// Verify GetCriticalStates works
	criticalStates := gs.GetCriticalStates()
	if len(criticalStates) == 0 {
		t.Error("Expected critical states to be returned")
	}
	if criticalStates[0] != "hunger" {
		t.Errorf("Expected 'hunger' in critical states, got %v", criticalStates)
	}
}

// TestApplyInteractionEffects verifies stat modification from interactions
func TestApplyInteractionEffects(t *testing.T) {
	config := map[string]StatConfig{
		"hunger":    {Initial: 50, Max: 100, DegradationRate: 1.0, CriticalThreshold: 20},
		"happiness": {Initial: 30, Max: 100, DegradationRate: 0.5, CriticalThreshold: 15},
	}

	gs := NewGameState(config, nil)

	// Test positive effects
	effects := map[string]float64{
		"hunger":    25, // Feed the pet
		"happiness": 15, // Make it happy
	}

	gs.ApplyInteractionEffects(effects)

	// Verify changes applied
	if gs.GetStat("hunger") != 75 {
		t.Errorf("Expected hunger to be 75, got %f", gs.GetStat("hunger"))
	}
	if gs.GetStat("happiness") != 45 {
		t.Errorf("Expected happiness to be 45, got %f", gs.GetStat("happiness"))
	}

	// Test boundary conditions - overflow
	overflowEffects := map[string]float64{
		"hunger": 50, // Should cap at max (100)
	}

	gs.ApplyInteractionEffects(overflowEffects)

	if gs.GetStat("hunger") != 100 {
		t.Errorf("Expected hunger to cap at 100, got %f", gs.GetStat("hunger"))
	}

	// Test negative effects - underflow
	underflowEffects := map[string]float64{
		"happiness": -60, // Should floor at 0
	}

	gs.ApplyInteractionEffects(underflowEffects)

	if gs.GetStat("happiness") != 0 {
		t.Errorf("Expected happiness to floor at 0, got %f", gs.GetStat("happiness"))
	}
}

// TestCanSatisfyRequirements verifies interaction requirement checking
func TestCanSatisfyRequirements(t *testing.T) {
	config := map[string]StatConfig{
		"energy": {Initial: 50, Max: 100, DegradationRate: 1.0, CriticalThreshold: 20},
		"hunger": {Initial: 80, Max: 100, DegradationRate: 1.0, CriticalThreshold: 20},
	}

	gs := NewGameState(config, nil)

	// Test requirement satisfaction
	requirements := map[string]map[string]float64{
		"energy": {"min": 30}, // Should pass (50 >= 30)
		"hunger": {"max": 90}, // Should pass (80 <= 90)
	}

	if !gs.CanSatisfyRequirements(requirements) {
		t.Error("Expected requirements to be satisfied")
	}

	// Test requirement failure
	failRequirements := map[string]map[string]float64{
		"energy": {"min": 60}, // Should fail (50 < 60)
	}

	if gs.CanSatisfyRequirements(failRequirements) {
		t.Error("Expected requirements to fail")
	}

	// Test non-existent stat
	invalidRequirements := map[string]map[string]float64{
		"nonexistent": {"min": 10},
	}

	if gs.CanSatisfyRequirements(invalidRequirements) {
		t.Error("Expected requirements to fail for non-existent stat")
	}
}

// TestGetStatPercentage verifies percentage calculation
func TestGetStatPercentage(t *testing.T) {
	config := map[string]StatConfig{
		"hunger": {Initial: 75, Max: 100, DegradationRate: 1.0, CriticalThreshold: 20},
	}

	gs := NewGameState(config, nil)

	percentage := gs.GetStatPercentage("hunger")
	if percentage != 75.0 {
		t.Errorf("Expected hunger percentage to be 75.0, got %f", percentage)
	}

	// Test non-existent stat
	nonExistent := gs.GetStatPercentage("nonexistent")
	if nonExistent != 0.0 {
		t.Errorf("Expected non-existent stat percentage to be 0.0, got %f", nonExistent)
	}
}

// TestGetOverallMood verifies mood calculation
func TestGetOverallMood(t *testing.T) {
	config := map[string]StatConfig{
		"hunger":    {Initial: 80, Max: 100, DegradationRate: 1.0, CriticalThreshold: 20},
		"happiness": {Initial: 60, Max: 100, DegradationRate: 1.0, CriticalThreshold: 20},
	}

	gs := NewGameState(config, nil)

	mood := gs.GetOverallMood()
	expectedMood := (80.0 + 60.0) / 2.0 // Average of both stats
	if mood != expectedMood {
		t.Errorf("Expected overall mood to be %f, got %f", expectedMood, mood)
	}

	// Test with nil game state
	var nilGs *GameState
	nilMood := nilGs.GetOverallMood()
	if nilMood != 50.0 {
		t.Errorf("Expected nil game state mood to be 50.0, got %f", nilMood)
	}
}

// TestGameStateJSON verifies serialization and deserialization
func TestGameStateJSON(t *testing.T) {
	config := map[string]StatConfig{
		"hunger": {Initial: 75, Max: 100, DegradationRate: 1.5, CriticalThreshold: 20},
	}

	gameConfig := &GameConfig{
		StatsDecayInterval:             2 * time.Minute,
		CriticalStateAnimationPriority: true,
	}

	original := NewGameState(config, gameConfig)
	original.TotalPlayTime = 30 * time.Minute

	// Serialize
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal GameState: %v", err)
	}

	// Deserialize
	var restored GameState
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Failed to unmarshal GameState: %v", err)
	}

	// Verify key fields
	if restored.GetStat("hunger") != original.GetStat("hunger") {
		t.Errorf("Expected hunger stat to be preserved, got %f vs %f", restored.GetStat("hunger"), original.GetStat("hunger"))
	}

	if restored.TotalPlayTime != original.TotalPlayTime {
		t.Errorf("Expected play time to be preserved, got %v vs %v", restored.TotalPlayTime, original.TotalPlayTime)
	}

	if restored.Config.StatsDecayInterval != original.Config.StatsDecayInterval {
		t.Errorf("Expected decay interval to be preserved")
	}
}

// TestGameStateValidation verifies data consistency checking
func TestGameStateValidation(t *testing.T) {
	// Test valid game state
	config := map[string]StatConfig{
		"hunger": {Initial: 75, Max: 100, DegradationRate: 1.0, CriticalThreshold: 20},
	}

	validGs := NewGameState(config, nil)
	if err := validGs.Validate(); err != nil {
		t.Errorf("Expected valid game state to pass validation, got: %v", err)
	}

	// Test nil game state
	var nilGs *GameState
	if err := nilGs.Validate(); err == nil {
		t.Error("Expected nil game state to fail validation")
	}

	// Test empty stats
	emptyGs := &GameState{Stats: make(map[string]*Stat)}
	if err := emptyGs.Validate(); err == nil {
		t.Error("Expected empty stats to fail validation")
	}

	// Test invalid stat values
	invalidGs := NewGameState(config, nil)
	invalidGs.Stats["hunger"].Current = -10 // Invalid negative value
	if err := invalidGs.Validate(); err == nil {
		t.Error("Expected negative stat value to fail validation")
	}

	// Test current > max
	invalidGs2 := NewGameState(config, nil)
	invalidGs2.Stats["hunger"].Current = 150 // Exceeds max
	if err := invalidGs2.Validate(); err == nil {
		t.Error("Expected current > max to fail validation")
	}
}

// TestGameStateNilSafety verifies all methods handle nil gracefully
func TestGameStateNilSafety(t *testing.T) {
	var nilGs *GameState

	// All these should not panic and return sensible defaults
	if nilGs.Update(time.Second) != nil {
		t.Error("Expected nil game state Update to return nil")
	}

	nilGs.ApplyInteractionEffects(map[string]float64{"test": 10})

	if nilGs.GetStats() != nil {
		t.Error("Expected nil game state GetStats to return nil")
	}

	if nilGs.GetStat("test") != 0 {
		t.Error("Expected nil game state GetStat to return 0")
	}

	if nilGs.GetCriticalStates() != nil {
		t.Error("Expected nil game state GetCriticalStates to return nil")
	}

	if nilGs.GetAge() != 0 {
		t.Error("Expected nil game state GetAge to return 0")
	}

	if nilGs.GetPlayTime() != 0 {
		t.Error("Expected nil game state GetPlayTime to return 0")
	}

	if !nilGs.CanSatisfyRequirements(map[string]map[string]float64{"test": {"min": 10}}) {
		t.Error("Expected nil game state to satisfy any requirements")
	}

	if nilGs.GetStatPercentage("test") != 0 {
		t.Error("Expected nil game state GetStatPercentage to return 0")
	}

	if nilGs.GetOverallMood() != 50 {
		t.Error("Expected nil game state GetOverallMood to return 50")
	}
}

// TestStateDegradationRealTime verifies realistic degradation scenarios
func TestStateDegradationRealTime(t *testing.T) {
	config := map[string]StatConfig{
		"hunger": {
			Initial:           100,
			Max:               100,
			DegradationRate:   1.0, // 1 point per minute
			CriticalThreshold: 20,
		},
	}

	gs := NewGameState(config, &GameConfig{StatsDecayInterval: time.Minute})

	// Simulate 30 minutes of degradation
	gs.LastDecayUpdate = time.Now().Add(-30 * time.Minute)

	triggeredStates := gs.Update(time.Second)

	// Should have degraded by 30 points (100 - 30 = 70)
	expectedHunger := 70.0
	actualHunger := gs.GetStat("hunger")

	if !floatEquals(actualHunger, expectedHunger) {
		t.Errorf("Expected hunger after 30 minutes to be %f, got %f", expectedHunger, actualHunger)
	}

	// Should not trigger critical states yet
	if len(triggeredStates) > 0 {
		t.Errorf("Expected no critical states at hunger=70, got %v", triggeredStates)
	}

	// Simulate another 55 minutes (total 85 minutes, hunger should be 15)
	gs.LastDecayUpdate = time.Now().Add(-55 * time.Minute)
	triggeredStates = gs.Update(time.Second)

	expectedHunger = 15.0 // 70 - 55 = 15
	actualHunger = gs.GetStat("hunger")

	if !floatEquals(actualHunger, expectedHunger) {
		t.Errorf("Expected hunger after 85 total minutes to be %f, got %f", expectedHunger, actualHunger)
	}

	// Should trigger critical states now
	if len(triggeredStates) == 0 {
		t.Error("Expected critical states to be triggered at hunger=15")
	}
}
