package ui

import (
	"testing"

	"github.com/opd-ai/desktop-companion/internal/character"
	"github.com/opd-ai/desktop-companion/internal/network"
)

// MockNetworkManagerForCompatibility provides a mock network manager for compatibility testing
type MockNetworkManagerForCompatibility struct {
	peers []network.Peer
}

func (m *MockNetworkManagerForCompatibility) GetPeerCount() int {
	return len(m.peers)
}

func (m *MockNetworkManagerForCompatibility) GetPeers() []network.Peer {
	return m.peers
}

func (m *MockNetworkManagerForCompatibility) GetNetworkID() string {
	return "test-network"
}

func (m *MockNetworkManagerForCompatibility) SendMessage(msgType network.MessageType, payload []byte, targetPeerID string) error {
	return nil
}

func (m *MockNetworkManagerForCompatibility) RegisterMessageHandler(msgType network.MessageType, handler network.MessageHandler) {
}

// Helper function to create test characters with different personalities
func createTestCharacterWithPersonality(t *testing.T, name string, traits map[string]float64) *character.Character {
	t.Helper()

	card := &character.CharacterCard{
		Name:        name,
		Description: "Test character for compatibility testing",
		Personality: &character.PersonalityConfig{
			Traits: traits,
		},
		Animations: map[string]string{
			"idle": "test.gif",
		},
	}

	char, err := character.New(card, "test_data")
	if err != nil {
		t.Fatalf("Failed to create character %s: %v", name, err)
	}

	return char
}

// TestFeature5_SetCompatibilityCalculator tests setting the compatibility calculator
func TestFeature5_SetCompatibilityCalculator(t *testing.T) {
	t.Run("SetCalculatorSuccessfully", func(t *testing.T) {
		mockNetworkManager := &MockNetworkManagerForCompatibility{}
		overlay := NewNetworkOverlay(mockNetworkManager)

		// Create character with personality
		char := createTestCharacterWithPersonality(t, "TestChar", map[string]float64{
			"kindness":   0.7,
			"confidence": 0.8,
		})

		calculator := character.NewCompatibilityCalculator(char)
		overlay.SetCompatibilityCalculator(calculator)

		// Verify calculator was set (internal state, verified through behavior)
		// We'll test this by checking if compatibility scores can be calculated
		if overlay.compatibilityCalculator == nil {
			t.Error("Expected compatibility calculator to be set")
		}
	})

	t.Run("SetNilCalculator", func(t *testing.T) {
		mockNetworkManager := &MockNetworkManagerForCompatibility{}
		overlay := NewNetworkOverlay(mockNetworkManager)

		// Should handle nil calculator gracefully
		overlay.SetCompatibilityCalculator(nil)

		// Should not panic and should handle gracefully
		score := overlay.GetCompatibilityScore("test-peer")
		if score != 0.5 {
			t.Errorf("Expected neutral score (0.5) when no calculator set, got %f", score)
		}
	})
}

