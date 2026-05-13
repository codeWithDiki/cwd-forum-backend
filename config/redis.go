package config

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func InitRedis() (*redis.Client, error) {
	err_env := godotenv.Load()

	if err_env != nil {
		return nil, err_env
	}

	ctx := context.Background()

	address := fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT"))
	password := os.Getenv("REDIS_PASSWORD")
	db := 0 // Default DB

	if os.Getenv("REDIS_DB") != "" {
		fmt.Sscanf(os.Getenv("REDIS_DB"), "%d", &db)
	}

	// Initialize Redis client
	redisClient := redis.NewClient(
		&redis.Options{
			Addr:     address,  // Redis server address
			Password: password, // No password set
			DB:       db,       // Use default DB
		},
	)

	pingStatus := redisClient.Ping(ctx)

	if pingStatus.Err() != nil {
		return nil, pingStatus.Err()
	}

	return redisClient, nil

}
