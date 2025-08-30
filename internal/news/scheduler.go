package news

import (
	"sync"
	"time"
)

// UpdateScheduler manages smart feed update scheduling to minimize bandwidth usage
type UpdateScheduler struct {
	feeds     map[string]ScheduledFeed // Key: feed URL
	mu        sync.RWMutex
}

// ScheduledFeed tracks feed update scheduling information
type ScheduledFeed struct {
	Feed         RSSFeed   // Feed configuration
	LastUpdate   time.Time // When feed was last updated
	NextUpdate   time.Time // When feed should be updated next
	UpdateCount  int       // Number of successful updates
	Priority     int       // Update priority (1=highest, 5=lowest)
}

// NewUpdateScheduler creates a new smart update scheduler
func NewUpdateScheduler() *UpdateScheduler {
	return &UpdateScheduler{
		feeds: make(map[string]ScheduledFeed),
	}
}

// AddFeed registers a feed for scheduled updates with smart frequency calculation
func (us *UpdateScheduler) AddFeed(feed RSSFeed) error {
	us.mu.Lock()
	defer us.mu.Unlock()
	
	// Calculate initial priority based on category and update frequency
	priority := us.calculatePriority(feed)
	
	scheduled := ScheduledFeed{
		Feed:        feed,
		LastUpdate:  time.Time{}, // Never updated
		NextUpdate:  time.Now().Add(us.calculateInitialDelay(feed)), // Stagger initial updates
		UpdateCount: 0,
		Priority:    priority,
	}
	
	us.feeds[feed.URL] = scheduled
	return nil
}

// GetNextFeed returns the next feed to update and time until next check
func (us *UpdateScheduler) GetNextFeed() (*RSSFeed, time.Duration) {
	us.mu.RLock()
	defer us.mu.RUnlock()
	
	var nextFeed *RSSFeed
	var earliestTime time.Time
	now := time.Now()
	
	// Find the feed that should be updated next
	for _, scheduled := range us.feeds {
		if scheduled.NextUpdate.Before(now) || scheduled.NextUpdate.Equal(now) {
			if nextFeed == nil || scheduled.Priority < us.feeds[nextFeed.URL].Priority {
				feedCopy := scheduled.Feed
				nextFeed = &feedCopy
			}
		}
		
		// Track earliest next update time for scheduling
		if earliestTime.IsZero() || scheduled.NextUpdate.Before(earliestTime) {
			earliestTime = scheduled.NextUpdate
		}
	}
	
	// Calculate time until next check
	var nextCheckIn time.Duration
	if nextFeed != nil {
		nextCheckIn = 30 * time.Second // Quick recheck after update
	} else if !earliestTime.IsZero() {
		nextCheckIn = time.Until(earliestTime)
		if nextCheckIn < 30*time.Second {
			nextCheckIn = 30 * time.Second // Minimum interval
		}
	} else {
		nextCheckIn = 5 * time.Minute // Default when no feeds
	}
	
	return nextFeed, nextCheckIn
}

// RecordUpdate updates the scheduling information after a successful feed update
func (us *UpdateScheduler) RecordUpdate(feedURL string, updateTime time.Time) {
	us.mu.Lock()
	defer us.mu.Unlock()
	
	scheduled, exists := us.feeds[feedURL]
	if !exists {
		return
	}
	
	scheduled.LastUpdate = updateTime
	scheduled.UpdateCount++
	
	// Calculate next update time based on feed configuration and performance
	interval := us.calculateNextInterval(scheduled)
	scheduled.NextUpdate = updateTime.Add(interval)
	
	us.feeds[feedURL] = scheduled
}

// calculatePriority determines feed update priority based on category and frequency
func (us *UpdateScheduler) calculatePriority(feed RSSFeed) int {
	// Priority: 1=highest, 5=lowest
	switch feed.Category {
	case "breaking", "alerts":
		return 1 // Highest priority for breaking news
	case "tech", "gaming":
		return 2 // High priority for tech news
	case "general", "headlines":
		return 3 // Normal priority
	case "entertainment", "sports":
		return 4 // Lower priority
	default:
		return 5 // Lowest priority for unknown categories
	}
}

// calculateInitialDelay staggers initial feed updates to avoid startup congestion
func (us *UpdateScheduler) calculateInitialDelay(feed RSSFeed) time.Duration {
	// Stagger feeds by priority and hash of URL
	baseDelay := time.Duration(us.calculatePriority(feed)) * 30 * time.Second
	
	// Add small random component based on URL hash to spread updates
	urlHash := 0
	for _, char := range feed.URL {
		urlHash += int(char)
	}
	jitter := time.Duration(urlHash%60) * time.Second
	
	return baseDelay + jitter
}

// calculateNextInterval determines the optimal interval for the next update
func (us *UpdateScheduler) calculateNextInterval(scheduled ScheduledFeed) time.Duration {
	// Base interval from feed configuration
	baseInterval := time.Duration(scheduled.Feed.UpdateFreq) * time.Minute
	if baseInterval == 0 {
		baseInterval = 30 * time.Minute // Default 30 minutes
	}
	
	// Apply bandwidth-conscious policies
	minInterval := 15 * time.Minute // Never update more than every 15 minutes
	maxInterval := 4 * time.Hour    // Never wait more than 4 hours
	
	// Adjust based on update history (successful feeds can update more frequently)
	if scheduled.UpdateCount > 10 {
		baseInterval = baseInterval * 3 / 4 // 25% faster for reliable feeds
	} else if scheduled.UpdateCount < 3 {
		baseInterval = baseInterval * 5 / 4 // 25% slower for new/unreliable feeds
	}
	
	// Enforce limits
	if baseInterval < minInterval {
		baseInterval = minInterval
	}
	if baseInterval > maxInterval {
		baseInterval = maxInterval
	}
	
	return baseInterval
}
