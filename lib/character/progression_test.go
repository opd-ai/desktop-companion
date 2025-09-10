package character

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewProgressionState(t *testing.T) {
	config := &ProgressionConfig{
		Levels: []LevelConfig{
			{Name: "Baby", Requirement: map[string]int64{"age": 0}, Size: 64},
			{Name: "Child", Requirement: map[string]int64{"age": 86400}, Size: 96},
		},
		Achievements: []AchievementConfig{
			{Name: "Well Fed", Requirement: map[string]map[string]interface{}{
				"hunger": {"maintainAbove": 80.0, "duration": 3600.0},
			}},
		},
	}

	ps := NewProgressionState(config)

	if ps == nil {
		t.Fatal("NewProgressionState returned nil")
	}

	if ps.CurrentLevel != "Baby" {
		t.Errorf("Expected initial level 'Baby', got '%s'", ps.CurrentLevel)
	}

	if ps.Age != 0 {
		t.Errorf("Expected initial age 0, got %v", ps.Age)
	}

	if len(ps.Achievements) != 0 {
		t.Errorf("Expected no initial achievements, got %d", len(ps.Achievements))
	}

	if len(ps.AchievementProgress) != 1 {
		t.Errorf("Expected 1 achievement progress tracker, got %d", len(ps.AchievementProgress))
	}

	if ps.Config != config {
		t.Error("Expected config to be set correctly")
	}
}

func TestProgressionState_Update_LevelProgression(t *testing.T) {
	config := &ProgressionConfig{
		Levels: []LevelConfig{
			{Name: "Baby", Requirement: map[string]int64{"age": 0}, Size: 64},
			{Name: "Child", Requirement: map[string]int64{"age": 3600}, Size: 96},  // 1 hour
			{Name: "Adult", Requirement: map[string]int64{"age": 7200}, Size: 128}, // 2 hours
		},
	}

	ps := NewProgressionState(config)
	gameState := createTestGameState()

	// Test no level change for young character
	levelChanged, _ := ps.Update(gameState, time.Minute*30) // 30 minutes
	if levelChanged {
		t.Error("Should not level up after 30 minutes")
	}
	if ps.CurrentLevel != "Baby" {
		t.Errorf("Should still be 'Baby', got '%s'", ps.CurrentLevel)
	}

	// Test level change after 1 hour
	levelChanged, _ = ps.Update(gameState, time.Minute*30) // Total: 1 hour
	if !levelChanged {
		t.Error("Should level up after 1 hour")
	}
	if ps.CurrentLevel != "Child" {
		t.Errorf("Should be 'Child', got '%s'", ps.CurrentLevel)
	}

	// Test another level change after 2 hours total
	levelChanged, _ = ps.Update(gameState, time.Hour) // Total: 2 hours
	if !levelChanged {
		t.Error("Should level up after 2 hours")
	}
	if ps.CurrentLevel != "Adult" {
		t.Errorf("Should be 'Adult', got '%s'", ps.CurrentLevel)
	}

	// Test no further level changes
	levelChanged, _ = ps.Update(gameState, time.Hour) // Total: 3 hours
	if levelChanged {
		t.Error("Should not level up beyond Adult")
	}
	if ps.CurrentLevel != "Adult" {
		t.Errorf("Should still be 'Adult', got '%s'", ps.CurrentLevel)
	}
}

