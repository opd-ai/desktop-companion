package ui

import (
	"testing"
	"time"

	"fyne.io/fyne/v2/test"

	"github.com/opd-ai/desktop-companion/lib/network"
)

func TestNewActivityFeed(t *testing.T) {
	t.Run("Valid tracker", func(t *testing.T) {
		tracker := network.NewActivityTracker(50)
		feed := NewActivityFeed(tracker)

		if feed == nil {
			t.Errorf("Expected feed to not be nil")
		}

		if feed.tracker != tracker {
			t.Errorf("Expected feed tracker to match provided tracker")
		}

		if feed.maxEvents != 50 {
			t.Errorf("Expected maxEvents to be 50, got %d", feed.maxEvents)
		}
	})

	t.Run("Nil tracker", func(t *testing.T) {
		feed := NewActivityFeed(nil)

		if feed != nil {
			t.Errorf("Expected feed to be nil when tracker is nil")
		}
	})
}

func TestActivityFeed_CreateRenderer(t *testing.T) {
	tracker := network.NewActivityTracker(10)
	feed := NewActivityFeed(tracker)

	renderer := feed.CreateRenderer()
	if renderer == nil {
		t.Errorf("Expected renderer to not be nil")
	}
}

func TestActivityFeed_AddTestEvent(t *testing.T) {
	tracker := network.NewActivityTracker(10)
	feed := NewActivityFeed(tracker)

	// Add test event
	feed.AddTestEvent(network.ActivityJoined, "Test character joined")

	// Verify event was added to tracker
	events := tracker.GetAllEvents()
	if len(events) != 1 {
		t.Errorf("Expected 1 event in tracker, got %d", len(events))
	}

	if events[0].Type != network.ActivityJoined {
		t.Errorf("Expected ActivityJoined, got %v", events[0].Type)
	}

	if events[0].Description != "Test character joined" {
		t.Errorf("Expected 'Test character joined', got '%s'", events[0].Description)
	}
}

func TestActivityFeed_AddTestEvent_NilTracker(t *testing.T) {
	// Create feed with nil tracker (shouldn't happen in practice but test safety)
	feed := &ActivityFeed{tracker: nil}

	// Should not panic
	feed.AddTestEvent(network.ActivityChat, "Test message")

	// No way to verify since tracker is nil, but test passes if no panic
}

func TestActivityFeed_Clear(t *testing.T) {
	tracker := network.NewActivityTracker(10)
	feed := NewActivityFeed(tracker)

	// Add some events through the tracker
	for i := 0; i < 3; i++ {
		event := network.ActivityEvent{
			Type:          network.ActivityChat,
			PeerID:        "peer1",
			CharacterName: "TestChar",
			Description:   "Test message",
			Timestamp:     time.Now(),
		}
		tracker.AddEvent(event)
	}

	// Give time for listener to process all events
	time.Sleep(50 * time.Millisecond)

	// Verify container has some items
	if feed.GetEventCount() == 0 {
		t.Errorf("Expected some objects in container before clear")
	}

	// Clear the tracker first to prevent new events from listener goroutines
	tracker.Clear()

	// Clear the feed
	feed.Clear()

	// Give extra time for any remaining goroutines to complete
	time.Sleep(20 * time.Millisecond)

	// Verify container is empty
	if feed.GetEventCount() != 0 {
		t.Errorf("Expected 0 objects in container after clear, got %d", feed.GetEventCount())
	}
}

func TestActivityFeed_SetMaxEvents(t *testing.T) {
	tracker := network.NewActivityTracker(10)
	feed := NewActivityFeed(tracker)

	originalMax := feed.maxEvents

	t.Run("Valid max events", func(t *testing.T) {
		feed.SetMaxEvents(25)
		if feed.maxEvents != 25 {
			t.Errorf("Expected maxEvents to be 25, got %d", feed.maxEvents)
		}
	})

	t.Run("Zero max events uses default", func(t *testing.T) {
		feed.SetMaxEvents(0)
		if feed.maxEvents != 50 {
			t.Errorf("Expected maxEvents to be 50 (default), got %d", feed.maxEvents)
		}
	})

	t.Run("Negative max events uses default", func(t *testing.T) {
		feed.SetMaxEvents(-10)
		if feed.maxEvents != 50 {
			t.Errorf("Expected maxEvents to be 50 (default), got %d", feed.maxEvents)
		}
	})

	// Restore original for other tests
	feed.maxEvents = originalMax
}

func TestActivityFeed_GetContainer(t *testing.T) {
	tracker := network.NewActivityTracker(10)
	feed := NewActivityFeed(tracker)

	container := feed.GetContainer()
	if container == nil {
		t.Errorf("Expected container to not be nil")
	}
}

