package character

import (
	"fmt"
	"testing"
	"time"
)

func TestProgressionState_DebugAchievement(t *testing.T) {
	config := &ProgressionConfig{
		Levels: []LevelConfig{
			{Name: "Baby", Requirement: map[string]int64{"age": 0}, Size: 64},
		},
		Achievements: []AchievementConfig{
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
	gameState.Stats["happiness"].Current = 95.0

	fmt.Printf("Before update: happiness = %f\n", gameState.Stats["happiness"].Current)
	fmt.Printf("Achievement requirement: %+v\n", config.Achievements[0].Requirement)

	// Test evaluation directly
	met := ps.evaluateAchievementRequirement(config.Achievements[0].Requirement, gameState)
	fmt.Printf("Requirement met: %t\n", met)

	// Test the update
	_, achievements := ps.Update(gameState, time.Second*30)
	fmt.Printf("Achievements returned: %v (length: %d)\n", achievements, len(achievements))

	// Check achievement progress
	progress := ps.AchievementProgress["Happy Pet"]
	fmt.Printf("Progress: MetCriteria=%t, RequiredTime=%v\n", progress.MetCriteria, progress.RequiredTime)
}