func TestProgressionState_Update_AchievementTracking(t *testing.T) {
	config := &ProgressionConfig{
		Levels: []LevelConfig{
			{Name: "Baby", Requirement: map[string]int64{"age": 0}, Size: 64},
		},
		Achievements: []AchievementConfig{
			{
				Name: "Well Fed",
				Requirement: map[string]map[string]interface{}{
					"hunger":        {"maintainAbove": 80.0},
					"maintainAbove": {"duration": 60.0}, // 1 minute
				},
			},
			{
				Name: "Happy Pet",
				Requirement: map[string]map[string]interface{}{
					"happiness": {"min": 90.0},
				},
			},
		},
	}

	ps := NewProgressionState(config)
	gameState := createTestGameState()

	// Set stats to meet achievement criteria
	gameState.Stats["hunger"].Current = 85.0
	gameState.Stats["happiness"].Current = 95.0

	// First update - should start tracking "Well Fed"
	_, achievements := ps.Update(gameState, time.Second*30)
	if len(achievements) != 1 {
		t.Errorf("Expected 1 achievement (Happy Pet), got %d: %v", len(achievements), achievements)
	}
	if len(achievements) > 0 && achievements[0].Name != "Happy Pet" {
		t.Errorf("Expected 'Happy Pet' achievement, got '%s'", achievements[0].Name)
	}

	// Second update - should complete "Well Fed" after duration
	_, achievements = ps.Update(gameState, time.Second*30) // Total: 1 minute
	if len(achievements) != 1 {
		t.Errorf("Expected 1 achievement (Well Fed), got %d: %v", len(achievements), achievements)
	}
	if len(achievements) > 0 && achievements[0].Name != "Well Fed" {
		t.Errorf("Expected 'Well Fed' achievement, got '%s'", achievements[0].Name)
	}

	// Verify both achievements are recorded
	allAchievements := ps.GetAchievements()
	if len(allAchievements) != 2 {
		t.Errorf("Expected 2 total achievements, got %d: %v", len(allAchievements), allAchievements)
	}
}

func TestProgressionState_Update_AchievementReset(t *testing.T) {
	config := &ProgressionConfig{
		Levels: []LevelConfig{
			{Name: "Baby", Requirement: map[string]int64{"age": 0}, Size: 64},
		},
		Achievements: []AchievementConfig{
			{
				Name: "Well Fed",
				Requirement: map[string]map[string]interface{}{
					"hunger":        {"maintainAbove": 80.0},
					"maintainAbove": {"duration": 120.0}, // 2 minutes
				},
			},
		},
	}

	ps := NewProgressionState(config)
	gameState := createTestGameState()

	// Start meeting criteria
	gameState.Stats["hunger"].Current = 85.0
	ps.Update(gameState, time.Minute) // 1 minute of progress

	// Check progress is being tracked
	progress := ps.AchievementProgress["Well Fed"]
	if !progress.MetCriteria {
		t.Error("Should be meeting criteria")
	}

	// Stop meeting criteria
	gameState.Stats["hunger"].Current = 75.0 // Below threshold
	ps.Update(gameState, time.Second*30)

	// Progress should be reset
	progress = ps.AchievementProgress["Well Fed"]
	if progress.MetCriteria {
		t.Error("Should no longer be meeting criteria")
	}
	if progress.Duration != 0 {
		t.Error("Duration should be reset to 0")
	}
}

func TestProgressionState_RecordInteraction(t *testing.T) {
	ps := NewProgressionState(nil)

	ps.RecordInteraction("feed")
	ps.RecordInteraction("feed")
	ps.RecordInteraction("play")

	counts := ps.GetInteractionCounts()
	if counts["feed"] != 2 {
		t.Errorf("Expected 2 feed interactions, got %d", counts["feed"])
	}
	if counts["play"] != 1 {
		t.Errorf("Expected 1 play interaction, got %d", counts["play"])
	}
}

func TestProgressionState_GetCurrentLevel(t *testing.T) {
	config := &ProgressionConfig{
		Levels: []LevelConfig{
			{Name: "Baby", Requirement: map[string]int64{"age": 0}, Size: 64},
			{Name: "Child", Requirement: map[string]int64{"age": 3600}, Size: 96},
		},
	}

	ps := NewProgressionState(config)

	level := ps.GetCurrentLevel()
	if level == nil {
		t.Fatal("GetCurrentLevel returned nil")
	}
	if level.Name != "Baby" {
		t.Errorf("Expected level 'Baby', got '%s'", level.Name)
	}
	if level.Size != 64 {
		t.Errorf("Expected size 64, got %d", level.Size)
	}

	// Change level and test again
	ps.CurrentLevel = "Child"
	level = ps.GetCurrentLevel()
	if level.Name != "Child" {
		t.Errorf("Expected level 'Child', got '%s'", level.Name)
	}
	if level.Size != 96 {
		t.Errorf("Expected size 96, got %d", level.Size)
	}
}

