package ui

import (
	"github.com/opd-ai/desktop-companion/internal/character"
	"fmt"
	"image/color"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// GiftSelectionDialog provides a UI for selecting and giving gifts to characters
// Follows existing UI patterns from DialogBubble and ContextMenu for consistency
//
// Design Philosophy:
// - Reuses existing widget patterns (container layouts, button styling)
// - Uses standard library and Fyne components only
// - Maintains thread safety with proper error handling
// - Integrates seamlessly with existing interaction system
//
// Usage:
//
//	dialog := NewGiftSelectionDialog(giftManager)
//	dialog.SetOnGiftGiven(func(response *character.GiftResponse) {
//	    // Handle gift giving result
//	})
//	dialog.Show()
type GiftSelectionDialog struct {
	widget.BaseWidget
	giftManager    *character.GiftManager
	background     *canvas.Rectangle
	content        *fyne.Container
	giftList       *widget.List
	notesEntry     *widget.Entry
	giveButton     *widget.Button
	cancelButton   *widget.Button
	visible        bool
	selectedGift   *character.GiftDefinition
	onGiftGiven    func(*character.GiftResponse)
	onCancel       func()
	cooldownTimers map[string]*CooldownTimer // Track cooldown timers by gift ID
}

// NewGiftSelectionDialog creates a new gift selection dialog
// Returns a fully initialized dialog that integrates with the existing gift system
func NewGiftSelectionDialog(giftManager *character.GiftManager) *GiftSelectionDialog {
	dialog := &GiftSelectionDialog{
		giftManager:    giftManager,
		visible:        false,
		cooldownTimers: make(map[string]*CooldownTimer),
	}

	// Create background with dialog styling similar to DialogBubble
	dialog.background = canvas.NewRectangle(color.RGBA{R: 250, G: 250, B: 250, A: 240})
	dialog.background.StrokeColor = color.RGBA{R: 120, G: 120, B: 120, A: 255}
	dialog.background.StrokeWidth = 2

	// Initialize UI components
	dialog.createGiftList()
	dialog.createNotesEntry()
	dialog.createButtons()
	dialog.createLayout()

	dialog.ExtendBaseWidget(dialog)
	return dialog
}

// createGiftList initializes the gift selection list widget
// Uses Fyne's List widget for efficient rendering of available gifts
func (gsd *GiftSelectionDialog) createGiftList() {
	// Create list widget with dynamic data binding
	gsd.giftList = widget.NewList(
		func() int {
			// Return number of available gifts
			return len(gsd.getAvailableGifts())
		},
		func() fyne.CanvasObject {
			// Create template list item with icon and text
			icon := widget.NewIcon(nil)
			icon.SetResource(nil) // Will be set per item

			nameLabel := widget.NewLabel("Gift Name")
			nameLabel.TextStyle = fyne.TextStyle{Bold: true}

			descLabel := widget.NewLabel("Description")
			descLabel.TextStyle = fyne.TextStyle{Italic: true}

			rarityLabel := widget.NewLabel("Rarity")

			// Create a cooldown timer placeholder (will be shown/hidden as needed)
			cooldownTimer := NewCooldownTimer()
			cooldownTimer.Hide() // Hidden by default

			itemContainer := container.NewVBox(
				container.NewHBox(icon, nameLabel),
				descLabel,
				rarityLabel,
				cooldownTimer, // Add cooldown timer to list item
			)

			return itemContainer
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			// Update list item with gift data
			gifts := gsd.getAvailableGifts()
			if id >= len(gifts) {
				return
			}

			gift := gifts[id]
			itemContainer := obj.(*fyne.Container)

			// Update name label
			nameContainer := itemContainer.Objects[0].(*fyne.Container)
			nameLabel := nameContainer.Objects[1].(*widget.Label)
			nameLabel.SetText(gift.Name)

			// Update description
			descLabel := itemContainer.Objects[1].(*widget.Label)
			desc := gift.Description
			if len(desc) > 50 {
				desc = desc[:47] + "..."
			}
			descLabel.SetText(desc)

			// Update rarity with simple text (keep it simple - color coding can be added later)
			rarityLabel := itemContainer.Objects[2].(*widget.Label)
			rarityLabel.SetText(fmt.Sprintf("Rarity: %s", strings.Title(gift.Rarity)))

			// Handle cooldown display
			cooldownTimer := itemContainer.Objects[3].(*CooldownTimer)
			if gsd.giftManager.IsGiftOnCooldown(gift.ID) {
				remaining := gsd.giftManager.GetGiftCooldownRemaining(gift.ID)
				cooldownTimer.StartCooldown(remaining)
				cooldownTimer.Show()

				// Set completion callback to refresh the list when cooldown expires
				cooldownTimer.SetOnComplete(func() {
					cooldownTimer.Hide()
					gsd.giftList.Refresh()
				})
			} else {
				cooldownTimer.Hide()
			}
		},
	)

	// Handle gift selection
	gsd.giftList.OnSelected = func(id widget.ListItemID) {
		gifts := gsd.getAvailableGifts()
		if id < len(gifts) {
			gsd.selectedGift = gifts[id]
			gsd.updateGiveButtonState()
		}
	}

	// Set initial size for the list
	gsd.giftList.Resize(fyne.NewSize(300, 200))
}

// createNotesEntry initializes the notes input field
// Follows the character notes configuration if available
func (gsd *GiftSelectionDialog) createNotesEntry() {
	gsd.notesEntry = widget.NewMultiLineEntry()
	gsd.notesEntry.SetPlaceHolder("Add a personal message with your gift...")
	gsd.notesEntry.Wrapping = fyne.TextWrapWord

	// Set size constraints
	gsd.notesEntry.Resize(fyne.NewSize(300, 80))

	// Update give button state when notes change
	gsd.notesEntry.OnChanged = func(string) {
		gsd.updateGiveButtonState()
	}
}

// createButtons initializes the action buttons
// Uses existing button styling patterns from ContextMenu
func (gsd *GiftSelectionDialog) createButtons() {
	// Give button - primary action
	gsd.giveButton = widget.NewButton("Give Gift", func() {
		gsd.handleGiveGift()
	})
	gsd.giveButton.Importance = widget.HighImportance
	gsd.giveButton.Disable() // Initially disabled until gift selected

	// Cancel button - secondary action
	gsd.cancelButton = widget.NewButton("Cancel", func() {
		gsd.Hide()
		if gsd.onCancel != nil {
			gsd.onCancel()
		}
	})
	gsd.cancelButton.Importance = widget.LowImportance
}

// createLayout assembles the dialog layout
// Uses container layout patterns similar to DialogBubble
func (gsd *GiftSelectionDialog) createLayout() {
	// Title label
	titleLabel := widget.NewLabel("Select a Gift")
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}
	titleLabel.Alignment = fyne.TextAlignCenter

	// Instructions
	instructionLabel := widget.NewLabel("Choose a gift to give and add a personal note:")
	instructionLabel.Wrapping = fyne.TextWrapWord

	// Button container
	buttonContainer := container.NewHBox(
		gsd.cancelButton,
		widget.NewSeparator(), // Visual separator
		gsd.giveButton,
	)

	// Main content layout
	contentContainer := container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
		instructionLabel,
		gsd.giftList,
		widget.NewLabel("Personal Message:"),
		gsd.notesEntry,
		widget.NewSeparator(),
		buttonContainer,
	)

	// Add padding around content
	paddedContent := container.NewPadded(contentContainer)

	// Create final layout with background
	gsd.content = container.NewBorder(
		nil, nil, nil, nil,
		gsd.background,
		paddedContent,
	)

	// Set initial size
	gsd.content.Resize(fyne.NewSize(350, 450))
}

