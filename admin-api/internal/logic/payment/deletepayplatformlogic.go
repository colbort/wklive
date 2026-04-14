// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeletePayPlatformLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeletePayPlatformLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeletePayPlatformLogic {
	return &DeletePayPlatformLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeletePayPlatformLogic) DeletePayPlatform(req *types.DeletePayPlatformReq) (resp *types.RespBase, err error) {
	return
}
