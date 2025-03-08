package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"pkg/mysqlconn"
	"pkg/redisclient"
	"time"

	"golang.org/x/sync/errgroup"

	"auth_service/config"
	"auth_service/internal/handler"
	"auth_service/internal/repository"
	"auth_service/internal/service"
)

const postMethod = "POST"

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	mc, err := mysqlconn.New(cfg.DbUser, cfg.DbPw, cfg.DbHost, cfg.DbName)
	if err != nil {
		log.Fatalf("failed to connect mc: %v", err)
	}
	rc := redisclient.New(cfg.RedisHost, cfg.RedisPw, 0)

	if err := run(cfg, context.Background(), mc, rc); err != nil {
		log.Fatalf("failed to run: %v", err)
	}
}

func run(cfg *config.Config, ctx context.Context, mc *mysqlconn.MySQLConn, rc *redisclient.RedisClient) error {
	userRepo := repository.NewSqlUserRepository()
	authService := service.NewAuthService(mc.Conn(), userRepo, rc)
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

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		log.Fatalf("failed to listen port %d: %v", cfg.Port, err)
	}

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
