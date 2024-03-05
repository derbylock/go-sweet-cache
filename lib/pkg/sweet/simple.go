package sweet

import (
	"context"
	"time"

	"resenje.org/singleflight"
)

var _ Cacher[any] = &Cache[any]{}

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

type CacheItem[V any] struct {
	value  V
	ok     bool
	actual time.Time
}

type Cache[V any] struct {
	back SimpleCache
	sfg  *singleflight.Group[any, CacheItem[V]]
	now  func() time.Time
}

func NewCache[V any](back SimpleCache, now func() time.Time) *Cache[V] {
	return &Cache[V]{
		back: back,
		sfg:  &singleflight.Group[any, CacheItem[V]]{},
		now:  now,
	}
}

func (c Cache[V]) GetOrProvide(ctx context.Context, key any, valueProvider ValueProvider[V]) (V, bool) {
	now := c.now()
	if cachedVal, ok := c.back.Get(key); ok {
		if item, ok := cachedVal.(CacheItem[V]); ok {
			if now.Before(item.actual) {
				return item.value, item.ok
			}
			// it is not actual but usable
			go func() {
				c.updateValueFromProvider(ctx, key, valueProvider)
			}()
			return item.value, item.ok
		}
	}

	v, ok := c.updateValueFromProvider(ctx, key, valueProvider)
	return v, ok
}

func (c Cache[V]) updateValueFromProvider(
	ctx context.Context,
	key any,
	valueProvider ValueProvider[V],
) (V, bool) {
	singleItem, _, _ := c.sfg.Do(ctx, key, func(ctx context.Context) (CacheItem[V], error) {
		ok, v, actualTll, usableTtl := valueProvider(ctx, key)
		// use new now after the value provided
		now := c.now()
		item := CacheItem[V]{
			value:  v,
			ok:     ok,
			actual: now.Add(actualTll),
		}
		c.back.SetWithTTL(key, item, usableTtl)
		return item, nil
	})
	return singleItem.value, singleItem.ok
}

func (c Cache[V]) GetOrProvideAsync(
	ctx context.Context,
	key any,
	valueProvider ValueProvider[V],
	defaultValue V,
) (V, bool) {
	now := c.now()
	if cachedVal, ok := c.back.Get(key); ok {
		if item, ok := cachedVal.(CacheItem[V]); ok {
			if now.Before(item.actual) {
				return item.value, item.ok
			}
			// it is not actual but usable
			go func() {
				c.updateValueFromProvider(ctx, key, valueProvider)
			}()
			return item.value, item.ok
		}
	}

	go func() {
		c.updateValueFromProvider(ctx, key, valueProvider)
	}()
	return defaultValue, false
}

func (c Cache[V]) Get(ctx context.Context, key any) (V, bool) {
	if cachedVal, ok := c.back.Get(key); ok {
		if item, ok := cachedVal.(CacheItem[V]); ok {
			return item.value, item.ok
		}
	}
	return *new(V), false
}

func (c Cache[V]) Remove(ctx context.Context, key any) {
	c.back.Remove(key)
}

func (c Cache[V]) Clear(ctx context.Context) {
	c.back.Clear()
}
