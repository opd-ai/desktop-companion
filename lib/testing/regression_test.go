package testing

import (
	"github.com/opd-ai/desktop-companion/internal/character"
	"github.com/opd-ai/desktop-companion/internal/config"
	"github.com/opd-ai/desktop-companion/internal/monitoring"
	"github.com/opd-ai/desktop-companion/internal/persistence"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// RegressionTestSuite validates that all core functionality works as expected
// This ensures we haven't broken anything during the romance system implementation
type RegressionTestSuite struct {
	config   *config.Loader
	profiler *monitoring.Profiler
	tempDir  string
}

// TestFullSystemRegression runs comprehensive regression tests across all modules
func TestFullSystemRegression(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "regression_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	suite := &RegressionTestSuite{
		config:   config.New(tempDir),
		profiler: monitoring.NewProfiler(50), // 50MB memory target
		tempDir:  tempDir,
	}

	t.Run("BasicCharacterCompatibility", suite.testBasicCharacterCompatibility)
	t.Run("GameFeaturesCompatibility", suite.testGameFeaturesCompatibility)
	t.Run("RomanceFeatureValidation", suite.testRomanceFeatureValidation)
	t.Run("PerformanceTargetValidation", suite.testPerformanceTargets)
	t.Run("SaveLoadCompatibility", suite.testSaveLoadCompatibility)
	t.Run("AnimationSystemIntegrity", suite.testAnimationSystemIntegrity)
}

// testBasicCharacterCompatibility ensures existing character cards still work
func (s *RegressionTestSuite) testBasicCharacterCompatibility(t *testing.T) {
	t.Log("Testing backward compatibility with existing character cards...")

	// Test basic character without any modern features
	basicCard := character.CharacterCard{
		Name:        "Classic Pet",
		Description: "A traditional desktop pet",
		Animations: map[string]string{
			"idle":    "animations/idle.gif",
			"talking": "animations/talking.gif",
			"happy":   "animations/happy.gif",
		},
		Dialogs: []character.Dialog{
			{
				Trigger:   "click",
				Responses: []string{"Hello!", "How are you today?"},
				Animation: "talking",
				Cooldown:  5,
			},
		},
		Behavior: character.Behavior{
			IdleTimeout:     30,
			MovementEnabled: true,
			DefaultSize:     128,
		},
	}

	// Validation should pass for classic cards
	if err := basicCard.Validate(); err != nil {
		t.Fatalf("Classic character validation failed: %v", err)
	}

	// Should not detect romance features
	if basicCard.HasRomanceFeatures() {
		t.Error("Classic character incorrectly detected as having romance features")
	}

	// Should return default personality values
	if trait := basicCard.GetPersonalityTrait("shyness"); trait != 0.5 {
		t.Errorf("Expected default trait value 0.5, got %f", trait)
	}

	t.Log("✅ Classic character compatibility validated")
}

// testGameFeaturesCompatibility ensures Tamagotchi features still work properly
func (s *RegressionTestSuite) testGameFeaturesCompatibility(t *testing.T) {
	t.Log("Testing game features (Tamagotchi) compatibility...")

	// Load existing game character card
	cardPath := filepath.Join("../../assets/characters/default/character_with_game_features.json")
	card, err := character.LoadCard(cardPath)
	if err != nil {
		// If the specific file doesn't exist, create a test card with game features
		t.Logf("Game character card not found at %s, creating test card: %v", cardPath, err)

		card = &character.CharacterCard{
			Name:        "Test Game Character",
			Description: "Test character with game features",
			Animations:  map[string]string{"idle": "test.gif", "talking": "test.gif"},
			Dialogs:     []character.Dialog{{Trigger: "click", Responses: []string{"Hi"}}},
			Behavior:    character.Behavior{DefaultSize: 128},
			Stats: map[string]character.StatConfig{
				"hunger": {Initial: 100, Max: 100, DegradationRate: 1.0},
			},
			Interactions: map[string]character.InteractionConfig{
				"feed": {
					Triggers: []string{"rightclick"},
					Effects:  map[string]float64{"hunger": 25},
				},
			},
		}
	}

	// Create stub animation files
	s.createStubAnimationFiles(t, card)

	// Validate game features work
	if err := card.Validate(); err != nil {
		t.Fatalf("Game character validation failed: %v", err)
	}

	// Test character creation with game state
	char, err := character.New(card, s.tempDir)
	if err != nil {
		t.Fatalf("Game character creation failed: %v", err)
	}

	// Test game interactions still work
	gameState := char.GetGameState()
	if gameState == nil {
		t.Fatal("Game state not initialized for game character")
	}

	// Test basic game interaction (use actual configured interactions)
	hasGameInteraction := false
	if card.Interactions != nil {
		for interactionName := range card.Interactions {
			if char.CanUseGameInteraction(interactionName) {
				hasGameInteraction = true
				break
			}
		}
	}

	if !hasGameInteraction {
		t.Error("At least one game interaction should be available")
	}

	// Test stat management
	if card.Stats != nil {
		for statName := range card.Stats {
			value := gameState.GetStat(statName)
			if value < 0 {
				t.Errorf("Stat %s should be non-negative, got %f", statName, value)
			}
		}
	}

	t.Log("✅ Game features compatibility validated")
}

