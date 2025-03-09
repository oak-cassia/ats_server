package repository

import (
	"context"

	"auth_service/internal/model"
)

type UserRepository struct{}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

func (r *UserRepository) CreateUser(ctx context.Context, exec SQLExecutor, user *model.User) error {
	query := "INSERT INTO account (email, password, created_at, last_login) VALUES (?, ?, ?, ?)"
	result, err := exec.ExecContext(ctx, query, user.Email, user.Password, user.CreatedAt, user.LastLogin)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	user.ID = id
	return nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, exec SQLExecutor, email string) (*model.User, error) {
	var user model.User
	query := "SELECT id, email, password, created_at, last_login FROM account WHERE email = ?"
	err := exec.QueryRowContext(ctx, query, email).Scan(&user.ID, &user.Email, &user.Password, &user.CreatedAt, &user.LastLogin)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, exec SQLExecutor, user *model.User) error {
	query := "UPDATE account SET last_login = ? WHERE id = ?"
	_, err := exec.ExecContext(ctx, query, user.LastLogin, user.ID)
	return err
}
