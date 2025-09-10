package main

import (
	"os"
	"strings"
	"testing"
)

// TestBug1HeadlessDisplayCrash reproduces the critical bug where the application
// crashes when run in headless environments without display support
func TestBug1HeadlessDisplayCrash(t *testing.T) {
	// Save original DISPLAY
	originalDisplay := os.Getenv("DISPLAY")
	defer func() {
		if originalDisplay != "" {
			os.Setenv("DISPLAY", originalDisplay)
		} else {
			os.Unsetenv("DISPLAY")
		}
	}()

	// Unset DISPLAY to simulate headless environment
	os.Unsetenv("DISPLAY")

	// This should now gracefully fail instead of panicking
	err := checkDisplayAvailable()
	if err == nil {
		t.Fatal("Expected error when no display available, but checkDisplayAvailable returned nil")
	}

	// Verify it's the expected error message (updated for enhanced Wayland + X11 support)
	expectedMsg := "no display available - neither X11 (DISPLAY) nor Wayland (WAYLAND_DISPLAY) environment is available"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Fatalf("Expected error containing %q, got: %v", expectedMsg, err)
	}

	t.Logf("Got expected error: %v", err)
}

// TestDisplayAvailableWithX11 tests that checkDisplayAvailable works when X11 display is set
func TestDisplayAvailableWithX11(t *testing.T) {
	// Save original DISPLAY
	originalDisplay := os.Getenv("DISPLAY")
	defer func() {
		if originalDisplay != "" {
			os.Setenv("DISPLAY", originalDisplay)
		} else {
			os.Unsetenv("DISPLAY")
		}
	}()

	// Set a mock DISPLAY value
	os.Setenv("DISPLAY", ":0")

	// This should succeed
	err := checkDisplayAvailable()
	if err != nil {
		t.Fatalf("Expected no error when DISPLAY is set, got: %v", err)
	}
}
