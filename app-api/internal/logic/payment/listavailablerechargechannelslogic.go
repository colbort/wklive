// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAvailableRechargeChannelsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListAvailableRechargeChannelsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAvailableRechargeChannelsLogic {
	return &ListAvailableRechargeChannelsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListAvailableRechargeChannelsLogic) ListAvailableRechargeChannels(req *types.ListAvailableRechargeChannelsReq) (resp *types.ListAvailableRechargeChannelsResp, err error) {
	// todo: add your logic here and delete this line

	return
}
