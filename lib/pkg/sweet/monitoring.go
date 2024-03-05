package sweet

import "context"

type CacheMonitoring interface {
	Miss(ctx context.Context, key any)
	Hit(ctx context.Context, key any)
	GetFailed(ctx context.Context, key any, err error)
	PutFailed(ctx context.Context, key any, err error)
	RemoveFailed(ctx context.Context, key any)
	ClearFailed(ctx context.Context)
}

var _ CacheMonitoring = NopCacheMonitoring{}

type NopCacheMonitoring struct {
}

func (n NopCacheMonitoring) Miss(ctx context.Context, key any) {
	// do nothing
}

func (n NopCacheMonitoring) Hit(ctx context.Context, key any) {
	// do nothing
}

func (n NopCacheMonitoring) GetFailed(ctx context.Context, key any, err error) {
	// do nothing
}

func (n NopCacheMonitoring) PutFailed(ctx context.Context, key any, err error) {
	// do nothing
}

func (n NopCacheMonitoring) RemoveFailed(ctx context.Context, key any) {
	// do nothing
}

func (n NopCacheMonitoring) ClearFailed(ctx context.Context) {
	// do nothing
}
