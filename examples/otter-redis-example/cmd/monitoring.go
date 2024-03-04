package main

import (
	"context"
	"fmt"
)

type LogMonitoring struct {
}

func (n LogMonitoring) GetFailed(ctx context.Context, key string, err error) {
	fmt.Printf("GetFailed key:%s err:%w\n", key, err)
}

func (n LogMonitoring) SetFailed(ctx context.Context, key string, err error) {
	fmt.Printf("SetFailed key:%s err:%w\n", key, err)
}

func (n LogMonitoring) RemoveFailed(ctx context.Context, key string, err error) {
	fmt.Printf("RemoveFailed key:%s err:%w\n", key, err)
}
