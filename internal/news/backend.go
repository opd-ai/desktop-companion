package news

import (
	"context"
	"desktop-companion/internal/dialog"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// NewsBlogBackend implements the DialogBackend interface for news summarization
// It follows the existing plugin pattern established by the markov chain backend
// Phase 4: Now includes background updating, smart scheduling, and error recovery
type NewsBlogBackend struct {
	// Core backend properties
	enabled    bool
	confidence float64
	debug      bool

	// Phase 4: Production-ready components
	feedManager *FeedManager
	config      *NewsBackendConfig
	ctx         context.Context
	cancel      context.CancelFunc

	// Legacy components (for backward compatibility)
	fetcher *FeedFetcher
	cache   *NewsCache
	feeds   []RSSFeed

	// Concurrency protection
	mu          sync.RWMutex
	updateTimer *time.Timer

	// Integration with character personality
	personalityInfluence bool
}

// NewsBackendConfig defines configuration for the news backend
type NewsBackendConfig struct {
	Enabled              bool     `json:"enabled"`
	SummaryLength        int      `json:"summaryLength"`        // Maximum length for news summaries
	PersonalityInfluence bool     `json:"personalityInfluence"` // Whether to adapt responses to personality
	CacheTimeout         int      `json:"cacheTimeout"`         // Seconds to cache responses
	UpdateInterval       int      `json:"updateInterval"`       // Minutes between feed updates
	MaxNewsPerResponse   int      `json:"maxNewsPerResponse"`   // Maximum news items per response
	DebugMode            bool     `json:"debugMode"`            // Enable debug logging
	PreferredCategories  []string `json:"preferredCategories"`  // Preferred news categories
	
	// Phase 4: New configuration options
	BackgroundUpdates    bool     `json:"backgroundUpdates"`    // Enable background feed updating
	SmartScheduling      bool     `json:"smartScheduling"`      // Enable intelligent update scheduling
	ErrorRecovery        bool     `json:"errorRecovery"`        // Enable comprehensive error handling
	MaxCacheItems        int      `json:"maxCacheItems"`        // Maximum items in cache
	BandwidthConscious   bool     `json:"bandwidthConscious"`   // Enable bandwidth-conscious policies
}

// NewNewsBlogBackend creates a new news blog backend with Phase 4 enhancements
func NewNewsBlogBackend() *NewsBlogBackend {
	// Create context for background operations
	ctx, cancel := context.WithCancel(context.Background())
	
	// Phase 4: Create production-ready feed manager
	feedManager := NewFeedManager()
	
	backend := &NewsBlogBackend{
		enabled:              false,
		confidence:           0.7, // Default confidence level
		feedManager:          feedManager,
		ctx:                  ctx,
		cancel:               cancel,
		
		// Legacy components for backward compatibility
		fetcher:              NewFeedFetcher(30 * time.Second),
		cache:                NewNewsCache(100), // Default: store up to 100 news items
		feeds:                []RSSFeed{},
		personalityInfluence: true,
		debug:                false,
	}
	
	return backend
}

// Initialize sets up the backend with JSON configuration
func (nb *NewsBlogBackend) Initialize(config json.RawMessage) error {
	if config == nil {
		return fmt.Errorf("news backend requires configuration")
	}

	var backendConfig NewsBackendConfig
	if err := json.Unmarshal(config, &backendConfig); err != nil {
		return fmt.Errorf("failed to parse news backend config: %w", err)
	}

	nb.mu.Lock()
	defer nb.mu.Unlock()

	nb.config = &backendConfig
	nb.enabled = backendConfig.Enabled
	nb.debug = backendConfig.DebugMode
	nb.personalityInfluence = backendConfig.PersonalityInfluence

	// Phase 4: Configure production-ready cache
	cacheSize := backendConfig.SummaryLength
	if backendConfig.MaxCacheItems > 0 {
		cacheSize = backendConfig.MaxCacheItems
	}
	if cacheSize > 0 {
		nb.cache = NewNewsCache(cacheSize)
	}

	// Phase 4: Start background feed manager if enabled
	if nb.enabled && backendConfig.BackgroundUpdates {
		if err := nb.feedManager.Start(nb.ctx); err != nil {
			return fmt.Errorf("failed to start feed manager: %w", err)
		}
		
		if nb.debug {
			fmt.Printf("[DEBUG] News backend: background updates enabled\n")
		}
	}

	// Log initialization if debug mode is enabled
	if nb.debug {
		fmt.Printf("[DEBUG] News backend initialized with %d preferred categories\n",
			len(backendConfig.PreferredCategories))
	}

	return nil
}

// GenerateResponse produces a dialog response for the given context
func (nb *NewsBlogBackend) GenerateResponse(context dialog.DialogContext) (dialog.DialogResponse, error) {
	nb.mu.RLock()
	defer nb.mu.RUnlock()

	if !nb.enabled {
		return dialog.DialogResponse{}, fmt.Errorf("news backend is disabled")
	}

	// Determine what type of news to fetch based on context
	newsCategory := "headlines" // Default category
	maxNews := 3                // Default number of news items

	// Check if this is a news-specific request
	if topicContext, exists := context.TopicContext["newsCategory"]; exists {
		if category, ok := topicContext.(string); ok {
			newsCategory = category
		}
	}

	if maxContext, exists := context.TopicContext["maxNews"]; exists {
		if max, ok := maxContext.(int); ok {
			maxNews = max
		}
	}

	// Get relevant news items
	newsItems := nb.getRelevantNews(newsCategory, maxNews)
	if len(newsItems) == 0 {
		return dialog.DialogResponse{
			Text:          "I don't have any recent news to share right now.",
			Confidence:    0.3,
			ResponseType:  "informative",
			EmotionalTone: "neutral",
		}, nil
	}

	// Generate response based on personality and news items
	response := nb.generateNewsResponse(newsItems, context)

	return response, nil
}

// GetBackendInfo returns metadata about this backend implementation
func (nb *NewsBlogBackend) GetBackendInfo() dialog.BackendInfo {
	return dialog.BackendInfo{
		Name:        "news_blog",
		Version:     "1.0.0",
		Description: "RSS/Atom news feed integration for Desktop Dating Simulator",
		Capabilities: []string{
			"news_summarization",
			"category_filtering",
			"personality_adaptation",
			"feed_management",
		},
		Author:  "Desktop Companion DDS",
		License: "BSD-3-Clause",
	}
}

// CanHandle checks if this backend can process the given trigger/context
func (nb *NewsBlogBackend) CanHandle(context dialog.DialogContext) bool {
	nb.mu.RLock()
	defer nb.mu.RUnlock()

	if !nb.enabled {
		return false
	}

	// Check if this is a news-related request
	if context.Trigger == "news" || context.Trigger == "news_update" {
		return true
	}

	// Check topic context for news requests
	if _, exists := context.TopicContext["newsCategory"]; exists {
		return true
	}

	// Check if we have recent news and this is a general conversation
	if context.Trigger == "click" || context.Trigger == "rightclick" {
		return len(nb.cache.GetRecentItems(1)) > 0
	}

	return false
}

// UpdateMemory allows the backend to record interaction outcomes for learning
func (nb *NewsBlogBackend) UpdateMemory(context dialog.DialogContext, response dialog.DialogResponse, userFeedback *dialog.UserFeedback) error {
	// For now, we don't implement learning, but this could be extended
	// to track which news categories users prefer based on engagement
	if nb.debug && userFeedback != nil {
		fmt.Printf("[DEBUG] News backend received feedback: positive=%v, engagement=%.2f\n",
			userFeedback.Positive, userFeedback.Engagement)
	}
	return nil
}

// AddFeed adds a new RSS feed to the backend with Phase 4 enhancements
func (nb *NewsBlogBackend) AddFeed(feed RSSFeed) error {
	nb.mu.Lock()
	defer nb.mu.Unlock()

	// Validate the feed URL
	if err := nb.fetcher.ValidateFeedURL(feed.URL); err != nil {
		return fmt.Errorf("invalid feed URL: %w", err)
	}

	// Phase 4: Register with feed manager for background updating
	if nb.config != nil && nb.config.BackgroundUpdates {
		if err := nb.feedManager.AddFeed(feed); err != nil {
			return fmt.Errorf("failed to register feed with manager: %w", err)
		}
		
		if nb.debug {
			fmt.Printf("[DEBUG] Registered feed with background manager: %s\n", feed.Name)
		}
	}

	// Maintain backward compatibility with legacy feeds list
	nb.feeds = append(nb.feeds, feed)

	if nb.debug {
		fmt.Printf("[DEBUG] Added news feed: %s (%s)\n", feed.Name, feed.URL)
	}

	return nil
}

// UpdateFeeds fetches latest news from all configured feeds
func (nb *NewsBlogBackend) UpdateFeeds() error {
	nb.mu.Lock()
	defer nb.mu.Unlock()

	var totalItems int
	var errors []string

	for _, feed := range nb.feeds {
		if !feed.Enabled {
			continue
		}

		// Check if enough time has passed since last update
		lastUpdate := nb.cache.GetLastUpdate(feed.Name)
		updateInterval := time.Duration(feed.UpdateFreq) * time.Minute
		if time.Since(lastUpdate) < updateInterval {
			continue // Skip this feed, too soon to update
		}

		// Fetch news items
		items, err := nb.fetcher.FetchFeed(feed)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Feed %s: %v", feed.Name, err))
			continue
		}

		// Add items to cache
		for _, item := range items {
			nb.cache.AddItem(item)
			totalItems++
		}

		// Update timestamp
		nb.cache.UpdateFeedTimestamp(feed.Name)

		if nb.debug {
			fmt.Printf("[DEBUG] Updated feed %s: %d new items\n", feed.Name, len(items))
		}
	}

	if len(errors) > 0 && nb.debug {
		fmt.Printf("[DEBUG] Feed update errors: %v\n", errors)
	}

	if nb.debug {
		fmt.Printf("[DEBUG] Feed update complete: %d total new items\n", totalItems)
	}

	return nil
}

