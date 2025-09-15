// frame_cache.go: Image frame caching for animation performance optimization
// Implements LRU-based caching to reduce frame processing overhead and improve rendering performance

package performance

import (
	"container/list"
	"image"
	"sync"
)

// FrameCache provides cached access to processed animation frames.
// Uses LRU eviction policy to maintain memory efficiency while improving frame access performance.
// Designed to optimize the animation rendering pipeline by caching processed frames.
type FrameCache struct {
	mu       sync.RWMutex
	maxSize  int
	cache    map[string]*cacheEntry
	lruList  *list.List
	hitCount int64
	missCount int64
}

// cacheEntry holds a cached frame with LRU list element for efficient eviction
type cacheEntry struct {
	key       string
	frame     image.Image
	listElem  *list.Element
}

// FrameCacheStats provides cache performance metrics
type FrameCacheStats struct {
	HitCount  int64
	MissCount int64
	HitRatio  float64
	Size      int
	MaxSize   int
}

// NewFrameCache creates a new LRU frame cache with specified maximum size.
// maxSize determines how many frames to keep in memory before evicting least recently used frames.
func NewFrameCache(maxSize int) *FrameCache {
	if maxSize <= 0 {
		maxSize = 100 // Default to reasonable cache size
	}
	
	return &FrameCache{
		maxSize: maxSize,
		cache:   make(map[string]*cacheEntry),
		lruList: list.New(),
	}
}

// Get retrieves a cached frame by key, updating LRU position.
// Returns the cached frame and true if found, nil and false if not cached.
func (fc *FrameCache) Get(key string) (image.Image, bool) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	
	entry, exists := fc.cache[key]
	if !exists {
		fc.missCount++
		return nil, false
	}
	
	// Move to front of LRU list (most recently used)
	fc.lruList.MoveToFront(entry.listElem)
	fc.hitCount++
	
	return entry.frame, true
}

// Put stores a frame in the cache with the given key.
// If cache is full, evicts the least recently used frame to make space.
func (fc *FrameCache) Put(key string, frame image.Image) {
	if frame == nil {
		return // Don't cache nil frames
	}
	
	fc.mu.Lock()
	defer fc.mu.Unlock()
	
	// Check if key already exists
	if entry, exists := fc.cache[key]; exists {
		// Update existing entry and move to front
		entry.frame = frame
		fc.lruList.MoveToFront(entry.listElem)
		return
	}
	
	// Create new entry
	elem := fc.lruList.PushFront(key)
	entry := &cacheEntry{
		key:      key,
		frame:    frame,
		listElem: elem,
	}
	fc.cache[key] = entry
	
	// Evict oldest entries if over capacity
	fc.evictIfNeeded()
}

// evictIfNeeded removes least recently used entries when cache exceeds maxSize
func (fc *FrameCache) evictIfNeeded() {
	for fc.lruList.Len() > fc.maxSize {
		// Remove from back of list (least recently used)
		oldest := fc.lruList.Back()
		if oldest != nil {
			key := oldest.Value.(string)
			delete(fc.cache, key)
			fc.lruList.Remove(oldest)
		}
	}
}

// Clear removes all cached frames and resets statistics
func (fc *FrameCache) Clear() {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	
	fc.cache = make(map[string]*cacheEntry)
	fc.lruList.Init()
	fc.hitCount = 0
	fc.missCount = 0
}

// GetStats returns current cache performance statistics
func (fc *FrameCache) GetStats() FrameCacheStats {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	
	total := fc.hitCount + fc.missCount
	hitRatio := 0.0
	if total > 0 {
		hitRatio = float64(fc.hitCount) / float64(total)
	}
	
	return FrameCacheStats{
		HitCount:  fc.hitCount,
		MissCount: fc.missCount,
		HitRatio:  hitRatio,
		Size:      len(fc.cache),
		MaxSize:   fc.maxSize,
	}
}

// GetCacheKey generates a consistent cache key for animation frames.
// Combines animation name, frame index, and size for unique identification.
func GetCacheKey(animationName string, frameIndex int, width, height int) string {
	// Use efficient string formatting for cache key generation
	// Format: "animName:frameIdx:widthxheight"
	return animationName + ":" + 
		   intToString(frameIndex) + ":" + 
		   intToString(width) + "x" + 
		   intToString(height)
}

// intToString converts int to string efficiently for cache key generation
func intToString(n int) string {
	if n == 0 {
		return "0"
	}
	
	// Simple int to string conversion without fmt.Sprintf overhead
	negative := n < 0
	if negative {
		n = -n
	}
	
	var buf [20]byte // enough for 64-bit int
	i := len(buf)
	
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	
	if negative {
		i--
		buf[i] = '-'
	}
	
	return string(buf[i:])
}