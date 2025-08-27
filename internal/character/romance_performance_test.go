package character

import (
	"fmt"
	"runtime"
	"testing"
	"time"
)

// BenchmarkRomanceInteractionHandling benchmarks romance interaction processing performance
func BenchmarkRomanceInteractionHandling(b *testing.B) {
	// Create romance character for testing
	romanceCard := createTestRomanceCharacter()

	// Create character instance
	character, err := New(&romanceCard, ".")
	if err != nil {
		b.Fatalf("Failed to create character: %v", err)
	}

	// Pre-populate some interaction history for realistic testing
	if character.gameState != nil {
		character.gameState.InteractionHistory = map[string][]time.Time{
			"compliment": make([]time.Time, 10),
			"gift":       make([]time.Time, 5),
		}

		// Pre-populate romance memories
		for i := 0; i < 20; i++ {
			character.gameState.RomanceMemories = append(character.gameState.RomanceMemories, RomanceMemory{
				Timestamp:       time.Now().Add(-time.Duration(i) * time.Minute),
				InteractionType: "compliment",
				Response:        "Thank you!",
			})
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Test different romance interactions
		interaction := []string{"compliment", "gift", "deep_conversation"}[i%3]
		character.HandleRomanceInteraction(interaction)
	}
}

// BenchmarkRomanceDialogSelection benchmarks dialog selection performance
func BenchmarkRomanceDialogSelection(b *testing.B) {
	romanceCard := createTestRomanceCharacter()
	character, err := New(&romanceCard, ".")
	if err != nil {
		b.Fatalf("Failed to create character: %v", err)
	}

	// Add many romance dialogs to stress test selection algorithm
	for i := 0; i < 50; i++ {
		romanceCard.RomanceDialogs = append(romanceCard.RomanceDialogs, DialogExtended{
			Dialog: Dialog{
				Trigger:   "click",
				Responses: []string{fmt.Sprintf("Romance response %d", i)},
				Animation: "romantic",
			},
			Requirements: &RomanceRequirement{
				Stats: map[string]map[string]float64{
					"affection": {"min": float64(i * 2)},
				},
			},
		})
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		character.selectRomanceDialog("click")
	}
}

// BenchmarkMemorySystemPerformance benchmarks romance memory management
func BenchmarkMemorySystemPerformance(b *testing.B) {
	romanceCard := createTestRomanceCharacter()
	character, err := New(&romanceCard, ".")
	if err != nil {
		b.Fatalf("Failed to create character: %v", err)
	}

	if character.gameState == nil {
		b.Skip("Character does not have game state enabled")
	}

	// Fill memory to near capacity (testing memory limit handling)
	for i := 0; i < 45; i++ {
		character.gameState.RecordRomanceInteraction(
			"compliment",
			"Test response",
			map[string]float64{"affection": 10},
			map[string]float64{"affection": 15},
		)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		character.gameState.RecordRomanceInteraction(
			"gift",
			"Gift response",
			map[string]float64{"affection": 20},
			map[string]float64{"affection": 30},
		)
	}
}

// BenchmarkRelationshipLevelCalculation benchmarks relationship progression performance
func BenchmarkRelationshipLevelCalculation(b *testing.B) {
	romanceCard := createTestRomanceCharacter()
	character, err := New(&romanceCard, ".")
	if err != nil {
		b.Fatalf("Failed to create character: %v", err)
	}

	if character.gameState == nil {
		b.Skip("Character does not have game state enabled")
	}

	// Set stats to trigger level calculations
	character.gameState.Stats["affection"].Current = 75
	character.gameState.Stats["trust"].Current = 60
	character.gameState.Stats["intimacy"].Current = 40

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		character.gameState.UpdateRelationshipLevel(romanceCard.Progression)
	}
}

// BenchmarkRomanceEventProcessing benchmarks romance events system performance
func BenchmarkRomanceEventProcessing(b *testing.B) {
	romanceCard := createTestRomanceCharacter()
	character, err := New(&romanceCard, ".")
	if err != nil {
		b.Fatalf("Failed to create character: %v", err)
	}

	if character.romanceEventManager == nil {
		b.Skip("Character does not have romance event manager")
	}

	// Add multiple romance events for testing
	character.romanceEventManager.events = []RandomEventConfig{
		{
			Name:        "Sweet Memory",
			Probability: 0.1,
			Effects:     map[string]float64{"affection": 2},
			Responses:   []string{"I was thinking of you..."},
		},
		{
			Name:        "Romantic Moment",
			Probability: 0.05,
			Effects:     map[string]float64{"intimacy": 3},
			Responses:   []string{"This moment is perfect..."},
		},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		character.processRomanceEvents(time.Second)
	}
}

