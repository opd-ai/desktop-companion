package ui

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"fyne.io/fyne/v2"

	"github.com/opd-ai/desktop-companion/lib/character"
)

// TestNewStatsOverlay tests the creation of a new stats overlay
func TestNewStatsOverlay(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "stats_overlay_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create character with game features
	char := createTestCharacterWithGame(t, tmpDir)

	// Create stats overlay
	overlay := NewStatsOverlay(char)

	if overlay == nil {
		t.Fatal("Expected non-nil stats overlay")
	}

	if overlay.character != char {
		t.Error("Expected overlay to reference the correct character")
	}

	if overlay.visible {
		t.Error("Expected overlay to start hidden")
	}

	if len(overlay.progressBars) == 0 {
		t.Error("Expected progress bars to be created for game stats")
	}

	if len(overlay.statLabels) == 0 {
		t.Error("Expected stat labels to be created for game stats")
	}
}

// TestNewStatsOverlayWithoutGameFeatures tests overlay creation for character without game features
func TestNewStatsOverlayWithoutGameFeatures(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "stats_overlay_nogame_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create character without game features
	char := createTestCharacterWithoutGame(t, tmpDir)

	// Create stats overlay
	overlay := NewStatsOverlay(char)

	if overlay == nil {
		t.Fatal("Expected non-nil stats overlay")
	}

	if overlay.character != char {
		t.Error("Expected overlay to reference the correct character")
	}

	// Should have empty widgets for non-game characters
	if len(overlay.progressBars) != 0 {
		t.Error("Expected no progress bars for character without game features")
	}

	if len(overlay.statLabels) != 0 {
		t.Error("Expected no stat labels for character without game features")
	}
}

// TestStatsOverlayToggle tests showing and hiding the overlay
func TestStatsOverlayToggle(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "stats_overlay_toggle_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create character with game features
	char := createTestCharacterWithGame(t, tmpDir)
	overlay := NewStatsOverlay(char)

	// Initially hidden
	if overlay.IsVisible() {
		t.Error("Expected overlay to start hidden")
	}

	// Toggle to show
	overlay.Toggle()
	if !overlay.IsVisible() {
		t.Error("Expected overlay to be visible after first toggle")
	}

	// Toggle to hide
	overlay.Toggle()
	if overlay.IsVisible() {
		t.Error("Expected overlay to be hidden after second toggle")
	}
}

// TestStatsOverlayShow tests explicit show functionality
func TestStatsOverlayShow(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "stats_overlay_show_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create character with game features
	char := createTestCharacterWithGame(t, tmpDir)
	overlay := NewStatsOverlay(char)

	// Ensure cleanup
	defer overlay.Hide()

	// Show overlay
	overlay.Show()

	if !overlay.IsVisible() {
		t.Error("Expected overlay to be visible after Show()")
	}
}

// TestStatsOverlayHide tests explicit hide functionality
func TestStatsOverlayHide(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "stats_overlay_hide_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create character with game features
	char := createTestCharacterWithGame(t, tmpDir)
	overlay := NewStatsOverlay(char)

	// Ensure cleanup
	defer overlay.Hide()

	// Show then hide
	overlay.Show()
	overlay.Hide()

	if overlay.IsVisible() {
		t.Error("Expected overlay to be hidden after Hide()")
	}
}

// TestStatsOverlayUpdateDisplay tests the stat display update functionality
func TestStatsOverlayUpdateDisplay(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "stats_overlay_update_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create character with game features
	char := createTestCharacterWithGame(t, tmpDir)
	overlay := NewStatsOverlay(char)

	// Ensure cleanup
	defer overlay.Hide()

	// Initial stat values should be reflected
	overlay.updateStatDisplay()

	// Verify progress bars have expected values
	gameState := char.GetGameState()
	if gameState != nil {
		stats := gameState.GetStats()
		for statName, expectedValue := range stats {
			if progressBar, exists := overlay.progressBars[statName]; exists {
				expectedPercentage := expectedValue / 100.0
				if progressBar.Value != expectedPercentage {
					t.Errorf("Expected progress bar for %s to have value %.2f, got %.2f",
						statName, expectedPercentage, progressBar.Value)
				}
			}
		}
	}
}

// TestStatsOverlayWithoutGameState tests overlay behavior when character has no game state
func TestStatsOverlayWithoutGameState(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "stats_overlay_nogame_update_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create character without game features
	char := createTestCharacterWithoutGame(t, tmpDir)
	overlay := NewStatsOverlay(char)

	// Show should be safe even without game state
	overlay.Show()
	if overlay.IsVisible() {
		t.Error("Expected overlay to remain hidden for character without game features")
	}

	// Update should be safe even without game state
	overlay.updateStatDisplay() // Should not panic
}

// TestStatsOverlayRenderer tests the Fyne renderer functionality
func TestStatsOverlayRenderer(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "stats_overlay_renderer_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create character with game features
	char := createTestCharacterWithGame(t, tmpDir)
	overlay := NewStatsOverlay(char)

	// Ensure cleanup
	defer overlay.Hide()

	// Test renderer creation
	renderer := overlay.CreateRenderer()
	if renderer == nil {
		t.Fatal("Expected non-nil renderer")
	}

	// Test renderer methods
	minSize := renderer.MinSize()
	if minSize.Width <= 0 || minSize.Height <= 0 {
		t.Error("Expected positive minimum size")
	}

	// Test objects method
	objects := renderer.Objects()
	if len(objects) == 0 {
		t.Error("Expected renderer to have objects")
	}

	// Test layout (should not panic)
	testSize := fyne.NewSize(200, 300)
	renderer.Layout(testSize)

	// Test refresh (should not panic)
	renderer.Refresh()

	// Test destroy (should not panic)
	renderer.Destroy()
}

