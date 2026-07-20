package provider

import (
	"slices"
	"sync"
	"time"
)

// FetchCache memoizes DiscoveredModel lists per provider name so the
// "Add models" modal can re-filter without re-hitting the provider API.
type FetchCache struct {
	mu  sync.Mutex
	ttl time.Duration
	m   map[string]fetchEntry
}

type fetchEntry struct {
	models  []DiscoveredModel
	expires time.Time
}

// NewFetchCache returns a cache with the given TTL.
func NewFetchCache(ttl time.Duration) *FetchCache {
	return &FetchCache{ttl: ttl, m: map[string]fetchEntry{}}
}

// Get returns a copy of the cached list and true if present and not expired.
// Callers may mutate the returned slice freely; the cached entry is not
// affected (the cache stores its own copy via Set).
func (c *FetchCache) Get(provider string) ([]DiscoveredModel, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.m[provider]
	if !ok || time.Now().After(e.expires) {
		return nil, false
	}
	return slices.Clone(e.models), true
}

// Set stores a copy of the list with the configured TTL.
func (c *FetchCache) Set(provider string, models []DiscoveredModel) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[provider] = fetchEntry{models: slices.Clone(models), expires: time.Now().Add(c.ttl)}
}

// Invalidate drops one entry. Called when the Add-models modal closes.
func (c *FetchCache) Invalidate(provider string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.m, provider)
}
