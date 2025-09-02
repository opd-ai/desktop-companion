package ui

import (
	"fmt"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"desktop-companion/internal/network"
)

// ActivityFeed displays a scrollable log of network peer activities
// Uses Fyne standard widgets following the project's "library-first" philosophy
type ActivityFeed struct {
	widget.BaseWidget
	vbox      *fyne.Container
	scroll    *container.Scroll
	tracker   *network.ActivityTracker
	maxEvents int
	cleared   bool         // Flag to ignore events after clear
	mu        sync.RWMutex // Protects UI operations from concurrent access
}

// NewActivityFeed creates a new activity feed widget with the given tracker
func NewActivityFeed(tracker *network.ActivityTracker) *ActivityFeed {
	if tracker == nil {
		return nil
	}

	feed := &ActivityFeed{
		tracker:   tracker,
		maxEvents: 50, // Display up to 50 events
	}

	feed.vbox = container.NewVBox()
	feed.scroll = container.NewScroll(feed.vbox)
	feed.scroll.SetMinSize(fyne.NewSize(300, 150))

	// Load initial events
	feed.refreshEvents()

	// Register listener for new events
	tracker.AddListener(func(event network.ActivityEvent) {
		feed.addEventToFeed(event)
	})

	feed.ExtendBaseWidget(feed)
	return feed
}

// CreateRenderer creates the widget renderer
func (af *ActivityFeed) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(af.scroll)
}

// refreshEvents reloads all events from the tracker
func (af *ActivityFeed) refreshEvents() {
	af.mu.Lock()
	defer af.mu.Unlock()

	if af.tracker == nil {
		return
	}

	events := af.tracker.GetRecentEvents(af.maxEvents)
	af.vbox.RemoveAll()

	for _, event := range events {
		af.addEventWidget(event)
	}

	af.Refresh()
}

// addEventToFeed adds a new event to the feed (called by listener)
// Uses mutex protection for thread-safe UI operations
func (af *ActivityFeed) addEventToFeed(event network.ActivityEvent) {
	af.mu.Lock()
	defer af.mu.Unlock()

	// Ignore events if feed has been cleared
	if af.cleared {
		return
	}

	// Add new event widget
	af.addEventWidget(event)

	// Remove old events if we exceed the limit
	if len(af.vbox.Objects) > af.maxEvents {
		af.vbox.Remove(af.vbox.Objects[0])
	}

	// Auto-scroll to bottom to show newest events
	af.scroll.ScrollToBottom()
	af.Refresh()
}

// addEventWidget creates and adds a single event widget to the container
// Must be called with mutex held
func (af *ActivityFeed) addEventWidget(event network.ActivityEvent) {
	timeStr := event.Timestamp.Format("15:04")
	eventText := fmt.Sprintf("[%s] %s", timeStr, event.Description)

	// Create styled label based on event type
	label := widget.NewLabel(eventText)
	label.Wrapping = fyne.TextWrapWord

	// Style based on activity type
	switch event.Type {
	case network.ActivityJoined:
		label.Importance = widget.SuccessImportance
	case network.ActivityLeft:
		label.Importance = widget.MediumImportance
	case network.ActivityBattle:
		label.Importance = widget.WarningImportance
	case network.ActivityChat:
		label.Importance = widget.LowImportance
	default:
		label.Importance = widget.LowImportance
	}

	af.vbox.Add(label)
}

// Clear removes all events from the display
func (af *ActivityFeed) Clear() {
	af.mu.Lock()
	defer af.mu.Unlock()

	af.cleared = true
	af.vbox.RemoveAll()
	af.Refresh()
}

// SetMaxEvents updates the maximum number of events to display
func (af *ActivityFeed) SetMaxEvents(max int) {
	af.mu.Lock()
	defer af.mu.Unlock()

	if max <= 0 {
		max = 50
	}
	af.maxEvents = max
	af.refreshEventsUnsafe() // Call internal method that doesn't take mutex
}

// refreshEventsUnsafe reloads all events from the tracker (must be called with mutex held)
func (af *ActivityFeed) refreshEventsUnsafe() {
	if af.tracker == nil {
		return
	}

	events := af.tracker.GetRecentEvents(af.maxEvents)
	af.vbox.RemoveAll()

	for _, event := range events {
		af.addEventWidget(event)
	}

	af.Refresh()
}

// GetContainer returns the main container for layout purposes
func (af *ActivityFeed) GetContainer() *fyne.Container {
	return container.NewWithoutLayout(af.scroll)
}

// GetEventCount returns the current number of events in the feed (thread-safe)
func (af *ActivityFeed) GetEventCount() int {
	af.mu.RLock()
	defer af.mu.RUnlock()
	return len(af.vbox.Objects)
}

// AddTestEvent adds a test event for UI testing purposes
func (af *ActivityFeed) AddTestEvent(eventType network.ActivityType, description string) {
	if af.tracker == nil {
		return
	}

	event := network.ActivityEvent{
		Type:          eventType,
		PeerID:        "test-peer",
		CharacterName: "Test Character",
		Description:   description,
		Timestamp:     time.Now(),
	}

	af.tracker.AddEvent(event)
}
