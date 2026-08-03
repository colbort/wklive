package applogic

import (
	"context"
	"wklive/services/trade/internal/logic/helpers"

	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CheckOrderRiskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckOrderRiskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckOrderRiskLogic {
	return &CheckOrderRiskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 校验订单风控
func (l *CheckOrderRiskLogic) CheckOrderRisk(in *trade.CheckOrderRiskReq) (*trade.CheckOrderRiskResp, error) {
	return helpers.EvaluateAndLogUserOrderRisk(l.ctx, l.svcCtx, in)
}