func TestProgressionState_GetCurrentSize(t *testing.T) {
	config := &ProgressionConfig{
		Levels: []LevelConfig{
			{Name: "Baby", Requirement: map[string]int64{"age": 0}, Size: 64},
		},
	}

	ps := NewProgressionState(config)

	size := ps.GetCurrentSize()
	if size != 64 {
		t.Errorf("Expected size 64, got %d", size)
	}

	// Test with nil config (should return default)
	ps.Config = nil
	size = ps.GetCurrentSize()
	if size != 128 {
		t.Errorf("Expected default size 128, got %d", size)
	}
}

func TestProgressionState_GetLevelAnimation(t *testing.T) {
	config := &ProgressionConfig{
		Levels: []LevelConfig{
			{Name: "Baby", Requirement: map[string]int64{"age": 0}, Size: 64, Animations: map[string]string{
				"idle":  "baby_idle.gif",
				"happy": "baby_happy.gif",
			}},
		},
	}

	ps := NewProgressionState(config)

	animation, exists := ps.GetLevelAnimation("idle")
	if !exists {
		t.Error("Should find level-specific animation")
	}
	if animation != "baby_idle.gif" {
		t.Errorf("Expected 'baby_idle.gif', got '%s'", animation)
	}

	animation, exists = ps.GetLevelAnimation("talking")
	if exists {
		t.Error("Should not find non-existent level animation")
	}
	if animation != "" {
		t.Errorf("Expected empty string for non-existent animation, got '%s'", animation)
	}
}

func TestProgressionState_MarshalUnmarshalJSON(t *testing.T) {
	config := &ProgressionConfig{
		Levels: []LevelConfig{
			{Name: "Baby", Requirement: map[string]int64{"age": 0}, Size: 64},
		},
		Achievements: []AchievementConfig{
			{Name: "Test Achievement", Requirement: map[string]map[string]interface{}{
				"hunger": {"min": 50.0},
			}},
		},
	}

	original := NewProgressionState(config)
	original.CurrentLevel = "Baby"
	original.Age = time.Hour * 2
	original.TotalCareTime = time.Hour * 3
	original.Achievements = []string{"Test Achievement"}
	original.InteractionCounts["feed"] = 5
	original.RecordInteraction("play")

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal from JSON
	var restored ProgressionState
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Verify important fields
	if restored.CurrentLevel != original.CurrentLevel {
		t.Errorf("CurrentLevel mismatch: expected '%s', got '%s'", original.CurrentLevel, restored.CurrentLevel)
	}

	if restored.Age != original.Age {
		t.Errorf("Age mismatch: expected %v, got %v", original.Age, restored.Age)
	}

	if restored.TotalCareTime != original.TotalCareTime {
		t.Errorf("TotalCareTime mismatch: expected %v, got %v", original.TotalCareTime, restored.TotalCareTime)
	}

	if len(restored.Achievements) != len(original.Achievements) {
		t.Errorf("Achievements length mismatch: expected %d, got %d", len(original.Achievements), len(restored.Achievements))
	}

	if restored.InteractionCounts["feed"] != 5 {
		t.Errorf("Expected feed count 5, got %d", restored.InteractionCounts["feed"])
	}

	if restored.InteractionCounts["play"] != 1 {
		t.Errorf("Expected play count 1, got %d", restored.InteractionCounts["play"])
	}
}

func TestProgressionState_Validate(t *testing.T) {
	// Test nil state
	var ps *ProgressionState
	if err := ps.Validate(); err == nil {
		t.Error("Should reject nil progression state")
	}

	// Test negative age
	ps = &ProgressionState{Age: -time.Hour}
	if err := ps.Validate(); err == nil {
		t.Error("Should reject negative age")
	}

	// Test negative care time
	ps = &ProgressionState{TotalCareTime: -time.Hour}
	if err := ps.Validate(); err == nil {
		t.Error("Should reject negative care time")
	}

	// Test invalid config
	ps = &ProgressionState{
		Config: &ProgressionConfig{
			Levels: []LevelConfig{}, // Empty levels
		},
	}
	if err := ps.Validate(); err == nil {
		t.Error("Should reject empty levels config")
	}

	// Test valid state
	ps = NewProgressionState(&ProgressionConfig{
		Levels: []LevelConfig{
			{Name: "Baby", Requirement: map[string]int64{"age": 0}, Size: 64},
		},
	})
	if err := ps.Validate(); err != nil {
		t.Errorf("Should accept valid state: %v", err)
	}
}

