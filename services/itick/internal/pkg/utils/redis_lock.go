package utils

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrLockNotAcquired = errors.New("lock not acquired")

type RedisLock struct {
	client *redis.Client
}

func NewRedisLock(client *redis.Client) *RedisLock {
	return &RedisLock{
		client: client,
	}
}

// 加锁
func (l *RedisLock) Acquire(ctx context.Context, key, value string, ttl time.Duration) error {
	ok, err := l.client.SetNX(ctx, key, value, ttl).Result()
	if err != nil {
		return err
	}
	if !ok {
		return ErrLockNotAcquired
	}
	return nil
}

// 续期，只给自己的锁续期
func (l *RedisLock) Refresh(ctx context.Context, key, value string, ttl time.Duration) error {
	lua := `
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("PEXPIRE", KEYS[1], ARGV[2])
else
	return 0
end
`
	ret, err := l.client.Eval(ctx, lua, []string{key}, value, ttl.Milliseconds()).Int()
	if err != nil {
		return err
	}
	if ret == 0 {
		return ErrLockNotAcquired
	}
	return nil
}

// 释放锁，只删除自己的锁
func (l *RedisLock) Release(ctx context.Context, key, value string) error {
	lua := `
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
else
	return 0
end
`
	_, err := l.client.Eval(ctx, lua, []string{key}, value).Result()
	return err
}
