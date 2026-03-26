// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateRechargeOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateRechargeOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateRechargeOrderLogic {
	return &CreateRechargeOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateRechargeOrderLogic) CreateRechargeOrder(req *types.CreateRechargeOrderReq) (resp *types.CreateRechargeOrderResp, err error) {
	// todo: add your logic here and delete this line

	return
}
