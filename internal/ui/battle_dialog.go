// Package ui provides battle-related user interface components
// This file implements battle UI widgets for the JRPG battle system
package ui

import (
	"fmt"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// BattleActionType represents the type of battle action
// Using string type for JSON serialization compatibility
type BattleActionType string

// Battle action constants matching the battle system design
const (
	ActionAttack  BattleActionType = "attack"
	ActionDefend  BattleActionType = "defend"
	ActionStun    BattleActionType = "stun"
	ActionHeal    BattleActionType = "heal"
	ActionBoost   BattleActionType = "boost"
	ActionCounter BattleActionType = "counter"
	ActionDrain   BattleActionType = "drain"
	ActionShield  BattleActionType = "shield"
	ActionCharge  BattleActionType = "charge"
	ActionEvade   BattleActionType = "evade"
	ActionTaunt   BattleActionType = "taunt"
)

// BattleActionDialog displays a dialog for selecting battle actions
// Design Philosophy:
// - Uses standard Fyne components for consistency
// - Provides clear action selection with descriptions
// - Includes timeout handling for turn-based gameplay
// - Follows the same widget architecture as DialogBubble and ContextMenu
type BattleActionDialog struct {
	widget.BaseWidget
	background     *canvas.Rectangle
	content        *fyne.Container
	titleLabel     *widget.Label
	actionButtons  []*widget.Button
	cancelButton   *widget.Button
	timerLabel     *widget.Label
	visible        bool
	onActionSelect func(BattleActionType)
	onCancel       func()
	turnTimeout    time.Duration
	timeRemaining  time.Duration
	timerRunning   bool
	timerStop      chan bool
}

// NewBattleActionDialog creates a new battle action selection dialog
// turnTimeout specifies how long the player has to select an action
// If timeout is 0, no timer is displayed
func NewBattleActionDialog(turnTimeout time.Duration) *BattleActionDialog {
	dialog := &BattleActionDialog{
		turnTimeout:   turnTimeout,
		timeRemaining: turnTimeout,
		timerStop:     make(chan bool, 1),
		visible:       false,
	}

	// Create background with battle theme styling
	dialog.background = canvas.NewRectangle(color.RGBA{R: 240, G: 240, B: 255, A: 250})
	dialog.background.StrokeColor = color.RGBA{R: 100, G: 100, B: 150, A: 255}
	dialog.background.StrokeWidth = 2

	// Create title label
	dialog.titleLabel = widget.NewLabel("Select Battle Action")
	dialog.titleLabel.TextStyle = fyne.TextStyle{Bold: true}
	dialog.titleLabel.Alignment = fyne.TextAlignCenter

	// Create timer label if timeout is specified
	if turnTimeout > 0 {
		dialog.timerLabel = widget.NewLabel(fmt.Sprintf("Time: %.0fs", turnTimeout.Seconds()))
		dialog.timerLabel.Alignment = fyne.TextAlignCenter
	}

	// Create cancel button
	dialog.cancelButton = widget.NewButton("Cancel", func() {
		dialog.Hide()
		if dialog.onCancel != nil {
			dialog.onCancel()
		}
	})
	dialog.cancelButton.Importance = widget.LowImportance

	dialog.createActionButtons()
	dialog.rebuildContent()

	dialog.ExtendBaseWidget(dialog)
	return dialog
}

// createActionButtons creates buttons for all available battle actions
// Each button includes the action name and a brief description
func (d *BattleActionDialog) createActionButtons() {
	actions := []struct {
		action      BattleActionType
		label       string
		description string
	}{
		{ActionAttack, "Attack", "Deal damage to opponent"},
		{ActionDefend, "Defend", "Reduce incoming damage"},
		{ActionHeal, "Heal", "Restore hit points"},
		{ActionStun, "Stun", "Disable opponent temporarily"},
		{ActionBoost, "Boost", "Increase attack power"},
		{ActionCounter, "Counter", "Reactive counter-attack"},
		{ActionDrain, "Drain", "Absorb opponent's energy"},
		{ActionShield, "Shield", "Create protective barrier"},
		{ActionCharge, "Charge", "Build energy for next attack"},
		{ActionEvade, "Evade", "Avoid next attack"},
		{ActionTaunt, "Taunt", "Provoke opponent"},
	}

	d.actionButtons = make([]*widget.Button, len(actions))
	for i, action := range actions {
		// Capture action in closure to avoid loop variable issues
		selectedAction := action.action

		// Create button with action name and description
		buttonText := fmt.Sprintf("%s\n%s", action.label, action.description)
		btn := widget.NewButton(buttonText, func() {
			d.Hide()
			if d.onActionSelect != nil {
				d.onActionSelect(selectedAction)
			}
		})

		btn.Importance = widget.MediumImportance
		d.actionButtons[i] = btn
	}
}

// rebuildContent recreates the content container with current elements
func (d *BattleActionDialog) rebuildContent() {
	// Create action buttons grid (3 columns for better layout)
	actionGrid := container.NewGridWithColumns(3)
	for _, btn := range d.actionButtons {
		actionGrid.Add(btn)
	}

	// Create content sections
	var contentObjects []fyne.CanvasObject
	contentObjects = append(contentObjects, d.titleLabel)

	if d.timerLabel != nil {
		contentObjects = append(contentObjects, d.timerLabel)
	}

	contentObjects = append(contentObjects, actionGrid)
	contentObjects = append(contentObjects, d.cancelButton)

	// Create main content container
	mainContent := container.NewVBox(contentObjects...)

	// Create container with background and content
	d.content = container.NewBorder(nil, nil, nil, nil, d.background, mainContent)

	d.updateSize()
}

// updateSize calculates appropriate dialog size
func (d *BattleActionDialog) updateSize() {
	// Calculate size based on content
	width := float32(480)  // Wide enough for 3-column layout
	height := float32(320) // Height for title, timer, actions, and cancel

	if d.timerLabel == nil {
		height -= 30 // Reduce height if no timer
	}

	// Center the dialog
	dialogX := float32(-240) // Half width for centering
	dialogY := float32(-160) // Half height for centering

	// Apply layout
	d.content.Resize(fyne.NewSize(width, height))
	d.content.Move(fyne.NewPos(dialogX, dialogY))
	d.background.Resize(fyne.NewSize(width, height))
}

// SetOnActionSelect sets the callback for action selection
func (d *BattleActionDialog) SetOnActionSelect(callback func(BattleActionType)) {
	d.onActionSelect = callback
}

// SetOnCancel sets the callback for dialog cancellation
func (d *BattleActionDialog) SetOnCancel(callback func()) {
	d.onCancel = callback
}

// Show displays the battle action dialog and starts the timer if configured
func (d *BattleActionDialog) Show() {
	d.visible = true
	d.content.Show()

	// Start timer if configured
	if d.turnTimeout > 0 {
		d.startTimer()
	}

	d.Refresh()
}

// Hide hides the battle action dialog and stops the timer
func (d *BattleActionDialog) Hide() {
	d.visible = false
	d.content.Hide()
	d.stopTimer()
	d.Refresh()
}

// startTimer begins the turn countdown timer
func (d *BattleActionDialog) startTimer() {
	if d.timerRunning {
		return
	}

	d.timerRunning = true
	d.timeRemaining = d.turnTimeout

	go func() {
		ticker := time.NewTicker(100 * time.Millisecond) // Update every 100ms for smooth countdown
		defer ticker.Stop()

		for {
			select {
			case <-d.timerStop:
				return
			case <-ticker.C:
				d.timeRemaining -= 100 * time.Millisecond

				if d.timeRemaining <= 0 {
					// Timer expired - trigger timeout action
					d.timerRunning = false
					d.Hide()
					if d.onCancel != nil {
						d.onCancel() // Treat timeout as cancellation
					}
					return
				}

				// Update timer display
				if d.timerLabel != nil {
					d.timerLabel.SetText(fmt.Sprintf("Time: %.1fs", d.timeRemaining.Seconds()))
				}
			}
		}
	}()
}

// stopTimer stops the turn countdown timer
func (d *BattleActionDialog) stopTimer() {
	if d.timerRunning {
		d.timerRunning = false
		select {
		case d.timerStop <- true:
		default:
		}
	}
}

// IsVisible returns whether the dialog is currently visible
func (d *BattleActionDialog) IsVisible() bool {
	return d.visible
}

// CreateRenderer creates the Fyne renderer for the battle action dialog
func (d *BattleActionDialog) CreateRenderer() fyne.WidgetRenderer {
	return &battleActionDialogRenderer{
		dialog:  d,
		content: d.content,
	}
}

// battleActionDialogRenderer implements fyne.WidgetRenderer for battle action dialogs
type battleActionDialogRenderer struct {
	dialog  *BattleActionDialog
	content *fyne.Container
}

// Layout arranges the battle action dialog components
func (r *battleActionDialogRenderer) Layout(size fyne.Size) {
	if r.dialog.visible && r.content != nil {
		r.content.Resize(r.content.Size())
		r.content.Move(r.content.Position())
	}
}

// MinSize returns the minimum size for the battle action dialog
func (r *battleActionDialogRenderer) MinSize() fyne.Size {
	if r.dialog.visible {
		return fyne.NewSize(480, 320)
	}
	return fyne.NewSize(0, 0)
}

// Objects returns the canvas objects for rendering
func (r *battleActionDialogRenderer) Objects() []fyne.CanvasObject {
	if r.dialog.visible && r.content != nil {
		return []fyne.CanvasObject{r.content}
	}
	return []fyne.CanvasObject{}
}

// Refresh redraws the battle action dialog
func (r *battleActionDialogRenderer) Refresh() {
	if r.dialog.visible && r.content != nil {
		r.content.Refresh()
	}
}

// Destroy cleans up battle action dialog resources
func (r *battleActionDialogRenderer) Destroy() {
	if r.dialog != nil {
		r.dialog.stopTimer()
	}
}
