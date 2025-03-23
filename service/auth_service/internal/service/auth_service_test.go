package service

import (
	"context"
	"database/sql"
	"pkg/redisclient"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"auth_service/internal/model"
	"auth_service/internal/repository"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, q repository.Queryer, email string) (*model.User, error) {
	args := m.Called(ctx, q, email)
	if u := args.Get(0); u != nil {
		return u.(*model.User), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockUserRepository) CreateUser(ctx context.Context, exec repository.Execer, user *model.User) error {
	args := m.Called(ctx, exec, user)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, exec repository.Execer, user *model.User) error {
	args := m.Called(ctx, exec, user)
	return args.Error(0)
}

type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Save(ctx context.Context, key, token string, expire time.Duration) error {
	args := m.Called(ctx, key, token, expire)
	return args.Error(0)
}

func (m *MockRedisClient) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func TestRegisterUser_Success(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)
	xdb := sqlx.NewDb(db, "mysql")

	email := "test@example.com"
	password := "password123"

	ctx := context.WithValue(context.Background(), "time", time.Now())

	mockRepo := new(MockUserRepository)
	mockRepo.
		On("GetUserByEmail", ctx, xdb, email).
		Return(nil, nil)

	mockRepo.
		On("CreateUser", ctx, xdb, mock.AnythingOfType("*model.User")).
		Return(nil)

	service := NewAuthService(xdb, mockRepo, nil)
	err = service.RegisterUser(ctx, email, password)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestRegisterUser_UserExists(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)
	xdb := sqlx.NewDb(db, "mysql")

	email := "test@example.com"
	password := "password123"
	existingUser := &model.User{ID: 1, Email: email, Password: password, Role: "admin" /*TODO:role*/}

	ctx := context.Background()

	mockUserRepo := new(MockUserRepository)
	mockUserRepo.
		On("GetUserByEmail", ctx, xdb, email).
		Return(existingUser, nil)

	service := NewAuthService(xdb, mockUserRepo, nil)
	err = service.RegisterUser(ctx, email, password)
	assert.Error(t, err)
	assert.Equal(t, "user already exists", err.Error())
	mockUserRepo.AssertExpectations(t)
}

func TestLoginUser_Success(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	assert.NoError(t, err)
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)
	xdb := sqlx.NewDb(db, "mysql")

	email := "test@example.com"
	password := "password123"

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	assert.NoError(t, err)

	user := &model.User{
		ID:        1,
		Email:     email,
		Password:  string(hash),
		Role:      "admin", //TODO: role
		CreatedAt: time.Now(),
		LastLogin: time.Now(),
	}

	ctx := context.WithValue(context.Background(), "time", time.Now())

	mockUserRepo := new(MockUserRepository)
	mockUserRepo.
		On("GetUserByEmail", ctx, xdb, email).
		Return(user, nil)

	mockUserRepo.
		On("UpdateLastLogin", ctx, mock.Anything, user).
		Return(nil)

	sessionKey := redisclient.GetSessionKey(email)
	mockRedisClient := new(MockRedisClient)
	mockRedisClient.
		On("Save", ctx, sessionKey, mock.AnythingOfType("string"), sessionExpire).
		Return(nil)

	mockDB.ExpectBegin()
	mockDB.ExpectCommit()

	service := NewAuthService(xdb, mockUserRepo, mockRedisClient)
	token, err := service.LoginUser(ctx, email, password)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	mockUserRepo.AssertExpectations(t)
	mockRedisClient.AssertExpectations(t)
	assert.NoError(t, mockDB.ExpectationsWereMet())
}

func TestLoginUser_InvalidPassword(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)
	xdb := sqlx.NewDb(db, "mysql")

	email := "test@example.com"
	correctPassword := "correctpassword"
	wrongPassword := "wrongpassword"

	hash, err := bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.DefaultCost)
	assert.NoError(t, err)

	user := &model.User{
		ID:        1,
		Email:     email,
		Password:  string(hash),
		Role:      "admin", // TODO: role
		CreatedAt: time.Now(),
		LastLogin: time.Now(),
	}

	ctx := context.Background()

	mockUserRepo := new(MockUserRepository)
	mockUserRepo.
		On("GetUserByEmail", ctx, xdb, email).
		Return(user, nil)

	service := NewAuthService(xdb, mockUserRepo, nil)
	token, err := service.LoginUser(ctx, email, wrongPassword)
	assert.Error(t, err)
	assert.Equal(t, "invalid password", err.Error())
	assert.Empty(t, token)

	mockUserRepo.AssertExpectations(t)
}
func TestGenerateToken(t *testing.T) {
	token := generateToken()
	if len(token) != 44 {
		t.Errorf("token size is not 32, got: %d", len(token))
	}
}
