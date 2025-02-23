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

type Middleware func(http.HandlerFunc) http.HandlerFunc

func Method(m string) Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.Method != m {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			next(w, r)
		}
	}
}

func Chain(f http.HandlerFunc, middlewares ...Middleware) http.HandlerFunc {
	for _, m := range middlewares {
		f = m(f)
	}
	return f
}
func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	defer cfg.Close()

	userRepo := repository.NewUserRepository(cfg.DB)
	authService := service.NewAuthService(userRepo, cfg.Redis)
	authHandler := handler.NewAuthHandler(authService)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("hi")) })
	http.HandleFunc("/register", Chain(authHandler.RegisterHandler, Method(postMethod)))
	http.HandleFunc("/login", Chain(authHandler.LoginHandler, Method(postMethod)))

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