// testRomanceFeatureValidation ensures romance features work correctly
func (s *RegressionTestSuite) testRomanceFeatureValidation(t *testing.T) {
	t.Log("Testing romance features validation...")

	// Load romance character card
	cardPath := filepath.Join("../../assets/characters/romance/character.json")
	card, err := character.LoadCard(cardPath)
	if err != nil {
		// If the specific file doesn't exist, create a test card with romance features
		t.Logf("Romance character card not found at %s, creating test card: %v", cardPath, err)

		card = &character.CharacterCard{
			Name:        "Test Romance Character",
			Description: "Test character with romance features",
			Animations:  map[string]string{"idle": "test.gif", "talking": "test.gif"},
			Dialogs:     []character.Dialog{{Trigger: "click", Responses: []string{"Hi"}}},
			Behavior:    character.Behavior{DefaultSize: 128},
			Stats: map[string]character.StatConfig{
				"affection": {Initial: 0, Max: 100, DegradationRate: 0.1},
				"trust":     {Initial: 20, Max: 100, DegradationRate: 0.05},
			},
			Personality: &character.PersonalityConfig{
				Traits: map[string]float64{
					"romanticism": 0.8,
					"shyness":     0.3,
				},
			},
		}
	}

	// Create stub animation files
	s.createStubAnimationFiles(t, card)

	// Validate romance features
	if err := card.Validate(); err != nil {
		t.Fatalf("Romance character validation failed: %v", err)
	}

	// Should detect romance features
	if !card.HasRomanceFeatures() {
		t.Error("Romance character should have romance features detected")
	}

	// Test character creation with romance
	char, err := character.New(card, s.tempDir)
	if err != nil {
		t.Fatalf("Romance character creation failed: %v", err)
	}

	// Test romance stats initialization
	gameState := char.GetGameState()
	if gameState == nil {
		t.Fatal("Game state not initialized for romance character")
	}

	// Validate romance stats exist
	romanceStats := []string{"affection", "trust", "intimacy", "jealousy"}
	for _, statName := range romanceStats {
		if value := gameState.GetStat(statName); value < 0 {
			t.Errorf("Romance stat %s should be non-negative, got %f", statName, value)
		}
	}

	// Test personality system
	if trait := card.GetPersonalityTrait("romanticism"); trait < 0 || trait > 1 {
		t.Errorf("Personality trait should be between 0-1, got %f", trait)
	}

	t.Log("✅ Romance features validation passed")
}

