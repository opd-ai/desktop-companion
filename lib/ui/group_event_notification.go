package ui

import (
	"fmt"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/opd-ai/desktop-companion/internal/network"
)

// GroupEventNotification displays floating group event invitation notifications
// Follows the achievement notification pattern using existing Fyne widgets
type GroupEventNotification struct {
	widget.BaseWidget
	container        *fyne.Container
	background       *canvas.Rectangle
	titleLabel       *widget.RichText
	descLabel        *widget.RichText
	acceptButton     *widget.Button
	declineButton    *widget.Button
	visible          bool
	timeoutTimer     *time.Timer
	hideCallback     func()
	responseCallback func(accepted bool)
}

// NewGroupEventNotification creates a new group event notification widget
func NewGroupEventNotification() *GroupEventNotification {
	gen := &GroupEventNotification{
		visible: false,
	}

	// Create blue background for network event feel
	gen.background = canvas.NewRectangle(color.RGBA{R: 70, G: 130, B: 180, A: 200}) // Steel blue with transparency
	gen.background.StrokeColor = color.RGBA{R: 25, G: 25, B: 112, A: 255}           // Navy blue border
	gen.background.StrokeWidth = 2

	// Create invitation title with network icon
	gen.titleLabel = widget.NewRichTextFromMarkdown("**🌐 Group Event Invitation**")
	gen.titleLabel.Wrapping = fyne.TextWrapWord

	// Create description label
	gen.descLabel = widget.NewRichTextFromMarkdown("*Event details*")
	gen.descLabel.Wrapping = fyne.TextWrapWord

	// Create action buttons
	gen.acceptButton = widget.NewButton("Accept", func() {
		if gen.responseCallback != nil {
			gen.responseCallback(true)
		}
		gen.Hide()
	})

	gen.declineButton = widget.NewButton("Decline", func() {
		if gen.responseCallback != nil {
			gen.responseCallback(false)
		}
		gen.Hide()
	})

	// Style buttons
	gen.acceptButton.Importance = widget.HighImportance
	gen.declineButton.Importance = widget.MediumImportance

	// Create button container
	buttonContainer := container.NewHBox(
		gen.acceptButton,
		gen.declineButton,
	)

	// Create main content container with padding
	content := container.NewVBox(
		gen.titleLabel,
		gen.descLabel,
		buttonContainer,
	)

	// Create container with background
	gen.container = container.NewWithoutLayout(
		gen.background,
		content,
	)

	// Set initial size and position (top-right corner)
	gen.container.Resize(fyne.NewSize(300, 120))
	gen.container.Move(fyne.NewPos(20, 20)) // Will be repositioned by parent
	gen.container.Hide()

	return gen
}

// ShowInvitation displays a group event invitation notification
func (gen *GroupEventNotification) ShowInvitation(invitation network.GroupEventInvitation, onResponse func(accepted bool)) {
	if gen.visible {
		gen.Hide() // Hide any existing notification
	}

	// Update content
	gen.titleLabel.ParseMarkdown("**🌐 Group Event Invitation**")

	description := fmt.Sprintf("**From:** %s\n**Event:** %s\n**Details:** %s",
		invitation.SenderID,
		invitation.TemplateName,
		invitation.Message)
	gen.descLabel.ParseMarkdown(description)

	// Store response callback
	gen.responseCallback = onResponse

	// Show notification
	gen.visible = true
	gen.container.Show()
	gen.Refresh()

	// Set auto-decline timeout (30 seconds)
	if gen.timeoutTimer != nil {
		gen.timeoutTimer.Stop()
	}
	gen.timeoutTimer = time.AfterFunc(30*time.Second, func() {
		if gen.visible && gen.responseCallback != nil {
			gen.responseCallback(false) // Auto-decline on timeout
		}
		gen.Hide()
	})
}

// Hide hides the group event notification
func (gen *GroupEventNotification) Hide() {
	if !gen.visible {
		return
	}

	gen.visible = false
	gen.container.Hide()

	// Stop timeout timer
	if gen.timeoutTimer != nil {
		gen.timeoutTimer.Stop()
		gen.timeoutTimer = nil
	}

	// Clear callback
	gen.responseCallback = nil

	if gen.hideCallback != nil {
		gen.hideCallback()
	}

	gen.Refresh()
}

// IsVisible returns whether the notification is currently visible
func (gen *GroupEventNotification) IsVisible() bool {
	return gen.visible
}

// SetHideCallback sets a callback to be called when the notification is hidden
func (gen *GroupEventNotification) SetHideCallback(callback func()) {
	gen.hideCallback = callback
}

// CreateRenderer creates the Fyne renderer for the group event notification
func (gen *GroupEventNotification) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(gen.container)
}

// Move positions the notification at the specified coordinates
func (gen *GroupEventNotification) Move(pos fyne.Position) {
	gen.container.Move(pos)
}

// Resize changes the notification size
func (gen *GroupEventNotification) Resize(size fyne.Size) {
	gen.container.Resize(size)
	gen.background.Resize(size)
}
