package itick

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisConnectLimiter enforces iTick's connection-rate limit across every
// service instance, not merely within one process.
type RedisConnectLimiter struct {
	rdb         *redis.Client
	gateKey     string
	coolDownKey string
}

func NewRedisConnectLimiter(rdb *redis.Client) *RedisConnectLimiter {
	return &RedisConnectLimiter{
		rdb: rdb, gateKey: "market:v1:ws:connect_gate", coolDownKey: "market:v1:ws:connect_cool_down",
	}
}

func (l *RedisConnectLimiter) Wait(ctx context.Context) error {
	for {
		if ttl, err := l.rdb.PTTL(ctx, l.coolDownKey).Result(); err != nil && err != redis.Nil {
			return err
		} else if ttl > 0 {
			if err := waitContext(ctx, ttl+100*time.Millisecond); err != nil {
				return err
			}
			continue
		}
		token, err := randomToken()
		if err != nil {
			return err
		}
		ok, err := l.rdb.SetNX(ctx, l.gateKey, token, 3*time.Second).Result()
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
		ttl, err := l.rdb.PTTL(ctx, l.gateKey).Result()
		if err != nil && err != redis.Nil {
			return err
		}
		if ttl <= 0 {
			ttl = 500 * time.Millisecond
		}
		if err := waitContext(ctx, ttl+100*time.Millisecond); err != nil {
			return err
		}
	}
}

func (l *RedisConnectLimiter) Penalize(ctx context.Context, duration time.Duration) error {
	if duration < 30*time.Second {
		duration = 30 * time.Second
	}
	const script = `
local ttl = redis.call('PTTL', KEYS[1])
if ttl < tonumber(ARGV[1]) then
  redis.call('SET', KEYS[1], '1', 'PX', ARGV[1])
end
return 1`
	return l.rdb.Eval(ctx, script, []string{l.coolDownKey}, duration.Milliseconds()).Err()
}

func waitContext(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
