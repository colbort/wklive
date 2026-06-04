// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"wklive/admin-api/internal/logicutil"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPayPlatformLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPayPlatformLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPayPlatformLogic {
	return &GetPayPlatformLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPayPlatformLogic) GetPayPlatform(req *types.GetPayPlatformReq) (resp *types.GetPayPlatformResp, err error) {
	return logicutil.Proxy[types.GetPayPlatformResp](l.ctx, req, l.svcCtx.PaymentCli.GetPayPlatform)
}
