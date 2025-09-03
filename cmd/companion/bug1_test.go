package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/opd-ai/desktop-companion/internal/character"
)

// TestBug1MoodBasedAnimationIntegration reproduces the bug where mood-based
// animation selection doesn't work correctly due to platform adapter overriding
// character card's idle timeout setting.
func TestBug1MoodBasedAnimationIntegration(t *testing.T) {
	// Create temporary directory for test assets
	tmpDir, err := os.MkdirTemp("", "bug1_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test character card with mood-based animations enabled
	card := createTestGameCharacterCard()
	card.GameRules.MoodBasedAnimations = true
	// Set short idle timeout for faster testing
	card.Behavior.IdleTimeout = 1

	// Create required animation files
	createTestAnimationFiles(t, tmpDir)

	char, err := character.New(card, tmpDir)
	if err != nil {
		t.Fatalf("Failed to create character: %v", err)
	}

	err = char.EnableGameMode(nil, "")
	if err != nil {
		t.Fatalf("Failed to enable game mode: %v", err)
	}

	// Set high mood stats that should trigger "happy" animation
	// Overall mood = (90+95+80)/3 = 88.33 > 80, should be "happy"
	gameState := char.GetGameState()
	if gameState != nil {
		gameState.Stats["hunger"].Current = 90.0
		gameState.Stats["happiness"].Current = 95.0
		gameState.Stats["energy"].Current = 80.0
	}

	// Force character into a non-idle state
	err = char.ForceState("talking")
	if err != nil {
		t.Fatalf("Failed to force talking state: %v", err)
	}

	// Wait for idle timeout and update
	time.Sleep(time.Second * 2)
	char.Update()

	// Should now be in mood-based idle animation ("happy")
	currentState := char.GetCurrentState()
	if currentState != "happy" {
		t.Errorf("After idle timeout with high mood, should be in 'happy' state, got: %s", currentState)
	}
}

// TestBug1AnimationLoadingGracefulDegradation is a regression test for
// the critical bug where character creation failed entirely if any required
// animation files were missing. Now it should succeed with graceful degradation.
func TestBug1AnimationLoadingGracefulDegradation(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir := t.TempDir()

	// Create character card directory
	cardDir := filepath.Join(tempDir, "test_character")
	if err := os.MkdirAll(cardDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create animations directory and invalid animation files
	animDir := filepath.Join(cardDir, "animations")
	if err := os.MkdirAll(animDir, 0755); err != nil {
		t.Fatalf("Failed to create animations directory: %v", err)
	}

	// Create animation files that exist but are empty (will fail to load as GIFs)
	idlePath := filepath.Join(animDir, "idle.gif")
	talkingPath := filepath.Join(animDir, "talking.gif")

	// Create invalid GIF files (empty files that will fail to load)
	if err := os.WriteFile(idlePath, []byte("invalid gif data"), 0644); err != nil {
		t.Fatalf("Failed to create idle animation file: %v", err)
	}
	if err := os.WriteFile(talkingPath, []byte("invalid gif data"), 0644); err != nil {
		t.Fatalf("Failed to create talking animation file: %v", err)
	}

	// Create a basic character card file
	cardPath := filepath.Join(cardDir, "character.json")
	cardContent := `{
		"name": "Test Character",
		"description": "A test character for bug reproduction",
		"animations": {
			"idle": "animations/idle.gif",
			"talking": "animations/talking.gif"
		},
		"dialogs": [
			{
				"trigger": "click",
				"responses": ["Hello!"],
				"animation": "talking",
				"cooldown": 5
			}
		],
		"behavior": {
			"defaultSize": 200,
			"idleTimeout": 30,
			"actionsEnabled": true
		}
	}`

	if err := os.WriteFile(cardPath, []byte(cardContent), 0644); err != nil {
		t.Fatalf("Failed to create character card: %v", err)
	}

	// Load the character card
	card, err := character.LoadCard(cardPath)
	if err != nil {
		t.Fatalf("Failed to load character card: %v", err)
	}

	// Test character creation - this should now succeed with graceful degradation
	char, err := character.New(card, cardDir)

	// FIXED: Character creation should succeed with graceful degradation
	// even when all animations fail to load
	if err != nil {
		t.Errorf("Character creation should succeed with graceful degradation, but failed: %v", err)
	} else {
		// After the fix, this should succeed with graceful degradation
		t.Logf("Character creation succeeded with graceful degradation")
		if char == nil {
			t.Errorf("Character should not be nil after successful creation")
		}
	}
}

// Helper function to create test character card with game features
func createTestGameCharacterCard() *character.CharacterCard {
	return &character.CharacterCard{
		Name:        "Test Game Pet",
		Description: "A test pet with game features",
		Animations: map[string]string{
			"idle":    "idle.gif",
			"talking": "talking.gif",
			"happy":   "happy.gif",
			"sad":     "sad.gif",
			"hungry":  "hungry.gif",
			"eating":  "eating.gif",
		},
		Dialogs: []character.Dialog{
			{
				Trigger:   "click",
				Responses: []string{"Hello!"},
				Animation: "talking",
				Cooldown:  5,
			},
		},
		Behavior: character.Behavior{
			IdleTimeout:     30,
			MovementEnabled: false,
			DefaultSize:     128,
		},
		Stats: map[string]character.StatConfig{
			"hunger": {
				Initial:           100,
				Max:               100,
				DegradationRate:   1.0,
				CriticalThreshold: 20,
			},
			"happiness": {
				Initial:           100,
				Max:               100,
				DegradationRate:   0.8,
				CriticalThreshold: 15,
			},
			"energy": {
				Initial:           100,
				Max:               100,
				DegradationRate:   1.5,
				CriticalThreshold: 25,
			},
		},
		GameRules: &character.GameRulesConfig{
			StatsDecayInterval:             60,
			AutoSaveInterval:               300,
			CriticalStateAnimationPriority: true,
			MoodBasedAnimations:            true,
		},
		Interactions: map[string]character.InteractionConfig{
			"feed": {
				Triggers:   []string{"rightclick"},
				Effects:    map[string]float64{"hunger": 25, "happiness": 5},
				Animations: []string{"eating", "happy"},
				Responses:  []string{"Yum! Thank you!", "That was delicious!"},
				Cooldown:   30,
				Requirements: map[string]map[string]float64{
					"hunger": {"max": 80},
				},
			},
		},
	}
}

// Helper function to create test animation files
func createTestAnimationFiles(t *testing.T, dir string) {
	files := []string{"idle.gif", "talking.gif", "happy.gif", "sad.gif", "hungry.gif", "eating.gif"}

	for _, filename := range files {
		createTestAnimationFile(t, dir, filename)
	}
}

// Helper function to create a single test animation file
func createTestAnimationFile(t *testing.T, dir, filename string) {
	// Create minimal valid GIF data (1x1 pixel, single frame)
	validGIF := []byte{71, 73, 70, 56, 57, 97, 1, 0, 1, 0, 128, 0, 0, 255, 255, 255, 0, 0, 0, 44, 0, 0, 0, 0, 1, 0, 1, 0, 0, 2, 2, 68, 1, 0, 59}

	err := os.WriteFile(filepath.Join(dir, filename), validGIF, 0644)
	if err != nil {
		t.Fatalf("Failed to create test animation file %s: %v", filename, err)
	}
}