func TestProgressionState_NilSafety(t *testing.T) {
	var ps *ProgressionState

	// Test all methods with nil receiver
	if changed, achievements := ps.Update(nil, time.Minute); changed || len(achievements) > 0 {
		t.Error("Nil progression state should not report changes")
	}

	ps.RecordInteraction("test") // Should not panic

	if level := ps.GetCurrentLevel(); level != nil {
		t.Error("Nil progression state should return nil level")
	}

	if age := ps.GetAge(); age != 0 {
		t.Error("Nil progression state should return 0 age")
	}

	if achievements := ps.GetAchievements(); achievements != nil {
		t.Error("Nil progression state should return nil achievements")
	}

	if counts := ps.GetInteractionCounts(); counts != nil {
		t.Error("Nil progression state should return nil counts")
	}

	if size := ps.GetCurrentSize(); size != 128 {
		t.Error("Nil progression state should return default size")
	}

	if animation, exists := ps.GetLevelAnimation("test"); exists || animation != "" {
		t.Error("Nil progression state should not find animations")
	}
}

func TestProgressionConfig_Validation(t *testing.T) {
	// Test invalid level size
	config := &ProgressionConfig{
		Levels: []LevelConfig{
			{Name: "Baby", Requirement: map[string]int64{"age": 0}, Size: 10}, // Too small
		},
	}
	ps := &ProgressionState{Config: config}
	if err := ps.Validate(); err == nil {
		t.Error("Should reject invalid level size")
	}

	// Test empty level name
	config = &ProgressionConfig{
		Levels: []LevelConfig{
			{Name: "", Requirement: map[string]int64{"age": 0}, Size: 64},
		},
	}
	ps = &ProgressionState{Config: config}
	if err := ps.Validate(); err == nil {
		t.Error("Should reject empty level name")
	}

	// Test empty achievement name
	config = &ProgressionConfig{
		Levels: []LevelConfig{
			{Name: "Baby", Requirement: map[string]int64{"age": 0}, Size: 64},
		},
		Achievements: []AchievementConfig{
			{Name: "", Requirement: map[string]map[string]interface{}{"hunger": {"min": 50.0}}},
		},
	}
	ps = &ProgressionState{Config: config}
	if err := ps.Validate(); err == nil {
		t.Error("Should reject empty achievement name")
	}

	// Test empty achievement requirement
	config = &ProgressionConfig{
		Levels: []LevelConfig{
			{Name: "Baby", Requirement: map[string]int64{"age": 0}, Size: 64},
		},
		Achievements: []AchievementConfig{
			{Name: "Test", Requirement: map[string]map[string]interface{}{}},
		},
	}
	ps = &ProgressionState{Config: config}
	if err := ps.Validate(); err == nil {
		t.Error("Should reject empty achievement requirement")
	}
}

// createTestGameState creates a game state for testing
func createTestGameState() *GameState {
	statConfigs := map[string]StatConfig{
		"hunger":    {Initial: 100, Max: 100, DegradationRate: 1.0, CriticalThreshold: 20},
		"happiness": {Initial: 100, Max: 100, DegradationRate: 0.8, CriticalThreshold: 15},
		"health":    {Initial: 100, Max: 100, DegradationRate: 0.3, CriticalThreshold: 10},
		"energy":    {Initial: 100, Max: 100, DegradationRate: 1.5, CriticalThreshold: 25},
	}

	config := &GameConfig{
		StatsDecayInterval:             time.Minute,
		CriticalStateAnimationPriority: true,
		MoodBasedAnimations:            true,
	}

	return NewGameState(statConfigs, config)
}
