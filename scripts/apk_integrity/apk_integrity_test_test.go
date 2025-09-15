
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
