package client

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisConnectLimiter enforces iTick's connection-rate limit across every
// service instance, not merely within one process.
type RedisConnectLimiter struct {
	rdb *redis.Client
	key string
}

func NewRedisConnectLimiter(rdb *redis.Client) *RedisConnectLimiter {
	return &RedisConnectLimiter{rdb: rdb, key: "itick:v1:ws:connect_rate"}
}

func (l *RedisConnectLimiter) Wait(ctx context.Context) error {
	const script = `
local nowParts = redis.call('TIME')
local now = nowParts[1] * 1000 + math.floor(nowParts[2] / 1000)
redis.call('ZREMRANGEBYSCORE', KEYS[1], '-inf', now - 1000)
if redis.call('ZCARD', KEYS[1]) < 2 then
  redis.call('ZADD', KEYS[1], now, ARGV[1])
  redis.call('PEXPIRE', KEYS[1], 2000)
  return 0
end
local oldest = redis.call('ZRANGE', KEYS[1], 0, 0, 'WITHSCORES')
return math.max(1, 1000 - (now - tonumber(oldest[2])))
`
	for {
		token, err := randomToken()
		if err != nil {
			return err
		}
		waitMs, err := l.rdb.Eval(ctx, script, []string{l.key}, token).Int64()
		if err != nil {
			return err
		}
		if waitMs <= 0 {
			return nil
		}
		timer := time.NewTimer(time.Duration(waitMs+50) * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}
