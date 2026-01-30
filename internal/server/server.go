package server

import (
	"github.com/go-redis/redis"
	_ "github.com/joho/godotenv/autoload"
)

type Server struct {
	rdb *redis.Client
}

func NewServer(rdb *redis.Client) *Server {
	return &Server{
		rdb: rdb,
	}
}
