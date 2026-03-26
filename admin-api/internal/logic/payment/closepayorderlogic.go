// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ClosePayOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewClosePayOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ClosePayOrderLogic {
	return &ClosePayOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ClosePayOrderLogic) ClosePayOrder(req *types.ClosePayOrderReq) (resp *types.RespBase, err error) {
	// todo: add your logic here and delete this line

	return
}
