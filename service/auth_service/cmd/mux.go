package main

import (
	"auth_service/internal/handler"
	"auth_service/internal/repository"
	"auth_service/internal/service"
	"net/http"
	"pkg/auth"
	"pkg/mysqlconn"
	"pkg/redisclient"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
)

func NewMux(mc *mysqlconn.MySQLConn, rc *redisclient.RedisClient) *http.ServeMux {
	userRepo := repository.NewUserRepository()

	// JWT 생성기 설정
	jwtConfig := auth.JWTConfig{
		Issuer:     "auth_service",
		ExpiresIn:  24 * time.Hour,
		SignMethod: jwa.RS256,
	}

	jwtGenerator, err := auth.NewJWTManager(jwtConfig)
	if err != nil {
		panic("failed to initialize JWT generator: " + err.Error())
	}

	authService := service.NewAuthService(mc.Conn(), userRepo, rc, jwtGenerator)
	authHandler := handler.NewAuthHandler(authService)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	})
	mux.HandleFunc("/register", Chain(
		authHandler.RegisterHandler,
		Method(postMethod),
		Timeout(5*time.Second),
		TimeNow(),
	))
	mux.HandleFunc("/login", Chain(
		authHandler.LoginHandler,
		Method(postMethod),
		Timeout(5*time.Second),
		TimeNow(),
	))
	return mux
}
