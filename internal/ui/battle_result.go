// Battle result overlay UI component
// This file implements battle result display widgets
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

// BattleResult represents the outcome of a battle action
type BattleResult struct {
	Success       bool
	ActionType    BattleActionType
	Damage        float64
	Healing       float64
	StatusEffects []string
	Animation     string
	Response      string
}

// BattleResultOverlay displays the results of battle actions
// Design Philosophy:
// - Temporary overlay that auto-dismisses after showing result
// - Clear visual feedback for battle actions
// - Consistent with other UI components in the project
// - Uses color coding for different result types
type BattleResultOverlay struct {
	widget.BaseWidget
	background    *canvas.Rectangle
	content       *fyne.Container
	titleLabel    *widget.Label
	messageLabel  *widget.Label
	detailsLabel  *widget.Label
	visible       bool
	autoHideTimer *time.Timer
}

// NewBattleResultOverlay creates a new battle result overlay
func NewBattleResultOverlay() *BattleResultOverlay {
	overlay := &BattleResultOverlay{
		visible: false,
	}

	// Create background with semi-transparent styling
	overlay.background = canvas.NewRectangle(color.RGBA{R: 0, G: 0, B: 0, A: 180})
	overlay.background.StrokeColor = color.RGBA{R: 255, G: 255, B: 255, A: 200}
	overlay.background.StrokeWidth = 2

	// Create title label
	overlay.titleLabel = widget.NewLabel("Battle Result")
	overlay.titleLabel.TextStyle = fyne.TextStyle{Bold: true}
	overlay.titleLabel.Alignment = fyne.TextAlignCenter

	// Create message label for main result
	overlay.messageLabel = widget.NewLabel("")
	overlay.messageLabel.Alignment = fyne.TextAlignCenter
	overlay.messageLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Create details label for additional information
	overlay.detailsLabel = widget.NewLabel("")
	overlay.detailsLabel.Alignment = fyne.TextAlignCenter
	overlay.detailsLabel.Wrapping = fyne.TextWrapWord

	overlay.rebuildContent()
	overlay.ExtendBaseWidget(overlay)
	return overlay
}

// ShowResult displays a battle result with auto-hide after 3 seconds
func (o *BattleResultOverlay) ShowResult(result BattleResult) {
	o.setResultContent(result)
	o.Show()

	// Auto-hide after 3 seconds
	if o.autoHideTimer != nil {
		o.autoHideTimer.Stop()
	}
	o.autoHideTimer = time.AfterFunc(3*time.Second, func() {
		o.Hide()
	})
}

// ShowMessage displays a simple message with auto-hide
func (o *BattleResultOverlay) ShowMessage(title, message string) {
	o.titleLabel.SetText(title)
	o.messageLabel.SetText(message)
	o.detailsLabel.SetText("")

	// Set neutral styling
	o.background.FillColor = color.RGBA{R: 100, G: 100, B: 100, A: 180}

	o.Show()

	// Auto-hide after 2 seconds for simple messages
	if o.autoHideTimer != nil {
		o.autoHideTimer.Stop()
	}
	o.autoHideTimer = time.AfterFunc(2*time.Second, func() {
		o.Hide()
	})
}

// setResultContent configures the overlay content based on battle result
func (o *BattleResultOverlay) setResultContent(result BattleResult) {
	// Set title based on success
	if result.Success {
		o.titleLabel.SetText("Action Successful!")
		o.background.FillColor = color.RGBA{R: 0, G: 150, B: 0, A: 180} // Green for success
	} else {
		o.titleLabel.SetText("Action Failed!")
		o.background.FillColor = color.RGBA{R: 150, G: 0, B: 0, A: 180} // Red for failure
	}

	// Set main message
	message := o.formatActionMessage(result)
	o.messageLabel.SetText(message)

	// Set details
	details := o.formatActionDetails(result)
	o.detailsLabel.SetText(details)
}

// formatActionMessage creates a descriptive message for the action result
func (o *BattleResultOverlay) formatActionMessage(result BattleResult) string {
	if !result.Success {
		return fmt.Sprintf("%s failed!", string(result.ActionType))
	}

	switch result.ActionType {
	case ActionAttack, ActionDrain, ActionCounter:
		if result.Damage > 0 {
			return fmt.Sprintf("Dealt %.0f damage!", result.Damage)
		}
		return "Attack connected!"
	case ActionHeal:
		if result.Healing > 0 {
			return fmt.Sprintf("Restored %.0f HP!", result.Healing)
		}
		return "Healing applied!"
	case ActionDefend:
		return "Defense stance activated!"
	case ActionStun:
		return "Opponent stunned!"
	case ActionBoost:
		return "Attack power increased!"
	case ActionShield:
		return "Protective barrier created!"
	case ActionCharge:
		return "Energy building up!"
	case ActionEvade:
		return "Evasion ready!"
	case ActionTaunt:
		return "Opponent provoked!"
	default:
		return fmt.Sprintf("%s successful!", string(result.ActionType))
	}
}

