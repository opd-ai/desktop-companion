package character

import (
	"math"
	"testing"
)

// TestNewCompatibilityCalculator tests the creation of a compatibility calculator
func TestNewCompatibilityCalculator(t *testing.T) {
	t.Run("ValidCharacterWithPersonality", func(t *testing.T) {
		// Create character with personality
		card := &CharacterCard{
			Name: "TestChar",
			Personality: &PersonalityConfig{
				Traits: map[string]float64{
					"kindness":    0.7,
					"confidence":  0.8,
					"playfulness": 0.6,
				},
			},
		}

		char, err := New(card, "test_data")
		if err != nil {
			t.Fatalf("Failed to create character: %v", err)
		}

		calculator := NewCompatibilityCalculator(char)
		if calculator == nil {
			t.Fatal("Expected non-nil compatibility calculator")
		}

		if calculator.localPersonality == nil {
			t.Error("Expected local personality to be set")
		}

		if calculator.localPersonality.Traits == nil {
			t.Error("Expected personality traits to be set")
		}
	})

	t.Run("NilCharacter", func(t *testing.T) {
		calculator := NewCompatibilityCalculator(nil)
		if calculator == nil {
			t.Fatal("Expected non-nil compatibility calculator even for nil character")
		}

		if calculator.localPersonality != nil {
			t.Error("Expected nil local personality for nil character")
		}
	})

	t.Run("CharacterWithoutPersonality", func(t *testing.T) {
		card := &CharacterCard{
			Name: "TestChar",
			// No personality config
		}

		char, err := New(card, "test_data")
		if err != nil {
			t.Fatalf("Failed to create character: %v", err)
		}

		calculator := NewCompatibilityCalculator(char)
		if calculator == nil {
			t.Fatal("Expected non-nil compatibility calculator")
		}

		if calculator.localPersonality != nil {
			t.Error("Expected nil local personality for character without personality")
		}
	})
}

// TestCalculateCompatibility tests the core compatibility calculation logic
func TestCalculateCompatibility(t *testing.T) {
	// Create test calculator with known personality
	localPersonality := &PersonalityConfig{
		Traits: map[string]float64{
			"kindness":     0.7,
			"confidence":   0.8,
			"playfulness":  0.6,
			"intelligence": 0.9,
		},
	}

	calculator := &CompatibilityCalculator{
		localPersonality: localPersonality,
	}

	t.Run("PerfectMatch", func(t *testing.T) {
		// Identical personality should give perfect compatibility
		peerPersonality := &PersonalityConfig{
			Traits: map[string]float64{
				"kindness":     0.7,
				"confidence":   0.8,
				"playfulness":  0.6,
				"intelligence": 0.9,
			},
		}

		score := calculator.CalculateCompatibility(peerPersonality)
		if score != 1.0 {
			t.Errorf("Expected perfect compatibility (1.0), got %f", score)
		}
	})

	t.Run("CompleteOpposite", func(t *testing.T) {
		// Completely opposite personality should give zero compatibility
		peerPersonality := &PersonalityConfig{
			Traits: map[string]float64{
				"kindness":     0.0, // local: 0.7, difference: 0.7, score: 0.3
				"confidence":   0.0, // local: 0.8, difference: 0.8, score: 0.2
				"playfulness":  0.0, // local: 0.6, difference: 0.6, score: 0.4
				"intelligence": 0.0, // local: 0.9, difference: 0.9, score: 0.1
			},
		}

		score := calculator.CalculateCompatibility(peerPersonality)
		expectedScore := (0.3 + 0.2 + 0.4 + 0.1) / 4.0 // Average: 0.25

		// Use tolerance for floating point comparison
		tolerance := 0.000001
		if math.Abs(score-expectedScore) > tolerance {
			t.Errorf("Expected compatibility score %f, got %f", expectedScore, score)
		}
	})

	t.Run("PartialMatch", func(t *testing.T) {
		// Some matching traits, some different
		peerPersonality := &PersonalityConfig{
			Traits: map[string]float64{
				"kindness":   0.7, // Perfect match
				"confidence": 0.5, // Difference: 0.3, score: 0.7
				"creativity": 0.8, // Not in local traits, ignored
			},
		}

		score := calculator.CalculateCompatibility(peerPersonality)
		expectedScore := (1.0 + 0.7) / 2.0 // Average: 0.85

		// Use tolerance for floating point comparison
		tolerance := 0.000001
		if math.Abs(score-expectedScore) > tolerance {
			t.Errorf("Expected compatibility score %f, got %f", expectedScore, score)
		}
	})

	t.Run("NoMatchingTraits", func(t *testing.T) {
		// No overlapping traits should return neutral compatibility
		peerPersonality := &PersonalityConfig{
			Traits: map[string]float64{
				"creativity":  0.8,
				"persistence": 0.6,
				"humor":       0.9,
			},
		}

		score := calculator.CalculateCompatibility(peerPersonality)
		if score != 0.5 {
			t.Errorf("Expected neutral compatibility (0.5) for no matching traits, got %f", score)
		}
	})

	t.Run("NilPeerPersonality", func(t *testing.T) {
		score := calculator.CalculateCompatibility(nil)
		if score != 0.5 {
			t.Errorf("Expected neutral compatibility (0.5) for nil peer personality, got %f", score)
		}
	})

	t.Run("NilPeerTraits", func(t *testing.T) {
		peerPersonality := &PersonalityConfig{
			Traits: nil,
		}

		score := calculator.CalculateCompatibility(peerPersonality)
		if score != 0.5 {
			t.Errorf("Expected neutral compatibility (0.5) for nil peer traits, got %f", score)
		}
	})
}

