package character

import (
	"desktop-companion/internal/dialog"
	"desktop-companion/internal/news"
	"encoding/json"
	"testing"
)

// TestNewsEventsInitialization tests that news events are properly initialized
func TestNewsEventsInitialization(t *testing.T) {
	// Create a character with news features enabled
	card := &CharacterCard{
		Name:        "News Test Character",
		Description: "A character for testing news functionality",
		Animations: map[string]string{
			"idle": "idle.gif",
		},
		Dialogs: []Dialog{
			{Trigger: "click", Responses: []string{"Hello!"}, Animation: "idle"},
		},
		Behavior: Behavior{
			IdleTimeout:     30,
			MovementEnabled: false,
			DefaultSize:     128,
		},
		DialogBackend: &dialog.DialogBackendConfig{
			Enabled:        true,
			DefaultBackend: "news_blog",
			FallbackChain:  []string{"news_blog", "simple_random"},
			Backends: map[string]json.RawMessage{
				"news_blog": json.RawMessage(`{
					"enabled": true,
					"summaryLength": 100,
					"personalityInfluence": true,
					"cacheTimeout": 1800,
					"debugMode": false
				}`),
			},
		},
		NewsFeatures: &news.NewsConfig{
			Enabled:             true,
			UpdateInterval:      30,
			MaxStoredItems:      50,
			ReadingPersonality:  "casual",
			PreferredCategories: []string{"tech", "gaming"},
			Feeds: []news.RSSFeed{
				{
					URL:        "https://feeds.feedburner.com/TechCrunch",
					Name:       "TechCrunch",
					Category:   "tech",
					UpdateFreq: 60,
					MaxItems:   10,
					Enabled:    true,
				},
			},
			ReadingEvents: []news.NewsEvent{
				{
					Name:           "morning_news",
					Category:       "conversation",
					Trigger:        "daily_news",
					NewsCategory:   "headlines",
					MaxNews:        3,
					IncludeSummary: false,
					ReadingStyle:   "casual",
					Responses:      []string{"Good morning! Here's what's happening: {NEWS_HEADLINES}"},
					Animations:     []string{"talking"},
					Cooldown:       3600,
				},
			},
		},
	}

	char, err := New(card, "../../testdata")
	if err != nil {
		t.Fatalf("Failed to create character: %v", err)
	}

	// Test that news features are recognized
	if !char.card.HasNewsFeatures() {
		t.Error("Character should have news features enabled")
	}

	// Test that dialog manager is initialized
	if char.dialogManager == nil {
		t.Error("Dialog manager should be initialized for news-enabled character")
	}

	// Test that news backend is registered
	backend, exists := char.dialogManager.GetBackend("news_blog")
	if !exists {
		t.Error("News backend should be registered")
	}

	if backend == nil {
		t.Error("News backend should not be nil")
	}
}

// TestNewsEventsWithoutNewsFeatures tests characters without news features
func TestNewsEventsWithoutNewsFeatures(t *testing.T) {
	// Create a character without news features
	card := &CharacterCard{
		Name:        "Regular Character",
		Description: "A regular character without news",
		Animations: map[string]string{
			"idle": "idle.gif",
		},
		Dialogs: []Dialog{
			{Trigger: "click", Responses: []string{"Hello!"}, Animation: "idle"},
		},
		Behavior: Behavior{
			IdleTimeout:     30,
			MovementEnabled: false,
			DefaultSize:     128,
		},
		// No NewsFeatures field
	}

	char, err := New(card, "../../testdata")
	if err != nil {
		t.Fatalf("Failed to create character: %v", err)
	}

	// Test that news features are not recognized
	if char.card.HasNewsFeatures() {
		t.Error("Character should not have news features enabled")
	}

	// Even without news features, dialog manager might exist for other backends
	if char.dialogManager != nil {
		// Test that news backend is not registered
		_, exists := char.dialogManager.GetBackend("news_blog")
		if exists {
			t.Error("News backend should not be registered for non-news character")
		}
	}
}

// TestNewsEventHandlingWithoutNewsFeatures tests error handling for non-news characters
func TestNewsEventHandlingWithoutNewsFeatures(t *testing.T) {
	// Create a character without news features
	card := &CharacterCard{
		Name:        "Regular Character",
		Description: "A regular character without news",
		Animations: map[string]string{
			"idle": "idle.gif",
		},
		Dialogs: []Dialog{
			{Trigger: "click", Responses: []string{"Hello!"}, Animation: "idle"},
		},
		Behavior: Behavior{
			IdleTimeout:     30,
			MovementEnabled: false,
			DefaultSize:     128,
		},
	}

	char, err := New(card, "../../testdata")
	if err != nil {
		t.Fatalf("Failed to create character: %v", err)
	}

	// Test handling news event on non-news character
	_, err = char.HandleNewsEvent("any_event")
	if err == nil {
		t.Error("HandleNewsEvent should error for characters without news features")
	}
}
