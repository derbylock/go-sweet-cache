package main

import (
	"context"
	"encoding"
	"encoding/json"
	"time"

	"github.com/derbylock/go-sweet-cache/lib/v2/pkg/sweet"
)

type UserRepository struct {
}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

type GetUserParams struct {
	Name    string `json:"name"`
	Surname string `json:"surname"`
}

func (u *GetUserParams) MarshalBinary() (data []byte, err error) {
	return "userRepository:" + json.Marshal(u)
}

func (r *UserRepository) GetUser(ctx context.Context, params GetUserParams) (
	user User,
	err error,
) {
	cntExec.Add(1)
	return User{
		Name:    params.Name,
		Surname: params.Surname,
		Age:     len(params.Name)*3 + len(params.Surname)*2,
	}, nil
}

func Cached[K encoding.BinaryMarshaler, V any](
	cacheNamespace string,
	actualTTL time.Duration,
	negativeTTL time.Duration,
	provider func(ctx context.Context, params K) (V, error)) {
	cacheNamespace

	return sweet.SimpleFixedTTLProvider(
		time.Second*20,
		time.Second*5,
		provider,
	)
}