// getAvailableGifts returns sorted list of gifts that can be given
// Uses the gift manager's filtering logic
func (gsd *GiftSelectionDialog) getAvailableGifts() []*character.GiftDefinition {
	if gsd.giftManager == nil {
		return []*character.GiftDefinition{}
	}

	// Get available gifts from manager
	gifts := gsd.giftManager.GetAvailableGifts()

	// Sort by rarity and name for consistent ordering
	sort.Slice(gifts, func(i, j int) bool {
		// Sort by rarity first (common to legendary)
		rarityOrder := map[string]int{
			"common":    1,
			"uncommon":  2,
			"rare":      3,
			"epic":      4,
			"legendary": 5,
		}

		iRarity := rarityOrder[gifts[i].Rarity]
		jRarity := rarityOrder[gifts[j].Rarity]

		if iRarity != jRarity {
			return iRarity < jRarity
		}

		// Then sort by name
		return gifts[i].Name < gifts[j].Name
	})

	return gifts
}

// updateGiveButtonState enables/disables the give button based on selection
func (gsd *GiftSelectionDialog) updateGiveButtonState() {
	if gsd.selectedGift != nil && !gsd.giftManager.IsGiftOnCooldown(gsd.selectedGift.ID) {
		gsd.giveButton.Enable()
	} else {
		gsd.giveButton.Disable()
	}
}