// getRelevantNews retrieves news items based on category and limits (Phase 4 enhanced)
func (nb *NewsBlogBackend) getRelevantNews(category string, maxItems int) []*NewsItem {
	// Phase 4: Try to get news from background feed manager first
	if nb.config != nil && nb.config.BackgroundUpdates && nb.feedManager != nil {
		if items, err := nb.feedManager.GetLatestNews(category, maxItems); err == nil && len(items) > 0 {
			if nb.debug {
				fmt.Printf("[DEBUG] Retrieved %d items from feed manager for category '%s'\n", len(items), category)
			}
			return items
		}
	}
	
	// Fallback to legacy cache for backward compatibility
	if category == "headlines" || category == "recent" {
		return nb.cache.GetRecentItems(maxItems)
	}

	return nb.cache.GetItemsByCategory(category, maxItems)
}

// generateNewsResponse creates a personalized news response
func (nb *NewsBlogBackend) generateNewsResponse(newsItems []*NewsItem, context dialog.DialogContext) dialog.DialogResponse {
	if len(newsItems) == 0 {
		return dialog.DialogResponse{
			Text:          "I don't have any news to share right now.",
			Confidence:    0.3,
			ResponseType:  "informative",
			EmotionalTone: "neutral",
		}
	}

	// Determine reading style based on personality
	readingStyle := nb.determineReadingStyle(context)

	// Generate response text
	var responseText string
	if len(newsItems) == 1 {
		responseText = nb.generateSingleNewsResponse(newsItems[0], readingStyle)
	} else {
		responseText = nb.generateMultiNewsResponse(newsItems, readingStyle)
	}

	// Determine emotional tone based on news content and personality
	emotionalTone := nb.determineEmotionalTone(newsItems, context)

	return dialog.DialogResponse{
		Text:             responseText,
		Confidence:       nb.confidence,
		ResponseType:     "informative",
		EmotionalTone:    emotionalTone,
		Topics:           nb.extractTopics(newsItems),
		MemoryImportance: 0.6, // News is moderately important for memory
	}
}

