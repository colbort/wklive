// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user_private

import (
	"context"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteBankLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteBankLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteBankLogic {
	return &DeleteBankLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteBankLogic) DeleteBank(req *types.DeleteBankReq) (resp *types.CommonResp, err error) {
	// todo: add your logic here and delete this line

	return
}
