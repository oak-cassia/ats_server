package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"pkg/mysqlconn"
	"pkg/redisclient"
)

func TestHealth(t *testing.T) {
	assertions := assert.New(t)

	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)

	mux := NewMux(&mysqlconn.MySQLConn{}, &redisclient.RedisClient{})
	mux.ServeHTTP(res, req)

	assertions.Equal(http.StatusOK, res.Code)

	data, _ := io.ReadAll(res.Body)
	assertions.Equal(`{"status": "ok"}`, string(data))
}
