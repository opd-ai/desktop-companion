package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"desktop-companion/internal/dialog"
)

func main() {
	fmt.Println("Dialog Backend Example")

	// Create dialog manager
	manager := dialog.NewDialogManager(true)

	// Create and register backends
	simpleBackend := dialog.NewSimpleRandomBackend()
	markovBackend := dialog.NewMarkovChainBackend()

	// Initialize simple backend with basic config
	simpleConfig := map[string]interface{}{
		"type":                 "basic",
		"personalityInfluence": 0.5,
		"responseVariation":    0.3,
		"preferRomanceDialogs": true,
	}
	simpleConfigJSON, _ := json.Marshal(simpleConfig)
	if err := simpleBackend.Initialize(json.RawMessage(simpleConfigJSON)); err != nil {
		log.Fatal("Failed to initialize simple backend:", err)
	}

	// Initialize Markov backend with basic config
	markovConfig := map[string]interface{}{
		"chainOrder":     2,
		"minWords":       3,
		"maxWords":       15,
		"temperatureMin": 0.3,
		"temperatureMax": 0.8,
		"trainingData": []string{
			"Hello! How are you doing today?",
			"It's nice to see you again!",
			"Thank you for spending time with me.",
			"I hope you're having a wonderful day!",
			"What would you like to talk about?",
			"You always know how to make me smile.",
		},
	}
	markovConfigJSON, _ := json.Marshal(markovConfig)
	if err := markovBackend.Initialize(json.RawMessage(markovConfigJSON)); err != nil {
		log.Fatal("Failed to initialize Markov backend:", err)
	}

	// Register backends
	manager.RegisterBackend("simple_random", simpleBackend)
	manager.RegisterBackend("markov_chain", markovBackend)

	// Set default backend
	if err := manager.SetDefaultBackend("simple_random"); err != nil {
		log.Fatal("Failed to set default backend:", err)
	}

	// Set fallback chain
	if err := manager.SetFallbackChain([]string{"markov_chain"}); err != nil {
		log.Fatal("Failed to set fallback chain:", err)
	}

	// Create test context
	context := dialog.DialogContext{
		Trigger:       "click",
		InteractionID: "test_001",
		Timestamp:     time.Now(),
		CurrentStats: map[string]float64{
			"happiness": 75.0,
			"energy":    60.0,
			"affection": 80.0,
		},
		PersonalityTraits: map[string]float64{
			"shyness":     0.3,
			"romanticism": 0.7,
			"helpfulness": 0.8,
			"confidence":  0.6,
		},
		CurrentMood:       75.0,
		RelationshipLevel: "Friend",
		FallbackResponses: []string{
			"Hello there!",
			"Nice to see you!",
			"How can I help you?",
		},
		FallbackAnimation: "talking",
	}

	// Test dialog generation
	fmt.Println("\n--- Testing Dialog Generation ---")
	for i := 0; i < 5; i++ {
		context.InteractionID = fmt.Sprintf("test_%03d", i+1)
		context.ConversationTurn = i + 1

		response, err := manager.GenerateDialog(context)
		if err != nil {
			log.Printf("Error generating dialog: %v", err)
			continue
		}

		fmt.Printf("Turn %d:\n", i+1)
		fmt.Printf("  Text: %s\n", response.Text)
		fmt.Printf("  Animation: %s\n", response.Animation)
		fmt.Printf("  Confidence: %.2f\n", response.Confidence)
		fmt.Printf("  Type: %s\n", response.ResponseType)
		fmt.Printf("  Tone: %s\n", response.EmotionalTone)
		fmt.Println()
	}

	// Test different triggers
	fmt.Println("--- Testing Different Triggers ---")
	triggers := []string{"rightclick", "hover", "compliment", "give_gift", "deep_conversation"}

	for _, trigger := range triggers {
		context.Trigger = trigger
		context.InteractionID = fmt.Sprintf("test_%s", trigger)

		response, err := manager.GenerateDialog(context)
		if err != nil {
			log.Printf("Error generating dialog for %s: %v", trigger, err)
			continue
		}

		fmt.Printf("%s: %s (Animation: %s)\n", trigger, response.Text, response.Animation)
	}

	// Test backend info
	fmt.Println("\n--- Backend Information ---")
	for _, backendName := range manager.GetRegisteredBackends() {
		info, err := manager.GetBackendInfo(backendName)
		if err != nil {
			log.Printf("Error getting info for %s: %v", backendName, err)
			continue
		}

		fmt.Printf("%s v%s: %s\n", info.Name, info.Version, info.Description)
		fmt.Printf("  Capabilities: %v\n", info.Capabilities)
		fmt.Printf("  Author: %s, License: %s\n", info.Author, info.License)
		fmt.Println()
	}
}
