// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRechargeOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetRechargeOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRechargeOrderLogic {
	return &GetRechargeOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetRechargeOrderLogic) GetRechargeOrder(req *types.GetRechargeOrderReq) (resp *types.GetRechargeOrderResp, err error) {
	// todo: add your logic here and delete this line

	return
}
