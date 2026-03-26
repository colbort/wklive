// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user_private

import (
	"context"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"

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
	// todo: add your logic here and delete this line

	return
}