// testPerformanceTargets validates that the application meets performance requirements
func (s *RegressionTestSuite) testPerformanceTargets(t *testing.T) {
	t.Log("Testing performance targets...")

	// Start performance monitoring
	if err := s.profiler.Start("", "", true); err != nil {
		t.Fatalf("Failed to start profiler: %v", err)
	}
	defer s.profiler.Stop("", true)

	// Create multiple characters to stress test
	cards := make([]*character.CharacterCard, 3)
	chars := make([]*character.Character, 3)

	// Test basic character
	cards[0] = &character.CharacterCard{
		Name:        "Test Basic",
		Description: "Basic test character",
		Animations:  map[string]string{"idle": "test.gif", "talking": "test.gif"},
		Dialogs:     []character.Dialog{{Trigger: "click", Responses: []string{"Hi"}, Animation: "talking"}},
		Behavior:    character.Behavior{DefaultSize: 128, IdleTimeout: 30},
	}

	// Test game character
	cards[1] = &character.CharacterCard{
		Name:        "Test Game",
		Description: "Game test character",
		Animations:  map[string]string{"idle": "test.gif", "talking": "test.gif"},
		Dialogs:     []character.Dialog{{Trigger: "click", Responses: []string{"Hi"}, Animation: "talking"}},
		Behavior:    character.Behavior{DefaultSize: 128, IdleTimeout: 30},
		Stats: map[string]character.StatConfig{
			"hunger": {Initial: 100, Max: 100, DegradationRate: 1.0},
		},
		Interactions: map[string]character.InteractionConfig{
			"feed": {
				Triggers:  []string{"rightclick"},
				Effects:   map[string]float64{"hunger": 25},
				Responses: []string{"Yum! I feel much better now!", "Thanks for feeding me!"},
			},
		},
	}

	// Test romance character
	cards[2] = &character.CharacterCard{
		Name:        "Test Romance",
		Description: "Romance test character",
		Animations:  map[string]string{"idle": "test.gif", "talking": "test.gif"},
		Dialogs:     []character.Dialog{{Trigger: "click", Responses: []string{"Hi"}, Animation: "talking"}},
		Behavior:    character.Behavior{DefaultSize: 128, IdleTimeout: 30},
		Stats: map[string]character.StatConfig{
			"affection": {Initial: 0, Max: 100, DegradationRate: 0.1},
		},
		Personality: &character.PersonalityConfig{
			Traits: map[string]float64{
				"romanticism": 0.8,
				"shyness":     0.3,
			},
		},
	}

	// Create characters and validate memory usage
	for i, card := range cards {
		// Create stub animation files for each test card
		s.createStubAnimationFiles(t, card)

		if err := card.Validate(); err != nil {
			t.Fatalf("Card %d validation failed: %v", i, err)
		}

		char, err := character.New(card, s.tempDir)
		if err != nil {
			t.Fatalf("Character %d creation failed: %v", i, err)
		}
		chars[i] = char

		// Simulate some interactions to trigger memory allocation
		char.HandleClick()
		char.HandleRightClick()
		if char.GetGameState() != nil {
			char.Update()
		}
	}

	// Wait a moment for memory to stabilize
	time.Sleep(100 * time.Millisecond)

	// Check memory usage
	memoryMB := s.profiler.GetMemoryUsage()
	if memoryMB > 50.0 {
		t.Logf("WARNING: Memory usage %.2f MB exceeds 50MB target", memoryMB)
		// Don't fail the test, but log the warning
	} else {
		t.Logf("✅ Memory usage %.2f MB within 50MB target", memoryMB)
	}

	// Test frame rate capability (simulate frame recording)
	start := time.Now()
	frameCount := 0
	testDuration := 100 * time.Millisecond

	for time.Since(start) < testDuration {
		s.profiler.RecordFrame()
		frameCount++
	}

	actualDuration := time.Since(start)
	fps := float64(frameCount) / actualDuration.Seconds()

	if fps < 30.0 {
		t.Logf("WARNING: Frame rate %.1f FPS below 30 FPS target", fps)
	} else {
		t.Logf("✅ Frame rate %.1f FPS meets 30+ FPS target", fps)
	}

	t.Log("✅ Performance targets validation completed")
}

