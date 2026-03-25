// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetDefaultUserBankLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSetDefaultUserBankLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetDefaultUserBankLogic {
	return &SetDefaultUserBankLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SetDefaultUserBankLogic) SetDefaultUserBank(req *types.SetDefaultUserBankReq) (resp *types.RespBase, err error) {
	// todo: add your logic here and delete this line

	return
}
