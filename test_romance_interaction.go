package main

import (
	"encoding/json"
	"fmt"
	"log"

	"desktop-companion/internal/character"
)

// Test script to verify HandleRomanceInteraction is working
func main() {
	// Create a romance character card
	cardJSON := `{
		"name": "Romance Test Character",
		"description": "A character for testing romance interactions",
		"animations": {
			"idle": "idle.gif",
			"happy": "happy.gif",
			"blushing": "blushing.gif"
		},
		"dialogs": [
			{
				"trigger": "click",
				"responses": ["Hello there!"],
				"animation": "happy"
			}
		],
		"behavior": {
			"idleTimeout": 30,
			"defaultSize": 128
		},
		"stats": {
			"affection": {"initial": 20, "max": 100, "degradationRate": 0.1},
			"trust": {"initial": 15, "max": 100, "degradationRate": 0.05},
			"happiness": {"initial": 50, "max": 100, "degradationRate": 0.2}
		},
		"interactions": {
			"compliment": {
				"romanceCategory": "verbal_affection",
				"responses": ["That's so sweet of you to say!", "You really think so?"],
				"animations": ["blushing", "happy"],
				"effects": {"affection": 5, "trust": 2},
				"cooldown": 10,
				"requirements": {}
			},
			"give_gift": {
				"romanceCategory": "physical_gift",
				"responses": ["This is beautiful, thank you!", "You remembered!"],
				"animations": ["happy"],
				"effects": {"affection": 8, "happiness": 5},
				"cooldown": 30,
				"requirements": {}
			}
		},
		"personality": {
			"traits": {
				"shyness": 0.6,
				"romanticism": 0.8,
				"affection_responsiveness": 1.2
			},
			"compatibility": {
				"gift_appreciation": 1.5,
				"conversation_lover": 1.3
			}
		}
	}`

	var card character.CharacterCard
	if err := json.Unmarshal([]byte(cardJSON), &card); err != nil {
		log.Fatal("Failed to parse character card:", err)
	}

	// Create character with game features enabled
	char, err := character.NewCharacter(&card, true, true)
	if err != nil {
		log.Fatal("Failed to create character:", err)
	}

	fmt.Println("=== Testing HandleRomanceInteraction ===")
	
	// Test compliment interaction
	fmt.Println("\nTesting compliment interaction:")
	response := char.HandleRomanceInteraction("compliment")
	fmt.Printf("Response: %s\n", response)
	
	if response == "" {
		fmt.Println("❌ No response from compliment interaction")
	} else {
		fmt.Println("✅ Compliment interaction working")
	}

	// Test gift interaction
	fmt.Println("\nTesting gift interaction:")
	response = char.HandleRomanceInteraction("give_gift")
	fmt.Printf("Response: %s\n", response)
	
	if response == "" {
		fmt.Println("❌ No response from gift interaction")
	} else {
		fmt.Println("✅ Gift interaction working")
	}

	// Test invalid interaction
	fmt.Println("\nTesting invalid interaction:")
	response = char.HandleRomanceInteraction("invalid_interaction")
	fmt.Printf("Response: %s\n", response)
	
	if response == "" {
		fmt.Println("✅ Correctly handled invalid interaction (empty response)")
	} else {
		fmt.Println("❌ Should return empty for invalid interaction")
	}

	// Check if romance features are detected
	fmt.Printf("\nRomance features enabled: %v\n", card.HasRomanceFeatures())
	
	fmt.Println("\n=== Romance Interaction Test Complete ===")
}
