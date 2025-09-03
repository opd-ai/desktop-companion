package character

import (
	"testing"
	"time"
)

// createDialogFavoritesTestGameState creates a GameState for testing with minimal configuration
func createDialogFavoritesTestGameState() *GameState {
	config := map[string]StatConfig{
		"hunger": {
			Initial:           100,
			Max:               100,
			DegradationRate:   1.0,
			CriticalThreshold: 20,
		},
	}
	return NewGameState(config, nil)
}

// TestDialogMemoryFavoriteTracking tests the dialog memory favorite functionality
func TestDialogMemoryFavoriteTracking(t *testing.T) {
	gs := createDialogFavoritesTestGameState()

	// Test initial state - no favorites
	favorites := gs.GetFavoriteDialogResponses()
	if len(favorites) != 0 {
		t.Errorf("Expected no favorites initially, got %d", len(favorites))
	}

	// Add a dialog memory
	memory := DialogMemory{
		Timestamp:        time.Now(),
		Trigger:          "chat",
		Response:         "Hello! How are you today?",
		EmotionalTone:    "friendly",
		MemoryImportance: 0.7,
		BackendUsed:      "markov",
		Confidence:       0.8,
	}
	gs.RecordDialogMemory(memory)

	// Test marking as favorite
	success := gs.MarkDialogResponseFavorite("Hello! How are you today?", 4.0)
	if !success {
		t.Error("Failed to mark dialog response as favorite")
	}

	// Verify favorite status
	isFavorite, rating := gs.IsDialogResponseFavorite("Hello! How are you today?")
	if !isFavorite {
		t.Error("Response should be marked as favorite")
	}
	if rating != 4.0 {
		t.Errorf("Expected rating 4.0, got %f", rating)
	}

	// Test getting favorites
	favorites = gs.GetFavoriteDialogResponses()
	if len(favorites) != 1 {
		t.Errorf("Expected 1 favorite, got %d", len(favorites))
	}

	// Test favorites by rating
	highRatedFavorites := gs.GetFavoriteResponsesByRating(4.0)
	if len(highRatedFavorites) != 1 {
		t.Errorf("Expected 1 high-rated favorite, got %d", len(highRatedFavorites))
	}

	lowRatedFavorites := gs.GetFavoriteResponsesByRating(5.0)
	if len(lowRatedFavorites) != 0 {
		t.Errorf("Expected 0 favorites with rating >= 5.0, got %d", len(lowRatedFavorites))
	}
}

// TestDialogMemoryFavoriteUnmarking tests removing favorite status
func TestDialogMemoryFavoriteUnmarking(t *testing.T) {
	gs := createDialogFavoritesTestGameState()

	// Add and mark as favorite
	memory := DialogMemory{
		Timestamp:        time.Now(),
		Trigger:          "chat",
		Response:         "Test response",
		EmotionalTone:    "neutral",
		MemoryImportance: 0.5,
		BackendUsed:      "simple",
		Confidence:       0.6,
	}
	gs.RecordDialogMemory(memory)
	gs.MarkDialogResponseFavorite("Test response", 3.0)

	// Verify it's marked as favorite
	isFavorite, rating := gs.IsDialogResponseFavorite("Test response")
	if !isFavorite || rating != 3.0 {
		t.Error("Response should be marked as favorite with rating 3.0")
	}

	// Unmark as favorite
	success := gs.UnmarkDialogResponseFavorite("Test response")
	if !success {
		t.Error("Failed to unmark dialog response as favorite")
	}

	// Verify it's no longer a favorite
	isFavorite, rating = gs.IsDialogResponseFavorite("Test response")
	if isFavorite || rating != 0 {
		t.Error("Response should no longer be marked as favorite")
	}

	// Verify no favorites exist
	favorites := gs.GetFavoriteDialogResponses()
	if len(favorites) != 0 {
		t.Errorf("Expected no favorites, got %d", len(favorites))
	}
}

// TestDialogMemoryFavoriteEdgeCases tests edge cases and error conditions
func TestDialogMemoryFavoriteEdgeCases(t *testing.T) {
	gs := createDialogFavoritesTestGameState()

	// Test marking non-existent response as favorite
	success := gs.MarkDialogResponseFavorite("Non-existent response", 5.0)
	if success {
		t.Error("Should not be able to mark non-existent response as favorite")
	}

	// Test unmarking non-existent response
	success = gs.UnmarkDialogResponseFavorite("Non-existent response")
	if success {
		t.Error("Should not be able to unmark non-existent response")
	}

	// Test with nil dialog memories
	gs.DialogMemories = nil
	success = gs.MarkDialogResponseFavorite("Test", 5.0)
	if success {
		t.Error("Should not be able to mark favorite with nil dialog memories")
	}

	isFavorite, rating := gs.IsDialogResponseFavorite("Test")
	if isFavorite || rating != 0 {
		t.Error("Should return false for nil dialog memories")
	}

	favorites := gs.GetFavoriteDialogResponses()
	if len(favorites) != 0 {
		t.Error("Should return empty slice for nil dialog memories")
	}
}

