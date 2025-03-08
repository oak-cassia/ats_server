package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
)

type Server struct {
	server   *http.Server
	listener net.Listener
}

func NewServer(l net.Listener, mux *http.ServeMux) *Server {
	return &Server{
		server:   &http.Server{Handler: mux},
		listener: l,
	}
}

func (s *Server) Run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 반환 값으로 오류를 받을 수 없어서 errgroup 패키지를 사용하여 오류를 반환
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		if err := s.server.Serve(s.listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("failed to close: %+v", err)
			return err
		}
		return nil
	})

	// 컨텍스트 취소 시 서버 종료
	<-ctx.Done()
	if err := s.server.Shutdown(context.Background()); err != nil {
		log.Printf("failed to shutdown: %+v", err)
	}

	return eg.Wait()
}
