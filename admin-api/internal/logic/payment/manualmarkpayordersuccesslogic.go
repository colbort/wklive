// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ManualMarkPayOrderSuccessLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewManualMarkPayOrderSuccessLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ManualMarkPayOrderSuccessLogic {
	return &ManualMarkPayOrderSuccessLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ManualMarkPayOrderSuccessLogic) ManualMarkPayOrderSuccess(req *types.ManualMarkPayOrderSuccessReq) (resp *types.RespBase, err error) {
	// todo: add your logic here and delete this line

	return
}
