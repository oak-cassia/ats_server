package service

import (
	"context"
	"time"
)

type RedisClient interface {
	SetData(ctx context.Context, key, value string, expiration time.Duration) error
	DelData(ctx context.Context, key string) error
}
