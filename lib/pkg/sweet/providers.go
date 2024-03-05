package sweet

import (
	"context"
	"time"
)

type SimpleValueProvider[V any] func(ctx context.Context, key any) (V, error)

func FixedTTLProvider[V any](
	defaultActualTTL time.Duration,
	defaultUsableTTL time.Duration,
	defaultActualNegativeTTL time.Duration,
	defaultUsableNegativeTTL time.Duration,
	simpleProvider SimpleValueProvider[V],
) ValueProvider[V] {
	return func(ctx context.Context, key any) (
		ok bool,
		val V,
		actualTTL time.Duration,
		usableTTL time.Duration,
	) {
		v, e := simpleProvider(ctx, key)
		if e != nil {
			return false, v, defaultActualNegativeTTL, defaultUsableNegativeTTL
		}
		return true, v, defaultActualTTL, defaultUsableTTL
	}
}

func SimpleFixedTTLProvider[V any](
	defaultUsableTTL time.Duration,
	defaultUsableNegativeTTL time.Duration,
	f func(ctx context.Context, key any) (V, error),
) ValueProvider[V] {
	return FixedTTLProvider(
		defaultUsableTTL/2,
		defaultUsableTTL,
		defaultUsableNegativeTTL/2,
		defaultUsableNegativeTTL,
		f,
	)
}

func (p *ValueProvider[V]) WithRemoteCache(
	remoteCache Cacher[V],
	baseProvider ValueProvider[V],
) ValueProvider[V] {
	return func(ctx context.Context, key any) (
		ok bool,
		val V,
		actualTTL time.Duration,
		usableTTL time.Duration,
	) {
		var newActualTTL time.Duration
		var newUsableTTL time.Duration

		val, ok = remoteCache.GetOrProvide(
			ctx,
			key,
			func(ctx context.Context, key any) (ok bool, val V, actualTTL time.Duration, usableTTL time.Duration) {
				ok, val, newActualTTL, newUsableTTL = baseProvider(ctx, key)
				return ok, val, newActualTTL, newUsableTTL
			},
		)
		return ok, val, newActualTTL, newUsableTTL
	}
}
