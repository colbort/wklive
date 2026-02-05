// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package system

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type Google2FABindInitLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGoogle2FABindInitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *Google2FABindInitLogic {
	return &Google2FABindInitLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *Google2FABindInitLogic) Google2FABindInit(req *types.Google2FABindInitReq) (resp *types.Google2FABindInitResp, err error) {
	// todo: add your logic here and delete this line

	return
}
