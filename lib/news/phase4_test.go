package news

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestFeedManager(t *testing.T) {
	t.Run("new_feed_manager", func(t *testing.T) {
		fm := NewFeedManager()
		if fm == nil {
			t.Fatal("NewFeedManager() returned nil")
		}

		if fm.fetcher == nil {
			t.Error("FeedManager fetcher is nil")
		}

		if fm.cache == nil {
			t.Error("FeedManager cache is nil")
		}

		if fm.updateScheduler == nil {
			t.Error("FeedManager updateScheduler is nil")
		}

		if fm.errorTracker == nil {
			t.Error("FeedManager errorTracker is nil")
		}
	})

	t.Run("start_and_stop", func(t *testing.T) {
		fm := NewFeedManager()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Test start
		err := fm.Start(ctx)
		if err != nil {
			t.Fatalf("Failed to start FeedManager: %v", err)
		}

		// Test double start (should fail)
		err = fm.Start(ctx)
		if err == nil {
			t.Error("Expected error when starting already running FeedManager")
		}

		// Test stop
		err = fm.Stop()
		if err != nil {
			t.Fatalf("Failed to stop FeedManager: %v", err)
		}

		// Test double stop (should not error)
		err = fm.Stop()
		if err != nil {
			t.Errorf("Unexpected error on double stop: %v", err)
		}
	})

	t.Run("add_feed", func(t *testing.T) {
		fm := NewFeedManager()

		validFeed := RSSFeed{
			URL:        "https://example.com/feed.xml",
			Name:       "Test Feed",
			Category:   "tech",
			UpdateFreq: 30,
			MaxItems:   10,
			Enabled:    true,
		}

		// Test valid feed
		err := fm.AddFeed(validFeed)
		if err != nil {
			t.Errorf("Failed to add valid feed: %v", err)
		}

		// Test empty URL
		invalidFeed := validFeed
		invalidFeed.URL = ""
		err = fm.AddFeed(invalidFeed)
		if err == nil {
			t.Error("Expected error for feed with empty URL")
		}

		// Test disabled feed
		disabledFeed := validFeed
		disabledFeed.Enabled = false
		err = fm.AddFeed(disabledFeed)
		if err != nil {
			t.Errorf("Unexpected error for disabled feed: %v", err)
		}
	})

	t.Run("get_latest_news", func(t *testing.T) {
		fm := NewFeedManager()

		// Test with empty cache
		news, err := fm.GetLatestNews("tech", 5)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if len(news) != 0 {
			t.Errorf("Expected 0 news items, got %d", len(news))
		}
	})
}

