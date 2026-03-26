// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMyPayOrdersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListMyPayOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyPayOrdersLogic {
	return &ListMyPayOrdersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListMyPayOrdersLogic) ListMyPayOrders(req *types.ListMyPayOrdersReq) (resp *types.ListMyPayOrdersResp, err error) {
	// todo: add your logic here and delete this line

	return
}
