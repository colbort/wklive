// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserBankLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserBankLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserBankLogic {
	return &GetUserBankLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserBankLogic) GetUserBank(req *types.GetUserBankReq) (resp *types.GetUserBankResp, err error) {
	// todo: add your logic here and delete this line

	return
}
