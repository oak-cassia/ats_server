package service

import (
	"auth_service/internal/model"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"redisclient"
	"time"
)

const tokenSize = 32
const sessionExpire = 24 * time.Hour

type AuthService struct {
	userRepo    UserRepository
	redisClient RedisClient
}

func NewAuthService(ur UserRepository, rc RedisClient) AuthService {
	return AuthService{
		userRepo:    ur,
		redisClient: rc,
	}
}

func (s *AuthService) RegisterUser(email, password string) error {
	existing, _ := s.userRepo.GetUserByEmail(email)
	if existing != nil {
		return errors.New("user already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &model.User{
		Email:     email,
		Password:  string(hash),
		CreatedAt: time.Now(),
		LastLogin: time.Now(),
	}
	if err = s.userRepo.CreateUser(user); err != nil {
		return errors.New("failed to create user")
	}

	return nil
}

func (s *AuthService) LoginUser(email, password string) (string, error) {
	user, err := s.userRepo.GetUserByEmail(email)
	if err != nil {
		return "", errors.New("user not found")
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("invalid password")
	}

	tx, err := s.userRepo.BeginTx()
	if err != nil {
		return "", errors.New("failed to begin transaction")
	}
	defer func(tx *sql.Tx) {
		_ = tx.Rollback()
	}(tx)

	user.LastLogin = time.Now()
	if err = s.userRepo.UpdateLastLoginTx(tx, user); err != nil {
		return "", errors.New("failed to update last login")
	}

	token := generateToken()
	if token == "" {
		return "", errors.New("failed to generate token")
	}

	if err = setSession(email, token, s.redisClient); err != nil {
		return "", errors.New("failed to set session")
	}

	if err = tx.Commit(); err != nil {
		return "", errors.New("failed to commit transaction")
	}

	return token, nil
}

func generateToken() string {
	b := make([]byte, tokenSize)
	if _, err := rand.Read(b); err != nil {
		return ""
	}

	return base64.StdEncoding.EncodeToString(b)
}

func setSession(email, token string, rc RedisClient) error {
	sk := redisclient.GetSessionKey(email)
	return rc.SetData(sk, token, sessionExpire)
}
