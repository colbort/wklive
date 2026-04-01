// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListWithdrawOrdersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListWithdrawOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListWithdrawOrdersLogic {
	return &ListWithdrawOrdersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListWithdrawOrdersLogic) ListWithdrawOrders(req *types.ListWithdrawOrdersReq) (resp *types.ListWithdrawOrdersResp, err error) {
	// todo: add your logic here and delete this line

	return
}
