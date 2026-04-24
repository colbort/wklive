package logic

import (
	"context"

	"wklive/proto/tenant"
	"wklive/services/tenant/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChangeUserStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewChangeUserStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangeUserStatusLogic {
	return &ChangeUserStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 修改租户状态
func (l *ChangeUserStatusLogic) ChangeUserStatus(in *tenant.ChangeUserStatusReq) (*tenant.RespBase, error) {
	// todo: add your logic here and delete this line

	return &tenant.RespBase{}, nil
}
