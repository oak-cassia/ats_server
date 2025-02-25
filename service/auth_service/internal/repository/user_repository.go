package repository

import (
	"auth_service/internal/model"
	"database/sql"
	"errors"
)

type UserRepository struct{}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

func (r *UserRepository) CreateUser(exec SQLExecutor, user *model.User) error {
	query := "INSERT INTO account (email, password, created_at, last_login) VALUES (?, ?, ?, ?)"
	result, err := exec.Exec(query, user.Email, user.Password, user.CreatedAt, user.LastLogin)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	user.ID = uint(id)
	return nil
}

func (r *UserRepository) GetUserByEmail(exec SQLExecutor, email string) (*model.User, error) {
	var user model.User
	query := "SELECT id, email, password, created_at, last_login FROM account WHERE email = ?"
	err := exec.QueryRow(query, email).Scan(&user.ID, &user.Email, &user.Password, &user.CreatedAt, &user.LastLogin)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("failed to find user")
	}

	return &user, nil
}

func (r *UserRepository) UpdateLastLogin(exec SQLExecutor, user *model.User) error {
	query := "UPDATE account SET last_login = ? WHERE id = ?"
	_, err := exec.Exec(query, user.LastLogin, user.ID)
	return err
}
