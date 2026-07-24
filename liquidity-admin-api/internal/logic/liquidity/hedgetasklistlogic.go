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

type HedgeTaskListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHedgeTaskListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HedgeTaskListLogic {
	return &HedgeTaskListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HedgeTaskListLogic) HedgeTaskList(req *types.PageQuery) (resp *types.HedgeListResp, err error) {
	tenantID, err := logicutil.TenantID(l.ctx)
	if err != nil {
		return nil, err
	}
	out, err := l.svcCtx.LiquidityCli.GetHedgeTaskList(l.ctx, &pb.GetHedgeTaskListReq{
		TenantId: tenantID, Status: pb.HedgeStatus(req.Status), Cursor: req.Cursor, Limit: listLimit(req.Limit),
	})
	if err != nil {
		return nil, err
	}
	return logicutil.Convert[types.HedgeListResp](out), nil
}