// determineReadingStyle analyzes personality to determine how to present news
func (nb *NewsBlogBackend) determineReadingStyle(context dialog.DialogContext) string {
	if !nb.personalityInfluence {
		return "casual"
	}

	// Analyze personality traits if available
	if traits := context.PersonalityTraits; len(traits) > 0 {
		// High energy/extroversion = enthusiastic
		if energy, exists := traits["energy"]; exists && energy > 0.7 {
			return "enthusiastic"
		}

		// High intellect/seriousness = formal
		if intellect, exists := traits["intellect"]; exists && intellect > 0.8 {
			return "formal"
		}
	}

	// Default to casual
	return "casual"
}

// generateSingleNewsResponse creates a response for a single news item
func (nb *NewsBlogBackend) generateSingleNewsResponse(item *NewsItem, style string) string {
	templates := nb.getSingleNewsTemplates(style)

	// Use simple template selection based on content
	templateIndex := len(item.Title) % len(templates)
	template := templates[templateIndex]

	// Replace placeholders
	response := strings.ReplaceAll(template, "{TITLE}", item.Title)
	response = strings.ReplaceAll(response, "{SOURCE}", item.Source)

	if len(item.Summary) > 0 && len(item.Summary) < 100 {
		response = strings.ReplaceAll(response, "{SUMMARY}", item.Summary)
	} else {
		response = strings.ReplaceAll(response, " {SUMMARY}", "")
		response = strings.ReplaceAll(response, "{SUMMARY}", "")
	}

	return response
}

// generateMultiNewsResponse creates a response for multiple news items
func (nb *NewsBlogBackend) generateMultiNewsResponse(items []*NewsItem, style string) string {
	templates := nb.getMultiNewsTemplates(style)

	// Use simple template selection
	templateIndex := len(items) % len(templates)
	template := templates[templateIndex]

	// Create headlines list
	var headlines []string
	for i, item := range items {
		if i >= 3 { // Limit to 3 headlines for readability
			break
		}
		headlines = append(headlines, fmt.Sprintf("• %s", item.Title))
	}

	headlinesList := strings.Join(headlines, "\n")

	// Replace placeholders
	response := strings.ReplaceAll(template, "{HEADLINES}", headlinesList)
	response = strings.ReplaceAll(response, "{COUNT}", fmt.Sprintf("%d", len(items)))

	return response
}