// TestFeature5_UpdateCompatibilityScores tests compatibility score calculation
func TestFeature5_UpdateCompatibilityScores(t *testing.T) {
	t.Run("UpdateWithValidCharacters", func(t *testing.T) {
		mockNetworkManager := &MockNetworkManagerForCompatibility{
			peers: []network.Peer{
				{ID: "peer1", Conn: nil}, // Use nil for mock connection
				{ID: "peer2", Conn: nil}, // Use nil for mock connection
			},
		}
		overlay := NewNetworkOverlay(mockNetworkManager)

		// Create local character with personality
		localChar := createTestCharacterWithPersonality(t, "LocalChar", map[string]float64{
			"kindness":   0.7,
			"confidence": 0.8,
			"humor":      0.6,
		})

		calculator := character.NewCompatibilityCalculator(localChar)
		overlay.SetCompatibilityCalculator(calculator)

		// Add network characters with personalities manually for testing
		overlay.characterMutex.Lock()
		overlay.characters = []CharacterInfo{
			{
				Name:     "LocalChar",
				Location: "Local",
				IsLocal:  true,
				IsActive: true,
				PeerID:   "",
			},
			{
				Name:     "Peer1's Character",
				Location: "peer1",
				IsLocal:  false,
				IsActive: true,
				PeerID:   "peer1",
				Personality: &character.PersonalityConfig{
					Traits: map[string]float64{
						"kindness":   0.8, // Close match
						"confidence": 0.7, // Close match
						"humor":      0.9, // Close match
					},
				},
			},
			{
				Name:     "Peer2's Character",
				Location: "peer2",
				IsLocal:  false,
				IsActive: true,
				PeerID:   "peer2",
				Personality: &character.PersonalityConfig{
					Traits: map[string]float64{
						"kindness":   0.1, // Very different
						"confidence": 0.2, // Very different
						"humor":      0.0, // Very different
					},
				},
			},
		}
		overlay.characterMutex.Unlock()

		// Update compatibility scores
		overlay.UpdateCompatibilityScores()

		// Check peer1 compatibility (should be high)
		peer1Score := overlay.GetCompatibilityScore("peer1")
		if peer1Score < 0.7 {
			t.Errorf("Expected high compatibility score for peer1 (>0.7), got %f", peer1Score)
		}

		// Check peer2 compatibility (should be low)
		peer2Score := overlay.GetCompatibilityScore("peer2")
		if peer2Score > 0.5 {
			t.Errorf("Expected low compatibility score for peer2 (<0.5), got %f", peer2Score)
		}
	})

	t.Run("UpdateWithoutCalculator", func(t *testing.T) {
		mockNetworkManager := &MockNetworkManagerForCompatibility{}
		overlay := NewNetworkOverlay(mockNetworkManager)

		// Don't set calculator - should handle gracefully
		overlay.UpdateCompatibilityScores()

		// Should return neutral scores
		score := overlay.GetCompatibilityScore("any-peer")
		if score != 0.5 {
			t.Errorf("Expected neutral score (0.5) without calculator, got %f", score)
		}
	})
}

// TestFeature5_GetCompatibilityScore tests score retrieval
func TestFeature5_GetCompatibilityScore(t *testing.T) {
	t.Run("GetExistingScore", func(t *testing.T) {
		mockNetworkManager := &MockNetworkManagerForCompatibility{}
		overlay := NewNetworkOverlay(mockNetworkManager)

		// Manually set a compatibility score
		overlay.compatibilityMutex.Lock()
		overlay.compatibilityScores["test-peer"] = 0.85
		overlay.compatibilityMutex.Unlock()

		score := overlay.GetCompatibilityScore("test-peer")
		if score != 0.85 {
			t.Errorf("Expected score 0.85, got %f", score)
		}
	})

	t.Run("GetNonExistentScore", func(t *testing.T) {
		mockNetworkManager := &MockNetworkManagerForCompatibility{}
		overlay := NewNetworkOverlay(mockNetworkManager)

		score := overlay.GetCompatibilityScore("unknown-peer")
		if score != 0.5 {
			t.Errorf("Expected neutral score (0.5) for unknown peer, got %f", score)
		}
	})
}

// TestFeature5_GetCompatibilityCategory tests category determination
func TestFeature5_GetCompatibilityCategory(t *testing.T) {
	mockNetworkManager := &MockNetworkManagerForCompatibility{}
	overlay := NewNetworkOverlay(mockNetworkManager)

	// Create character and calculator for category testing
	char := createTestCharacterWithPersonality(t, "TestChar", map[string]float64{
		"kindness": 0.7,
	})
	calculator := character.NewCompatibilityCalculator(char)
	overlay.SetCompatibilityCalculator(calculator)

	testCases := []struct {
		peerID   string
		score    float64
		expected string
	}{
		{"excellent-peer", 0.95, "Excellent"},
		{"very-good-peer", 0.85, "Very Good"},
		{"good-peer", 0.65, "Good"},
		{"fair-peer", 0.45, "Fair"},
		{"poor-peer", 0.25, "Poor"},
		{"very-poor-peer", 0.05, "Very Poor"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			// Set score
			overlay.compatibilityMutex.Lock()
			overlay.compatibilityScores[tc.peerID] = tc.score
			overlay.compatibilityMutex.Unlock()

			category := overlay.GetCompatibilityCategory(tc.peerID)
			if category != tc.expected {
				t.Errorf("Score %f: expected category %s, got %s", tc.score, tc.expected, category)
			}
		})
	}
}

