package character

import (
	"desktop-companion/internal/dialog"
	"desktop-companion/internal/news"
	"fmt"
	"strings"
	"time"
)

// NewsDialogContext extends DialogContext for news-specific dialog generation
// This provides additional context for news-related conversations
type NewsDialogContext struct {
	dialog.DialogContext                 // Embed existing context
	NewsItems            []news.NewsItem `json:"newsItems"`
	RequestedCategory    string          `json:"requestedCategory"`
	MaxItems             int             `json:"maxItems"`
	IncludeSummary       bool            `json:"includeSummary"`
	ReadingStyle         string          `json:"readingStyle"`
}

// initializeNewsEvents sets up news-specific event handling
// This method is called during character initialization if news features are enabled
func (c *Character) initializeNewsEvents() error {
	if !c.card.HasNewsFeatures() {
		return nil
	}

	// Initialize news backend with feeds from character configuration
	if err := c.initializeNewsBackend(); err != nil {
		return fmt.Errorf("failed to initialize news backend: %w", err)
	}

	if c.debug {
		fmt.Printf("[DEBUG] Initialized news events for character %s\n", c.card.Name)
	}

	return nil
}

// initializeNewsBackend configures the news backend with character's feed configuration
func (c *Character) initializeNewsBackend() error {
	if c.dialogManager == nil {
		return fmt.Errorf("dialog manager not available")
	}

	// Get the news backend
	backend, exists := c.dialogManager.GetBackend("news_blog")
	if !exists {
		return fmt.Errorf("news backend not registered")
	}

	newsBackend, ok := backend.(*news.NewsBlogBackend)
	if !ok {
		return fmt.Errorf("invalid news backend type")
	}

	// Add feeds from character configuration
	for _, feed := range c.card.NewsFeatures.Feeds {
		if err := newsBackend.AddFeed(feed); err != nil {
			if c.debug {
				fmt.Printf("[DEBUG] Failed to add feed %s: %v\n", feed.Name, err)
			}
			continue
		}
	}

	// Perform initial feed update if enabled
	if len(c.card.NewsFeatures.Feeds) > 0 {
		go func() {
			if err := newsBackend.UpdateFeeds(); err != nil && c.debug {
				fmt.Printf("[DEBUG] Initial feed update failed: %v\n", err)
			}
		}()
	}

	return nil
}

// HandleNewsEvent processes a news-specific event trigger
// This method creates appropriate dialog context and generates news-aware responses
func (c *Character) HandleNewsEvent(eventName string) (string, error) {
	if !c.card.HasNewsFeatures() {
		return "", fmt.Errorf("news features not enabled for this character")
	}

	// Find the news event configuration
	var newsEventConfig *news.NewsEvent
	for i := range c.card.NewsFeatures.ReadingEvents {
		if c.card.NewsFeatures.ReadingEvents[i].Name == eventName {
			newsEventConfig = &c.card.NewsFeatures.ReadingEvents[i]
			break
		}
	}

	if newsEventConfig == nil {
		return "", fmt.Errorf("news event %q not found", eventName)
	}

	// Check if event is on cooldown
	if c.isNewsEventOnCooldown(eventName) {
		return "", fmt.Errorf("event %q is on cooldown", eventName)
	}

	// Create news-specific dialog context
	newsContext := c.createNewsDialogContext(newsEventConfig)

	// Generate response using news backend if available
	response, err := c.generateNewsResponse(newsContext)
	if err != nil {
		// Fallback to regular dialog generation
		return c.generateFallbackNewsResponse(newsEventConfig)
	}

	// Record event usage for cooldown tracking
	c.recordNewsEventUsage(eventName)

	return response.Text, nil
}

