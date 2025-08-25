package character

import (
	"os"
	"testing"
	"time"
)

// TestMoodBasedAnimationSelection tests mood-based idle animation selection
func TestMoodBasedAnimationSelection(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mood_animation_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create character card with mood-based animations enabled
	card := createTestGameCharacterCard()
	card.GameRules.MoodBasedAnimations = true

	createTestAnimationFiles(t, tmpDir)

	char, err := New(card, tmpDir)
	if err != nil {
		t.Fatalf("Failed to create character: %v", err)
	}

	err = char.EnableGameMode(nil, "")
	if err != nil {
		t.Fatalf("Failed to enable game mode: %v", err)
	}

	// Test high mood (80+) should select happy animation
	char.gameState.Stats["hunger"].Current = 90.0
	char.gameState.Stats["happiness"].Current = 95.0
	char.gameState.Stats["health"].Current = 85.0
	char.gameState.Stats["energy"].Current = 80.0

	animation := char.selectIdleAnimation()
	if animation != "happy" {
		t.Errorf("High mood should select 'happy' animation, got: %s", animation)
	}

	// Test normal mood (60-79) should select idle animation
	char.gameState.Stats["hunger"].Current = 70.0
	char.gameState.Stats["happiness"].Current = 65.0
	char.gameState.Stats["health"].Current = 70.0
	char.gameState.Stats["energy"].Current = 60.0

	animation = char.selectIdleAnimation()
	if animation != "idle" {
		t.Errorf("Normal mood should select 'idle' animation, got: %s", animation)
	}

	// Test low mood (20-39) should select sad animation
	char.gameState.Stats["hunger"].Current = 30.0
	char.gameState.Stats["happiness"].Current = 25.0
	char.gameState.Stats["health"].Current = 35.0
	char.gameState.Stats["energy"].Current = 20.0

	animation = char.selectIdleAnimation()
	if animation != "sad" {
		t.Errorf("Low mood should select 'sad' animation, got: %s", animation)
	}

	// Test very low mood (<20) should select sad animation
	char.gameState.Stats["hunger"].Current = 10.0
	char.gameState.Stats["happiness"].Current = 15.0
	char.gameState.Stats["health"].Current = 5.0
	char.gameState.Stats["energy"].Current = 18.0

	animation = char.selectIdleAnimation()
	if animation != "sad" {
		t.Errorf("Very low mood should select 'sad' animation, got: %s", animation)
	}
}

// TestMoodBasedAnimationDisabled tests that mood-based animations respect the config flag
func TestMoodBasedAnimationDisabled(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mood_disabled_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create character card with mood-based animations disabled
	card := createTestGameCharacterCard()
	card.GameRules.MoodBasedAnimations = false

	createTestAnimationFiles(t, tmpDir)

	char, err := New(card, tmpDir)
	if err != nil {
		t.Fatalf("Failed to create character: %v", err)
	}

	err = char.EnableGameMode(nil, "")
	if err != nil {
		t.Fatalf("Failed to enable game mode: %v", err)
	}

	// Set very low mood but mood-based animations are disabled
	char.gameState.Stats["hunger"].Current = 5.0
	char.gameState.Stats["happiness"].Current = 10.0
	char.gameState.Stats["health"].Current = 8.0
	char.gameState.Stats["energy"].Current = 12.0

	animation := char.selectIdleAnimation()
	if animation != "idle" {
		t.Errorf("With mood-based animations disabled, should always return 'idle', got: %s", animation)
	}
}

