package character

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestGameInteractionFeed tests the feed interaction functionality
func TestGameInteractionFeed(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "game_interaction_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test character card with game features
	card := createTestGameCharacterCard()

	// Create required animation files
	createTestAnimationFiles(t, tmpDir)

	// Create character with game features
	char, err := New(card, tmpDir)
	if err != nil {
		t.Fatalf("Failed to create character: %v", err)
	}

	// Enable game mode
	err = char.EnableGameMode(nil, "")
	if err != nil {
		t.Fatalf("Failed to enable game mode: %v", err)
	}

	// Get initial hunger stat
	initialHunger := char.gameState.GetStat("hunger")

	// Test feed interaction
	response := char.HandleGameInteraction("feed")
	if response == "" {
		t.Error("Feed interaction should return a response")
	}

	// Verify hunger stat increased
	newHunger := char.gameState.GetStat("hunger")
	if newHunger <= initialHunger {
		t.Errorf("Hunger should increase after feeding, got %f -> %f", initialHunger, newHunger)
	}

	// Test cooldown prevention
	response2 := char.HandleGameInteraction("feed")
	if response2 != "" {
		t.Error("Feed interaction should be on cooldown")
	}
}

// TestGameInteractionPlay tests the play interaction functionality
func TestGameInteractionPlay(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "game_interaction_play_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	card := createTestGameCharacterCard()
	createTestAnimationFiles(t, tmpDir)

	char, err := New(card, tmpDir)
	if err != nil {
		t.Fatalf("Failed to create character: %v", err)
	}

	err = char.EnableGameMode(nil, "")
	if err != nil {
		t.Fatalf("Failed to enable game mode: %v", err)
	}

	// Get initial stats
	initialHappiness := char.gameState.GetStat("happiness")
	initialEnergy := char.gameState.GetStat("energy")

	// Test play interaction
	response := char.HandleGameInteraction("play")
	if response == "" {
		t.Error("Play interaction should return a response")
	}

	// Verify happiness increased and energy decreased
	newHappiness := char.gameState.GetStat("happiness")
	newEnergy := char.gameState.GetStat("energy")

	if newHappiness <= initialHappiness {
		t.Errorf("Happiness should increase after playing, got %f -> %f", initialHappiness, newHappiness)
	}

	if newEnergy >= initialEnergy {
		t.Errorf("Energy should decrease after playing, got %f -> %f", initialEnergy, newEnergy)
	}
}

// TestGameInteractionRequirements tests that interaction requirements are enforced
func TestGameInteractionRequirements(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "game_interaction_req_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	card := createTestGameCharacterCard()
	createTestAnimationFiles(t, tmpDir)

	char, err := New(card, tmpDir)
	if err != nil {
		t.Fatalf("Failed to create character: %v", err)
	}

	err = char.EnableGameMode(nil, "")
	if err != nil {
		t.Fatalf("Failed to enable game mode: %v", err)
	}

	// Set hunger to high level (above feed requirement max)
	char.gameState.Stats["hunger"].Current = 90.0

	// Try to feed when hunger is too high
	response := char.HandleGameInteraction("feed")
	if response != "" {
		t.Error("Feed interaction should fail when hunger requirements not met")
	}

	// Set energy to low level (below play requirement min)
	char.gameState.Stats["energy"].Current = 10.0

	// Try to play when energy is too low
	response = char.HandleGameInteraction("play")
	if response != "" {
		t.Error("Play interaction should fail when energy requirements not met")
	}
}

// TestGameInteractionInvalidType tests handling of invalid interaction types
func TestGameInteractionInvalidType(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "game_interaction_invalid_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	card := createTestGameCharacterCard()
	createTestAnimationFiles(t, tmpDir)

	char, err := New(card, tmpDir)
	if err != nil {
		t.Fatalf("Failed to create character: %v", err)
	}

	err = char.EnableGameMode(nil, "")
	if err != nil {
		t.Fatalf("Failed to enable game mode: %v", err)
	}

	// Test invalid interaction type
	response := char.HandleGameInteraction("invalid_interaction")
	if response != "" {
		t.Error("Invalid interaction should return empty response")
	}
}

