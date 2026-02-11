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
	var script = redis.NewScript(`
	local current = redis.call("INCR", KEYS[1])
	if current == 1 then
	redis.call("PEXPIRE", KEYS[1], ARGV[1])
	end
	return current
	`)
	// count, err := l.rdb.Incr(ctx, redisKey).Result() //counter is incremented here.
	count, err := script.Run(ctx, l.rdb, []string{redisKey}, window.Milliseconds()).Int64()

	if err != nil {
		return false, 0, err
	}
	if count == 1 {
		if err := l.rdb.Expire(ctx, redisKey, window).Err(); err != nil {
			return false, 0, err
		}
	}
	remaining := max(0, limit-int(count))
	allowed := count <= int64(limit)
	return allowed, remaining, nil

}
