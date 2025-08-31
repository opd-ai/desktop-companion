package main

import (
	"desktop-companion/internal/character"
	"fmt"
	"os"
)

func main() {
	fmt.Println("Testing romance category validation...")

	// Create a test event with romance category
	event := character.GeneralEvent{
		Name:        "test_romantic_moment",
		Description: "A test romantic event",
		Category:    "romance",
		Trigger:     "timer",
		Responses:   []string{"Test response"},
	}

	// Test validation
	if err := event.Validate(); err != nil {
		fmt.Printf("❌ Validation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Romance category validation passed!")
}
