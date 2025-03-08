package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
)

func TestRun(t *testing.T) {
	assertions := assert.New(t)

	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to listen port: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "Hello")
	})

	ctx, cancel := context.WithCancel(context.Background())
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		s := NewServer(listener, mux)
		return s.Run(ctx)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	mux.ServeHTTP(res, req)
	assertions.Equal(http.StatusOK, res.Code)

	data, _ := io.ReadAll(res.Body)
	assertions.Equal("Hello", string(data))

	time.Sleep(1 * time.Second)
	cancel()

	if err := eg.Wait(); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}