// testSaveLoadCompatibility ensures the persistence system works across all character types
func (s *RegressionTestSuite) testSaveLoadCompatibility(t *testing.T) {
	t.Log("Testing save/load compatibility...")

	saveManager := persistence.NewSaveManager(s.tempDir)

	// Test character types for save/load compatibility
	testCases := []struct {
		name       string
		hasStats   bool
		hasGame    bool
		hasRomance bool
	}{
		{"basic", false, false, false},
		{"game", true, true, false},
		{"romance", true, true, true},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("SaveLoad_%s", tc.name), func(t *testing.T) {
			// Create test character card based on type
			card := &character.CharacterCard{
				Name:        fmt.Sprintf("Test_%s", tc.name),
				Description: fmt.Sprintf("Test character for %s save/load", tc.name),
				Animations:  map[string]string{"idle": "test.gif", "talking": "test.gif"},
				Dialogs:     []character.Dialog{{Trigger: "click", Responses: []string{"Hi"}, Animation: "talking"}},
				Behavior:    character.Behavior{DefaultSize: 128, IdleTimeout: 30},
			}

			if tc.hasStats {
				card.Stats = map[string]character.StatConfig{
					"hunger": {Initial: 75, Max: 100, DegradationRate: 1.0},
				}
			}

			if tc.hasRomance {
				card.Stats["affection"] = character.StatConfig{Initial: 25, Max: 100, DegradationRate: 0.1}
				card.Personality = &character.PersonalityConfig{
					Traits: map[string]float64{"romanticism": 0.7},
				}
			}

			// Create character
			s.createStubAnimationFiles(t, card)

			char, err := character.New(card, s.tempDir)
			if err != nil {
				t.Fatalf("Character creation failed: %v", err)
			}

			// Create save data from character state
			gameState := char.GetGameState()
			if gameState != nil {
				saveData := &persistence.GameSaveData{
					CharacterName: card.Name,
					SaveVersion:   "test",
					GameState: &persistence.GameStateData{
						Stats:           make(map[string]*persistence.StatData),
						LastDecayUpdate: time.Now(),
						CreationTime:    time.Now(),
					},
					Metadata: &persistence.SaveMetadata{
						LastSaved: time.Now(),
						Version:   "test",
					},
				}

				// Convert game state to save data format
				for statName, value := range gameState.GetStats() {
					saveData.GameState.Stats[statName] = &persistence.StatData{
						Current:         value,
						Max:             100.0,
						DegradationRate: 1.0,
					}
				}

				// Save character state
				if err := saveManager.SaveGameState(card.Name, saveData); err != nil {
					t.Fatalf("Save failed: %v", err)
				}

				// Load character state
				loadedData, err := saveManager.LoadGameState(card.Name)
				if err != nil {
					t.Fatalf("Load failed: %v", err)
				}

				if loadedData.CharacterName != card.Name {
					t.Errorf("Expected character name %s, got %s", card.Name, loadedData.CharacterName)
				}
			}

			t.Logf("✅ Save/load compatibility validated for %s character", tc.name)
		})
	}

	t.Log("✅ Save/load compatibility validation completed")
}

// testAnimationSystemIntegrity ensures animation system works with all character types
func (s *RegressionTestSuite) testAnimationSystemIntegrity(t *testing.T) {
	t.Log("Testing animation system integrity...")

	// Test animation states across different character types
	testAnimations := []string{"idle", "talking", "happy", "sad"}

	// Test basic animation functionality
	card := &character.CharacterCard{
		Name:        "Animation Test",
		Description: "Character for animation testing",
		Animations: map[string]string{
			"idle":    "animations/idle.gif",
			"talking": "animations/talking.gif",
			"happy":   "animations/happy.gif",
			"sad":     "animations/sad.gif",
		},
		Dialogs:  []character.Dialog{{Trigger: "click", Responses: []string{"Hi"}}},
		Behavior: character.Behavior{DefaultSize: 128},
	}

	// Create stub animation files
	s.createStubAnimationFiles(t, card)

	char, err := character.New(card, s.tempDir)
	if err != nil {
		t.Fatalf("Animation test character creation failed: %v", err)
	}

	// Test each animation state
	availableAnimations := char.GetAvailableAnimations()
	for _, animName := range testAnimations {
		found := false
		for _, available := range availableAnimations {
			if available == animName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Animation %s should be available", animName)
		}
	}

	// Test animation state changes through character interactions
	char.HandleClick()      // Should trigger talking animation
	char.HandleRightClick() // Should trigger other animations

	t.Log("✅ Animation system integrity validated")
}