// handleGiveGift processes the gift giving action
// Integrates with existing gift manager and provides user feedback
func (gsd *GiftSelectionDialog) handleGiveGift() {
	if gsd.selectedGift == nil {
		return
	}

	// Check cooldown before proceeding (double-check for safety)
	if gsd.giftManager.IsGiftOnCooldown(gsd.selectedGift.ID) {
		// This shouldn't happen if UI is working correctly, but add safety check
		return
	}

	// Get notes text (may be empty)
	notes := strings.TrimSpace(gsd.notesEntry.Text)

	// Validate notes length if gift has restrictions
	if gsd.selectedGift.Notes.Enabled && gsd.selectedGift.Notes.MaxLength > 0 {
		if len(notes) > gsd.selectedGift.Notes.MaxLength {
			// Show error without crashing - truncate instead
			notes = notes[:gsd.selectedGift.Notes.MaxLength]
		}
	}

	// Give the gift using existing gift manager
	response, err := gsd.giftManager.GiveGift(gsd.selectedGift.ID, notes)

	// Hide dialog first
	gsd.Hide()

	// Handle result
	if err != nil {
		// Create error response for consistent handling
		errorResponse := &character.GiftResponse{
			ErrorMessage: fmt.Sprintf("Failed to give gift: %v", err),
		}
		if gsd.onGiftGiven != nil {
			gsd.onGiftGiven(errorResponse)
		}
		return
	}

	// Success - notify callback with response
	if gsd.onGiftGiven != nil {
		gsd.onGiftGiven(response)
	}
}

// Show displays the gift selection dialog
// Refreshes available gifts and resets selection state
func (gsd *GiftSelectionDialog) Show() {
	gsd.visible = true

	// Reset selection state
	gsd.selectedGift = nil
	gsd.notesEntry.SetText("")
	gsd.giftList.UnselectAll()
	gsd.updateGiveButtonState()

	// Refresh gift list data
	gsd.giftList.Refresh()

	// Show the dialog
	gsd.content.Show()
	gsd.Refresh()
}

// Hide hides the gift selection dialog
func (gsd *GiftSelectionDialog) Hide() {
	gsd.visible = false
	gsd.content.Hide()
	gsd.Refresh()
}

// IsVisible returns whether the dialog is currently visible
func (gsd *GiftSelectionDialog) IsVisible() bool {
	return gsd.visible
}

// SetOnGiftGiven sets the callback for when a gift is successfully given
func (gsd *GiftSelectionDialog) SetOnGiftGiven(callback func(*character.GiftResponse)) {
	gsd.onGiftGiven = callback
}

// SetOnCancel sets the callback for when the dialog is cancelled
func (gsd *GiftSelectionDialog) SetOnCancel(callback func()) {
	gsd.onCancel = callback
}

// CreateRenderer creates the Fyne renderer for the gift dialog
func (gsd *GiftSelectionDialog) CreateRenderer() fyne.WidgetRenderer {
	return &giftDialogRenderer{
		dialog:  gsd,
		content: gsd.content,
	}
}

// giftDialogRenderer implements fyne.WidgetRenderer for gift dialogs
type giftDialogRenderer struct {
	dialog  *GiftSelectionDialog
	content *fyne.Container
}

// Layout arranges the gift dialog components
func (r *giftDialogRenderer) Layout(size fyne.Size) {
	if r.dialog.visible && r.content != nil {
		// Center the dialog on the screen
		r.content.Resize(r.content.Size())

		// Position dialog in center of available space
		dialogSize := r.content.Size()
		x := (size.Width - dialogSize.Width) / 2
		y := (size.Height - dialogSize.Height) / 2
		r.content.Move(fyne.NewPos(x, y))
	}
}

// MinSize returns the minimum size for the gift dialog
func (r *giftDialogRenderer) MinSize() fyne.Size {
	if r.dialog.visible {
		return fyne.NewSize(350, 450)
	}
	return fyne.NewSize(0, 0)
}

// Objects returns the canvas objects for rendering
func (r *giftDialogRenderer) Objects() []fyne.CanvasObject {
	if r.dialog.visible && r.content != nil {
		return []fyne.CanvasObject{r.content}
	}
	return []fyne.CanvasObject{}
}

// Refresh redraws the gift dialog
func (r *giftDialogRenderer) Refresh() {
	if r.dialog.visible && r.content != nil {
		r.content.Refresh()
	}
}

// Destroy cleans up gift dialog resources
func (r *giftDialogRenderer) Destroy() {
	// No special cleanup needed - Fyne handles widget cleanup
}
