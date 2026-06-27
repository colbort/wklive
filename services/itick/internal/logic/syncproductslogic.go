package logic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/itick"
	"wklive/services/itick/internal/pkg/utils"
	"wklive/services/itick/internal/svc"
	"wklive/services/itick/models"

	"github.com/zeromicro/go-zero/core/logx"
)

const syncProductsLockKey = "itick:sync_products"

type SyncProductsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSyncProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncProductsLogic {
	return &SyncProductsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 同步产品列表 （定时任务）
func (l *SyncProductsLogic) SyncProducts(in *itick.SyncProductsReq) (*itick.SyncProductsResp, error) {
	lockValue := fmt.Sprintf("%d", time.Now().UnixNano())
	distLock := utils.NewRedisLock(l.svcCtx.LockRedis)
	if err := distLock.Acquire(l.ctx, syncProductsLockKey, lockValue, 30*time.Second); err != nil {
		if errors.Is(err, utils.ErrLockNotAcquired) {
			return &itick.SyncProductsResp{
				Base: helper.ErrResp(i18n.SyncTaskAlreadyRunning, i18n.Translate(i18n.SyncTaskAlreadyRunning, l.ctx)),
			}, nil
		}

		logx.Errorf("acquire lock failed, key=%s err=%v", syncProductsLockKey, err)
		return &itick.SyncProductsResp{
			Base: helper.ErrResp(i18n.DistributedLockAcquireFailed, i18n.Translate(i18n.DistributedLockAcquireFailed, l.ctx)),
		}, nil
	}

	renewCtx, renewCancel := context.WithCancel(l.ctx)
	go l.autoRenewLock(renewCtx, distLock, syncProductsLockKey, lockValue)
	defer func() {
		renewCancel()
		if err := distLock.Release(context.Background(), syncProductsLockKey, lockValue); err != nil {
			logx.Errorf("release lock failed, key=%s err=%v", syncProductsLockKey, err)
		}
	}()

	categories, err := l.svcCtx.ItickCategoryModel.FindAll(l.ctx)
	if err != nil {
		return nil, err
	}

	worker := NewSyncCategoryProductsWorker(l.ctx, l.svcCtx)
	for _, category := range categories {
		if category == nil {
			continue
		}
		if _, err := l.svcCtx.ItickCategoryModel.FindOne(l.ctx, category.Id); err != nil && !errors.Is(err, models.ErrNotFound) {
			return nil, err
		}
		if err := worker.doSync(&itick.SyncCategoryProductsReq{Id: category.Id}); err != nil {
			logx.Errorf("sync products failed, categoryId=%d err=%v", category.Id, err)
			return &itick.SyncProductsResp{
				Base: helper.ErrResp(i18n.SyncTaskFailed, i18n.Translate(i18n.SyncTaskFailed, l.ctx)),
			}, nil
		}
	}

	return &itick.SyncProductsResp{
		Base: helper.OkResp(),
	}, nil
}

func (l *SyncProductsLogic) autoRenewLock(ctx context.Context, lock *utils.RedisLock, key, value string) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := lock.Refresh(context.Background(), key, value, 30*time.Second); err != nil {
				logx.Errorf("refresh lock failed, key=%s err=%v", key, err)
				return
			}
		}
	}
}
