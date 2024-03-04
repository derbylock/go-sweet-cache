package sweet

import (
	"context"
)

var _ Cacher[string, any] = &TwoLevelCache[string, any]{}

type TwoLevelCache[K comparable, V any] struct {
	front Cacher[K, V]
	back  Cacher[K, V]
}

func NewTwoLevelCache[K comparable, V any](front Cacher[K, V], back Cacher[K, V]) *TwoLevelCache[K, V] {
	return &TwoLevelCache[K, V]{front: front, back: back}
}

func (t TwoLevelCache[K, V]) GetOrProvide(ctx context.Context, key K, valueProvider ValueProvider[K, V]) (V, bool) {
	return t.front.GetOrProvide(ctx, key, valueProvider.WithRemoteCache(t.back, valueProvider))
}

func (t TwoLevelCache[K, V]) GetOrProvideAsync(ctx context.Context, key K, valueProvider ValueProvider[K, V], defaultValue V) (V, bool) {
	return t.front.GetOrProvideAsync(ctx, key, valueProvider.WithRemoteCache(t.back, valueProvider), defaultValue)
}

func (t TwoLevelCache[K, V]) Get(ctx context.Context, key K) (V, bool) {
	v, ok := t.front.Get(ctx, key)
	if !ok {
		v, ok = t.back.Get(ctx, key)
	}
	return v, ok
}

func (t TwoLevelCache[K, V]) Remove(ctx context.Context, key K) {
	t.front.Remove(ctx, key)
	t.back.Remove(ctx, key)
}

func (t TwoLevelCache[K, V]) Clear(ctx context.Context) {
	t.front.Clear(ctx)
	t.front.Clear(ctx)
}
