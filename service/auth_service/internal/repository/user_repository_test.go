package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	"auth_service/internal/model"
)

func TestUserRepository_CreateUser(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testUser := &model.User{
		ID:        1,
		Email:     "test",
		Password:  "test",
		Role:      "role", // TODO: role
		CreatedAt: time.Now(),
		LastLogin: time.Now(),
	}
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	mock.ExpectExec(
		`INSERT INTO account \(email, password, role, created_at, last_login\) VALUES \(\?, \?, \?, \?, \?\)`,
	).WithArgs(testUser.Email, testUser.Password, testUser.Role, testUser.CreatedAt, testUser.LastLogin).
		WillReturnResult(sqlmock.NewResult(1, 1))

	xdb := sqlx.NewDb(db, "mysql")
	r := &UserRepository{}
	if err = r.CreateUser(ctx, xdb, testUser); err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
}
