package news

import (
	"time"
)

// RSSFeed represents a single news source configuration
type RSSFeed struct {
	URL        string   `json:"url"`        // Feed URL (required)
	Name       string   `json:"name"`       // Display name for the feed
	Category   string   `json:"category"`   // "tech", "gaming", "general"
	UpdateFreq int      `json:"updateFreq"` // Minutes between updates
	MaxItems   int      `json:"maxItems"`   // Maximum items to store
	Keywords   []string `json:"keywords"`   // Filter keywords (optional)
	Enabled    bool     `json:"enabled"`    // Whether this feed is active
}

// NewsItem represents a single news article
type NewsItem struct {
	Title      string    `json:"title"`     // Article title
	Summary    string    `json:"summary"`   // Article summary/description
	URL        string    `json:"url"`       // Article URL
	Published  time.Time `json:"published"` // Publication timestamp
	Category   string    `json:"category"`  // Feed category
	Source     string    `json:"source"`    // Feed name
	ReadStatus bool      `json:"read"`      // Whether user has seen this item

	// For deduplication and caching
	ID string `json:"id"` // Unique identifier (typically URL or GUID)
}

// NewsConfig defines RSS/Atom newsfeed configuration for character cards
type NewsConfig struct {
	Enabled             bool        `json:"enabled"`             // Enable news features
	UpdateInterval      int         `json:"updateInterval"`      // Minutes between feed updates
	MaxStoredItems      int         `json:"maxStoredItems"`      // Maximum news items to keep in memory
	ReadingPersonality  string      `json:"readingPersonality"`  // "casual", "formal", "enthusiastic"
	PreferredCategories []string    `json:"preferredCategories"` // Preferred news categories
	Feeds               []RSSFeed   `json:"feeds"`               // List of RSS feeds
	ReadingEvents       []NewsEvent `json:"readingEvents"`       // News-specific events
}

// NewsEvent extends general dialog events for news-specific scenarios
type NewsEvent struct {
	Name             string   `json:"name"`             // Event identifier
	Category         string   `json:"category"`         // Event category (usually "conversation")
	Trigger          string   `json:"trigger"`          // How to trigger this event
	NewsCategory     string   `json:"newsCategory"`     // Which news category to read
	MaxNews          int      `json:"maxNews"`          // Maximum news items to mention
	IncludeSummary   bool     `json:"includeSummary"`   // Include article summaries
	ReadingStyle     string   `json:"readingStyle"`     // "casual", "formal", "enthusiastic"
	TriggerFrequency string   `json:"triggerFrequency"` // "daily", "hourly", "manual"
	Responses        []string `json:"responses"`        // Response templates
	Animations       []string `json:"animations"`       // Animations to play
	Cooldown         int      `json:"cooldown"`         // Seconds between triggers
}

// NewsCache manages news item storage and retrieval
type NewsCache struct {
	items       map[string]*NewsItem   // Key: item ID, Value: news item
	itemsByFeed map[string][]*NewsItem // Key: feed name, Value: items from that feed
	lastUpdate  map[string]time.Time   // Key: feed name, Value: last update time
	maxItems    int                    // Maximum items to store
}

// NewNewsCache creates a new news cache with the specified maximum items
func NewNewsCache(maxItems int) *NewsCache {
	return &NewsCache{
		items:       make(map[string]*NewsItem),
		itemsByFeed: make(map[string][]*NewsItem),
		lastUpdate:  make(map[string]time.Time),
		maxItems:    maxItems,
	}
}

// AddItem adds a news item to the cache, handling deduplication
func (nc *NewsCache) AddItem(item *NewsItem) {
	if item.ID == "" {
		// Use URL as fallback ID
		item.ID = item.URL
	}

	// Skip if already exists
	if _, exists := nc.items[item.ID]; exists {
		return
	}

	nc.items[item.ID] = item
	nc.itemsByFeed[item.Source] = append(nc.itemsByFeed[item.Source], item)

	// Enforce cache size limits
	nc.enforceMaxItems()
}

