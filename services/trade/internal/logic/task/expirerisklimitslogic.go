package tasklogic

import (
	"context"

	"wklive/proto/trade"
	"wklive/services/trade/internal/logic/helpers"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ExpireRiskLimitsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewExpireRiskLimitsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExpireRiskLimitsLogic {
	return &ExpireRiskLimitsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 风控限制过期恢复
func (l *ExpireRiskLimitsLogic) ExpireRiskLimits(in *trade.TradeTaskReq) (*trade.TradeTaskResp, error) {
	return helpers.RunTaskWithLock(l.ctx, l.svcCtx, "expire_risk_limits", func(taskCtx context.Context) (*trade.TradeTaskResp, error) {
		l.ctx = taskCtx
		now := expiryNow()
		cursor := int64(0)
		for {
			items, _, err := l.svcCtx.RiskUserTradeLimitModel.FindPage(l.ctx, cursor, 100)
			if err != nil {
				return nil, err
			}
			if len(items) == 0 {
				break
			}
			for _, item := range items {
				cursor = item.Id
				if in.GetTenantId() > 0 && item.TenantId != in.GetTenantId() {
					continue
				}
				if item.EffectiveEndTime > 0 && item.EffectiveEndTime <= now {
					if _, err := expireProductControl(l.ctx, l.svcCtx, item.Id, in.GetTenantId(), 0, now); err != nil {
						return nil, err
					}
				}
			}
			if len(items) < 100 {
				break
			}
		}
		cursor = 0
		for {
			items, _, err := l.svcCtx.RiskUserSymbolLimitModel.FindPage(l.ctx, cursor, 100)
			if err != nil {
				return nil, err
			}
			if len(items) == 0 {
				break
			}
			for _, item := range items {
				cursor = item.Id
				if in.GetTenantId() > 0 && item.TenantId != in.GetTenantId() {
					continue
				}
				if item.EffectiveEndTime > 0 && item.EffectiveEndTime <= now {
					if _, err := expireSymbolControl(l.ctx, l.svcCtx, item.Id, in.GetTenantId(), 0, now); err != nil {
						return nil, err
					}
				}
			}
			if len(items) < 100 {
				break
			}
		}
		return helpers.OkTaskResp(), nil
	})
}
