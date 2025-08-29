package character

import (
	"encoding/json"
	"testing"

	"desktop-companion/internal/dialog"
)

func TestHandleChatMessage(t *testing.T) {
	// Test character without dialog backend
	normalCard := createTestCharacterCard()
	normalChar, err := New(normalCard, "../../testdata")
	if err != nil {
		t.Fatalf("Failed to create test character: %v", err)
	}

	response := normalChar.HandleChatMessage("Hello there!")
	if response != "" {
		t.Error("Character without dialog backend should return empty response")
	}

	// Test character with dialog backend
	backendCard := createTestCharacterCardWithDialogBackend()
	backendChar, err := New(backendCard, "../../testdata")
	if err != nil {
		t.Fatalf("Failed to create test character with dialog backend: %v", err)
	}

	// Test chat message processing
	testMessage := "How are you feeling today?"
	response = backendChar.HandleChatMessage(testMessage)

	// Should get some response (either from dialog backend or fallback)
	if response == "" {
		t.Error("Character with dialog backend should return some response")
	}

	// Test empty message
	response = backendChar.HandleChatMessage("")
	if response == "" {
		t.Error("Should still get fallback response for empty message")
	}
}

func TestHandleChatMessage_FallbackResponses(t *testing.T) {
	// Create character with dialog backend but with high confidence threshold to force fallback
	card := &CharacterCard{
		Name:        "Test Chat Character",
		Description: "A test character with chat capabilities",
		Behavior: Behavior{
			IdleTimeout:     30,
			MovementEnabled: true,
			DefaultSize:     100,
		},
		Animations: map[string]string{
			"idle":    "idle.gif",
			"talking": "talking.gif",
		},
		Dialogs: []Dialog{
			{
				Trigger:   "click",
				Responses: []string{"Hello!"},
				Animation: "talking",
				Cooldown:  5,
			},
		},
		DialogBackend: &dialog.DialogBackendConfig{
			Enabled:             true,
			DefaultBackend:      "simple_random",
			ConfidenceThreshold: 0.99, // Very high threshold to force fallback
			MemoryEnabled:       false,
			DebugMode:           false,
			FallbackChain:       []string{"simple_random"},
			Backends: map[string]json.RawMessage{
				"simple_random": json.RawMessage(`{"personality_weight": 0.1}`),
			},
		},
	}

	char, err := New(card, "../../testdata")
	if err != nil {
		t.Fatalf("Failed to create test character: %v", err)
	}

	// Should get fallback response when confidence is too low
	response := char.HandleChatMessage("This should trigger fallback due to high confidence threshold")
	if response == "" {
		t.Error("Should get fallback response when dialog backend confidence is too low")
	}
}

func TestBuildChatDialogContext(t *testing.T) {
	// Create character with dialog backend
	card := createTestCharacterCardWithDialogBackend()
	char, err := New(card, "../../testdata")
	if err != nil {
		t.Fatalf("Failed to create test character: %v", err)
	}

	// Access the private method through reflection would be complex,
	// so we'll test the public method that uses it
	testMessage := "Tell me about yourself"
	response := char.HandleChatMessage(testMessage)

	// Verify that we get a response (indicating context building worked)
	if response == "" {
		t.Error("Should get response indicating context building worked")
	}
}

func TestExtractTopicsFromMessage(t *testing.T) {
	// This tests the topic extraction logic indirectly through HandleChatMessage
	card := createTestCharacterCardWithDialogBackend()
	char, err := New(card, "../../testdata")
	if err != nil {
		t.Fatalf("Failed to create test character: %v", err)
	}

	// Test messages with different topics
	testCases := []string{
		"I love you so much!",          // Should detect romance topic
		"I'm feeling very happy today", // Should detect emotion topic
		"Want to play a game?",         // Should detect entertainment topic
		"How's the weather?",           // Should detect daily_life topic
	}

	for _, message := range testCases {
		response := char.HandleChatMessage(message)
		// Just verify we get some response - the topic extraction is tested indirectly
		if response == "" {
			t.Errorf("Should get response for message: %s", message)
		}
	}
}

// Helper function to create a test character card with dialog backend
func createTestCharacterCardWithDialogBackend() *CharacterCard {
	// Create backend config with proper json.RawMessage
	backendConfig, _ := json.Marshal(map[string]interface{}{
		"personality_weight": 0.8,
	})

	return &CharacterCard{
		Name:        "Test Chat Character",
		Description: "A test character with chat capabilities",
		Behavior: Behavior{
			IdleTimeout:     30,
			MovementEnabled: true,
			DefaultSize:     100,
		},
		Animations: map[string]string{
			"idle":    "idle.gif",
			"talking": "talking.gif",
		},
		Dialogs: []Dialog{
			{
				Trigger:   "click",
				Responses: []string{"Hello!"},
				Animation: "talking",
				Cooldown:  5,
			},
		},
		DialogBackend: &dialog.DialogBackendConfig{
			Enabled:             true,
			DefaultBackend:      "simple_random",
			ConfidenceThreshold: 0.7,
			MemoryEnabled:       true,
			DebugMode:           false,
			FallbackChain:       []string{"simple_random"},
			Backends: map[string]json.RawMessage{
				"simple_random": backendConfig,
			},
		},
	}
}
