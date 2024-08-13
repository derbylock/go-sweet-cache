package main

import (
	"context"
	"fmt"
	"time"

	"github.com/derbylock/go-sweet-cache/lib/v2/pkg/sweet"
)

type CachedUserRepository struct {
	repository *UserRepository
	getCache   sweet.Cacher[User]
}

func NewCachedUserRepository(repository *UserRepository, getCache sweet.Cacher[User]) *CachedUserRepository {
	return &CachedUserRepository{repository: repository, getCache: getCache}
}

func (r *CachedUserRepository) GetUser(ctx context.Context, params GetUserParams) (
	user User,
	err error,
) {
	u, ok := r.getCache.GetOrProvide(ctx, params, sweet.SimpleFixedTTLProvider[User](
		time.Second*20,
		time.Second*5,
		func(ctx context.Context, key any) (User, error) {
			u, err := r.repository.GetUser(ctx, params)
			return u, err
		},
	))
	if !ok {
		return User{}, fmt.Errorf("can't retrieve user")
	}
	return u, nil
}
