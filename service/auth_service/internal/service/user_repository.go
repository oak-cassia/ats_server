package service

import (
	"auth_service/internal/model"
	"auth_service/internal/repository"
)

type UserRepository interface {
	GetUserByEmail(exec repository.SQLExecutor, email string) (*model.User, error)
	CreateUser(exec repository.SQLExecutor, user *model.User) error
	UpdateLastLogin(exec repository.SQLExecutor, user *model.User) error
}
