package main

import (
	"auth_service/internal/handler"
	"auth_service/internal/repository"
	"auth_service/internal/service"
	"net/http"
	"pkg/mysqlconn"
	"pkg/redisclient"
	"time"
)

func NewMux(mc *mysqlconn.MySQLConn, rc *redisclient.RedisClient) *http.ServeMux {
	userRepo := repository.NewSqlUserRepository()
	authService := service.NewAuthService(mc.Conn(), userRepo, rc)
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
