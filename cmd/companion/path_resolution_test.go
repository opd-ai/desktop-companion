package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestDefaultCharacterPathResolutionInApplication tests the bug fix
// for default character path resolution when running from different directories
func TestDefaultCharacterPathResolutionInApplication(t *testing.T) {
	// Build the application first
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}

	binaryPath := filepath.Join(projectRoot, "build", "test_companion_path_resolution")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, "cmd/companion/main.go")
	buildCmd.Dir = projectRoot

	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build application: %v\nOutput: %s", err, output)
	}
	defer os.Remove(binaryPath) // Clean up

	// Test 1: Run from project root (should work)
	cmd1 := exec.Command(binaryPath, "-version")
	cmd1.Dir = projectRoot
	output1, err1 := cmd1.CombinedOutput()
	if err1 != nil {
		t.Errorf("Application failed when run from project root: %v\nOutput: %s", err1, output1)
	}
	if !strings.Contains(string(output1), "Desktop Companion") {
		t.Errorf("Expected version output, got: %s", output1)
	}

	// Test 2: Run from different directory (should still work after fix)
	tmpDir := t.TempDir()
	cmd2 := exec.Command(binaryPath, "-version")
	cmd2.Dir = tmpDir
	output2, err2 := cmd2.CombinedOutput()
	if err2 != nil {
		t.Errorf("Application failed when run from different directory: %v\nOutput: %s", err2, output2)
	}
	if !strings.Contains(string(output2), "Desktop Companion") {
		t.Errorf("Expected version output, got: %s", output2)
	}

	// Test 3: Run with debug flag to verify character loading works from different directory
	cmd3 := exec.Command(binaryPath, "-debug", "-version")
	cmd3.Dir = tmpDir
	output3, err3 := cmd3.CombinedOutput()
	if err3 != nil {
		t.Errorf("Application with debug failed when run from different directory: %v\nOutput: %s", err3, output3)
	}
	// Should not contain character loading errors
	if strings.Contains(string(output3), "Failed to load character card") {
		t.Errorf("Character loading failed: %s", output3)
	}
}

// findProjectRoot finds the project root by looking for go.mod
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}
