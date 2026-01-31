package main

import (
	"context"
	"log"
	"net/http"
	redisclient "throttle-flow/internal/redis"
	"throttle-flow/internal/server"
	"time"
)

func main() {
	ctx := context.Background()
	rdb := redisclient.New(ctx)

	err := rdb.Set(ctx, "redis:test", "it_works", 5*time.Minute).Err()
	if err != nil {
		log.Fatal("Redis SET failed:", err)
	}

	val, err := rdb.Get(ctx, "redis:test").Result()
	if err != nil {
		log.Fatal("Redis GET failed:", err)
	}

	log.Println("Redis test value:", val)

	srv := server.NewServer(rdb)
	handler := server.Routes(srv)
	log.Println("rate-limiter running on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))

}
