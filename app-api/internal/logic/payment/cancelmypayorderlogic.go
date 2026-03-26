// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelMyPayOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCancelMyPayOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelMyPayOrderLogic {
	return &CancelMyPayOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CancelMyPayOrderLogic) CancelMyPayOrder(req *types.CancelMyPayOrderReq) (resp *types.RespBase, err error) {
	// todo: add your logic here and delete this line

	return
}
