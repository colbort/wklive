// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetWithdrawOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetWithdrawOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetWithdrawOrderLogic {
	return &GetWithdrawOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetWithdrawOrderLogic) GetWithdrawOrder(req *types.GetWithdrawOrderReq) (resp *types.GetWithdrawOrderResp, err error) {
	// todo: add your logic here and delete this line

	return
}
