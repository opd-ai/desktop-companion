package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/opd-ai/desktop-companion/internal/dialog"
)

func main() {
	fmt.Println("Dialog Backend Example")

	// Create and configure dialog manager
	manager := setupDialogManager()

	// Create test context
	context := createTestContext()

	// Run all dialog tests
	runDialogGenerationTests(manager, context)
	runTriggerTests(manager, context)
	displayBackendInformation(manager)
}

// setupDialogManager creates and configures the dialog manager with backends.
func setupDialogManager() *dialog.DialogManager {
	manager := dialog.NewDialogManager(true)

	// Create and initialize backends
	simpleBackend := createSimpleBackend()
	markovBackend := createMarkovBackend()

	// Register backends
	manager.RegisterBackend("simple_random", simpleBackend)
	manager.RegisterBackend("markov_chain", markovBackend)

	// Configure manager settings
	if err := manager.SetDefaultBackend("simple_random"); err != nil {
		log.Fatal("Failed to set default backend:", err)
	}

	if err := manager.SetFallbackChain([]string{"markov_chain"}); err != nil {
		log.Fatal("Failed to set fallback chain:", err)
	}

	return manager
}

// createSimpleBackend creates and initializes a simple random backend.
func createSimpleBackend() dialog.DialogBackend {
	backend := dialog.NewSimpleRandomBackend()

	config := map[string]interface{}{
		"type":                 "basic",
		"personalityInfluence": 0.5,
		"responseVariation":    0.3,
		"preferRomanceDialogs": true,
	}

	configJSON, _ := json.Marshal(config)
	if err := backend.Initialize(json.RawMessage(configJSON)); err != nil {
		log.Fatal("Failed to initialize simple backend:", err)
	}

	return backend
}

// createMarkovBackend creates and initializes a Markov chain backend.
func createMarkovBackend() dialog.DialogBackend {
	backend := dialog.NewMarkovChainBackend()

	config := map[string]interface{}{
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

	configJSON, _ := json.Marshal(config)
	if err := backend.Initialize(json.RawMessage(configJSON)); err != nil {
		log.Fatal("Failed to initialize Markov backend:", err)
	}

	return backend
}

// createTestContext creates a test dialog context with sample data.
func createTestContext() dialog.DialogContext {
	return dialog.DialogContext{
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
}

// runDialogGenerationTests tests dialog generation with multiple turns.
func runDialogGenerationTests(manager *dialog.DialogManager, context dialog.DialogContext) {
	fmt.Println("\n--- Testing Dialog Generation ---")

	for i := 0; i < 5; i++ {
		context.InteractionID = fmt.Sprintf("test_%03d", i+1)
		context.ConversationTurn = i + 1

		response, err := manager.GenerateDialog(context)
		if err != nil {
			log.Printf("Error generating dialog: %v", err)
			continue
		}

		displayDialogResponse(i+1, response)
	}
}

// displayDialogResponse formats and displays a dialog response.
func displayDialogResponse(turn int, response dialog.DialogResponse) {
	fmt.Printf("Turn %d:\n", turn)
	fmt.Printf("  Text: %s\n", response.Text)
	fmt.Printf("  Animation: %s\n", response.Animation)
	fmt.Printf("  Confidence: %.2f\n", response.Confidence)
	fmt.Printf("  Type: %s\n", response.ResponseType)
	fmt.Printf("  Tone: %s\n", response.EmotionalTone)
	fmt.Println()
}

// runTriggerTests tests dialog generation with different triggers.
func runTriggerTests(manager *dialog.DialogManager, context dialog.DialogContext) {
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
}

// displayBackendInformation shows information about all registered backends.
func displayBackendInformation(manager *dialog.DialogManager) {
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