// TestGameInteractionWithoutGameMode tests interactions when game mode is disabled
func TestGameInteractionWithoutGameMode(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "game_interaction_nogame_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	card := createTestGameCharacterCard()
	createTestAnimationFiles(t, tmpDir)

	char, err := New(card, tmpDir)
	if err != nil {
		t.Fatalf("Failed to create character: %v", err)
	}

	// Don't enable game mode

	// Test game interaction without game mode
	response := char.HandleGameInteraction("feed")
	if response != "" {
		t.Error("Game interactions should not work without game mode enabled")
	}
}

// TestGameStateDegradationIntegration tests that game state updates work in character update loop
func TestGameStateDegradationIntegration(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "game_degradation_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	card := createTestGameCharacterCard()
	createTestAnimationFiles(t, tmpDir)

	char, err := New(card, tmpDir)
	if err != nil {
		t.Fatalf("Failed to create character: %v", err)
	}

	err = char.EnableGameMode(nil, "")
	if err != nil {
		t.Fatalf("Failed to enable game mode: %v", err)
	}

	// Force last decay update to be in the past
	char.gameState.LastDecayUpdate = time.Now().Add(-2 * time.Minute)

	// Get initial hunger
	initialHunger := char.gameState.GetStat("hunger")

	// Update character multiple times
	for i := 0; i < 10; i++ {
		char.Update()
		time.Sleep(10 * time.Millisecond)
	}

	// Verify hunger decreased due to degradation
	finalHunger := char.gameState.GetStat("hunger")
	if finalHunger >= initialHunger {
		t.Errorf("Hunger should decrease over time, got %f -> %f", initialHunger, finalHunger)
	}
}

// TestGameStateAnimationSelection tests that critical states trigger appropriate animations
func TestGameStateAnimationSelection(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "game_animation_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	card := createTestGameCharacterCard()
	createTestAnimationFiles(t, tmpDir)

	char, err := New(card, tmpDir)
	if err != nil {
		t.Fatalf("Failed to create character: %v", err)
	}

	err = char.EnableGameMode(nil, "")
	if err != nil {
		t.Fatalf("Failed to enable game mode: %v", err)
	}

	// Set hunger to critical level
	char.gameState.Stats["hunger"].Current = 10.0 // Below critical threshold of 20

	// Force decay update to trigger critical state
	char.gameState.LastDecayUpdate = time.Now().Add(-2 * time.Minute)

	// Update character to process state changes
	char.Update()

	// Check if character state changed to hungry
	currentState := char.GetCurrentState()
	if currentState != "hungry" && currentState != "sad" {
		t.Errorf("Character should be in critical state animation, got: %s", currentState)
	}
}

// Helper function to create test character card with game features
func createTestGameCharacterCard() *CharacterCard {
	return &CharacterCard{
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
		GameRules: &GameRulesConfig{
			StatsDecayInterval:             60,
			AutoSaveInterval:               300,
			CriticalStateAnimationPriority: true,
			MoodBasedAnimations:            true,
		},
		Interactions: map[string]InteractionConfig{
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
			"play": {
				Triggers:   []string{"doubleclick"},
				Effects:    map[string]float64{"happiness": 20, "energy": -15},
				Animations: []string{"happy"},
				Responses:  []string{"This is fun!", "I love playing!"},
				Cooldown:   45,
				Requirements: map[string]map[string]float64{
					"energy": {"min": 20},
				},
			},
			"pet": {
				Triggers:   []string{"click"},
				Effects:    map[string]float64{"happiness": 10},
				Animations: []string{"happy"},
				Responses:  []string{"That feels nice!", "I love attention!"},
				Cooldown:   15,
			},
		},
	}
}

// Helper function to create test animation files
func createTestAnimationFiles(t *testing.T, dir string) {
	// Create minimal valid GIF data
	validGIF := []byte{71, 73, 70, 56, 57, 97, 1, 0, 1, 0, 128, 0, 0, 255, 255, 255, 0, 0, 0, 44, 0, 0, 0, 0, 1, 0, 1, 0, 0, 2, 2, 68, 1, 0, 59}

	animations := []string{"idle.gif", "talking.gif", "happy.gif", "sad.gif", "hungry.gif", "eating.gif"}
	for _, filename := range animations {
		err := os.WriteFile(filepath.Join(dir, filename), validGIF, 0644)
		if err != nil {
			t.Fatalf("Failed to create test animation file %s: %v", filename, err)
		}
	}
}