// createNewsDialogContext creates a specialized context for news dialog generation
func (c *Character) createNewsDialogContext(config *news.NewsEvent) NewsDialogContext {
	// Create base dialog context
	baseContext := dialog.DialogContext{
		Trigger:           config.Trigger,
		InteractionID:     fmt.Sprintf("news_%s_%d", config.Name, time.Now().Unix()),
		Timestamp:         time.Now(),
		CurrentStats:      c.getCurrentStatsMap(),
		PersonalityTraits: c.getPersonalityTraitsMap(),
		CurrentMood:       c.getCurrentMood(),
		CurrentAnimation:  c.GetCurrentState(),
		TimeOfDay:         c.getTimeOfDay(),
		TopicContext: map[string]interface{}{
			"newsCategory":   config.NewsCategory,
			"maxNews":        config.MaxNews,
			"includeSummary": config.IncludeSummary,
			"readingStyle":   config.ReadingStyle,
		},
	}

	// Create news-specific context
	return NewsDialogContext{
		DialogContext:     baseContext,
		RequestedCategory: config.NewsCategory,
		MaxItems:          config.MaxNews,
		IncludeSummary:    config.IncludeSummary,
		ReadingStyle:      config.ReadingStyle,
	}
}

// generateNewsResponse generates a news response using the dialog manager
func (c *Character) generateNewsResponse(context NewsDialogContext) (dialog.DialogResponse, error) {
	if c.dialogManager == nil {
		return dialog.DialogResponse{}, fmt.Errorf("dialog manager not available")
	}

	// Use the news backend if available, otherwise fall back to default
	response, err := c.dialogManager.GenerateDialog(context.DialogContext)
	if err != nil {
		return dialog.DialogResponse{}, fmt.Errorf("failed to generate news response: %w", err)
	}

	return response, nil
}

// generateFallbackNewsResponse provides a simple fallback when news backend is unavailable
func (c *Character) generateFallbackNewsResponse(config *news.NewsEvent) (string, error) {
	if len(config.Responses) == 0 {
		return "I'd love to share some news with you, but I'm having trouble accessing my news sources right now.", nil
	}

	// Select a random response from configured options
	responseIndex := time.Now().Second() % len(config.Responses)
	response := config.Responses[responseIndex]

	// Simple template replacement for fallback
	response = strings.ReplaceAll(response, "{NEWS_HEADLINES}", "the latest headlines")
	response = strings.ReplaceAll(response, "{NEWS_SUMMARY}", "some interesting news")

	return response, nil
}

// isNewsEventOnCooldown checks if a news event is currently on cooldown
func (c *Character) isNewsEventOnCooldown(eventName string) bool {
	if lastUsed, exists := c.dialogCooldowns[eventName]; exists {
		// Find the event configuration to get cooldown duration
		for _, event := range c.card.NewsFeatures.ReadingEvents {
			if event.Name == eventName {
				cooldownDuration := time.Duration(event.Cooldown) * time.Second
				return time.Since(lastUsed) < cooldownDuration
			}
		}
	}
	return false
}

// recordNewsEventUsage records when a news event was used for cooldown tracking
func (c *Character) recordNewsEventUsage(eventName string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.dialogCooldowns[eventName] = time.Now()
}

// Helper methods for context creation

func (c *Character) getCurrentStatsMap() map[string]float64 {
	if c.gameState == nil {
		return map[string]float64{}
	}

	stats := c.gameState.GetStats()
	return stats
}

func (c *Character) getPersonalityTraitsMap() map[string]float64 {
	if c.card.Personality == nil {
		return map[string]float64{}
	}

	// PersonalityConfig uses a Traits map[string]float64 field
	return c.card.Personality.Traits
}

func (c *Character) getCurrentMood() float64 {
	if c.gameState == nil {
		return 50.0 // Neutral mood
	}

	stats := c.gameState.GetStats()
	// Calculate overall mood from stats (simple average)
	happiness := stats["happiness"]
	health := stats["health"]
	energy := stats["energy"]
	jealousy := stats["jealousy"]

	moodSum := happiness + health + energy - jealousy
	return moodSum / 3.0
}
