package news

import (
	"desktop-companion/internal/dialog"
	"encoding/json"
	"testing"
	"time"
)

func TestNewsBlogBackend_Initialize(t *testing.T) {
	backend := NewNewsBlogBackend()

	// Test with valid configuration
	config := NewsBackendConfig{
		Enabled:              true,
		SummaryLength:        50,
		PersonalityInfluence: true,
		CacheTimeout:         1800,
		UpdateInterval:       30,
		MaxNewsPerResponse:   3,
		DebugMode:           false,
		PreferredCategories: []string{"tech", "gaming"},
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	err = backend.Initialize(json.RawMessage(configJSON))
	if err != nil {
		t.Errorf("Expected successful initialization, got error: %v", err)
	}

	if !backend.enabled {
		t.Errorf("Expected backend to be enabled after initialization")
	}

	if !backend.personalityInfluence {
		t.Errorf("Expected personality influence to be enabled")
	}
}

func TestNewsBlogBackend_InitializeInvalidConfig(t *testing.T) {
	backend := NewNewsBlogBackend()

	// Test with nil configuration
	err := backend.Initialize(nil)
	if err == nil {
		t.Errorf("Expected error with nil configuration")
	}

	// Test with invalid JSON
	err = backend.Initialize(json.RawMessage(`{invalid json}`))
	if err == nil {
		t.Errorf("Expected error with invalid JSON configuration")
	}
}

func TestNewsBlogBackend_GetBackendInfo(t *testing.T) {
	backend := NewNewsBlogBackend()
	info := backend.GetBackendInfo()

	if info.Name != "news_blog" {
		t.Errorf("Expected backend name 'news_blog', got '%s'", info.Name)
	}

	if info.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", info.Version)
	}

	if len(info.Capabilities) == 0 {
		t.Errorf("Expected backend to have capabilities")
	}

	expectedCapabilities := []string{
		"news_summarization",
		"category_filtering",
		"personality_adaptation", 
		"feed_management",
	}

	for _, expected := range expectedCapabilities {
		found := false
		for _, capability := range info.Capabilities {
			if capability == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected capability '%s' not found in backend info", expected)
		}
	}
}

func TestNewsBlogBackend_CanHandle(t *testing.T) {
	backend := NewNewsBlogBackend()

	// Backend should not handle anything when disabled
	context := dialog.DialogContext{
		Trigger: "news",
	}

	if backend.CanHandle(context) {
		t.Errorf("Disabled backend should not handle any requests")
	}

	// Enable backend
	config := NewsBackendConfig{
		Enabled: true,
	}
	configJSON, _ := json.Marshal(config)
	backend.Initialize(json.RawMessage(configJSON))

	// Should handle news triggers
	context.Trigger = "news"
	if !backend.CanHandle(context) {
		t.Errorf("Enabled backend should handle 'news' trigger")
	}

	context.Trigger = "news_update"
	if !backend.CanHandle(context) {
		t.Errorf("Enabled backend should handle 'news_update' trigger")
	}

	// Should handle contexts with news topic
	context.Trigger = "click"
	context.TopicContext = map[string]interface{}{
		"newsCategory": "tech",
	}
	if !backend.CanHandle(context) {
		t.Errorf("Enabled backend should handle contexts with news topics")
	}
}

func TestNewsBlogBackend_GenerateResponse(t *testing.T) {
	backend := NewNewsBlogBackend()

	// Enable backend
	config := NewsBackendConfig{
		Enabled: true,
	}
	configJSON, _ := json.Marshal(config)
	backend.Initialize(json.RawMessage(configJSON))

	// Test response generation without news items
	context := dialog.DialogContext{
		Trigger: "news",
		PersonalityTraits: map[string]float64{
			"energy": 0.8,
		},
	}

	response, err := backend.GenerateResponse(context)
	if err != nil {
		t.Errorf("Expected successful response generation, got error: %v", err)
	}

	if response.Text == "" {
		t.Errorf("Expected non-empty response text")
	}

	if response.ResponseType != "informative" {
		t.Errorf("Expected response type 'informative', got '%s'", response.ResponseType)
	}
}

func TestNewsBlogBackend_AddFeed(t *testing.T) {
	backend := NewNewsBlogBackend()

	// Create a test feed (note: this will fail validation since it's not a real URL)
	feed := RSSFeed{
		URL:      "https://example.com/feed.rss",
		Name:     "Test Feed",
		Category: "tech",
		Enabled:  true,
	}

	// This should fail since the URL doesn't exist
	err := backend.AddFeed(feed)
	if err == nil {
		t.Errorf("Expected error when adding invalid feed URL")
	}
}

