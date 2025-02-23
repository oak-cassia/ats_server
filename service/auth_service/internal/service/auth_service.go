package service

import (
	"auth_service/internal/model"
	"auth_service/internal/repository"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"redisclient"
	"time"
)

const tokenSize = 32
const sessionExpire = 24 * time.Hour

type AuthService struct {
	userRepo    *repository.UserRepository
	redisClient *redisclient.RedisClient
}

func NewAuthService(ur *repository.UserRepository, rc *redisclient.RedisClient) AuthService {
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

	user.LastLogin = time.Now()
	if err = s.userRepo.UpdateLastLogin(user); err != nil {
		return "", errors.New("failed to update last login")
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("invalid password")
	}

	token := generateToken()
	if token == "" {
		return "", errors.New("failed to generate token")
	}

	if err = setSession(email, token, s.redisClient); err != nil {
		return "", errors.New("failed to set session")
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

func setSession(email, token string, rc *redisclient.RedisClient) error {
	sk := redisclient.GetSessionKey(email)
	return rc.SetData(sk, token, sessionExpire)
}
