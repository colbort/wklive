// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user_private

import (
	"context"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChangeLoginPasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChangeLoginPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangeLoginPasswordLogic {
	return &ChangeLoginPasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChangeLoginPasswordLogic) ChangeLoginPassword(req *types.ChangeLoginPasswordReq) (resp *types.CommonResp, err error) {
	// todo: add your logic here and delete this line

	return
}
