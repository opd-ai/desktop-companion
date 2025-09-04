package ui

import (
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// BattleInvitationDialog provides a simple confirmation dialog for battle invitations
// Follows existing UI patterns for consistency and minimal code changes
type BattleInvitationDialog struct {
	widget.BaseWidget
	content       *fyne.Container
	titleLabel    *widget.Label
	messageLabel  *widget.Label
	acceptButton  *widget.Button
	declineButton *widget.Button
	visible       bool
	onResponse    func(accepted bool)
	mu            sync.Mutex
}

// NewBattleInvitationDialog creates a new battle invitation confirmation dialog
func NewBattleInvitationDialog() *BattleInvitationDialog {
	dialog := &BattleInvitationDialog{
		visible: false,
	}

	dialog.initializeComponents()
	dialog.setupLayout()

	return dialog
}

// initializeComponents creates the dialog UI components
func (bid *BattleInvitationDialog) initializeComponents() {
	bid.titleLabel = widget.NewLabel("Battle Invitation")
	bid.messageLabel = widget.NewLabel("You have received a battle invitation. Do you accept?")

	bid.acceptButton = widget.NewButton("Accept", func() {
		bid.respond(true)
	})

	bid.declineButton = widget.NewButton("Decline", func() {
		bid.respond(false)
	})
}

// setupLayout creates the dialog layout
func (bid *BattleInvitationDialog) setupLayout() {
	buttonContainer := container.NewHBox(
		bid.acceptButton,
		bid.declineButton,
	)

	bid.content = container.NewVBox(
		bid.titleLabel,
		bid.messageLabel,
		buttonContainer,
	)

	bid.content.Hide() // Start hidden
}

// Show displays the battle invitation dialog
func (bid *BattleInvitationDialog) Show(fromCharacter string, onResponse func(accepted bool)) {
	bid.mu.Lock()
	defer bid.mu.Unlock()

	bid.onResponse = onResponse
	bid.messageLabel.SetText("Battle invitation from " + fromCharacter + ". Do you accept?")
	bid.visible = true
	bid.content.Show()
	bid.Refresh()
}

// Hide hides the battle invitation dialog
func (bid *BattleInvitationDialog) Hide() {
	bid.mu.Lock()
	defer bid.mu.Unlock()

	bid.visible = false
	bid.content.Hide()
	bid.Refresh()
}

// respond handles button click responses
func (bid *BattleInvitationDialog) respond(accepted bool) {
	bid.mu.Lock()
	callback := bid.onResponse
	bid.mu.Unlock()

	bid.Hide()

	if callback != nil {
		callback(accepted)
	}
}

// IsVisible returns whether the dialog is currently visible
func (bid *BattleInvitationDialog) IsVisible() bool {
	bid.mu.Lock()
	defer bid.mu.Unlock()
	return bid.visible
}

// GetContainer returns the container for embedding in the window
func (bid *BattleInvitationDialog) GetContainer() *fyne.Container {
	return bid.content
}

// CreateRenderer creates the Fyne renderer for the battle invitation dialog
func (bid *BattleInvitationDialog) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(bid.content)
}
