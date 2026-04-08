package logic

import (
	"context"

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
	// todo: add your logic here and delete this line

	return &trade.CheckOrderRiskResp{}, nil
}
