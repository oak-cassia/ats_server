package service

import (
	"time"
)

type RedisClient interface {
	SetData(key, value string, expiration time.Duration) error
	DelData(key string) error
}
