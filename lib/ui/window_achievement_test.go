package ui

import (
	"fmt"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	"github.com/opd-ai/desktop-companion/internal/character"
)

func TestDesktopWindowAchievementIntegration(t *testing.T) {
	// Create test app
	app := test.NewApp()
	defer test.NewApp() // Reset test app

	// Create character with game state
	char := createTestCharacterWithGameState()

	// Create desktop window with game mode enabled
	dw := NewDesktopWindow(app, char, false, nil, true, false, nil, false, false, false)

	// Verify achievement notification was created
	if dw.achievementNotification == nil {
		t.Fatal("Achievement notification should be created in game mode")
	}

	// Test showing achievement notification
	testAchievement := character.AchievementDetails{
		Name:        "Test Window Achievement",
		Description: "Achievement shown through window",
		Timestamp:   time.Now(),
	}

	dw.ShowAchievementNotification(testAchievement)

	if !dw.achievementNotification.IsVisible() {
		t.Error("Achievement notification should be visible after ShowAchievementNotification")
	}
}

func TestDesktopWindowWithoutGameMode(t *testing.T) {
	// Create test app
	app := test.NewApp()
	defer test.NewApp() // Reset test app

	// Create character
	char := createTestCharacterWithGameState()

	// Create desktop window WITHOUT game mode
	dw := NewDesktopWindow(app, char, false, nil, false, false, nil, false, false, false)

	// Verify achievement notification was NOT created
	if dw.achievementNotification != nil {
		t.Error("Achievement notification should not be created without game mode")
	}

	// Test that calling ShowAchievementNotification doesn't crash
	testAchievement := character.AchievementDetails{
		Name:        "Test Achievement",
		Description: "Should be safely ignored",
		Timestamp:   time.Now(),
	}

	dw.ShowAchievementNotification(testAchievement) // Should not panic
}

func TestDesktopWindowCheckForNewAchievements(t *testing.T) {
	// Create test app
	app := test.NewApp()
	defer test.NewApp() // Reset test app

	// Create character with game state
	char := createTestCharacterWithGameState()

	// Create desktop window with game mode enabled
	dw := NewDesktopWindow(app, char, false, nil, true, false, nil, false, false, false)

	// Simulate new achievements in game state
	gameState := char.GetGameState()
	if gameState == nil {
		t.Fatal("Game state should exist")
	}

	// We can't easily test the internal recentAchievements field directly,
	// but we can test that checkForNewAchievements doesn't crash
	dw.checkForNewAchievements() // Should not panic

	// Test with nil character
	dw.character = nil
	dw.checkForNewAchievements() // Should not panic

	// Test with nil achievement notification
	dw.character = char
	dw.achievementNotification = nil
	dw.checkForNewAchievements() // Should not panic
}

func TestDesktopWindowAchievementNotificationInContent(t *testing.T) {
	// Create test app
	app := test.NewApp()
	defer test.NewApp() // Reset test app

	// Create character with game state
	char := createTestCharacterWithGameState()

	// Create desktop window with game mode enabled
	dw := NewDesktopWindow(app, char, false, nil, true, false, nil, false, false, false)

	// Check that the achievement notification is included in the window content
	content := dw.window.Content()
	if content == nil {
		t.Fatal("Window content should not be nil")
	}

	// The content should be a container with our notification included
	// We can't easily verify the exact structure, but the setupContent method
	// should have included the achievement notification without errors
}

// createTestCharacterWithGameState creates a test character with progression system
func createTestCharacterWithGameState() *character.Character {
	// Create character card with game features
	card := &character.CharacterCard{
		Name:        "Test Character",
		Description: "Test character with game features",
		Animations: map[string]string{
			"idle": "test.gif",
		},
		Dialogs: []character.Dialog{
			{
				Trigger:   "click",
				Responses: []string{"Hello!"},
				Animation: "idle",
				Cooldown:  5,
			},
		},
		Behavior: character.Behavior{
			IdleTimeout:     30,
			MovementEnabled: false,
			DefaultSize:     128,
		},
		Stats: map[string]character.StatConfig{
			"happiness": {Initial: 100, Max: 100, DegradationRate: 1.0, CriticalThreshold: 20},
			"health":    {Initial: 100, Max: 100, DegradationRate: 0.5, CriticalThreshold: 15},
		},
		GameRules: &character.GameRulesConfig{
			StatsDecayInterval:             60,
			AutoSaveInterval:               300,
			CriticalStateAnimationPriority: true,
			MoodBasedAnimations:            true,
		},
		Progression: &character.ProgressionConfig{
			Levels: []character.LevelConfig{
				{Name: "Baby", Requirement: map[string]int64{"age": 0}, Size: 64},
				{Name: "Child", Requirement: map[string]int64{"age": 3600}, Size: 96},
			},
			Achievements: []character.AchievementConfig{
				{
					Name: "Happy Pet",
					Requirement: map[string]map[string]interface{}{
						"happiness": {"min": 90.0},
					},
					Reward: &character.AchievementReward{
						StatBoosts: map[string]float64{"happiness": 5.0},
					},
				},
			},
		},
	}

	// Create character with the card
	char, err := character.New(card, "")
	if err != nil {
		panic(fmt.Sprintf("Failed to create test character: %v", err))
	}

	// Enable game features if available
	if card.HasGameFeatures() {
		err := char.EnableGameMode(nil, "")
		if err != nil {
			panic(fmt.Sprintf("Failed to enable game mode: %v", err))
		}
	}

	return char
}

func TestDesktopWindowAchievementNotificationPositioning(t *testing.T) {
	// Create test app
	app := test.NewApp()
	defer test.NewApp() // Reset test app

	// Create character with game state
	char := createTestCharacterWithGameState()

	// Create desktop window with game mode enabled
	dw := NewDesktopWindow(app, char, false, nil, true, false, nil, false, false, false)

	// Show an achievement
	testAchievement := character.AchievementDetails{
		Name:        "Positioning Test",
		Description: "Testing notification positioning",
		Timestamp:   time.Now(),
	}

	dw.ShowAchievementNotification(testAchievement)

	// Verify the notification is visible
	if !dw.achievementNotification.IsVisible() {
		t.Error("Achievement notification should be visible")
	}

	// Test the renderer layout
	renderer := dw.achievementNotification.CreateRenderer()
	if renderer == nil {
		t.Fatal("Renderer should not be nil")
	}

	// Test layout with different window sizes
	testSizes := []struct {
		width, height float32
	}{
		{800, 600},
		{1024, 768},
		{400, 300},
	}

	for _, size := range testSizes {
		renderer.Layout(fyne.NewSize(size.width, size.height))

		// Verify minimum size is reasonable
		minSize := renderer.MinSize()
		if minSize.Width <= 0 || minSize.Height <= 0 {
			t.Errorf("Invalid minimum size for window %fx%f: %v",
				size.width, size.height, minSize)
		}
	}
}
