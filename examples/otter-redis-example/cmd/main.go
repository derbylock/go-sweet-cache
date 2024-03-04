package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"otter-redis-example/pkg/adapters"

	"github.com/derbylock/go-sweet-cache/pkg/simple"
	"github.com/derbylock/go-sweet-cache/pkg/sweet"
	"github.com/maypok86/otter"
	"github.com/redis/go-redis/v9"
)

type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func (u *User) MarshalBinary() (data []byte, err error) {
	return json.Marshal(u)
}

func (u *User) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, u)
}

var cntExec = atomic.Int32{}

func main() {
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     "odin-redis.sbmt:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	redisCache := adapters.NewRedis[string, *User](rdb, func(key any) context.Context {
		return context.Background()
	})

	localCache := createLocalCache()
	cache := simple.NewCache[string, *User](localCache, time.Now)

	userProvider := func(ctx context.Context, key string) (
		val *User,
		err error,
	) {
		i := cntExec.Add(1)
		if i == 2 {
			return nil, errors.New("cached error")
		}

		return &User{
			Name: key,
			Age:  len(key) * 3,
		}, nil
	}

	cachedUserProvider := SimpleFixedTTLProvider(
		redisCache,
		time.Second*20,
		time.Second*5,
		userProvider,
	)

	v1, err := cache.GetOrProvide(ctx, "key1", cachedUserProvider)
	fmt.Printf("%v || %w || %d\n", v1, err, cntExec.Load())

	v1, err = cache.GetOrProvide(ctx, "key23", cachedUserProvider)
	fmt.Printf("%v || %w || %d\n", v1, err, cntExec.Load())

	v1, err = cache.GetOrProvide(ctx, "key1", cachedUserProvider)
	fmt.Printf("%v || %w || %d\n", v1, err, cntExec.Load())

	v1, err = cache.GetOrProvide(ctx, "key23", cachedUserProvider)
	fmt.Printf("%v || %w || %d\n", v1, err, cntExec.Load())

	time.Sleep(time.Second * 10)
	fmt.Println("After 10 sec")
	v1, err = cache.GetOrProvide(ctx, "key1", cachedUserProvider)
	fmt.Printf("%v || %w || %d\n", v1, err, cntExec.Load())

	v1, err = cache.GetOrProvide(ctx, "key23", cachedUserProvider)
	fmt.Printf("%v || %w || %d\n", v1, err, cntExec.Load())

	time.Sleep(time.Second)
	fmt.Println("After 1 sec")
	v1, err = cache.GetOrProvide(ctx, "key1", cachedUserProvider)
	fmt.Printf("%v || %w || %d\n", v1, err, cntExec.Load())

	v1, err = cache.GetOrProvide(ctx, "key23", cachedUserProvider)
	fmt.Printf("%v || %w || %d\n", v1, err, cntExec.Load())

	time.Sleep(time.Second * 30)
	fmt.Println("After 30 seconds")
	v1, err = cache.GetOrProvide(ctx, "key1", cachedUserProvider)
	fmt.Printf("%v || %w || %d\n", v1, err, cntExec.Load())

	v1, err = cache.GetOrProvide(ctx, "key23", cachedUserProvider)
	fmt.Printf("%v || %w || %d\n", v1, err, cntExec.Load())
}

func remoteCacheProvider[K comparable, V any](remoteCache sweet.Cacher[K, V], userProvider func(ctx context.Context, key K) (val V, actualTTL time.Duration, usableTTL time.Duration, err error)) func(ctx context.Context, key K) (val V, actualTTL time.Duration, usableTTL time.Duration, err error) {
	return func(ctx context.Context, key K) (
		val V,
		actualTTL time.Duration,
		usableTTL time.Duration,
		err error,
	) {
		var newActualTTL time.Duration
		var newUsableTTL time.Duration

		val, err = remoteCache.GetOrProvide(
			ctx,
			key,
			func(ctx context.Context, key K) (val V, actualTTL time.Duration, usableTTL time.Duration, err error) {
				val, newActualTTL, newUsableTTL, err = userProvider(ctx, key)
				// for remote cache usable ttl = actual ttl to avoid cache actuality prolongation
				return val, newActualTTL, newActualTTL, err
			},
		)
		return val, newActualTTL, newUsableTTL, err
	}
}

func createLocalCache() *adapters.Otter {
	// create a cache with capacity equal to 10000 elements
	otterCache, err := otter.MustBuilder[any, any](10_000).
		CollectStats().
		Cost(func(key any, value any) uint32 {
			return 1
		}).
		WithVariableTTL().
		Build()
	if err != nil {
		panic(err)
	}
	localCache := adapters.NewOtter(otterCache, time.Hour)
	return localCache
}

func FixedTTLProvider[K comparable, V any](
	cache sweet.Cacher[K, V],
	defaultActualTTL time.Duration,
	defaultUsableTTL time.Duration,
	defaultActualNegativeTTL time.Duration,
	defaultUsableNegativeTTL time.Duration,
	f func(ctx context.Context, key K) (V, error),
) sweet.ValueProvider[K, V] {
	return remoteCacheProvider[K, V](cache, func(ctx context.Context, key K) (
		val V,
		actualTTL time.Duration,
		usableTTL time.Duration,
		err error,
	) {
		v, e := f(ctx, key)
		if e != nil {
			return v, defaultActualNegativeTTL, defaultUsableNegativeTTL, e
		}
		return v, defaultActualTTL, defaultUsableTTL, e
	})
}

func SimpleFixedTTLProvider[K comparable, V any](
	cache sweet.Cacher[K, V],
	defaultUsableTTL time.Duration,
	defaultUsableNegativeTTL time.Duration,
	f func(ctx context.Context, key K) (V, error),
) sweet.ValueProvider[K, V] {
	return FixedTTLProvider(
		cache,
		defaultUsableTTL/2,
		defaultUsableTTL,
		defaultUsableNegativeTTL/2,
		defaultUsableNegativeTTL,
		f,
	)
}
