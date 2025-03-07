package main

import (
	"bytes"
	"context"
	"encoding/json"
	"golang.org/x/sync/errgroup"
	"log"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	// 테스트 실행 전에 작업 디렉토리를 프로젝트 루트로 변경
	if err := os.Chdir(".."); err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return run(ctx)
	})

	in := "login"
	var res struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	res.Email = "myEmail1@gmail.com"
	res.Password = "myPassword"

	reqBody, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	response, err := http.Post("http://localhost:10001/"+in, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}

	if response.StatusCode != http.StatusOK {
		t.Errorf("expected %d, got %d", http.StatusOK, response.StatusCode)
	}

	var resBody struct {
		Message string `json:"message"`
		Token   string `json:"token"`
	}

	if err := json.NewDecoder(response.Body).Decode(&resBody); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}

	if resBody.Message != "success login user" {
		t.Errorf("expected %s, got %s", "success login user", resBody.Message)
	}

	time.Sleep(1 * time.Second)
	cancel()

	if err := eg.Wait(); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}
