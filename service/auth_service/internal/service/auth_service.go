package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"

	"auth_service/internal/model"
	"pkg/redisclient"
)

const tokenSize = 32
const sessionExpire = 24 * time.Hour

type AuthService struct {
	db          *sqlx.DB
	userRepo    UserRepository
	redisClient RedisClient
}

func NewAuthService(db *sqlx.DB, ur UserRepository, rc RedisClient) *AuthService {
	return &AuthService{
		db:          db,
		userRepo:    ur,
		redisClient: rc,
	}
}

func (s *AuthService) RegisterUser(ctx context.Context, email, password string) error {
	existing, _ := s.userRepo.GetUserByEmail(ctx, s.db, email)
	if existing != nil {
		return fmt.Errorf("user already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	now := ctx.Value("time" /* TODO ctx key package*/).(time.Time)
	if now == (time.Time{}) {
		now = time.Now()
	}

	user := &model.User{
		Email:     email,
		Password:  string(hash),
		Role:      "admin", // TODO: role
		CreatedAt: now,
		LastLogin: now,
	}
	if err = s.userRepo.CreateUser(ctx, s.db, user); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (s *AuthService) LoginUser(ctx context.Context, email, password string) (string, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, s.db, email)
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

	now := ctx.Value("time" /* TODO ctx key package*/).(time.Time)
	if now == (time.Time{}) {
		now = time.Now()
	}

	user.LastLogin = now
	if err = s.userRepo.UpdateLastLogin(ctx, tx, user); err != nil {
		return "", fmt.Errorf("failed to update last login: %w", err)
	}

	token := generateToken()
	if token == "" {
		return "", fmt.Errorf("failed to generate token")
	}

	if err = s.setSession(ctx, email, token); err != nil {
		return "", fmt.Errorf("failed to set session: %w", err)
	}

	if err = tx.Commit(); err != nil {
		_ = s.deleteSession(ctx, email)
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

func (s *AuthService) setSession(ctx context.Context, email, token string) error {
	sk := redisclient.GetSessionKey(email)
	return s.redisClient.Save(ctx, sk, token, sessionExpire)
}

func (s *AuthService) deleteSession(ctx context.Context, email string) error {
	sk := redisclient.GetSessionKey(email)
	return s.redisClient.Delete(ctx, sk)
}
