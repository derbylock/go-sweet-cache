package adapters

import (
	"time"

	"github.com/derbylock/go-sweet-cache/pkg/simple"
	"github.com/maypok86/otter"
)

var _ simple.SimpleCache = &Otter{}

type Otter struct {
	back       otter.CacheWithVariableTTL[any, any]
	defaultTTL time.Duration
}

func NewOtter(back otter.CacheWithVariableTTL[any, any], defaultTTL time.Duration) *Otter {
	return &Otter{back: back, defaultTTL: defaultTTL}
}

func (o Otter) Set(key any, value any) bool {
	return o.back.Set(key, value, o.defaultTTL)
}

func (o Otter) Get(key any) (any, bool) {
	return o.back.Get(key)
}

func (o Otter) Remove(key any) {
	o.back.Delete(key)
}

func (o Otter) Clear() {
	o.back.Clear()
}

func (o Otter) SetWithTTL(key any, value any, ttl time.Duration) bool {
	return o.back.Set(key, value, ttl)
}
