// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user_private

import (
	"context"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"

	"wklive/app-api/internal/logicutil"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddBankLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddBankLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddBankLogic {
	return &AddBankLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddBankLogic) AddBank(req *types.AddBankReq) (resp *types.AddBankResp, err error) {
	return logicutil.Proxy[types.AddBankResp](l.ctx, req, l.svcCtx.UserCli.AddBank)
}
