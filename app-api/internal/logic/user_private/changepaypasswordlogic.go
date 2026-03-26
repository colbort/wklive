// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user_private

import (
	"context"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChangePayPasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChangePayPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangePayPasswordLogic {
	return &ChangePayPasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChangePayPasswordLogic) ChangePayPassword(req *types.ChangePayPasswordReq) (resp *types.RespBase, err error) {
	// todo: add your logic here and delete this line

	return
}
