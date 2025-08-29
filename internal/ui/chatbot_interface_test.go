package ui

import (
	"encoding/json"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"

	"desktop-companion/internal/character"
	"desktop-companion/internal/dialog"
)

func TestNewChatbotInterface(t *testing.T) {
	// Test with character card without dialog backend
	normalCard := createTestCharacterCard()

	// Create a mock character for testing (bypass animation loading)
	normalChar := createMockCharacter(normalCard)

	chatbot := NewChatbotInterface(normalChar)

	if chatbot == nil {
		t.Fatal("NewChatbotInterface returned nil")
	}

	if chatbot.IsAvailable() {
		t.Error("Chatbot should not be available for character without dialog backend")
	}

	if chatbot.IsVisible() {
		t.Error("New chatbot interface should be initially hidden")
	}
}

func TestChatbotInterfaceWithDialogBackend(t *testing.T) {
	// Test with character that has dialog backend
	card := createTestCharacterCardWithDialogBackend()
	char := createMockCharacter(card)

	chatbot := NewChatbotInterface(char)

	if !chatbot.IsAvailable() {
		t.Error("Chatbot should be available for character with dialog backend")
	}

	// Test UI components initialization
	if chatbot.available && chatbot.messageInput == nil {
		t.Error("Message input should be initialized for available chatbot")
	}

	if chatbot.available && chatbot.sendButton == nil {
		t.Error("Send button should be initialized for available chatbot")
	}

	if chatbot.available && chatbot.conversationHistory == nil {
		t.Error("Conversation history should be initialized for available chatbot")
	}
}

func TestChatbotInterface_ShowHide(t *testing.T) {
	card := createTestCharacterCardWithDialogBackend()
	char := createMockCharacter(card)

	chatbot := NewChatbotInterface(char)

	// Test show
	chatbot.Show()
	if !chatbot.IsVisible() {
		t.Error("Chatbot should be visible after Show()")
	}

	// Test hide
	chatbot.Hide()
	if chatbot.IsVisible() {
		t.Error("Chatbot should be hidden after Hide()")
	}

	// Test toggle
	chatbot.Toggle()
	if !chatbot.IsVisible() {
		t.Error("Chatbot should be visible after Toggle() from hidden")
	}

	chatbot.Toggle()
	if chatbot.IsVisible() {
		t.Error("Chatbot should be hidden after Toggle() from visible")
	}
}

func TestChatbotInterface_SendMessage(t *testing.T) {
	card := createTestCharacterCardWithDialogBackend()
	char := createMockCharacter(card)

	chatbot := NewChatbotInterface(char)

	// Test empty message (should not add to conversation)
	initialLength := chatbot.GetConversationLength()
	chatbot.messageInput.SetText("")
	chatbot.sendMessage()

	if chatbot.GetConversationLength() != initialLength {
		t.Error("Empty message should not be added to conversation")
	}

	// Test valid message
	testMessage := "Hello, how are you?"
	chatbot.messageInput.SetText(testMessage)
	chatbot.sendMessage()

	if chatbot.GetConversationLength() == initialLength {
		t.Error("Valid message should be added to conversation")
	}

	if chatbot.messageInput.Text != "" {
		t.Error("Message input should be cleared after sending")
	}
}

func TestChatbotInterface_AddMessage(t *testing.T) {
	card := createTestCharacterCardWithDialogBackend()
	char := createMockCharacter(card)

	chatbot := NewChatbotInterface(char)

	// Test adding user message
	userMessage := ChatMessage{
		IsUser:    true,
		Text:      "Test user message",
		Timestamp: time.Now(),
	}

	chatbot.addMessage(userMessage)

	if chatbot.GetConversationLength() != 1 {
		t.Error("Message should be added to conversation")
	}

	// Test adding character message
	charMessage := ChatMessage{
		IsUser:    false,
		Text:      "Test character response",
		Timestamp: time.Now(),
		Animation: "talking",
	}

	chatbot.addMessage(charMessage)

	if chatbot.GetConversationLength() != 2 {
		t.Error("Character message should be added to conversation")
	}
}

func TestChatbotInterface_HistoryManagement(t *testing.T) {
	card := createTestCharacterCardWithDialogBackend()
	char := createMockCharacter(card)

	chatbot := NewChatbotInterface(char)
	chatbot.maxHistoryLength = 3 // Set small limit for testing

	// Add messages exceeding the limit
	for i := 0; i < 5; i++ {
		message := ChatMessage{
			IsUser:    true,
			Text:      "Test message",
			Timestamp: time.Now(),
		}
		chatbot.addMessage(message)
	}

	if chatbot.GetConversationLength() != 3 {
		t.Errorf("Conversation length should be limited to %d, got %d",
			chatbot.maxHistoryLength, chatbot.GetConversationLength())
	}
}

