package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type RedisCache struct {
	Client *redis.Client
}

func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{
		Client: client,
	}
}

func (c *RedisCache) Manage(key string, value interface{}, expiration time.Duration) (interface{}, error) {
	getStatus := c.Client.Get(ctx, key)

	if getStatus.Err() == nil {
		return getStatus.Val(), nil
	}

	err := c.Client.Set(ctx, key, value, time.Hour).Err()
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (c *RedisCache) Del(ctx context.Context, key string) error {
	return c.Client.Del(ctx, key).Err()
}
