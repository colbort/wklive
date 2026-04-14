// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeletePayProductLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeletePayProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeletePayProductLogic {
	return &DeletePayProductLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeletePayProductLogic) DeletePayProduct() (resp *types.RespBase, err error) {
	// todo: add your logic here and delete this line

	return
}
