package client

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrLockNotObtained = errors.New("lock not obtained")

type RedisLeaderLock struct {
	rdb *redis.Client
	key string
}

func NewRedisLeaderLock(rdb *redis.Client, key string) *RedisLeaderLock {
	return &RedisLeaderLock{
		rdb: rdb,
		key: key,
	}
}

func randomToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (l *RedisLeaderLock) Acquire(ctx context.Context, ttl time.Duration) (string, error) {
	token, err := randomToken()
	if err != nil {
		return "", err
	}

	ok, err := l.rdb.SetNX(ctx, l.key, token, ttl).Result()
	if err != nil {
		return "", err
	}
	if !ok {
		return "", ErrLockNotObtained
	}

	return token, nil
}

func (l *RedisLeaderLock) Refresh(ctx context.Context, token string, ttl time.Duration) (bool, error) {
	lua := `
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("PEXPIRE", KEYS[1], ARGV[2])
else
	return 0
end
`
	n, err := l.rdb.Eval(ctx, lua, []string{l.key}, token, ttl.Milliseconds()).Int()
	if err != nil {
		return false, err
	}
	return n == 1, nil
}

func (l *RedisLeaderLock) Release(ctx context.Context, token string) error {
	lua := `
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
else
	return 0
end
`
	return l.rdb.Eval(ctx, lua, []string{l.key}, token).Err()
}
