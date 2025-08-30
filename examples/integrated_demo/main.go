package main

import (
	"encoding/json"
	"log"

	"fyne.io/fyne/v2/app"

	"desktop-companion/internal/character"
	"desktop-companion/internal/dialog"
	"desktop-companion/internal/monitoring"
	"desktop-companion/internal/ui"
)

func main() {
	// Create Fyne application
	myApp := app.New()
	myApp.SetIcon(nil) // Remove default icon for cleaner desktop companion appearance

	// Create character with dialog backend enabled
	card := createChatbotDemoCharacter()
	char := createCharacterFromCard(card)
	if char == nil {
		log.Fatal("Failed to create character")
	}

	// Create profiler for monitoring
	profiler := monitoring.NewProfiler(100)

	// Create desktop window with chatbot integration
	// Enable debug mode to see keyboard shortcuts in logs
	// Note: Network mode disabled (false) and no network manager (nil) for this demo
	window := ui.NewDesktopWindow(myApp, char, true, profiler, false, false, nil, false, false)

	log.Println("Desktop companion with chatbot interface started!")
	log.Println("Use keyboard shortcuts:")
	log.Println("  - Press 'C' to toggle chatbot interface")
	log.Println("  - Right-click character for context menu with 'Open Chat' option")
	log.Println("  - Left-click character for basic interaction")

	// Show window and run application
	window.Show()
	myApp.Run()
}

func createChatbotDemoCharacter() *character.CharacterCard {
	// Create backend config for simple random responses
	backendConfig, _ := json.Marshal(map[string]interface{}{
		"personality_weight": 0.8,
	})

	return &character.CharacterCard{
		Name:        "AI Companion",
		Description: "An AI-powered desktop companion with chatbot capabilities",
		Behavior: character.Behavior{
			IdleTimeout:     30,
			MovementEnabled: true,
			DefaultSize:     120,
		},
		Animations: map[string]string{
			"idle":    "idle.gif",
			"talking": "talking.gif",
			"happy":   "happy.gif",
		},
		Dialogs: []character.Dialog{
			{
				Trigger: "click",
				Responses: []string{
					"Hello! I'm your AI companion!",
					"Hi there! Press 'C' to open chat with me!",
					"Greetings! Right-click for more options!",
					"Hey! I have AI chat capabilities now!",
				},
				Animation: "talking",
				Cooldown:  3,
			},
			{
				Trigger: "rightclick",
				Responses: []string{
					"Right-click menu activated! Try the 'Open Chat' option!",
					"Context menu with chat functionality available!",
					"Access my AI chat through the context menu!",
				},
				Animation: "happy",
				Cooldown:  3,
			},
		},
		// Enable AI dialog backend for chatbot functionality
		DialogBackend: &dialog.DialogBackendConfig{
			Enabled:             true,
			DefaultBackend:      "simple_random",
			ConfidenceThreshold: 0.7,
			MemoryEnabled:       true,
			DebugMode:           true, // Enable debug for demo
			FallbackChain:       []string{"simple_random"},
			Backends: map[string]json.RawMessage{
				"simple_random": backendConfig,
			},
		},
	}
}

func createCharacterFromCard(card *character.CharacterCard) *character.Character {
	// Try to create character with animation files from assets
	char, err := character.New(card, "../../assets/characters/default/animations")
	if err != nil {
		// Fallback to testdata if assets not available
		char, err = character.New(card, "../../testdata")
		if err != nil {
			log.Printf("Warning: Failed to create character with animations: %v", err)
			log.Println("Character will work but may not have visual animations")
			return nil
		}
	}

	return char
}