// TestStatsOverlayGetContainer tests the container accessor
func TestStatsOverlayGetContainer(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "stats_overlay_container_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create character with game features
	char := createTestCharacterWithGame(t, tmpDir)
	overlay := NewStatsOverlay(char)

	// Ensure cleanup
	defer overlay.Hide()

	container := overlay.GetContainer()
	if container == nil {
		t.Error("Expected non-nil container")
	}

	if container != overlay.container {
		t.Error("Expected GetContainer to return the internal container")
	}
}

// TestStatsOverlayUpdateLoop tests the automatic update functionality
func TestStatsOverlayUpdateLoop(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "stats_overlay_loop_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create character with game features
	char := createTestCharacterWithGame(t, tmpDir)
	overlay := NewStatsOverlay(char)

	// Ensure cleanup
	defer overlay.Hide()

	// Show overlay to start update loop
	overlay.Show()

	// Wait briefly to ensure update loop starts
	time.Sleep(100 * time.Millisecond)

	// Verify update loop is running
	if overlay.updateTicker == nil {
		t.Error("Expected update ticker to be started when overlay is shown")
	}

	// Hide overlay to stop update loop
	overlay.Hide()

	// Wait briefly to ensure update loop stops
	time.Sleep(100 * time.Millisecond)

	// Verify update loop is stopped
	if overlay.updateTicker != nil {
		t.Error("Expected update ticker to be stopped when overlay is hidden")
	}
}

// Helper function to create a test character with game features
func createTestCharacterWithGame(t *testing.T, tmpDir string) *character.Character {
	// Create required animation files
	createTestAnimationFiles(t, tmpDir)

	// Create character card with game features
	cardContent := `{
		"name": "Test Pet",
		"description": "A test character for stats overlay",
		"animations": {
			"idle": "idle.gif",
			"talking": "talking.gif",
			"happy": "happy.gif",
			"eating": "eating.gif"
		},
		"stats": {
			"hunger": {"initial": 100, "max": 100, "degradationRate": 1.0, "criticalThreshold": 20},
			"happiness": {"initial": 80, "max": 100, "degradationRate": 0.8, "criticalThreshold": 15}
		},
		"gameRules": {
			"statsDecayInterval": 60,
			"autoSaveInterval": 300
		},
		"interactions": {
			"feed": {
				"triggers": ["rightclick"],
				"effects": {"hunger": 25, "happiness": 5},
				"animations": ["eating"],
				"responses": ["Yum!"],
				"cooldown": 30
			}
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
			"idleTimeout": 30,
			"movementEnabled": false,
			"defaultSize": 128
		}
	}`

	// Write test character card
	cardPath := filepath.Join(tmpDir, "character.json")
	if err := os.WriteFile(cardPath, []byte(cardContent), 0644); err != nil {
		t.Fatalf("Failed to write test character card: %v", err)
	}

	// Load character
	card, err := character.LoadCard(cardPath)
	if err != nil {
		t.Fatalf("Failed to load test character card: %v", err)
	}

	// Create character
	char, err := character.New(card, tmpDir)
	if err != nil {
		t.Fatalf("Failed to create test character: %v", err)
	}

	return char
}

// Helper function to create a test character without game features
func createTestCharacterWithoutGame(t *testing.T, tmpDir string) *character.Character {
	// Create required animation files
	createTestAnimationFiles(t, tmpDir)

	// Create character card without game features
	cardContent := `{
		"name": "Simple Pet",
		"description": "A simple character without game features",
		"animations": {
			"idle": "idle.gif",
			"talking": "talking.gif"
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
			"idleTimeout": 30,
			"movementEnabled": false,
			"defaultSize": 128
		}
	}`

	// Write test character card
	cardPath := filepath.Join(tmpDir, "character.json")
	if err := os.WriteFile(cardPath, []byte(cardContent), 0644); err != nil {
		t.Fatalf("Failed to write test character card: %v", err)
	}

	// Load character
	card, err := character.LoadCard(cardPath)
	if err != nil {
		t.Fatalf("Failed to load test character card: %v", err)
	}

	// Create character
	char, err := character.New(card, tmpDir)
	if err != nil {
		t.Fatalf("Failed to create test character: %v", err)
	}

	return char
}

// TestCapitalizeFirst tests the helper function
func TestCapitalizeFirst(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hunger", "Hunger"},
		{"happiness", "Happiness"},
		{"health", "Health"},
		{"energy", "Energy"},
		{"", ""},
		{"a", "A"},
	}

	for _, test := range tests {
		result := capitalizeFirst(test.input)
		if result != test.expected {
			t.Errorf("capitalizeFirst(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

// TestContains tests the helper function
func TestContains(t *testing.T) {
	slice := []string{"hunger", "health", "energy"}

	if !contains(slice, "hunger") {
		t.Error("Expected contains to return true for 'hunger'")
	}

	if contains(slice, "happiness") {
		t.Error("Expected contains to return false for 'happiness'")
	}

	if contains([]string{}, "anything") {
		t.Error("Expected contains to return false for empty slice")
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