// GetItemsByCategory returns news items filtered by category
func (nc *NewsCache) GetItemsByCategory(category string, limit int) []*NewsItem {
	var result []*NewsItem
	for _, item := range nc.items {
		if item.Category == category || category == "headlines" {
			result = append(result, item)
		}
		if len(result) >= limit {
			break
		}
	}
	return result
}

// GetRecentItems returns the most recent news items across all feeds
func (nc *NewsCache) GetRecentItems(limit int) []*NewsItem {
	var result []*NewsItem
	for _, item := range nc.items {
		result = append(result, item)
	}

	// Sort by publication date (most recent first)
	// Using simple insertion sort for small datasets
	for i := 1; i < len(result); i++ {
		key := result[i]
		j := i - 1
		for j >= 0 && result[j].Published.Before(key.Published) {
			result[j+1] = result[j]
			j--
		}
		result[j+1] = key
	}

	if len(result) > limit {
		result = result[:limit]
	}
	return result
}

// UpdateFeedTimestamp records the last update time for a feed
func (nc *NewsCache) UpdateFeedTimestamp(feedName string) {
	nc.lastUpdate[feedName] = time.Now()
}

// GetLastUpdate returns the last update time for a feed
func (nc *NewsCache) GetLastUpdate(feedName string) time.Time {
	return nc.lastUpdate[feedName]
}

// enforceMaxItems removes oldest items if cache exceeds maximum size
func (nc *NewsCache) enforceMaxItems() {
	if len(nc.items) <= nc.maxItems {
		return
	}

	// Get all items sorted by publication date (oldest first)
	var allItems []*NewsItem
	for _, item := range nc.items {
		allItems = append(allItems, item)
	}

	// Sort by publication date (oldest first for removal)
	for i := 1; i < len(allItems); i++ {
		key := allItems[i]
		j := i - 1
		for j >= 0 && allItems[j].Published.After(key.Published) {
			allItems[j+1] = allItems[j]
			j--
		}
		allItems[j+1] = key
	}

	// Remove oldest items until we're under the limit
	itemsToRemove := len(allItems) - nc.maxItems
	for i := 0; i < itemsToRemove; i++ {
		item := allItems[i]
		delete(nc.items, item.ID)

		// Remove from feed-specific list
		feedItems := nc.itemsByFeed[item.Source]
		for j, feedItem := range feedItems {
			if feedItem.ID == item.ID {
				nc.itemsByFeed[item.Source] = append(feedItems[:j], feedItems[j+1:]...)
				break
			}
		}
	}
}

// GetStats returns cache statistics for monitoring
func (nc *NewsCache) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})
	stats["totalItems"] = len(nc.items)
	stats["feedCount"] = len(nc.itemsByFeed)
	stats["maxItems"] = nc.maxItems

	feedStats := make(map[string]int)
	for feedName, items := range nc.itemsByFeed {
		feedStats[feedName] = len(items)
	}
	stats["itemsByFeed"] = feedStats

	return stats
}

// Clear removes all items from the cache
func (nc *NewsCache) Clear() {
	nc.items = make(map[string]*NewsItem)
	nc.itemsByFeed = make(map[string][]*NewsItem)
	nc.lastUpdate = make(map[string]time.Time)
}

// AddNews is an alias for AddItem for compatibility with FeedManager
func (nc *NewsCache) AddNews(item *NewsItem) {
	nc.AddItem(item)
}

// GetLatestNews returns the most recent news items, optionally filtered by category
func (nc *NewsCache) GetLatestNews(category string, maxItems int) ([]*NewsItem, error) {
	if category == "" || category == "all" {
		return nc.GetRecentItems(maxItems), nil
	}

	return nc.GetItemsByCategory(category, maxItems), nil
}
