// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteUserBankLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteUserBankLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteUserBankLogic {
	return &DeleteUserBankLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteUserBankLogic) DeleteUserBank(req *types.DeleteUserBankReq) (resp *types.RespBase, err error) {
	// todo: add your logic here and delete this line

	return
}
