package news

import (
	"sync"
	"time"
)

// ErrorTracker manages feed reliability and implements error recovery policies
type ErrorTracker struct {
	feedErrors map[string]*FeedErrorInfo // Key: feed URL
	mu         sync.RWMutex
}

// FeedErrorInfo tracks error statistics and recovery state for a feed
type FeedErrorInfo struct {
	URL               string    // Feed URL
	ConsecutiveErrors int       // Number of consecutive errors
	LastError         error     // Most recent error
	LastErrorTime     time.Time // When the last error occurred
	LastSuccess       time.Time // When the feed last succeeded
	TotalAttempts     int       // Total update attempts
	TotalErrors       int       // Total errors encountered
	BackoffUntil      time.Time // When feed can be retried (exponential backoff)
}

// NewErrorTracker creates a new error tracking system
func NewErrorTracker() *ErrorTracker {
	return &ErrorTracker{
		feedErrors: make(map[string]*FeedErrorInfo),
	}
}

// RecordAttempt logs an update attempt for a feed
func (et *ErrorTracker) RecordAttempt(feedURL string) {
	et.mu.Lock()
	defer et.mu.Unlock()

	info, exists := et.feedErrors[feedURL]
	if !exists {
		info = &FeedErrorInfo{
			URL: feedURL,
		}
		et.feedErrors[feedURL] = info
	}

	info.TotalAttempts++
}

// RecordError logs an error for a feed and implements exponential backoff
func (et *ErrorTracker) RecordError(feedURL string, err error) {
	et.mu.Lock()
	defer et.mu.Unlock()

	info, exists := et.feedErrors[feedURL]
	if !exists {
		info = &FeedErrorInfo{
			URL: feedURL,
		}
		et.feedErrors[feedURL] = info
	}

	info.ConsecutiveErrors++
	info.TotalErrors++
	info.LastError = err
	info.LastErrorTime = time.Now()

	// Implement exponential backoff
	backoffDuration := et.calculateBackoffDuration(info.ConsecutiveErrors)
	info.BackoffUntil = time.Now().Add(backoffDuration)
}

// RecordSuccess logs a successful update and resets error counters
func (et *ErrorTracker) RecordSuccess(feedURL string) {
	et.mu.Lock()
	defer et.mu.Unlock()

	info, exists := et.feedErrors[feedURL]
	if !exists {
		info = &FeedErrorInfo{
			URL: feedURL,
		}
		et.feedErrors[feedURL] = info
	}

	info.ConsecutiveErrors = 0
	info.LastSuccess = time.Now()
	info.BackoffUntil = time.Time{} // Clear backoff
}

// ShouldSkipFeed determines if a feed should be temporarily skipped due to errors
func (et *ErrorTracker) ShouldSkipFeed(feedURL string) bool {
	et.mu.RLock()
	defer et.mu.RUnlock()

	info, exists := et.feedErrors[feedURL]
	if !exists {
		return false // No error history, allow update
	}

	// Skip if in backoff period
	if !info.BackoffUntil.IsZero() && time.Now().Before(info.BackoffUntil) {
		return true
	}

	// Skip if too many consecutive errors (circuit breaker)
	if info.ConsecutiveErrors >= 10 {
		// Disable feed for longer period after 10 consecutive errors
		if time.Since(info.LastErrorTime) < 24*time.Hour {
			return true
		}
		// Reset after 24 hours
		info.ConsecutiveErrors = 0
		et.feedErrors[feedURL] = info
	}

	return false
}

// GetFeedHealth returns health statistics for a feed
func (et *ErrorTracker) GetFeedHealth(feedURL string) FeedHealthStats {
	et.mu.RLock()
	defer et.mu.RUnlock()

	info, exists := et.feedErrors[feedURL]
	if !exists {
		return FeedHealthStats{
			URL:         feedURL,
			HealthScore: 100, // Perfect health for unknown feeds
			ErrorRate:   0,
			IsHealthy:   true,
		}
	}

	// Calculate health score (0-100)
	healthScore := et.calculateHealthScore(info)

	// Calculate error rate
	errorRate := 0.0
	if info.TotalAttempts > 0 {
		errorRate = float64(info.TotalErrors) / float64(info.TotalAttempts)
	}

	return FeedHealthStats{
		URL:               feedURL,
		HealthScore:       healthScore,
		ErrorRate:         errorRate,
		ConsecutiveErrors: info.ConsecutiveErrors,
		LastSuccess:       info.LastSuccess,
		LastError:         info.LastError,
		IsHealthy:         healthScore >= 70 && info.ConsecutiveErrors < 5,
	}
}

// FeedHealthStats provides health information about a feed
type FeedHealthStats struct {
	URL               string    // Feed URL
	HealthScore       int       // Health score (0-100)
	ErrorRate         float64   // Error rate (0.0-1.0)
	ConsecutiveErrors int       // Current consecutive error count
	LastSuccess       time.Time // Last successful update
	LastError         error     // Most recent error
	IsHealthy         bool      // Overall health status
}

// calculateBackoffDuration implements exponential backoff with jitter
func (et *ErrorTracker) calculateBackoffDuration(consecutiveErrors int) time.Duration {
	// Exponential backoff: 1min, 2min, 4min, 8min, 16min, 32min, max 1hour
	baseDelay := time.Minute
	for i := 1; i < consecutiveErrors && i < 6; i++ {
		baseDelay *= 2
	}

	// Cap at 1 hour
	if baseDelay > time.Hour {
		baseDelay = time.Hour
	}

	// Add jitter (±25%) to avoid thundering herd
	jitter := baseDelay / 4
	jitterSeconds := int(jitter.Seconds())
	if jitterSeconds > 0 {
		// Simple pseudo-random jitter based on current time
		jitterAmount := time.Now().UnixNano() % int64(jitterSeconds*2)
		baseDelay = baseDelay - jitter + time.Duration(jitterAmount)*time.Second
	}

	return baseDelay
}

// calculateHealthScore computes a health score (0-100) based on error history
func (et *ErrorTracker) calculateHealthScore(info *FeedErrorInfo) int {
	if info.TotalAttempts == 0 {
		return 100 // No attempts yet
	}

	// Base score on error rate
	successRate := float64(info.TotalAttempts-info.TotalErrors) / float64(info.TotalAttempts)
	baseScore := int(successRate * 100)

	// Penalize consecutive errors
	consecutivePenalty := info.ConsecutiveErrors * 10
	if consecutivePenalty > 50 {
		consecutivePenalty = 50 // Max 50 point penalty
	}

	// Bonus for recent success
	recentSuccessBonus := 0
	if !info.LastSuccess.IsZero() && time.Since(info.LastSuccess) < 24*time.Hour {
		recentSuccessBonus = 10
	}

	healthScore := baseScore - consecutivePenalty + recentSuccessBonus

	// Ensure score is within bounds
	if healthScore < 0 {
		healthScore = 0
	}
	if healthScore > 100 {
		healthScore = 100
	}

	return healthScore
}