// TestDialogMemoryFavoriteMultipleResponses tests handling multiple responses with same text
func TestDialogMemoryFavoriteMultipleResponses(t *testing.T) {
	gs := createDialogFavoritesTestGameState()

	// Add same response multiple times
	for i := 0; i < 3; i++ {
		memory := DialogMemory{
			Timestamp:        time.Now().Add(time.Duration(i) * time.Second),
			Trigger:          "chat",
			Response:         "Common response",
			EmotionalTone:    "neutral",
			MemoryImportance: 0.5,
			BackendUsed:      "simple",
			Confidence:       0.6,
		}
		gs.RecordDialogMemory(memory)
	}

	// Mark as favorite - should mark the most recent one
	success := gs.MarkDialogResponseFavorite("Common response", 4.5)
	if !success {
		t.Error("Failed to mark response as favorite")
	}

	// Count how many are marked as favorite (should be only the most recent)
	favorites := gs.GetFavoriteDialogResponses()
	favoriteCount := 0
	for _, memory := range favorites {
		if memory.Response == "Common response" && memory.IsFavorite {
			favoriteCount++
		}
	}

	if favoriteCount != 1 {
		t.Errorf("Expected 1 favorite response, got %d", favoriteCount)
	}
}

// TestDialogMemoryFavoriteThreadSafety tests thread safety of favorite operations
func TestDialogMemoryFavoriteThreadSafety(t *testing.T) {
	gs := createDialogFavoritesTestGameState()

	// Add a dialog memory
	memory := DialogMemory{
		Timestamp:        time.Now(),
		Trigger:          "chat",
		Response:         "Thread safety test",
		EmotionalTone:    "neutral",
		MemoryImportance: 0.5,
		BackendUsed:      "simple",
		Confidence:       0.6,
	}
	gs.RecordDialogMemory(memory)

	// Test concurrent access
	done := make(chan bool, 2)

	// Goroutine 1: Mark as favorite
	go func() {
		gs.MarkDialogResponseFavorite("Thread safety test", 3.0)
		done <- true
	}()

	// Goroutine 2: Check favorite status
	go func() {
		gs.IsDialogResponseFavorite("Thread safety test")
		gs.GetFavoriteDialogResponses()
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// Verify final state
	isFavorite, rating := gs.IsDialogResponseFavorite("Thread safety test")
	if !isFavorite || rating != 3.0 {
		t.Error("Expected response to be marked as favorite with rating 3.0")
	}
}

// TestDialogMemoryFavoriteBackwardCompatibility tests that existing functionality still works
func TestDialogMemoryFavoriteBackwardCompatibility(t *testing.T) {
	gs := createDialogFavoritesTestGameState()

	// Test that existing dialog memory methods work unchanged
	memory := DialogMemory{
		Timestamp:        time.Now(),
		Trigger:          "click",
		Response:         "Backward compatibility test",
		EmotionalTone:    "happy",
		MemoryImportance: 0.8,
		BackendUsed:      "markov",
		Confidence:       0.9,
		// IsFavorite and FavoriteRating intentionally omitted (default false/0)
	}
	gs.RecordDialogMemory(memory)

	// Test existing methods
	memories := gs.GetDialogMemories()
	if len(memories) != 1 {
		t.Errorf("Expected 1 dialog memory, got %d", len(memories))
	}

	recentMemories := gs.GetRecentDialogMemories(1)
	if len(recentMemories) != 1 {
		t.Errorf("Expected 1 recent memory, got %d", len(recentMemories))
	}

	triggerMemories := gs.GetDialogMemoriesByTrigger("click")
	if len(triggerMemories) != 1 {
		t.Errorf("Expected 1 trigger memory, got %d", len(triggerMemories))
	}

	importantMemories := gs.GetHighImportanceDialogMemories(0.7)
	if len(importantMemories) != 1 {
		t.Errorf("Expected 1 important memory, got %d", len(importantMemories))
	}

	// Verify new fields have default values
	memory = memories[0]
	if memory.IsFavorite != false {
		t.Error("Expected IsFavorite to default to false")
	}
	if memory.FavoriteRating != 0 {
		t.Error("Expected FavoriteRating to default to 0")
	}
}
