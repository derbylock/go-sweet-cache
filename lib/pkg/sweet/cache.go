package sweet

import (
	"context"
	"time"
)

// ValueProvider is a function that provides a value for a given key, along with the actual and usable TTLs.
//
// The function takes a key of type K and returns a value of type V, along with the actual TTL and usable TTL,
// or an error if the value could not be provided.
//
// Context should be used by provider to properly process cancellation, e.g. because of timeout.
type ValueProvider[V any] func(ctx context.Context, key any) (ok bool, val V, actualTTL time.Duration, usableTTL time.Duration)

type Cacher[V any] interface {
	GetOrProvide(ctx context.Context, key any, valueProvider ValueProvider[V]) (V, bool)
	GetOrProvideAsync(ctx context.Context, key any, valueProvider ValueProvider[V], defaultValue V) (V, bool)
	Get(ctx context.Context, key any) (V, bool)
	Remove(ctx context.Context, key any)
	Clear(ctx context.Context)
}
