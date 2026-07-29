// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package market

import (
	"context"

	"wklive/admin-api/internal/logicutil"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SyncProductKlineHistoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSyncProductKlineHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncProductKlineHistoryLogic {
	return &SyncProductKlineHistoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SyncProductKlineHistoryLogic) SyncProductKlineHistory(req *types.SyncProductKlineHistoryReq) (resp *types.SyncProductKlineHistoryResp, err error) {
	return logicutil.Proxy[types.SyncProductKlineHistoryResp](l.ctx, req, l.svcCtx.MarketCli.SyncProductKlineHistory)
}
