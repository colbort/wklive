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

type AddUserBankLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddUserBankLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddUserBankLogic {
	return &AddUserBankLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddUserBankLogic) AddUserBank(req *types.AddUserBankReq) (resp *types.AddUserBankResp, err error) {
	return logicutil.Proxy[types.AddUserBankResp](l.ctx, req, l.svcCtx.UserCli.AddUserBank)
}
