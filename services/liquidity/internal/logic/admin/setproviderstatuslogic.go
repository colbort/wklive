package adminlogic

import (
	"context"
	"fmt"
	"time"

	"wklive/common/helper"
	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetProviderStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetProviderStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetProviderStatusLogic {
	return &SetProviderStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SetProviderStatusLogic) SetProviderStatus(in *liquidity.SetProviderStatusReq) (*liquidity.CommonResp, error) {
	if in.Status != liquidity.ProviderStatus_PROVIDER_STATUS_ENABLED &&
		in.Status != liquidity.ProviderStatus_PROVIDER_STATUS_DISABLED {
		return nil, fmt.Errorf("invalid provider status")
	}
	row, err := l.svcCtx.ProviderModel.FindOne(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}
	if row.Version != in.Version {
		return nil, fmt.Errorf("provider version conflict")
	}
	if in.Status == liquidity.ProviderStatus_PROVIDER_STATUS_ENABLED &&
		row.LastHealthStatus == int64(liquidity.HealthStatus_HEALTH_STATUS_UNHEALTHY) {
		return nil, fmt.Errorf("unhealthy provider cannot be enabled")
	}
	row.Status, row.Version, row.UpdateTimes = int64(in.Status), row.Version+1, time.Now().UnixMilli()
	if err := l.svcCtx.ProviderModel.Update(l.ctx, row); err != nil {
		return nil, err
	}
	return &liquidity.CommonResp{Base: helper.OkResp()}, nil
}
