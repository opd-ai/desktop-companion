package character

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

// Helper function to copy a file
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// TestAnimationManagerStateCorruptionOnLoadFailure reproduces the bug where
// animation loading fails after some animations are loaded successfully,
// leaving the animation manager in an inconsistent state
func TestAnimationManagerStateCorruptionOnLoadFailure(t *testing.T) {
	// Create a test directory with mixed animation files
	tmpDir := t.TempDir()
	animDir := filepath.Join(tmpDir, "animations")
	if err := os.MkdirAll(animDir, 0755); err != nil {
		t.Fatalf("Failed to create animation directory: %v", err)
	}

	// Create some valid GIF files by copying from existing ones
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}

	sourceGif := filepath.Join(projectRoot, "assets", "characters", "default", "animations", "idle.gif")

	// Create one valid animation file
	validFile1 := filepath.Join(animDir, "idle.gif")
	if err := copyFile(sourceGif, validFile1); err != nil {
		t.Fatalf("Failed to create valid animation file: %v", err)
	}

	// Create another valid animation file
	validFile2 := filepath.Join(animDir, "happy.gif")
	if err := copyFile(sourceGif, validFile2); err != nil {
		t.Fatalf("Failed to create valid animation file: %v", err)
	}

	// Create an invalid animation file (corrupted GIF)
	invalidFile := filepath.Join(animDir, "corrupted.gif")
	if err := os.WriteFile(invalidFile, []byte("not a gif file"), 0644); err != nil {
		t.Fatalf("Failed to create corrupted animation file: %v", err)
	}

	// Create animation manager to test
	am := NewAnimationManager()

	// Try to load animations - this should trigger the bug
	// In the current implementation, if any animation fails to load,
	// the animation manager might be left in an inconsistent state

	// Load the valid animations first
	if err := am.LoadAnimation("idle", validFile1); err != nil {
		t.Fatalf("Failed to load valid idle animation: %v", err)
	}

	if err := am.LoadAnimation("happy", validFile2); err != nil {
		t.Fatalf("Failed to load valid happy animation: %v", err)
	}

	// Now try to load the invalid animation - this should fail gracefully
	if err := am.LoadAnimation("corrupted", invalidFile); err == nil {
		t.Error("Expected error when loading corrupted animation, but succeeded")
	}

	// Check that valid animations still work after the failure
	if err := am.SetCurrentAnimation("idle"); err != nil {
		t.Errorf("Valid 'idle' animation should still work after load failure: %v", err)
	}

	if err := am.SetCurrentAnimation("happy"); err != nil {
		t.Errorf("Valid 'happy' animation should still work after load failure: %v", err)
	}

	// Try to set the corrupted animation - should fail gracefully
	if err := am.SetCurrentAnimation("corrupted"); err == nil {
		t.Error("Expected error when setting corrupted animation, but succeeded")
	}
}

// TestCharacterCreationWithPartialAnimationFailure tests the actual bug scenario:
// character creation should succeed even when some animations fail to load
func TestCharacterCreationWithPartialAnimationFailure(t *testing.T) {
	// Create a test directory with mixed animation files
	tmpDir := t.TempDir()
	animDir := filepath.Join(tmpDir, "animations")
	if err := os.MkdirAll(animDir, 0755); err != nil {
		t.Fatalf("Failed to create animation directory: %v", err)
	}

	// Create some valid GIF files by copying from existing ones
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}

	sourceGif := filepath.Join(projectRoot, "assets", "characters", "default", "animations", "idle.gif")

	// Create valid animation files
	validFile1 := filepath.Join(animDir, "idle.gif")
	if err := copyFile(sourceGif, validFile1); err != nil {
		t.Fatalf("Failed to create valid animation file: %v", err)
	}

	validFile2 := filepath.Join(animDir, "happy.gif")
	if err := copyFile(sourceGif, validFile2); err != nil {
		t.Fatalf("Failed to create valid animation file: %v", err)
	}

	// Create invalid animation files
	invalidFile1 := filepath.Join(animDir, "corrupted.gif")
	if err := os.WriteFile(invalidFile1, []byte("not a gif file"), 0644); err != nil {
		t.Fatalf("Failed to create corrupted animation file: %v", err)
	}

	// Create character card with mixed valid/invalid animations
	card := &CharacterCard{
		Name: "Test Character",
		Animations: map[string]string{
			"idle":      "idle.gif",
			"happy":     "happy.gif",
			"corrupted": "corrupted.gif",
			"missing":   "missing.gif", // This file doesn't exist
		},
		Behavior: Behavior{
			IdleTimeout:     300,
			MovementEnabled: true,
			DefaultSize:     200,
		},
	}

	// This should succeed despite some animations failing to load
	char, err := New(card, tmpDir)
	if err != nil {
		t.Fatalf("Character creation should succeed with partial animation failures: %v", err)
	}

	// Check that valid animations were loaded
	loadedAnims := char.animationManager.GetLoadedAnimations()
	if len(loadedAnims) < 2 {
		t.Errorf("Expected at least 2 animations to load, got %d: %v", len(loadedAnims), loadedAnims)
	}

	// Check that valid animations work
	if err := char.animationManager.SetCurrentAnimation("idle"); err != nil {
		t.Errorf("Valid 'idle' animation should work: %v", err)
	}

	if err := char.animationManager.SetCurrentAnimation("happy"); err != nil {
		t.Errorf("Valid 'happy' animation should work: %v", err)
	}

	// Check that invalid animations fail gracefully
	if err := char.animationManager.SetCurrentAnimation("corrupted"); err == nil {
		t.Error("Expected error when setting corrupted animation")
	}

	if err := char.animationManager.SetCurrentAnimation("missing"); err == nil {
		t.Error("Expected error when setting missing animation")
	}
}
