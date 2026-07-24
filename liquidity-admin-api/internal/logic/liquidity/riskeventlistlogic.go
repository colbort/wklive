// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package liquidity

import (
	"context"

	"wklive/liquidity-admin-api/internal/logicutil"
	"wklive/liquidity-admin-api/internal/svc"
	"wklive/liquidity-admin-api/internal/types"
	pb "wklive/proto/liquidity"

	"github.com/zeromicro/go-zero/core/logx"
)

type RiskEventListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRiskEventListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RiskEventListLogic {
	return &RiskEventListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RiskEventListLogic) RiskEventList(req *types.PageQuery) (resp *types.RiskListResp, err error) {
	out, err := l.svcCtx.LiquidityCli.GetRiskEventList(l.ctx, &pb.GetRiskEventListReq{
		Status: pb.RiskEventStatus(req.Status), Cursor: req.Cursor, Limit: listLimit(req.Limit),
	})
	if err != nil {
		return nil, err
	}
	return logicutil.Convert[types.RiskListResp](out), nil
}
