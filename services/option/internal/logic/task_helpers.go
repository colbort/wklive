package logic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	optionTaskLockTTL           = 30 * time.Second
	optionTaskLockRenewInterval = 10 * time.Second
)

var errOptionTaskLockNotAcquired = errors.New("option task lock not acquired")

func okOptionTaskResp() *option.OptionTaskResp {
	return &option.OptionTaskResp{Base: helper.OkResp()}
}

func runOptionTaskWithLock(
	ctx context.Context,
	svcCtx *svc.ServiceContext,
	taskName string,
	fn func() (*option.OptionTaskResp, error),
) (*option.OptionTaskResp, error) {
	lockKey := fmt.Sprintf("option:task:%s", taskName)
	lockValue := fmt.Sprintf("%d", time.Now().UnixNano())

	if err := acquireOptionTaskLock(ctx, svcCtx.Redis, lockKey, lockValue); err != nil {
		if errors.Is(err, errOptionTaskLockNotAcquired) {
			return &option.OptionTaskResp{
				Base: helper.GetErrResp(1, i18n.Translate(i18n.SyncTaskAlreadyRunning, ctx)),
			}, nil
		}
		logx.Errorf("acquire option task lock failed, key=%s err=%v", lockKey, err)
		return &option.OptionTaskResp{
			Base: helper.GetErrResp(1, i18n.Translate(i18n.DistributedLockAcquireFailed, ctx)),
		}, nil
	}

	renewCtx, renewCancel := context.WithCancel(ctx)
	renewDone := make(chan struct{})
	go func() {
		defer close(renewDone)
		autoRenewOptionTaskLock(renewCtx, svcCtx.Redis, lockKey, lockValue)
	}()
	defer func() {
		renewCancel()
		<-renewDone
		if err := releaseOptionTaskLock(context.Background(), svcCtx.Redis, lockKey, lockValue); err != nil {
			logx.Errorf("release option task lock failed, key=%s err=%v", lockKey, err)
		}
	}()

	return fn()
}

func acquireOptionTaskLock(ctx context.Context, rds *redis.Redis, key, value string) error {
	ok, err := rds.SetnxExCtx(ctx, key, value, int(optionTaskLockTTL.Seconds()))
	if err != nil {
		return err
	}
	if !ok {
		return errOptionTaskLockNotAcquired
	}
	return nil
}

func autoRenewOptionTaskLock(ctx context.Context, rds *redis.Redis, key, value string) {
	ticker := time.NewTicker(optionTaskLockRenewInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := refreshOptionTaskLock(context.Background(), rds, key, value); err != nil {
				logx.Errorf("refresh option task lock failed, key=%s err=%v", key, err)
				return
			}
		}
	}
}

func refreshOptionTaskLock(ctx context.Context, rds *redis.Redis, key, value string) error {
	lua := `
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("PEXPIRE", KEYS[1], ARGV[2])
else
	return 0
end
`
	ret, err := rds.EvalCtx(ctx, lua, []string{key}, value, optionTaskLockTTL.Milliseconds())
	if err != nil {
		return err
	}
	if !redisEvalOK(ret) {
		return errOptionTaskLockNotAcquired
	}
	return nil
}

func releaseOptionTaskLock(ctx context.Context, rds *redis.Redis, key, value string) error {
	lua := `
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
else
	return 0
end
`
	_, err := rds.EvalCtx(ctx, lua, []string{key}, value)
	return err
}

func redisEvalOK(ret any) bool {
	switch v := ret.(type) {
	case int64:
		return v > 0
	case int:
		return v > 0
	case int32:
		return v > 0
	case uint64:
		return v > 0
	default:
		return false
	}
}
