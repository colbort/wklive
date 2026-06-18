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
	if strings.TrimSpace(in.ApiUrl) == "" {
		return &itick.SyncKlinesResp{
			Base: helper.GetErrResp(i18n.ApiURLRequired, i18n.Translate(i18n.ApiURLRequired, l.ctx)),
		}, nil
	}
	if strings.TrimSpace(in.ApiToken) == "" {
		return &itick.SyncKlinesResp{
			Base: helper.GetErrResp(i18n.ApiTokenRequired, i18n.Translate(i18n.ApiTokenRequired, l.ctx)),
		}, nil
	}

	// 业务维度锁，避免相同源重复同步
	sum := md5.Sum([]byte(strings.TrimSpace(in.ApiUrl) + "|" + strings.TrimSpace(in.ApiToken)))
	lockKey := fmt.Sprintf("itick:sync_klines:%x", sum)
	lockValue := fmt.Sprintf("%d", time.Now().UnixNano())

	distLock := utils.NewRedisLock(l.svcCtx.LockRedis)

	// 先抢锁，锁初始 30s，后续由 worker 自动续期
	if err := distLock.Acquire(l.ctx, lockKey, lockValue, 30*time.Second); err != nil {
		if errors.Is(err, utils.ErrLockNotAcquired) {
			return &itick.SyncKlinesResp{
				Base: helper.GetErrResp(i18n.SyncTaskAlreadyRunning, i18n.Translate(i18n.SyncTaskAlreadyRunning, l.ctx)),
			}, nil
		}

		logx.Errorf("acquire lock failed, key=%s err=%v", lockKey, err)
		return &itick.SyncKlinesResp{
			Base: helper.GetErrResp(i18n.DistributedLockAcquireFailed, i18n.Translate(i18n.DistributedLockAcquireFailed, l.ctx)),
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
			Base: helper.GetErrResp(i18n.SyncTaskCreateFailed, i18n.Translate(i18n.SyncTaskCreateFailed, l.ctx)),
		}, nil
	}

	reqCopy := &itick.SyncKlinesReq{
		ApiUrl:   in.GetApiUrl(),
		ApiToken: in.GetApiToken(),
		WsUrl:    in.GetWsUrl(),
	}

	go func(taskNo string, reqCopy *itick.SyncKlinesReq, lockKey, lockValue string) {
		bgCtx, cancel := context.WithTimeout(context.Background(), 12*time.Hour)
		defer cancel()

		worker := NewSyncKlinesWorker(bgCtx, l.svcCtx, distLock, lockKey, lockValue)
		worker.Run(taskNo, reqCopy)
	}(taskNo, reqCopy, lockKey, lockValue)

	return &itick.SyncKlinesResp{
		Base: helper.OkResp(),
	}, nil
}
