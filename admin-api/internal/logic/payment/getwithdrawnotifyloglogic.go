// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetWithdrawNotifyLogLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetWithdrawNotifyLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetWithdrawNotifyLogLogic {
	return &GetWithdrawNotifyLogLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetWithdrawNotifyLogLogic) GetWithdrawNotifyLog(req *types.GetWithdrawNotifyLogReq) (resp *types.GetWithdrawNotifyLogResp, err error) {
	// todo: add your logic here and delete this line

	return
}
