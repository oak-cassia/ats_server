package main

import (
	"auth_service/config"
	"auth_service/internal/handler"
	"auth_service/internal/repository"
	"auth_service/internal/service"
	"fmt"
	"log"
	"net/http"
)

const port = 10001
const postMethod = "POST"

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	defer cfg.Close()

	userRepo := repository.NewSqlUserRepository()
	authService := service.NewAuthService(cfg.DB, userRepo, cfg.Redis)
	authHandler := handler.NewAuthHandler(authService)

	http.HandleFunc("/register", Chain(authHandler.RegisterHandler, Method(postMethod)))
	http.HandleFunc("/login", Chain(authHandler.LoginHandler, Method(postMethod)))

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
