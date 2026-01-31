package limiter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type WindowLimiter struct {
	rdb *redis.Client
}

func NewWindowLimiter(rdb *redis.Client) *WindowLimiter {
	return &WindowLimiter{
		rdb: rdb,
	}
}	

func (l *WindowLimiter) Allow(
	ctx context.Context,
	key string,
	limit int,
	window time.Duration,
) (bool, int, error) {
	now := time.Now().UTC()
	windowStart := now.Truncate(window).Unix()
	redisKey := fmt.Sprintf("rl:%s:%d", key, windowStart)
	count, err := l.rdb.Incr(ctx, redisKey).Result() //counter is incremented here.
	if err != nil {
		return false, 0, err
	}
	if count == 1 {
		if err := l.rdb.Expire(ctx, redisKey, window).Err(); err != nil {
			return false, 0, err
		}
	}
	remaining := limit - int(count)
	if remaining < 0 {
		return false, 0, nil
	}
	return true, remaining, nil
}

