package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"golang.org/x/sync/errgroup"

	"auth_service/config"
	"auth_service/internal/handler"
	"auth_service/internal/repository"
	"auth_service/internal/service"
)

const postMethod = "POST"

func main() {
	if len(os.Args) != 2 {
		log.Printf("need port number")
		os.Exit(1)
	}
	port := os.Args[1]
	l, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen port %s: %v", port, err)
	}
	if err := run(context.Background(), l); err != nil {
		log.Fatalf("failed to run: %v", err)
	}
}

func run(ctx context.Context, listener net.Listener) error {
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
		Handler: mux,
	}

	// 반환 값으로 오류를 받을 수 없어서 errgroup 패키지를 사용하여 오류를 반환
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		if err := s.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
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
