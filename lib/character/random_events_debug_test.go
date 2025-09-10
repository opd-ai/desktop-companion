package character

import (
	"testing"
	"time"
)

// TestRandomEventLogic - test the basic logic without randomness
func TestRandomEventLogic(t *testing.T) {
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

	// Test canTriggerEvent directly
	events := []RandomEventConfig{
		{
			Name:        "test_event",
			Description: "Test event",
			Probability: 1.0,
			Effects:     map[string]float64{"hunger": -10},
			Cooldown:    1,
		},
	}

	rem := NewRandomEventManager(events, true, 1*time.Second)

	// Test if the event can trigger
	now := time.Now()
	canTrigger := rem.canTriggerEvent(events[0], now, gameState)
	t.Logf("Can trigger event: %v", canTrigger)

	if !canTrigger {
		t.Error("Event should be able to trigger")
		return
	}

	// Test the full Update method with a small check interval
	rem2 := NewRandomEventManager(events, true, 100*time.Millisecond)

	// Update with sufficient elapsed time
	triggeredEvent := rem2.Update(200*time.Millisecond, gameState)

	t.Logf("Triggered event: %v", triggeredEvent)

	// With probability 1.0, it should eventually trigger
	// Let's try a few times to account for random variation
	attempts := 0
	maxAttempts := 20
	for triggeredEvent == nil && attempts < maxAttempts {
		time.Sleep(10 * time.Millisecond)
		rem2 = NewRandomEventManager(events, true, 100*time.Millisecond)
		triggeredEvent = rem2.Update(200*time.Millisecond, gameState)
		attempts++
	}

	if triggeredEvent == nil {
		t.Errorf("Expected event to trigger within %d attempts", maxAttempts)
	} else {
		t.Logf("Event triggered successfully after %d attempts: %s", attempts+1, triggeredEvent.Name)
	}
}
