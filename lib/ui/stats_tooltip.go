package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/opd-ai/desktop-companion/lib/character"
)

// StatsTooltip displays quick stats as a lightweight tooltip
// Leverages existing StatsOverlay patterns for consistency
type StatsTooltip struct {
	widget.BaseWidget
	character *character.Character
	container *fyne.Container
	visible   bool
}

// NewStatsTooltip creates a new stats tooltip widget
func NewStatsTooltip(char *character.Character) *StatsTooltip {
	st := &StatsTooltip{
		character: char,
		visible:   false,
	}

	st.ExtendBaseWidget(st)
	st.createTooltipContent()

	return st
}

// createTooltipContent creates compact stat display for tooltip
func (st *StatsTooltip) createTooltipContent() {
	gameState := st.character.GetGameState()
	if gameState == nil {
		// Create empty container for characters without game features
		st.container = container.NewVBox(
			widget.NewLabel("No stats available"),
		)
		return
	}

	// Get current stats
	stats := gameState.GetStats()
	if len(stats) == 0 {
		st.container = container.NewVBox(
			widget.NewLabel("No stats available"),
		)
		return
	}

	widgets := []fyne.CanvasObject{}

	// Add title
	title := widget.NewLabel("Quick Stats")
	title.TextStyle.Bold = true
	widgets = append(widgets, title)

	// Create compact stat display
	for statName, value := range stats {
		// Format: "Stat: 85/100" (assuming max 100 like StatsOverlay)
		statText := fmt.Sprintf("%s: %.0f/100", capitalizeFirst(statName), value)

		label := widget.NewLabel(statText)

		// Simple percentage-based formatting (no color since Fyne doesn't support it easily)
		if value < 30 {
			// Low stats - add indicator
			statText = fmt.Sprintf("%s: %.0f/100 (LOW)", capitalizeFirst(statName), value)
		} else if value < 60 {
			// Medium stats
			statText = fmt.Sprintf("%s: %.0f/100", capitalizeFirst(statName), value)
		} else {
			// High stats - add indicator
			statText = fmt.Sprintf("%s: %.0f/100 (GOOD)", capitalizeFirst(statName), value)
		}

		label.SetText(statText)
		widgets = append(widgets, label)
	} // Create container with padding and background
	st.container = container.NewVBox(widgets...)
}

// Show makes the tooltip visible
func (st *StatsTooltip) Show() {
	st.visible = true
	st.Refresh()
}

// Hide makes the tooltip invisible
func (st *StatsTooltip) Hide() {
	st.visible = false
	st.Refresh()
}

// IsVisible returns whether the tooltip is currently visible
func (st *StatsTooltip) IsVisible() bool {
	return st.visible
}

// UpdateContent refreshes the tooltip content with current stats
func (st *StatsTooltip) UpdateContent() {
	st.createTooltipContent()
	st.Refresh()
}

// GetContainer returns the container for rendering
func (st *StatsTooltip) GetContainer() *fyne.Container {
	if st.visible {
		return st.container
	}
	return container.NewWithoutLayout() // Empty container when hidden
}

// CreateRenderer creates the Fyne renderer for the tooltip
func (st *StatsTooltip) CreateRenderer() fyne.WidgetRenderer {
	return &statsTooltipRenderer{
		tooltip: st,
	}
}

// statsTooltipRenderer implements fyne.WidgetRenderer for the stats tooltip
type statsTooltipRenderer struct {
	tooltip *StatsTooltip
}

// Layout arranges the tooltip content
func (r *statsTooltipRenderer) Layout(size fyne.Size) {
	if r.tooltip.container != nil {
		r.tooltip.container.Resize(size)
	}
}

// MinSize returns the minimum size for the tooltip
func (r *statsTooltipRenderer) MinSize() fyne.Size {
	if r.tooltip.container != nil {
		return r.tooltip.container.MinSize()
	}
	return fyne.NewSize(120, 80)
}

// Objects returns the canvas objects for rendering
func (r *statsTooltipRenderer) Objects() []fyne.CanvasObject {
	if r.tooltip.visible && r.tooltip.container != nil {
		return []fyne.CanvasObject{r.tooltip.container}
	}
	return []fyne.CanvasObject{}
}

// Refresh redraws the tooltip
func (r *statsTooltipRenderer) Refresh() {
	if r.tooltip.container != nil {
		r.tooltip.container.Refresh()
	}
}

// Destroy cleans up tooltip resources
func (r *statsTooltipRenderer) Destroy() {
	// No special cleanup needed
}
