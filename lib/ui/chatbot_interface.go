// Package ui provides desktop companion user interface components
// This file implements a chatbot interface widget for AI-enabled characters
package ui

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/opd-ai/desktop-companion/lib/character"
)

// ChatbotInterface provides a multi-line chat interface for AI-enabled characters.
//
// Design Philosophy:
// - Conditional activation only for characters with dialog backend enabled
// - Follows existing widget patterns (DialogBubble, ContextMenu, StatsOverlay)
// - Uses standard Fyne components for library-first approach
// - Integrates with existing character dialog backend infrastructure
// - Provides conversation history for better user experience
//
// Usage:
//
//	chatbot := NewChatbotInterface(character)
//	if chatbot.IsAvailable() {
//	    chatbot.Show()
//	}
//	// User types message and clicks send or presses Enter
//	// Character responds using existing dialog backend system
type ChatbotInterface struct {
	widget.BaseWidget
	character  *character.Character
	background *canvas.Rectangle
	content    *fyne.Container
	visible    bool
	available  bool

	// UI components
	conversationContainer *fyne.Container
	messageInput          *widget.Entry
	sendButton            *widget.Button
	toggleButton          *widget.Button
	historyScroll         *container.Scroll

	// State management
	conversationLog  []ChatMessage
	messageWidgets   []*ChatMessageWidget
	maxHistoryLength int
	lastMessageTime  time.Time
	inputPlaceholder string
}

// ChatMessage represents a single message in the conversation
type ChatMessage struct {
	IsUser     bool      `json:"isUser"`     // true for user messages, false for character responses
	Text       string    `json:"text"`       // Message content
	Timestamp  time.Time `json:"timestamp"`  // When the message was sent/received
	Animation  string    `json:"animation"`  // Animation triggered with character response (if any)
	IsFavorite bool      `json:"isFavorite"` // Whether this response is marked as favorite
	Rating     float64   `json:"rating"`     // User rating for this response (1-5 stars)
}

// NewChatbotInterface creates a new chatbot interface widget.
// The interface is only available for characters with dialog backend enabled.
//
// Returns a fully initialized ChatbotInterface widget that follows Fyne's
// widget pattern and can be added to any container layout.
func NewChatbotInterface(char *character.Character) *ChatbotInterface {
	chatbot := &ChatbotInterface{
		character:        char,
		visible:          false,
		available:        char.GetCard().HasDialogBackend(),
		conversationLog:  make([]ChatMessage, 0),
		maxHistoryLength: 50, // Limit to prevent memory issues
		inputPlaceholder: "Type a message...",
	}

	// ENHANCEMENT: Load recent conversation history from character memory
	if chatbot.available {
		chatbot.loadRecentConversations()
	}

	// Only initialize UI components if chatbot is available
	if chatbot.available {
		chatbot.initializeComponents()
		chatbot.setupLayout()
		chatbot.setupInteractions()
	}

	chatbot.ExtendBaseWidget(chatbot)
	return chatbot
}

// initializeComponents creates the UI components for the chatbot interface
func (c *ChatbotInterface) initializeComponents() {
	// Create background with chat interface styling
	c.background = canvas.NewRectangle(color.RGBA{R: 250, G: 250, B: 250, A: 240})
	c.background.StrokeColor = color.RGBA{R: 150, G: 150, B: 150, A: 255}
	c.background.StrokeWidth = 1

	// Create conversation container for message widgets
	c.conversationContainer = container.NewVBox()
	c.messageWidgets = make([]*ChatMessageWidget, 0)

	// Add initial empty state
	c.updateConversationDisplay()

	// Create scrollable container for conversation history
	c.historyScroll = container.NewScroll(c.conversationContainer)
	c.historyScroll.SetMinSize(fyne.NewSize(300, 150))

	// Create message input field
	c.messageInput = widget.NewMultiLineEntry()
	c.messageInput.SetPlaceHolder(c.inputPlaceholder)
	c.messageInput.Wrapping = fyne.TextWrapWord
	c.messageInput.SetMinRowsVisible(2)

	// Create send button
	c.sendButton = widget.NewButtonWithIcon("Send", theme.MailSendIcon(), func() {
		c.sendMessage()
	})
	c.sendButton.Importance = widget.HighImportance

	// Create toggle visibility button
	c.toggleButton = widget.NewButtonWithIcon("Chat", theme.ComputerIcon(), func() {
		c.Toggle()
	})
	c.toggleButton.Importance = widget.MediumImportance
}

// setupLayout arranges the UI components in the chatbot interface
func (c *ChatbotInterface) setupLayout() {
	// Create input area with send button
	inputArea := container.NewBorder(nil, nil, nil, c.sendButton, c.messageInput)

	// Create main chat content (history + input)
	chatContent := container.NewVBox(c.historyScroll, inputArea)

	// Create main container with background
	c.content = container.NewBorder(nil, nil, nil, nil, c.background, chatContent)

	// Set initial size
	c.content.Resize(fyne.NewSize(320, 220))
	c.content.Move(fyne.NewPos(10, 60)) // Position below character
}

