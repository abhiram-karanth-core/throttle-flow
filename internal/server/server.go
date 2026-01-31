package server

import (
	"throttle-flow/internal/limiter"

	_ "github.com/joho/godotenv/autoload"
	"github.com/redis/go-redis/v9"
)

type Server struct {
	limiter limiter.Limiter
}

func NewServer(rdb *redis.Client) *Server {
	return &Server{

		limiter: limiter.NewWindowLimiter(rdb),
	}
}
