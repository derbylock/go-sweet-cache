package simple

import (
	"context"
	"time"

	"github.com/derbylock/go-sweet-cache/lib/v2/pkg/sweet"
	"resenje.org/singleflight"
)

var _ sweet.Cacher[string, any] = &Cache[string, any]{}

// SimpleCache is an interface for a simple cache that can store key-value pairs.
type SimpleCache interface {
	// Set adds a key-value pair to the cache.
	// Returns false if key was not set because of some reason
	Set(key any, value any) bool

	// Get retrieves a value from the cache by its key.
	Get(key any) (any, bool)

	// Remove removes a key-value pair from the cache.
	Remove(key any)

	// Clear clears all key-value pairs from the cache.
	Clear()

	// SetWithTTL works like Set but adds a key-value pair to the cache that will expire after the specified TTL
	// (time to live) has passed. A zero value means the value never expires, which is identical to calling Set.
	// A negative value is a no-op and the value is discarded.
	SetWithTTL(key any, value any, ttl time.Duration) bool
}

type CacheItem[K comparable, V any] struct {
	value  V
	ok     bool
	actual time.Time
	usable time.Time
}

type Cache[K comparable, V any] struct {
	back SimpleCache
	sfg  *singleflight.Group[K, V]
	now  func() time.Time
}

func NewCache[K comparable, V any](back SimpleCache, now func() time.Time) *Cache[K, V] {
	return &Cache[K, V]{
		back: back,
		sfg:  &singleflight.Group[K, V]{},
		now:  now,
	}
}

func (c Cache[K, V]) GetOrProvide(ctx context.Context, key K, valueProvider sweet.ValueProvider[K, V]) (V, bool) {
	now := c.now()
	if cachedVal, ok := c.back.Get(key); ok {
		if item, ok := cachedVal.(CacheItem[K, V]); ok {
			if now.Before(item.actual) {
				return item.value, item.ok
			}
			if now.Before(item.usable) {
				go func() {
					c.updateValueFromProvider(ctx, key, valueProvider)
				}()
				return item.value, item.ok
			}
		}
	}

	v, _, err := c.updateValueFromProvider(ctx, key, valueProvider)
	return v, err == nil
}

func (c Cache[K, V]) updateValueFromProvider(
	ctx context.Context,
	key K,
	valueProvider sweet.ValueProvider[K, V],
) (V, bool, error) {
	return c.sfg.Do(ctx, key, func(ctx context.Context) (V, error) {
		v, actualTll, usableTtl, err := valueProvider(ctx, key)
		// use new now after the value provided
		now := c.now()
		item := CacheItem[K, V]{
			value:  v,
			ok:     err == nil,
			actual: now.Add(actualTll),
			usable: now.Add(usableTtl),
		}
		c.back.SetWithTTL(key, item, usableTtl)
		return v, err
	})
}

func (c Cache[K, V]) GetOrProvideAsync(
	ctx context.Context,
	key K,
	valueProvider sweet.ValueProvider[K, V],
	defaultValue V,
) (V, bool) {
	now := c.now()
	if cachedVal, ok := c.back.Get(key); ok {
		if item, ok := cachedVal.(CacheItem[K, V]); ok {
			if now.Before(item.actual) {
				return item.value, item.ok
			}
			if now.Before(item.usable) {
				go func() {
					c.updateValueFromProvider(ctx, key, valueProvider)
				}()
				return item.value, item.ok
			}
		}
	}

	go func() {
		c.updateValueFromProvider(ctx, key, valueProvider)
	}()
	return defaultValue, false
}

func (c Cache[K, V]) Get(ctx context.Context, key K) (V, bool) {
	now := c.now()
	if cachedVal, ok := c.back.Get(key); ok {
		if item, ok := cachedVal.(CacheItem[K, V]); ok {
			if now.Before(item.usable) {
				return item.value, item.ok
			}
		}
	}
	return *new(V), false
}

func (c Cache[K, V]) Remove(ctx context.Context, key K) {
	c.back.Remove(key)
}

func (c Cache[K, V]) Clear(ctx context.Context) {
	c.back.Clear()
}
