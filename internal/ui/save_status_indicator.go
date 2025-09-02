package ui

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// SaveStatus represents the current state of the save operation
// Used to drive visual feedback through icons and animations
type SaveStatus int

const (
	// SaveStatusIdle indicates no save operations in progress
	SaveStatusIdle SaveStatus = iota
	// SaveStatusSaving indicates a save operation is currently in progress
	SaveStatusSaving
	// SaveStatusSaved indicates the last save operation completed successfully
	SaveStatusSaved
	// SaveStatusError indicates the last save operation failed with an error
	SaveStatusError
)

// SaveStatusIndicator is a small widget that shows save operation status
// Designed to be unobtrusive in the corner of the character window
// Uses standard Fyne icons for consistent visual language
type SaveStatusIndicator struct {
	widget.BaseWidget
	icon      *widget.Icon
	status    SaveStatus
	lastSaved time.Time
	errorMsg  string
}

// NewSaveStatusIndicator creates a new save status indicator widget
// Returns widget in idle state, ready to be positioned in window corner
func NewSaveStatusIndicator() *SaveStatusIndicator {
	indicator := &SaveStatusIndicator{
		status: SaveStatusIdle,
	}

	indicator.updateIcon()
	indicator.ExtendBaseWidget(indicator)
	return indicator
}

// SetStatus updates the current save status and refreshes the display
// message parameter is used for error messages when status is SaveStatusError
// For SaveStatusSaved, automatically records timestamp for display purposes
func (ssi *SaveStatusIndicator) SetStatus(status SaveStatus, message string) {
	ssi.status = status

	if status == SaveStatusSaved {
		ssi.lastSaved = time.Now()
		ssi.errorMsg = "" // Clear any previous error
	} else if status == SaveStatusError {
		ssi.errorMsg = message
	}

	ssi.updateIcon()
	ssi.Refresh()
}

// GetStatus returns the current save status for testing and integration
func (ssi *SaveStatusIndicator) GetStatus() SaveStatus {
	return ssi.status
}

// GetLastSaved returns the timestamp of the last successful save operation
func (ssi *SaveStatusIndicator) GetLastSaved() time.Time {
	return ssi.lastSaved
}

// GetErrorMessage returns the current error message if status is SaveStatusError
func (ssi *SaveStatusIndicator) GetErrorMessage() string {
	return ssi.errorMsg
}

// updateIcon selects appropriate icon based on current save status
// Uses standard Fyne theme icons for consistency with system UI
func (ssi *SaveStatusIndicator) updateIcon() {
	switch ssi.status {
	case SaveStatusSaving:
		// Use refresh icon to indicate active operation
		ssi.icon = widget.NewIcon(theme.ViewRefreshIcon())
	case SaveStatusSaved:
		// Use confirm icon to indicate successful save
		ssi.icon = widget.NewIcon(theme.ConfirmIcon())
	case SaveStatusError:
		// Use error icon to indicate save failure
		ssi.icon = widget.NewIcon(theme.ErrorIcon())
	default:
		// Use document save icon for idle state
		ssi.icon = widget.NewIcon(theme.DocumentSaveIcon())
	}
}

// CreateRenderer creates the widget renderer for the save status indicator
// Returns a simple renderer that displays just the icon
func (ssi *SaveStatusIndicator) CreateRenderer() fyne.WidgetRenderer {
	return &saveStatusRenderer{
		indicator: ssi,
		objects:   []fyne.CanvasObject{ssi.icon},
	}
}

// saveStatusRenderer implements fyne.WidgetRenderer for SaveStatusIndicator
// Simple renderer that displays the status icon with minimal visual footprint
type saveStatusRenderer struct {
	indicator *SaveStatusIndicator
	objects   []fyne.CanvasObject
}

// Layout positions the icon to fill the available space
func (r *saveStatusRenderer) Layout(size fyne.Size) {
	if r.indicator.icon != nil {
		r.indicator.icon.Resize(size)
		r.indicator.icon.Move(fyne.NewPos(0, 0))
	}
}

// MinSize returns the minimum size needed for the status indicator
// Small 16x16 icon suitable for corner placement
func (r *saveStatusRenderer) MinSize() fyne.Size {
	return fyne.NewSize(16, 16)
}

// Objects returns the list of objects to render (just the icon)
func (r *saveStatusRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

// Refresh updates the renderer when the widget changes
func (r *saveStatusRenderer) Refresh() {
	// Icon updates are handled by updateIcon() which creates new icon instances
	// No additional refresh logic needed
}

// Destroy cleans up any resources used by the renderer
func (r *saveStatusRenderer) Destroy() {
	// No persistent resources to clean up
}