func TestChatbotInterface_ClearHistory(t *testing.T) {
	card := createTestCharacterCardWithDialogBackend()
	char := createMockCharacter(card)

	chatbot := NewChatbotInterface(char)

	// Add some messages
	for i := 0; i < 3; i++ {
		message := ChatMessage{
			IsUser:    true,
			Text:      "Test message",
			Timestamp: time.Now(),
		}
		chatbot.addMessage(message)
	}

	if chatbot.GetConversationLength() == 0 {
		t.Error("Should have messages before clearing")
	}

	chatbot.ClearHistory()

	if chatbot.GetConversationLength() != 0 {
		t.Error("Conversation should be empty after clearing history")
	}
}

func TestChatbotInterface_SetPosition(t *testing.T) {
	card := createTestCharacterCardWithDialogBackend()
	char := createMockCharacter(card)

	chatbot := NewChatbotInterface(char)

	// Test position setting
	testX, testY := float32(100), float32(200)
	chatbot.SetPosition(testX, testY)

	if chatbot.content != nil {
		pos := chatbot.content.Position()
		if pos.X != testX || pos.Y != testY {
			t.Errorf("Position not set correctly: expected (%f, %f), got (%f, %f)",
				testX, testY, pos.X, pos.Y)
		}
	}
}

func TestChatbotInterface_GetToggleButton(t *testing.T) {
	card := createTestCharacterCardWithDialogBackend()
	char := createMockCharacter(card)

	chatbot := NewChatbotInterface(char)

	toggleButton := chatbot.GetToggleButton()
	if toggleButton == nil {
		t.Error("Toggle button should not be nil")
	}

	// Test toggle button functionality
	if chatbot.IsVisible() {
		t.Error("Chatbot should be initially hidden")
	}

	test.Tap(toggleButton)
	if !chatbot.IsVisible() {
		t.Error("Chatbot should be visible after tapping toggle button")
	}

	test.Tap(toggleButton)
	if chatbot.IsVisible() {
		t.Error("Chatbot should be hidden after second tap on toggle button")
	}
}

func TestChatbotInterface_Renderer(t *testing.T) {
	card := createTestCharacterCardWithDialogBackend()
	char := createMockCharacter(card)

	chatbot := NewChatbotInterface(char)
	renderer := chatbot.CreateRenderer()

	if renderer == nil {
		t.Fatal("CreateRenderer should not return nil")
	}

	// Test minimum size
	minSize := renderer.MinSize()
	if chatbot.IsVisible() && (minSize.Width == 0 || minSize.Height == 0) {
		t.Error("MinSize should have non-zero dimensions when visible")
	}

	// Test objects when hidden
	objects := renderer.Objects()
	if chatbot.IsVisible() != (len(objects) > 0) {
		t.Error("Objects should match visibility state")
	}
}

// Helper function to create a test character card with dialog backend
func createTestCharacterCardWithDialogBackend() *character.CharacterCard {
	// Create backend config with proper json.RawMessage
	backendConfig, _ := json.Marshal(map[string]interface{}{
		"personality_weight": 0.8,
	})

	return &character.CharacterCard{
		Name:        "Test Chat Character",
		Description: "A test character with chat capabilities",
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

// Helper function to create a test character card without dialog backend
func createTestCharacterCard() *character.CharacterCard {
	return &character.CharacterCard{
		Name:        "Test Character",
		Description: "A basic test character",
		Behavior: character.Behavior{
			IdleTimeout:     30,
			MovementEnabled: true,
			DefaultSize:     100,
		},
		Animations: map[string]string{
			"idle": "idle.gif",
		},
		Dialogs: []character.Dialog{
			{
				Trigger:   "click",
				Responses: []string{"Hello!"},
				Animation: "idle",
				Cooldown:  5,
			},
		},
	}
}

// Benchmark test for chatbot interface operations
func BenchmarkChatbotInterface_SendMessage(b *testing.B) {
	card := createTestCharacterCardWithDialogBackend()
	char, err := character.New(card, "testpath")
	if err != nil {
		b.Fatalf("Failed to create test character: %v", err)
	}

	chatbot := NewChatbotInterface(char)
	testMessage := "Benchmark test message"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		chatbot.messageInput.SetText(testMessage)
		chatbot.sendMessage()
	}
}

func BenchmarkChatbotInterface_UpdateDisplay(b *testing.B) {
	card := createTestCharacterCardWithDialogBackend()
	char, err := character.New(card, "testpath")
	if err != nil {
		b.Fatalf("Failed to create test character: %v", err)
	}

	chatbot := NewChatbotInterface(char)

	// Add some messages for testing
	for i := 0; i < 10; i++ {
		message := ChatMessage{
			IsUser:    i%2 == 0,
			Text:      "Benchmark message",
			Timestamp: time.Now(),
		}
		chatbot.addMessage(message)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		chatbot.updateConversationDisplay()
	}
}

// createMockCharacter creates a character for testing with actual animation files
func createMockCharacter(card *character.CharacterCard) *character.Character {
	// Use real animation files from testdata directory
	char, err := character.New(card, "../../testdata")
	if err != nil {
		// Return nil if creation fails - tests will handle this
		return nil
	}

	return char
}
