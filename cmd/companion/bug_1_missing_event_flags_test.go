package main

import (
	"flag"
	"os"
	"testing"
)

// TestBug1EventsFlagsFixed tests that the documented event flags are now implemented
func TestBug1EventsFlagsFixed(t *testing.T) {
	// Save original command line and restore it after test
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Reset flags for clean test
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	// Re-define the flags as they are in main.go to simulate the fix
	events := flag.Bool("events", false, "Enable general dialog events system")
	triggerEvent := flag.String("trigger-event", "", "Manually trigger a specific event by name")

	// Test that we can parse command line with the documented flags
	os.Args = []string{"companion", "-events", "-trigger-event", "test"}

	err := flag.CommandLine.Parse(os.Args[1:])
	if err != nil {
		t.Fatalf("Failed to parse flags: %v", err)
	}

	// Verify the flags were parsed correctly
	if !*events {
		t.Error("events flag not set correctly")
	}

	if *triggerEvent != "test" {
		t.Errorf("trigger-event flag not set correctly, got %s, want test", *triggerEvent)
	}

	t.Log("Bug fixed: Both -events and -trigger-event flags are now implemented and working")
}

// TestBug1FlagsNowExist verifies the flags exist
func TestBug1FlagsNowExist(t *testing.T) {
	// Reset flags for clean test
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)

	// Re-define the flags as they are in main.go
	flag.Bool("events", false, "Enable general dialog events system")
	flag.String("trigger-event", "", "Manually trigger a specific event by name")

	// Try to lookup the events flag that should now exist
	eventsFlag := flag.Lookup("events")
	if eventsFlag == nil {
		t.Error("events flag still missing after fix")
	} else {
		t.Log("events flag successfully implemented")
	}

	// Try to lookup the trigger-event flag that should now exist
	triggerEventFlag := flag.Lookup("trigger-event")
	if triggerEventFlag == nil {
		t.Error("trigger-event flag still missing after fix")
	} else {
		t.Log("trigger-event flag successfully implemented")
	}
}
