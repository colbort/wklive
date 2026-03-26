// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserRechargeStatLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserRechargeStatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserRechargeStatLogic {
	return &GetUserRechargeStatLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserRechargeStatLogic) GetUserRechargeStat(req *types.GetUserRechargeStatReq) (resp *types.GetUserRechargeStatResp, err error) {
	// todo: add your logic here and delete this line

	return
}
