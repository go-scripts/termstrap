package termstrap

import (
	"image"
	"sync"
	"time"
)

// CachePolicy defines how images and rendered ANSI output are cached.
type CachePolicy int

const (
	// CacheDefault uses the cache with TTL expiration.
	CacheDefault CachePolicy = iota

	// CacheReload forces reloading remote images and recomputing ANSI rendering,
	// updating the cache with the new result.
	CacheReload

	// CacheNoStore bypasses the cache entirely (neither reading nor writing).
	CacheNoStore
)

// ImageCache defines the interface for caching raw images and rendered ANSI strings.
type ImageCache interface {
	GetImage(key string) (image.Image, bool)
	SetImage(key string, img image.Image, ttl time.Duration)
	GetANSI(key string) (string, bool)
	SetANSI(key string, rendered string, ttl time.Duration)
	Delete(key string)
	Clear()
}

type cacheEntry[T any] struct {
	value     T
	expiresAt time.Time
}

func (e cacheEntry[T]) isExpired() bool {
	if e.expiresAt.IsZero() {
		return false
	}
	return time.Now().After(e.expiresAt)
}

// MemoryImageCache is a thread-safe, in-memory implementation of ImageCache.
type MemoryImageCache struct {
	mu         sync.RWMutex
	images     map[string]cacheEntry[image.Image]
	ansi       map[string]cacheEntry[string]
	maxEntries int
	defaultTTL time.Duration
}

// NewMemoryImageCache creates a new in-memory cache.
func NewMemoryImageCache(maxEntries int, defaultTTL time.Duration) *MemoryImageCache {
	if maxEntries <= 0 {
		maxEntries = 1000
	}
	if defaultTTL <= 0 {
		defaultTTL = 2 * time.Hour
	}
	return &MemoryImageCache{
		images:     make(map[string]cacheEntry[image.Image]),
		ansi:       make(map[string]cacheEntry[string]),
		maxEntries: maxEntries,
		defaultTTL: defaultTTL,
	}
}

// Global default cache instance
var (
	defaultCacheMu sync.RWMutex
	defaultCache   = NewMemoryImageCache(1000, 2*time.Hour)
)

// DefaultCache returns the global default image cache instance.
func DefaultCache() ImageCache {
	defaultCacheMu.RLock()
	defer defaultCacheMu.RUnlock()
	return defaultCache
}

// SetDefaultCache sets a custom global default cache.
func SetDefaultCache(c ImageCache) {
	defaultCacheMu.Lock()
	defer defaultCacheMu.Unlock()
	defaultCache = c.(*MemoryImageCache)
}

// ClearImageCache clears all entries from the global default image cache.
func ClearImageCache() {
	defaultCache.Clear()
}

// InvalidateImage removes an image and its ANSI renders from the global cache.
func InvalidateImage(key string) {
	defaultCache.Delete(key)
}

// GetImage retrieves a decoded image from cache if not expired.
func (c *MemoryImageCache) GetImage(key string) (image.Image, bool) {
	c.mu.RLock()
	entry, ok := c.images[key]
	c.mu.RUnlock()

	if !ok || entry.isExpired() {
		if ok && entry.isExpired() {
			c.mu.Lock()
			delete(c.images, key)
			c.mu.Unlock()
		}
		return nil, false
	}
	return entry.value, true
}

// SetImage stores a decoded image in cache.
func (c *MemoryImageCache) SetImage(key string, img image.Image, ttl time.Duration) {
	if img == nil {
		return
	}
	if ttl <= 0 {
		ttl = c.defaultTTL
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.images) >= c.maxEntries {
		// Evict oldest entries when full
		count := 0
		for k := range c.images {
			delete(c.images, k)
			count++
			if count >= c.maxEntries/4 {
				break
			}
		}
	}

	c.images[key] = cacheEntry[image.Image]{
		value:     img,
		expiresAt: time.Now().Add(ttl),
	}
}

// GetANSI retrieves rendered ANSI content from cache if not expired.
func (c *MemoryImageCache) GetANSI(key string) (string, bool) {
	c.mu.RLock()
	entry, ok := c.ansi[key]
	c.mu.RUnlock()

	if !ok || entry.isExpired() {
		if ok && entry.isExpired() {
			c.mu.Lock()
			delete(c.ansi, key)
			c.mu.Unlock()
		}
		return "", false
	}
	return entry.value, true
}

// SetANSI stores rendered ANSI content in cache.
func (c *MemoryImageCache) SetANSI(key string, rendered string, ttl time.Duration) {
	if rendered == "" {
		return
	}
	if ttl <= 0 {
		ttl = c.defaultTTL
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.ansi) >= c.maxEntries {
		count := 0
		for k := range c.ansi {
			delete(c.ansi, k)
			count++
			if count >= c.maxEntries/4 {
				break
			}
		}
	}

	c.ansi[key] = cacheEntry[string]{
		value:     rendered,
		expiresAt: time.Now().Add(ttl),
	}
}

// Delete removes an image and any related ANSI renders from cache.
func (c *MemoryImageCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.images, key)
	for k := range c.ansi {
		if len(k) >= len(key) && k[:len(key)] == key {
			delete(c.ansi, k)
		}
	}
}

// Clear purges all cached images and ANSI entries.
func (c *MemoryImageCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.images = make(map[string]cacheEntry[image.Image])
	c.ansi = make(map[string]cacheEntry[string])
}

// NoopCache implements ImageCache with all caching disabled.
type NoopCache struct{}

func (NoopCache) GetImage(key string) (image.Image, bool)                  { return nil, false }
func (NoopCache) SetImage(key string, img image.Image, ttl time.Duration)  {}
func (NoopCache) GetANSI(key string) (string, bool)                       { return "", false }
func (NoopCache) SetANSI(key string, rendered string, ttl time.Duration) {}
func (NoopCache) Delete(key string)                                       {}
func (NoopCache) Clear()                                                  {}