// getSingleNewsTemplates returns templates for single news item responses
func (nb *NewsBlogBackend) getSingleNewsTemplates(style string) []string {
	switch style {
	case "enthusiastic":
		return []string{
			"Oh wow! I just read about this: {TITLE}! {SUMMARY}",
			"This is so interesting! {TITLE} from {SOURCE}. {SUMMARY}",
			"You've got to hear about this! {TITLE}! {SUMMARY}",
		}
	case "formal":
		return []string{
			"I came across this article: {TITLE}. {SUMMARY}",
			"There's an interesting development: {TITLE} from {SOURCE}. {SUMMARY}",
			"I thought you might find this informative: {TITLE}. {SUMMARY}",
		}
	default: // casual
		return []string{
			"I saw this news: {TITLE}. {SUMMARY}",
			"Hey, check this out: {TITLE} from {SOURCE}. {SUMMARY}",
			"Interesting news: {TITLE}. {SUMMARY}",
		}
	}
}

// getMultiNewsTemplates returns templates for multiple news items
func (nb *NewsBlogBackend) getMultiNewsTemplates(style string) []string {
	switch style {
	case "enthusiastic":
		return []string{
			"I've been reading the news and found some amazing stories!\n{HEADLINES}",
			"So much interesting stuff happening today!\n{HEADLINES}",
			"The news is really exciting today! Here are the highlights:\n{HEADLINES}",
		}
	case "formal":
		return []string{
			"Here are today's key headlines:\n{HEADLINES}",
			"I've compiled the most relevant news stories:\n{HEADLINES}",
			"Current news summary ({COUNT} items):\n{HEADLINES}",
		}
	default: // casual
		return []string{
			"Here's what's happening in the news:\n{HEADLINES}",
			"I've been browsing the news. Here's what caught my attention:\n{HEADLINES}",
			"Some interesting news today:\n{HEADLINES}",
		}
	}
}

// determineEmotionalTone analyzes news content to determine appropriate emotional response
func (nb *NewsBlogBackend) determineEmotionalTone(newsItems []*NewsItem, context dialog.DialogContext) string {
	// Simple sentiment analysis based on keywords
	positiveKeywords := []string{"breakthrough", "success", "achievement", "win", "launch", "growth"}
	negativeKeywords := []string{"crisis", "failure", "problem", "decline", "crash", "concern"}

	var positiveCount, negativeCount int

	for _, item := range newsItems {
		content := strings.ToLower(item.Title + " " + item.Summary)

		for _, keyword := range positiveKeywords {
			if strings.Contains(content, keyword) {
				positiveCount++
			}
		}

		for _, keyword := range negativeKeywords {
			if strings.Contains(content, keyword) {
				negativeCount++
			}
		}
	}

	// Determine tone based on content analysis
	if positiveCount > negativeCount {
		return "optimistic"
	} else if negativeCount > positiveCount {
		return "concerned"
	}

	return "neutral"
}

// extractTopics identifies key topics from news items
func (nb *NewsBlogBackend) extractTopics(newsItems []*NewsItem) []string {
	topicMap := make(map[string]bool)

	for _, item := range newsItems {
		// Add source as a topic
		topicMap[item.Source] = true

		// Add category as a topic
		if item.Category != "" {
			topicMap[item.Category] = true
		}
	}

	var topics []string
	for topic := range topicMap {
		topics = append(topics, topic)
	}

	// Sort for consistent ordering
	sort.Strings(topics)

	return topics
}

// GetCacheStats returns current cache statistics
func (nb *NewsBlogBackend) GetCacheStats() map[string]interface{} {
	nb.mu.RLock()
	defer nb.mu.RUnlock()

	return nb.cache.GetStats()
}

// ClearCache removes all cached news items
func (nb *NewsBlogBackend) ClearCache() {
	nb.mu.Lock()
	defer nb.mu.Unlock()

	nb.cache.Clear()

	if nb.debug {
		fmt.Println("[DEBUG] News cache cleared")
	}
}

// Shutdown gracefully stops all background operations (Phase 4)
func (nb *NewsBlogBackend) Shutdown() error {
	nb.mu.Lock()
	defer nb.mu.Unlock()

	if nb.debug {
		fmt.Println("[DEBUG] Shutting down news backend...")
	}

	// Stop background feed manager
	if nb.feedManager != nil {
		if err := nb.feedManager.Stop(); err != nil {
			return fmt.Errorf("failed to stop feed manager: %w", err)
		}
	}

	// Cancel context to stop all background operations
	if nb.cancel != nil {
		nb.cancel()
	}

	if nb.debug {
		fmt.Println("[DEBUG] News backend shutdown complete")
	}

	return nil
}