// TestCalculateCompatibilityNilLocalPersonality tests behavior with nil local personality
func TestCalculateCompatibilityNilLocalPersonality(t *testing.T) {
	calculator := &CompatibilityCalculator{
		localPersonality: nil,
	}

	peerPersonality := &PersonalityConfig{
		Traits: map[string]float64{
			"kindness":   0.7,
			"confidence": 0.8,
		},
	}

	score := calculator.CalculateCompatibility(peerPersonality)
	if score != 0.5 {
		t.Errorf("Expected neutral compatibility (0.5) for nil local personality, got %f", score)
	}
}

// TestGetCompatibilityCategory tests the human-readable categorization of scores
func TestGetCompatibilityCategory(t *testing.T) {
	calculator := &CompatibilityCalculator{}

	testCases := []struct {
		score    float64
		expected string
	}{
		{1.0, "Excellent"},
		{0.95, "Excellent"},
		{0.9, "Excellent"},
		{0.85, "Very Good"},
		{0.8, "Very Good"},
		{0.7, "Good"},
		{0.6, "Good"},
		{0.5, "Fair"},
		{0.4, "Fair"},
		{0.3, "Poor"},
		{0.2, "Poor"},
		{0.1, "Very Poor"},
		{0.0, "Very Poor"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			category := calculator.GetCompatibilityCategory(tc.score)
			if category != tc.expected {
				t.Errorf("Score %f: expected category %s, got %s", tc.score, tc.expected, category)
			}
		})
	}
}

// TestCompatibilityCalculatorEdgeCases tests edge cases and error conditions
func TestCompatibilityCalculatorEdgeCases(t *testing.T) {
	t.Run("EmptyTraits", func(t *testing.T) {
		localPersonality := &PersonalityConfig{
			Traits: map[string]float64{}, // Empty traits
		}

		calculator := &CompatibilityCalculator{
			localPersonality: localPersonality,
		}

		peerPersonality := &PersonalityConfig{
			Traits: map[string]float64{
				"kindness": 0.7,
			},
		}

		score := calculator.CalculateCompatibility(peerPersonality)
		if score != 0.5 {
			t.Errorf("Expected neutral compatibility (0.5) for empty local traits, got %f", score)
		}
	})

	t.Run("ExtremeValues", func(t *testing.T) {
		localPersonality := &PersonalityConfig{
			Traits: map[string]float64{
				"trait1": 2.0,  // Above normal range [0,1]
				"trait2": -0.5, // Below normal range [0,1]
			},
		}

		calculator := &CompatibilityCalculator{
			localPersonality: localPersonality,
		}

		peerPersonality := &PersonalityConfig{
			Traits: map[string]float64{
				"trait1": 1.0,
				"trait2": 0.0,
			},
		}

		score := calculator.CalculateCompatibility(peerPersonality)

		// Should handle extreme values gracefully
		// trait1: difference = |2.0 - 1.0| = 1.0, score = 1.0 - 1.0 = 0.0
		// trait2: difference = |-0.5 - 0.0| = 0.5, score = 1.0 - 0.5 = 0.5
		// Average: (0.0 + 0.5) / 2 = 0.25
		expectedScore := 0.25

		// Use tolerance for floating point comparison
		tolerance := 0.000001
		if math.Abs(score-expectedScore) > tolerance {
			t.Errorf("Expected compatibility score %f for extreme values, got %f", expectedScore, score)
		}
	})

	t.Run("SingleTrait", func(t *testing.T) {
		localPersonality := &PersonalityConfig{
			Traits: map[string]float64{
				"kindness": 0.8,
			},
		}

		calculator := &CompatibilityCalculator{
			localPersonality: localPersonality,
		}

		peerPersonality := &PersonalityConfig{
			Traits: map[string]float64{
				"kindness": 0.6,
			},
		}

		score := calculator.CalculateCompatibility(peerPersonality)
		expectedScore := 1.0 - 0.2 // difference = 0.2, score = 0.8

		// Use tolerance for floating point comparison
		tolerance := 0.000001
		if math.Abs(score-expectedScore) > tolerance {
			t.Errorf("Expected compatibility score %f for single trait, got %f", expectedScore, score)
		}
	})
}

// TestCompatibilityCalculatorConcurrency tests thread safety
func TestCompatibilityCalculatorConcurrency(t *testing.T) {
	localPersonality := &PersonalityConfig{
		Traits: map[string]float64{
			"kindness":   0.7,
			"confidence": 0.8,
		},
	}

	calculator := &CompatibilityCalculator{
		localPersonality: localPersonality,
	}

	peerPersonality := &PersonalityConfig{
		Traits: map[string]float64{
			"kindness":   0.6,
			"confidence": 0.9,
		},
	}

	// Run multiple goroutines to test concurrent access
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			// Should not panic or race
			score := calculator.CalculateCompatibility(peerPersonality)
			if score < 0.0 || score > 1.0 {
				t.Errorf("Invalid compatibility score: %f", score)
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
