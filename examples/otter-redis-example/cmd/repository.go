package main

import (
	"context"
	"encoding/json"
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
	return json.Marshal(u)
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
