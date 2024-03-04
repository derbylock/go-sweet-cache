package go_redis

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
	monitoring  RedisClientErrorsMonitoring
}

type RedisClientErrorsMonitoring interface {
	SetFailed(ctx context.Context, key string, err error)
	RemoveFailed(ctx context.Context, key string, err error)
}

func (r *Redis[K, V]) GetOrProvide(ctx context.Context, key K, valueProvider sweet.ValueProvider[K, V]) (V, bool) {
	v := *new(V)
	err := r.cli.Get(ctx, r.keyString(key)).Scan(v)
	if err != nil {
		var actualTTL time.Duration
		v, actualTTL, _, err = valueProvider(ctx, key)
		if err == nil {
			err := r.cli.Set(ctx, r.keyString(key), v, actualTTL).Err()
			if err != nil {
				r.monitoring.SetFailed(ctx, r.keyString(key), err)
			}
		}
	}
	return v, true
}

func (r *Redis[K, V]) GetOrProvideAsync(ctx context.Context, key K, valueProvider sweet.ValueProvider[K, V], defaultValue V) (V, bool) {
	v := *new(V)
	err := r.cli.Get(ctx, r.keyString(key)).Scan(v)
	if err != nil {
		go func() {
			var actualTTL time.Duration
			v, actualTTL, _, err = valueProvider(ctx, key)
			if err == nil {
				err = r.cli.Set(ctx, r.keyString(key), v, actualTTL).Err()
				if err != nil {
					r.monitoring.SetFailed(ctx, r.keyString(key), err)
				}
			}
		}()
		return defaultValue, false
	}
	return v, true
}

func (r *Redis[K, V]) Get(ctx context.Context, key K) (V, bool) {
	// always return cache miss because we can't return value without remote calls
	return *new(V), false
}

func (r *Redis[K, V]) Remove(ctx context.Context, key K) {
	keyString := r.keyString(key)
	err := r.cli.Del(context.Background(), keyString).Err()
	if err != nil {
		r.monitoring.RemoveFailed(ctx, keyString, err)
	}
}

func NewRedis[K comparable, V any](cli *redis.Client, ctxProvider func(key any) context.Context) *Redis[K, V] {
	return &Redis[K, V]{cli: cli, ctxProvider: ctxProvider}
}

func (r *Redis[K, V]) Clear(ctx context.Context) {
	// doing nothing, distributed cache can't be fully cleared
}

func (r *Redis[K, V]) keyString(key any) string {
	k, err := r.encode(key)
	if err != nil {
		return fmt.Sprintf("%v", k)
	}
	return k
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
