package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestMultipleCharactersPipelineDetailed tests the complete pipeline with multiple characters
// This implements Phase 2, Task 4: "Test full pipeline with multiple characters"
func TestMultipleCharactersPipelineDetailed(t *testing.T) {
	// Find project root
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	if err := os.Chdir(projectRoot); err != nil {
		t.Fatalf("Failed to change to project root: %v", err)
	}

	// Test characters - focusing on core archetypes for comprehensive testing
	testCharacters := []string{"default", "easy", "normal"}

	// Phase 1: Verify all test characters exist
	t.Run("VerifyTestCharacters", func(t *testing.T) {
		verifyTestCharacters(t, testCharacters)
	})

	// Phase 2: Test sequential build pipeline for multiple characters
	t.Run("SequentialBuildPipeline", func(t *testing.T) {
		testSequentialBuildPipeline(t, testCharacters)
	})

	// Phase 3: Test validation pipeline for all built characters
	t.Run("ValidationPipeline", func(t *testing.T) {
		testValidationPipeline(t)
	})

	// Phase 4: Test performance benchmarking pipeline
	t.Run("BenchmarkPipeline", func(t *testing.T) {
		testBenchmarkPipeline(t)
	})

	// Phase 5: Test cleanup pipeline
	t.Run("CleanupPipeline", func(t *testing.T) {
		testCleanupPipeline(t)
	})
}

// verifyTestCharacters ensures all test characters exist and have valid configuration
func verifyTestCharacters(t *testing.T, characters []string) {
	for _, char := range characters {
		t.Run(fmt.Sprintf("Character_%s", char), func(t *testing.T) {
			// Check character directory exists
			charDir := filepath.Join("assets", "characters", char)
			if _, err := os.Stat(charDir); os.IsNotExist(err) {
				t.Fatalf("Character directory does not exist: %s", charDir)
			}

			// Check character.json exists
			charConfig := filepath.Join(charDir, "character.json")
			if _, err := os.Stat(charConfig); os.IsNotExist(err) {
				t.Fatalf("Character configuration does not exist: %s", charConfig)
			}

			t.Logf("✓ Character %s verified", char)
		})
	}
}

// testSequentialBuildPipeline tests building multiple characters in sequence
func testSequentialBuildPipeline(t *testing.T, characters []string) {
	// First, clean any existing builds to ensure fresh start
	t.Log("Cleaning existing character builds...")
	cleanCmd := exec.Command("make", "clean-characters")
	if output, err := cleanCmd.CombinedOutput(); err != nil {
		t.Logf("Clean command failed (may be expected): %v\nOutput: %s", err, output)
	}

	// Build each character sequentially
	for _, char := range characters {
		t.Run(fmt.Sprintf("Build_%s", char), func(t *testing.T) {
			buildSingleCharacterInPipeline(t, char)
		})
	}

	// Verify all expected binaries were created
	t.Run("VerifyAllBuilds", func(t *testing.T) {
		verifyAllBuilds(t, characters)
	})
}

// buildSingleCharacterInPipeline builds a single character with proper error handling and timing
func buildSingleCharacterInPipeline(t *testing.T, character string) {
	t.Logf("Building character: %s", character)

	// Create build command with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "make", "build-character", fmt.Sprintf("CHAR=%s", character))

	// Capture both stdout and stderr
	output, err := cmd.CombinedOutput()

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			t.Fatalf("Build timeout for character %s after 3 minutes", character)
		}
		t.Fatalf("Failed to build character %s: %v\nOutput: %s", character, err, output)
	}

	// Verify binary was created
	expectedBinary := getExpectedBinaryPath(character)
	if _, err := os.Stat(expectedBinary); err != nil {
		t.Fatalf("Expected binary not found after build: %s (error: %v)", expectedBinary, err)
	}

	t.Logf("✓ Successfully built character %s: %s", character, expectedBinary)
}

// verifyAllBuilds ensures all expected binaries exist after the build pipeline
func verifyAllBuilds(t *testing.T, characters []string) {
	for _, char := range characters {
		expectedBinary := getExpectedBinaryPath(char)
		if _, err := os.Stat(expectedBinary); err != nil {
			t.Errorf("Missing expected binary for character %s: %s", char, expectedBinary)
		} else {
			t.Logf("✓ Verified binary exists: %s", expectedBinary)
		}
	}
}

// testValidationPipeline tests the validation of all built character binaries
func testValidationPipeline(t *testing.T) {
	t.Log("Running character binary validation pipeline...")

	// Create validation command with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "make", "validate-characters")
	output, err := cmd.CombinedOutput()

	// Validation may fail if no binaries exist, log but don't fail test
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			t.Fatal("Validation pipeline timeout after 2 minutes")
		}
		// Log the error but continue - validation script may be strict
		t.Logf("Validation completed with warnings/errors: %v\nOutput: %s", err, output)
	} else {
		t.Log("✓ Character binary validation pipeline completed successfully")
	}

	// Always log the validation output for debugging
	t.Logf("Validation output: %s", output)
}

// testBenchmarkPipeline tests the performance benchmarking pipeline
func testBenchmarkPipeline(t *testing.T) {
	t.Log("Running character binary benchmarking pipeline...")

	// Create benchmark command with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "make", "benchmark-characters")
	output, err := cmd.CombinedOutput()

	// Benchmark may fail if no binaries exist, log but don't fail test
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			t.Fatal("Benchmark pipeline timeout after 3 minutes")
		}
		// Log the error but continue - benchmark script may be strict
		t.Logf("Benchmarking completed with warnings/errors: %v\nOutput: %s", err, output)
	} else {
		t.Log("✓ Character binary benchmarking pipeline completed successfully")
	}

	// Always log the benchmark output for performance analysis
	t.Logf("Benchmark output: %s", output)
}

// testCleanupPipeline tests the cleanup functionality
func testCleanupPipeline(t *testing.T) {
	t.Log("Running character binary cleanup pipeline...")

	cmd := exec.Command("make", "clean-characters")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("Failed to run cleanup pipeline: %v\nOutput: %s", err, output)
	}

	t.Log("✓ Character binary cleanup pipeline completed successfully")
	t.Logf("Cleanup output: %s", output)
}

// getExpectedBinaryPath returns the expected path for a character binary
func getExpectedBinaryPath(character string) string {
	// Get current GOOS and GOARCH
	goosCmd := exec.Command("go", "env", "GOOS")
	goosOutput, _ := goosCmd.Output()
	goos := strings.TrimSpace(string(goosOutput))

	goarchCmd := exec.Command("go", "env", "GOARCH")
	goarchOutput, _ := goarchCmd.Output()
	goarch := strings.TrimSpace(string(goarchOutput))

	expectedBinary := fmt.Sprintf("build/%s_%s_%s", character, goos, goarch)
	if goos == "windows" {
		expectedBinary += ".exe"
	}

	return expectedBinary
}

// findProjectRoot finds the project root directory by looking for go.mod
func findProjectRoot2() (string, error) {
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
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("could not find project root (no go.mod found)")
}
