package character

import (
	"testing"
	"time"
)

// TestNewRandomEventManager tests creation of random event manager
func TestNewRandomEventManager(t *testing.T) {
	tests := []struct {
		name          string
		events        []RandomEventConfig
		enabled       bool
		interval      time.Duration
		expectNil     bool
		expectEnabled bool
	}{
		{
			name:          "disabled manager",
			events:        []RandomEventConfig{},
			enabled:       false,
			interval:      30 * time.Second,
			expectNil:     false,
			expectEnabled: false,
		},
		{
			name:          "no events",
			events:        []RandomEventConfig{},
			enabled:       true,
			interval:      30 * time.Second,
			expectNil:     false,
			expectEnabled: false,
		},
		{
			name: "enabled with events",
			events: []RandomEventConfig{
				{
					Name:        "test_event",
					Description: "Test event",
					Probability: 0.1,
					Effects:     map[string]float64{"hunger": -10},
					Cooldown:    60,
				},
			},
			enabled:       true,
			interval:      30 * time.Second,
			expectNil:     false,
			expectEnabled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rem := NewRandomEventManager(tt.events, tt.enabled, tt.interval)

			if rem == nil {
				t.Fatal("NewRandomEventManager returned nil")
			}

			if rem.IsEnabled() != tt.expectEnabled {
				t.Errorf("expected enabled=%v, got enabled=%v", tt.expectEnabled, rem.IsEnabled())
			}

			if rem.GetEventCount() != len(tt.events) {
				t.Errorf("expected event count=%d, got count=%d", len(tt.events), rem.GetEventCount())
			}
		})
	}
}

// TestRandomEventManagerUpdate tests the Update method
func TestRandomEventManagerUpdate(t *testing.T) {
	// Create test game state
	statConfigs := map[string]StatConfig{
		"hunger": {
			Initial:           50.0,
			Max:               100.0,
			DegradationRate:   1.0,
			CriticalThreshold: 20.0,
		},
	}
	gameState := NewGameState(statConfigs, nil)

	tests := []struct {
		name            string
		events          []RandomEventConfig
		enabled         bool
		interval        time.Duration
		elapsed         time.Duration
		expectTriggered bool
	}{
		{
			name:            "disabled manager",
			events:          []RandomEventConfig{},
			enabled:         false,
			interval:        1 * time.Second,
			elapsed:         2 * time.Second,
			expectTriggered: false,
		},
		{
			name: "high probability event",
			events: []RandomEventConfig{
				{
					Name:        "test_event",
					Probability: 1.0, // 100% chance
					Effects:     map[string]float64{"hunger": -10},
					Cooldown:    1,
				},
			},
			enabled:         true,
			interval:        1 * time.Second,
			elapsed:         2 * time.Second,
			expectTriggered: true, // Should trigger with 100% probability
		},
		{
			name: "low probability event",
			events: []RandomEventConfig{
				{
					Name:        "test_event",
					Probability: 0.0, // 0% chance
					Effects:     map[string]float64{"hunger": -10},
					Cooldown:    1,
				},
			},
			enabled:         true,
			interval:        1 * time.Second,
			elapsed:         2 * time.Second,
			expectTriggered: false, // Should never trigger with 0% probability
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rem := NewRandomEventManager(tt.events, tt.enabled, tt.interval)

			// Update with specified elapsed time
			triggeredEvent := rem.Update(tt.elapsed, gameState)

			if tt.expectTriggered && triggeredEvent == nil {
				t.Error("expected event to trigger, but got nil")
			}

			if !tt.expectTriggered && triggeredEvent != nil {
				t.Errorf("expected no event to trigger, but got: %v", triggeredEvent.Name)
			}
		})
	}
}

