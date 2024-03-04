package sweet

import (
	"context"
	"fmt"
	"time"
)

type SimpleValueProvider[K comparable, V any] func(ctx context.Context, key K) (V, error)

func FixedTTLProvider[K comparable, V any](
	defaultActualTTL time.Duration,
	defaultUsableTTL time.Duration,
	defaultActualNegativeTTL time.Duration,
	defaultUsableNegativeTTL time.Duration,
	simpleProvider SimpleValueProvider[K, V],
) ValueProvider[K, V] {
	return func(ctx context.Context, key K) (
		val V,
		actualTTL time.Duration,
		usableTTL time.Duration,
		err error,
	) {
		v, e := simpleProvider(ctx, key)
		if e != nil {
			return v, defaultActualNegativeTTL, defaultUsableNegativeTTL, e
		}
		return v, defaultActualTTL, defaultUsableTTL, e
	}
}

func SimpleFixedTTLProvider[K comparable, V any](
	defaultUsableTTL time.Duration,
	defaultUsableNegativeTTL time.Duration,
	f func(ctx context.Context, key K) (V, error),
) ValueProvider[K, V] {
	return FixedTTLProvider(
		defaultUsableTTL/2,
		defaultUsableTTL,
		defaultUsableNegativeTTL/2,
		defaultUsableNegativeTTL,
		f,
	)
}

func (p *ValueProvider[K, V]) WithRemoteCache(
	remoteCache Cacher[K, V],
	baseProvider ValueProvider[K, V],
) ValueProvider[K, V] {
	return func(ctx context.Context, key K) (
		val V,
		actualTTL time.Duration,
		usableTTL time.Duration,
		err error,
	) {
		var newActualTTL time.Duration
		var newUsableTTL time.Duration

		var ok bool
		val, ok = remoteCache.GetOrProvide(
			ctx,
			key,
			func(ctx context.Context, key K) (val V, actualTTL time.Duration, usableTTL time.Duration, err error) {
				val, newActualTTL, newUsableTTL, err = baseProvider(ctx, key)
				return val, newActualTTL, newUsableTTL, err
			},
		)
		if !ok {
			err = fmt.Errorf("fail get value from remote cache")
		}
		return val, newActualTTL, newUsableTTL, err
	}
}
