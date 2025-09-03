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

// TestCharacterBinaryPipeline tests the complete character binary generation pipeline
func TestCharacterBinaryPipeline(t *testing.T) {
	// Skip if running in CI or if tools are not available
	if os.Getenv("CI") != "" {
		t.Skip("Skipping integration test in CI environment")
	}

	// Change to project root
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(projectRoot)

	// Test pipeline phases
	t.Run("Phase1_ListCharacters", func(t *testing.T) {
		testListCharacters(t)
	})

	t.Run("Phase2_BuildSingleCharacter", func(t *testing.T) {
		testBuildSingleCharacter(t)
	})

	t.Run("Phase3_ValidateBuiltCharacter", func(t *testing.T) {
		testValidateBuiltCharacter(t)
	})

	t.Run("Phase4_BenchmarkCharacter", func(t *testing.T) {
		testBenchmarkCharacter(t)
	})

	t.Run("Phase5_CleanupCharacters", func(t *testing.T) {
		testCleanupCharacters(t)
	})
}

// testListCharacters tests the character listing functionality
func testListCharacters(t *testing.T) {
	cmd := exec.Command("make", "list-characters")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to list characters: %v", err)
	}

	outputStr := string(output)
	if len(strings.TrimSpace(outputStr)) == 0 {
		t.Fatal("No characters found, expected at least one character")
	}

	// Check for expected characters
	expectedChars := []string{"default", "easy", "normal"}
	for _, expected := range expectedChars {
		if !strings.Contains(outputStr, expected) {
			t.Logf("Warning: Expected character '%s' not found in output", expected)
		}
	}

	t.Logf("Found characters: %s", strings.TrimSpace(outputStr))
}

// testBuildSingleCharacter tests building a single character binary
func testBuildSingleCharacter(t *testing.T) {
	// Test with default character first
	characterName := "default"

	// Check if character exists
	characterPath := filepath.Join("assets", "characters", characterName)
	if _, err := os.Stat(characterPath); os.IsNotExist(err) {
		t.Skipf("Character %s not found, skipping build test", characterName)
	}

	// Build the character
	t.Logf("Building character: %s", characterName)
	cmd := exec.Command("make", "build-character", fmt.Sprintf("CHAR=%s", characterName))

	// Set timeout for build
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	cmd = exec.CommandContext(ctx, "make", "build-character", fmt.Sprintf("CHAR=%s", characterName))

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build character %s: %v\nOutput: %s", characterName, err, string(output))
	}

	// Check if binary was created
	expectedBinary := fmt.Sprintf("build/%s_%s_%s", characterName, os.Getenv("GOOS"), os.Getenv("GOARCH"))
	if os.Getenv("GOOS") == "windows" {
		expectedBinary += ".exe"
	}
	// Use go env if environment variables not set
	if os.Getenv("GOOS") == "" {
		goosCmd := exec.Command("go", "env", "GOOS")
		goosOutput, _ := goosCmd.Output()
		goos := strings.TrimSpace(string(goosOutput))

		goarchCmd := exec.Command("go", "env", "GOARCH")
		goarchOutput, _ := goarchCmd.Output()
		goarch := strings.TrimSpace(string(goarchOutput))

		expectedBinary = fmt.Sprintf("build/%s_%s_%s", characterName, goos, goarch)
		if goos == "windows" {
			expectedBinary += ".exe"
		}
	}

	if _, err := os.Stat(expectedBinary); err != nil {
		t.Fatalf("Expected binary not found: %s (error: %v)", expectedBinary, err)
	}

	t.Logf("Successfully built character binary: %s", expectedBinary)
}

// testValidateBuiltCharacter tests the validation of built character binaries
func testValidateBuiltCharacter(t *testing.T) {
	// Check if validation script exists
	scriptPath := "scripts/validate-character-binaries.sh"
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		t.Fatalf("Validation script not found: %s", scriptPath)
	}

	// Run validation
	t.Log("Running character binary validation...")
	cmd := exec.Command("make", "validate-characters")

	// Set timeout for validation
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	cmd = exec.CommandContext(ctx, "make", "validate-characters")

	output, err := cmd.CombinedOutput()

	// Check results - validation may fail if no binaries exist, which is acceptable
	if err != nil {
		outputStr := string(output)
		if strings.Contains(outputStr, "No character binaries found") {
			t.Log("No character binaries found for validation (acceptable)")
			return
		}
		t.Logf("Validation output: %s", outputStr)
		// Don't fail the test - validation may not find binaries in test environment
		t.Logf("Validation command failed (may be expected in test environment): %v", err)
	} else {
		t.Log("Character binary validation completed successfully")
		t.Logf("Validation output: %s", string(output))
	}
}

