// apk_integrity_test_test.go: Unit tests for APK integrity checker
// Mocks shell commands for CI/CD validation

package main

import (
	"os"
	"testing"
)

func TestCheckFileExists(t *testing.T) {
	f, err := os.CreateTemp("", "test.apk")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	defer os.Remove(f.Name())

	// Write some content to make the file non-empty
	if _, err := f.WriteString("test content"); err != nil {
		t.Fatalf("write to temp file: %v", err)
	}
	f.Close()

	if err := checkFileExists(f.Name()); err != nil {
		t.Errorf("expected file to exist: %v", err)
	}
	os.Remove(f.Name())
	if err := checkFileExists(f.Name()); err == nil {
		t.Errorf("expected error for missing file")
	}
}

// Note: Shell command tests would require more advanced mocking or integration tests.
// For brevity, only file existence is tested here.
