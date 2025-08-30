package character

import (
	"desktop-companion/internal/news"
	"testing"
)

func TestCharacterCardNewsFeatures(t *testing.T) {
	// Create a basic character card with news features for testing
	testCard := &CharacterCard{
		Name:        "TestNewsBot",
		Description: "Test character with news features",
		Animations: map[string]string{
			"idle":    "idle.gif",
			"talking": "talking.gif",
		},
		Dialogs: []Dialog{
			{
				Trigger:   "click",
				Responses: []string{"Hello! I've been reading the news."},
				Animation: "talking",
				Cooldown:  5,
			},
		},
		Behavior: Behavior{
			IdleTimeout:     30,
			MovementEnabled: false,
			DefaultSize:     128,
		},
	}

	// Test character without news features
	if testCard.HasNewsFeatures() {
		t.Errorf("Character without news config should not have news features")
	}

	if testCard.GetNewsConfig() != nil {
		t.Errorf("Character without news config should return nil for GetNewsConfig()")
	}

	// Add news features
	testCard.NewsFeatures = &news.NewsConfig{
		Enabled:             true,
		UpdateInterval:      30,
		MaxStoredItems:      50,
		ReadingPersonality:  "casual",
		PreferredCategories: []string{"tech", "gaming"},
		Feeds: []news.RSSFeed{
			{
				URL:        "https://example.com/feed.rss",
				Name:       "Test Feed",
				Category:   "tech",
				UpdateFreq: 60,
				MaxItems:   10,
				Enabled:    true,
			},
		},
		ReadingEvents: []news.NewsEvent{
			{
				Name:         "test_headlines",
				Category:     "conversation",
				Trigger:      "news",
				NewsCategory: "headlines",
				MaxNews:      3,
				ReadingStyle: "casual",
				Responses:    []string{"Here are today's headlines: {NEWS_HEADLINES}"},
				Cooldown:     300,
			},
		},
	}

	// Test character with news features
	if !testCard.HasNewsFeatures() {
		t.Errorf("Character with news config should have news features")
	}

	newsConfig := testCard.GetNewsConfig()
	if newsConfig == nil {
		t.Fatalf("Character with news config should return valid config from GetNewsConfig()")
	}

	// Verify news configuration properties
	if !newsConfig.Enabled {
		t.Errorf("Expected news features to be enabled")
	}

	if newsConfig.UpdateInterval != 30 {
		t.Errorf("Expected update interval of 30, got %d", newsConfig.UpdateInterval)
	}

	if newsConfig.MaxStoredItems != 50 {
		t.Errorf("Expected max stored items of 50, got %d", newsConfig.MaxStoredItems)
	}

	if newsConfig.ReadingPersonality != "casual" {
		t.Errorf("Expected reading personality 'casual', got '%s'", newsConfig.ReadingPersonality)
	}

	// Verify feeds
	if len(newsConfig.Feeds) != 1 {
		t.Errorf("Expected 1 feed, got %d", len(newsConfig.Feeds))
	}

	feed := newsConfig.Feeds[0]
	if feed.Name != "Test Feed" {
		t.Errorf("Expected feed name 'Test Feed', got '%s'", feed.Name)
	}

	if feed.Category != "tech" {
		t.Errorf("Expected feed category 'tech', got '%s'", feed.Category)
	}

	// Verify reading events
	if len(newsConfig.ReadingEvents) != 1 {
		t.Errorf("Expected 1 reading event, got %d", len(newsConfig.ReadingEvents))
	}

	event := newsConfig.ReadingEvents[0]
	if event.Name != "test_headlines" {
		t.Errorf("Expected event name 'test_headlines', got '%s'", event.Name)
	}

	if event.NewsCategory != "headlines" {
		t.Errorf("Expected news category 'headlines', got '%s'", event.NewsCategory)
	}
}

func TestCharacterCardNewsDisabled(t *testing.T) {
	// Test character card with news features disabled
	testCard := &CharacterCard{
		Name:        "TestBot",
		Description: "Test character",
		Animations: map[string]string{
			"idle": "idle.gif",
		},
		Dialogs: []Dialog{
			{
				Trigger:   "click",
				Responses: []string{"Hello!"},
				Animation: "idle",
				Cooldown:  5,
			},
		},
		Behavior: Behavior{
			IdleTimeout: 30,
			DefaultSize: 128,
		},
		NewsFeatures: &news.NewsConfig{
			Enabled: false, // Explicitly disabled
		},
	}

	// Even though news config exists, it's disabled
	if testCard.HasNewsFeatures() {
		t.Errorf("Character with disabled news config should not have news features")
	}

	// But GetNewsConfig should still return the config
	newsConfig := testCard.GetNewsConfig()
	if newsConfig == nil {
		t.Errorf("GetNewsConfig should return config even when disabled")
	}

	if newsConfig.Enabled {
		t.Errorf("News config should be disabled")
	}
}
