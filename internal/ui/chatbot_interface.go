// Package ui provides desktop companion user interface components
// This file implements a chatbot interface widget for AI-enabled characters
package ui

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"desktop-companion/internal/character"
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
//   chatbot := NewChatbotInterface(character)
//   if chatbot.IsAvailable() {
//       chatbot.Show()
//   }
//   // User types message and clicks send or presses Enter
//   // Character responds using existing dialog backend system
type ChatbotInterface struct {
	widget.BaseWidget
	character          *character.Character
	background         *canvas.Rectangle
	content            *fyne.Container
	visible            bool
	available          bool

	// UI components
	conversationHistory *widget.RichText
	messageInput        *widget.Entry
	sendButton          *widget.Button
	toggleButton        *widget.Button
	historyScroll       *container.Scroll

	// State management
	conversationLog     []ChatMessage
	maxHistoryLength    int
	lastMessageTime     time.Time
	inputPlaceholder    string
}

// ChatMessage represents a single message in the conversation
type ChatMessage struct {
	IsUser    bool      `json:"isUser"`    // true for user messages, false for character responses
	Text      string    `json:"text"`      // Message content
	Timestamp time.Time `json:"timestamp"` // When the message was sent/received
	Animation string    `json:"animation"` // Animation triggered with character response (if any)
}

// NewChatbotInterface creates a new chatbot interface widget.
// The interface is only available for characters with dialog backend enabled.
// 
// Returns a fully initialized ChatbotInterface widget that follows Fyne's
// widget pattern and can be added to any container layout.
func NewChatbotInterface(char *character.Character) *ChatbotInterface {
	chatbot := &ChatbotInterface{
		character:           char,
		visible:             false,
		available:           char.GetCard().HasDialogBackend(),
		conversationLog:     make([]ChatMessage, 0),
		maxHistoryLength:    50, // Limit to prevent memory issues
		inputPlaceholder:    "Type a message...",
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

	// Create conversation history display
	c.conversationHistory = widget.NewRichText()
	c.conversationHistory.Wrapping = fyne.TextWrapWord
	c.updateConversationDisplay()

	// Create scrollable container for conversation history
	c.historyScroll = container.NewScroll(c.conversationHistory)
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
		// Add character response to conversation
		characterMessage := ChatMessage{
			IsUser:    false,
			Text:      response,
			Timestamp: time.Now(),
			Animation: c.character.GetCurrentState(), // Capture animation used
		}
		c.addMessage(characterMessage)
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
}

// updateConversationDisplay refreshes the conversation history display
func (c *ChatbotInterface) updateConversationDisplay() {
	if len(c.conversationLog) == 0 {
		c.conversationHistory.ParseMarkdown("*Start a conversation...*")
		return
	}

	// Build conversation text with styling
	var content strings.Builder
	
	for i, message := range c.conversationLog {
		// Add timestamp for readability
		timeStr := message.Timestamp.Format("15:04")
		
		if message.IsUser {
			// User messages in blue
			content.WriteString(fmt.Sprintf("**You** (%s): %s\n\n", timeStr, message.Text))
		} else {
			// Character messages in default color
			characterName := c.character.GetName()
			content.WriteString(fmt.Sprintf("**%s** (%s): %s\n\n", characterName, timeStr, message.Text))
		}

		// Add separator between conversation sessions
		if i < len(c.conversationLog)-1 {
			nextMessage := c.conversationLog[i+1]
			if nextMessage.Timestamp.Sub(message.Timestamp) > 5*time.Minute {
				content.WriteString("---\n\n")
			}
		}
	}

	c.conversationHistory.ParseMarkdown(content.String())
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
