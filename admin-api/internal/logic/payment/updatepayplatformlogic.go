// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdatePayPlatformLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdatePayPlatformLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePayPlatformLogic {
	return &UpdatePayPlatformLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdatePayPlatformLogic) UpdatePayPlatform(req *types.UpdatePayPlatformReq) (resp *types.RespBase, err error) {
	// todo: add your logic here and delete this line

	return
}
