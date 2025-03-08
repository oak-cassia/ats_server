package main

import (
	"auth_service/config"
	"context"
	"fmt"
	"log"
	"net"
	"pkg/mysqlconn"
	"pkg/redisclient"
)

const postMethod = "POST"

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatalf("failed to run: %v", err)
	}
}

func run(ctx context.Context) error {
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		log.Fatalf("failed to listen port %d: %v", cfg.Port, err)
	}

	mc, err := mysqlconn.New(cfg.DbUser, cfg.DbPw, cfg.DbHost, cfg.DbPort, cfg.DbName)
	if err != nil {
		log.Fatalf("failed to connect mc: %v", err)
	}
	rc := redisclient.New(cfg.RedisHost, cfg.RedisPw, 0)
	mux := NewMux(mc, rc)

	s := NewServer(listener, mux)
	return s.Run(ctx)
}
