// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryMyPayOrderStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryMyPayOrderStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryMyPayOrderStatusLogic {
	return &QueryMyPayOrderStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryMyPayOrderStatusLogic) QueryMyPayOrderStatus(req *types.QueryMyPayOrderStatusReq) (resp *types.QueryMyPayOrderStatusResp, err error) {
	// todo: add your logic here and delete this line

	return
}
