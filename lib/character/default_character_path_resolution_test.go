package character

import (
	"os"
	"path/filepath"
	"testing"
)

// test_default_character_path_resolution_bug reproduces the bug where
// the default character path is resolved relative to current working directory
// instead of the executable location, causing failures when run from different directories
// NOTE: This test verifies the bug exists at the LoadCard level, but the fix
// is implemented at the application level (cmd/companion/main.go)
func TestDefaultCharacterPathResolutionBug(t *testing.T) {
	// Save original working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)

	// Create a temporary directory and change to it
	tmpDir := t.TempDir()
	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Try to load the default character path (this should fail in current implementation)
	defaultPath := "assets/characters/default/character.json"

	// This should fail because assets directory doesn't exist in tmpDir
	_, err = LoadCard(defaultPath)
	if err == nil {
		t.Error("Expected error when loading character from non-existent relative path, but got none")
	}

	// The error should be about file not found, confirming the bug exists at LoadCard level
	if !os.IsNotExist(err) {
		t.Logf("Expected file not found error, got: %v", err)
	}

	// NOTE: The fix for this bug is implemented in cmd/companion/main.go
	// which resolves the path before calling LoadCard, so this test demonstrates
	// the issue exists at the LoadCard level but is fixed at the application level
}

// TestProposedCharacterPathResolutionFix tests the proposed fix
func TestProposedCharacterPathResolutionFix(t *testing.T) {
	// This test demonstrates how the path resolution should work
	// when using executable directory as base instead of working directory

	// Get the current working directory (which should be the project root in tests)
	projectRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Navigate to project root if we're in a subdirectory
	for {
		if _, err := os.Stat(filepath.Join(projectRoot, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(projectRoot)
		if parent == projectRoot {
			t.Fatal("Could not find project root with go.mod")
		}
		projectRoot = parent
	}

	// Construct absolute path to default character
	absolutePath := filepath.Join(projectRoot, "assets", "characters", "default", "character.json")

	// This should work regardless of current working directory
	card, err := LoadCard(absolutePath)
	if err != nil {
		t.Fatalf("Failed to load character with absolute path: %v", err)
	}

	if card == nil {
		t.Error("Character card should not be nil")
	}
}
