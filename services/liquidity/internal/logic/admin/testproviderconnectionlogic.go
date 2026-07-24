package adminlogic

import (
	"context"
	"time"

	"wklive/common/helper"
	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type TestProviderConnectionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTestProviderConnectionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TestProviderConnectionLogic {
	return &TestProviderConnectionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *TestProviderConnectionLogic) TestProviderConnection(in *liquidity.TestProviderConnectionReq) (*liquidity.ProviderHealthResp, error) {
	row, err := l.svcCtx.ProviderModel.FindOne(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}
	now := time.Now().UnixMilli()
	health := liquidity.HealthStatus_HEALTH_STATUS_HEALTHY
	message := "internal provider configuration is valid"
	if liquidity.ProviderType(row.ProviderType) == liquidity.ProviderType_PROVIDER_TYPE_EXTERNAL {
		adapter, adapterErr := l.svcCtx.ProviderAdapters.Get(row.VenueCode)
		if adapterErr != nil {
			health = liquidity.HealthStatus_HEALTH_STATUS_UNHEALTHY
			message = adapterErr.Error()
		} else if healthErr := adapter.Health(l.ctx, row); healthErr != nil {
			health = liquidity.HealthStatus_HEALTH_STATUS_UNHEALTHY
			message = healthErr.Error()
		} else {
			message = "external provider connection is healthy"
		}
	}
	row.LastHealthStatus, row.LastHealthAt, row.UpdateTimes = int64(health), now, now
	if health == liquidity.HealthStatus_HEALTH_STATUS_HEALTHY {
		row.LastErrorMsg = ""
	} else {
		row.LastErrorMsg = message
	}
	row.Version++
	if err := l.svcCtx.ProviderModel.Update(l.ctx, row); err != nil {
		return nil, err
	}
	return &liquidity.ProviderHealthResp{
		Base: helper.OkResp(), HealthStatus: health, CheckedAt: now, Message: message,
	}, nil
}
