package network

import (
	"fmt"
	"testing"
	"time"
)

func TestNewActivityTracker(t *testing.T) {
	tests := []struct {
		name           string
		maxEvents      int
		expectedMax    int
		shouldNotBeNil bool
	}{
		{
			name:           "Valid max events",
			maxEvents:      50,
			expectedMax:    50,
			shouldNotBeNil: true,
		},
		{
			name:           "Zero max events uses default",
			maxEvents:      0,
			expectedMax:    100,
			shouldNotBeNil: true,
		},
		{
			name:           "Negative max events uses default",
			maxEvents:      -5,
			expectedMax:    100,
			shouldNotBeNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewActivityTracker(tt.maxEvents)

			if !tt.shouldNotBeNil && tracker != nil {
				t.Errorf("Expected tracker to be nil")
			}
			if tt.shouldNotBeNil && tracker == nil {
				t.Errorf("Expected tracker to not be nil")
			}

			if tracker != nil && tracker.maxEvents != tt.expectedMax {
				t.Errorf("Expected maxEvents %d, got %d", tt.expectedMax, tracker.maxEvents)
			}
		})
	}
}

func TestActivityTracker_AddEvent(t *testing.T) {
	tracker := NewActivityTracker(3) // Small limit for testing

	// Test adding events
	event1 := ActivityEvent{
		Type:          ActivityJoined,
		PeerID:        "peer1",
		CharacterName: "Character1",
		Description:   "Test event 1",
	}

	tracker.AddEvent(event1)

	if tracker.GetEventCount() != 1 {
		t.Errorf("Expected 1 event, got %d", tracker.GetEventCount())
	}

	// Test automatic timestamp setting
	events := tracker.GetAllEvents()
	if len(events) != 1 {
		t.Errorf("Expected 1 event in GetAllEvents, got %d", len(events))
	}

	if events[0].Timestamp.IsZero() {
		t.Errorf("Expected timestamp to be set automatically")
	}
}

func TestActivityTracker_MaxEventsLimit(t *testing.T) {
	tracker := NewActivityTracker(2) // Very small limit

	// Add more events than the limit
	for i := 0; i < 5; i++ {
		event := ActivityEvent{
			Type:          ActivityInteraction,
			PeerID:        "peer1",
			CharacterName: "Character1",
			Description:   fmt.Sprintf("Event %d", i),
		}
		tracker.AddEvent(event)
	}

	// Should only keep the last 2 events
	if tracker.GetEventCount() != 2 {
		t.Errorf("Expected 2 events (limit), got %d", tracker.GetEventCount())
	}

	events := tracker.GetAllEvents()
	if len(events) != 2 {
		t.Errorf("Expected 2 events in GetAllEvents, got %d", len(events))
	}

	// Verify we kept the most recent events (events 3 and 4)
	if events[0].Description != "Event 3" {
		t.Errorf("Expected oldest kept event to be 'Event 3', got '%s'", events[0].Description)
	}
	if events[1].Description != "Event 4" {
		t.Errorf("Expected newest event to be 'Event 4', got '%s'", events[1].Description)
	}
}

func TestActivityTracker_GetRecentEvents(t *testing.T) {
	tracker := NewActivityTracker(10)

	// Add some events
	for i := 0; i < 5; i++ {
		event := ActivityEvent{
			Type:          ActivityChat,
			PeerID:        "peer1",
			CharacterName: "Character1",
			Description:   fmt.Sprintf("Message %d", i),
		}
		tracker.AddEvent(event)
	}

	tests := []struct {
		name          string
		count         int
		expectedCount int
	}{
		{
			name:          "Get 3 recent events",
			count:         3,
			expectedCount: 3,
		},
		{
			name:          "Get more events than available",
			count:         10,
			expectedCount: 5,
		},
		{
			name:          "Get zero events",
			count:         0,
			expectedCount: 0,
		},
		{
			name:          "Get negative count",
			count:         -1,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events := tracker.GetRecentEvents(tt.count)
			if len(events) != tt.expectedCount {
				t.Errorf("Expected %d events, got %d", tt.expectedCount, len(events))
			}

			// Verify events are in chronological order (oldest first)
			if len(events) > 1 {
				for i := 1; i < len(events); i++ {
					if events[i].Timestamp.Before(events[i-1].Timestamp) {
						t.Errorf("Events not in chronological order")
					}
				}
			}
		})
	}
}

func TestActivityTracker_AddListener(t *testing.T) {
	tracker := NewActivityTracker(10)
	called := false

	// Add listener
	tracker.AddListener(func(event ActivityEvent) {
		called = true
		if event.Type != ActivityBattle {
			t.Errorf("Expected ActivityBattle, got %v", event.Type)
		}
	})

	// Test listener is called
	event := ActivityEvent{
		Type:          ActivityBattle,
		PeerID:        "peer1",
		CharacterName: "Character1",
		Description:   "Battle started",
	}

	tracker.AddEvent(event)

	// Give goroutine time to execute
	time.Sleep(10 * time.Millisecond)

	if !called {
		t.Errorf("Expected listener to be called")
	}
}

func TestActivityTracker_AddListener_NilHandling(t *testing.T) {
	tracker := NewActivityTracker(10)

	// Adding nil listener should not panic
	tracker.AddListener(nil)

	// Adding event should not panic
	event := ActivityEvent{
		Type:          ActivityJoined,
		PeerID:        "peer1",
		CharacterName: "Character1",
		Description:   "Test event",
	}

	tracker.AddEvent(event)

	if tracker.GetEventCount() != 1 {
		t.Errorf("Expected 1 event after adding with nil listener, got %d", tracker.GetEventCount())
	}
}

