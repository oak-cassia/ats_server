package service

import (
	"context"

	"auth_service/internal/model"
	"auth_service/internal/repository"
)

type UserRepository interface {
	GetUserByEmail(ctx context.Context, q repository.Queryer, email string) (*model.User, error)
	CreateUser(ctx context.Context, exec repository.Execer, user *model.User) error
	UpdateLastLogin(ctx context.Context, exec repository.Execer, user *model.User) error
}
