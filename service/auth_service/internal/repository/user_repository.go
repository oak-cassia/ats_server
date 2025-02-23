package repository

import (
	"auth_service/internal/model"
	"database/sql"
	"errors"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) CreateUser(user *model.User) error {
	query := "INSERT INTO account (email, password) VALUES (?, ?)"
	result, err := r.db.Exec(query, user.Email, user.Password)
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

func (r *UserRepository) GetUserByEmail(email string) (*model.User, error) {
	var user model.User
	query := "SELECT id, email, password FROM account WHERE email = ?"
	err := r.db.QueryRow(query, email).Scan(&user.ID, &user.Email, &user.Password)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("failed to find user")
	} else if err != nil {
		return nil, err
	}

	return &user, nil
}
