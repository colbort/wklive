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

type InitGoogle2FALogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewInitGoogle2FALogic(ctx context.Context, svcCtx *svc.ServiceContext) *InitGoogle2FALogic {
	return &InitGoogle2FALogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *InitGoogle2FALogic) InitGoogle2FA() (resp *types.InitGoogle2FAResp, err error) {
	return logicutil.Proxy[types.InitGoogle2FAResp](l.ctx, nil, l.svcCtx.UserCli.InitGoogle2FA)
}
