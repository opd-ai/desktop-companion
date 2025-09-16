// Package testdemo provides a deliberate test failure for debugging demonstration
package testdemo

import (
	"testing"
)

// TestDemoFailure deliberately fails to demonstrate debugging process
func TestDemoFailure(t *testing.T) {
	// This test will fail to demonstrate the systematic debugging process
	expected := 42
	actual := 42 // Fixed: corrected value to match expected

	if actual != expected {
		t.Errorf("Expected %d, but got %d", expected, actual)
	}
}

// TestDemoPass passes to show contrast
func TestDemoPass(t *testing.T) {
	expected := "hello"
	actual := "hello"

	if actual != expected {
		t.Errorf("Expected %s, but got %s", expected, actual)
	}
}
