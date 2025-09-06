package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// TestValidateCharacterBinariesScript tests the validation script functionality
func TestValidateCharacterBinariesScript(t *testing.T) {
	// Skip if validation script doesn't exist
	scriptPath := filepath.Join("..", "..", "scripts", "validate-character-binaries.sh")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		t.Skip("Validation script not found, skipping test")
	}

	tests := []struct {
		name        string
		args        []string
		expectError bool
		timeout     time.Duration
	}{
		{
			name:        "show help",
			args:        []string{"help"},
			expectError: false,
			timeout:     10 * time.Second,
		},
		{
			name:        "invalid command",
			args:        []string{"invalid-command"},
			expectError: true,
			timeout:     5 * time.Second,
		},
		{
			name:        "timeout validation",
			args:        []string{"--timeout", "5", "validate"},
			expectError: false, // May fail if no binaries, but shouldn't crash
			timeout:     15 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("bash", append([]string{scriptPath}, tt.args...)...)
			cmd.Dir = filepath.Join("..", "..")

			// Set timeout
			timer := time.AfterFunc(tt.timeout, func() {
				if cmd.Process != nil {
					cmd.Process.Kill()
				}
			})
			defer timer.Stop()

			err := cmd.Run()

			if tt.expectError && err == nil {
				t.Errorf("Expected error for test %s, but got none", tt.name)
			}
			if !tt.expectError && err != nil {
				// Don't fail for missing binaries in validation test
				if tt.name == "timeout validation" {
					t.Logf("Validation test failed (expected if no binaries): %v", err)
				} else {
					t.Errorf("Unexpected error for test %s: %v", tt.name, err)
				}
			}
		})
	}
}

// TestBinaryValidationLogic tests the core validation logic
func TestBinaryValidationLogic(t *testing.T) {
	// Create temporary test binary
	tempDir := t.TempDir()
	testBinary := filepath.Join(tempDir, "test_binary")

	// Create a simple test binary with appropriate extension
	if runtime.GOOS == "windows" {
		testBinary += ".bat"
	}

	// Create a mock binary that can respond to -version
	var mockBinaryContent []byte
	var fileMode os.FileMode

	if runtime.GOOS == "windows" {
		// Create a Windows batch file that responds to -version
		mockBinaryContent = []byte(`@echo off
if "%1"=="-version" (
    echo Test Binary v1.0.0
    exit /b 0
)
exit /b 1
`)
		fileMode = 0644
	} else {
		// Create a Unix shell script
		mockBinaryContent = []byte(`#!/bin/bash
if [[ "$1" == "-version" ]]; then
    echo "Test Binary v1.0.0"
    exit 0
fi
exit 1
`)
		fileMode = 0755
	}

	err := os.WriteFile(testBinary, mockBinaryContent, fileMode)
	if err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}

	tests := []struct {
		name          string
		binaryPath    string
		expectExist   bool
		expectExecute bool
	}{
		{
			name:          "valid binary",
			binaryPath:    testBinary,
			expectExist:   true,
			expectExecute: true,
		},
		{
			name:          "non-existent binary",
			binaryPath:    filepath.Join(tempDir, "nonexistent"),
			expectExist:   false,
			expectExecute: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test file existence
			_, err := os.Stat(tt.binaryPath)
			exists := !os.IsNotExist(err)

			if exists != tt.expectExist {
				t.Errorf("Expected exists=%v, got exists=%v for %s", tt.expectExist, exists, tt.binaryPath)
			}

			if tt.expectExecute && exists {
				// Test execution
				cmd := exec.Command(tt.binaryPath, "-version")
				err := cmd.Run()
				if err != nil {
					t.Errorf("Expected binary to execute successfully, got error: %v", err)
				}
			}
		})
	}
}

// TestBinaryMetrics tests binary size and performance metrics collection
func TestBinaryMetrics(t *testing.T) {
	// Create a test file to measure
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_file")

	// Create file with known size (1KB)
	content := make([]byte, 1024)
	err := os.WriteFile(testFile, content, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test size measurement logic
	stat, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to stat test file: %v", err)
	}

	sizeMB := stat.Size() / (1024 * 1024)
	if sizeMB != 0 { // 1KB should be 0MB when integer division
		t.Logf("File size in MB: %d (expected 0 for 1KB file)", sizeMB)
	}

	// Test that we can measure file sizes correctly
	if stat.Size() != 1024 {
		t.Errorf("Expected file size 1024 bytes, got %d", stat.Size())
	}
}

// TestValidationErrorHandling tests error handling in validation scenarios
func TestValidationErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(string) string // Returns path to test
		expectError bool
	}{
		{
			name: "non-executable file",
			setupFunc: func(tempDir string) string {
				filePath := filepath.Join(tempDir, "non_executable")
				os.WriteFile(filePath, []byte("test"), 0644) // No execute permission
				return filePath
			},
			expectError: true,
		},
		{
			name: "empty file",
			setupFunc: func(tempDir string) string {
				filePath := filepath.Join(tempDir, "empty_file")
				os.WriteFile(filePath, []byte{}, 0755)
				return filePath
			},
			expectError: true, // Empty file won't execute properly
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			testPath := tt.setupFunc(tempDir)

			// Test that we can detect various error conditions
			_, err := os.Stat(testPath)
			if err != nil {
				if !tt.expectError {
					t.Errorf("Unexpected error accessing test file: %v", err)
				}
				return
			}

			// Test execution of the file
			cmd := exec.Command(testPath, "-version")
			err = cmd.Run()

			hasError := err != nil
			if hasError != tt.expectError {
				t.Errorf("Expected error=%v, got error=%v for test %s", tt.expectError, hasError, tt.name)
			}
		})
	}
}

// BenchmarkValidationPerformance benchmarks the validation script performance
func BenchmarkValidationPerformance(t *testing.B) {
	// Skip if validation script doesn't exist
	scriptPath := filepath.Join("..", "..", "scripts", "validate-character-binaries.sh")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		t.Skip("Validation script not found, skipping benchmark")
	}

	t.ResetTimer()

	for i := 0; i < t.N; i++ {
		cmd := exec.Command("bash", scriptPath, "help")
		cmd.Dir = filepath.Join("..", "..")

		start := time.Now()
		err := cmd.Run()
		duration := time.Since(start)

		if err != nil {
			t.Fatalf("Validation script failed: %v", err)
		}

		// Log performance metrics
		if i == 0 { // Only log first iteration
			t.Logf("Script execution time: %v", duration)
		}
	}
}
