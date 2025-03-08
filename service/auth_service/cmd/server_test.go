package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
)

func TestRun(t *testing.T) {
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

	url := fmt.Sprintf("http://%s/test", listener.Addr().String())
	t.Logf("try request to %s", url)

	response, err := http.Get(url)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	if response.StatusCode != http.StatusOK {
		t.Errorf("expected %d, got %d", http.StatusOK, response.StatusCode)
	}

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	bodyStr := string(bodyBytes)
	if bodyStr != "Hello" {
		t.Errorf("expected response body %q, got %q", "Hello", bodyStr)
	}

	time.Sleep(1 * time.Second)
	cancel()

	if err := eg.Wait(); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}
