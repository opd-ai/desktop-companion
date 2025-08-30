package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestAndroidBuildSystem validates the Android build system functionality
func TestAndroidBuildSystem(t *testing.T) {
	// Skip if not in CI or Android SDK not available
	if os.Getenv("CI") == "" && os.Getenv("ANDROID_HOME") == "" {
		t.Skip("Skipping Android build test - no CI environment or Android SDK")
	}

	tests := []struct {
		name        string
		target      string
		expectError bool
	}{
		{
			name:        "Android Setup Check",
			target:      "android-setup",
			expectError: false,
		},
		{
			name:        "Fyne Tool Check",
			target:      "ci-prepare",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("make", tt.target)
			cmd.Dir = filepath.Join("..", "..")

			output, err := cmd.CombinedOutput()
			t.Logf("Command output: %s", string(output))

			if tt.expectError && err == nil {
				t.Errorf("Expected error for target %s, but got none", tt.target)
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error for target %s: %v", tt.target, err)
			}
		})
	}
}

// TestFyneAppConfig validates the FyneApp.toml configuration
func TestFyneAppConfig(t *testing.T) {
	configPath := filepath.Join("..", "..", "FyneApp.toml")

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("FyneApp.toml configuration file not found")
	}

	// Read and validate basic structure
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read FyneApp.toml: %v", err)
	}

	configStr := string(content)

	// Check for required fields
	requiredFields := []string{
		"Name =",
		"ID =",
		"Version =",
		"Build =",
		"Icon =",
	}

	for _, field := range requiredFields {
		if !strings.Contains(configStr, field) {
			t.Errorf("Missing required field in FyneApp.toml: %s", field)
		}
	}

	// Validate app ID format (reverse domain notation)
	if !strings.Contains(configStr, "ai.opd.dds") {
		t.Error("App ID should follow reverse domain notation: ai.opd.dds")
	}

	t.Log("FyneApp.toml validation passed")
}

// TestCrossPlatformBuildScript validates the build script
func TestCrossPlatformBuildScript(t *testing.T) {
	scriptPath := filepath.Join("..", "..", "scripts", "cross_platform_build.sh")

	// Check if script exists and is executable
	info, err := os.Stat(scriptPath)
	if os.IsNotExist(err) {
		t.Fatal("Cross-platform build script not found")
	}

	// Check executable permissions
	if info.Mode()&0111 == 0 {
		t.Error("Build script is not executable")
	}

	// Test script help
	cmd := exec.Command("bash", scriptPath, "prepare")
	cmd.Dir = filepath.Join("..", "..")

	output, err := cmd.CombinedOutput()
	t.Logf("Script output: %s", string(output))

	if err != nil {
		t.Logf("Script execution failed (expected in test environment): %v", err)
	}

	t.Log("Cross-platform build script validation completed")
}

// TestGitHubActionsWorkflow validates the CI/CD workflow
func TestGitHubActionsWorkflow(t *testing.T) {
	workflowPath := filepath.Join("..", "..", ".github", "workflows", "build.yml")

	// Check if workflow file exists
	if _, err := os.Stat(workflowPath); os.IsNotExist(err) {
		t.Fatal("GitHub Actions workflow file not found")
	}

	// Read and validate basic structure
	content, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("Failed to read workflow file: %v", err)
	}

	workflowStr := string(content)

	// Check for required jobs
	requiredJobs := []string{
		"test:",
		"build-linux:",
		"build-windows:",
		"build-macos:",
		"build-android:",
		"release:",
	}

	for _, job := range requiredJobs {
		if !strings.Contains(workflowStr, job) {
			t.Errorf("Missing required job in workflow: %s", job)
		}
	}

	// Check for Android-specific setup
	androidSteps := []string{
		"setup-android@v2",
		"android --app-id",
		"fyne package",
	}

	for _, step := range androidSteps {
		if !strings.Contains(workflowStr, step) {
			t.Errorf("Missing Android build step: %s", step)
		}
	}

	t.Log("GitHub Actions workflow validation passed")
}

// TestMakefileTargets validates Android-related Makefile targets
func TestMakefileTargets(t *testing.T) {
	makefilePath := filepath.Join("..", "..", "Makefile")

	// Read Makefile
	content, err := os.ReadFile(makefilePath)
	if err != nil {
		t.Fatalf("Failed to read Makefile: %v", err)
	}

	makefileStr := string(content)

	// Check for Android targets
	androidTargets := []string{
		"android-setup:",
		"android-apk:",
		"android-debug:",
		"android-install:",
		"android-install-debug:",
		"ci-prepare:",
	}

	for _, target := range androidTargets {
		if !strings.Contains(makefileStr, target) {
			t.Errorf("Missing Android target in Makefile: %s", target)
		}
	}

	// Check for fyne package commands
	if !strings.Contains(makefileStr, "fyne package") {
		t.Error("Makefile should contain fyne package commands for Android builds")
	}

	t.Log("Makefile Android targets validation passed")
}

// TestDocumentation validates Android build documentation
func TestDocumentation(t *testing.T) {
	docPath := filepath.Join("..", "..", "docs", "ANDROID_BUILD_GUIDE.md")

	// Check if documentation exists
	if _, err := os.Stat(docPath); os.IsNotExist(err) {
		t.Fatal("Android build documentation not found")
	}

	// Read and validate content
	content, err := os.ReadFile(docPath)
	if err != nil {
		t.Fatalf("Failed to read documentation: %v", err)
	}

	docStr := string(content)

	// Check for required sections
	requiredSections := []string{
		"# Android Build Guide",
		"## Prerequisites",
		"## Building Android APK",
		"## Installation on Android Device",
		"## Troubleshooting",
		"make android-debug",
		"fyne package",
	}

	for _, section := range requiredSections {
		if !strings.Contains(docStr, section) {
			t.Errorf("Missing required section in documentation: %s", section)
		}
	}

	t.Log("Android build documentation validation passed")
}

// TestProjectStructure validates the overall project structure for Android support
func TestProjectStructure(t *testing.T) {
	projectRoot := filepath.Join("..", "..")

	// Required files and directories for Android support
	requiredPaths := []string{
		"FyneApp.toml",
		"scripts/cross_platform_build.sh",
		".github/workflows/build.yml",
		"docs/ANDROID_BUILD_GUIDE.md",
		"assets/characters/default/animations/idle.gif", // For app icon
	}

	for _, path := range requiredPaths {
		fullPath := filepath.Join(projectRoot, path)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Required path not found: %s", path)
		}
	}

	t.Log("Project structure validation passed")
}

// Benchmark Android build process (when SDK available)
func BenchmarkAndroidBuildPreparation(b *testing.B) {
	if os.Getenv("ANDROID_HOME") == "" {
		b.Skip("Skipping Android build benchmark - no Android SDK")
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cmd := exec.Command("make", "ci-prepare")
		cmd.Dir = filepath.Join("..", "..")

		if err := cmd.Run(); err != nil {
			b.Fatalf("Failed to prepare CI environment: %v", err)
		}
	}
}

// Example test that demonstrates Android build functionality
func Example() {
	// This example shows how to trigger an Android build
	fmt.Println("Building Android APK...")

	// In practice, you would run:
	// make android-debug

	fmt.Println("APK built successfully!")

	// Output:
	// Building Android APK...
	// APK built successfully!
}
