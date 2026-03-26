// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user_private

import (
	"context"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetDefaultBankLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSetDefaultBankLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetDefaultBankLogic {
	return &SetDefaultBankLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SetDefaultBankLogic) SetDefaultBank(req *types.SetDefaultBankReq) (resp *types.CommonResp, err error) {
	// todo: add your logic here and delete this line

	return
}