// BenchmarkPersonalityModifierCalculation benchmarks personality trait processing
func BenchmarkPersonalityModifierCalculation(b *testing.B) {
	romanceCard := createTestRomanceCharacter()
	character, err := New(&romanceCard, ".")
	if err != nil {
		b.Fatalf("Failed to create character: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		character.calculatePersonalityModifier("compliment")
		character.calculatePersonalityModifier("gift")
		character.calculatePersonalityModifier("romantic_gesture")
	}
}

// TestRomanceMemoryUsage tests memory usage stays within bounds during heavy romance usage
func TestRomanceMemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory usage test in short mode")
	}

	// Record initial memory
	var initialMem runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&initialMem)

	// Create character with heavy romance configuration
	romanceCard := createTestRomanceCharacter()
	character, err := New(&romanceCard, ".")
	if err != nil {
		t.Fatalf("Failed to create character: %v", err)
	}

	if character.gameState == nil {
		t.Skip("Character does not have game state enabled")
	}

	// Simulate heavy romance interaction usage
	for i := 0; i < 1000; i++ {
		// Rotate through different interaction types
		interactions := []string{"compliment", "gift", "deep_conversation", "romantic_gesture"}
		interaction := interactions[i%len(interactions)]

		character.HandleRomanceInteraction(interaction)

		// Force some events to trigger
		if i%10 == 0 && character.romanceEventManager != nil {
			character.processRomanceEvents(time.Second)
		}

		// Update relationship levels periodically
		if i%25 == 0 {
			character.gameState.UpdateRelationshipLevel(romanceCard.Progression)
		}
	}

	// Record final memory usage
	var finalMem runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&finalMem)

	// Calculate memory increase
	memoryIncreaseMB := float64(finalMem.Alloc-initialMem.Alloc) / 1024 / 1024

	// Memory increase should be reasonable (< 10MB for heavy usage)
	if memoryIncreaseMB > 10.0 {
		t.Errorf("Memory usage increased by %.2fMB during heavy romance interactions, expected < 10MB", memoryIncreaseMB)
	}

	t.Logf("Memory usage after 1000 romance interactions: %.2fMB increase", memoryIncreaseMB)

	// Verify memory limits are respected (romance memories should be capped at 50)
	if len(character.gameState.RomanceMemories) > 50 {
		t.Errorf("Romance memories not properly limited: %d entries (expected ≤ 50)", len(character.gameState.RomanceMemories))
	}
}