func TestActivityFeed_EventListener(t *testing.T) {
	tracker := network.NewActivityTracker(10)
	feed := NewActivityFeed(tracker)

	// Create test app to enable widget rendering
	testApp := test.NewApp()
	defer testApp.Quit()

	// Add an event through the tracker (should trigger listener)
	event := network.ActivityEvent{
		Type:          network.ActivityBattle,
		PeerID:        "peer1",
		CharacterName: "BattleChar",
		Description:   "Started battle",
		Timestamp:     time.Now(),
	}

	tracker.AddEvent(event)

	// Give time for async listener to process
	time.Sleep(20 * time.Millisecond)

	// Check that the feed container has the event
	eventCount := feed.GetEventCount()
	if eventCount != 1 {
		t.Errorf("Expected 1 object in feed container, got %d", eventCount)
	}
}

func TestActivityFeed_MaxEventsLimit(t *testing.T) {
	tracker := network.NewActivityTracker(100)
	feed := NewActivityFeed(tracker)
	feed.SetMaxEvents(3) // Set small limit for testing

	// Create test app
	testApp := test.NewApp()
	defer testApp.Quit()

	// Add more events than the limit
	for i := 0; i < 5; i++ {
		event := network.ActivityEvent{
			Type:          network.ActivityChat,
			PeerID:        "peer1",
			CharacterName: "TestChar",
			Description:   "Test message",
			Timestamp:     time.Now(),
		}
		tracker.AddEvent(event)
		time.Sleep(5 * time.Millisecond) // Small delay between events
	}

	// Give time for all listeners to process
	time.Sleep(50 * time.Millisecond)

	// Should only show maxEvents items
	eventCount := feed.GetEventCount()
	if eventCount > feed.maxEvents {
		t.Errorf("Expected at most %d objects in feed container, got %d", feed.maxEvents, eventCount)
	}
}

func TestActivityFeed_EventStyling(t *testing.T) {
	tracker := network.NewActivityTracker(10)
	feed := NewActivityFeed(tracker)

	// Create test app
	testApp := test.NewApp()
	defer testApp.Quit()

	// Test different event types to verify styling
	eventTypes := []network.ActivityType{
		network.ActivityJoined,
		network.ActivityLeft,
		network.ActivityBattle,
		network.ActivityChat,
		network.ActivityInteraction,
	}

	for _, eventType := range eventTypes {
		event := network.ActivityEvent{
			Type:          eventType,
			PeerID:        "peer1",
			CharacterName: "TestChar",
			Description:   "Test event",
			Timestamp:     time.Now(),
		}
		tracker.AddEvent(event)
		time.Sleep(5 * time.Millisecond) // Small delay between events
	}

	// Give time for listeners to process
	time.Sleep(100 * time.Millisecond)

	// Verify we have the expected number of events displayed
	expectedCount := len(eventTypes)
	eventCount := feed.GetEventCount()
	if eventCount != expectedCount {
		t.Errorf("Expected %d objects in feed container, got %d", expectedCount, eventCount)
	}
}

func TestActivityFeed_Integration(t *testing.T) {
	// Integration test with realistic usage
	tracker := network.NewActivityTracker(50)
	feed := NewActivityFeed(tracker)

	// Create test app
	testApp := test.NewApp()
	defer testApp.Quit()

	// Simulate various network activities
	activities := []struct {
		eventType   network.ActivityType
		description string
	}{
		{network.ActivityJoined, "Player1 joined the network"},
		{network.ActivityChat, "Player1: Hello everyone!"},
		{network.ActivityInteraction, "Player1 petted their character"},
		{network.ActivityBattle, "Player1 started a battle"},
		{network.ActivityLeft, "Player1 left the network"},
	}

	for _, activity := range activities {
		event := network.ActivityEvent{
			Type:          activity.eventType,
			PeerID:        "player1",
			CharacterName: "Player1",
			Description:   activity.description,
			Timestamp:     time.Now(),
		}
		tracker.AddEvent(event)
		time.Sleep(10 * time.Millisecond) // Small delay between events
	}

	// Give time for all listeners to process
	time.Sleep(100 * time.Millisecond)

	// Verify all activities are displayed
	eventCount := feed.GetEventCount()
	if eventCount != len(activities) {
		t.Errorf("Expected %d activities displayed, got %d", len(activities), eventCount)
	}

	// Verify tracker and feed are in sync
	trackerEvents := tracker.GetAllEvents()
	if len(trackerEvents) != len(activities) {
		t.Errorf("Expected %d events in tracker, got %d", len(activities), len(trackerEvents))
	}
}
