// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package system

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type Google2FADisableLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGoogle2FADisableLogic(ctx context.Context, svcCtx *svc.ServiceContext) *Google2FADisableLogic {
	return &Google2FADisableLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *Google2FADisableLogic) Google2FADisable(req *types.Google2FADisableReq) (resp *types.SimpleResp, err error) {
	// todo: add your logic here and delete this line

	return
}
