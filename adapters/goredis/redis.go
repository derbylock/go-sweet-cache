package goredis

import (
	"context"
	"encoding"
	"errors"
	"fmt"
	"time"

	"github.com/derbylock/go-sweet-cache/lib/v2/pkg/sweet"
	"github.com/redis/go-redis/v9"
)

var _ sweet.Cacher[any] = &Redis[any]{}

type Redis[V any] struct {
	cli        *redis.Client
	keyPrefix  string
	monitoring RedisClientErrorsMonitoring
}

type RedisClientErrorsMonitoring interface {
	GetFailed(ctx context.Context, key string, err error)
	SetFailed(ctx context.Context, key string, err error)
	RemoveFailed(ctx context.Context, key string, err error)
}

func (r *Redis[V]) GetOrProvide(ctx context.Context, key any, valueProvider sweet.ValueProvider[V]) (V, bool) {
	var v V
	var ok bool
	keyString := r.keyString(key)
	err := r.cli.Get(ctx, keyString).Scan(&v)
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			r.monitoring.GetFailed(ctx, keyString, err)
		}
		var actualTTL time.Duration
		ok, v, actualTTL, _ = valueProvider(ctx, key)
		if ok {
			err := r.cli.Set(ctx, r.keyString(key), &v, actualTTL).Err()
			if err != nil {
				r.monitoring.SetFailed(ctx, r.keyString(key), err)
			}
		}
	}
	return v, err == nil || ok
}

func (r *Redis[V]) GetOrProvideAsync(ctx context.Context, key any, valueProvider sweet.ValueProvider[V], defaultValue V) (V, bool) {
	var v V
	keyString := r.keyString(key)
	err := r.cli.Get(ctx, keyString).Scan(&v)
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			r.monitoring.GetFailed(ctx, keyString, err)
		}
		go func() {
			var actualTTL time.Duration
			var ok bool
			ok, v, actualTTL, _ = valueProvider(ctx, key)
			if ok {
				err = r.cli.Set(ctx, r.keyString(key), &v, actualTTL).Err()
				if err != nil {
					r.monitoring.SetFailed(ctx, r.keyString(key), err)
				}
			}
		}()
		return defaultValue, false
	}
	return v, true
}

func (r *Redis[V]) Get(ctx context.Context, key any) (V, bool) {
	// always return cache miss because we can't return value without remote calls
	return *new(V), false
}

func (r *Redis[V]) Remove(ctx context.Context, key any) {
	keyString := r.keyString(key)
	err := r.cli.Del(ctx, keyString).Err()
	if err != nil {
		r.monitoring.RemoveFailed(ctx, keyString, err)
	}
}

func NewRedis[V any](cli *redis.Client, keyPrefix string, monitoring RedisClientErrorsMonitoring) *Redis[V] {
	return &Redis[V]{cli: cli, keyPrefix: keyPrefix, monitoring: monitoring}
}

func (r *Redis[V]) Clear(ctx context.Context) {
	// doing nothing, distributed cache can't be fully cleared
}

func (r *Redis[V]) keyString(key any) string {
	k, err := r.encode(key)
	if err != nil {
		return fmt.Sprintf("%s%v", r.keyPrefix, k)
	}
	return r.keyPrefix + k
}

func (r *Redis[V]) encode(value any) (string, error) {
	switch s := value.(type) {
	case string:
		return s, nil
	case []byte:
		return string(s), nil
	case encoding.BinaryMarshaler:
		b, err := s.MarshalBinary()
		if err != nil {
			return "", fmt.Errorf("marshal binary: %w", err)
		}
		return string(b), nil
	default:
		return fmt.Sprintf("%v", value), nil
	}
}
