package service

import (
	"auth_service/internal/model"
	"auth_service/internal/repository"
	"context"
)

type UserRepository interface {
	GetUserByEmail(ctx context.Context, exec repository.SQLExecutor, email string) (*model.User, error)
	CreateUser(ctx context.Context, exec repository.SQLExecutor, user *model.User) error
	UpdateLastLogin(ctx context.Context, exec repository.SQLExecutor, user *model.User) error
}