func TestActivityTracker_Clear(t *testing.T) {
	tracker := NewActivityTracker(10)

	// Add some events and listeners
	for i := 0; i < 3; i++ {
		event := ActivityEvent{
			Type:          ActivityDiscovery,
			PeerID:        fmt.Sprintf("peer%d", i),
			CharacterName: fmt.Sprintf("Character%d", i),
			Description:   fmt.Sprintf("Discovery %d", i),
		}
		tracker.AddEvent(event)
	}

	tracker.AddListener(func(event ActivityEvent) {})

	if tracker.GetEventCount() != 3 {
		t.Errorf("Expected 3 events before clear, got %d", tracker.GetEventCount())
	}

	// Clear everything
	tracker.Clear()

	if tracker.GetEventCount() != 0 {
		t.Errorf("Expected 0 events after clear, got %d", tracker.GetEventCount())
	}

	events := tracker.GetAllEvents()
	if len(events) != 0 {
		t.Errorf("Expected 0 events in GetAllEvents after clear, got %d", len(events))
	}
}

func TestActivityType_String(t *testing.T) {
	tests := []struct {
		activityType ActivityType
		expected     string
	}{
		{ActivityJoined, "joined"},
		{ActivityLeft, "left"},
		{ActivityInteraction, "interaction"},
		{ActivityStateChange, "state_change"},
		{ActivityChat, "chat"},
		{ActivityBattle, "battle"},
		{ActivityDiscovery, "discovery"},
		{ActivityType(999), "unknown"}, // Invalid type
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.activityType.String()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestCreateEventHelpers(t *testing.T) {
	// Test CreateCharacterActionEvent
	t.Run("CreateCharacterActionEvent", func(t *testing.T) {
		event := CreateCharacterActionEvent("peer1", "TestChar", "clicked", map[string]string{"key": "value"})

		if event.Type != ActivityInteraction {
			t.Errorf("Expected ActivityInteraction, got %v", event.Type)
		}
		if event.PeerID != "peer1" {
			t.Errorf("Expected peer1, got %s", event.PeerID)
		}
		if event.CharacterName != "TestChar" {
			t.Errorf("Expected TestChar, got %s", event.CharacterName)
		}
		if event.Description != "TestChar performed clicked" {
			t.Errorf("Expected 'TestChar performed clicked', got '%s'", event.Description)
		}
	})

	// Test CreatePeerJoinedEvent
	t.Run("CreatePeerJoinedEvent", func(t *testing.T) {
		event := CreatePeerJoinedEvent("peer2", "JoinedChar")

		if event.Type != ActivityJoined {
			t.Errorf("Expected ActivityJoined, got %v", event.Type)
		}
		if event.Description != "JoinedChar joined the network" {
			t.Errorf("Expected 'JoinedChar joined the network', got '%s'", event.Description)
		}
	})

	// Test CreatePeerLeftEvent
	t.Run("CreatePeerLeftEvent", func(t *testing.T) {
		event := CreatePeerLeftEvent("peer3", "LeftChar")

		if event.Type != ActivityLeft {
			t.Errorf("Expected ActivityLeft, got %v", event.Type)
		}
		if event.Description != "LeftChar left the network" {
			t.Errorf("Expected 'LeftChar left the network', got '%s'", event.Description)
		}
	})

	// Test CreateChatEvent
	t.Run("CreateChatEvent", func(t *testing.T) {
		event := CreateChatEvent("peer4", "ChatChar", "Hello world")

		if event.Type != ActivityChat {
			t.Errorf("Expected ActivityChat, got %v", event.Type)
		}
		if event.Description != "ChatChar: Hello world" {
			t.Errorf("Expected 'ChatChar: Hello world', got '%s'", event.Description)
		}
	})

	// Test CreateBattleEvent
	t.Run("CreateBattleEvent", func(t *testing.T) {
		event := CreateBattleEvent("peer5", "BattleChar", "started a battle")

		if event.Type != ActivityBattle {
			t.Errorf("Expected ActivityBattle, got %v", event.Type)
		}
		if event.Description != "BattleChar started a battle" {
			t.Errorf("Expected 'BattleChar started a battle', got '%s'", event.Description)
		}
	})
}

func TestActivityTracker_ThreadSafety(t *testing.T) {
	tracker := NewActivityTracker(100)

	// Test concurrent access
	done := make(chan bool, 2)

	// Goroutine 1: Add events
	go func() {
		for i := 0; i < 50; i++ {
			event := ActivityEvent{
				Type:          ActivityChat,
				PeerID:        fmt.Sprintf("peer%d", i),
				CharacterName: fmt.Sprintf("Char%d", i),
				Description:   fmt.Sprintf("Message %d", i),
			}
			tracker.AddEvent(event)
			time.Sleep(time.Microsecond) // Small delay to encourage race conditions
		}
		done <- true
	}()

	// Goroutine 2: Read events
	go func() {
		for i := 0; i < 50; i++ {
			_ = tracker.GetRecentEvents(10)
			_ = tracker.GetEventCount()
			time.Sleep(time.Microsecond)
		}
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// Verify final state
	finalCount := tracker.GetEventCount()
	if finalCount != 50 {
		t.Errorf("Expected 50 events after concurrent access, got %d", finalCount)
	}
}
