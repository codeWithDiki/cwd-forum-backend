package middleware

import (
	"gin-quickstart/pkg/cache"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func RedisCacheMiddleware(redis *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("redisCache", cache.NewRedisCache(redis))
		c.Next()
	}
}