// TestFeature5_GetCompatibilityCategoryFallback tests fallback categorization
func TestFeature5_GetCompatibilityCategoryFallback(t *testing.T) {
	mockNetworkManager := &MockNetworkManagerForCompatibility{}
	overlay := NewNetworkOverlay(mockNetworkManager)

	// Don't set calculator - should use fallback categorization
	overlay.compatibilityMutex.Lock()
	overlay.compatibilityScores["test-peer"] = 0.75
	overlay.compatibilityMutex.Unlock()

	category := overlay.GetCompatibilityCategory("test-peer")
	if category != "Good" {
		t.Errorf("Expected fallback category 'Good' for score 0.75, got %s", category)
	}
}

// TestFeature5_CharacterListDisplayCompatibility tests UI display of compatibility
func TestFeature5_CharacterListDisplayCompatibility(t *testing.T) {
	t.Run("DisplayCompatibilityInCharacterList", func(t *testing.T) {
		mockNetworkManager := &MockNetworkManagerForCompatibility{
			peers: []network.Peer{
				{ID: "peer1", Conn: nil}, // Use nil for mock connection
			},
		}
		overlay := NewNetworkOverlay(mockNetworkManager)

		// Set up compatibility data
		overlay.compatibilityMutex.Lock()
		overlay.compatibilityScores["peer1"] = 0.85 // Very Good
		overlay.compatibilityMutex.Unlock()

		// Create character with peer data
		overlay.characterMutex.Lock()
		overlay.characters = []CharacterInfo{
			{
				Name:     "Peer1's Character",
				Location: "peer1",
				IsLocal:  false,
				IsActive: true,
				PeerID:   "peer1",
			},
		}
		overlay.characterMutex.Unlock()

		// Get the display text (this tests the character list formatting logic)
		score := overlay.GetCompatibilityScore("peer1")
		category := overlay.GetCompatibilityCategory("peer1")

		if score != 0.85 {
			t.Errorf("Expected score 0.85, got %f", score)
		}

		if category != "Very Good" {
			t.Errorf("Expected category 'Very Good', got %s", category)
		}

		// Verify that the character list would display the compatibility info
		// (The actual UI rendering is tested through the character list widget)
	})
}

// TestFeature5_ConcurrentAccess tests thread safety
func TestFeature5_ConcurrentAccess(t *testing.T) {
	mockNetworkManager := &MockNetworkManagerForCompatibility{}
	overlay := NewNetworkOverlay(mockNetworkManager)

	// Set up calculator
	char := createTestCharacterWithPersonality(t, "TestChar", map[string]float64{
		"kindness": 0.7,
	})
	calculator := character.NewCompatibilityCalculator(char)
	overlay.SetCompatibilityCalculator(calculator)

	// Test concurrent access to compatibility functions
	done := make(chan bool, 20)

	// Concurrent score updates
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			overlay.UpdateCompatibilityScores()
		}()
	}

	// Concurrent score reads
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			_ = overlay.GetCompatibilityScore("test-peer")
			_ = overlay.GetCompatibilityCategory("test-peer")
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 20; i++ {
		<-done
	}

	// Should not have panicked due to race conditions
}

// TestFeature5_CharacterInfoExtension tests the extended CharacterInfo struct
func TestFeature5_CharacterInfoExtension(t *testing.T) {
	t.Run("CharacterInfoHasRequiredFields", func(t *testing.T) {
		char := CharacterInfo{
			Name:     "TestChar",
			Location: "peer1",
			IsLocal:  false,
			IsActive: true,
			CharType: "Network",
			PeerID:   "peer1",
			Personality: &character.PersonalityConfig{
				Traits: map[string]float64{
					"kindness": 0.7,
				},
			},
		}

		if char.PeerID != "peer1" {
			t.Errorf("Expected PeerID 'peer1', got %s", char.PeerID)
		}

		if char.Personality == nil {
			t.Error("Expected non-nil personality")
		}

		if char.Personality.Traits["kindness"] != 0.7 {
			t.Errorf("Expected kindness trait 0.7, got %f", char.Personality.Traits["kindness"])
		}
	})
}
