// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListUserRechargeStatsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListUserRechargeStatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUserRechargeStatsLogic {
	return &ListUserRechargeStatsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListUserRechargeStatsLogic) ListUserRechargeStats(req *types.ListUserRechargeStatsReq) (resp *types.ListUserRechargeStatsResp, err error) {
	// todo: add your logic here and delete this line

	return
}
