// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMyPayOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMyPayOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyPayOrderLogic {
	return &GetMyPayOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMyPayOrderLogic) GetMyPayOrder(req *types.GetMyPayOrderReq) (resp *types.GetMyPayOrderResp, err error) {
	// todo: add your logic here and delete this line

	return
}