// TestRandomEventManagerCooldowns tests cooldown functionality
func TestRandomEventManagerCooldowns(t *testing.T) {
	events := []RandomEventConfig{
		{
			Name:        "test_event",
			Probability: 1.0, // 100% chance
			Effects:     map[string]float64{"hunger": -10},
			Cooldown:    60, // 60 second cooldown
		},
	}

	statConfigs := map[string]StatConfig{
		"hunger": {Initial: 50.0, Max: 100.0, DegradationRate: 1.0, CriticalThreshold: 20.0},
	}
	gameState := NewGameState(statConfigs, nil)

	rem := NewRandomEventManager(events, true, 1*time.Second)

	// First trigger - should work
	triggeredEvent := rem.Update(2*time.Second, gameState)
	if triggeredEvent == nil {
		t.Fatal("expected first event to trigger")
	}

	// Second trigger immediately - should be blocked by cooldown
	triggeredEvent = rem.Update(2*time.Second, gameState)
	if triggeredEvent != nil {
		t.Error("expected second event to be blocked by cooldown")
	}
}

// TestRandomEventManagerConditions tests condition checking
func TestRandomEventManagerConditions(t *testing.T) {
	tests := []struct {
		name            string
		event           RandomEventConfig
		hungerValue     float64
		expectTriggered bool
	}{
		{
			name: "condition met",
			event: RandomEventConfig{
				Name:        "low_hunger_event",
				Probability: 1.0,
				Conditions: map[string]map[string]float64{
					"hunger": {"max": 30.0}, // Trigger when hunger <= 30
				},
				Effects:  map[string]float64{"hunger": 10},
				Cooldown: 1,
			},
			hungerValue:     20.0, // Below threshold
			expectTriggered: true,
		},
		{
			name: "condition not met",
			event: RandomEventConfig{
				Name:        "low_hunger_event",
				Probability: 1.0,
				Conditions: map[string]map[string]float64{
					"hunger": {"max": 30.0}, // Trigger when hunger <= 30
				},
				Effects:  map[string]float64{"hunger": 10},
				Cooldown: 1,
			},
			hungerValue:     50.0, // Above threshold
			expectTriggered: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statConfigs := map[string]StatConfig{
				"hunger": {Initial: tt.hungerValue, Max: 100.0, DegradationRate: 1.0, CriticalThreshold: 20.0},
			}
			gameState := NewGameState(statConfigs, nil)

			rem := NewRandomEventManager([]RandomEventConfig{tt.event}, true, 1*time.Second)

			triggeredEvent := rem.Update(2*time.Second, gameState)

			if tt.expectTriggered && triggeredEvent == nil {
				t.Error("expected event to trigger based on conditions, but got nil")
			}

			if !tt.expectTriggered && triggeredEvent != nil {
				t.Errorf("expected event to be blocked by conditions, but got: %v", triggeredEvent.Name)
			}
		})
	}
}