// setupInteractions configures event handlers and keyboard shortcuts
func (c *ChatbotInterface) setupInteractions() {
	// Enter key sends message (Shift+Enter for new line)
	c.messageInput.OnSubmitted = func(text string) {
		if strings.TrimSpace(text) != "" {
			c.sendMessage()
		}
	}

	// Focus management
	c.messageInput.OnCursorChanged = func() {
		// Keep input focused when typing
	}
}

// sendMessage processes the user's message and gets character response
func (c *ChatbotInterface) sendMessage() {
	message := strings.TrimSpace(c.messageInput.Text)
	if message == "" {
		return
	}

	// Clear input field
	c.messageInput.SetText("")

	// Add user message to conversation
	userMessage := ChatMessage{
		IsUser:    true,
		Text:      message,
		Timestamp: time.Now(),
	}
	c.addMessage(userMessage)

	// Get character response
	response := c.character.HandleChatMessage(message)
	if response != "" {
		// Check if this response is already marked as favorite
		isFavorite, rating := false, float64(0)
		if gameState := c.character.GetGameState(); gameState != nil {
			isFavorite, rating = gameState.IsDialogResponseFavorite(response)
		}

		// Add character response to conversation
		characterMessage := ChatMessage{
			IsUser:     false,
			Text:       response,
			Timestamp:  time.Now(),
			Animation:  c.character.GetCurrentState(), // Capture animation used
			IsFavorite: isFavorite,
			Rating:     rating,
		}
		c.addMessage(characterMessage)

		// Record this chat interaction in character memory
		c.character.RecordChatMemory(message, response)
	}

	// Update display and scroll to bottom
	c.updateConversationDisplay()
	c.scrollToBottom()
}

// addMessage adds a message to the conversation log with history management
func (c *ChatbotInterface) addMessage(message ChatMessage) {
	c.conversationLog = append(c.conversationLog, message)
	c.lastMessageTime = message.Timestamp

	// Trim history if it exceeds maximum length
	if len(c.conversationLog) > c.maxHistoryLength {
		// Remove oldest messages, keeping the most recent ones
		c.conversationLog = c.conversationLog[len(c.conversationLog)-c.maxHistoryLength:]
	}

	// Create and add message widget
	messageWidget := NewChatMessageWidget(c.character, message, func(rating float64) {
		c.onMessageRated(message.Text, rating)
	})
	c.messageWidgets = append(c.messageWidgets, messageWidget)
	c.conversationContainer.Add(messageWidget)

	// Trim widget history to match conversation log
	if len(c.messageWidgets) > c.maxHistoryLength {
		// Remove oldest widgets
		removeCount := len(c.messageWidgets) - c.maxHistoryLength
		for i := 0; i < removeCount; i++ {
			c.conversationContainer.Remove(c.messageWidgets[i])
		}
		c.messageWidgets = c.messageWidgets[removeCount:]
	}

	c.conversationContainer.Refresh()
}

// updateConversationDisplay refreshes the conversation history display
func (c *ChatbotInterface) updateConversationDisplay() {
	if len(c.conversationLog) == 0 {
		// Show empty state
		emptyLabel := widget.NewLabel("Start a conversation...")
		emptyLabel.TextStyle.Italic = true
		c.conversationContainer.RemoveAll()
		c.conversationContainer.Add(emptyLabel)
		return
	}

	// Clear existing widgets
	c.conversationContainer.RemoveAll()
	c.messageWidgets = make([]*ChatMessageWidget, 0)

	// Create widgets for all messages
	for _, message := range c.conversationLog {
		messageWidget := NewChatMessageWidget(c.character, message, func(rating float64) {
			c.onMessageRated(message.Text, rating)
		})
		c.messageWidgets = append(c.messageWidgets, messageWidget)
		c.conversationContainer.Add(messageWidget)
	}

	c.conversationContainer.Refresh()
}

// onMessageRated handles when a user rates a message
func (c *ChatbotInterface) onMessageRated(messageText string, rating float64) {
	// Update the conversation log with the new rating
	for i := range c.conversationLog {
		if c.conversationLog[i].Text == messageText {
			c.conversationLog[i].IsFavorite = rating > 0
			c.conversationLog[i].Rating = rating
			break
		}
	}

	// The rating is automatically saved to character memory by the ChatMessageWidget
	// No additional action needed here
}

// scrollToBottom scrolls conversation history to show the latest messages
func (c *ChatbotInterface) scrollToBottom() {
	if c.historyScroll != nil {
		// Scroll to bottom with a small delay to ensure content is rendered
		go func() {
			time.Sleep(50 * time.Millisecond)
			c.historyScroll.ScrollToBottom()
		}()
	}
}

// Show displays the chatbot interface
func (c *ChatbotInterface) Show() {
	if !c.available {
		return
	}

	c.visible = true
	c.content.Show()

	// Focus the input field when showing
	if c.messageInput != nil {
		c.messageInput.FocusGained()
	}

	c.Refresh()
}

// Hide conceals the chatbot interface
func (c *ChatbotInterface) Hide() {
	c.visible = false
	if c.content != nil {
		c.content.Hide()
	}
	c.Refresh()
}