func TestUpdateScheduler(t *testing.T) {
	t.Run("new_update_scheduler", func(t *testing.T) {
		scheduler := NewUpdateScheduler()
		if scheduler == nil {
			t.Fatal("NewUpdateScheduler() returned nil")
		}

		if scheduler.feeds == nil {
			t.Error("UpdateScheduler feeds map is nil")
		}
	})

	t.Run("add_feed", func(t *testing.T) {
		scheduler := NewUpdateScheduler()

		feed := RSSFeed{
			URL:        "https://example.com/feed.xml",
			Name:       "Test Feed",
			Category:   "tech",
			UpdateFreq: 30,
			MaxItems:   10,
			Enabled:    true,
		}

		err := scheduler.AddFeed(feed)
		if err != nil {
			t.Errorf("Failed to add feed: %v", err)
		}

		// Verify feed was added
		scheduler.mu.RLock()
		_, exists := scheduler.feeds[feed.URL]
		scheduler.mu.RUnlock()

		if !exists {
			t.Error("Feed was not added to scheduler")
		}
	})

	t.Run("get_next_feed", func(t *testing.T) {
		scheduler := NewUpdateScheduler()

		// Test with no feeds
		feed, duration := scheduler.GetNextFeed()
		if feed != nil {
			t.Error("Expected nil feed when no feeds are scheduled")
		}

		if duration <= 0 {
			t.Error("Expected positive duration")
		}

		// Add a feed that should be updated immediately
		testFeed := RSSFeed{
			URL:        "https://example.com/feed.xml",
			Name:       "Test Feed",
			Category:   "tech",
			UpdateFreq: 1, // 1 minute
			MaxItems:   10,
			Enabled:    true,
		}

		scheduler.AddFeed(testFeed)

		// Force next update to be in the past
		scheduler.mu.Lock()
		scheduled := scheduler.feeds[testFeed.URL]
		scheduled.NextUpdate = time.Now().Add(-1 * time.Minute)
		scheduler.feeds[testFeed.URL] = scheduled
		scheduler.mu.Unlock()

		feed, duration = scheduler.GetNextFeed()
		if feed == nil {
			t.Error("Expected feed to be returned for immediate update")
		}

		if duration <= 0 {
			t.Error("Expected positive duration")
		}
	})

	t.Run("record_update", func(t *testing.T) {
		scheduler := NewUpdateScheduler()

		testFeed := RSSFeed{
			URL:        "https://example.com/feed.xml",
			Name:       "Test Feed",
			Category:   "tech",
			UpdateFreq: 30,
			MaxItems:   10,
			Enabled:    true,
		}

		scheduler.AddFeed(testFeed)

		updateTime := time.Now()
		scheduler.RecordUpdate(testFeed.URL, updateTime)

		scheduler.mu.RLock()
		scheduled, exists := scheduler.feeds[testFeed.URL]
		scheduler.mu.RUnlock()

		if !exists {
			t.Fatal("Feed not found after recording update")
		}

		if scheduled.LastUpdate != updateTime {
			t.Error("LastUpdate was not recorded correctly")
		}

		if scheduled.UpdateCount != 1 {
			t.Errorf("Expected UpdateCount to be 1, got %d", scheduled.UpdateCount)
		}

		if scheduled.NextUpdate.Before(updateTime) {
			t.Error("NextUpdate should be after LastUpdate")
		}
	})
}

func TestErrorTracker(t *testing.T) {
	t.Run("new_error_tracker", func(t *testing.T) {
		tracker := NewErrorTracker()
		if tracker == nil {
			t.Fatal("NewErrorTracker() returned nil")
		}

		if tracker.feedErrors == nil {
			t.Error("ErrorTracker feedErrors map is nil")
		}
	})

	t.Run("record_attempt", func(t *testing.T) {
		tracker := NewErrorTracker()
		feedURL := "https://example.com/feed.xml"

		tracker.RecordAttempt(feedURL)

		tracker.mu.RLock()
		info, exists := tracker.feedErrors[feedURL]
		tracker.mu.RUnlock()

		if !exists {
			t.Fatal("Feed info not created after RecordAttempt")
		}

		if info.TotalAttempts != 1 {
			t.Errorf("Expected TotalAttempts to be 1, got %d", info.TotalAttempts)
		}
	})

	t.Run("record_error", func(t *testing.T) {
		tracker := NewErrorTracker()
		feedURL := "https://example.com/feed.xml"
		testError := &TestError{message: "test error"}

		tracker.RecordError(feedURL, testError)

		tracker.mu.RLock()
		info, exists := tracker.feedErrors[feedURL]
		tracker.mu.RUnlock()

		if !exists {
			t.Fatal("Feed info not created after RecordError")
		}

		if info.ConsecutiveErrors != 1 {
			t.Errorf("Expected ConsecutiveErrors to be 1, got %d", info.ConsecutiveErrors)
		}

		if info.TotalErrors != 1 {
			t.Errorf("Expected TotalErrors to be 1, got %d", info.TotalErrors)
		}

		if info.LastError != testError {
			t.Error("LastError was not recorded correctly")
		}

		if info.BackoffUntil.IsZero() {
			t.Error("BackoffUntil should be set after error")
		}
	})

	t.Run("record_success", func(t *testing.T) {
		tracker := NewErrorTracker()
		feedURL := "https://example.com/feed.xml"
		testError := &TestError{message: "test error"}

		// Record an error first
		tracker.RecordError(feedURL, testError)

		// Then record success
		tracker.RecordSuccess(feedURL)

		tracker.mu.RLock()
		info, exists := tracker.feedErrors[feedURL]
		tracker.mu.RUnlock()

		if !exists {
			t.Fatal("Feed info not found after RecordSuccess")
		}

		if info.ConsecutiveErrors != 0 {
			t.Errorf("Expected ConsecutiveErrors to be reset to 0, got %d", info.ConsecutiveErrors)
		}

		if info.BackoffUntil.IsZero() == false {
			t.Error("BackoffUntil should be cleared after success")
		}
	})

	t.Run("should_skip_feed", func(t *testing.T) {
		tracker := NewErrorTracker()
		feedURL := "https://example.com/feed.xml"

		// Should not skip new feed
		if tracker.ShouldSkipFeed(feedURL) {
			t.Error("Should not skip new feed")
		}

		// Record multiple errors to trigger backoff
		testError := &TestError{message: "test error"}
		for i := 0; i < 3; i++ {
			tracker.RecordError(feedURL, testError)
		}

		// Should skip during backoff period
		if !tracker.ShouldSkipFeed(feedURL) {
			t.Error("Should skip feed during backoff period")
		}
	})

	t.Run("get_feed_health", func(t *testing.T) {
		tracker := NewErrorTracker()
		feedURL := "https://example.com/feed.xml"

		// Test health for unknown feed
		health := tracker.GetFeedHealth(feedURL)
		if health.HealthScore != 100 {
			t.Errorf("Expected health score 100 for unknown feed, got %d", health.HealthScore)
		}

		if !health.IsHealthy {
			t.Error("Unknown feed should be considered healthy")
		}

		// Record some attempts and errors
		for i := 0; i < 5; i++ {
			tracker.RecordAttempt(feedURL)
		}

		testError := &TestError{message: "test error"}
		tracker.RecordError(feedURL, testError)

		health = tracker.GetFeedHealth(feedURL)
		if health.HealthScore < 0 || health.HealthScore > 100 {
			t.Errorf("Health score should be between 0-100, got %d", health.HealthScore)
		}

		if health.ErrorRate < 0 || health.ErrorRate > 1 {
			t.Errorf("Error rate should be between 0-1, got %f", health.ErrorRate)
		}
	})
}

