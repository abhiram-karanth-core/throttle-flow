package redisclient

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

func New(ctx context.Context) *redis.Client {

	redisURL := os.Getenv("REDIS_ADDR")
	if redisURL == "" {
		log.Fatal("REDIS_ADDR not set")
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("Invalid REDIS_ADDR: %v", err)
	}

	client := redis.NewClient(opt)

	if err := client.Ping(context.Background()).Err(); err != nil {
		if err := client.Ping(ctx).Err(); err != nil {
			log.Fatalf("Redis connection failed: %v", err)
		}
	}
	return client
}
