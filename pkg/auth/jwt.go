package auth

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

//go:embed cert/secret.pem
var rawPrivateKey []byte

//go:embed cert/public.pem
var rawPublicKey []byte

type JWTConfig struct {
	Issuer     string
	ExpiresIn  time.Duration
	SignMethod jwa.SignatureAlgorithm
}

// User represents minimal user information for JWT token
type User struct {
	ID    int64
	Email string
	Role  string
}

// JWTManager handles JWT token generation and verification
type JWTManager struct {
	privateKey, publicKey jwk.Key
	config                JWTConfig
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(config JWTConfig) (*JWTManager, error) {
	manager := &JWTManager{
		config: config,
	}

	var err error
	manager.privateKey, err = parse(rawPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	manager.publicKey, err = parse(rawPublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	return manager, nil
}

func parse(rawKey []byte) (jwk.Key, error) {
	key, err := jwk.ParseKey(rawKey, jwk.WithPEM(true))
	if err != nil {
		return nil, err
	}
	return key, nil
}

// GenerateToken creates a new JWT token for the given user
func (j *JWTManager) GenerateToken(_ context.Context, u User) (string, error) {
	now := time.Now()
	tok, err := jwt.NewBuilder().
		Issuer(j.config.Issuer).
		IssuedAt(now).
		Expiration(now.Add(j.config.ExpiresIn)).
		Claim("sub", u.Email).
		Claim("user_id", u.ID).
		Claim("role", u.Role).
		Claim("jti", uuid.New().String()).
		Build()

	if err != nil {
		return "", fmt.Errorf("failed to build token: %w", err)
	}

	// JWT 토큰에 서명
	signed, err := jwt.Sign(tok, jwt.WithKey(j.config.SignMethod, j.privateKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return string(signed), nil
}

// VerifyToken verifies and parses the JWT token
func (j *JWTManager) VerifyToken(ctx context.Context, tokenString string) (jwt.Token, error) {
	tok, err := jwt.Parse(
		[]byte(tokenString),
		jwt.WithKey(j.config.SignMethod, j.publicKey),
		jwt.WithValidate(true),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}

	return tok, nil
}

// GetTokenKey returns Redis key for JWT token
func GetTokenKey(email string) string {
	return fmt.Sprintf("jwt:%s", email)
}
