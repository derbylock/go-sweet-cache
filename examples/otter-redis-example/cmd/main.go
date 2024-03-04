package main

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	adaptersRedis "github.com/derbylock/go-sweet-cache/adapters/goredis/v2"
	adaptersOtter "github.com/derbylock/go-sweet-cache/adapters/otter/v2"
	"github.com/derbylock/go-sweet-cache/lib/v2/pkg/simple"
	"github.com/derbylock/go-sweet-cache/lib/v2/pkg/sweet"
	"github.com/maypok86/otter"
	"github.com/redis/go-redis/v9"
)

var cntExec = atomic.Int32{}

func main() {
	ctx, userRepository, cache := initServices()

	getUserByNameCached := sweet.SimpleFixedTTLProvider(
		time.Second*20,
		time.Second*5,
		userRepository.getUserByNameAndSurname,
	)

	const key1 = "keyX1"
	const key2 = "keyX2"

	v1, err := cache.GetOrProvide(ctx, key1, getUserByNameCached)
	fmt.Printf("%v || %w || %d\n", v1, err, cntExec.Load())

	v1, err = cache.GetOrProvide(ctx, key2, getUserByNameCached)
	fmt.Printf("%v || %w || %d\n", v1, err, cntExec.Load())

	v1, err = cache.GetOrProvide(ctx, key1, getUserByNameCached)
	fmt.Printf("%v || %w || %d\n", v1, err, cntExec.Load())

	v1, err = cache.GetOrProvide(ctx, key2, getUserByNameCached)
	fmt.Printf("%v || %w || %d\n", v1, err, cntExec.Load())

	time.Sleep(time.Second * 10)
	fmt.Println("After 10 sec")
	v1, err = cache.GetOrProvide(ctx, key1, getUserByNameCached)
	fmt.Printf("%v || %w || %d\n", v1, err, cntExec.Load())

	v1, err = cache.GetOrProvide(ctx, key2, getUserByNameCached)
	fmt.Printf("%v || %w || %d\n", v1, err, cntExec.Load())

	time.Sleep(time.Second)
	fmt.Println("After 1 sec")
	v1, err = cache.GetOrProvide(ctx, key1, getUserByNameCached)
	fmt.Printf("%v || %w || %d\n", v1, err, cntExec.Load())

	v1, err = cache.GetOrProvide(ctx, key2, getUserByNameCached)
	fmt.Printf("%v || %w || %d\n", v1, err, cntExec.Load())

	time.Sleep(time.Second * 30)
	fmt.Println("After 30 seconds")
	v1, err = cache.GetOrProvide(ctx, key1, getUserByNameCached)
	fmt.Printf("%v || %w || %d\n", v1, err, cntExec.Load())

	v1, err = cache.GetOrProvide(ctx, key2, getUserByNameCached)
	fmt.Printf("%v || %w || %d\n", v1, err, cntExec.Load())
}

func initServices() (context.Context, *UserRepository, *simple.TwoLevelCache[string, User]) {
	ctx := context.Background()
	rdb := redisClient()
	userRepository := NewUserRepository()

	localCache := simple.NewCache[string, User](otterSimpleCache(), time.Now)
	redisCache := adaptersRedis.NewRedis[string, User](rdb, &LogMonitoring{})
	cache := simple.NewTwoLevelCache[string, User](localCache, redisCache)
	return ctx, userRepository, cache
}

func redisClient() *redis.Client {
	redisHost, ok := os.LookupEnv("REDIS_HOST")
	if !ok {
		redisHost = "localhost:6379"
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisHost,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	return rdb
}

func otterSimpleCache() *adaptersOtter.Otter {
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
	localCache := adaptersOtter.NewOtter(otterCache, time.Hour)
	return localCache
}
