package auth

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisStore adapts redis.Client to the auth.Store interface used by AuthService.
type RedisStore struct {
	client *redis.Client
}

func NewRedisStore(c *redis.Client) *RedisStore {
	return &RedisStore{client: c}
}

func (r *RedisStore) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *RedisStore) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *RedisStore) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}
