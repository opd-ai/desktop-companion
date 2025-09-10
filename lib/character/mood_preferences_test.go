package character

import (
	"os"
	"path/filepath"
	"testing"
)

// TestGetMoodCategory tests the new mood category functionality
func TestGetMoodCategory(t *testing.T) {
	config := map[string]StatConfig{
		"hunger":    {Initial: 80, Max: 100, DegradationRate: 1.0, CriticalThreshold: 20},
		"happiness": {Initial: 60, Max: 100, DegradationRate: 1.0, CriticalThreshold: 20},
		"energy":    {Initial: 70, Max: 100, DegradationRate: 1.0, CriticalThreshold: 20},
	}

	gs := NewGameState(config, nil)

	tests := []struct {
		name          string
		hunger        float64
		happiness     float64
		energy        float64 // Changed from health to energy
		expectedMood  string
		expectedValue float64 // For verification
	}{
		{
			name:          "Happy mood (>80)",
			hunger:        90,
			happiness:     85,
			energy:        85,
			expectedMood:  "happy",
			expectedValue: (90.0 + 85.0 + 85.0) / 3.0, // 86.67
		},
		{
			name:          "Content mood (60-79)",
			hunger:        70,
			happiness:     65,
			energy:        70,
			expectedMood:  "content",
			expectedValue: (70.0 + 65.0 + 70.0) / 3.0, // 68.33
		},
		{
			name:          "Neutral mood (40-59)",
			hunger:        50,
			happiness:     45,
			energy:        55,
			expectedMood:  "neutral",
			expectedValue: (50.0 + 45.0 + 55.0) / 3.0, // 50
		},
		{
			name:          "Sad mood (20-39)",
			hunger:        30,
			happiness:     25,
			energy:        35,
			expectedMood:  "sad",
			expectedValue: (30.0 + 25.0 + 35.0) / 3.0, // 30
		},
		{
			name:          "Depressed mood (<20)",
			hunger:        15,
			happiness:     10,
			energy:        18,
			expectedMood:  "depressed",
			expectedValue: (15.0 + 10.0 + 18.0) / 3.0, // 14.33
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set stat values
			gs.Stats["hunger"].Current = tt.hunger
			gs.Stats["happiness"].Current = tt.happiness
			gs.Stats["energy"].Current = tt.energy

			// Verify mood calculation
			mood := gs.GetOverallMood()
			if mood < tt.expectedValue-0.1 || mood > tt.expectedValue+0.1 {
				t.Errorf("Expected mood value around %f, got %f", tt.expectedValue, mood)
			}

			// Test mood category
			category := gs.GetMoodCategory()
			if category != tt.expectedMood {
				t.Errorf("Expected mood category '%s', got '%s'", tt.expectedMood, category)
			}
		})
	}
}

// TestMoodBasedAnimationPreferences tests the new mood preferences system
func TestMoodBasedAnimationPreferences(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mood_preferences_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create character card with mood animation preferences
	card := createTestGameCharacterCard()
	card.GameRules.MoodBasedAnimations = true
	card.Behavior.MoodAnimationPreferences = map[string][]string{
		"happy":     {"happy", "excited", "dance"},
		"content":   {"idle", "content"},
		"neutral":   {"idle", "thinking"},
		"sad":       {"sad", "crying"},
		"depressed": {"sad", "depressed", "exhausted"},
	}

	// Create animation files including mood-specific ones
	createTestAnimationFiles(t, tmpDir)
	createAdditionalMoodAnimations(t, tmpDir)

	char, err := New(card, tmpDir)
	if err != nil {
		t.Fatalf("Failed to create character: %v", err)
	}

	err = char.EnableGameMode(nil, "")
	if err != nil {
		t.Fatalf("Failed to enable game mode: %v", err)
	}

	tests := []struct {
		name              string
		hunger            float64
		happiness         float64
		energy            float64 // Using energy instead of health
		expectedMoodState string
	}{
		{
			name:              "Happy mood should use happy animation",
			hunger:            90,
			happiness:         95,
			energy:            85,
			expectedMoodState: "happy",
		},
		{
			name:              "Content mood should use content/idle animation",
			hunger:            70,
			happiness:         65,
			energy:            70,
			expectedMoodState: "idle", // content animation doesn't exist, should fallback
		},
		{
			name:              "Sad mood should use sad animation",
			hunger:            30,
			happiness:         25,
			energy:            35,
			expectedMoodState: "sad",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Debug: Check if gameState exists
			if char.gameState == nil {
				t.Fatal("GameState is nil")
			}

			// Set mood stats (check if stats exist first)
			if _, exists := char.gameState.Stats["hunger"]; !exists {
				t.Fatal("hunger stat not found")
			}
			if _, exists := char.gameState.Stats["happiness"]; !exists {
				t.Fatal("happiness stat not found")
			}
			if _, exists := char.gameState.Stats["energy"]; !exists {
				t.Fatal("energy stat not found - using this instead of health")
			}

			char.gameState.Stats["hunger"].Current = tt.hunger
			char.gameState.Stats["happiness"].Current = tt.happiness
			char.gameState.Stats["energy"].Current = tt.energy // Use energy instead of health

			// Test mood-based animation selection
			selectedAnimation := char.selectMoodAppropriateAnimation("idle")
			if selectedAnimation != tt.expectedMoodState {
				t.Errorf("Expected mood animation '%s', got '%s'", tt.expectedMoodState, selectedAnimation)
			}
		})
	}
}

