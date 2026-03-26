// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RetryNotifyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRetryNotifyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RetryNotifyLogic {
	return &RetryNotifyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RetryNotifyLogic) RetryNotify(req *types.RetryNotifyReq) (resp *types.RespBase, err error) {
	// todo: add your logic here and delete this line

	return
}
