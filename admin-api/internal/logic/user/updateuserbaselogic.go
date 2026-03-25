// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserBaseLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateUserBaseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserBaseLogic {
	return &UpdateUserBaseLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateUserBaseLogic) UpdateUserBase(req *types.UpdateUserBaseReq) (resp *types.UpdateUserBaseResp, err error) {
	// todo: add your logic here and delete this line

	return
}
