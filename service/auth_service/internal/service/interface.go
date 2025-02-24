package service

import (
	"auth_service/internal/model"
	"database/sql"
	"time"
)

type UserRepository interface {
	GetUserByEmail(email string) (*model.User, error)
	CreateUser(user *model.User) error

	BeginTx() (*sql.Tx, error)
	UpdateLastLoginTx(tx *sql.Tx, user *model.User) error
}

type RedisClient interface {
	SetData(key, value string, expiration time.Duration) error
}
