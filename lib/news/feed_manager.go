package news

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// FeedManager manages background RSS feed updates with caching and performance optimization
type FeedManager struct {
	fetcher         *FeedFetcher
	cache           *NewsCache
	updateScheduler *UpdateScheduler
	errorTracker    *ErrorTracker
	mu              sync.RWMutex
	running         bool
	stopCh          chan struct{}
	wg              sync.WaitGroup
}

// NewFeedManager creates a new feed manager with performance optimization
func NewFeedManager() *FeedManager {
	fetcher := NewFeedFetcher(30 * time.Second) // 30s timeout for feeds

	return &FeedManager{
		fetcher:         fetcher,
		cache:           NewNewsCache(1000), // Cache up to 1000 news items
		updateScheduler: NewUpdateScheduler(),
		errorTracker:    NewErrorTracker(),
		stopCh:          make(chan struct{}),
	}
}

// Start begins background feed updating with smart scheduling
func (fm *FeedManager) Start(ctx context.Context) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if fm.running {
		return fmt.Errorf("feed manager already running")
	}

	fm.running = true

	// Start background update goroutine
	fm.wg.Add(1)
	go fm.backgroundUpdateLoop(ctx)

	return nil
}

// Stop gracefully shuts down background feed updating
func (fm *FeedManager) Stop() error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if !fm.running {
		return nil
	}

	fm.running = false
	close(fm.stopCh)

	// Wait for background goroutines to finish with timeout
	done := make(chan struct{})
	go func() {
		fm.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timeout waiting for feed manager to stop")
	}
}

// AddFeed registers a new RSS feed for background updating
func (fm *FeedManager) AddFeed(feed RSSFeed) error {
	if feed.URL == "" {
		return fmt.Errorf("feed URL cannot be empty")
	}

	if !feed.Enabled {
		return nil // Skip disabled feeds
	}

	fm.mu.Lock()
	defer fm.mu.Unlock()

	return fm.updateScheduler.AddFeed(feed)
}

// GetLatestNews retrieves the most recent news items with deduplication
func (fm *FeedManager) GetLatestNews(category string, maxItems int) ([]*NewsItem, error) {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	return fm.cache.GetLatestNews(category, maxItems)
}

// backgroundUpdateLoop runs the main feed update loop with smart scheduling
func (fm *FeedManager) backgroundUpdateLoop(ctx context.Context) {
	defer fm.wg.Done()

	// Initial update delay to avoid startup congestion
	updateTimer := time.NewTimer(5 * time.Second)
	defer updateTimer.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-fm.stopCh:
			return
		case <-updateTimer.C:
			// Get next feed to update based on smart scheduling
			feed, nextUpdate := fm.updateScheduler.GetNextFeed()
			if feed != nil {
				fm.updateFeedWithErrorHandling(*feed)
			}

			// Schedule next update
			updateTimer.Reset(nextUpdate)
		}
	}
}

// updateFeedWithErrorHandling updates a single feed with comprehensive error recovery
func (fm *FeedManager) updateFeedWithErrorHandling(feed RSSFeed) {
	// Track update attempt
	fm.errorTracker.RecordAttempt(feed.URL)

	// Check if feed should be temporarily disabled due to errors
	if fm.errorTracker.ShouldSkipFeed(feed.URL) {
		return
	}

	// Fetch feed with timeout and error handling
	newsItems, err := fm.fetcher.FetchFeed(feed)
	if err != nil {
		fm.errorTracker.RecordError(feed.URL, err)
		return
	}

	// Success - reset error count
	fm.errorTracker.RecordSuccess(feed.URL)

	// Store items in cache with deduplication
	for _, item := range newsItems {
		if item != nil {
			fm.cache.AddNews(item)
		}
	}

	// Update scheduler with successful fetch
	fm.updateScheduler.RecordUpdate(feed.URL, time.Now())
}