func TestNewsBlogBackend_ResponseGeneration(t *testing.T) {
	backend := NewNewsBlogBackend()

	// Enable backend
	config := NewsBackendConfig{
		Enabled: true,
		PersonalityInfluence: true,
	}
	configJSON, _ := json.Marshal(config)
	backend.Initialize(json.RawMessage(configJSON))

	// Add some test news items to the cache
	testItems := []*NewsItem{
		{
			ID:        "item1",
			Title:     "Breakthrough in AI Technology",
			Summary:   "Researchers achieve new milestone in artificial intelligence.",
			URL:       "https://example.com/ai-breakthrough",
			Published: time.Now(),
			Source:    "TechNews",
			Category:  "tech",
		},
		{
			ID:        "item2",
			Title:     "New Gaming Console Released",
			Summary:   "Latest gaming hardware hits the market.",
			URL:       "https://example.com/gaming-console",
			Published: time.Now().Add(-1 * time.Hour),
			Source:    "GameNews",
			Category:  "gaming",
		},
	}

	for _, item := range testItems {
		backend.cache.AddItem(item)
	}

	// Test response with different personality traits
	contexts := []struct {
		name   string
		traits map[string]float64
		expected string
	}{
		{
			name: "enthusiastic",
			traits: map[string]float64{
				"energy": 0.9,
			},
			expected: "enthusiastic",
		},
		{
			name: "formal",
			traits: map[string]float64{
				"intellect": 0.9,
			},
			expected: "formal",
		},
		{
			name: "casual",
			traits: map[string]float64{
				"energy": 0.3,
			},
			expected: "casual",
		},
	}

	for _, tc := range contexts {
		t.Run(tc.name, func(t *testing.T) {
			context := dialog.DialogContext{
				Trigger:           "news",
				PersonalityTraits: tc.traits,
			}

			response, err := backend.GenerateResponse(context)
			if err != nil {
				t.Errorf("Expected successful response generation, got error: %v", err)
			}

			if response.Text == "" {
				t.Errorf("Expected non-empty response text")
			}

			// Check that response contains some news content
			hasNewsContent := false
			for _, item := range testItems {
				if contains(response.Text, item.Title) {
					hasNewsContent = true
					break
				}
			}

			if !hasNewsContent {
				t.Errorf("Expected response to contain news content, got: %s", response.Text)
			}
		})
	}
}

func TestNewsBlogBackend_EmotionalTone(t *testing.T) {
	backend := NewNewsBlogBackend()

	// Enable backend
	config := NewsBackendConfig{
		Enabled: true,
	}
	configJSON, _ := json.Marshal(config)
	backend.Initialize(json.RawMessage(configJSON))

	// Test positive news
	positiveNews := []*NewsItem{
		{
			ID:        "positive1",
			Title:     "Major Breakthrough Achieved in Technology",
			Summary:   "Scientists achieve remarkable success in new research.",
			Published: time.Now(),
			Source:    "TechNews",
			Category:  "tech",
		},
	}

	context := dialog.DialogContext{
		Trigger: "news",
	}

	backend.cache.AddItem(positiveNews[0])
	response, err := backend.GenerateResponse(context)
	if err != nil {
		t.Errorf("Expected successful response generation, got error: %v", err)
	}

	// Should have optimistic tone for positive news
	if response.EmotionalTone != "optimistic" {
		t.Errorf("Expected 'optimistic' tone for positive news, got '%s'", response.EmotionalTone)
	}
}

func TestNewsBlogBackend_UpdateMemory(t *testing.T) {
	backend := NewNewsBlogBackend()

	// Enable debug mode to test memory updates
	config := NewsBackendConfig{
		Enabled:   true,
		DebugMode: true,
	}
	configJSON, _ := json.Marshal(config)
	backend.Initialize(json.RawMessage(configJSON))

	context := dialog.DialogContext{
		Trigger: "news",
	}

	response := dialog.DialogResponse{
		Text: "Test news response",
	}

	feedback := &dialog.UserFeedback{
		Positive:   true,
		Engagement: 0.8,
	}

	// Should not return an error
	err := backend.UpdateMemory(context, response, feedback)
	if err != nil {
		t.Errorf("Expected successful memory update, got error: %v", err)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		 containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
