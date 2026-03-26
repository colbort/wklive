// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdatePayProductLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdatePayProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePayProductLogic {
	return &UpdatePayProductLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdatePayProductLogic) UpdatePayProduct(req *types.UpdatePayProductReq) (resp *types.RespBase, err error) {
	// todo: add your logic here and delete this line

	return
}