// TestSelectMoodAppropriateAnimation tests the mood animation selection logic
func TestSelectMoodAppropriateAnimation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mood_appropriate_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	card := createTestGameCharacterCard()
	// Add mood-specific animations to the card
	card.Animations["thinking"] = "thinking.gif"
	card.Animations["content"] = "content.gif"
	card.Animations["excited"] = "excited.gif"
	card.Animations["depressed"] = "depressed.gif"

	card.Behavior.MoodAnimationPreferences = map[string][]string{
		"happy":     {"happy", "excited", "dance"},
		"content":   {"content", "idle"},
		"neutral":   {"thinking", "idle"},
		"sad":       {"sad", "crying"},
		"depressed": {"depressed", "sad"},
	}

	createTestAnimationFiles(t, tmpDir)
	createAdditionalMoodAnimations(t, tmpDir)

	char, err := New(card, tmpDir)
	if err != nil {
		t.Fatalf("Failed to create character: %v", err)
	}

	err = char.EnableGameMode(nil, "")
	if err != nil {
		t.Fatalf("Failed to enable game mode: %v", err)
	}

	tests := []struct {
		name           string
		moodCategory   string
		preferredState string
		expectedResult string
	}{
		{
			name:           "Happy mood with available animation",
			moodCategory:   "happy",
			preferredState: "idle",
			expectedResult: "happy", // Should find "happy" animation
		},
		{
			name:           "Neutral mood with available thinking animation",
			moodCategory:   "neutral",
			preferredState: "idle",
			expectedResult: "thinking", // Should find "thinking" before "idle"
		},
		{
			name:           "Sad mood with available animation",
			moodCategory:   "sad",
			preferredState: "idle",
			expectedResult: "sad", // Should find "sad" animation
		},
		{
			name:           "Content mood with available content animation",
			moodCategory:   "content",
			preferredState: "idle",
			expectedResult: "content", // Should find "content" animation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the mood by directly testing the method
			// Set up mood category in game state
			switch tt.moodCategory {
			case "happy":
				char.gameState.Stats["hunger"].Current = 90
				char.gameState.Stats["happiness"].Current = 85
				char.gameState.Stats["energy"].Current = 85
			case "neutral":
				char.gameState.Stats["hunger"].Current = 50
				char.gameState.Stats["happiness"].Current = 45
				char.gameState.Stats["energy"].Current = 55
			case "sad":
				char.gameState.Stats["hunger"].Current = 30
				char.gameState.Stats["happiness"].Current = 25
				char.gameState.Stats["energy"].Current = 35
			case "content":
				char.gameState.Stats["hunger"].Current = 70
				char.gameState.Stats["happiness"].Current = 65
				char.gameState.Stats["energy"].Current = 70
			}

			result := char.selectMoodAppropriateAnimation(tt.preferredState)
			if result != tt.expectedResult {
				t.Errorf("Expected '%s', got '%s'", tt.expectedResult, result)
			}
		})
	}
}

