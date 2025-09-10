package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/opd-ai/desktop-companion/lib/character"
	"github.com/opd-ai/desktop-companion/lib/dialog"
	"github.com/opd-ai/desktop-companion/lib/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// Example demonstrating the new chatbot interface functionality
func main() {
	// Create Fyne app
	myApp := app.New()
	myWindow := myApp.NewWindow("Chatbot Demo")
	myWindow.Resize(fyne.NewSize(600, 400))

	// Create a character with dialog backend enabled
	card := createTestCharacterWithChatbot()
	char, err := character.New(card, "/workspaces/DDS/testdata")
	if err != nil {
		log.Fatalf("Failed to create character: %v", err)
	}

	// Create chatbot interface
	chatbot := ui.NewChatbotInterface(char)

	// Create demo UI
	content := createDemoUI(chatbot, char)

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}

func createDemoUI(chatbot *ui.ChatbotInterface, char *character.Character) fyne.CanvasObject {
	// Title
	title := widget.NewLabel("Desktop Companion Chatbot Demo")
	title.TextStyle.Bold = true

	// Status info
	status := widget.NewLabel(fmt.Sprintf("Character: %s\nChatbot Available: %v",
		char.GetName(), chatbot.IsAvailable()))

	// Show chatbot interface
	chatbot.Show()

	// Demo buttons
	sendDemo := widget.NewButton("Send Demo Message", func() {
		// Simulate sending a message
		if chatbot.IsAvailable() {
			// We can't directly access private methods, so this demonstrates the concept
			fmt.Println("Demo: User would type message and click send")
		}
	})

	toggleVisibility := widget.NewButton("Toggle Chatbot", func() {
		chatbot.Toggle()
	})

	clearHistory := widget.NewButton("Clear History", func() {
		chatbot.ClearHistory()
	})

	// Instructions
	instructions := widget.NewLabel(
		"Instructions:\n" +
			"1. Type messages in the chat input below\n" +
			"2. Press Enter or click Send to send messages\n" +
			"3. Character will respond using AI dialog backend\n" +
			"4. Conversation history is maintained automatically")

	// Create layout
	buttons := container.NewHBox(sendDemo, toggleVisibility, clearHistory)

	// Get the chatbot container
	chatbotContainer := chatbot.GetContainer()

	return container.NewVBox(
		title,
		status,
		instructions,
		buttons,
		widget.NewSeparator(),
		chatbotContainer,
	)
}

func createTestCharacterWithChatbot() *character.CharacterCard {
	return &character.CharacterCard{
		Name:        "Demo Chatbot Character",
		Description: "A demonstration character with chatbot capabilities",
		Behavior: character.Behavior{
			IdleTimeout:     30,
			MovementEnabled: true,
			DefaultSize:     100,
		},
		Animations: map[string]string{
			"idle":    "idle.gif",
			"talking": "talking.gif",
		},
		Dialogs: []character.Dialog{
			{
				Trigger:   "click",
				Responses: []string{"Hello! You can chat with me using the text box below."},
				Animation: "talking",
				Cooldown:  5,
			},
		},
		DialogBackend: &dialog.DialogBackendConfig{
			Enabled:             true,
			DefaultBackend:      "simple_random",
			ConfidenceThreshold: 0.5,
			MemoryEnabled:       true,
			DebugMode:           true,
			FallbackChain:       []string{"simple_random"},
			Backends:            map[string]json.RawMessage{},
		},
	}
}
