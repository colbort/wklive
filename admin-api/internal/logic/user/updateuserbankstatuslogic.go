// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"wklive/admin-api/internal/logicutil"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserBankStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateUserBankStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserBankStatusLogic {
	return &UpdateUserBankStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateUserBankStatusLogic) UpdateUserBankStatus(req *types.UpdateUserBankStatusReq) (resp *types.RespBase, err error) {
	return logicutil.Proxy[types.RespBase](l.ctx, req, l.svcCtx.UserCli.UpdateUserBankStatus)
}
