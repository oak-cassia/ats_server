package service

import (
	"auth_service/internal/model"
	"auth_service/internal/repository"
	"context"
	"time"
)

type UserRepository interface {
	GetUserByEmail(ctx context.Context, q repository.Queryer, email string) (*model.User, error)
	CreateUser(ctx context.Context, exec repository.Execer, user *model.User) error
	UpdateLastLogin(ctx context.Context, exec repository.Execer, user *model.User) error
}

type RedisClient interface {
	Save(ctx context.Context, key, value string, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
}
