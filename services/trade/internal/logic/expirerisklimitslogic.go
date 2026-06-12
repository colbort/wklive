package logic

import (
	"context"

	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/trade"
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
	return runTradeTaskWithLock(l.ctx, l.svcCtx, "expire_risk_limits", func() (*trade.TradeTaskResp, error) {
		now := utils.NowMillis()
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
				if item.Enabled == enableToModel(common.Enable_ENABLE_ENABLED, int64(common.Enable_ENABLE_ENABLED)) && item.EffectiveEndTime > 0 && item.EffectiveEndTime <= now {
					item.Enabled = enableToModel(common.Enable_ENABLE_DISABLED, int64(common.Enable_ENABLE_DISABLED))
					item.UpdateTimes = now
					if err := l.svcCtx.RiskUserTradeLimitModel.Update(l.ctx, item); err != nil {
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
				if item.Enabled == enableToModel(common.Enable_ENABLE_ENABLED, int64(common.Enable_ENABLE_ENABLED)) && item.EffectiveEndTime > 0 && item.EffectiveEndTime <= now {
					item.Enabled = enableToModel(common.Enable_ENABLE_DISABLED, int64(common.Enable_ENABLE_DISABLED))
					item.UpdateTimes = now
					if err := l.svcCtx.RiskUserSymbolLimitModel.Update(l.ctx, item); err != nil {
						return nil, err
					}
				}
			}
			if len(items) < 100 {
				break
			}
		}
		return okTradeTaskResp(), nil
	})
}
