package service

import (
	"context"
	"time"
)

type RedisClient interface {
	Save(ctx context.Context, key, value string, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
}