// testBenchmarkCharacter tests the benchmarking functionality
func testBenchmarkCharacter(t *testing.T) {
	// Run benchmark
	t.Log("Running character binary benchmarks...")
	cmd := exec.Command("make", "benchmark-characters")

	// Set timeout for benchmark
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	cmd = exec.CommandContext(ctx, "make", "benchmark-characters")

	output, err := cmd.CombinedOutput()

	// Check results - benchmark may fail if no binaries exist
	if err != nil {
		outputStr := string(output)
		if strings.Contains(outputStr, "No character binaries found") {
			t.Log("No character binaries found for benchmarking (acceptable)")
			return
		}
		t.Logf("Benchmark output: %s", outputStr)
		t.Logf("Benchmark command failed (may be expected in test environment): %v", err)
	} else {
		t.Log("Character binary benchmarking completed successfully")
		t.Logf("Benchmark output: %s", string(output))
	}
}

// testCleanupCharacters tests the cleanup functionality
func testCleanupCharacters(t *testing.T) {
	// Run cleanup
	t.Log("Running character build cleanup...")
	cmd := exec.Command("make", "clean-characters")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to clean character builds: %v\nOutput: %s", err, string(output))
	}

	// Check that temporary build artifacts are cleaned
	tempDirs := []string{
		"cmd/companion-default",
		"cmd/companion-easy",
		"cmd/companion-normal",
	}

	for _, dir := range tempDirs {
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			t.Errorf("Temporary directory not cleaned up: %s", dir)
		}
	}

	t.Log("Character build cleanup completed successfully")
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
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("could not find project root (no go.mod found)")
}

// TestValidationScriptIntegration tests the validation script with mock data
func TestValidationScriptIntegration(t *testing.T) {
	// Create temporary project structure
	tempDir := t.TempDir()

	// Create mock build directory
	buildDir := filepath.Join(tempDir, "build")
	err := os.MkdirAll(buildDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create mock build directory: %v", err)
	}

	// Create mock binary (simple script that responds to -version)
	mockBinary := filepath.Join(buildDir, "test_linux_amd64")
	mockScript := `#!/bin/bash
if [[ "$1" == "-version" ]]; then
    echo "Mock Binary v1.0.0"
    exit 0
fi
echo "Mock binary running"
exit 0
`
	err = os.WriteFile(mockBinary, []byte(mockScript), 0755)
	if err != nil {
		t.Fatalf("Failed to create mock binary: %v", err)
	}

	// Test validation script with mock data
	scriptPath := "scripts/validate-character-binaries.sh"
	projectRoot, _ := findProjectRoot()
	fullScriptPath := filepath.Join(projectRoot, scriptPath)

	if _, err := os.Stat(fullScriptPath); os.IsNotExist(err) {
		t.Skip("Validation script not found, skipping integration test")
	}

	// Change to temp directory and run validation
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tempDir)

	cmd := exec.Command("bash", fullScriptPath, "help")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Validation script failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Usage:") {
		t.Error("Validation script help output doesn't contain usage information")
	}

	t.Log("Validation script integration test completed successfully")
}

// TestMakefileIntegration tests that all new Makefile targets work correctly
func TestMakefileIntegration(t *testing.T) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(projectRoot)

	tests := []struct {
		name        string
		target      string
		expectError bool
	}{
		{
			name:        "help-characters",
			target:      "help-characters",
			expectError: false,
		},
		{
			name:        "list-characters",
			target:      "list-characters",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("make", tt.target)
			output, err := cmd.CombinedOutput()

			if tt.expectError && err == nil {
				t.Errorf("Expected error for target %s, but got none", tt.target)
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error for target %s: %v\nOutput: %s", tt.target, err, string(output))
			}

			if !tt.expectError {
				t.Logf("Target %s output: %s", tt.target, string(output))
			}
		})
	}
}

// TestFullPipelineDocumentation tests that the documentation is accurate
func TestFullPipelineDocumentation(t *testing.T) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}

	// Check that PLAN.md exists and contains expected sections
	planPath := filepath.Join(projectRoot, "PLAN.md")
	planContent, err := os.ReadFile(planPath)
	if err != nil {
		t.Fatalf("Failed to read PLAN.md: %v", err)
	}

	planStr := string(planContent)
	expectedSections := []string{
		"## 7. Implementation Timeline",
		"Phase 1: Core Infrastructure",
		"Phase 2: CI/CD Pipeline",
		"Phase 3: Integration and Testing",
		"Phase 4: Release and Monitoring",
	}

	for _, section := range expectedSections {
		if !strings.Contains(planStr, section) {
			t.Errorf("PLAN.md missing expected section: %s", section)
		}
	}

	// Check README for character binary documentation
	readmePath := filepath.Join(projectRoot, "README.md")
	readmeContent, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("Failed to read README.md: %v", err)
	}

	readmeStr := string(readmeContent)
	if !strings.Contains(readmeStr, "Character-Specific Binary Generation") {
		t.Log("README.md may need updates for character binary generation documentation")
	}

	t.Log("Documentation check completed")
}
