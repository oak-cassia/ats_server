package redisclient

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(addr, password string, db int) *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	})

	return &RedisClient{
		client: client,
	}
}

func (r *RedisClient) SetData(key, value string, expire time.Duration) error {
	ctx := context.Background()

	return r.client.Set(ctx, key, value, expire).Err()
}

func (r *RedisClient) GetData(key string) (string, error) {
	ctx := context.Background()
	return r.client.Get(ctx, key).Result()
}

func (r *RedisClient) DelData(key string) error {
	ctx := context.Background()
	return r.client.Del(ctx, key).Err()
}

func GetSessionKey(email string) string {
	return "session:" + email
}