// TestSetStateWithMoodPreferences tests that setState uses mood preferences
func TestSetStateWithMoodPreferences(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "setstate_mood_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	card := createTestGameCharacterCard()
	card.GameRules.MoodBasedAnimations = true
	card.Behavior.MoodAnimationPreferences = map[string][]string{
		"happy": {"happy", "excited"},
		"sad":   {"sad", "crying"},
	}

	createTestAnimationFiles(t, tmpDir)
	createAdditionalMoodAnimations(t, tmpDir)

	char, err := New(card, tmpDir)
	if err != nil {
		t.Fatalf("Failed to create character: %v", err)
	}

	err = char.EnableGameMode(nil, "")
	if err != nil {
		t.Fatalf("Failed to enable game mode: %v", err)
	}

	// Test happy mood
	char.gameState.Stats["hunger"].Current = 90
	char.gameState.Stats["happiness"].Current = 85
	char.gameState.Stats["energy"].Current = 85

	char.setState("idle")
	currentState := char.GetCurrentState()
	if currentState != "happy" {
		t.Errorf("setState with happy mood should use 'happy' animation, got '%s'", currentState)
	}

	// Test sad mood
	char.gameState.Stats["hunger"].Current = 30
	char.gameState.Stats["happiness"].Current = 25
	char.gameState.Stats["energy"].Current = 35

	char.setState("idle")
	currentState = char.GetCurrentState()
	if currentState != "sad" {
		t.Errorf("setState with sad mood should use 'sad' animation, got '%s'", currentState)
	}
}

// TestMoodPreferencesBackwardCompatibility tests that the system works without preferences
func TestMoodPreferencesBackwardCompatibility(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "backwards_compat_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Character without mood preferences (backward compatibility)
	card := createTestGameCharacterCard()
	card.GameRules.MoodBasedAnimations = true
	// Don't set MoodAnimationPreferences - should use legacy system

	createTestAnimationFiles(t, tmpDir)
	createAdditionalMoodAnimations(t, tmpDir)

	char, err := New(card, tmpDir)
	if err != nil {
		t.Fatalf("Failed to create character: %v", err)
	}

	err = char.EnableGameMode(nil, "")
	if err != nil {
		t.Fatalf("Failed to enable game mode: %v", err)
	}

	// Test high mood - should use legacy mood system
	char.gameState.Stats["hunger"].Current = 90
	char.gameState.Stats["happiness"].Current = 85
	char.gameState.Stats["energy"].Current = 85

	animation := char.selectIdleAnimation()
	if animation != "happy" {
		t.Errorf("Legacy mood system should select 'happy', got '%s'", animation)
	}

	// Test low mood - should use legacy mood system
	char.gameState.Stats["hunger"].Current = 30
	char.gameState.Stats["happiness"].Current = 25
	char.gameState.Stats["energy"].Current = 35

	animation = char.selectIdleAnimation()
	if animation != "sad" {
		t.Errorf("Legacy mood system should select 'sad', got '%s'", animation)
	}
}

// TestMoodPreferencesWithoutGameState tests behavior without game state
func TestMoodPreferencesWithoutGameState(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "no_gamestate_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	card := createTestCharacterCard()
	card.Behavior.MoodAnimationPreferences = map[string][]string{
		"happy": {"happy", "excited"},
	}

	createTestAnimationFiles(t, tmpDir)

	char, err := New(card, tmpDir)
	if err != nil {
		t.Fatalf("Failed to create character: %v", err)
	}

	// Don't enable game mode - no game state
	result := char.selectMoodAppropriateAnimation("idle")
	if result != "idle" {
		t.Errorf("Without game state, should return preferred state, got '%s'", result)
	}
}

// Helper function to create additional mood-specific animations for testing
func createAdditionalMoodAnimations(t *testing.T, basePath string) {
	t.Helper()

	additionalAnimations := []string{"happy", "sad", "thinking", "content", "excited", "depressed"}

	for _, animName := range additionalAnimations {
		// Create the animation file in the correct directory
		gifPath := createTestGIF(t, animName+".gif", 2, []int{100, 100})

		// Move the file to the basePath directory with the correct name
		finalPath := filepath.Join(basePath, animName+".gif")

		// Read the created file
		data, err := os.ReadFile(gifPath)
		if err != nil {
			t.Fatalf("Failed to read test GIF: %v", err)
		}

		// Write it to the target location
		if err := os.WriteFile(finalPath, data, 0o644); err != nil {
			t.Fatalf("Failed to write animation file: %v", err)
		}
	}
}
