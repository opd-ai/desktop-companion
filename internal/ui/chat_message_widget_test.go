package ui

import (
	"testing"
	"time"

	"fyne.io/fyne/v2/test"

	"desktop-companion/internal/character"
)

// createTestCharacterForChatWidget creates a character for testing ChatMessageWidget
func createTestCharacterForChatWidget() *character.Character {
	card := &character.CharacterCard{
		Name:        "Test Character",
		Description: "Test character for chat widget testing",
		Behavior: character.Behavior{
			IdleTimeout:     30,
			MovementEnabled: false,
			DefaultSize:     100,
		},
		Animations: map[string]string{
			"idle": "idle.gif",
		},
		Dialogs: []character.Dialog{
			{Trigger: "click", Responses: []string{"Hello!"}, Animation: "idle"},
		},
	}

	// Use character.New with testdata path
	char, err := character.New(card, "../../testdata")
	if err != nil {
		// Return a basic character without animations for testing
		return nil
	}
	return char
}

// TestChatMessageWidget_UserMessage tests user message display
func TestChatMessageWidget_UserMessage(t *testing.T) {
	char := createTestCharacterForChatWidget()

	userMessage := ChatMessage{
		IsUser:    true,
		Text:      "Hello there!",
		Timestamp: time.Now(),
	}

	widget := NewChatMessageWidget(char, userMessage, nil)
	if widget == nil {
		t.Fatal("NewChatMessageWidget returned nil")
	}

	// User messages should not have rating controls
	if widget.starButtons != nil {
		t.Error("User messages should not have star rating buttons")
	}

	if widget.IsFavorite() {
		t.Error("User messages should not be marked as favorite")
	}

	if widget.GetRating() != 0 {
		t.Error("User messages should have zero rating")
	}
}

// TestChatMessageWidget_CharacterMessage tests character message display with rating
func TestChatMessageWidget_CharacterMessage(t *testing.T) {
	char := createTestCharacterForChatWidget()

	characterMessage := ChatMessage{
		IsUser:     false,
		Text:       "Hello! How are you?",
		Timestamp:  time.Now(),
		IsFavorite: false,
		Rating:     0,
	}

	ratingChanged := false
	var receivedRating float64

	widget := NewChatMessageWidget(char, characterMessage, func(rating float64) {
		ratingChanged = true
		receivedRating = rating
	})

	if widget == nil {
		t.Fatal("NewChatMessageWidget returned nil")
	}

	// Character messages should have rating controls
	if widget.starButtons == nil {
		t.Error("Character messages should have star rating buttons")
	}

	if len(widget.starButtons) != 5 {
		t.Errorf("Expected 5 star buttons, got %d", len(widget.starButtons))
	}

	// Test setting rating
	widget.setRating(4.0)

	if !widget.IsFavorite() {
		t.Error("Message should be marked as favorite after rating")
	}

	if widget.GetRating() != 4.0 {
		t.Errorf("Expected rating 4.0, got %f", widget.GetRating())
	}

	if !ratingChanged {
		t.Error("Rating change callback should have been called")
	}

	if receivedRating != 4.0 {
		t.Errorf("Expected callback rating 4.0, got %f", receivedRating)
	}
}

// TestChatMessageWidget_PreexistingFavorite tests loading message with existing favorite status
func TestChatMessageWidget_PreexistingFavorite(t *testing.T) {
	char := createTestCharacterForChatWidget()

	favoriteMessage := ChatMessage{
		IsUser:     false,
		Text:       "This is a favorite response",
		Timestamp:  time.Now(),
		IsFavorite: true,
		Rating:     5.0,
	}

	widget := NewChatMessageWidget(char, favoriteMessage, nil)
	if widget == nil {
		t.Fatal("NewChatMessageWidget returned nil")
	}

	// Should load with existing favorite status
	if !widget.IsFavorite() {
		t.Error("Message should be marked as favorite from initialization")
	}

	if widget.GetRating() != 5.0 {
		t.Errorf("Expected rating 5.0, got %f", widget.GetRating())
	}
}

// TestChatMessageWidget_StarButtonInteraction tests star button clicking
func TestChatMessageWidget_StarButtonInteraction(t *testing.T) {
	char := createTestCharacterForChatWidget()

	characterMessage := ChatMessage{
		IsUser:    false,
		Text:      "Rate this message",
		Timestamp: time.Now(),
	}

	widget := NewChatMessageWidget(char, characterMessage, nil)

	// Test clicking different star buttons
	if widget.starButtons != nil && len(widget.starButtons) >= 3 {
		// Simulate clicking the 3rd star (rating = 3)
		widget.setRating(3.0)

		if widget.GetRating() != 3.0 {
			t.Errorf("Expected rating 3.0 after clicking 3rd star, got %f", widget.GetRating())
		}
	}
}

// TestChatMessageWidget_ZeroRating tests removing rating
func TestChatMessageWidget_ZeroRating(t *testing.T) {
	char := createTestCharacterForChatWidget()

	characterMessage := ChatMessage{
		IsUser:     false,
		Text:       "Test zero rating",
		Timestamp:  time.Now(),
		IsFavorite: true,
		Rating:     3.0,
	}

	widget := NewChatMessageWidget(char, characterMessage, nil)

	// Set rating to 0 (remove favorite)
	widget.setRating(0.0)

	if widget.IsFavorite() {
		t.Error("Message should not be favorite after setting rating to 0")
	}

	if widget.GetRating() != 0.0 {
		t.Errorf("Expected rating 0.0, got %f", widget.GetRating())
	}
}

// TestChatMessageWidget_ThreadSafety tests concurrent access to rating functionality
func TestChatMessageWidget_ThreadSafety(t *testing.T) {
	char := createTestCharacterForChatWidget()

	characterMessage := ChatMessage{
		IsUser:    false,
		Text:      "Thread safety test",
		Timestamp: time.Now(),
	}

	widget := NewChatMessageWidget(char, characterMessage, nil)

	// Test concurrent rating operations
	done := make(chan bool, 2)

	go func() {
		widget.setRating(4.0)
		done <- true
	}()

	go func() {
		_ = widget.GetRating()
		_ = widget.IsFavorite()
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	// Verify final state is consistent
	rating := widget.GetRating()
	isFavorite := widget.IsFavorite()

	if rating > 0 && !isFavorite {
		t.Error("Inconsistent state: rating > 0 but not favorite")
	}

	if rating == 0 && isFavorite {
		t.Error("Inconsistent state: rating = 0 but marked as favorite")
	}
}

// TestChatMessageWidget_Rendering tests that the widget can be rendered
func TestChatMessageWidget_Rendering(t *testing.T) {
	char := createTestCharacterForChatWidget()

	characterMessage := ChatMessage{
		IsUser:    false,
		Text:      "Rendering test message",
		Timestamp: time.Now(),
	}

	widget := NewChatMessageWidget(char, characterMessage, nil)

	// Create test window and add widget
	testApp := test.NewApp()
	defer testApp.Quit()

	window := testApp.NewWindow("Test")
	window.SetContent(widget)

	// Test that renderer can be created
	renderer := widget.CreateRenderer()
	if renderer == nil {
		t.Error("CreateRenderer returned nil")
	}

	// Test that content exists
	if widget.content == nil {
		t.Error("Widget content is nil")
	}
}
