package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Simplified character card structure for validation
type CharacterCard struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Animations  map[string]string `json:"animations"`
	Dialogs     []map[string]any  `json:"dialogs"`
	Behavior    map[string]any    `json:"behavior"`
	Stats       map[string]any    `json:"stats,omitempty"`
	Personality map[string]any    `json:"personality,omitempty"`
}

func validateCharacterCard(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var card CharacterCard
	if err := json.Unmarshal(data, &card); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Basic validation
	if card.Name == "" {
		return fmt.Errorf("character name is required")
	}

	if len(card.Animations) == 0 {
		return fmt.Errorf("at least one animation is required")
	}

	// Check required animations
	required := []string{"idle", "talking", "happy"}
	for _, anim := range required {
		if _, exists := card.Animations[anim]; !exists {
			return fmt.Errorf("required animation '%s' is missing", anim)
		}
	}

	fmt.Printf("✅ %s: Valid character card\n", filepath.Base(filePath))
	fmt.Printf("   Name: %s\n", card.Name)
	fmt.Printf("   Animations: %d\n", len(card.Animations))
	fmt.Printf("   Dialogs: %d\n", len(card.Dialogs))

	// Check for romance features
	if card.Personality != nil {
		fmt.Printf("   🌹 Romance features detected\n")
	}

	return nil
}

func main() {
	characters := []string{
		"assets/characters/tsundere/character.json",
		"assets/characters/flirty/character.json",
		"assets/characters/slow_burn/character.json",
	}

	fmt.Println("=== Character Card Validation ===")

	for _, charPath := range characters {
		if err := validateCharacterCard(charPath); err != nil {
			fmt.Printf("❌ %s: %v\n", filepath.Base(charPath), err)
			os.Exit(1)
		}
	}

	fmt.Println("\n🎉 All character cards are valid!")
}