// formatActionDetails creates additional details text for the action result
func (o *BattleResultOverlay) formatActionDetails(result BattleResult) string {
	var details []string

	// Add damage and healing info
	if result.Damage > 0 && result.Healing > 0 {
		details = append(details, fmt.Sprintf("Damage: %.0f • Healing: %.0f", result.Damage, result.Healing))
	} else if result.Damage > 0 {
		details = append(details, fmt.Sprintf("Damage: %.0f", result.Damage))
	} else if result.Healing > 0 {
		details = append(details, fmt.Sprintf("Healing: %.0f", result.Healing))
	}

	// Add status effects
	if len(result.StatusEffects) > 0 {
		effects := "Effects: "
		for i, effect := range result.StatusEffects {
			if i > 0 {
				effects += ", "
			}
			effects += effect
		}
		details = append(details, effects)
	}

	// Add response if available
	if result.Response != "" {
		details = append(details, fmt.Sprintf("Response: %s", result.Response))
	}

	// Join all details
	result_details := ""
	for i, detail := range details {
		if i > 0 {
			result_details += "\n"
		}
		result_details += detail
	}

	return result_details
}

// rebuildContent recreates the content container
func (o *BattleResultOverlay) rebuildContent() {
	// Create content container
	mainContent := container.NewVBox(
		o.titleLabel,
		o.messageLabel,
		o.detailsLabel,
	)

	// Create container with background and content
	o.content = container.NewBorder(nil, nil, nil, nil, o.background, mainContent)

	o.updateSize()
}

// updateSize calculates appropriate overlay size and position
func (o *BattleResultOverlay) updateSize() {
	// Center overlay on screen
	width := float32(300)
	height := float32(150)

	overlayX := float32(-150) // Half width for centering
	overlayY := float32(-75)  // Half height for centering

	// Apply layout
	o.content.Resize(fyne.NewSize(width, height))
	o.content.Move(fyne.NewPos(overlayX, overlayY))
	o.background.Resize(fyne.NewSize(width, height))
}

// Show displays the battle result overlay
func (o *BattleResultOverlay) Show() {
	o.visible = true
	o.content.Show()
	o.Refresh()
}

// Hide hides the battle result overlay
func (o *BattleResultOverlay) Hide() {
	o.visible = false
	o.content.Hide()

	// Cancel auto-hide timer if running
	if o.autoHideTimer != nil {
		o.autoHideTimer.Stop()
		o.autoHideTimer = nil
	}

	o.Refresh()
}

// IsVisible returns whether the overlay is currently visible
func (o *BattleResultOverlay) IsVisible() bool {
	return o.visible
}

// CreateRenderer creates the Fyne renderer for the battle result overlay
func (o *BattleResultOverlay) CreateRenderer() fyne.WidgetRenderer {
	return &battleResultOverlayRenderer{
		overlay: o,
		content: o.content,
	}
}

// battleResultOverlayRenderer implements fyne.WidgetRenderer for battle result overlays
type battleResultOverlayRenderer struct {
	overlay *BattleResultOverlay
	content *fyne.Container
}

// Layout arranges the battle result overlay components
func (r *battleResultOverlayRenderer) Layout(size fyne.Size) {
	if r.overlay.visible && r.content != nil {
		r.content.Resize(r.content.Size())
		r.content.Move(r.content.Position())
	}
}

// MinSize returns the minimum size for the battle result overlay
func (r *battleResultOverlayRenderer) MinSize() fyne.Size {
	if r.overlay.visible {
		return fyne.NewSize(300, 150)
	}
	return fyne.NewSize(0, 0)
}

// Objects returns the canvas objects for rendering
func (r *battleResultOverlayRenderer) Objects() []fyne.CanvasObject {
	if r.overlay.visible && r.content != nil {
		return []fyne.CanvasObject{r.content}
	}
	return []fyne.CanvasObject{}
}

// Refresh redraws the battle result overlay
func (r *battleResultOverlayRenderer) Refresh() {
	if r.overlay.visible && r.content != nil {
		r.content.Refresh()
	}
}

// Destroy cleans up battle result overlay resources
func (r *battleResultOverlayRenderer) Destroy() {
	if r.overlay != nil && r.overlay.autoHideTimer != nil {
		r.overlay.autoHideTimer.Stop()
	}
}
