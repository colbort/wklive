package logic

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"strings"
	"time"
	"wklive/common/helper"
	"wklive/common/i18n"
	cutils "wklive/common/utils"
	"wklive/proto/itick"
	"wklive/services/itick/internal/pkg/utils"
	"wklive/services/itick/internal/svc"
	"wklive/services/itick/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type SyncKlinesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSyncKlinesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncKlinesLogic {
	return &SyncKlinesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 同步K线数据 （定时任务）
func (l *SyncKlinesLogic) SyncKlines(in *itick.SyncKlinesReq) (*itick.SyncKlinesResp, error) {
	if strings.TrimSpace(l.svcCtx.Config.Itick.ApiUrl) == "" {
		return &itick.SyncKlinesResp{
			Base: helper.ErrResp(i18n.ApiURLRequired, i18n.Translate(i18n.ApiURLRequired, l.ctx)),
		}, nil
	}
	if strings.TrimSpace(l.svcCtx.Config.Itick.Token) == "" {
		return &itick.SyncKlinesResp{
			Base: helper.ErrResp(i18n.ApiTokenRequired, i18n.Translate(i18n.ApiTokenRequired, l.ctx)),
		}, nil
	}

	// 业务维度锁，避免相同源重复同步
	sum := md5.Sum([]byte(strings.TrimSpace(l.svcCtx.Config.Itick.ApiUrl) + "|" + strings.TrimSpace(l.svcCtx.Config.Itick.Token)))
	lockKey := fmt.Sprintf("itick:sync_klines:%x", sum)
	lockValue := fmt.Sprintf("%d", time.Now().UnixNano())

	distLock := utils.NewRedisLock(l.svcCtx.LockRedis)

	// 先抢锁，锁初始 30s，后续由 worker 自动续期
	if err := distLock.Acquire(l.ctx, lockKey, lockValue, 30*time.Second); err != nil {
		if errors.Is(err, utils.ErrLockNotAcquired) {
			return &itick.SyncKlinesResp{
				Base: helper.ErrResp(i18n.SyncTaskAlreadyRunning, i18n.Translate(i18n.SyncTaskAlreadyRunning, l.ctx)),
			}, nil
		}

		logx.Errorf("acquire lock failed, key=%s err=%v", lockKey, err)
		return &itick.SyncKlinesResp{
			Base: helper.ErrResp(i18n.DistributedLockAcquireFailed, i18n.Translate(i18n.DistributedLockAcquireFailed, l.ctx)),
		}, nil
	}

	taskNo := fmt.Sprintf("sync_klines_%d", time.Now().UnixNano())
	now := cutils.NowMillis()

	_, err := l.svcCtx.ItickSyncTaskModel.Insert(l.ctx, &models.TItickSyncTask{
		TaskNo:      taskNo,
		TaskType:    "sync_klines",
		BizId:       0,
		Status:      0,
		Message:     "任务已提交",
		CreateTimes: now,
		UpdateTimes: now,
	})
	if err != nil {
		_ = distLock.Release(l.ctx, lockKey, lockValue)

		logx.Errorf("create sync task failed, err=%v", err)
		return &itick.SyncKlinesResp{
			Base: helper.ErrResp(i18n.SyncTaskCreateFailed, i18n.Translate(i18n.SyncTaskCreateFailed, l.ctx)),
		}, nil
	}

	go func(taskNo string, apiUrl string, token string, lockKey, lockValue string) {
		bgCtx, cancel := context.WithTimeout(context.Background(), 12*time.Hour)
		defer cancel()

		worker := NewSyncKlinesWorker(bgCtx, l.svcCtx, distLock, lockKey, lockValue)
		worker.Run(taskNo, apiUrl, token)
	}(taskNo, l.svcCtx.Config.Itick.ApiUrl, l.svcCtx.Config.Itick.Token, lockKey, lockValue)

	return &itick.SyncKlinesResp{
		Base: helper.OkResp(),
	}, nil
}
