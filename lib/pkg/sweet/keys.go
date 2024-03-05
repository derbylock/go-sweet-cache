package sweet

import (
	"context"
	"errors"
	"fmt"
)

var InvalidCacheKeyErr = errors.New("cache key should be string or implement \"GetCacheKey() string\" method")

type CacheKeyGenerator interface {
	GetCacheKey() string
}

type GlobalCacheKeyGenerator struct {
	cacheKey string
}

func (g *GlobalCacheKeyGenerator) GetCacheKey() string {
	return g.cacheKey
}

func NewGlobalCacheKeyGenerator(namespace string, cacheName string, cacheKeySeparator string, key any) (CacheKeyGenerator, error) {
	fullKeyPrefix := fmt.Sprintf("%s%s%s%s", namespace, cacheKeySeparator, cacheName, cacheKeySeparator)
	return NewGlobalCacheKeyGeneratorFromPrefix(fullKeyPrefix, key)
}

func NewGlobalCacheKeyGeneratorFromPrefix(prefix string, key any) (CacheKeyGenerator, error) {
	var keyString string
	switch v := key.(type) {
	case string:
		keyString = v
	case CacheKeyGenerator:
		keyString = v.GetCacheKey()
	default:
		return nil, InvalidCacheKeyErr
	}
	return &GlobalCacheKeyGenerator{cacheKey: prefix + keyString}, nil
}

var _ Cacher[any] = &GlobalKeysCache[any]{}

type GlobalKeysCache[V any] struct {
	back      Cacher[V]
	keyPrefix string
}

func NewGlobalKeysCache[V any](namespace string, cacheName string, cacheKeySeparator string, back Cacher[V]) *GlobalKeysCache[V] {
	fullKeyPrefix := fmt.Sprintf("%s%s%s%s", namespace, cacheKeySeparator, cacheName, cacheKeySeparator)
	return &GlobalKeysCache[V]{back: back, keyPrefix: fullKeyPrefix}
}

func (t GlobalKeysCache[V]) GetOrProvide(ctx context.Context, key any, valueProvider ValueProvider[V]) (V, bool) {
	key, err := NewGlobalCacheKeyGeneratorFromPrefix(t.keyPrefix, key)
	if err != nil {
		return *new(V), false
	}
	return t.back.GetOrProvide(ctx, key, valueProvider)
}

func (t GlobalKeysCache[V]) GetOrProvideAsync(ctx context.Context, key any, valueProvider ValueProvider[V], defaultValue V) (V, bool) {
	return t.back.GetOrProvideAsync(ctx, key, valueProvider, defaultValue)
}

func (t GlobalKeysCache[V]) Get(ctx context.Context, key any) (V, bool) {
	return t.back.Get(ctx, key)
}

func (t GlobalKeysCache[V]) Remove(ctx context.Context, key any) {
	t.back.Remove(ctx, key)
}

func (t GlobalKeysCache[V]) Clear(ctx context.Context) {
	t.back.Clear(ctx)
}
