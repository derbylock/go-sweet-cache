package sweet

import (
	"context"
)

var _ Cacher[any] = &TwoLevelCache[any]{}

type TwoLevelCache[V any] struct {
	front Cacher[V]
	back  Cacher[V]
}

func NewTwoLevelCache[V any](front Cacher[V], back Cacher[V]) *TwoLevelCache[V] {
	return &TwoLevelCache[V]{front: front, back: back}
}

func (t TwoLevelCache[V]) GetOrProvide(ctx context.Context, key any, valueProvider ValueProvider[V]) (V, bool) {
	return t.front.GetOrProvide(ctx, key, valueProvider.WithRemoteCache(t.back, valueProvider))
}

func (t TwoLevelCache[V]) GetOrProvideAsync(ctx context.Context, key any, valueProvider ValueProvider[V], defaultValue V) (V, bool) {
	return t.front.GetOrProvideAsync(ctx, key, valueProvider.WithRemoteCache(t.back, valueProvider), defaultValue)
}

func (t TwoLevelCache[V]) Get(ctx context.Context, key any) (V, bool) {
	v, ok := t.front.Get(ctx, key)
	if !ok {
		v, ok = t.back.Get(ctx, key)
	}
	return v, ok
}

func (t TwoLevelCache[V]) Remove(ctx context.Context, key any) {
	t.front.Remove(ctx, key)
	t.back.Remove(ctx, key)
}

func (t TwoLevelCache[V]) Clear(ctx context.Context) {
	t.front.Clear(ctx)
	t.back.Clear(ctx)
}