// Toggle switches the chatbot interface visibility
func (c *ChatbotInterface) Toggle() {
	if c.visible {
		c.Hide()
	} else {
		c.Show()
	}
}

// IsVisible returns whether the chatbot interface is currently visible
func (c *ChatbotInterface) IsVisible() bool {
	return c.visible
}

// IsAvailable returns whether the chatbot interface is available for this character
// Only characters with dialog backend enabled can use the chatbot interface
func (c *ChatbotInterface) IsAvailable() bool {
	return c.available
}

// GetToggleButton returns the button used to toggle chatbot visibility
// This can be added to context menus or other UI areas
func (c *ChatbotInterface) GetToggleButton() *widget.Button {
	return c.toggleButton
}

// ClearHistory clears the conversation history
func (c *ChatbotInterface) ClearHistory() {
	c.conversationLog = make([]ChatMessage, 0)
	c.updateConversationDisplay()
}

// GetConversationLength returns the number of messages in the conversation
func (c *ChatbotInterface) GetConversationLength() int {
	return len(c.conversationLog)
}

// SetPosition moves the chatbot interface to a specific location
func (c *ChatbotInterface) SetPosition(x, y float32) {
	if c.content != nil {
		c.content.Move(fyne.NewPos(x, y))
	}
}

// GetContainer returns the main container for adding to window layouts
func (c *ChatbotInterface) GetContainer() *fyne.Container {
	return c.content
}

// FocusInput sets focus to the message input field for better user experience
func (c *ChatbotInterface) FocusInput() {
	if c.messageInput != nil && c.visible {
		c.messageInput.FocusGained()
	}
}

// CreateRenderer creates the Fyne renderer for the chatbot interface
func (c *ChatbotInterface) CreateRenderer() fyne.WidgetRenderer {
	return &chatbotRenderer{
		chatbot: c,
		content: c.content,
	}
}

// chatbotRenderer implements fyne.WidgetRenderer for the chatbot interface
type chatbotRenderer struct {
	chatbot *ChatbotInterface
	content *fyne.Container
}

// Layout arranges the chatbot interface components
func (r *chatbotRenderer) Layout(size fyne.Size) {
	if r.chatbot.visible && r.content != nil {
		r.content.Resize(r.content.Size())
		r.content.Move(r.content.Position())
	}
}

// MinSize returns the minimum size for the chatbot interface
func (r *chatbotRenderer) MinSize() fyne.Size {
	if r.chatbot.visible {
		return fyne.NewSize(300, 200)
	}
	return fyne.NewSize(0, 0)
}

// Objects returns the canvas objects for rendering
func (r *chatbotRenderer) Objects() []fyne.CanvasObject {
	if r.chatbot.visible && r.content != nil {
		return []fyne.CanvasObject{r.content}
	}
	return []fyne.CanvasObject{}
}

// Refresh redraws the chatbot interface
func (r *chatbotRenderer) Refresh() {
	if r.chatbot.visible && r.content != nil {
		r.content.Refresh()
	}
}

// Destroy cleans up chatbot interface resources
func (r *chatbotRenderer) Destroy() {
	// No special cleanup needed for standard Fyne components
}

// ENHANCEMENT: Load recent conversations from character memory
func (ci *ChatbotInterface) loadRecentConversations() {
	// Get recent dialog interactions from character memory
	if ci.character == nil {
		return
	}

	// Load recent dialog memories from character
	memories := ci.character.GetRecentDialogMemories(10)

	// Convert dialog memories to chat messages
	for _, memory := range memories {
		if memory.Trigger == "chat" { // Only load chat interactions
			// Add character response to conversation log
			ci.conversationLog = append(ci.conversationLog, ChatMessage{
				Text:      memory.Response,
				IsUser:    false,
				Timestamp: memory.Timestamp,
			})
		}
	}
}

// ENHANCEMENT: Export conversation history to file
func (ci *ChatbotInterface) ExportConversation() error {
	if len(ci.conversationLog) == 0 {
		return fmt.Errorf("no conversation to export")
	}

	// Create filename with character name and timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	var name string
	if ci.character != nil {
		name = ci.character.GetName()
	} else {
		name = "Unknown"
	}
	filename := fmt.Sprintf("%s_chat_%s.txt", name, timestamp)

	// Build conversation text
	var conversation strings.Builder
	conversation.WriteString(fmt.Sprintf("Chat Conversation with %s\n", name))
	conversation.WriteString(fmt.Sprintf("Exported on: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	for _, msg := range ci.conversationLog {
		speaker := "Character"
		if msg.IsUser {
			speaker = "You"
		}
		conversation.WriteString(fmt.Sprintf("[%s] %s: %s\n",
			msg.Timestamp.Format("15:04:05"), speaker, msg.Text))
	}

	// Write to file in user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %v", err)
	}

	filepath := filepath.Join(homeDir, filename)
	err = os.WriteFile(filepath, []byte(conversation.String()), 0o644)
	if err != nil {
		return fmt.Errorf("failed to write conversation file: %v", err)
	}

	return nil
}
