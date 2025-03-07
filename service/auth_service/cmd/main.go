package main

import (
	"auth_service/config"
	"auth_service/internal/handler"
	"auth_service/internal/repository"
	"auth_service/internal/service"
	"context"
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"
	"log"
	"net/http"
	"time"
)

const port = 10001
const postMethod = "POST"

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatalf("failed to run: %v", err)
	}
}

func run(ctx context.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	defer cfg.Close()

	userRepo := repository.NewSqlUserRepository()
	authService := service.NewAuthService(cfg.DB, userRepo, cfg.Redis)
	authHandler := handler.NewAuthHandler(authService)

	mux := http.NewServeMux()
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

	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	// 반환 값으로 오류를 받을 수 없어서 errgroup 패키지를 사용하여 오류를 반환
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("failed to close: %+v", err)
			return err
		}
		return nil
	})

	// 컨텍스트 취소 시 서버 종료
	<-ctx.Done()
	if err := s.Shutdown(context.Background()); err != nil {
		log.Printf("failed to shutdown: %+v", err)
	}

	return eg.Wait()
}
