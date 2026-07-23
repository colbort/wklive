package tasklogic

import (
	"context"
	"fmt"
	"time"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/staking"
	"wklive/services/staking/internal/svc"
	"wklive/services/staking/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	stakingTaskLockTTL           = 30 * time.Second
	stakingTaskLockRenewInterval = 10 * time.Second
)

func okTaskResp() *staking.StakingTaskResp {
	return &staking.StakingTaskResp{Base: helper.OkResp()}
}

func runTaskWithLock(
	ctx context.Context,
	svcCtx *svc.ServiceContext,
	taskName string,
	fn func() (*staking.StakingTaskResp, error),
) (*staking.StakingTaskResp, error) {
	lockKey := fmt.Sprintf("staking:task:%s", taskName)
	lockValue := fmt.Sprintf("%d", time.Now().UnixNano())

	if err := acquireTaskLock(ctx, svcCtx.Redis, lockKey, lockValue); err != nil {
		if i18n.IsStatusError(err, i18n.SyncTaskAlreadyRunning) {
			return &staking.StakingTaskResp{
				Base: helper.ErrResp(i18n.SyncTaskAlreadyRunning, i18n.Translate(i18n.SyncTaskAlreadyRunning, ctx)),
			}, nil
		}
		logx.Errorf("acquire staking task lock failed, key=%s err=%v", lockKey, err)
		return &staking.StakingTaskResp{
			Base: helper.ErrResp(i18n.DistributedLockAcquireFailed, i18n.Translate(i18n.DistributedLockAcquireFailed, ctx)),
		}, nil
	}

	renewCtx, renewCancel := context.WithCancel(ctx)
	renewDone := make(chan struct{})
	go func() {
		defer close(renewDone)
		autoRenewTaskLock(renewCtx, svcCtx.Redis, lockKey, lockValue)
	}()
	defer func() {
		renewCancel()
		<-renewDone
		if err := releaseTaskLock(context.Background(), svcCtx.Redis, lockKey, lockValue); err != nil {
			logx.Errorf("release staking task lock failed, key=%s err=%v", lockKey, err)
		}
	}()

	return fn()
}

func acquireTaskLock(ctx context.Context, rds *redis.Redis, key, value string) error {
	ok, err := rds.SetnxExCtx(ctx, key, value, int(stakingTaskLockTTL.Seconds()))
	if err != nil {
		return err
	}
	if !ok {
		return i18n.StatusError(ctx, i18n.SyncTaskAlreadyRunning)
	}
	return nil
}

func autoRenewTaskLock(ctx context.Context, rds *redis.Redis, key, value string) {
	ticker := time.NewTicker(stakingTaskLockRenewInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := refreshTaskLock(context.Background(), rds, key, value); err != nil {
				logx.Errorf("refresh staking task lock failed, key=%s err=%v", key, err)
				return
			}
		}
	}
}

func refreshTaskLock(ctx context.Context, rds *redis.Redis, key, value string) error {
	lua := `
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("PEXPIRE", KEYS[1], ARGV[2])
else
	return 0
end
`
	ret, err := rds.EvalCtx(ctx, lua, []string{key}, value, stakingTaskLockTTL.Milliseconds())
	if err != nil {
		return err
	}
	if !redisEvalOK(ret) {
		return i18n.StatusError(ctx, i18n.SyncTaskAlreadyRunning)
	}
	return nil
}

func releaseTaskLock(ctx context.Context, rds *redis.Redis, key, value string) error {
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

func calcTaskReward(order *models.TStakeOrder, days int64) decimal.Decimal {
	if order == nil || !order.StakeAmount.IsPositive() || !order.Apr.IsPositive() || days <= 0 {
		return decimal.Zero
	}
	return order.StakeAmount.Mul(order.Apr).Mul(decimal.NewFromInt(days)).Div(decimal.NewFromInt(36500))
}
