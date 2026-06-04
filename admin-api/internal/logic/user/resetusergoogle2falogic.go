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

type ResetUserGoogle2FALogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewResetUserGoogle2FALogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetUserGoogle2FALogic {
	return &ResetUserGoogle2FALogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ResetUserGoogle2FALogic) ResetUserGoogle2FA(req *types.ResetUserGoogle2FAReq) (resp *types.RespBase, err error) {
	return logicutil.Proxy[types.RespBase](l.ctx, req, l.svcCtx.UserCli.ResetUserGoogle2FA)
}
