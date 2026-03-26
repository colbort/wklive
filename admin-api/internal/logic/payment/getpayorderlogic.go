// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPayOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPayOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPayOrderLogic {
	return &GetPayOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPayOrderLogic) GetPayOrder(req *types.GetPayOrderReq) (resp *types.GetPayOrderResp, err error) {
	// todo: add your logic here and delete this line

	return
}
