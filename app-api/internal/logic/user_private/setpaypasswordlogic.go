// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user_private

import (
	"context"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetPayPasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSetPayPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetPayPasswordLogic {
	return &SetPayPasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SetPayPasswordLogic) SetPayPassword(req *types.SetPayPasswordReq) (resp *types.RespBase, err error) {
	// todo: add your logic here and delete this line

	return
}
