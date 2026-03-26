// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPayNotifyLogLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPayNotifyLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPayNotifyLogLogic {
	return &GetPayNotifyLogLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPayNotifyLogLogic) GetPayNotifyLog(req *types.GetPayNotifyLogReq) (resp *types.GetPayNotifyLogResp, err error) {
	// todo: add your logic here and delete this line

	return
}
