package news

import (
	"fmt"
	"testing"
	"time"
)

func TestNewsCache(t *testing.T) {
	cache := NewNewsCache(5) // Small cache for testing

	// Test adding items
	item1 := &NewsItem{
		ID:        "item1",
		Title:     "Test Article 1",
		URL:       "https://example.com/1",
		Published: time.Now().Add(-1 * time.Hour),
		Source:    "TestFeed",
		Category:  "tech",
	}

	item2 := &NewsItem{
		ID:        "item2",
		Title:     "Test Article 2", 
		URL:       "https://example.com/2",
		Published: time.Now(),
		Source:    "TestFeed",
		Category:  "gaming",
	}

	cache.AddItem(item1)
	cache.AddItem(item2)

	// Test retrieving recent items
	recent := cache.GetRecentItems(2)
	if len(recent) != 2 {
		t.Errorf("Expected 2 recent items, got %d", len(recent))
	}

	// Most recent should be first
	if recent[0].ID != "item2" {
		t.Errorf("Expected most recent item to be item2, got %s", recent[0].ID)
	}

	// Test category filtering
	techItems := cache.GetItemsByCategory("tech", 10)
	if len(techItems) != 1 || techItems[0].ID != "item1" {
		t.Errorf("Expected 1 tech item with ID item1, got %d items", len(techItems))
	}

	gamingItems := cache.GetItemsByCategory("gaming", 10)
	if len(gamingItems) != 1 || gamingItems[0].ID != "item2" {
		t.Errorf("Expected 1 gaming item with ID item2, got %d items", len(gamingItems))
	}

	// Test deduplication
	cache.AddItem(item1) // Add same item again
	recent = cache.GetRecentItems(10)
	if len(recent) != 2 {
		t.Errorf("Expected 2 items after deduplication, got %d", len(recent))
	}
}

func TestNewsCacheMaxItems(t *testing.T) {
	cache := NewNewsCache(2) // Very small cache

	// Add more items than the cache can hold
	for i := 0; i < 5; i++ {
		item := &NewsItem{
			ID:        fmt.Sprintf("item%d", i),
			Title:     fmt.Sprintf("Article %d", i),
			URL:       fmt.Sprintf("https://example.com/%d", i),
			Published: time.Now().Add(time.Duration(-i) * time.Hour),
			Source:    "TestFeed",
			Category:  "tech",
		}
		cache.AddItem(item)
	}

	// Should only have 2 items (the most recent ones)
	stats := cache.GetStats()
	totalItems := stats["totalItems"].(int)
	if totalItems != 2 {
		t.Errorf("Expected cache to contain 2 items, got %d", totalItems)
	}

	// Should have the most recent items (item0 and item1)
	recent := cache.GetRecentItems(10)
	if len(recent) != 2 {
		t.Errorf("Expected 2 recent items, got %d", len(recent))
	}

	// Most recent should be item0
	if recent[0].ID != "item0" {
		t.Errorf("Expected most recent item to be item0, got %s", recent[0].ID)
	}
}

func TestNewsCacheTimestamps(t *testing.T) {
	cache := NewNewsCache(10)

	// Test timestamp tracking
	feedName := "TestFeed"
	
	// Initially no timestamp
	lastUpdate := cache.GetLastUpdate(feedName)
	if !lastUpdate.IsZero() {
		t.Errorf("Expected zero timestamp for new feed, got %v", lastUpdate)
	}

	// Update timestamp
	cache.UpdateFeedTimestamp(feedName)
	
	// Should now have a recent timestamp
	lastUpdate = cache.GetLastUpdate(feedName)
	if lastUpdate.IsZero() {
		t.Errorf("Expected non-zero timestamp after update")
	}

	// Timestamp should be very recent (within last second)
	if time.Since(lastUpdate) > time.Second {
		t.Errorf("Expected timestamp to be very recent, got %v ago", time.Since(lastUpdate))
	}
}

func TestNewsCacheStats(t *testing.T) {
	cache := NewNewsCache(10)

	// Add items from different feeds
	item1 := &NewsItem{
		ID:     "item1",
		Source: "Feed1",
		Title:  "Article 1",
		URL:    "https://example.com/1",
		Published: time.Now(),
	}

	item2 := &NewsItem{
		ID:     "item2", 
		Source: "Feed2",
		Title:  "Article 2",
		URL:    "https://example.com/2",
		Published: time.Now(),
	}

	item3 := &NewsItem{
		ID:     "item3",
		Source: "Feed1", // Same feed as item1
		Title:  "Article 3",
		URL:    "https://example.com/3",
		Published: time.Now(),
	}

	cache.AddItem(item1)
	cache.AddItem(item2)
	cache.AddItem(item3)

	stats := cache.GetStats()

	// Check total items
	if stats["totalItems"].(int) != 3 {
		t.Errorf("Expected 3 total items, got %d", stats["totalItems"].(int))
	}

	// Check feed count
	if stats["feedCount"].(int) != 2 {
		t.Errorf("Expected 2 feeds, got %d", stats["feedCount"].(int))
	}

	// Check items by feed
	itemsByFeed := stats["itemsByFeed"].(map[string]int)
	if itemsByFeed["Feed1"] != 2 {
		t.Errorf("Expected 2 items from Feed1, got %d", itemsByFeed["Feed1"])
	}
	if itemsByFeed["Feed2"] != 1 {
		t.Errorf("Expected 1 item from Feed2, got %d", itemsByFeed["Feed2"])
	}
}

func TestNewsCacheClear(t *testing.T) {
	cache := NewNewsCache(10)

	// Add some items
	item := &NewsItem{
		ID:    "item1",
		Title: "Test Article",
		URL:   "https://example.com/1",
		Source: "TestFeed",
		Published: time.Now(),
	}
	cache.AddItem(item)

	// Verify item was added
	stats := cache.GetStats()
	if stats["totalItems"].(int) != 1 {
		t.Errorf("Expected 1 item before clear, got %d", stats["totalItems"].(int))
	}

	// Clear cache
	cache.Clear()

	// Verify cache is empty
	stats = cache.GetStats()
	if stats["totalItems"].(int) != 0 {
		t.Errorf("Expected 0 items after clear, got %d", stats["totalItems"].(int))
	}

	recent := cache.GetRecentItems(10)
	if len(recent) != 0 {
		t.Errorf("Expected no recent items after clear, got %d", len(recent))
	}
}
