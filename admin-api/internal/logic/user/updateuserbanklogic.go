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

type UpdateUserBankLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateUserBankLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserBankLogic {
	return &UpdateUserBankLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateUserBankLogic) UpdateUserBank(req *types.UpdateUserBankReq) (resp *types.UpdateUserBankResp, err error) {
	return logicutil.Proxy[types.UpdateUserBankResp](l.ctx, req, l.svcCtx.UserCli.UpdateUserBank)
}
