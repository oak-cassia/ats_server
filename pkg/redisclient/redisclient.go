package redisclient

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

func New(host, port, password string, db int) *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: password,
		DB:       db,
	})

	return &RedisClient{
		client: client,
	}
}

func (r *RedisClient) Save(ctx context.Context, key, value string, expire time.Duration) error {
	return r.client.Set(ctx, key, value, expire).Err()
}

func (r *RedisClient) Load(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *RedisClient) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func GetSessionKey(email string) string {
	return "session:" + email
}
