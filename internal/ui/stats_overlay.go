package ui

import (
	"fmt"
	"image/color"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"desktop-companion/internal/character"
)

// StatsOverlay displays pet stats as an optional UI overlay
// Uses Fyne widgets to avoid custom implementations - follows "lazy programmer" approach
type StatsOverlay struct {
	widget.BaseWidget
	character    *character.Character
	container    *fyne.Container
	progressBars map[string]*widget.ProgressBar
	statLabels   map[string]*widget.Label
	visible      bool
	updateTicker *time.Ticker
	stopUpdate   chan bool
	mu           sync.RWMutex // Protects updateTicker and background goroutine state
}

// NewStatsOverlay creates a new stats overlay widget
// Only creates UI elements when character has game features enabled
func NewStatsOverlay(char *character.Character) *StatsOverlay {
	so := &StatsOverlay{
		character:    char,
		progressBars: make(map[string]*widget.ProgressBar),
		statLabels:   make(map[string]*widget.Label),
		visible:      false,
		stopUpdate:   make(chan bool, 1),
	}

	so.ExtendBaseWidget(so)
	so.createStatsWidgets()

	return so
}

// createStatsWidgets creates progress bars and labels for each stat
// Uses standard Fyne widgets to minimize custom code
func (so *StatsOverlay) createStatsWidgets() {
	gameState := so.character.GetGameState()
	if gameState == nil {
		// Create empty container for characters without game features
		so.container = container.NewVBox()
		return
	}

	widgets := []fyne.CanvasObject{}

	// Get current stats to determine which progress bars to create
	stats := gameState.GetStats()

	// Create a progress bar and label for each stat
	for statName := range stats {
		// Create label for stat name and value
		label := widget.NewLabel(fmt.Sprintf("%s: 0", capitalizeFirst(statName)))
		so.statLabels[statName] = label

		// Create progress bar for stat value
		progressBar := widget.NewProgressBar()
		progressBar.Min = 0
		progressBar.Max = 100 // Assuming max stat is 100, will be adjusted dynamically

		// ProgressBar will be updated in the update loop

		so.progressBars[statName] = progressBar

		// Add label and progress bar to widgets list
		widgets = append(widgets, label, progressBar)
	}

	// Create container with vertical layout for compact display
	so.container = container.NewVBox(widgets...)
	so.container.Hide() // Start hidden
}

// CreateRenderer creates the Fyne renderer for this widget
func (so *StatsOverlay) CreateRenderer() fyne.WidgetRenderer {
	return &statsOverlayRenderer{
		overlay:   so,
		container: so.container,
	}
}

// Toggle shows/hides the stats overlay
func (so *StatsOverlay) Toggle() {
	so.visible = !so.visible

	if so.visible {
		so.Show()
	} else {
		so.Hide()
	}
}

// Show displays the stats overlay and starts update loop
func (so *StatsOverlay) Show() {
	gameState := so.character.GetGameState()
	if gameState == nil {
		return // No stats to show for characters without game features
	}

	so.visible = true
	so.container.Show()
	so.startUpdateLoop()
	so.Refresh()
}

// Hide hides the stats overlay and stops update loop
func (so *StatsOverlay) Hide() {
	so.visible = false
	so.container.Hide()
	so.stopUpdateLoop()
	so.Refresh()
}

// IsVisible returns whether the overlay is currently visible
func (so *StatsOverlay) IsVisible() bool {
	return so.visible
}

// startUpdateLoop begins periodic updates of stat display
// Updates every 2 seconds to balance responsiveness with performance
func (so *StatsOverlay) startUpdateLoop() {
	so.mu.Lock()
	if so.updateTicker != nil {
		so.mu.Unlock()
		return // Already running
	}

	so.updateTicker = time.NewTicker(2 * time.Second)
	ticker := so.updateTicker // Capture ticker under lock
	so.mu.Unlock()

	go func() {
		if ticker == nil {
			return
		}

		for {
			select {
			case <-ticker.C:
				if so.character != nil {
					so.updateStatDisplay()
				}
			case <-so.stopUpdate:
				return
			}
		}
	}()
}

// stopUpdateLoop stops the periodic update of stat display
func (so *StatsOverlay) stopUpdateLoop() {
	so.mu.Lock()
	defer so.mu.Unlock()

	if so.updateTicker != nil {
		so.updateTicker.Stop()
		so.updateTicker = nil

		// Signal stop to goroutine
		select {
		case so.stopUpdate <- true:
		default:
		}
	}
}

// updateStatDisplay refreshes the progress bars and labels with current stat values
// Uses character's GetGameState method to get current values
func (so *StatsOverlay) updateStatDisplay() {
	if so.character == nil {
		return
	}

	gameState := so.character.GetGameState()
	if gameState == nil {
		return
	}

	stats := gameState.GetStats()
	if stats == nil {
		return
	}

	criticalStates := gameState.GetCriticalStates()

	for statName, currentValue := range stats {
		// Update progress bar
		if progressBar, exists := so.progressBars[statName]; exists {
			// Calculate percentage (assuming max is 100 for simplicity)
			percentage := currentValue / 100.0
			progressBar.SetValue(percentage)

			// ProgressBar text is not settable in Fyne, use labels instead
		}

		// Update label with critical state indication
		if label, exists := so.statLabels[statName]; exists {
			isCritical := contains(criticalStates, statName)
			if isCritical {
				label.SetText(fmt.Sprintf("%s: %.0f CRITICAL", capitalizeFirst(statName), currentValue))
			} else {
				label.SetText(fmt.Sprintf("%s: %.0f", capitalizeFirst(statName), currentValue))
			}
		}
	}
}

// GetContainer returns the container for external positioning
func (so *StatsOverlay) GetContainer() *fyne.Container {
	return so.container
}

// statsOverlayRenderer implements fyne.WidgetRenderer for the stats overlay
type statsOverlayRenderer struct {
	overlay   *StatsOverlay
	container *fyne.Container
}

// Layout arranges the stats overlay within the widget bounds
func (r *statsOverlayRenderer) Layout(size fyne.Size) {
	if r.container != nil {
		r.container.Resize(size)
	}
}

// MinSize returns the minimum size required for the stats overlay
func (r *statsOverlayRenderer) MinSize() fyne.Size {
	if r.container != nil {
		return r.container.MinSize()
	}
	return fyne.NewSize(200, 100) // Default minimum size
}

// Refresh updates the visual representation
func (r *statsOverlayRenderer) Refresh() {
	if r.container != nil {
		r.container.Refresh()
	}
}

// Objects returns the list of objects to render
func (r *statsOverlayRenderer) Objects() []fyne.CanvasObject {
	if r.container != nil {
		return []fyne.CanvasObject{r.container}
	}
	return []fyne.CanvasObject{}
}

// Destroy cleans up resources
func (r *statsOverlayRenderer) Destroy() {
	// Cleanup is handled by the overlay widget
}

// Helper function to capitalize first letter of stat names
func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return fmt.Sprintf("%c%s", s[0]-32, s[1:])
}

// Helper function to check if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// getStatColor returns appropriate color based on stat value and critical threshold
// Returns red for critical values, orange for low values, green for healthy values
func getStatColor(current, max, critical float64) color.RGBA {
	percentage := current / max

	if current <= critical {
		// Critical: red
		return color.RGBA{255, 0, 0, 255}
	} else if percentage < 0.5 {
		// Low: orange
		return color.RGBA{255, 165, 0, 255}
	} else {
		// Healthy: green
		return color.RGBA{0, 255, 0, 255}
	}
}
