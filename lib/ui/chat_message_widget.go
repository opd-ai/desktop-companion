// Package ui provides desktop companion user interface components
// This file implements a chat message widget with favorite star rating functionality
package ui

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/opd-ai/desktop-companion/lib/character"
)

// ChatMessageWidget represents a single chat message with rating functionality
//
// Design Philosophy:
// - Simple star rating system (1-5 stars) for character responses only
// - Uses existing Fyne widgets for library-first approach
// - Integrates with character memory system for favorite tracking
// - Visual distinction between user and character messages
type ChatMessageWidget struct {
	widget.BaseWidget
	character *character.Character
	message   ChatMessage
	content   *fyne.Container

	// UI components
	messageText  *widget.RichText
	starButtons  []*widget.Button
	favoriteIcon *widget.Icon
	ratingLabel  *widget.Label
	messageBox   *fyne.Container

	// State
	currentRating   float64
	onRatingChanged func(rating float64)
}

// NewChatMessageWidget creates a chat message widget with optional rating functionality
func NewChatMessageWidget(char *character.Character, msg ChatMessage, onRatingChanged func(float64)) *ChatMessageWidget {
	widget := &ChatMessageWidget{
		character:       char,
		message:         msg,
		currentRating:   msg.Rating,
		onRatingChanged: onRatingChanged,
	}

	widget.ExtendBaseWidget(widget)
	widget.setupComponents()
	widget.setupLayout()

	return widget
}

// setupComponents initializes the UI components
func (c *ChatMessageWidget) setupComponents() {
	// Create message text display
	c.messageText = widget.NewRichText()
	c.messageText.Wrapping = fyne.TextWrapWord

	// Format message with timestamp and styling
	timeStr := c.message.Timestamp.Format("15:04")
	if c.message.IsUser {
		// User messages in blue
		c.messageText.ParseMarkdown(fmt.Sprintf("**You** (%s): %s", timeStr, c.message.Text))
	} else {
		// Character messages with rating controls
		characterName := c.character.GetName()
		c.messageText.ParseMarkdown(fmt.Sprintf("**%s** (%s): %s", characterName, timeStr, c.message.Text))

		// Create star rating buttons for character responses only
		c.setupStarRating()
	}
}

// setupStarRating creates the star rating interface for character messages
func (c *ChatMessageWidget) setupStarRating() {
	if c.message.IsUser {
		return // No rating for user messages
	}

	// Create star rating buttons (1-5 stars) using available icons
	c.starButtons = make([]*widget.Button, 5)
	for i := 0; i < 5; i++ {
		starIndex := i + 1 // 1-based rating
		c.starButtons[i] = widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
			c.setRating(float64(starIndex))
		})
		c.starButtons[i].Resize(fyne.NewSize(20, 20))
		c.starButtons[i].Importance = widget.LowImportance
	}

	// Create favorite icon indicator using available confirmation icon
	c.favoriteIcon = widget.NewIcon(theme.ConfirmIcon())
	c.favoriteIcon.Resize(fyne.NewSize(16, 16))

	// Create rating label
	c.ratingLabel = widget.NewLabel("")
	c.ratingLabel.TextStyle.Italic = true

	// Update visual state based on current rating
	c.updateStarVisuals()
}

// setupLayout arranges the components
func (c *ChatMessageWidget) setupLayout() {
	// Create main message container
	c.messageBox = container.NewVBox(c.messageText)

	// Add rating controls for character messages
	if !c.message.IsUser && c.starButtons != nil {
		// Create star rating row
		starContainer := container.NewHBox()
		for _, star := range c.starButtons {
			starContainer.Add(star)
		}

		// Add favorite indicator and rating label
		statusContainer := container.NewHBox(c.favoriteIcon, c.ratingLabel)

		// Add rating controls to message
		ratingRow := container.NewBorder(nil, nil, starContainer, statusContainer)
		c.messageBox.Add(ratingRow)
	}

	// Create background with message bubble styling
	background := canvas.NewRectangle(color.RGBA{R: 245, G: 245, B: 245, A: 200})
	background.CornerRadius = 8

	if c.message.IsUser {
		// User message styling (light blue)
		background.FillColor = color.RGBA{R: 200, G: 230, B: 255, A: 200}
	} else {
		// Character message styling (light gray)
		background.FillColor = color.RGBA{R: 245, G: 245, B: 245, A: 200}
	}

	// Combine background and content
	c.content = container.NewBorder(nil, nil, nil, nil, background, c.messageBox)
	c.content.Resize(fyne.NewSize(300, 60)) // Default size
}

// setRating sets the star rating for this message
func (c *ChatMessageWidget) setRating(rating float64) {
	if c.message.IsUser {
		return // Cannot rate user messages
	}

	c.currentRating = rating
	c.message.Rating = rating
	c.message.IsFavorite = rating > 0

	// Update character memory with favorite status
	if c.character != nil && c.character.GetGameState() != nil {
		if rating > 0 {
			c.character.GetGameState().MarkDialogResponseFavorite(c.message.Text, rating)
		} else {
			c.character.GetGameState().UnmarkDialogResponseFavorite(c.message.Text)
		}
	}

	// Update visual state
	c.updateStarVisuals()

	// Notify parent of rating change
	if c.onRatingChanged != nil {
		c.onRatingChanged(rating)
	}
}

// updateStarVisuals updates the appearance of star buttons and indicators
func (c *ChatMessageWidget) updateStarVisuals() {
	if c.starButtons == nil {
		return
	}

	// Update star button appearance using available icons
	for i, button := range c.starButtons {
		if float64(i+1) <= c.currentRating {
			// Filled star (rated) - use confirm icon for filled state
			button.SetIcon(theme.ConfirmIcon())
			button.Importance = widget.HighImportance
		} else {
			// Empty star (not rated) - use add icon for empty state
			button.SetIcon(theme.ContentAddIcon())
			button.Importance = widget.LowImportance
		}
		button.Refresh()
	}

	// Update favorite icon visibility
	if c.favoriteIcon != nil {
		if c.currentRating > 0 {
			c.favoriteIcon.Show()
		} else {
			c.favoriteIcon.Hide()
		}
	}

	// Update rating label
	if c.ratingLabel != nil {
		if c.currentRating > 0 {
			c.ratingLabel.SetText(fmt.Sprintf("%.0f/5 stars", c.currentRating))
			c.ratingLabel.Show()
		} else {
			c.ratingLabel.SetText("")
			c.ratingLabel.Hide()
		}
	}
}

// GetRating returns the current star rating
func (c *ChatMessageWidget) GetRating() float64 {
	return c.currentRating
}

// IsFavorite returns whether this message is marked as favorite
func (c *ChatMessageWidget) IsFavorite() bool {
	return c.currentRating > 0
}

// CreateRenderer implements fyne.Widget interface
func (c *ChatMessageWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(c.content)
}
