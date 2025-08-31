package scripts

import (
	"os"
	"path/filepath"
	"testing"
)

// Test for Gap #5: Android Build Icon Path Reference Error
func TestGap5AndroidIconPathExists(t *testing.T) {
	// Get the path that Makefile references for Android icon
	iconPath := "assets/characters/default/animations/idle.gif"

	// Check if file exists
	if _, err := os.Stat(iconPath); os.IsNotExist(err) {
		t.Errorf("Android build icon file does not exist: %s", iconPath)
		t.Errorf("This will cause 'make android-apk' to fail")
		t.Errorf("Bug reproduction confirmed: Icon path in Makefile references non-existent file")
		return
	}

	// Additional validation: Check if file is actually a valid GIF
	file, err := os.Open(iconPath)
	if err != nil {
		t.Errorf("Cannot open icon file: %v", err)
		return
	}
	defer file.Close()

	// Read first few bytes to check GIF signature
	header := make([]byte, 3)
	_, err = file.Read(header)
	if err != nil {
		t.Errorf("Cannot read icon file header: %v", err)
		return
	}

	if string(header) != "GIF" {
		t.Errorf("Icon file is not a valid GIF: %s", iconPath)
		t.Errorf("Header: %s (expected: GIF)", string(header))
	}

	t.Logf("SUCCESS: Android icon file exists and appears to be valid GIF: %s", iconPath)
}

// Test that simulates the scenario where user hasn't set up animations yet
func TestGap5AndroidIconPathMissing(t *testing.T) {
	// Temporarily move the file to simulate fresh setup
	iconPath := "assets/characters/default/animations/idle.gif"
	backupPath := iconPath + ".test_backup"

	// Check if file exists first
	if _, err := os.Stat(iconPath); os.IsNotExist(err) {
		t.Logf("Bug confirmed: Icon file does not exist: %s", iconPath)
		t.Logf("Android build will fail with: icon file not found")
		return
	}

	// File exists, so let's test the scenario where it doesn't
	err := os.Rename(iconPath, backupPath)
	if err != nil {
		t.Skipf("Cannot simulate missing file scenario: %v", err)
		return
	}

	// Ensure cleanup
	defer func() {
		os.Rename(backupPath, iconPath)
	}()

	// Now test the missing file scenario
	if _, err := os.Stat(iconPath); os.IsNotExist(err) {
		t.Logf("Bug reproduction successful: Icon file missing: %s", iconPath)
		t.Logf("This scenario occurs when users run 'make android-apk' before animation setup")
		t.Logf("Expected failure: fyne package will fail with icon file not found error")
	}
}

func TestGap5FixValidation(t *testing.T) {
	// Test the proposed fix: Check that we have fallback icon or validation
	iconPath := "assets/characters/default/animations/idle.gif"

	// The fix should ensure either:
	// 1. The icon file always exists (provide default)
	// 2. The Makefile validates the icon exists before building
	// 3. Use a different icon that's guaranteed to exist

	// For now, let's verify current state
	absPath, err := filepath.Abs(iconPath)
	if err != nil {
		t.Errorf("Cannot resolve absolute path for icon: %v", err)
		return
	}

	if _, err := os.Stat(iconPath); err != nil {
		t.Errorf("Icon file issue: %v", err)
		t.Errorf("Absolute path: %s", absPath)

		// Suggest fixes
		t.Logf("Possible fixes:")
		t.Logf("1. Provide a default icon file that always exists")
		t.Logf("2. Add validation to Makefile before build")
		t.Logf("3. Use project icon instead of character animation")
		return
	}

	t.Logf("Current icon file exists: %s", absPath)
}
