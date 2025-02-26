package service

import (
	"auth_service/internal/model"
	"auth_service/internal/repository"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"redisclient"
	"testing"
	"time"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetUserByEmail(exec repository.SQLExecutor, email string) (*model.User, error) {
	args := m.Called(exec, email)
	if u := args.Get(0); u != nil {
		return u.(*model.User), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockUserRepository) CreateUser(exec repository.SQLExecutor, user *model.User) error {
	args := m.Called(exec, user)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateLastLogin(exec repository.SQLExecutor, user *model.User) error {
	args := m.Called(exec, user)
	return args.Error(0)
}

type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) SetData(key, token string, expire time.Duration) error {
	args := m.Called(key, token, expire)
	return args.Error(0)
}

func (m *MockRedisClient) DelData(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

func TestRegisterUser_Success(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	email := "test@example.com"
	password := "password123"

	mockRepo := new(MockUserRepository)
	mockRepo.
		On("GetUserByEmail", db, email).
		Return(nil, nil)

	mockRepo.
		On("CreateUser", db, mock.AnythingOfType("*model.User")).
		Return(nil)

	service := NewAuthService(db, mockRepo, nil)
	err = service.RegisterUser(email, password)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestRegisterUser_UserExists(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	email := "test@example.com"
	password := "password123"
	existingUser := &model.User{ID: 1, Email: email, Password: password}

	mockUserRepo := new(MockUserRepository)
	mockUserRepo.
		On("GetUserByEmail", db, email).
		Return(existingUser, nil)

	service := NewAuthService(db, mockUserRepo, nil)
	err = service.RegisterUser(email, password)
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

	email := "test@example.com"
	password := "password123"

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	assert.NoError(t, err)

	user := &model.User{
		ID:        1,
		Email:     email,
		Password:  string(hash),
		CreatedAt: time.Now(),
		LastLogin: time.Now(),
	}

	mockUserRepo := new(MockUserRepository)
	mockUserRepo.
		On("GetUserByEmail", db, email).
		Return(user, nil)

	mockUserRepo.
		On("UpdateLastLogin", mock.Anything, user).
		Return(nil)

	sessionKey := redisclient.GetSessionKey(email)
	mockRedisClient := new(MockRedisClient)
	mockRedisClient.
		On("SetData", sessionKey, mock.AnythingOfType("string"), sessionExpire).
		Return(nil)

	mockDB.ExpectBegin()
	mockDB.ExpectCommit()

	service := NewAuthService(db, mockUserRepo, mockRedisClient)
	token, err := service.LoginUser(email, password)
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

	email := "test@example.com"
	correctPassword := "correctpassword"
	wrongPassword := "wrongpassword"

	hash, err := bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.DefaultCost)
	assert.NoError(t, err)

	user := &model.User{
		ID:        1,
		Email:     email,
		Password:  string(hash),
		CreatedAt: time.Now(),
		LastLogin: time.Now(),
	}

	mockUserRepo := new(MockUserRepository)
	mockUserRepo.
		On("GetUserByEmail", db, email).
		Return(user, nil)

	service := NewAuthService(db, mockUserRepo, nil)
	token, err := service.LoginUser(email, wrongPassword)
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