// TestError is a simple error implementation for testing
type TestError struct {
	message string
}

func (e *TestError) Error() string {
	return e.message
}

func TestNewsCache_Phase4Extensions(t *testing.T) {
	t.Run("add_news_alias", func(t *testing.T) {
		cache := NewNewsCache(10)

		item := &NewsItem{
			ID:       "test-1",
			Title:    "Test News",
			URL:      "https://example.com/1",
			Category: "tech",
			Source:   "test-feed",
		}

		// Test AddNews method (alias for AddItem)
		cache.AddNews(item)

		// Verify item was added
		items, err := cache.GetLatestNews("tech", 5)
		if err != nil {
			t.Fatalf("GetLatestNews failed: %v", err)
		}

		if len(items) != 1 {
			t.Errorf("Expected 1 item, got %d", len(items))
		}

		if items[0].ID != "test-1" {
			t.Errorf("Expected item ID 'test-1', got '%s'", items[0].ID)
		}
	})

	t.Run("get_latest_news", func(t *testing.T) {
		cache := NewNewsCache(10)

		// Add test items
		for i := 0; i < 3; i++ {
			item := &NewsItem{
				ID:       fmt.Sprintf("test-%d", i),
				Title:    fmt.Sprintf("Test News %d", i),
				URL:      fmt.Sprintf("https://example.com/%d", i),
				Category: "tech",
				Source:   "test-feed",
			}
			cache.AddNews(item)
		}

		// Test getting all categories
		items, err := cache.GetLatestNews("", 5)
		if err != nil {
			t.Fatalf("GetLatestNews failed: %v", err)
		}

		if len(items) != 3 {
			t.Errorf("Expected 3 items for all categories, got %d", len(items))
		}

		// Test getting specific category
		items, err = cache.GetLatestNews("tech", 5)
		if err != nil {
			t.Fatalf("GetLatestNews failed: %v", err)
		}

		if len(items) != 3 {
			t.Errorf("Expected 3 items for tech category, got %d", len(items))
		}

		// Test limit
		items, err = cache.GetLatestNews("tech", 2)
		if err != nil {
			t.Fatalf("GetLatestNews failed: %v", err)
		}

		if len(items) != 2 {
			t.Errorf("Expected 2 items with limit, got %d", len(items))
		}
	})
}
