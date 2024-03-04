package sweet

import "context"

type CacheMonitoring[K comparable] interface {
	Miss(ctx context.Context, key K)
	Hit(ctx context.Context, key K)
	GetFailed(ctx context.Context, key K, err error)
	PutFailed(ctx context.Context, key K, err error)
	RemoveFailed(ctx context.Context, key K)
	ClearFailed(ctx context.Context)
}

var _ CacheMonitoring[string] = NopCacheMonitoring[string]{}

type NopCacheMonitoring[K comparable] struct {
}

func (n NopCacheMonitoring[K]) Miss(ctx context.Context, key K) {
	// do nothing
}

func (n NopCacheMonitoring[K]) Hit(ctx context.Context, key K) {
	// do nothing
}

func (n NopCacheMonitoring[K]) GetFailed(ctx context.Context, key K, err error) {
	// do nothing
}

func (n NopCacheMonitoring[K]) PutFailed(ctx context.Context, key K, err error) {
	// do nothing
}

func (n NopCacheMonitoring[K]) RemoveFailed(ctx context.Context, key K) {
	// do nothing
}

func (n NopCacheMonitoring[K]) ClearFailed(ctx context.Context) {
	// do nothing
}