// TestTriggeredEvent tests the TriggeredEvent struct and its methods
func TestTriggeredEvent(t *testing.T) {
	tests := []struct {
		name                string
		event               TriggeredEvent
		expectHasEffects    bool
		expectHasAnimations bool
		expectHasResponses  bool
	}{
		{
			name: "complete event",
			event: TriggeredEvent{
				Name:        "complete_event",
				Description: "A complete event",
				Effects:     map[string]float64{"hunger": 10},
				Animations:  []string{"happy", "eating"},
				Responses:   []string{"Yay!", "This is great!"},
				Duration:    30 * time.Second,
			},
			expectHasEffects:    true,
			expectHasAnimations: true,
			expectHasResponses:  true,
		},
		{
			name: "effects only",
			event: TriggeredEvent{
				Name:        "effects_only",
				Description: "Only has effects",
				Effects:     map[string]float64{"hunger": 10},
				Animations:  []string{},
				Responses:   []string{},
			},
			expectHasEffects:    true,
			expectHasAnimations: false,
			expectHasResponses:  false,
		},
		{
			name: "no effects",
			event: TriggeredEvent{
				Name:        "no_effects",
				Description: "No effects",
				Effects:     map[string]float64{},
				Animations:  []string{"idle"},
				Responses:   []string{"Hello!"},
			},
			expectHasEffects:    false,
			expectHasAnimations: true,
			expectHasResponses:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.event.HasEffects() != tt.expectHasEffects {
				t.Errorf("expected HasEffects()=%v, got %v", tt.expectHasEffects, tt.event.HasEffects())
			}

			if tt.event.HasAnimations() != tt.expectHasAnimations {
				t.Errorf("expected HasAnimations()=%v, got %v", tt.expectHasAnimations, tt.event.HasAnimations())
			}

			if tt.event.HasResponses() != tt.expectHasResponses {
				t.Errorf("expected HasResponses()=%v, got %v", tt.expectHasResponses, tt.event.HasResponses())
			}

			// Test random selection methods
			if tt.event.HasAnimations() {
				animation := tt.event.GetRandomAnimation()
				if animation == "" {
					t.Error("expected non-empty animation from GetRandomAnimation")
				}
				// Verify it's from the list
				found := false
				for _, a := range tt.event.Animations {
					if a == animation {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("GetRandomAnimation returned '%s' which is not in the animations list", animation)
				}
			} else {
				animation := tt.event.GetRandomAnimation()
				if animation != "" {
					t.Error("expected empty animation from GetRandomAnimation when no animations")
				}
			}

			if tt.event.HasResponses() {
				response := tt.event.GetRandomResponse()
				if response == "" {
					t.Error("expected non-empty response from GetRandomResponse")
				}
				// Verify it's from the list
				found := false
				for _, r := range tt.event.Responses {
					if r == response {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("GetRandomResponse returned '%s' which is not in the responses list", response)
				}
			} else {
				response := tt.event.GetRandomResponse()
				if response != "" {
					t.Error("expected empty response from GetRandomResponse when no responses")
				}
			}
		})
	}
}

// TestRandomEventManagerThreadSafety tests concurrent access
func TestRandomEventManagerThreadSafety(t *testing.T) {
	events := []RandomEventConfig{
		{
			Name:        "thread_test_event",
			Probability: 0.5,
			Effects:     map[string]float64{"hunger": -5},
			Cooldown:    1,
		},
	}

	statConfigs := map[string]StatConfig{
		"hunger": {Initial: 50.0, Max: 100.0, DegradationRate: 1.0, CriticalThreshold: 20.0},
	}
	gameState := NewGameState(statConfigs, nil)

	rem := NewRandomEventManager(events, true, 1*time.Second)

	// Test concurrent access
	done := make(chan struct{})
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer func() { done <- struct{}{} }()

			// Multiple operations
			rem.IsEnabled()
			rem.GetEventCount()
			rem.GetLastCheckTime()
			rem.Update(2*time.Second, gameState)
			rem.SetEnabled(true)
			rem.GetRandomResponse([]string{"test"})
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// No panic or race condition means success
}

// TestRandomEventManagerEdgeCases tests edge cases and error conditions
func TestRandomEventManagerEdgeCases(t *testing.T) {
	t.Run("nil game state", func(t *testing.T) {
		events := []RandomEventConfig{
			{Name: "test", Probability: 1.0, Effects: map[string]float64{"hunger": 10}},
		}
		rem := NewRandomEventManager(events, true, 1*time.Second)

		triggeredEvent := rem.Update(2*time.Second, nil)
		if triggeredEvent != nil {
			t.Error("expected nil when game state is nil")
		}
	})

	t.Run("empty responses", func(t *testing.T) {
		rem := NewRandomEventManager([]RandomEventConfig{}, true, 1*time.Second)

		response := rem.GetRandomResponse([]string{})
		if response != "" {
			t.Error("expected empty response for empty responses list")
		}
	})

	t.Run("disabled manager operations", func(t *testing.T) {
		rem := NewRandomEventManager([]RandomEventConfig{}, false, 1*time.Second)

		response := rem.GetRandomResponse([]string{"test"})
		if response != "" {
			t.Error("expected empty response when manager is disabled")
		}
	})

	t.Run("toggle enable/disable", func(t *testing.T) {
		// Create manager with events so it can be enabled
		events := []RandomEventConfig{
			{Name: "test", Probability: 0.1, Effects: map[string]float64{"hunger": 10}},
		}
		rem := NewRandomEventManager(events, true, 1*time.Second)

		if !rem.IsEnabled() {
			t.Error("expected manager to be enabled initially")
		}

		rem.SetEnabled(false)
		if rem.IsEnabled() {
			t.Error("expected manager to be disabled after SetEnabled(false)")
		}

		rem.SetEnabled(true)
		if !rem.IsEnabled() {
			t.Error("expected manager to be enabled after SetEnabled(true)")
		}
	})
}