// TestRomanceConcurrencyStress tests romance features under concurrent access stress
func TestRomanceConcurrencyStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrency stress test in short mode")
	}

	romanceCard := createTestRomanceCharacter()
	character, err := New(&romanceCard, ".")
	if err != nil {
		t.Fatalf("Failed to create character: %v", err)
	}

	if character.gameState == nil {
		t.Skip("Character does not have game state enabled")
	}

	// Number of concurrent operations
	numGoroutines := 10
	operationsPerGoroutine := 100

	// Use channel to coordinate completion
	done := make(chan bool, numGoroutines)

	// Launch concurrent romance interactions
	for i := 0; i < numGoroutines; i++ {
		go func(routineID int) {
			defer func() { done <- true }()

			for j := 0; j < operationsPerGoroutine; j++ {
				// Simulate different concurrent operations
				switch j % 4 {
				case 0:
					character.HandleRomanceInteraction("compliment")
				case 1:
					character.gameState.GetRelationshipLevel()
				case 2:
					character.gameState.UpdateRelationshipLevel(romanceCard.Progression)
				case 3:
					if character.romanceEventManager != nil {
						character.processRomanceEvents(time.Second)
					}
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete with timeout
	timeout := time.After(30 * time.Second)
	completed := 0

	for completed < numGoroutines {
		select {
		case <-done:
			completed++
		case <-timeout:
			t.Fatal("Concurrency stress test timed out - possible deadlock")
		}
	}

	t.Logf("Successfully completed %d concurrent operations across %d goroutines",
		numGoroutines*operationsPerGoroutine, numGoroutines)
}

// TestRomanceFeatureLoadTesting simulates extended usage patterns
func TestRomanceFeatureLoadTesting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	romanceCard := createTestRomanceCharacter()
	character, err := New(&romanceCard, ".")
	if err != nil {
		t.Fatalf("Failed to create character: %v", err)
	}

	if character.gameState == nil {
		t.Skip("Character does not have game state enabled")
	}

	// Simulate 1 hour of continuous usage
	startTime := time.Now()
	iterations := 3600 // 1 per second for 1 hour

	for i := 0; i < iterations; i++ {
		// Simulate realistic usage patterns
		switch i % 10 {
		case 0, 1, 2: // 30% compliments
			character.HandleRomanceInteraction("compliment")
		case 3, 4: // 20% gifts
			character.HandleRomanceInteraction("gift")
		case 5: // 10% deep conversations
			character.HandleRomanceInteraction("deep_conversation")
		case 6: // 10% romantic gestures
			character.HandleRomanceInteraction("romantic_gesture")
		case 7, 8: // 20% just checking dialog
			character.selectRomanceDialog("click")
		case 9: // 10% level checks
			character.gameState.UpdateRelationshipLevel(romanceCard.Progression)
		}

		// Process events periodically
		if i%30 == 0 && character.romanceEventManager != nil {
			character.processRomanceEvents(time.Second)
		}

		// Simulate time passage (fast-forward)
		if i%100 == 0 {
			// Force stat degradation simulation
			for statName := range character.gameState.Stats {
				if character.gameState.Stats[statName].Current > 10 {
					character.gameState.Stats[statName].Current -= 0.1
				}
			}
		}
	}

	duration := time.Since(startTime)
	t.Logf("Load test completed %d operations in %v (%.2f ops/sec)",
		iterations, duration, float64(iterations)/duration.Seconds())

	// Verify system state remains healthy
	stats := character.gameState.GetStats()
	if len(stats) == 0 {
		t.Error("Character stats were lost during load testing")
	}

	if character.gameState.GetRelationshipLevel() == "" {
		t.Error("Relationship level was lost during load testing")
	}

	if len(character.gameState.InteractionHistory) == 0 {
		t.Error("Interaction history was lost during load testing")
	}
}

// createTestRomanceCharacter creates a comprehensive test character for performance testing
func createTestRomanceCharacter() CharacterCard {
	return CharacterCard{
		Name:        "Performance Test Romance",
		Description: "Character for performance testing",
		Animations: map[string]string{
			"idle":            "idle.gif",
			"talking":         "talking.gif",
			"romantic":        "romantic.gif",
			"blushing":        "blushing.gif",
			"heart_eyes":      "heart_eyes.gif",
			"excited_romance": "excited_romance.gif",
		},
		Dialogs: []Dialog{
			{Trigger: "click", Responses: []string{"Hello!"}, Animation: "talking"},
		},
		Behavior: Behavior{IdleTimeout: 30, DefaultSize: 128},
		Stats: map[string]StatConfig{
			"affection": {Initial: 0, Max: 100, DegradationRate: 0.1, CriticalThreshold: 10},
			"trust":     {Initial: 20, Max: 100, DegradationRate: 0.05, CriticalThreshold: 5},
			"intimacy":  {Initial: 0, Max: 100, DegradationRate: 0.2, CriticalThreshold: 0},
			"jealousy":  {Initial: 0, Max: 100, DegradationRate: 2.0, CriticalThreshold: 80},
		},
		Interactions: map[string]InteractionConfig{
			"compliment": {
				Triggers:  []string{"shift+click"},
				Effects:   map[string]float64{"affection": 5, "trust": 1},
				Responses: []string{"Thank you! 💕"},
				Cooldown:  45,
			},
			"gift": {
				Triggers:  []string{"ctrl+click"},
				Effects:   map[string]float64{"affection": 10, "happiness": 5},
				Responses: []string{"A gift! 🎁"},
				Cooldown:  120,
			},
			"deep_conversation": {
				Triggers:  []string{"alt+click"},
				Effects:   map[string]float64{"trust": 8, "intimacy": 3},
				Responses: []string{"I love our talks..."},
				Cooldown:  90,
			},
			"romantic_gesture": {
				Triggers:  []string{"ctrl+alt+click"},
				Effects:   map[string]float64{"affection": 15, "intimacy": 10},
				Responses: []string{"*melts* 💖"},
				Cooldown:  180,
			},
		},
		Personality: &PersonalityConfig{
			Traits: map[string]float64{
				"shyness":                  0.6,
				"romanticism":              0.8,
				"jealousy_prone":           0.3,
				"trust_difficulty":         0.4,
				"affection_responsiveness": 0.9,
			},
			Compatibility: map[string]float64{
				"gift_appreciation":  1.5,
				"conversation_lover": 1.3,
			},
		},
		Progression: &ProgressionConfig{
			Levels: []LevelConfig{
				{
					Name:        "Stranger",
					Requirement: map[string]int64{"affection": 0},
					Size:        128,
				},
				{
					Name:        "Friend",
					Requirement: map[string]int64{"affection": 15, "trust": 10},
					Size:        132,
				},
				{
					Name:        "Close Friend",
					Requirement: map[string]int64{"affection": 30, "trust": 25},
					Size:        136,
				},
				{
					Name:        "Romantic Interest",
					Requirement: map[string]int64{"affection": 50, "trust": 40, "intimacy": 20},
					Size:        140,
				},
			},
		},
		RomanceDialogs: []DialogExtended{
			{
				Dialog: Dialog{
					Trigger:   "click",
					Responses: []string{"Hi sweetheart! 💕", "I was hoping you'd visit!"},
					Animation: "romantic",
				},
				Requirements: &RomanceRequirement{
					Stats: map[string]map[string]float64{
						"affection": {"min": 30},
					},
				},
			},
		},
		RomanceEvents: []RandomEventConfig{
			{
				Name:        "Sweet Memory",
				Probability: 0.1,
				Effects:     map[string]float64{"affection": 2},
				Responses:   []string{"I was thinking about you... 💭"},
			},
			{
				Name:        "Romantic Daydream",
				Probability: 0.05,
				Effects:     map[string]float64{"intimacy": 3},
				Responses:   []string{"*sighs dreamily* 💖"},
			},
		},
	}
}