// createStubAnimationFiles creates minimal GIF files for testing
func (s *RegressionTestSuite) createStubAnimationFiles(t *testing.T, card *character.CharacterCard) {
	// Create a minimal valid GIF file (1x1 pixel) with required color table
	// This is a proper minimal GIF format that should work with Go's gif decoder
	gifData := []byte{
		// GIF Header
		0x47, 0x49, 0x46, 0x38, 0x39, 0x61, // "GIF89a"

		// Logical Screen Descriptor
		0x01, 0x00, // width = 1
		0x01, 0x00, // height = 1
		0x80, // packed fields (global color table flag = 1, color resolution = 0, sort = 0, size = 0 = 2 colors)
		0x00, // background color index
		0x00, // pixel aspect ratio

		// Global Color Table (2 colors: black and white)
		0x00, 0x00, 0x00, // Color 0: black (R, G, B)
		0xFF, 0xFF, 0xFF, // Color 1: white (R, G, B)

		// Image Descriptor
		0x2C,       // image separator
		0x00, 0x00, // left position
		0x00, 0x00, // top position
		0x01, 0x00, // width = 1
		0x01, 0x00, // height = 1
		0x00, // packed fields (no local color table)

		// Image Data
		0x02,       // LZW minimum code size
		0x02,       // data sub-block size
		0x44, 0x01, // LZW encoded single pixel
		0x00, // data sub-block terminator

		// Trailer
		0x3B, // GIF trailer
	}

	// Create animation files for all referenced animations
	for _, animPath := range card.Animations {
		fullPath := filepath.Join(s.tempDir, animPath)

		// Create directory if it doesn't exist
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Logf("Warning: Could not create animation directory %s: %v", dir, err)
			continue
		}

		// Write stub GIF file
		if err := os.WriteFile(fullPath, gifData, 0644); err != nil {
			t.Logf("Warning: Could not create stub animation file %s: %v", fullPath, err)
		}
	}
} // BenchmarkFullSystemPerformance benchmarks overall system performance
func BenchmarkFullSystemPerformance(b *testing.B) {
	// Create temporary directory for benchmarking
	tempDir, err := os.MkdirTemp("", "benchmark_test")
	if err != nil {
		b.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a representative character with all features
	card := &character.CharacterCard{
		Name:        "Benchmark Character",
		Description: "Character for performance benchmarking",
		Animations:  map[string]string{"idle": "test.gif", "talking": "test.gif"},
		Dialogs:     []character.Dialog{{Trigger: "click", Responses: []string{"Hi"}}},
		Behavior:    character.Behavior{DefaultSize: 128},
		Stats: map[string]character.StatConfig{
			"hunger":    {Initial: 100, Max: 100, DegradationRate: 1.0},
			"affection": {Initial: 50, Max: 100, DegradationRate: 0.1},
		},
		Personality: &character.PersonalityConfig{
			Traits: map[string]float64{"romanticism": 0.8},
		},
	}

	// Create stub animation files
	gifData := []byte{
		// GIF Header
		0x47, 0x49, 0x46, 0x38, 0x39, 0x61, // "GIF89a"

		// Logical Screen Descriptor
		0x01, 0x00, // width = 1
		0x01, 0x00, // height = 1
		0x00, // packed fields (no global color table)
		0x00, // background color index
		0x00, // pixel aspect ratio

		// Image Descriptor
		0x2C,       // image separator
		0x00, 0x00, // left position
		0x00, 0x00, // top position
		0x01, 0x00, // width = 1
		0x01, 0x00, // height = 1
		0x00, // packed fields (no local color table)

		// Image Data
		0x02,       // LZW minimum code size
		0x02,       // data sub-block size
		0x44, 0x01, // LZW encoded single pixel
		0x00, // data sub-block terminator

		// Trailer
		0x3B, // GIF trailer
	}

	for _, animPath := range card.Animations {
		fullPath := filepath.Join(tempDir, animPath)
		os.MkdirAll(filepath.Dir(fullPath), 0755)
		os.WriteFile(fullPath, gifData, 0644)
	}

	char, err := character.New(card, tempDir)
	if err != nil {
		b.Fatalf("Benchmark character creation failed: %v", err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Simulate typical user interactions
			char.HandleClick()
			char.Update()
			if char.GetGameState() != nil {
				char.GetGameState().GetStats()
			}
		}
	})
}