// TestMoodBasedAnimationWithMissingAnimations tests fallback when mood animations don't exist
func TestMoodBasedAnimationWithMissingAnimations(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mood_missing_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create character card with mood-based animations enabled but limited animations
	card := &CharacterCard{
		Name:        "Test Character",
		Description: "Test character for mood-based animation testing",
		Animations: map[string]string{
			"idle":    "animations/idle.gif",
			"talking": "animations/talking.gif",
			// Note: missing "happy" and "sad" animations
		},
		Dialogs: []Dialog{
			{
				Trigger:   "click",
				Responses: []string{"Hello!"},
				Animation: "talking",
				Cooldown:  5,
			},
		},
		Behavior: Behavior{
			IdleTimeout:     30,
			MovementEnabled: false,
			DefaultSize:     128,
		},
		Stats: map[string]StatConfig{
			"hunger":    {Initial: 50, Max: 100, DegradationRate: 1.0, CriticalThreshold: 20},
			"happiness": {Initial: 50, Max: 100, DegradationRate: 1.0, CriticalThreshold: 15},
			"health":    {Initial: 50, Max: 100, DegradationRate: 0.5, CriticalThreshold: 25},
			"energy":    {Initial: 50, Max: 100, DegradationRate: 1.5, CriticalThreshold: 30},
		},
		GameRules: &GameConfig{
			StatsDecayInterval:             time.Minute,
			AutoSaveInterval:               5 * time.Minute,
			CriticalStateAnimationPriority: true,
			DeathEnabled:                   false,
			EvolutionEnabled:               true,
			MoodBasedAnimations:            true, // Enabled but missing animations
		},
	}

	// Create only idle and talking animations
	createTestAnimationFile(t, tmpDir, "idle.gif")
	createTestAnimationFile(t, tmpDir, "talking.gif")

	char, err := New(card, tmpDir)
	if err != nil {
		t.Fatalf("Failed to create character: %v", err)
	}

	err = char.EnableGameMode(nil, "")
	if err != nil {
		t.Fatalf("Failed to enable game mode: %v", err)
	}

	// Set high mood - should want "happy" but fall back to "idle"
	char.gameState.Stats["hunger"].Current = 95.0
	char.gameState.Stats["happiness"].Current = 90.0
	char.gameState.Stats["health"].Current = 95.0
	char.gameState.Stats["energy"].Current = 85.0

	animation := char.selectIdleAnimation()
	if animation != "idle" {
		t.Errorf("Should fallback to 'idle' when mood animation doesn't exist, got: %s", animation)
	}

	// Set low mood - should want "sad" but fall back to "idle"
	char.gameState.Stats["hunger"].Current = 15.0
	char.gameState.Stats["happiness"].Current = 20.0
	char.gameState.Stats["health"].Current = 25.0
	char.gameState.Stats["energy"].Current = 10.0

	animation = char.selectIdleAnimation()
	if animation != "idle" {
		t.Errorf("Should fallback to 'idle' when mood animation doesn't exist, got: %s", animation)
	}
}

// TestMoodBasedAnimationWithoutGameState tests behavior when game state is not enabled
func TestMoodBasedAnimationWithoutGameState(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mood_nogame_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	card := createTestCharacterCard()
	createTestAnimationFiles(t, tmpDir)

	char, err := New(card, tmpDir)
	if err != nil {
		t.Fatalf("Failed to create character: %v", err)
	}

	// Don't enable game mode - no game state
	animation := char.selectIdleAnimation()
	if animation != "idle" {
		t.Errorf("Without game state, should always return 'idle', got: %s", animation)
	}
}

// TestMoodBasedAnimationIntegration tests mood-based animation in full update loop
func TestMoodBasedAnimationIntegration(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mood_integration_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	card := createTestGameCharacterCard()
	card.GameRules.MoodBasedAnimations = true
	// Set short idle timeout for faster testing
	card.Behavior.IdleTimeout = 1

	createTestAnimationFiles(t, tmpDir)

	char, err := New(card, tmpDir)
	if err != nil {
		t.Fatalf("Failed to create character: %v", err)
	}

	err = char.EnableGameMode(nil, "")
	if err != nil {
		t.Fatalf("Failed to enable game mode: %v", err)
	}

	// Set high mood
	char.gameState.Stats["hunger"].Current = 90.0
	char.gameState.Stats["happiness"].Current = 95.0
	char.gameState.Stats["health"].Current = 85.0
	char.gameState.Stats["energy"].Current = 80.0

	// Force character into a non-idle state
	err = char.ForceState("talking")
	if err != nil {
		t.Fatalf("Failed to force talking state: %v", err)
	}

	// Wait for idle timeout and update
	time.Sleep(time.Second * 2)
	char.Update()

	// Should now be in mood-based idle animation
	currentState := char.GetCurrentState()
	if currentState != "happy" {
		t.Errorf("After idle timeout with high mood, should be in 'happy' state, got: %s", currentState)
	}
}
