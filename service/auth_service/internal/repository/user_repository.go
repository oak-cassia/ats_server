package repository

import (
	"context"

	"auth_service/internal/model"
)

type UserRepository struct{}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

func (r *UserRepository) CreateUser(ctx context.Context, exec Execer, user *model.User) error {
	query := "INSERT INTO account (email, password, role, created_at, last_login) VALUES (?, ?, ?, ?, ?)"
	result, err := exec.ExecContext(ctx, query, user.Email, user.Password, user.Role, user.CreatedAt, user.LastLogin)
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

func (r *UserRepository) GetUserByEmail(ctx context.Context, q Queryer, email string) (*model.User, error) {
	var user model.User
	query := "SELECT id, email, password, role, created_at, last_login FROM account WHERE email = ?"
	err := q.GetContext(ctx, &user, query, email)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, exec Execer, user *model.User) error {
	query := "UPDATE account SET last_login = ? WHERE id = ?"
	_, err := exec.ExecContext(ctx, query, user.LastLogin, user.ID)
	return err
}
