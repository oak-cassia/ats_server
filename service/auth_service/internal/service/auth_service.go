package service

import (
	"auth_service/internal/model"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"redisclient"
	"time"
)

const tokenSize = 32
const sessionExpire = 24 * time.Hour

type AuthService struct {
	db          *sql.DB
	userRepo    UserRepository
	redisClient RedisClient
}

func NewAuthService(db *sql.DB, ur UserRepository, rc RedisClient) AuthService {
	return AuthService{
		db:          db,
		userRepo:    ur,
		redisClient: rc,
	}
}

func (s *AuthService) RegisterUser(email, password string) error {
	existing, _ := s.userRepo.GetUserByEmail(s.db, email)
	if existing != nil {
		return fmt.Errorf("user already exists")
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
	if err = s.userRepo.CreateUser(s.db, user); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (s *AuthService) LoginUser(email, password string) (string, error) {
	user, err := s.userRepo.GetUserByEmail(s.db, email)
	if err != nil {
		return "", fmt.Errorf("user not found: %w", err)
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", fmt.Errorf("invalid password")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func(tx *sql.Tx) {
		_ = tx.Rollback()
	}(tx)

	user.LastLogin = time.Now()
	if err = s.userRepo.UpdateLastLogin(tx, user); err != nil {
		return "", fmt.Errorf("failed to update last login: %w", err)
	}

	token := generateToken()
	if token == "" {
		return "", fmt.Errorf("failed to generate token")
	}

	if err = setSession(email, token, s.redisClient); err != nil {
		return "", fmt.Errorf("failed to set session: %w", err)
	}

	if err = tx.Commit(); err != nil {
		_ = s.deleteSession(email)
		return "", fmt.Errorf("failed to commit transaction: %w", err)
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

func (s *AuthService) deleteSession(email string) error {
	sk := redisclient.GetSessionKey(email)
	return s.redisClient.DelData(sk)
}
