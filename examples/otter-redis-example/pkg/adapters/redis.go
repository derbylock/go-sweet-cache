package adapters

import (
	"context"
	"encoding"
	"fmt"
	"time"

	"github.com/derbylock/go-sweet-cache/pkg/sweet"
	"github.com/redis/go-redis/v9"
)

var _ sweet.Cacher[string, any] = &Redis[string, any]{}

type Redis[K comparable, V any] struct {
	cli         *redis.Client
	ctxProvider func(key any) context.Context
}

func (r *Redis[K, V]) GetOrProvide(ctx context.Context, key K, valueProvider sweet.ValueProvider[K, V]) (V, error) {
	v := *new(V)
	err := r.cli.Get(ctx, r.keyString(key)).Scan(v)
	if err != nil {
		v, actual, _, err := valueProvider(ctx, key)
		if err == nil {
			ee := r.cli.Set(ctx, r.keyString(key), v, actual).Err()
			if ee != nil {
				// report to monitoring
			}
		}
		return v, err
	}
	return v, nil
}

func (r *Redis[K, V]) GetOrProvideAsync(ctx context.Context, key K, valueProvider sweet.ValueProvider[K, V], defaultValue V) (V, error) {
	v := *new(V)
	err := r.cli.Get(ctx, r.keyString(key)).Scan(v)
	if err != nil {
		go func() {
			v, actual, _, err := valueProvider(ctx, key)
			_ = r.cli.Set(ctx, r.keyString(key), &CacheItem[V]{
				Value: v,
				Err:   err,
			}, actual).Err()
		}()
	}
	return v.Value, v.Err
}

func (r *Redis[K, V]) Get(key K) (V, bool, error) {
	return *new(V), false, nil
}

func (r *Redis[K, V]) Remove(key K) {
	_ = r.cli.Del(context.Background(), r.keyString(key)).Err()
}

func NewRedis[K comparable, V any](cli *redis.Client, ctxProvider func(key any) context.Context) *Redis[K, V] {
	return &Redis[K, V]{cli: cli, ctxProvider: ctxProvider}
}

func (r *Redis[K, V]) Clear() {
	// doing nothing, distributed cache can't be fully cleared
}

func (r *Redis[K, V]) SetWithTTL(key K, value V, ttl time.Duration) bool {
	ctx := r.ctxProvider(key)
	_, err := r.cli.Set(ctx, r.keyString(key), value, ttl*10).Result()
	if err != nil {
		return false
	}
	return true
}

func (r *Redis[K, V]) encode(value any) (string, error) {
	switch s := value.(type) {
	case string:
		return s, nil
	case []byte:
		return string(s), nil
	case encoding.BinaryMarshaler:
		b, err := s.MarshalBinary()
		if err != nil {
			return "", fmt.Errorf("marshal binary: %w")
		}
		return string(b), nil
	default:
		return fmt.Sprint("%v", value), nil
	}
}

func (r *Redis[K, V]) keyString(key any) string {
	k, err := r.encode(key)
	if err != nil {
		return fmt.Sprintf("%v", k)
	}
	return k
}
