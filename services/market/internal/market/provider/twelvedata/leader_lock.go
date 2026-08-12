package twelvedata

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

var errLockNotObtained = errors.New("lock not obtained")

type redisLeaderLock struct {
	rdb *redis.Client
	key string
}

func newRedisLeaderLock(rdb *redis.Client, key string) *redisLeaderLock {
	return &redisLeaderLock{rdb: rdb, key: key}
}

func (l *redisLeaderLock) acquire(ctx context.Context, ttl time.Duration) (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(bytes)
	ok, err := l.rdb.SetNX(ctx, l.key, token, ttl).Result()
	if err != nil {
		return "", err
	}
	if !ok {
		return "", errLockNotObtained
	}
	return token, nil
}

func (l *redisLeaderLock) refresh(ctx context.Context, token string, ttl time.Duration) (bool, error) {
	const script = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("PEXPIRE", KEYS[1], ARGV[2])
end
return 0
`
	value, err := l.rdb.Eval(ctx, script, []string{l.key}, token, ttl.Milliseconds()).Int()
	return value == 1, err
}

func (l *redisLeaderLock) release(ctx context.Context, token string) error {
	const script = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
end
return 0
`
	return l.rdb.Eval(ctx, script, []string{l.key}, token).Err()
}
