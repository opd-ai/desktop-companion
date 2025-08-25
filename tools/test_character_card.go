package main

import (
	"desktop-companion/internal/character"
	"fmt"
	"log"
)

func main() {
	cardPath := "assets/characters/default/character_with_game_features.json"

	card, err := character.LoadCard(cardPath)
	if err != nil {
		log.Fatalf("Failed to load character card: %v", err)
	}

	fmt.Printf("Successfully loaded character: %s\n", card.Name)
	fmt.Printf("Description: %s\n", card.Description)

	if card.Progression != nil {
		fmt.Printf("Progression enabled with %d levels and %d achievements\n",
			len(card.Progression.Levels), len(card.Progression.Achievements))

		for i, level := range card.Progression.Levels {
			fmt.Printf("  Level %d: %s (size: %d, age req: %v)\n",
				i, level.Name, level.Size, level.Requirement)
		}

		for i, achievement := range card.Progression.Achievements {
			fmt.Printf("  Achievement %d: %s (req: %v)\n",
				i, achievement.Name, achievement.Requirement)
		}
	} else {
		fmt.Printf("No progression configured\n")
	}

	fmt.Printf("Validation successful!\n")
}
