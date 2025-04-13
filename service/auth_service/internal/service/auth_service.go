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
	"pkg/auth"
)

const tokenSize = 32
const sessionExpire = 24 * time.Hour

type AuthService struct {
	db          *sqlx.DB
	userRepo    UserRepository
	redisClient RedisClient
	jwtGen      JWTGenerator
}

func NewAuthService(db *sqlx.DB, ur UserRepository, rc RedisClient, jwtGen JWTGenerator) *AuthService {
	return &AuthService{
		db:          db,
		userRepo:    ur,
		redisClient: rc,
		jwtGen:      jwtGen,
	}
}

// GetTokenKey returns Redis key for JWT token
func GetTokenKey(email string) string {
	return fmt.Sprintf("jwt:%s", email)
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

	// JWT 토큰 생성
	tokenUser := auth.User{
		ID:    user.ID,
		Email: user.Email,
		Role:  user.Role,
	}

	token, err := s.jwtGen.GenerateToken(ctx, tokenUser)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	// Redis에 토큰 저장
	if err := s.redisClient.Save(
		ctx,
		GetTokenKey(email),
		token,
		sessionExpire,
	); err != nil {
		return "", fmt.Errorf("failed to store token: %w", err)
	}

	if err = tx.Commit(); err != nil {
		// 트랜잭션 실패 시 토큰 삭제
		_ = s.redisClient.Delete(ctx, GetTokenKey(email))
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	return token, nil
}

// 이 함수는 후방 호환성을 위해 남겨두지만 더 이상 사용되지 않습니다.
func generateToken() string {
	b := make([]byte, tokenSize)
	if _, err := rand.Read(b); err != nil {
		return ""
	}

	return base64.StdEncoding.EncodeToString(b)
}

// 토큰 폐기
func (s *AuthService) RevokeToken(ctx context.Context, email string) error {
	return s.redisClient.Delete(ctx, GetTokenKey(email))
}
